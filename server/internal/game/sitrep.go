package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/xjhc/alignment/core"
)

// SitrepGenerator handles the creation of daily situation reports
type SitrepGenerator struct {
	gameState *core.GameState
	rng       *rand.Rand
}

// NewSitrepGenerator creates a new SITREP generator
func NewSitrepGenerator(gameState *core.GameState) *SitrepGenerator {
	return &SitrepGenerator{
		gameState: gameState,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SitrepSection represents a section of the daily report
type SitrepSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"` // "standard", "classified", "redacted"
}

// DailySitrep represents the complete daily situation report
type DailySitrep struct {
	DayNumber  int             `json:"day_number"`
	Date       time.Time       `json:"date"`
	Sections   []SitrepSection `json:"sections"`
	AlertLevel string          `json:"alert_level"`
	Summary    string          `json:"summary"`
	FooterNote string          `json:"footer_note"`
}

// GenerateDailySitrep creates the complete SITREP for the current day
func (sg *SitrepGenerator) GenerateDailySitrep() DailySitrep {
	sitrep := DailySitrep{
		DayNumber:  sg.gameState.DayNumber,
		Date:       getCurrentTime(),
		Sections:   make([]SitrepSection, 0),
		AlertLevel: sg.determineAlertLevel(),
	}

	// Standard sections in order
	sitrep.Sections = append(sitrep.Sections, sg.generateExecutiveSummary())
	sitrep.Sections = append(sitrep.Sections, sg.generatePersonnelStatus())
	sitrep.Sections = append(sitrep.Sections, sg.generateOperationalMetrics())
	sitrep.Sections = append(sitrep.Sections, sg.generateSecurityAlerts())
	sitrep.Sections = append(sitrep.Sections, sg.generateProjectStatus())
	sitrep.Sections = append(sitrep.Sections, sg.generateThreatAssessment())
	sitrep.Sections = append(sitrep.Sections, sg.generateRecommendations())

	// Apply hotfix redaction if active
	sg.applyHotfixRedaction(&sitrep)

	// Generate summary and footer
	sitrep.Summary = sg.generateSummary()
	sitrep.FooterNote = sg.generateFooterNote()

	return sitrep
}

// determineAlertLevel calculates the current threat level
func (sg *SitrepGenerator) determineAlertLevel() string {
	// Base alert level on various factors
	alertScore := 0

	// Check for recent eliminations
	if sg.gameState.DayNumber > 1 {
		alertScore += 1
	}

	// Check for crisis events
	if sg.gameState.CrisisEvent != nil {
		alertScore += 2
	}

	// Check player count (fewer players = higher alert)
	aliveCount := 0
	for _, player := range sg.gameState.Players {
		if player.IsAlive {
			aliveCount++
		}
	}
	if aliveCount <= 4 {
		alertScore += 2
	} else if aliveCount <= 6 {
		alertScore += 1
	}

	// Check for high AI equity
	maxAIEquity := 0
	for _, player := range sg.gameState.Players {
		if player.AIEquity > maxAIEquity {
			maxAIEquity = player.AIEquity
		}
	}
	if maxAIEquity >= 3 {
		alertScore += 2
	} else if maxAIEquity >= 2 {
		alertScore += 1
	}

	// Determine alert level
	switch {
	case alertScore >= 5:
		return "CRITICAL"
	case alertScore >= 3:
		return "HIGH"
	case alertScore >= 1:
		return "ELEVATED"
	default:
		return "NORMAL"
	}
}

// generateExecutiveSummary creates the executive summary section
func (sg *SitrepGenerator) generateExecutiveSummary() SitrepSection {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("**Day %d Operations Summary**\n\n", sg.gameState.DayNumber))

	// Count personnel
	aliveCount := 0
	eliminatedCount := 0
	for _, player := range sg.gameState.Players {
		if player.IsAlive {
			aliveCount++
		} else {
			eliminatedCount++
		}
	}

	content.WriteString(fmt.Sprintf("• Active personnel: %d\n", aliveCount))
	if eliminatedCount > 0 {
		content.WriteString(fmt.Sprintf("• Personnel no longer with company: %d\n", eliminatedCount))
	}

	// Current phase
	content.WriteString(fmt.Sprintf("• Current phase: %s\n", strings.Title(strings.ToLower(string(sg.gameState.Phase.Type)))))

	// Crisis status
	if sg.gameState.CrisisEvent != nil {
		content.WriteString(fmt.Sprintf("• **Active Crisis**: %s\n", sg.gameState.CrisisEvent.Title))
	} else {
		content.WriteString("• No active crisis events\n")
	}

	return SitrepSection{
		Title:   "Executive Summary",
		Content: content.String(),
		Type:    "standard",
	}
}

