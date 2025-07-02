// Package admin provides secure administrative endpoints for server management
package admin

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/xjhc/alignment/server/internal/actors"
	"github.com/xjhc/alignment/server/internal/health"
	"github.com/xjhc/alignment/server/internal/logger"
)

// Handlers provides secure admin HTTP handlers
type Handlers struct {
	supervisor    *actors.Supervisor
	healthMonitor *health.Monitor
	logger        *slog.Logger
	adminUser     string
	adminPassword string
}

// NewHandlers creates new admin handlers with basic auth credentials
func NewHandlers(supervisor *actors.Supervisor, healthMonitor *health.Monitor) *Handlers {
	// Get admin credentials from environment variables
	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin" // Default username
	}
	
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123" // Default password (should be changed in production)
	}

	return &Handlers{
		supervisor:    supervisor,
		healthMonitor: healthMonitor,
		logger:        logger.WithField("component", "admin_handlers"),
		adminUser:     adminUser,
		adminPassword: adminPassword,
	}
}

// basicAuth middleware for protecting admin endpoints
func (h *Handlers) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			h.logger.Warn("Missing basic auth credentials", "remote_addr", r.RemoteAddr)
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Use constant-time comparison to prevent timing attacks
		userValid := subtle.ConstantTimeCompare([]byte(user), []byte(h.adminUser)) == 1
		passValid := subtle.ConstantTimeCompare([]byte(pass), []byte(h.adminPassword)) == 1

		if !userValid || !passValid {
			h.logger.Warn("Invalid admin credentials", "user", user, "remote_addr", r.RemoteAddr)
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h.logger.Info("Admin access granted", "user", user, "endpoint", r.URL.Path, "remote_addr", r.RemoteAddr)
		next.ServeHTTP(w, r)
	}
}

// ServerStatus represents the overall server status
type ServerStatus struct {
	Health      string              `json:"health"`
	ActiveGames []actors.GameInfo   `json:"active_games"`
	Stats       interface{}         `json:"stats"`
}

// GetGames returns a list of all active games
func (h *Handlers) GetGames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.supervisor.GetStats()

	h.logger.Info("Admin: Listed active games", "count", len(stats.Games))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"games": stats.Games,
		"count": len(stats.Games),
	})
}

// GetGameState returns the detailed state of a specific game
func (h *Handlers) GetGameState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract game ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/admin/games/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] != "state" {
		http.Error(w, "Invalid path. Use /admin/games/{gameId}/state", http.StatusBadRequest)
		return
	}

	gameID := parts[0]
	h.logger.Info("Admin: Getting game state", "game_id", gameID)

	// Get game state from supervisor
	gameState, err := h.supervisor.GetGameState(gameID)
	if err != nil {
		h.logger.Warn("Admin: Failed to get game state", "game_id", gameID, "error", err)
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameState)
}

// DeleteGame forcefully terminates a game
func (h *Handlers) DeleteGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract game ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/admin/games/")
	gameID := strings.Split(path, "/")[0]
	
	if gameID == "" {
		http.Error(w, "Invalid path. Use /admin/games/{gameId}", http.StatusBadRequest)
		return
	}

	h.logger.Warn("Admin: Forcefully terminating game", "game_id", gameID)

	// Terminate the game
	err := h.supervisor.TerminateGame(gameID)
	if err != nil {
		h.logger.Error("Admin: Failed to terminate game", "game_id", gameID, "error", err)
		http.Error(w, "Failed to terminate game", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Admin: Game terminated successfully", "game_id", gameID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Game terminated successfully",
		"game_id": gameID,
	})
}

// GetServerStatus returns comprehensive server status
func (h *Handlers) GetServerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.supervisor.GetStats()
	status := ServerStatus{
		Health:      string(h.healthMonitor.GetStatus()),
		ActiveGames: stats.Games,
		Stats:       stats,
	}

	h.logger.Info("Admin: Retrieved server status")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// SetupRoutes configures the admin routes with authentication
func (h *Handlers) SetupRoutes() {
	http.HandleFunc("/admin", h.ServeAdminUI)               // Public dashboard UI (requires auth via JS)
	http.HandleFunc("/admin/", h.handleAdminRoutes)         // Route admin sub-paths
	http.HandleFunc("/admin/status", h.basicAuth(h.GetServerStatus))
	http.HandleFunc("/admin/games", h.basicAuth(h.GetGames))
	http.HandleFunc("/admin/games/", h.handleGameRoutes)
}

// handleAdminRoutes routes admin sub-paths
func (h *Handlers) handleAdminRoutes(w http.ResponseWriter, r *http.Request) {
	// If the path is exactly "/admin/" (with trailing slash), redirect to "/admin"
	if r.URL.Path == "/admin/" {
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
		return
	}
	
	// Other admin routes are handled by their specific handlers
	http.NotFound(w, r)
}

// handleGameRoutes routes game-specific admin endpoints
func (h *Handlers) handleGameRoutes(w http.ResponseWriter, r *http.Request) {
	// Apply basic auth first
	user, pass, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userValid := subtle.ConstantTimeCompare([]byte(user), []byte(h.adminUser)) == 1
	passValid := subtle.ConstantTimeCompare([]byte(pass), []byte(h.adminPassword)) == 1

	if !userValid || !passValid {
		w.Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Route based on method and path
	path := strings.TrimPrefix(r.URL.Path, "/admin/games/")
	parts := strings.Split(path, "/")
	
	if len(parts) >= 2 && parts[1] == "state" {
		h.GetGameState(w, r)
	} else if len(parts) >= 1 && parts[0] != "" {
		h.DeleteGame(w, r)
	} else {
		http.Error(w, "Invalid admin game endpoint", http.StatusBadRequest)
	}
}

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}