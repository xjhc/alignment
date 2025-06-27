

import os
import time
import json
import requests
import websocket
import threading
import pytest
from typing import List, Dict, Any

# --- Configuration ---
BASE_URL = os.environ.get("E2E_BASE_URL", "http://localhost:8080")
WS_URL = os.environ.get("E2E_WS_URL", "ws://localhost:8080/ws")
NUM_PLAYERS = 6

# --- Helper Functions ---

def receive_websocket_messages(ws, received_messages, stop_event, player_id, timeout=20):
    """Receives WebSocket messages in a separate thread for a single player."""
    start_time = time.time()
    try:
        while not stop_event.is_set() and time.time() - start_time < timeout:
            try:
                ws.settimeout(0.5)
                message_str = ws.recv()
                if not message_str:
                    continue
                
                for line in message_str.strip().split('\n'):
                    if not line:
                        continue
                    try:
                        message = json.loads(line)
                        msg_type = message.get("type")
                        if msg_type:
                            if msg_type not in received_messages:
                                received_messages[msg_type] = []
                            received_messages[msg_type].append(message)
                            print(f"  [{player_id}] Received: {msg_type}")
                    except json.JSONDecodeError:
                        print(f"  [{player_id}] Failed to parse message line: {line}")
            except websocket.WebSocketTimeoutException:
                continue
            except Exception as e:
                print(f"  [{player_id}] Error receiving message: {e}")
                break
    finally:
        ws.settimeout(None)

class PlayerClient:
    """A helper class to manage a single player's state and connection."""
    def __init__(self, name: str, avatar: str):
        self.name = name
        self.avatar = avatar
        self.player_id: str = ""
        self.session_token: str = ""
        self.game_id: str = ""
        self.ws: websocket.WebSocket = None
        self.received_messages: Dict[str, List[Any]] = {}
        self.stop_event = threading.Event()
        self.receive_thread = None
        self.role: str = ""
        self.alignment: str = ""

    def create_lobby(self) -> str:
        """Creates a new lobby and sets this player as the host."""
        print(f"  Creating lobby for {self.name}...")
        response = requests.post(
            f"{BASE_URL}/api/games",
            json={"lobby_name": f"{self.name}'s Game", "player_name": self.name, "player_avatar": self.avatar}
        )
        assert response.status_code == 201, f"Failed to create lobby: {response.text}"
        data = response.json()
        self.game_id = data["game_id"]
        self.player_id = data["player_id"]
        self.session_token = data["session_token"]
        print(f"  ✅ Lobby {self.game_id} created by {self.player_id}")
        return self.game_id

    def join_lobby(self, game_id: str):
        """Joins an existing lobby."""
        print(f"  {self.name} joining lobby {game_id}...")
        self.game_id = game_id
        response = requests.post(
            f"{BASE_URL}/api/games/{game_id}/join",
            json={"player_name": self.name, "player_avatar": self.avatar}
        )
        assert response.status_code == 200, f"Failed to join lobby: {response.text}"
        data = response.json()
        self.player_id = data["player_id"]
        self.session_token = data["session_token"]
        print(f"  ✅ {self.name} ({self.player_id}) joined lobby.")

    def connect_websocket(self):
        """Connects to the WebSocket and starts listening for messages."""
        ws_url = f"{WS_URL}?gameId={self.game_id}&playerId={self.player_id}&sessionToken={self.session_token}"
        self.ws = websocket.create_connection(ws_url, timeout=10)
        self.receive_thread = threading.Thread(
            target=receive_websocket_messages,
            args=(self.ws, self.received_messages, self.stop_event, self.player_id)
        )
        self.receive_thread.start()
        print(f"  ✅ {self.name} ({self.player_id}) WebSocket connected.")

    def send_action(self, action_type: str, payload: Dict[str, Any]):
        """Sends a JSON action to the WebSocket."""
        action = {
            "type": action_type,
            "payload": payload
        }
        print(f"  [{self.player_id}] Sending: {action_type} with payload {payload}")
        self.ws.send(json.dumps(action))

    def stop(self):
        """Stops the message listener and closes the connection."""
        if self.receive_thread and self.receive_thread.is_alive():
            self.stop_event.set()
            self.receive_thread.join(timeout=2)
        if self.ws and self.ws.connected:
            self.ws.close()
        print(f"  [{self.player_id}] Connection stopped.")

    def get_events(self, event_type: str) -> List[Any]:
        """Gets all received events of a specific type."""
        return self.received_messages.get(event_type, [])

    def get_latest_event(self, event_type: str) -> Any:
        """Gets the most recently received event of a specific type."""
        events = self.get_events(event_type)
        return events[-1] if events else None