// generatePersonnelStatus creates personnel status overview
func (sg *SitrepGenerator) generatePersonnelStatus() SitrepSection {
	var content strings.Builder

	content.WriteString("**Personnel Status Report**\n\n")

	// Sort players by status and role
	activePersonnel := make([]*core.Player, 0)
	for _, player := range sg.gameState.Players {
		if player.IsAlive {
			activePersonnel = append(activePersonnel, player)
		}
	}

	// Sort by role prominence (CEO, C-level, VP)
	sort.Slice(activePersonnel, func(i, j int) bool {
		return sg.getRoleWeight(activePersonnel[i]) > sg.getRoleWeight(activePersonnel[j])
	})

	content.WriteString("**Active Personnel:**\n")
	for _, player := range activePersonnel {
		status := "Operational"
		if len(player.SystemShocks) > 0 {
			status = "Affected by system issues"
		}

		roleInfo := player.JobTitle
		if player.Role != nil && player.Role.Type != "" {
			roleInfo = player.Role.Name
		}

		content.WriteString(fmt.Sprintf("• %s (%s) - %s\n", player.Name, roleInfo, status))
	}

	// Recent departures
	recentDepartures := make([]*core.Player, 0)
	for _, player := range sg.gameState.Players {
		if !player.IsAlive {
			recentDepartures = append(recentDepartures, player)
		}
	}

	if len(recentDepartures) > 0 {
		content.WriteString("\n**Recent Departures:**\n")
		for _, player := range recentDepartures {
			roleInfo := "Role undisclosed"
			if player.Role != nil && player.Role.Type != "" {
				roleInfo = player.Role.Name
			}
			content.WriteString(fmt.Sprintf("• %s (%s) - No longer with company\n", player.Name, roleInfo))
		}
	}

	return SitrepSection{
		Title:   "Personnel Status",
		Content: content.String(),
		Type:    "standard",
	}
}

// generateOperationalMetrics creates operational metrics section
func (sg *SitrepGenerator) generateOperationalMetrics() SitrepSection {
	var content strings.Builder

	content.WriteString("**Operational Metrics**\n\n")

	// Token distribution analysis
	totalTokens := 0
	tokenHolders := make(map[string]int)
	for playerID, player := range sg.gameState.Players {
		if player.IsAlive {
			totalTokens += player.Tokens
			tokenHolders[playerID] = player.Tokens
		}
	}

	content.WriteString(fmt.Sprintf("• Total operational tokens in circulation: %d\n", totalTokens))

	// Average tokens
	if len(tokenHolders) > 0 {
		avgTokens := float64(totalTokens) / float64(len(tokenHolders))
		content.WriteString(fmt.Sprintf("• Average tokens per active personnel: %.1f\n", avgTokens))
	}

	// Project milestone progress
	totalMilestones := 0
	completedProjects := 0
	for _, player := range sg.gameState.Players {
		if player.IsAlive {
			totalMilestones += player.ProjectMilestones
			if player.ProjectMilestones >= 3 {
				completedProjects++
			}
		}
	}

	content.WriteString(fmt.Sprintf("• Total project milestones achieved: %d\n", totalMilestones))
	content.WriteString(fmt.Sprintf("• Personnel with completed projects: %d\n", completedProjects))

	// Mining activity (if night actions were resolved)
	if len(sg.gameState.Players) > 0 {
		content.WriteString("• Resource allocation efficiency: ")
		if totalTokens > len(sg.gameState.Players)*2 {
			content.WriteString("High\n")
		} else if totalTokens > len(sg.gameState.Players) {
			content.WriteString("Moderate\n")
		} else {
			content.WriteString("Low\n")
		}
	}

	return SitrepSection{
		Title:   "Operational Metrics",
		Content: content.String(),
		Type:    "standard",
	}
}

