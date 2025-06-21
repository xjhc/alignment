package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/actors"
	"github.com/xjhc/alignment/server/internal/comms"
	"github.com/xjhc/alignment/server/internal/game"
	"github.com/xjhc/alignment/server/internal/lobby"
	"github.com/xjhc/alignment/server/internal/store"
)

// Server represents the main application server
type Server struct {
	supervisor   *actors.Supervisor
	wsManager    *comms.WebSocketManager
	datastore    *store.RedisDataStore
	scheduler    *game.Scheduler
	lobbyManager *lobby.LobbyManager
}

// NewServer creates a new server instance
func NewServer() (*Server, error) {
	// Initialize Redis datastore
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0

	datastore, err := store.NewRedisDataStore(redisAddr, redisPassword, redisDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create datastore: %w", err)
	}

	// Create scheduler
	scheduler := game.NewScheduler(nil) // Timer callback will be set later

	// Create WebSocket manager with action handler
	actionHandler := &ActionHandler{}
	wsManager := comms.NewWebSocketManager(actionHandler, nil) // Will set token validator later

	// Create supervisor with application context
	ctx := context.Background()
	supervisor := actors.NewSupervisor(ctx, datastore, wsManager)

	// Create lobby manager (needs supervisor and wsManager as dependencies)
	lobbyManager := lobby.NewLobbyManager(wsManager, supervisor)

	// Set token validator on WebSocket manager
	wsManager.SetTokenValidator(lobbyManager)

	// Wire dependencies
	actionHandler.supervisor = supervisor
	actionHandler.lobbyManager = lobbyManager

	// Set scheduler callback
	scheduler = game.NewScheduler(func(timer game.Timer) {
		if supervisor != nil {
			handleTimerExpired(supervisor, timer)
		}
	})

	server := &Server{
		supervisor:   supervisor,
		wsManager:    wsManager,
		datastore:    datastore,
		scheduler:    scheduler,
		lobbyManager: lobbyManager,
	}

	return server, nil
}

// Start starts all server components
func (s *Server) Start() {
	log.Println("Starting Alignment game server...")

	// Start components
	s.scheduler.Start()
	s.supervisor.Start()
	s.wsManager.Start()

	log.Println("All components started successfully")
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	log.Println("Shutting down server...")

	s.scheduler.Stop()
	s.supervisor.Stop()
	s.datastore.Close()

	log.Println("Server stopped")
}

// ActionHandler implements the ActionHandler interface for WebSocket manager
type ActionHandler struct {
	supervisor   *actors.Supervisor
	lobbyManager *lobby.LobbyManager
}

// HandleAction processes game actions from WebSocket clients
func (ah *ActionHandler) HandleAction(action core.Action) error {
	log.Printf("Handling action: %s from player %s for game %s", action.Type, action.PlayerID, action.GameID)

	// Validate that action has a game ID
	if action.GameID == "" {
		return fmt.Errorf("action of type %s is missing game_id", action.Type)
	}

	switch action.Type {
	case core.ActionJoinGame:
		return ah.handleJoinGame(action)
	case core.ActionStartGame:
		return ah.handleStartGame(action)
	default:
		// Forward all other actions to the correct game actor
		if actor, exists := ah.supervisor.GetActor(action.GameID); exists {
			actor.SendAction(action)
			return nil
		}
		// Also check lobby actors for any other lobby-specific actions
		if lobbyActor, exists := ah.lobbyManager.GetLobbyActor(action.GameID); exists {
			lobbyActor.SendAction(action)
			return nil
		}
		return fmt.Errorf("game or lobby %s not found", action.GameID)
	}
}

func (ah *ActionHandler) handleJoinGame(action core.Action) error {
	// Joining is handled by the LobbyActor
	if lobbyActor, exists := ah.lobbyManager.GetLobbyActor(action.GameID); exists {
		lobbyActor.SendAction(action)
		return nil
	}

	// Or a reconnect to an existing game
	if gameActor, exists := ah.supervisor.GetActor(action.GameID); exists {
		gameActor.SendAction(action)
		return nil
	}

	return fmt.Errorf("game or lobby %s not found for join action", action.GameID)
}

func (ah *ActionHandler) handleStartGame(action core.Action) error {
	// Starting is handled by the LobbyActor
	if lobbyActor, exists := ah.lobbyManager.GetLobbyActor(action.GameID); exists {
		lobbyActor.SendAction(action)
		return nil
	}

	return fmt.Errorf("lobby %s not found to start game", action.GameID)
}

