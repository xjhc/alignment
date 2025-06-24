# Development: Testing Patterns & Tools

This document is a practical guide to the testing patterns and tools used in the `Alignment` project. Testing a highly concurrent, actor-based system presents unique challenges, such as race conditions and non-deterministic behavior. The patterns outlined here are designed to create a **stable, deterministic, and maintainable test suite**.

---

## 1. Table-Driven Tests

*   **What is it?** A standard Go pattern for testing a function with a variety of inputs and expected outputs. You define a slice of test case structs, and a single test loop iterates over them.

*   **Why do we use it?** It's the best way to test **pure functions**, like those in our `/core` package. It makes it easy to add new test cases and provides excellent coverage for all edge cases of our game's rules.

*   **How do we use it?** For functions like `core.ApplyEvent`, we define a table of initial states, the event to apply, and the expected final state.

    ```go
    // from: core/game_state_test.go
    func TestApplyEvent(t *testing.T) {
        testCases := []struct {
            name          string
            initialState  core.GameState
            event         core.Event
            expectedState core.GameState
        }{
            {
                name: "mining should add a token",
                initialState: GameState{Players: {"p1": {Tokens: 1}}},
                event: Event{Type: "MINING_SUCCESSFUL", Payload: {"target_id": "p1", "amount": 1}},
                expectedState: GameState{Players: {"p1": {Tokens: 2}}},
            },
            // ... dozens of other test cases for every event type ...
        }

        for _, tc := range testCases {
            t.Run(tc.name, func(t *testing.T) {
                newState := ApplyEvent(tc.initialState, tc.event)
                // Assert that newState matches tc.expectedState
            })
        }
    }
    ```

---

## 2. Mocks and Interfaces

*   **What is it?** An **interface** is a contract that defines a set of methods. A **mock** is a fake object we create in our tests that "pretends" to implement an interface.

*   **Why do we use it?** Mocks are the cornerstone of our integration testing strategy. They allow us to test a component (like a `PlayerActor`) in **complete isolation** from its real dependencies (like the `LobbyManager` or the Redis `DataStore`). This makes tests fast, reliable, and focused.

*   **How do we use it?**
    1.  **Define the contract:** In `server/internal/interfaces/interfaces.go`, we define what a `SessionManager` can do.
        ```go
        type SessionManagerInterface interface {
            SendActionToGame(gameID string, action core.Action) error
            // ... other methods
        }
        ```
    2.  **Create a central mock:** In `server/internal/mocks/`, we create a mock that fulfills this contract. The `var _ ...` line is a critical compile-time check that ensures our mock never goes out of sync with the interface.
        ```go
        // from: server/internal/mocks/session_manager.go
        import "github.com/stretchr/testify/mock"

        type MockSessionManager struct {
            mock.Mock
        }
        // This line will fail to compile if MockSessionManager is missing a method.
        var _ interfaces.SessionManagerInterface = (*MockSessionManager)(nil)

        func (m *MockSessionManager) SendActionToGame(gameID string, action core.Action) error {
            args := m.Called(gameID, action)
            return args.Error(0)
        }
        ```
    3.  **Inject the mock in tests:** In our tests, we create an instance of the `MockSessionManager` and pass it to the component we're testing.

---

## 3. The `testify/mock` Library

*   **What is it?** A popular Go library that provides a powerful framework for creating and using mocks. We have standardized on this library for all our mocks.

*   **Why do we use it?** It dramatically reduces the boilerplate of writing mocks and gives us powerful tools to:
    *   Specify what arguments we expect a method to be called with.
    *   Define what values a mocked method should return.
    *   Assert that a method was called exactly as expected.

*   **How do we use it?**
    ```go
    // from: player_actor_test.go
    func TestPlayerActor_GameAction(t *testing.T) {
        // 1. Create the mock
        mockSession := &mocks.MockSessionManager{}

        // 2. Set up an expectation:
        // "I expect the SendActionToGame method to be called once with
        // the gameID 'test-game' and any core.Action. When it is called,
        // it should return no error (nil)."
        mockSession.On("SendActionToGame", "test-game", mock.AnythingOfType("core.Action")).Return(nil).Once()

        // 3. Inject the mock and run the code being tested...
        actor := NewPlayerActor(...)
        actor.SetDependencies(nil, mockSession) // LobbyManager is nil for this test
        actor.handleGameAction(core.Action{Type: "SUBMIT_VOTE", GameID: "test-game"})

        // 4. Assert that all expectations were met.
        // This will fail the test if SendActionToGame was not called.
        mockSession.AssertExpectations(t)
    }
    ```