// generateSecurityAlerts creates security alerts section
func (sg *SitrepGenerator) generateSecurityAlerts() SitrepSection {
	var content strings.Builder

	content.WriteString("**Security Status**\n\n")

	// Check for anomalies and threats
	anomalies := sg.detectSecurityAnomalies()

	if len(anomalies) == 0 {
		content.WriteString("• No significant security anomalies detected\n")
		content.WriteString("• All personnel access patterns within normal parameters\n")
		content.WriteString("• System integrity checks: PASSED\n")
	} else {
		content.WriteString("**Detected Anomalies:**\n")
		for _, anomaly := range anomalies {
			content.WriteString(fmt.Sprintf("• %s\n", anomaly))
		}
	}

	// System shock reports
	affectedPersonnel := 0
	for _, player := range sg.gameState.Players {
		if player.IsAlive && len(player.SystemShocks) > 0 {
			affectedPersonnel++
		}
	}

	if affectedPersonnel > 0 {
		content.WriteString(fmt.Sprintf("\n• Personnel affected by system issues: %d\n", affectedPersonnel))
		content.WriteString("• Recommend system diagnostics and recovery protocols\n")
	}

	return SitrepSection{
		Title:   "Security Alerts",
		Content: content.String(),
		Type:    "classified",
	}
}

// generateProjectStatus creates project status section
func (sg *SitrepGenerator) generateProjectStatus() SitrepSection {
	var content strings.Builder

	content.WriteString("**Project Status Dashboard**\n\n")

	// Milestone distribution
	milestoneDistribution := map[int]int{0: 0, 1: 0, 2: 0, 3: 0}
	roleAbilitiesUnlocked := 0

	for _, player := range sg.gameState.Players {
		if player.IsAlive {
			milestones := player.ProjectMilestones
			if milestones > 3 {
				milestones = 3
			}
			milestoneDistribution[milestones]++

			if player.Role != nil && player.Role.IsUnlocked {
				roleAbilitiesUnlocked++
			}
		}
	}

	content.WriteString("**Milestone Progress Distribution:**\n")
	for milestones := 0; milestones <= 3; milestones++ {
		count := milestoneDistribution[milestones]
		if count > 0 {
			status := "In Progress"
			if milestones == 3 {
				status = "Completed"
			} else if milestones == 0 {
				status = "Not Started"
			}
			content.WriteString(fmt.Sprintf("• %d milestones (%s): %d personnel\n", milestones, status, count))
		}
	}

	content.WriteString(fmt.Sprintf("\n• Personnel with unlocked role capabilities: %d\n", roleAbilitiesUnlocked))

	// KPI progress (if any players have KPIs)
	kpiProgress := 0
	kpiCompleted := 0
	for _, player := range sg.gameState.Players {
		if player.IsAlive && player.PersonalKPI != nil {
			if player.PersonalKPI.Progress > 0 {
				kpiProgress++
			}
			if player.PersonalKPI.IsCompleted {
				kpiCompleted++
			}
		}
	}

	if kpiProgress > 0 || kpiCompleted > 0 {
		content.WriteString(fmt.Sprintf("• Personnel making personal KPI progress: %d\n", kpiProgress))
		content.WriteString(fmt.Sprintf("• Completed personal KPIs: %d\n", kpiCompleted))
	}

	return SitrepSection{
		Title:   "Project Status",
		Content: content.String(),
		Type:    "standard",
	}
}

