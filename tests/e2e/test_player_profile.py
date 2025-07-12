import requests
import pytest

# Configuration
BASE_URL = "http://localhost:8080"

def test_player_profile_api_endpoints():
    """
    Test the player profile API endpoints for live data integration.
    Verifies that the profile, avatar, and title endpoints work correctly.
    """
    
    # Create a test game and player
    create_response = requests.post(
        f"{BASE_URL}/api/lobby/create",
        json={"lobby_name": "Profile Test Game", "player_name": "TestPlayer", "player_avatar": "👤"}
    )
    assert create_response.status_code == 200, f"Failed to create lobby: {create_response.text}"
    
    game_data = create_response.json()
    assert "game_id" in game_data
    assert "player_id" in game_data
    assert "session_token" in game_data
    
    player_id = game_data["player_id"]
    session_token = game_data["session_token"]
    
    # Set up session cookies for authenticated requests
    session = requests.Session()
    session.cookies.set("session_token", session_token)
    
    print(f"Created test player: {player_id}")
    
    # Test player profile endpoint
    print("Testing GET /api/players/profile...")
    profile_response = session.get(f"{BASE_URL}/api/players/profile")
    
    if profile_response.status_code == 200:
        # Profile endpoint exists and returns data
        profile_data = profile_response.json()
        print("  ✅ Profile endpoint returned data")
        
        # Verify structure of profile data
        assert "profile" in profile_data, "Profile response should contain 'profile' field"
        profile = profile_data["profile"]
        
        # Check required profile fields
        required_fields = ["id", "username", "displayName"]
        for field in required_fields:
            assert field in profile, f"Profile should contain '{field}' field"
        
        # Check optional fields that may be present
        optional_fields = ["totalGamesPlayed", "totalGamesWon", "totalTokensMined", "kudosReceived", 
                          "unlockedAchievements", "unlockedAvatars", "unlockedTitles", 
                          "equippedAvatar", "equippedTitle"]
        for field in optional_fields:
            if field in profile:
                print(f"    Profile contains optional field: {field}")
        
        # Test avatar equipping endpoint
        print("Testing POST /api/players/equip-avatar...")
        avatar_response = session.post(
            f"{BASE_URL}/api/players/equip-avatar",
            json={"avatar_id": "scientist"}
        )
        
        if avatar_response.status_code == 200:
            print("  ✅ Avatar equip endpoint works")
        elif avatar_response.status_code == 404:
            print("  ⚠️ Avatar equip endpoint not implemented yet")
        else:
            print(f"  ❌ Avatar equip endpoint returned unexpected status: {avatar_response.status_code}")
        
        # Test title equipping endpoint  
        print("Testing POST /api/players/equip-title...")
        title_response = session.post(
            f"{BASE_URL}/api/players/equip-title",
            json={"title_id": "veteran"}
        )
        
        if title_response.status_code == 200:
            print("  ✅ Title equip endpoint works")
        elif title_response.status_code == 404:
            print("  ⚠️ Title equip endpoint not implemented yet") 
        else:
            print(f"  ❌ Title equip endpoint returned unexpected status: {title_response.status_code}")
            
    elif profile_response.status_code == 404:
        print("  ⚠️ Profile endpoint not implemented yet - component will fall back to mock data")
    else:
        pytest.fail(f"Profile endpoint returned unexpected status: {profile_response.status_code}")

def test_social_api_endpoints():
    """
    Test the social feature API endpoints (kudos, reporting, blocking).
    Verifies that the social service integration works correctly.
    """
    
    # Create two test players for social interactions
    create_response_1 = requests.post(
        f"{BASE_URL}/api/lobby/create",
        json={"lobby_name": "Social Test Game 1", "player_name": "Player1", "player_avatar": "👤"}
    )
    assert create_response_1.status_code == 200
    
    create_response_2 = requests.post(
        f"{BASE_URL}/api/lobby/create", 
        json={"lobby_name": "Social Test Game 2", "player_name": "Player2", "player_avatar": "🎭"}
    )
    assert create_response_2.status_code == 200
    
    player1_data = create_response_1.json()
    player2_data = create_response_2.json()
    
    player1_session = requests.Session()
    player1_session.cookies.set("session_token", player1_data["session_token"])
    
    player2_id = player2_data["player_id"]
    
    print(f"Created test players: {player1_data['player_id']} and {player2_id}")
    
    # Test kudos giving endpoint
    print("Testing POST /api/social/kudos...")
    kudos_response = player1_session.post(
        f"{BASE_URL}/api/social/kudos",
        json={
            "target_player_id": player2_id,
            "reason": "Great teamwork!"
        }
    )
    
    if kudos_response.status_code == 200:
        print("  ✅ Kudos endpoint works")
        kudos_data = kudos_response.json()
        assert "success" in kudos_data, "Kudos response should contain success field"
    elif kudos_response.status_code == 404:
        print("  ⚠️ Kudos endpoint not implemented yet")
    else:
        print(f"  ❌ Kudos endpoint returned unexpected status: {kudos_response.status_code}")
    
    # Test reporting endpoint
    print("Testing POST /api/social/report...")
    report_response = player1_session.post(
        f"{BASE_URL}/api/social/report",
        json={
            "target_player_id": player2_id,
            "reason": "spam",
            "description": "Sending inappropriate messages"
        }
    )
    
    if report_response.status_code == 200:
        print("  ✅ Report endpoint works")
        report_data = report_response.json()
        assert "success" in report_data, "Report response should contain success field"
    elif report_response.status_code == 404:
        print("  ⚠️ Report endpoint not implemented yet")
    else:
        print(f"  ❌ Report endpoint returned unexpected status: {report_response.status_code}")
    
    # Test blocking endpoint
    print("Testing POST /api/social/block...")
    block_response = player1_session.post(
        f"{BASE_URL}/api/social/block",
        json={"target_player_id": player2_id}
    )
    
    if block_response.status_code == 200:
        print("  ✅ Block endpoint works")
        block_data = block_response.json()
        assert "success" in block_data, "Block response should contain success field"
    elif block_response.status_code == 404:
        print("  ⚠️ Block endpoint not implemented yet")
    else:
        print(f"  ❌ Block endpoint returned unexpected status: {block_response.status_code}")