// HTTP handlers
func (s *Server) setupRoutes() {
	http.HandleFunc("/health", s.healthHandler)
	http.HandleFunc("/ws", s.wsManager.HandleWebSocket)
	http.HandleFunc("/api/games/", s.gameByIDHandler) // Handle paths with trailing slash
	http.HandleFunc("/api/games", s.gamesHandler)     // Handle exact match
	http.HandleFunc("/api/stats", s.statsHandler)
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Test endpoint works"))
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status": "healthy",
		"stats":  s.supervisor.GetStats(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) gamesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listLobbies(w, r)
	case http.MethodPost:
		s.createLobby(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listLobbies(w http.ResponseWriter, r *http.Request) {
	lobbies := s.lobbyManager.GetLobbyList()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lobbies": lobbies,
	})
}

func (s *Server) createLobby(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LobbyName    string `json:"lobby_name"`
		PlayerName   string `json:"player_name"`
		PlayerAvatar string `json:"player_avatar"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.LobbyName == "" {
		req.LobbyName = fmt.Sprintf("%s's Game", req.PlayerName)
	}

	if req.PlayerName == "" {
		http.Error(w, "Player name is required", http.StatusBadRequest)
		return
	}

	gameID, hostPlayerID, sessionToken, err := s.lobbyManager.CreateLobby(req.PlayerName, req.PlayerAvatar, req.LobbyName)
	if err != nil {
		http.Error(w, "Failed to create lobby", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"game_id":       gameID,
		"player_id":     hostPlayerID,
		"session_token": sessionToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) gameByIDHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("gameByIDHandler called with path: %s", r.URL.Path)

	// Extract game ID from URL path
	path := r.URL.Path
	if len(path) < 11 { // "/api/games/"
		log.Printf("Path too short: %d", len(path))
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	parts := strings.Split(path[11:], "/") // Remove "/api/games/" prefix
	log.Printf("Path parts: %v", parts)

	if len(parts) < 1 || parts[0] == "" {
		log.Printf("Invalid parts: %v", parts)
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	gameID := parts[0]
	log.Printf("Extracted gameID: %s", gameID)

	// Handle different sub-endpoints
	if len(parts) > 1 && parts[1] == "join" {
		log.Printf("Join endpoint requested for game: %s", gameID)
		if r.Method == http.MethodPost {
			s.joinLobby(w, r, gameID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	log.Printf("No matching endpoint for path: %s", path)
	http.Error(w, "Endpoint not found", http.StatusNotFound)
}

func (s *Server) joinLobby(w http.ResponseWriter, r *http.Request, gameID string) {
	var req struct {
		PlayerName   string `json:"player_name"`
		PlayerAvatar string `json:"player_avatar"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PlayerName == "" {
		http.Error(w, "Player name is required", http.StatusBadRequest)
		return
	}

	playerID, sessionToken, err := s.lobbyManager.JoinLobby(gameID, req.PlayerName, req.PlayerAvatar)
	if err != nil {
		switch err.Error() {
		case "lobby not found":
			http.Error(w, "Lobby not found", http.StatusNotFound)
		case "lobby is not accepting new players":
			http.Error(w, "Lobby is not accepting new players", http.StatusConflict)
		case "lobby is full":
			http.Error(w, "Lobby is full", http.StatusConflict)
		default:
			http.Error(w, "Failed to join lobby", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"game_id":       gameID,
		"player_id":     playerID,
		"session_token": sessionToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"supervisor": s.supervisor.GetStats(),
		"scheduler":  len(s.scheduler.GetActiveTimers()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleTimerExpired processes expired timers
func handleTimerExpired(supervisor *actors.Supervisor, timer game.Timer) {
	log.Printf("Timer expired: %s for game %s", timer.ID, timer.GameID)

	// Convert timer action to game action
	action := core.Action{
		Type:     core.ActionType(timer.Action.Type),
		GameID:   timer.GameID,
		PlayerID: "SYSTEM",
		Payload:  timer.Action.Payload,
	}

	// Send to game actor
	if actor, exists := supervisor.GetActor(timer.GameID); exists {
		actor.SendAction(action)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup routes
	server.setupRoutes()

	// Start server components
	server.Start()

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Received shutdown signal")
		server.Stop()
		os.Exit(0)
	}()

	fmt.Printf("Alignment server starting on port %s\n", port)
	fmt.Println("WebSocket endpoint: /ws")
	fmt.Println("API endpoints:")
	fmt.Println("  GET  /health")
	fmt.Println("  GET  /api/games               - List lobbies")
	fmt.Println("  POST /api/games              - Create lobby")
	fmt.Println("  POST /api/games/{gameId}/join - Join lobby")
	fmt.Println("  GET  /api/stats")

	log.Printf("Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