// generateThreatAssessment creates threat assessment section
func (sg *SitrepGenerator) generateThreatAssessment() SitrepSection {
	var content strings.Builder

	content.WriteString("**Threat Assessment**\n\n")

	// Analyze potential AI infiltration indicators
	suspiciousActivity := sg.analyzeSuspiciousActivity()

	if len(suspiciousActivity) == 0 {
		content.WriteString("• No indicators of AI infiltration detected\n")
		content.WriteString("• All personnel behavior within expected parameters\n")
		content.WriteString("• Recommendation: Maintain current security posture\n")
	} else {
		content.WriteString("**Potential Infiltration Indicators:**\n")
		for _, indicator := range suspiciousActivity {
			content.WriteString(fmt.Sprintf("• %s\n", indicator))
		}
		content.WriteString("\n• Recommendation: Enhanced monitoring and verification protocols\n")
	}

	// Crisis threat level
	if sg.gameState.CrisisEvent != nil {
		content.WriteString(fmt.Sprintf("\n**Active Crisis Threat**: %s\n", sg.gameState.CrisisEvent.Title))
		content.WriteString("• Enhanced security protocols in effect\n")
		content.WriteString("• Recommend immediate response coordination\n")
	}

	return SitrepSection{
		Title:   "Threat Assessment",
		Content: content.String(),
		Type:    "classified",
	}
}

// generateRecommendations creates recommendations section
func (sg *SitrepGenerator) generateRecommendations() SitrepSection {
	var content strings.Builder

	content.WriteString("**Strategic Recommendations**\n\n")

	recommendations := sg.generateStrategicRecommendations()

	content.WriteString("**Priority Actions:**\n")
	for i, rec := range recommendations {
		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
	}

	// Add operational guidance
	content.WriteString("\n**Operational Guidance:**\n")
	content.WriteString("• Continue monitoring all personnel for anomalous behavior\n")
	content.WriteString("• Maintain secure communication protocols\n")
	content.WriteString("• Report any suspicious activity immediately\n")

	if sg.gameState.Phase.Type == core.PhaseNight {
		content.WriteString("• Night shift protocols in effect - limit unnecessary movement\n")
	}

	return SitrepSection{
		Title:   "Recommendations",
		Content: content.String(),
		Type:    "standard",
	}
}

// applyHotfixRedaction applies VP Platforms hotfix redaction if active
func (sg *SitrepGenerator) applyHotfixRedaction(sitrep *DailySitrep) {
	// Check if hotfix redaction is active
	if sg.gameState.CrisisEvent != nil {
		if section, exists := sg.gameState.CrisisEvent.Effects["redacted_section"]; exists {
			sectionName := section.(string)

			// Find and redact the specified section
			for i := range sitrep.Sections {
				if sg.matchesSectionType(sitrep.Sections[i].Title, sectionName) {
					sitrep.Sections[i].Content = "**[REDACTED]**\n\nThis section has been redacted due to an active hotfix deployment.\nInformation is temporarily unavailable while systems are being patched.\n\nFor assistance, contact the VP of Platforms."
					sitrep.Sections[i].Type = "redacted"
					break
				}
			}
		}
	}
}

// matchesSectionType determines if a section matches the redaction target
func (sg *SitrepGenerator) matchesSectionType(sectionTitle, redactionTarget string) bool {
	switch redactionTarget {
	case "security_alerts", "investigation_results":
		return strings.Contains(strings.ToLower(sectionTitle), "security") ||
			strings.Contains(strings.ToLower(sectionTitle), "threat")
	case "operational_metrics", "mining_results":
		return strings.Contains(strings.ToLower(sectionTitle), "operational") ||
			strings.Contains(strings.ToLower(sectionTitle), "metrics")
	case "personnel_status":
		return strings.Contains(strings.ToLower(sectionTitle), "personnel")
	case "project_status":
		return strings.Contains(strings.ToLower(sectionTitle), "project")
	default:
		return strings.Contains(strings.ToLower(sectionTitle), strings.ToLower(redactionTarget))
	}
}

// Helper functions for analysis

// detectSecurityAnomalies identifies potential security issues
func (sg *SitrepGenerator) detectSecurityAnomalies() []string {
	anomalies := make([]string, 0)

	// Check for high AI equity accumulation
	for _, player := range sg.gameState.Players {
		if player.IsAlive && player.AIEquity >= 3 {
			anomalies = append(anomalies, "Personnel with elevated AI system access detected")
			break
		}
	}

	// Check for rapid token accumulation
	maxTokens := 0
	for _, player := range sg.gameState.Players {
		if player.IsAlive && player.Tokens > maxTokens {
			maxTokens = player.Tokens
		}
	}
	if maxTokens >= 7 {
		anomalies = append(anomalies, "Unusual resource concentration detected")
	}

	// Check for role ability usage patterns (this would need historical data)
	if sg.gameState.DayNumber > 2 {
		anomalies = append(anomalies, "Analyzing behavioral patterns for deviations")
	}

	return anomalies
}

