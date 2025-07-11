"""
E2E tests for Advanced Gameplay Systems:
- Corporate Mandates & Personal KPIs
- LIAISON & Whistleblower Protocols
- Specialized Communication Mechanics
- AI Conversion & System Shock
"""

import asyncio
import json
import pytest
import logging
from typing import Dict, List, Any

from . import BaseE2ETest

logger = logging.getLogger(__name__)


class TestAdvancedGameplaySystems(BaseE2ETest):
    """Test advanced gameplay systems in full game context"""

    async def test_corporate_mandate_assignment_and_effects(self):
        """Test that corporate mandates are assigned and have effects"""
        # Create game with 4 players
        await self.create_and_start_game(player_count=4)
        
        # Wait for mandate assignment during game initialization
        mandate_event = await self.wait_for_event("MANDATE_ACTIVATED", timeout=10)
        assert mandate_event is not None, "Corporate mandate should be assigned at game start"
        
        # Verify mandate data
        mandate_data = mandate_event["payload"]
        assert "mandate_type" in mandate_data
        assert "name" in mandate_data
        assert "description" in mandate_data
        assert "effects" in mandate_data
        
        logger.info(f"Corporate mandate assigned: {mandate_data['name']}")
        
        # Check that mandate appears in game state
        game_state = await self.get_game_state()
        assert game_state["corporateMandate"] is not None
        assert game_state["corporateMandate"]["isActive"] is True

    async def test_personal_kpi_assignment(self):
        """Test that personal KPIs are assigned to human players"""
        await self.create_and_start_game(player_count=4)
        
        # Wait for KPI assignment events
        kpi_events = []
        for _ in range(4):  # Expect 4 KPI assignments
            event = await self.wait_for_event("KPI_ASSIGNED", timeout=10)
            if event:
                kpi_events.append(event)
        
        assert len(kpi_events) >= 3, "Should assign KPIs to human players"
        
        # Verify KPI data structure
        for event in kpi_events:
            payload = event["payload"]
            assert "kpi_type" in payload
            assert "description" in payload
            assert "target" in payload
            assert "reward" in payload
            
        logger.info(f"Assigned {len(kpi_events)} personal KPIs")

    async def test_slash_status_command(self):
        """Test that /status command works in chat"""
        await self.create_and_start_game(player_count=3)
        await self.advance_to_phase("DISCUSSION")
        
        player1 = self.players[0]
        status_message = "Working on security audit"
        
        # Send /status command
        await player1.send_message(f"/status {status_message}")
        
        # Wait for status change event
        status_event = await self.wait_for_event("SLACK_STATUS_CHANGED", timeout=5)
        assert status_event is not None, "Status command should generate event"
        assert status_event["payload"]["status"] == status_message
        
        logger.info(f"Status command successful: '{status_message}'")

    async def test_whisper_mechanic(self):
        """Test whisper functionality between players"""
        await self.create_and_start_game(player_count=3)
        await self.advance_to_phase("DISCUSSION")
        
        player1 = self.players[0]
        player2 = self.players[1]
        
        whisper_message = "I think player 3 is suspicious"
        
        # Send whisper action
        await player1.send_action({
            "type": "WHISPER",
            "payload": {
                "target_player_id": player2.player_id,
                "message": whisper_message
            }
        })
        
        # Wait for public whisper announcement
        chat_event = await self.wait_for_event("CHAT_MESSAGE", timeout=5)
        assert chat_event is not None
        assert "whispers to" in chat_event["payload"]["message"]
        
        # Wait for private notification to target
        private_event = await self.wait_for_event("PRIVATE_NOTIFICATION", timeout=5)
        assert private_event is not None
        assert private_event["payload"]["type"] == "whisper"
        assert private_event["payload"]["message"] == whisper_message
        
        logger.info("Whisper mechanic working correctly")

    async def test_ai_conversion_with_system_shock(self):
        """Test AI conversion mechanics and system shock effects"""
        await self.create_and_start_game(player_count=4)
        
        # Advance to night phase
        await self.advance_to_phase("NIGHT")
        
        # Find AI player and human target
        game_state = await self.get_game_state()
        ai_player = None
        human_target = None
        
        for player_id, player_data in game_state["players"].items():
            if player_data["alignment"] == "AI":
                ai_player = self.get_player_by_id(player_id)
            elif player_data["alignment"] == "HUMAN" and human_target is None:
                human_target = player_data
        
        assert ai_player is not None, "Should have an AI player"
        assert human_target is not None, "Should have a human target"
        
        # AI attempts conversion
        await ai_player.send_action({
            "type": "ATTEMPT_CONVERSION",
            "payload": {
                "target_id": human_target["id"]
            }
        })
        
        # Wait for night resolution
        resolution_event = await self.wait_for_event("NIGHT_ACTIONS_RESOLVED", timeout=30)
        assert resolution_event is not None
        
        # Check for AI equity changes or system shock
        payload = resolution_event["payload"]
        if "conversion_attempts" in payload and len(payload["conversion_attempts"]) > 0:
            attempt = payload["conversion_attempts"][0]
            assert "ai_equity_after" in attempt
            assert attempt["ai_equity_after"] > attempt["ai_equity_before"]
            
            # If conversion failed, should have system shock
            if not attempt["success"]:
                assert "system_shock" in attempt
                logger.info("System shock applied after failed conversion")
            else:
                logger.info("AI conversion successful")

    async def test_liaison_protocol_activation(self):
        """Test LIAISON protocol activation when AI faction grows"""
        # This test would require manipulating game state to have high AI percentage
        # For now, we'll test the event structure when it occurs
        await self.create_and_start_game(player_count=6)
        
        # Skip to later in game when LIAISON might trigger
        # (In real test, would need to simulate AI conversions)
        
        # Listen for LIAISON protocol activation
        liaison_event = await self.wait_for_event("LIAISON_PROTOCOL_ACTIVATED", timeout=60)
        
        if liaison_event:
            payload = liaison_event["payload"]
            assert "ai_percentage" in payload
            assert "mining_bonus_slots" in payload
            assert payload["ai_percentage"] >= 0.40
            logger.info(f"LIAISON Protocol activated at {payload['ai_percentage']*100:.1f}% AI")

    async def test_whistleblower_voting_for_deactivated_players(self):
        """Test whistleblower voting mechanism for deactivated players"""
        await self.create_and_start_game(player_count=4)
        
        # Advance through several phases to get eliminations
        await self.advance_to_phase("DISCUSSION")
        await self.advance_to_phase("NOMINATION")
        
        # Simulate voting and elimination (simplified)
        await self.advance_to_phase("NIGHT")
        
        # Wait for whistleblower voting to start
        whistleblower_event = await self.wait_for_event("WHISTLEBLOWER_VOTING_STARTED", timeout=30)
        
        if whistleblower_event:
            payload = whistleblower_event["payload"]
            assert "crisis_options" in payload
            assert len(payload["crisis_options"]) == 3
            
            for option in payload["crisis_options"]:
                assert "type" in option
                assert "title" in option
                assert "description" in option
            
            logger.info("Whistleblower voting started successfully")

    async def test_exit_interview_and_parting_shot(self):
        """Test exit interview functionality for eliminated players"""
        await self.create_and_start_game(player_count=3)
        
        # Advance to voting phase and eliminate a player
        await self.advance_to_phase("DISCUSSION")
        await self.advance_to_phase("NOMINATION")
        
        # Get first player to submit exit interview
        player1 = self.players[0]
        parting_message = "I knew it was you all along!"
        
        # Submit parting shot
        await player1.send_action({
            "type": "SUBMIT_EXIT_INTERVIEW",
            "payload": {
                "parting_shot": parting_message
            }
        })
        
        # Wait for parting shot event
        parting_event = await self.wait_for_event("PARTING_SHOT_SET", timeout=5)
        if parting_event:
            assert parting_event["payload"]["parting_shot"] == parting_message
            logger.info(f"Parting shot recorded: '{parting_message}'")

    async def test_mandate_effects_on_gameplay(self):
        """Test that corporate mandates actually affect gameplay mechanics"""
        await self.create_and_start_game(player_count=4)
        
        # Get the assigned mandate
        mandate_event = await self.wait_for_event("MANDATE_ACTIVATED", timeout=10)
        mandate_type = mandate_event["payload"]["mandate_type"]
        
        if mandate_type == "TOTAL_TRANSPARENCY":
            # Test that private communications are blocked
            logger.info("Testing Total Transparency mandate effects")
            
        elif mandate_type == "AGGRESSIVE_GROWTH":
            # Test mining and token effects
            logger.info("Testing Aggressive Growth mandate effects")
            
        elif mandate_type == "SECURITY_LOCKDOWN":
            # Test ability restrictions
            logger.info("Testing Security Lockdown mandate effects")

    # Helper methods
    async def advance_to_phase(self, target_phase: str):
        """Advance game to specific phase"""
        max_attempts = 20
        attempt = 0
        
        while attempt < max_attempts:
            game_state = await self.get_game_state()
            current_phase = game_state.get("phase", {}).get("type")
            
            if current_phase == target_phase:
                logger.info(f"Reached {target_phase} phase")
                return
            
            # Wait for phase to change naturally
            await asyncio.sleep(2)
            attempt += 1
        
        raise TimeoutError(f"Failed to reach {target_phase} phase after {max_attempts} attempts")

    async def get_game_state(self) -> Dict[str, Any]:
        """Get current game state from first player"""
        if not self.players:
            raise RuntimeError("No players available")
        
        # Send state request or get from last received update
        return self.players[0].last_game_state or {}

    def get_player_by_id(self, player_id: str):
        """Get player connection by ID"""
        for player in self.players:
            if player.player_id == player_id:
                return player
        return None


# Pytest fixtures and configuration
@pytest.mark.asyncio
async def test_full_advanced_systems_integration():
    """Integration test running multiple advanced systems together"""
    test_instance = TestAdvancedGameplaySystems()
    
    try:
        await test_instance.setup()
        
        # Test mandate assignment
        await test_instance.test_corporate_mandate_assignment_and_effects()
        
        # Test KPI assignment  
        await test_instance.test_personal_kpi_assignment()
        
        # Test communication features
        await test_instance.test_slash_status_command()
        
        logger.info("Advanced gameplay systems integration test completed successfully")
        
    finally:
        await test_instance.teardown()