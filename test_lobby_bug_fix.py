#!/usr/bin/env python3
"""
Quick test to verify the lobby creation bug is fixed.
This tests the specific scenario that was broken: host creation and connection.
"""

import requests
import json
import websocket
import time
import threading

BASE_URL = "http://localhost:8080"
WS_URL = "ws://localhost:8080/ws"

def test_lobby_creation_bug_fix():
    """
    Test that demonstrates the lobby creation bug is fixed:
    1. Create lobby via REST API
    2. Connect as host via WebSocket using returned credentials
    3. Verify host is properly recognized and lobby becomes active
    """
    print("Testing lobby creation bug fix...")
    
    # Step 1: Create lobby
    print("Step 1: Creating lobby via REST API...")
    create_response = requests.post(
        f"{BASE_URL}/api/games",
        json={"lobby_name": "Bug Fix Test", "player_name": "TestHost", "player_avatar": "👑"}
    )
    
    if create_response.status_code != 201:
        print(f"❌ Failed to create lobby: {create_response.status_code} - {create_response.text}")
        return False
    
    create_data = create_response.json()
    game_id = create_data["game_id"]
    player_id = create_data["player_id"]
    session_token = create_data["session_token"]
    
    print(f"✅ Lobby created successfully:")
    print(f"   Game ID: {game_id}")
    print(f"   Player ID: {player_id}")
    print(f"   Session Token: {session_token[:10]}...")
    
    # Step 2: Connect via WebSocket as host
    print("Step 2: Connecting as host via WebSocket...")
    ws_url = f"{WS_URL}?gameId={game_id}&playerId={player_id}&sessionToken={session_token}"
    
    try:
        ws = websocket.create_connection(ws_url, timeout=10)
        print("✅ WebSocket connection established")
        
        # Step 3: Receive initial messages
        print("Step 3: Receiving initial messages...")
        messages_received = []
        
        # Collect messages for a few seconds
        start_time = time.time()
        while time.time() - start_time < 3:
            try:
                ws.settimeout(1)
                message = ws.recv()
                if message:
                    try:
                        parsed_msg = json.loads(message)
                        messages_received.append(parsed_msg)
                        print(f"   Received: {parsed_msg.get('type', 'unknown')}")
                    except json.JSONDecodeError:
                        print(f"   Received non-JSON: {message}")
            except websocket.WebSocketTimeoutException:
                continue
            except Exception as e:
                print(f"   Error receiving message: {e}")
                break
        
        # Step 4: Verify we received the expected messages
        print("Step 4: Verifying messages...")
        
        # Check for CLIENT_IDENTIFIED
        client_identified = next((msg for msg in messages_received if msg.get('type') == 'CLIENT_IDENTIFIED'), None)
        if client_identified:
            print("✅ Received CLIENT_IDENTIFIED")
            payload = client_identified.get('payload', {})
            if payload.get('your_player_id') == player_id:
                print("✅ Player ID matches in CLIENT_IDENTIFIED")
            else:
                print(f"❌ Player ID mismatch: expected {player_id}, got {payload.get('your_player_id')}")
                return False
        else:
            print("❌ Did not receive CLIENT_IDENTIFIED")
            return False
        
        # Check for LOBBY_STATE_UPDATE
        lobby_update = next((msg for msg in messages_received if msg.get('type') == 'LOBBY_STATE_UPDATE'), None)
        if lobby_update:
            print("✅ Received LOBBY_STATE_UPDATE")
            payload = lobby_update.get('payload', {})
            players = payload.get('players', [])
            
            if len(players) >= 1:
                print(f"✅ Lobby shows {len(players)} player(s)")
                
                # Check if host is in the lobby
                host_found = False
                for player in players:
                    if isinstance(player, dict) and player.get('id') == player_id:
                        host_found = True
                        print(f"✅ Host found in lobby: {player.get('name', 'Unknown')}")
                        break
                    elif player == player_id:  # Legacy format
                        host_found = True
                        print(f"✅ Host found in lobby (legacy format): {player_id}")
                        break
                
                if not host_found:
                    print(f"❌ Host not found in lobby. Players: {players}")
                    return False
                    
            else:
                print(f"❌ Lobby appears empty: {players}")
                return False
        else:
            print("❌ Did not receive LOBBY_STATE_UPDATE")
            return False
        
        print("✅ All checks passed! Lobby creation bug appears to be fixed.")
        ws.close()
        return True
        
    except Exception as e:
        print(f"❌ WebSocket connection failed: {e}")
        return False

if __name__ == "__main__":
    success = test_lobby_creation_bug_fix()
    if success:
        print("\n🎉 Lobby creation bug fix verification: PASSED")
        exit(0)
    else:
        print("\n💥 Lobby creation bug fix verification: FAILED") 
        exit(1)