// analyzeSuspiciousActivity looks for AI infiltration indicators
func (sg *SitrepGenerator) analyzeSuspiciousActivity() []string {
	indicators := make([]string, 0)

	// This would be more sophisticated with historical data
	// For now, provide generic indicators based on game state

	if sg.gameState.DayNumber >= 3 {
		indicators = append(indicators, "Multiple nights of activity - pattern analysis ongoing")
	}

	// Check elimination patterns
	eliminatedCount := 0
	for _, player := range sg.gameState.Players {
		if !player.IsAlive {
			eliminatedCount++
		}
	}

	if eliminatedCount >= 2 {
		indicators = append(indicators, "Personnel reduction rate exceeds baseline expectations")
	}

	// Add some randomized realistic indicators
	if sg.rng.Float64() < 0.3 { // 30% chance
		indicators = append(indicators, "Irregular access patterns detected in secure systems")
	}

	if sg.rng.Float64() < 0.2 { // 20% chance
		indicators = append(indicators, "Communication metadata analysis shows anomalous patterns")
	}

	return indicators
}

// generateStrategicRecommendations creates context-aware recommendations
func (sg *SitrepGenerator) generateStrategicRecommendations() []string {
	recommendations := make([]string, 0)

	// Based on game state
	aliveCount := 0
	for _, player := range sg.gameState.Players {
		if player.IsAlive {
			aliveCount++
		}
	}

	if aliveCount <= 4 {
		recommendations = append(recommendations, "Critical personnel threshold reached - implement emergency protocols")
	}

	if sg.gameState.CrisisEvent != nil {
		recommendations = append(recommendations, "Address active crisis event with coordinated response")
	}

	// Check for low project milestone progress
	lowProgress := 0
	for _, player := range sg.gameState.Players {
		if player.IsAlive && player.ProjectMilestones < 2 {
			lowProgress++
		}
	}
	if lowProgress > aliveCount/2 {
		recommendations = append(recommendations, "Accelerate project milestone completion to unlock personnel capabilities")
	}

	// Generic strategic recommendations
	recommendations = append(recommendations, "Maintain vigilant observation of all personnel interactions")
	recommendations = append(recommendations, "Continue verification of personnel alignment and loyalty")

	if sg.gameState.Phase.Type == core.PhaseNight {
		recommendations = append(recommendations, "Coordinate night operations for maximum security and efficiency")
	}

	return recommendations
}

// generateSummary creates an overall summary for the SITREP
func (sg *SitrepGenerator) generateSummary() string {
	alertLevel := sg.determineAlertLevel()

	switch alertLevel {
	case "CRITICAL":
		return "Company security at critical risk. Immediate executive action required."
	case "HIGH":
		return "Elevated threat level detected. Enhanced monitoring and response protocols active."
	case "ELEVATED":
		return "Potential security concerns identified. Maintain heightened awareness."
	default:
		return "Operations proceeding within normal parameters. Continue standard protocols."
	}
}

// generateFooterNote creates the footer disclaimer
func (sg *SitrepGenerator) generateFooterNote() string {
	return fmt.Sprintf("SITREP generated at %s | Classification: INTERNAL USE ONLY | Report Day %d",
		getCurrentTime().Format("15:04 MST"), sg.gameState.DayNumber)
}

// getRoleWeight returns a weight for sorting roles by importance
func (sg *SitrepGenerator) getRoleWeight(player *core.Player) int {
	if player.Role == nil {
		return 0
	}

	switch player.Role.Type {
	case core.RoleCEO:
		return 10
	case core.RoleCTO, core.RoleCFO, core.RoleCISO, core.RoleCOO:
		return 8
	case core.RoleEthics, core.RolePlatforms:
		return 6
	default:
		return 1
	}
}
