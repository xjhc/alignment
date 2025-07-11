import requests
import json
import websocket
import time
import pytest
import threading

# Configuration
BASE_URL = "http://localhost:8080"
WS_URL = "ws://localhost:8080/ws"

def receive_websocket_messages(ws, received_messages, stop_event, timeout=10):
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

def test_player_disconnect_corruption():
    """
    Tests Bug #2: Game State Corruption on Player Disconnect.
    
    This test reproduces the scenario where:
    1. Three players (A, B, C) join the same lobby
    2. Player C abruptly disconnects (connection closed)
    3. Players A and B should see Player C's status update to "Disconnected"
    4. The UI should not show a "ghost" of the disconnected player
    
    Currently, this test WILL FAIL due to the bug where the server
    fails to broadcast clear, authoritative state changes on disconnect.
    """
    print("\nSTEP 1: Player A creates lobby...")
    create_response = requests.post(
        f"{BASE_URL}/api/games",
        json={"lobby_name": "Disconnection Test Game", "player_name": "PlayerA", "player_avatar": "👑"}
    )
    assert create_response.status_code == 201, f"Failed to create lobby: {create_response.text}"
    
    create_data = create_response.json()
    game_id = create_data["game_id"]
    player_a_id = create_data["player_id"]
    player_a_token = create_data["session_token"]
    print(f"  ✅ Lobby created: {game_id}")
    
    # Connect Player A via WebSocket
    print("STEP 2: Player A connects via WebSocket...")
    player_a_ws_url = f"{WS_URL}?gameId={game_id}&playerId={player_a_id}&sessionToken={player_a_token}"
    player_a_ws = websocket.create_connection(player_a_ws_url, timeout=5)
    print("  ✅ Player A connected")
    
    # Wait for Player A to receive initial messages
    time.sleep(1)
    
    print("STEP 3: Player B joins the lobby...")
    join_response_b = requests.post(
        f"{BASE_URL}/api/games/{game_id}/join",
        json={"player_name": "PlayerB", "player_avatar": "🎭"}
    )
    assert join_response_b.status_code == 200, f"Failed to join lobby: {join_response_b.text}"
    
    join_data_b = join_response_b.json()
    player_b_id = join_data_b["player_id"]
    player_b_token = join_data_b["session_token"]
    print(f"  ✅ Player B joined: {player_b_id}")
    
    # Connect Player B via WebSocket
    player_b_ws_url = f"{WS_URL}?gameId={game_id}&playerId={player_b_id}&sessionToken={player_b_token}"
    player_b_ws = websocket.create_connection(player_b_ws_url, timeout=5)
    print("  ✅ Player B connected")
    
    print("STEP 4: Player C joins the lobby...")
    join_response_c = requests.post(
        f"{BASE_URL}/api/games/{game_id}/join",
        json={"player_name": "PlayerC", "player_avatar": "🎨"}
    )
    assert join_response_c.status_code == 200, f"Failed to join lobby: {join_response_c.text}"
    
    join_data_c = join_response_c.json()
    player_c_id = join_data_c["player_id"]
    player_c_token = join_data_c["session_token"]
    print(f"  ✅ Player C joined: {player_c_id}")
    
    # Connect Player C via WebSocket
    player_c_ws_url = f"{WS_URL}?gameId={game_id}&playerId={player_c_id}&sessionToken={player_c_token}"
    player_c_ws = websocket.create_connection(player_c_ws_url, timeout=5)
    print("  ✅ Player C connected")
    
    # Wait for all players to see each other
    time.sleep(2)
    
    # Verify all players can see each other before the disconnect test
    print("STEP 5: Verifying all players can see each other...")
    
    # Set up message receivers for all players
    player_a_messages = {}
    player_b_messages = {}
    player_c_messages = {}
    
    stop_event_a = threading.Event()
    stop_event_b = threading.Event()
    stop_event_c = threading.Event()
    
    receive_thread_a = threading.Thread(
        target=receive_websocket_messages,
        args=(player_a_ws, player_a_messages, stop_event_a, 3)
    )
    receive_thread_b = threading.Thread(
        target=receive_websocket_messages,
        args=(player_b_ws, player_b_messages, stop_event_b, 3)
    )
    receive_thread_c = threading.Thread(
        target=receive_websocket_messages,
        args=(player_c_ws, player_c_messages, stop_event_c, 3)
    )
    
    receive_thread_a.start()
    receive_thread_b.start()
    receive_thread_c.start()
    
    time.sleep(2)  # Give time for messages to arrive
    
    stop_event_a.set()
    stop_event_b.set()
    stop_event_c.set()
    
    receive_thread_a.join(timeout=2)
    receive_thread_b.join(timeout=2)
    receive_thread_c.join(timeout=2)
    
    # Verify all players received LOBBY_STATE_UPDATE with all players visible
    for player_name, messages in [("A", player_a_messages), ("B", player_b_messages), ("C", player_c_messages)]:
        assert "LOBBY_STATE_UPDATE" in messages, f"Player {player_name} should receive LOBBY_STATE_UPDATE"
        state = messages["LOBBY_STATE_UPDATE"].get("payload", {})
        players = state.get("players", [])
        assert len(players) == 3, f"Player {player_name} should see 3 players, but sees {len(players)}"
    
    print("  ✅ All players can see each other in the roster")
    
    # STEP 6: THE CRITICAL TEST - Simulate Player C's abrupt disconnect
    print("STEP 6: Simulating Player C's abrupt disconnect...")
    
    # Close Player C's WebSocket connection abruptly (simulating crash/network issue)
    player_c_ws.close()
    print("  ✅ Player C disconnected abruptly")
    
    # STEP 7: Check if Players A and B receive disconnect notifications
    print("STEP 7: Checking if Players A and B receive disconnect notifications...")
    
    # Set up message receivers for Players A and B to capture disconnect events
    player_a_disconnect_messages = {}
    player_b_disconnect_messages = {}
    
    stop_event_a_disconnect = threading.Event()
    stop_event_b_disconnect = threading.Event()
    
    receive_thread_a_disconnect = threading.Thread(
        target=receive_websocket_messages,
        args=(player_a_ws, player_a_disconnect_messages, stop_event_a_disconnect, 8)
    )
    receive_thread_b_disconnect = threading.Thread(
        target=receive_websocket_messages,
        args=(player_b_ws, player_b_disconnect_messages, stop_event_b_disconnect, 8)
    )
    
    receive_thread_a_disconnect.start()
    receive_thread_b_disconnect.start()
    
    # Wait for disconnect notification events
    time.sleep(5)  # Give time for server to process disconnect and broadcast updates
    
    stop_event_a_disconnect.set()
    stop_event_b_disconnect.set()
    
    receive_thread_a_disconnect.join(timeout=2)
    receive_thread_b_disconnect.join(timeout=2)
    
    # THE CRITICAL ASSERTIONS THAT WILL FAIL DUE TO THE BUG
    print("STEP 8: Verifying Players A and B receive connection status updates...")
    
    # Check if Players A and B received any connection status change events
    # This could be PLAYER_CONNECTION_STATUS_CHANGED or LOBBY_STATE_UPDATE with updated status
    
    player_a_received_disconnect_event = False
    player_b_received_disconnect_event = False
    
    # Check for specific disconnect events
    for event_type in player_a_disconnect_messages:
        if "CONNECTION_STATUS" in event_type or "PLAYER_DISCONNECTED" in event_type:
            player_a_received_disconnect_event = True
            print(f"  ✅ Player A received disconnect event: {event_type}")
            break
        elif event_type == "LOBBY_STATE_UPDATE":
            # Check if the lobby state shows Player C as disconnected
            state = player_a_disconnect_messages[event_type].get("payload", {})
            players = state.get("players", [])
            for player in players:
                if isinstance(player, dict) and player.get("id") == player_c_id:
                    if player.get("connection_status") == "DISCONNECTED":
                        player_a_received_disconnect_event = True
                        print(f"  ✅ Player A received LOBBY_STATE_UPDATE with Player C marked as DISCONNECTED")
                        break
    
    for event_type in player_b_disconnect_messages:
        if "CONNECTION_STATUS" in event_type or "PLAYER_DISCONNECTED" in event_type:
            player_b_received_disconnect_event = True
            print(f"  ✅ Player B received disconnect event: {event_type}")
            break
        elif event_type == "LOBBY_STATE_UPDATE":
            # Check if the lobby state shows Player C as disconnected
            state = player_b_disconnect_messages[event_type].get("payload", {})
            players = state.get("players", [])
            for player in players:
                if isinstance(player, dict) and player.get("id") == player_c_id:
                    if player.get("connection_status") == "DISCONNECTED":
                        player_b_received_disconnect_event = True
                        print(f"  ✅ Player B received LOBBY_STATE_UPDATE with Player C marked as DISCONNECTED")
                        break
    
    # THE ASSERTION THAT WILL FAIL DUE TO THE BUG
    assert player_a_received_disconnect_event, (
        f"Player A should receive a disconnect notification for Player C. "
        f"Without this, the UI shows a 'ghost' player. "
        f"Player A received: {list(player_a_disconnect_messages.keys())}"
    )
    
    assert player_b_received_disconnect_event, (
        f"Player B should receive a disconnect notification for Player C. "
        f"Without this, the UI shows a 'ghost' player. "
        f"Player B received: {list(player_b_disconnect_messages.keys())}"
    )
    
    print("  ✅ Both Players A and B received proper disconnect notifications")
    print("  ✅ The game state corruption bug is FIXED!")
    
    # Clean up
    try:
        player_a_ws.close()
        player_b_ws.close()
    except:
        pass

if __name__ == "__main__":
    test_player_disconnect_corruption()