# --- Test Fixture ---

@pytest.fixture(scope="module")
def game_setup():
    """Fixture to set up a game with the specified number of players and tear it down."""
    print("\n--- E2E Game Setup ---")
    
    # Health check loop
    for _ in range(15): # Try for 15 seconds
        try:
            response = requests.get(f"{BASE_URL}/health")
            if response.status_code == 200:
                print("  ✅ Backend is healthy.")
                break
        except requests.ConnectionError:
            pass
        time.sleep(1)
    else:
        pytest.fail("Backend did not become healthy in time.")

    players: List[PlayerClient] = []
    try:
        # Create host and lobby
        host = PlayerClient("Host", "👑")
        game_id = host.create_lobby()
        players.append(host)

        # Create and join other players
        for i in range(2, NUM_PLAYERS + 1):
            player = PlayerClient(f"Player{i}", "🤖")
            player.join_lobby(game_id)
            players.append(player)
        
        # Connect all players to WebSocket
        for player in players:
            player.connect_websocket()
        
        time.sleep(2) # Allow time for all LOBBY_STATE_UPDATEs to arrive

        yield game_id, players

    finally:
        print("\n--- E2E Game Teardown ---")
        for player in players:
            player.stop()

# --- Test Cases ---

def test_full_game_happy_path(game_setup):
    """
    Tests a full game flow from lobby creation to a valid win condition.
    """
    game_id, players = game_setup
    host = players[0]

    # 1. Verify lobby state
    print("\nSTEP 1: Verifying lobby state...")
    for player in players:
        lobby_state = player.get_latest_event("LOBBY_STATE_UPDATE")
        assert lobby_state is not None, f"{player.name} did not receive LOBBY_STATE_UPDATE"
        payload = lobby_state.get("payload", {})
        assert len(payload.get("players", [])) == NUM_PLAYERS, f"{player.name} sees incorrect player count"
        assert payload.get("host_id") == host.player_id, f"{player.name} has incorrect host_id"
    print("  ✅ All players see correct initial lobby state.")

    # 2. Host starts the game
    print("\nSTEP 2: Host starts the game...")
    host.send_action("START_GAME", {})
    time.sleep(5) # Allow time for game start processing and role assignment

    # 3. Verify all players receive game start events
    print("\nSTEP 3: Verifying game start for all players...")
    ai_player = None
    for player in players:
        role_assigned = player.get_latest_event("ROLE_ASSIGNED")
        snapshot = player.get_latest_event("GAME_STATE_UPDATE")
        
        assert role_assigned is not None, f"{player.name} did not receive ROLE_ASSIGNED"
        assert snapshot is not None, f"{player.name} did not receive GAME_STATE_UPDATE"
        
        role_payload = role_assigned.get("payload", {})
        assert "role" in role_payload, f"ROLE_ASSIGNED payload for {player.name} is missing 'role'"
        assert "alignment" in role_payload, f"ROLE_ASSIGNED payload for {player.name} is missing 'alignment'"
        
        player.role = role_payload["role"]["type"]
        player.alignment = role_payload["alignment"]
        
        if player.alignment == "ALIGNED":
            ai_player = player
        
        print(f"  ✅ {player.name} assigned role {player.role} and alignment {player.alignment}")

    assert ai_player is not None, "Could not identify the AI player"
    print(f"  Identified AI player: {ai_player.name} ({ai_player.player_id})")

    # 4. Simulate a round of voting to eliminate the AI
    print("\nSTEP 4: Simulating Day 1 voting...")
    
    # Nomination Phase
    print("  Nominating the AI player...")
    for p in players:
        if p.player_id != ai_player.player_id:  # All players except AI can vote
            p.send_action("SUBMIT_VOTE", {"vote_type": "NOMINATION", "target_id": ai_player.player_id})
            time.sleep(0.1) # Stagger votes slightly

    time.sleep(2) # Allow votes to be processed

    # Verdict Phase
    print("  Voting to eliminate the AI player...")
    for p in players:
        p.send_action("SUBMIT_VOTE", {"vote_type": "VERDICT", "verdict": "GUILTY"})
        time.sleep(0.1)

    time.sleep(3) # Allow verdict to be processed

    # 5. Verify game end
    print("\nSTEP 5: Verifying game end...")
    for p in players:
        game_ended = p.get_latest_event("GAME_ENDED")
        assert game_ended is not None, f"{p.name} did not receive GAME_ENDED event"
        
        end_payload = game_ended.get("payload", {})
        assert end_payload.get("winning_faction") == "HUMANS", f"Expected HUMAN faction to win, but got {end_payload.get('winning_faction')}"
        print(f"  ✅ {p.name} received correct GAME_ENDED event.")


