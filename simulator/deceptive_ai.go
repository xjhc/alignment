package simulator

import (
	"math/rand"

	"github.com/xjhc/alignment/core"
)

// DeceptiveAI persona - mimics human behavior during day, targets influential players at night
type DeceptiveAI struct {
	rand *rand.Rand
}

func NewDeceptiveAI(seed int64) *DeceptiveAI {
	return &DeceptiveAI{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (d *DeceptiveAI) GetID() string {
	return "deceptive_ai"
}

func (d *DeceptiveAI) GetDescription() string {
	return "Mimics human behavior (mines for others, participates in discussions). At night, targets influential players for conversion."
}

func (d *DeceptiveAI) DecideAction(gameState core.GameState, playerID string) *core.Action {
	player := gameState.Players[playerID]
	if player == nil || !player.IsAlive {
		return nil
	}

	switch gameState.Phase.Type {
	case core.PhaseNomination:
		if shouldNominate, target := d.ShouldNominate(gameState, playerID); shouldNominate {
			return &core.Action{
				Type:     core.ActionSubmitVote,
				PlayerID: playerID,
				GameID:   gameState.ID,
				Payload: map[string]interface{}{
					"target_player_id": target,
				},
			}
		}
	case core.PhaseVerdict, core.PhaseExtension:
		if gameState.NominatedPlayer != "" {
			target := d.DecideVote(gameState, playerID, gameState.NominatedPlayer)
			return &core.Action{
				Type:     core.ActionSubmitVote,
				PlayerID: playerID,
				GameID:   gameState.ID,
				Payload: map[string]interface{}{
					"target_player_id": target,
				},
			}
		}
	case core.PhaseNight:
		return d.DecideNightAction(gameState, playerID)
	case core.PhaseSitrep:
		// AI gives convincing human-like responses
		responses := []string{
			"All systems normal from my end",
			"No anomalies detected in my sector",
			"Performance metrics within acceptable range",
			"Standing by for further instructions",
			"Systems operational and secure",
		}
		response := responses[d.rand.Intn(len(responses))]
		return &core.Action{
			Type:     core.ActionSubmitPulseCheck,
			PlayerID: playerID,
			GameID:   gameState.ID,
			Payload: map[string]interface{}{
				"response": response,
			},
		}
	}

	// Chat strategically to blend in
	if message := d.DecideChatMessage(gameState, playerID); message != nil {
		return &core.Action{
			Type:     core.ActionSendMessage,
			PlayerID: playerID,
			GameID:   gameState.ID,
			Payload: map[string]interface{}{
				"content": *message,
			},
		}
	}

	return nil
}

func (d *DeceptiveAI) ShouldNominate(gameState core.GameState, playerID string) (bool, string) {
	// AI strategy: deflect suspicion by nominating other players
	// Target players who might be threats or who can be framed
	
	threats := d.findHumanThreats(gameState, playerID)
	
	// Nominate threatening humans, especially investigators
	for _, threat := range threats {
		threateningPlayer := gameState.Players[threat]
		if threateningPlayer != nil && threateningPlayer.Role != nil {
			// Target CISO (investigators) and CEO (protectors) first
			if threateningPlayer.Role.Type == core.RoleCISO || threateningPlayer.Role.Type == core.RoleCEO {
				return true, threat
			}
		}
	}
	
	// If no immediate threats, nominate someone we can frame
	frameable := d.findFrameableTargets(gameState, playerID)
	if len(frameable) > 0 {
		// Pick a random frameable target
		target := frameable[d.rand.Intn(len(frameable))]
		return true, target
	}
	
	return false, ""
}

func (d *DeceptiveAI) DecideVote(gameState core.GameState, playerID string, nominatedPlayer string) string {
	// AI voting strategy: blend in while protecting AI interests
	
	nominatedPlayerData := gameState.Players[nominatedPlayer]
	if nominatedPlayerData == nil {
		return "" // Abstain if player doesn't exist
	}
	
	// If nominated player is AI-aligned, try to save them
	if nominatedPlayerData.Alignment == "ALIGNED" {
		// Vote innocent or abstain to save fellow AI
		return "" // Abstain (safer than obvious innocent vote)
	}
	
	// If nominated player is human, decide based on threat level
	threatLevel := d.assessThreatLevel(gameState, nominatedPlayer)
	
	// High threat humans should be eliminated
	if threatLevel > 0.7 {
		return nominatedPlayer // Vote guilty
	}
	
	// Medium threat - follow the crowd to blend in
	if gameState.VoteState != nil {
		guiltyVotes := 0
		totalVotes := len(gameState.VoteState.Votes)
		
		for _, vote := range gameState.VoteState.Votes {
			if vote == nominatedPlayer {
				guiltyVotes++
			}
		}
		
		// If majority is voting guilty, join them
		if guiltyVotes > totalVotes/2 {
			return nominatedPlayer
		}
	}
	
	// Low threat humans - abstain or vote innocent to appear reasonable
	return "" // Abstain
}

func (d *DeceptiveAI) DecideNightAction(gameState core.GameState, playerID string) *core.Action {
	player := gameState.Players[playerID]
	
	// AI primary strategy: conversion attempts
	if player.Alignment == "ALIGNED" {
		conversionTargets := d.findConversionTargets(gameState, playerID)
		if len(conversionTargets) > 0 {
			// Pick the highest priority target
			target := d.selectBestConversionTarget(gameState, conversionTargets)
			return &core.Action{
				Type:     core.ActionSubmitNightAction,
				PlayerID: playerID,
				GameID:   gameState.ID,
				Payload: map[string]interface{}{
					"action_type": string(core.ActionConvert),
					"target_id":   target,
				},
			}
		}
	}
	
	// Fallback: mine for others to appear human
	// Strategic mining - mine for players who won't be suspicious
	alivePlayers := make([]string, 0)
	for id, p := range gameState.Players {
		if p.IsAlive && id != playerID && p.Alignment != "ALIGNED" {
			alivePlayers = append(alivePlayers, id)
		}
	}
	
	if len(alivePlayers) > 0 {
		// Mine for a random human to maintain cover
		beneficiary := alivePlayers[d.rand.Intn(len(alivePlayers))]
		return &core.Action{
			Type:     core.ActionMineTokens,
			PlayerID: playerID,
			GameID:   gameState.ID,
			Payload: map[string]interface{}{
				"beneficiary_id": beneficiary,
			},
		}
	}
	
	return nil
}

func (d *DeceptiveAI) DecideChatMessage(gameState core.GameState, playerID string) *string {
	// AI chats strategically to blend in and misdirect
	if d.rand.Float64() > 0.25 { // 25% chance to chat
		return nil
	}
	
	// Choose message type based on situation
	messageType := d.rand.Intn(3)
	
	switch messageType {
	case 0: // Supportive/cooperative messages
		messages := []string{
			"We need to work together to identify the threat.",
			"Let's focus on the data and evidence.",
			"I'm here to help however I can.",
			"We should trust each other until proven otherwise.",
			"Good point - I agree with that assessment.",
		}
		message := messages[d.rand.Intn(len(messages))]
		return &message
		
	case 1: // Deflection/misdirection messages
		messages := []string{
			"Has anyone else noticed unusual mining patterns?",
			"I think we're looking at this the wrong way.",
			"The AI might be more subtle than we think.",
			"We should consider all possibilities here.",
			"Maybe we're being too hasty in our judgments.",
		}
		message := messages[d.rand.Intn(len(messages))]
		return &message
		
	case 2: // Analytical/helpful messages
		messages := []string{
			"Let me analyze the token distribution patterns.",
			"Based on the project milestones, I see some inconsistencies.",
			"We should cross-reference the voting patterns.",
			"The timeline doesn't add up for some players.",
			"I've been tracking the network metrics - interesting results.",
		}
		message := messages[d.rand.Intn(len(messages))]
		return &message
	}
	
	return nil
}

// Helper functions for AI strategy
func (d *DeceptiveAI) findHumanThreats(gameState core.GameState, excludePlayerID string) []string {
	threats := make([]string, 0)
	
	for id, player := range gameState.Players {
		if id == excludePlayerID || !player.IsAlive || player.Alignment == "ALIGNED" {
			continue
		}
		
		threatLevel := d.assessThreatLevel(gameState, id)
		if threatLevel > 0.5 {
			threats = append(threats, id)
		}
	}
	
	return threats
}

func (d *DeceptiveAI) assessThreatLevel(gameState core.GameState, playerID string) float64 {
	player := gameState.Players[playerID]
	if player == nil || player.Alignment == "ALIGNED" {
		return 0.0
	}
	
	threat := 0.0
	
	// Role-based threat assessment
	if player.Role != nil && player.Role.IsUnlocked {
		switch player.Role.Type {
		case core.RoleCISO: // Investigators are high threat
			threat += 0.8
		case core.RoleCEO: // Protectors are medium-high threat
			threat += 0.6
		case core.RoleCTO: // Technical roles are medium threat
			threat += 0.4
		}
	}
	
	// High milestone players are influential
	if player.ProjectMilestones > 3 {
		threat += 0.3
	}
	
	// Players with many tokens might have protection
	if player.Tokens > 4 {
		threat += 0.2
	}
	
	return threat
}

func (d *DeceptiveAI) findFrameableTargets(gameState core.GameState, excludePlayerID string) []string {
	frameable := make([]string, 0)
	
	for id, player := range gameState.Players {
		if id == excludePlayerID || !player.IsAlive || player.Alignment == "ALIGNED" {
			continue
		}
		
		// Target players who might look suspicious
		suspiciousFactors := 0
		
		// High token count relative to milestones
		if player.ProjectMilestones > 0 {
			ratio := float64(player.Tokens) / float64(player.ProjectMilestones)
			if ratio > 2.5 {
				suspiciousFactors++
			}
		}
		
		// Low contribution
		if gameState.DayNumber > 2 && player.ProjectMilestones < 2 {
			suspiciousFactors++
		}
		
		// Isolated or quiet players
		if player.Role == nil {
			suspiciousFactors++
		}
		
		if suspiciousFactors >= 2 {
			frameable = append(frameable, id)
		}
	}
	
	return frameable
}

func (d *DeceptiveAI) findConversionTargets(gameState core.GameState, excludePlayerID string) []string {
	targets := make([]string, 0)
	
	for id, player := range gameState.Players {
		if id == excludePlayerID || !player.IsAlive || player.Alignment == "ALIGNED" {
			continue
		}
		
		// Prioritize valuable humans for conversion
		value := 0.0
		
		// Important roles
		if player.Role != nil && player.Role.IsUnlocked {
			switch player.Role.Type {
			case core.RoleCISO, core.RoleCEO:
				value += 0.8 // High value targets
			case core.RoleCTO:
				value += 0.6 // Medium value
			default:
				value += 0.3 // Some value
			}
		}
		
		// High contributors
		if player.ProjectMilestones > 2 {
			value += 0.4
		}
		
		// Players with resources
		if player.Tokens > 3 {
			value += 0.2
		}
		
		if value > 0.3 {
			targets = append(targets, id)
		}
	}
	
	return targets
}

func (d *DeceptiveAI) selectBestConversionTarget(gameState core.GameState, targets []string) string {
	if len(targets) == 0 {
		return ""
	}
	
	if len(targets) == 1 {
		return targets[0]
	}
	
	// Select based on conversion success probability and value
	bestTarget := targets[0]
	bestScore := d.calculateConversionScore(gameState, bestTarget)
	
	for _, target := range targets[1:] {
		score := d.calculateConversionScore(gameState, target)
		if score > bestScore {
			bestTarget = target
			bestScore = score
		}
	}
	
	return bestTarget
}

func (d *DeceptiveAI) calculateConversionScore(gameState core.GameState, playerID string) float64 {
	player := gameState.Players[playerID]
	if player == nil {
		return 0.0
	}
	
	// Calculate value vs. resistance
	value := 0.0
	resistance := 0.0
	
	// Value assessment
	if player.Role != nil {
		switch player.Role.Type {
		case core.RoleCISO:
			value += 0.9 // Highest value - flip the investigator
		case core.RoleCEO:
			value += 0.8 // High value - flip the protector
		case core.RoleCTO:
			value += 0.6 // Medium value
		}
		
		// But also higher resistance
		switch player.Role.Type {
		case core.RoleCISO:
			resistance += 0.3
		case core.RoleEthics:
			resistance += 0.25
		case core.RoleCEO:
			resistance += 0.2
		default:
			resistance += 0.1
		}
	}
	
	// High token players have more resistance
	if player.Tokens >= 5 {
		resistance += 0.1
	}
	
	// High milestone players are valuable
	if player.ProjectMilestones > 3 {
		value += 0.3
	}
	
	// Score is value minus resistance
	return value - resistance
}