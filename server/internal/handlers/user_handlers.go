package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/xjhc/alignment/server/internal/store"
)

// UserHandlers handles user profile related HTTP requests
type UserHandlers struct {
	postgresStore *store.PostgresStore
}

// NewUserHandlers creates a new user handlers instance
func NewUserHandlers(postgresStore *store.PostgresStore) *UserHandlers {
	return &UserHandlers{
		postgresStore: postgresStore,
	}
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// CreateUserResponse represents the response after creating a user
type CreateUserResponse struct {
	Player *store.Player `json:"player"`
}

// GetUserProfileResponse represents the user profile response
type GetUserProfileResponse struct {
	Player      *store.Player         `json:"player"`
	GameHistory []store.GameHistory   `json:"gameHistory"`
	Avatars     []store.Avatar        `json:"avatars"`
	Titles      []store.Title         `json:"titles"`
	Achievements []store.Achievement  `json:"achievements"`
}

// EquipItemRequest represents the request to equip an avatar or title
type EquipItemRequest struct {
	Avatar string `json:"avatar,omitempty"`
	Title  string `json:"title,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateUser handles user creation
func (uh *UserHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.DisplayName == "" || req.Email == "" {
		http.Error(w, "Username, displayName, and email are required", http.StatusBadRequest)
		return
	}

	// Create the player
	player, err := uh.postgresStore.CreatePlayer(req.Username, req.DisplayName, req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			http.Error(w, "Username or email already exists", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	response := CreateUserResponse{
		Player: player,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetUserProfile handles fetching user profile data
func (uh *UserHandlers) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path or query parameter
	// For now, we'll use a query parameter
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get player profile
	player, err := uh.postgresStore.GetPlayerByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get user profile: %v", err), http.StatusInternalServerError)
		return
	}

	// Get game history
	gameHistory, err := uh.postgresStore.GetPlayerGameHistory(userID, 20) // Last 20 games
	if err != nil {
		gameHistory = []store.GameHistory{} // Continue with empty history if fetch fails
	}

	// Get all avatars
	avatars, err := uh.postgresStore.GetAllAvatars()
	if err != nil {
		avatars = []store.Avatar{} // Continue with empty avatars if fetch fails
	}

	// Get all titles
	titles, err := uh.postgresStore.GetAllTitles()
	if err != nil {
		titles = []store.Title{} // Continue with empty titles if fetch fails
	}

	// Get all achievements
	achievements, err := uh.postgresStore.GetAllAchievements()
	if err != nil {
		achievements = []store.Achievement{} // Continue with empty achievements if fetch fails
	}

	response := GetUserProfileResponse{
		Player:       player,
		GameHistory:  gameHistory,
		Avatars:      avatars,
		Titles:       titles,
		Achievements: achievements,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// EquipItem handles equipping avatars and titles
func (uh *UserHandlers) EquipItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path or query parameter
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req EquipItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current player to verify ownership of items
	player, err := uh.postgresStore.GetPlayerByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get user: %v", err), http.StatusInternalServerError)
		return
	}

	// Validate that the player owns the items they're trying to equip
	currentAvatar := player.EquippedAvatar
	currentTitle := player.EquippedTitle

	if req.Avatar != "" {
		// Check if player owns this avatar
		owned := false
		for _, ownedAvatar := range player.UnlockedAvatars {
			if ownedAvatar == req.Avatar {
				owned = true
				break
			}
		}
		if !owned {
			http.Error(w, "Avatar not owned by player", http.StatusForbidden)
			return
		}
		currentAvatar = req.Avatar
	}

	if req.Title != "" {
		// Check if player owns this title
		owned := false
		for _, ownedTitle := range player.UnlockedTitles {
			if ownedTitle == req.Title {
				owned = true
				break
			}
		}
		if !owned {
			http.Error(w, "Title not owned by player", http.StatusForbidden)
			return
		}
		currentTitle = req.Title
	}

	// Update the player's equipment
	err = uh.postgresStore.UpdatePlayerEquipment(userID, currentAvatar, currentTitle)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update equipment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Equipment updated successfully",
		"avatar":  currentAvatar,
		"title":   currentTitle,
	})
}

// GiveKudosRequest represents the request to give kudos
type GiveKudosRequest struct {
	ReceiverID string `json:"receiverId"`
	GameID     string `json:"gameId"`
}

// GiveKudos handles giving kudos to another player
func (uh *UserHandlers) GiveKudos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract giver ID from URL path or query parameter
	giverID := r.URL.Query().Get("id")
	if giverID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req GiveKudosRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ReceiverID == "" || req.GameID == "" {
		http.Error(w, "ReceiverID and GameID are required", http.StatusBadRequest)
		return
	}

	// Prevent self-kudos
	if giverID == req.ReceiverID {
		http.Error(w, "Cannot give kudos to yourself", http.StatusBadRequest)
		return
	}

	err := uh.postgresStore.GiveKudos(giverID, req.ReceiverID, req.GameID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to give kudos: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Kudos given successfully",
	})
}

// ReportPlayerRequest represents the request to report a player
type ReportPlayerRequest struct {
	ReportedID  string `json:"reportedId"`
	GameID      string `json:"gameId"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

// ReportPlayer handles reporting a player for misconduct
func (uh *UserHandlers) ReportPlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract reporter ID from URL path or query parameter
	reporterID := r.URL.Query().Get("id")
	if reporterID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req ReportPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ReportedID == "" || req.Reason == "" {
		http.Error(w, "ReportedID and Reason are required", http.StatusBadRequest)
		return
	}

	// Prevent self-reporting
	if reporterID == req.ReportedID {
		http.Error(w, "Cannot report yourself", http.StatusBadRequest)
		return
	}

	err := uh.postgresStore.CreatePlayerReport(reporterID, req.ReportedID, req.GameID, req.Reason, req.Description)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to report player: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Player reported successfully",
	})
}

// UpdateSettingsRequest represents the request to update user settings
type UpdateSettingsRequest struct {
	DisableLoebmateHints bool `json:"disableLoebmateHints"`
}

// BlockPlayerRequest represents the request to block a player
type BlockPlayerRequest struct {
	BlockedID string `json:"blockedId"`
}

// UpdateSettings handles updating user settings
func (uh *UserHandlers) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path or query parameter
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the player settings in the database
	err := uh.postgresStore.UpdatePlayerSettings(userID, req.DisableLoebmateHints)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to update settings: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Settings updated successfully",
	})
}

// BlockPlayer handles blocking a player
func (uh *UserHandlers) BlockPlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract blocker ID from URL path or query parameter
	blockerID := r.URL.Query().Get("id")
	if blockerID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req BlockPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.BlockedID == "" {
		http.Error(w, "BlockedID is required", http.StatusBadRequest)
		return
	}

	// Prevent self-blocking
	if blockerID == req.BlockedID {
		http.Error(w, "Cannot block yourself", http.StatusBadRequest)
		return
	}

	err := uh.postgresStore.BlockPlayer(blockerID, req.BlockedID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to block player: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Player blocked successfully",
	})
}