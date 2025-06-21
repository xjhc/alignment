import requests
import json
import websocket
import time
import pytest
import threading

# Configuration
BASE_URL = "http://localhost:8080"
WS_URL = "ws://localhost:8080/ws"

def receive_websocket_messages(ws, received_messages, stop_event, timeout=5):
    """
    Helper function to receive WebSocket messages in a separate thread.
    Handles both individual messages and potentially batched messages.
    """
    start_time = time.time()
    try:
        while not stop_event.is_set() and time.time() - start_time < timeout:
            try:
                # Set a short timeout to allow checking stop_event
                ws.settimeout(0.5)
                message_str = ws.recv()
                
                if not message_str:
                    continue
                
                # Try to parse as a single JSON message first
                try:
                    message = json.loads(message_str)
                    msg_type = message.get("type")
                    if msg_type:
                        received_messages[msg_type] = message
                        print(f"  Received message of type: {msg_type}")
                except json.JSONDecodeError:
                    # If that fails, try to split by newlines (batched messages)
                    for line in message_str.strip().split('\n'):
                        if not line:
                            continue
                        try:
                            message = json.loads(line)
                            msg_type = message.get("type")
                            if msg_type:
                                received_messages[msg_type] = message
                                print(f"  Received message of type: {msg_type}")
                        except json.JSONDecodeError:
                            print(f"  Failed to parse message line: {line}")
                            
            except websocket.WebSocketTimeoutException:
                continue  # Check stop_event and continue
            except Exception as e:
                print(f"  Error receiving message: {e}")
                break
    finally:
        ws.settimeout(None)  # Reset timeout

def test_full_lobby_flow():
    """
    Tests the complete lobby flow:
    1. Create a game via REST API.
    2. List the new lobby.
    3. Join the lobby via WebSocket and receive initial state.
    """
    # 1. Create a lobby for the host
    print("\nSTEP 1: Creating lobby...")
    create_response = requests.post(
        f"{BASE_URL}/api/games",
        json={"lobby_name": "E2E Test Game", "player_name": "HostPlayer", "player_avatar": "👤"}
    )
    assert create_response.status_code == 201, f"Failed to create lobby: {create_response.text}"

    create_data = create_response.json()
    assert "game_id" in create_data
    assert "player_id" in create_data
    assert "session_token" in create_data

    game_id = create_data["game_id"]
    host_player_id = create_data["player_id"]
    host_session_token = create_data["session_token"]
    print(f"  ✅ Lobby created: {game_id}")

    # 2. List lobbies and verify the new one exists
    print("STEP 2: Listing lobbies...")
    list_response = requests.get(f"{BASE_URL}/api/games")
    assert list_response.status_code == 200

    lobbies = list_response.json().get("lobbies", [])
    assert any(lobby['id'] == game_id for lobby in lobbies), "Newly created lobby not found in list"
    print("  ✅ Lobby found in list.")

    # 3. Join via WebSocket and verify initial state for the host
    print("STEP 3: Host joining via WebSocket...")
    ws_url = f"{WS_URL}?gameId={game_id}&playerId={host_player_id}&sessionToken={host_session_token}"

    try:
        ws = websocket.create_connection(ws_url, timeout=5)
        print("  ✅ WebSocket connection successful.")

        # --- ROBUST TEST LOGIC ---
        # Use a separate thread to receive messages to avoid blocking
        received_messages = {}
        stop_event = threading.Event()
        
        receive_thread = threading.Thread(
            target=receive_websocket_messages,
            args=(ws, received_messages, stop_event, 5)
        )
        receive_thread.start()
        
        # Wait for expected messages or timeout
        start_time = time.time()
        messages_to_receive = 2  # CLIENT_IDENTIFIED and LOBBY_STATE_UPDATE
        
        while len(received_messages) < messages_to_receive and time.time() - start_time < 5:
            time.sleep(0.1)  # Small delay to avoid busy waiting
        
        stop_event.set()
        receive_thread.join(timeout=1)

        # Now, assert the presence and content of each expected message.
        assert "CLIENT_IDENTIFIED" in received_messages, f"Did not receive CLIENT_IDENTIFIED event. Got: {list(received_messages.keys())}"
        assert "LOBBY_STATE_UPDATE" in received_messages, f"Did not receive LOBBY_STATE_UPDATE event. Got: {list(received_messages.keys())}"

        # Validate CLIENT_IDENTIFIED payload
        id_payload = received_messages["CLIENT_IDENTIFIED"].get("payload", {})
        assert id_payload.get("your_player_id") == host_player_id, f"Expected player ID {host_player_id}, got {id_payload.get('your_player_id')}"
        print(f"  ✅ Verified host player ID: {host_player_id}")

        # Validate LOBBY_STATE_UPDATE payload
        state_payload = received_messages["LOBBY_STATE_UPDATE"].get("payload", {})
        assert state_payload.get("lobby_id") == game_id
        assert state_payload.get("host_id") == host_player_id
        players = state_payload.get("players", [])
        assert len(players) == 1, f"Expected 1 player in the lobby, got {len(players)}"
        
        # Players should now be objects with id, name, avatar
        if isinstance(players[0], dict):
            assert players[0].get('id') == host_player_id, f"Player ID in lobby state doesn't match identified ID: {players[0]}"
            assert players[0].get('name') == "HostPlayer", f"Expected host name 'HostPlayer', got {players[0].get('name')}"
            print(f"  ✅ Host player correctly present: {players[0]}")
        else:
            # Fallback for old format where players is just a list of IDs
            assert players[0] == host_player_id, f"Player ID in lobby state doesn't match identified ID: {players[0]}"
        
        print("  ✅ Verified correct LOBBY_STATE_UPDATE with host present from start.")

    except Exception as e:
        pytest.fail(f"WebSocket test failed: {e}")
    finally:
        if 'ws' in locals() and ws.connected:
            ws.close()