---

## 4. Synchronizing Asynchronous Tests with `sync.WaitGroup`

*   **What is it?** A `sync.WaitGroup` is a simple but powerful tool from Go's standard library. It's a counter that allows one goroutine to wait until a collection of other goroutines have finished their work.

*   **Why do we use it?** **This is the most critical pattern for testing our actors.** Our actors run in their own goroutines. A test cannot simply send an action to an actor's mailbox and immediately check the result. The test goroutine might finish its check *before* the actor goroutine has even processed the action. Using `time.Sleep()` is a recipe for flaky, unreliable tests. The `WaitGroup` solves this by making the test wait deterministically.

*   **How do we use it? (The "Wait, Defer, Add" Pattern)**

    **Step A: Add the WaitGroup to the Mock**
    ```go
    // server/internal/mocks/lobby_manager.go
    type MockLobbyManager struct {
        mock.Mock
        Wg sync.WaitGroup // Add this
    }

    func (m *MockLobbyManager) StartGame(hostPlayerID, lobbyID string) error {
        defer m.Wg.Done() // B. Defer Done() to signal completion
        args := m.Called(hostPlayerID, lobbyID)
        return args.Error(0)
    }
    ```

    **Step B: Use the WaitGroup in the Test**
    ```go
    // server/internal/actors/player_actor_test.go
    func TestPlayerActor_StartGame(t *testing.T) {
        mockLobby := &mocks.MockLobbyManager{}
        // ... other setup ...

        mockLobby.On("StartGame", "host-id", "lobby-id").Return(nil)
        mockLobby.Wg.Add(1) // A. Expect one call to a concurrent function

        // ... send the START_GAME action to the PlayerActor ...

        // C. Wait for the concurrent call to complete (with a timeout)
        waitWithTimeout(&mockLobby.Wg, 1*time.Second, t)

        // Now it's safe to assert
        mockLobby.AssertExpectations(t)
    }

    // Helper function to prevent tests from hanging forever
    func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration, t *testing.T) {
        c := make(chan struct{})
        go func() {
            defer close(c)
            wg.Wait()
        }()
        select {
        case <-c:
            // Completed successfully
        case <-time.After(timeout):
            t.Fatal("Test timed out waiting for WaitGroup")
        }
    }
    ```
This pattern provides a rock-solid, non-flaky way to test our concurrent actors and is the standard we will use going forward.

---

## 5. End-to-End (E2E) Testing Tools

*   **What is it?** A testing practice where we validate a complete user flow through the entire system, from the client UI to the database and back.

*   **Why do we use it?** E2E tests are the ultimate confidence check. They verify that all the isolated components (frontend, actors, database) work together correctly. They are essential for catching regressions in the client-server contract.

*   **How do we use it?**
    *   **Test Runner:** We use **Python** with the `pytest` framework and `websocket-client` library. This allows us to write clean, imperative test scripts that mimic real user behavior.
    *   **Execution Environment:** E2E tests are run against a **live, fully-composed application instance**, typically started with `docker-compose.dev.yml`. This ensures the test environment is identical to a real deployment.
    *   **Example Flow (`tests/e2e/test_lobby.py`):**
        1.  The test script makes a `POST` request to `/api/games` to create a lobby.
        2.  It parses the `game_id` and `session_token` from the response.
        3.  It uses these credentials to establish a real WebSocket connection to the running server.
        4.  It listens for the expected `LOBBY_STATE_UPDATE` event on the WebSocket.
        5.  It asserts that the payload of the event is correct.
    *   **CI Integration:** A dedicated GitHub Actions workflow (`e2e-tests.yml`) is responsible for building the Docker images, running `docker-compose up`, and then executing the `pytest` suite.