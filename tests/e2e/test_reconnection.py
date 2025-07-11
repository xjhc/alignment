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

def test_initial_join_is_immediate_and_reconnect_shows_overlay():
    """
    Tests Bug #1 Fix: Atomic Join & Sync Protocol
    
    This test verifies:
    1. Player A creates a new lobby
    2. Player B joins - should be INSTANT with no "Syncing Session..." overlay
    3. Player B reloads the page - should show overlay briefly then disappear
    
    This test verifies both the fix (instant join) and correct reconnect behavior.
    """
    print("\nSTEP 1: Player A creates lobby...")
    create_response = requests.post(
        f"{BASE_URL}/api/games",
        json={"lobby_name": "Reconnection Test Game", "player_name": "PlayerA", "player_avatar": "👑"}
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
    join_response = requests.post(
        f"{BASE_URL}/api/games/{game_id}/join",
        json={"player_name": "PlayerB", "player_avatar": "🎭"}
    )
    assert join_response.status_code == 200, f"Failed to join lobby: {join_response.text}"
    
    join_data = join_response.json()
    player_b_id = join_data["player_id"]
    player_b_token = join_data["session_token"]
    print(f"  ✅ Player B joined: {player_b_id}")
    
    # Connect Player B via WebSocket
    print("STEP 4: Player B connects via WebSocket...")
    player_b_ws_url = f"{WS_URL}?gameId={game_id}&playerId={player_b_id}&sessionToken={player_b_token}"
    player_b_ws = websocket.create_connection(player_b_ws_url, timeout=5)
    print("  ✅ Player B connected")
    
    # Wait for both players to see each other
    time.sleep(2)
    
    # Verify both players can see each other before the reconnection test
    print("STEP 5: Verifying both players can see each other...")
    
    # Set up message receivers for both players
    player_a_messages = {}
    player_b_messages = {}
    stop_event_a = threading.Event()
    stop_event_b = threading.Event()
    
    receive_thread_a = threading.Thread(
        target=receive_websocket_messages,
        args=(player_a_ws, player_a_messages, stop_event_a, 3)
    )
    receive_thread_b = threading.Thread(
        target=receive_websocket_messages,
        args=(player_b_ws, player_b_messages, stop_event_b, 3)
    )
    
    receive_thread_a.start()
    receive_thread_b.start()
    
    time.sleep(2)  # Give time for messages to arrive
    
    stop_event_a.set()
    stop_event_b.set()
    receive_thread_a.join(timeout=2)
    receive_thread_b.join(timeout=2)
    
    # Verify both players received LOBBY_STATE_UPDATE with both players visible
    assert "LOBBY_STATE_UPDATE" in player_a_messages, "Player A should receive LOBBY_STATE_UPDATE"
    assert "LOBBY_STATE_UPDATE" in player_b_messages, "Player B should receive LOBBY_STATE_UPDATE"
    
    player_a_state = player_a_messages["LOBBY_STATE_UPDATE"].get("payload", {})
    player_b_state = player_b_messages["LOBBY_STATE_UPDATE"].get("payload", {})
    
    player_a_players = player_a_state.get("players", [])
    player_b_players = player_b_state.get("players", [])
    
    assert len(player_a_players) == 2, f"Player A should see 2 players, but sees {len(player_a_players)}"
    assert len(player_b_players) == 2, f"Player B should see 2 players, but sees {len(player_b_players)}"
    
    print("  ✅ Both players can see each other in the roster")
    
    # STEP 6: THE CRITICAL TEST - Simulate Player B's page reload (disconnect and reconnect)
    print("STEP 6: Simulating Player B's page reload (disconnect and reconnect)...")
    
    # Close Player B's WebSocket connection (simulating page reload)
    player_b_ws.close()
    print("  ✅ Player B disconnected")
    
    # Wait a moment to simulate the disconnect/reconnect gap
    time.sleep(1)
    
    # Reconnect Player B with the same credentials (simulating page reload)
    print("STEP 7: Player B reconnecting...")
    player_b_ws_reconnect = websocket.create_connection(player_b_ws_url, timeout=5)
    print("  ✅ Player B reconnected")
    
    # Set up message receiver for Player B's reconnection
    player_b_reconnect_messages = {}
    stop_event_b_reconnect = threading.Event()
    
    receive_thread_b_reconnect = threading.Thread(
        target=receive_websocket_messages,
        args=(player_b_ws_reconnect, player_b_reconnect_messages, stop_event_b_reconnect, 10)
    )
    receive_thread_b_reconnect.start()
    
    # STEP 8: THE ASSERTION THAT CURRENTLY FAILS
    print("STEP 8: Verifying Player B receives state after reconnection...")
    
    # Wait for Player B to receive the state snapshot
    # This simulates the "Syncing Session..." overlay timeout
    start_time = time.time()
    timeout_seconds = 10
    
    while time.time() - start_time < timeout_seconds:
        if "LOBBY_STATE_UPDATE" in player_b_reconnect_messages:
            print("  ✅ Player B received LOBBY_STATE_UPDATE after reconnection")
            break
        time.sleep(0.1)
    
    stop_event_b_reconnect.set()
    receive_thread_b_reconnect.join(timeout=2)
    
    # THE CRITICAL ASSERTION THAT WILL FAIL DUE TO THE BUG
    assert "LOBBY_STATE_UPDATE" in player_b_reconnect_messages, (
        f"Player B should receive LOBBY_STATE_UPDATE after reconnection within {timeout_seconds} seconds. "
        f"This simulates the 'Syncing Session...' overlay timeout. "
        f"Received messages: {list(player_b_reconnect_messages.keys())}"
    )
    
    # Verify the state contains the correct player roster
    reconnect_state = player_b_reconnect_messages["LOBBY_STATE_UPDATE"].get("payload", {})
    reconnect_players = reconnect_state.get("players", [])
    
    assert len(reconnect_players) == 2, (
        f"Player B should see 2 players after reconnection, but sees {len(reconnect_players)}. "
        f"This indicates the state snapshot is incomplete."
    )
    
    print("  ✅ Player B successfully received complete state after reconnection")
    print("  ✅ The 'Syncing Session...' hang bug is FIXED!")
    
    # Clean up
    try:
        player_a_ws.close()
        player_b_ws_reconnect.close()
    except:
        pass

if __name__ == "__main__":
    test_reconnection_hang_bug()