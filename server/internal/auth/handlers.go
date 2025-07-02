package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// AuthHandlers handles HTTP authentication endpoints
type AuthHandlers struct {
	authService *AuthService
	// Store for temporary state validation (in production, use Redis)
	stateStore map[string]time.Time
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(authService *AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		stateStore:  make(map[string]time.Time),
	}
}

// LoginHandler redirects to Discord OAuth
func (ah *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authURL, state, err := ah.authService.GetAuthURL()
	if err != nil {
		log.Printf("Failed to generate auth URL: %v", err)
		http.Error(w, "Failed to initiate authentication", http.StatusInternalServerError)
		return
	}

	// Store state for validation (with expiration)
	ah.stateStore[state] = time.Now().Add(10 * time.Minute)

	// Clean up expired states
	ah.cleanupExpiredStates()

	// Redirect to Discord
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// CallbackHandler handles OAuth callback from Discord
func (ah *AuthHandlers) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		log.Printf("OAuth error: %s", errorParam)
		http.Error(w, "Authentication failed", http.StatusBadRequest)
		return
	}

	if code == "" || state == "" {
		http.Error(w, "Missing code or state parameter", http.StatusBadRequest)
		return
	}

	// Validate state
	if !ah.validateState(state) {
		log.Printf("Invalid or expired state: %s", state)
		http.Error(w, "Invalid authentication state", http.StatusBadRequest)
		return
	}

	// Exchange code for user info
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := ah.authService.ExchangeCodeForToken(ctx, code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to authenticate with Discord", http.StatusInternalServerError)
		return
	}

	// Generate JWT
	jwtToken, err := ah.authService.GenerateJWT(user)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set JWT as HttpOnly cookie
	cookie := &http.Cookie{
		Name:     "alignment_session",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   24 * 60 * 60, // 24 hours
	}
	http.SetCookie(w, cookie)

	log.Printf("User authenticated: %s (%s)", user.Username, user.ID)

	// Redirect to frontend
	http.Redirect(w, r, "/lobby-list", http.StatusTemporaryRedirect)
}

// UserInfoResponse represents the user info API response
type UserInfoResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Avatar       string `json:"avatar"`
	Provider     string `json:"provider"`
	IsAuthenticated bool `json:"is_authenticated"`
}

// MeHandler returns current user information from JWT
func (ah *AuthHandlers) MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get JWT from cookie
	cookie, err := r.Cookie("alignment_session")
	if err != nil {
		// No session cookie - user is not authenticated
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "not authenticated",
		})
		return
	}

	// Validate JWT
	claims, err := ah.authService.ValidateJWT(cookie.Value)
	if err != nil {
		log.Printf("Invalid JWT: %v", err)
		// Clear invalid cookie
		http.SetCookie(w, &http.Cookie{
			Name:   "alignment_session",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid session",
		})
		return
	}

	// Return user info
	response := UserInfoResponse{
		ID:           claims.UserID,
		Name:         claims.Username,
		Avatar:       claims.Avatar,
		Provider:     claims.Provider,
		IsAuthenticated: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LogoutHandler clears the authentication session
func (ah *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear the session cookie
	cookie := &http.Cookie{
		Name:   "alignment_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "logged out successfully",
	})
}

// validateState checks if the state is valid and not expired
func (ah *AuthHandlers) validateState(state string) bool {
	expiry, exists := ah.stateStore[state]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		delete(ah.stateStore, state)
		return false
	}

	// Remove used state
	delete(ah.stateStore, state)
	return true
}

// cleanupExpiredStates removes expired states from memory
func (ah *AuthHandlers) cleanupExpiredStates() {
	now := time.Now()
	for state, expiry := range ah.stateStore {
		if now.After(expiry) {
			delete(ah.stateStore, state)
		}
	}
}

// GetAuthenticatedUserID extracts the user ID from request JWT cookie
func (ah *AuthHandlers) GetAuthenticatedUserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie("alignment_session")
	if err != nil {
		return "", fmt.Errorf("no session cookie")
	}

	claims, err := ah.authService.ValidateJWT(cookie.Value)
	if err != nil {
		return "", fmt.Errorf("invalid session: %w", err)
	}

	return claims.UserID, nil
}