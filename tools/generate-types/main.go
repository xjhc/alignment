package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type TypeGenerator struct {
	output strings.Builder
}

func (tg *TypeGenerator) parseGoFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	// Regex patterns for parsing Go code
	eventTypeRegex := regexp.MustCompile(`^\s*Event\w+\s+EventType\s*=\s*"([^"]+)"`)
	actionTypeRegex := regexp.MustCompile(`^\s*Action\w+\s+ActionType\s*=\s*"([^"]+)"`)
	phaseTypeRegex := regexp.MustCompile(`^\s*Phase\w+\s+PhaseType\s*=\s*"([^"]+)"`)
	roleTypeRegex := regexp.MustCompile(`^\s*Role\w+\s+RoleType\s*=\s*"([^"]+)"`)
	kpiTypeRegex := regexp.MustCompile(`^\s*KPI\w+\s+KPIType\s*=\s*"([^"]+)"`)
	voteTypeRegex := regexp.MustCompile(`^\s*Vote\w+\s+VoteType\s*=\s*"([^"]+)"`)
	
	var eventTypes, actionTypes, phaseTypes, roleTypes, kpiTypes, voteTypes []string

	for scanner.Scan() {
		line := scanner.Text()
		
		if match := eventTypeRegex.FindStringSubmatch(line); match != nil {
			eventTypes = append(eventTypes, match[1])
		} else if match := actionTypeRegex.FindStringSubmatch(line); match != nil {
			actionTypes = append(actionTypes, match[1])
		} else if match := phaseTypeRegex.FindStringSubmatch(line); match != nil {
			phaseTypes = append(phaseTypes, match[1])
		} else if match := roleTypeRegex.FindStringSubmatch(line); match != nil {
			roleTypes = append(roleTypes, match[1])
		} else if match := kpiTypeRegex.FindStringSubmatch(line); match != nil {
			kpiTypes = append(kpiTypes, match[1])
		} else if match := voteTypeRegex.FindStringSubmatch(line); match != nil {
			voteTypes = append(voteTypes, match[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Generate TypeScript enums
	tg.generateEnum("ServerEventType", eventTypes)
	tg.generateEnum("ClientActionType", actionTypes)
	tg.generateEnum("PhaseType", phaseTypes)
	tg.generateEnum("RoleType", roleTypes)
	tg.generateEnum("KPIType", kpiTypes)
	tg.generateEnum("VoteType", voteTypes)

	return nil
}

func (tg *TypeGenerator) generateEnum(enumName string, values []string) {
	if len(values) == 0 {
		return
	}

	tg.output.WriteString(fmt.Sprintf("// Generated %s enum\n", enumName))
	tg.output.WriteString(fmt.Sprintf("export enum %s {\n", enumName))
	
	for _, value := range values {
		enumKey := tg.toEnumKey(value)
		tg.output.WriteString(fmt.Sprintf("  %s = %q,\n", enumKey, value))
	}
	
	tg.output.WriteString("}\n\n")
}

func (tg *TypeGenerator) toEnumKey(value string) string {
	// Convert UPPER_CASE to PascalCase for enum keys
	parts := strings.Split(value, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}
	return result.String()
}

func (tg *TypeGenerator) generateInterfaces() {
	// Generate basic interfaces that correspond to Go structs
	tg.output.WriteString(`// Generated interfaces from Go structs
export interface GeneratedEvent {
  id: string;
  type: ServerEventType;
  gameId: string;
  playerId?: string;
  timestamp: string;
  payload: Record<string, any>;
}

export interface GeneratedAction {
  type: ClientActionType;
  playerId: string;
  gameId: string;
  timestamp: string;
  payload: Record<string, any>;
}

export interface GeneratedPlayer {
  id: string;
  name: string;
  jobTitle: string;
  controlType: string;
  isAlive: boolean;
  tokens: number;
  projectMilestones: number;
  statusMessage: string;
  joinedAt: string;
  alignment?: string;
  role?: GeneratedRole;
  personalKPI?: GeneratedPersonalKPI;
  aiEquity?: number;
  hasUsedAbility?: boolean;
  lastNightAction?: GeneratedNightAction;
  hasSubmittedPulseCheck?: boolean;
  slackStatus?: string;
  partingShot?: string;
  systemShocks?: GeneratedSystemShock[];
}

export interface GeneratedRole {
  type: RoleType;
  name: string;
  description: string;
  isUnlocked: boolean;
  ability?: GeneratedAbility;
}

export interface GeneratedAbility {
  name: string;
  description: string;
  isReady: boolean;
}

export interface GeneratedPersonalKPI {
  type: KPIType;
  description: string;
  progress: number;
  target: number;
  isCompleted: boolean;
  reward: string;
}

export interface GeneratedSystemShock {
  type: string;
  description: string;
  expiresAt: string;
  isActive: boolean;
}

export interface GeneratedNightAction {
  type: string;
  targetId?: string;
}

export interface GeneratedChatMessage {
  id: string;
  playerID: string;
  playerName: string;
  message: string;
  timestamp: string;
  isSystem: boolean;
  channelID?: string;
}

export interface GeneratedPhase {
  type: PhaseType;
  startTime: string;
  duration: number;
}

export interface GeneratedVoteState {
  type: VoteType;
  votes: Record<string, string>;
  tokenWeights: Record<string, number>;
  results: Record<string, number>;
  isComplete: boolean;
}

export interface GeneratedGameSettings {
  maxPlayers: number;
  minPlayers: number;
  sitrepDuration: number;
  pulseCheckDuration: number;
  discussionDuration: number;
  extensionDuration: number;
  nominationDuration: number;
  trialDuration: number;
  verdictDuration: number;
  nightDuration: number;
  startingTokens: number;
  votingThreshold: number;
  customSettings?: Record<string, any>;
}

export interface GeneratedCrisisEvent {
  type: string;
  title: string;
  description: string;
  effects: Record<string, any>;
}

export interface GeneratedWinCondition {
  winner: string;
  condition: string;
  description: string;
}

// Union types for easier usage
export type AnyEventType = keyof typeof ServerEventType;
export type AnyActionType = keyof typeof ClientActionType;
export type AnyPhaseType = keyof typeof PhaseType;
export type AnyRoleType = keyof typeof RoleType;
export type AnyKPIType = keyof typeof KPIType;
export type AnyVoteType = keyof typeof VoteType;

`)
}

func (tg *TypeGenerator) generateAll() {
	tg.output.WriteString("// AUTO-GENERATED FILE - DO NOT EDIT\n")
	tg.output.WriteString("// Generated from Go core package types\n")
	tg.output.WriteString("// Run 'npm run generate:types' to update\n\n")
	
	// Parse the core types file
	if err := tg.parseGoFile("../../core/types.go"); err != nil {
		fmt.Printf("Error parsing Go file: %v\n", err)
		os.Exit(1)
	}
	
	// Generate interfaces
	tg.generateInterfaces()
}

func main() {
	generator := &TypeGenerator{}
	generator.generateAll()
	
	// Write to the output file
	outputPath := "../../client/src/types/generated.ts"
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}
	
	if err := os.WriteFile(outputPath, []byte(generator.output.String()), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Generated TypeScript definitions at %s\n", outputPath)
}