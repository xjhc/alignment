import requests
import json
import websocket
import time
import pytest
import threading
from typing import List, Dict, Any

# Configuration
BASE_URL = "http://localhost:8080"
WS_URL = "ws://localhost:8080/ws"

# --- Helper Functions ---

def receive_websocket_messages(ws, received_messages, stop_event, player_id, timeout=10):
    """Receives WebSocket messages in a separate thread for a single player."""
    start_time = time.time()
    try:
        while not stop_event.is_set() and time.time() - start_time < timeout:
            try:
                ws.settimeout(0.5)
                message_str = ws.recv()
                if not message_str:
                    continue
                
                # Handle potentially batched messages
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
        self.ws = websocket.create_connection(ws_url, timeout=5)
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
            "player_id": self.player_id,
            "game_id": self.game_id,
            "timestamp": time.time(),
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

# --- Test Suite ---

@pytest.fixture(scope="module")
def game_setup():
    """Fixture to set up a game with 6 players and tear it down."""
    print("\n--- E2E Game Setup ---")
    
    # Health check loop
    for _ in range(10): # Try for 10 seconds
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
        for i in range(2, 7):
            player = PlayerClient(f"Player{i}", "🤖")
            player.join_lobby(game_id)
            players.append(player)
        
        # Connect all players to WebSocket
        for player in players:
            player.connect_websocket()
        
        # Give a moment for all LOBBY_STATE_UPDATEs to arrive
        time.sleep(2)

        yield game_id, players

    finally:
        print("\n--- E2E Game Teardown ---")
        for player in players:
            player.stop()

def test_full_game_happy_path(game_setup):
    """
    Tests a full game flow from lobby to a win condition.
    """
    game_id, players = game_setup
    host = players[0]

    # 1. Host starts the game
    print("\nSTEP 1: Host starts the game...")
    host.send_action("START_GAME", {})
    time.sleep(3) # Allow time for game start processing

    # 2. Verify all players receive game start events
    print("\nSTEP 2: Verifying game start for all players...")
    for player in players:
        role_assigned = player.get_latest_event("ROLE_ASSIGNED")
        snapshot = player.get_latest_event("GAME_STATE_UPDATE") # Changed from GAME_STATE_SNAPSHOT
        
        assert role_assigned is not None, f"{player.name} did not receive ROLE_ASSIGNED"
        assert snapshot is not None, f"{player.name} did not receive GAME_STATE_UPDATE"
        
        role_payload = role_assigned.get("payload", {})
        assert "role" in role_payload
        assert "alignment" in role_payload
        player.role = role_payload["role"]["type"]
        player.alignment = role_payload["alignment"]
        print(f"  ✅ {player.name} assigned role {player.role} and alignment {player.alignment}")

    # 3. Simulate a few rounds of voting and night actions
    # This is a simplified simulation. A real test would be more dynamic.
    print("\nSTEP 3: Simulating game rounds...")
    
    # Find the AI player to target for elimination
    ai_player = None
    for p in players:
        # In this test, the AI is assigned the 'CTO' role based on game logic
        if p.role == 'CTO':
            ai_player = p
            break
    
    assert ai_player is not None, "Could not identify the AI player to target for elimination"
    print(f"  Identified AI player: {ai_player.name} ({ai_player.player_id}) with role {ai_player.role}")

    # Day 1: Nominate and vote out the AI player
    print("\n--- DAY 1: Voting Phase ---")
    
    # Nomination Phase
    for p in players:
        if p.player_id != ai_player.player_id: # Everyone votes for the AI
            p.send_action("SUBMIT_VOTE", {"vote_type": "NOMINATION", "target_id": ai_player.player_id})
    
    time.sleep(2) # Allow votes to be processed

    # Verdict Phase
    for p in players:
        p.send_action("SUBMIT_VOTE", {"vote_type": "VERDICT", "target_id": "GUILTY"})

    time.sleep(3) # Allow verdict to be processed

    # 4. Verify game end
    print("\nSTEP 4: Verifying game end...")
    for p in players:
        game_ended = p.get_latest_event("GAME_ENDED")
        assert game_ended is not None, f"{p.name} did not receive GAME_ENDED event"
        
        end_payload = game_ended.get("payload", {})
        assert end_payload.get("winning_faction") == "HUMANS", "Expected HUMAN faction to win"
        print(f"  ✅ {p.name} received correct GAME_ENDED event.")