def test_system_shock_flow(game_setup):
    """
    Tests the System Shock flow where a night action fails conversion 
    and the target player receives a PRIVATE_NOTIFICATION event.
    """
    game_id, players = game_setup
    host = players[0]

    # Start the game and get role assignments
    print("\nSTEP 1: Starting game and getting role assignments...")
    host.send_action("START_GAME", {})
    time.sleep(5)

    # Identify AI and human players
    ai_player = None
    human_players = []
    for player in players:
        role_assigned = player.get_latest_event("ROLE_ASSIGNED")
        assert role_assigned is not None, f"{player.name} did not receive ROLE_ASSIGNED"
        
        role_payload = role_assigned.get("payload", {})
        player.alignment = role_payload["alignment"]
        
        if player.alignment == "ALIGNED":
            ai_player = player
        else:
            human_players.append(player)

    assert ai_player is not None, "Could not identify the AI player"
    assert len(human_players) > 0, "No human players found"
    
    target_player = human_players[0]  # AI will try to convert this player
    print(f"  AI player: {ai_player.name}, Target: {target_player.name}")

    # Skip to night phase (this would normally happen through game progression)
    # For testing purposes, we'll simulate a conversion attempt that should fail
    print("\nSTEP 2: Simulating AI conversion attempt that will trigger System Shock...")
    
    # AI attempts conversion
    ai_player.send_action("ATTEMPT_CONVERSION", {"target_id": target_player.player_id})
    time.sleep(3)  # Allow time for night action processing

    # Verify target player receives PRIVATE_NOTIFICATION about system shock
    print("\nSTEP 3: Verifying System Shock notification...")
    private_notification = target_player.get_latest_event("PRIVATE_NOTIFICATION")
    assert private_notification is not None, f"{target_player.name} did not receive PRIVATE_NOTIFICATION"
    
    notification_payload = private_notification.get("payload", {})
    assert notification_payload.get("type") == "system_shock", f"Expected system_shock notification, got {notification_payload.get('type')}"
    
    print(f"  ✅ {target_player.name} received System Shock notification: {notification_payload.get('message')}")