def test_player_visibility_race_condition():
    """
    Tests the specific race condition where Player 4 can't see existing players.
    This creates a lobby with 3 players first, then adds a 4th to test visibility.
    """
    print("\nSTEP 1: Creating lobby with host...")
    create_response = requests.post(
        f"{BASE_URL}/api/games",
        json={"lobby_name": "Race Condition Test", "player_name": "Host", "player_avatar": "👑"}
    )
    assert create_response.status_code == 201
    
    create_data = create_response.json()
    game_id = create_data["game_id"]
    host_session_token = create_data["session_token"]
    print(f"  ✅ Lobby created: {game_id}")

    # Simulate 3 additional players joining quickly (to set up the race condition scenario)
    print("STEP 2: Adding 3 more players to create the visibility issue scenario...")
    
    player_tokens = []
    for i in range(2, 4):  # Players 2 and 3
        join_response = requests.post(
            f"{BASE_URL}/api/games/{game_id}/join",
            json={"player_name": f"Player{i}", "player_avatar": f"🎭"}
        )
        if join_response.status_code == 200:
            join_data = join_response.json()
            player_tokens.append((join_data["player_id"], join_data["session_token"]))
            print(f"  ✅ Player {i} joined")

    # Now test the 4th player (the one that typically can't see others)
    print("STEP 3: Testing 4th player visibility...")
    
    join_response = requests.post(
        f"{BASE_URL}/api/games/{game_id}/join", 
        json={"player_name": "Player4", "player_avatar": "🔍"}
    )
    assert join_response.status_code == 200
    
    player4_data = join_response.json()
    player4_id = player4_data["player_id"]
    player4_session_token = player4_data["session_token"]
    
    # Connect Player 4 via WebSocket and check if they can see all players
    ws_url = f"{WS_URL}?gameId={game_id}&playerId={player4_id}&sessionToken={player4_session_token}"
    
    try:
        ws = websocket.create_connection(ws_url, timeout=5)
        print("  ✅ Player 4 WebSocket connected.")

        received_messages = {}
        stop_event = threading.Event()
        
        receive_thread = threading.Thread(
            target=receive_websocket_messages,
            args=(ws, received_messages, stop_event, 5)
        )
        receive_thread.start()
        
        # Wait for messages
        time.sleep(2)  # Give time for messages to arrive
        stop_event.set()
        receive_thread.join(timeout=1)

        # Check that Player 4 receives LOBBY_STATE_UPDATE with all players visible
        assert "LOBBY_STATE_UPDATE" in received_messages, "Player 4 should receive LOBBY_STATE_UPDATE"
        
        state_payload = received_messages["LOBBY_STATE_UPDATE"].get("payload", {})
        players = state_payload.get("players", [])
        
        # The key test: Player 4 should see at least 4 players (including themselves)
        expected_players = 4  # Host + Player2 + Player3 + Player4
        assert len(players) >= expected_players, f"Player 4 should see {expected_players} players, but only sees {len(players)}: {players}"
        
        print(f"  ✅ Player 4 can see all {len(players)} players - race condition fixed!")

    except Exception as e:
        pytest.fail(f"Player 4 visibility test failed: {e}")
    finally:
        if 'ws' in locals() and ws.connected:
            ws.close()