def test_quote_reply_flow(game_setup):
    """
    Tests the Quote-Reply flow where one player replies to another 
    and other clients receive a CHAT_MESSAGE event with correctly formatted quote payload.
    """
    game_id, players = game_setup
    host = players[0]

    # Start the game
    print("\nSTEP 1: Starting game...")
    host.send_action("START_GAME", {})
    time.sleep(3)

    # Set up the quote-reply scenario
    original_sender = players[0]
    replier = players[1]
    observers = players[2:]

    print("\nSTEP 2: Sending original message...")
    original_message = "What do you think about the current situation?"
    original_sender.send_action("SEND_MESSAGE", {
        "message": original_message,
        "channel_id": "general"
    })
    time.sleep(1)

    # Get the original message ID from received messages
    original_chat_event = None
    for player in players:
        chat_events = player.get_events("CHAT_MESSAGE")
        if chat_events:
            original_chat_event = chat_events[-1]
            break
    
    assert original_chat_event is not None, "Original message was not received"
    original_message_id = original_chat_event.get("id") or original_chat_event.get("payload", {}).get("id")
    assert original_message_id is not None, "Could not get original message ID"

    print(f"  Original message ID: {original_message_id}")

    print("\nSTEP 3: Sending reply with quote...")
    reply_message = "I think we need to be more careful about who we trust."
    replier.send_action("SEND_MESSAGE", {
        "message": reply_message,
        "channel_id": "general",
        "quote_reply": {
            "message_id": original_message_id,
            "quoted_text": original_message,
            "quoted_author": original_sender.name
        }
    })
    time.sleep(1)

    print("\nSTEP 4: Verifying quote-reply format...")
    # Check that all observers receive the correctly formatted quote reply
    for observer in observers:
        chat_events = observer.get_events("CHAT_MESSAGE")
        reply_event = None
        
        # Find the reply message (should be the most recent)
        for event in reversed(chat_events):
            event_payload = event.get("payload", {})
            if event_payload.get("message") == reply_message:
                reply_event = event
                break
        
        assert reply_event is not None, f"{observer.name} did not receive the reply message"
        
        reply_payload = reply_event.get("payload", {})
        assert "[quote]" in reply_payload.get("message", ""), f"Reply message for {observer.name} does not contain [quote] format"
        
        # Verify quote structure - the message should contain the quoted content
        message_content = reply_payload.get("message", "")
        assert original_sender.name in message_content, f"Quote in {observer.name}'s message does not contain original author name"
        assert original_message in message_content, f"Quote in {observer.name}'s message does not contain original message text"
        
        print(f"  ✅ {observer.name} received correctly formatted quote-reply")


def test_expanded_game_mechanics(game_setup):
    """
    Tests additional game mechanics beyond the basic happy path.
    """
    game_id, players = game_setup
    host = players[0]

    print("\nSTEP 1: Starting game and role assignment...")
    host.send_action("START_GAME", {})
    time.sleep(5)

    # Verify pulse check mechanics
    print("\nSTEP 2: Testing Pulse Check mechanics...")
    pulse_check_events = []
    for player in players:
        pulse_check = player.get_latest_event("PULSE_CHECK_STARTED")
        if pulse_check:
            pulse_check_events.append(pulse_check)
    
    if pulse_check_events:
        print("  Pulse check initiated, testing responses...")
        # Have all players respond to pulse check
        responses = ["NOMINAL", "ELEVATED", "CRITICAL"]
        for i, player in enumerate(players):
            response = responses[i % len(responses)]
            player.send_action("SUBMIT_PULSE_CHECK", {"response": response})
            time.sleep(0.2)
        
        time.sleep(2)  # Allow pulse check to complete
        
        # Verify pulse check results are broadcast
        for player in players:
            pulse_update = player.get_latest_event("PULSE_CHECK_UPDATED")
            if pulse_update:
                update_payload = pulse_update.get("payload", {})
                assert "responses" in update_payload or "player_responses" in update_payload, "Pulse check update missing response data"
                print(f"  ✅ {player.name} received pulse check update")

    # Test night action mechanics if game progresses to night
    print("\nSTEP 3: Testing Night Action mechanics...")
    for player in players:
        # Check if player has role abilities unlocked
        role_assigned = player.get_latest_event("ROLE_ASSIGNED")
        if role_assigned:
            role_payload = role_assigned.get("payload", {})
            role_info = role_payload.get("role", {})
            if role_info.get("ability") and role_info["ability"].get("isReady"):
                print(f"  {player.name} has ability ready: {role_info['ability']['name']}")
                
                # Test using ability (this would normally be in night phase)
                target_player = players[(players.index(player) + 1) % len(players)]
                player.send_action("USE_ABILITY", {
                    "target_id": target_player.player_id,
                    "ability_type": role_info["ability"]["name"]
                })
                time.sleep(0.5)

    # Test token mining mechanics
    print("\nSTEP 4: Testing Token Mining mechanics...")
    test_player = players[0]
    test_player.send_action("MINE_TOKENS", {"target_player": players[1].player_id})
    time.sleep(1)
    
    # Check for mining result events
    mining_events = test_player.get_events("MINING_SUCCESSFUL") + test_player.get_events("MINING_FAILED")
    if mining_events:
        print(f"  ✅ Mining mechanic working: {len(mining_events)} mining event(s) received")

    print("  ✅ Extended game mechanics test completed")


