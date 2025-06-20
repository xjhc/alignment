package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alignment/server/internal/actors"
	"github.com/alignment/server/internal/comms"
	"github.com/alignment/server/internal/game"
	"github.com/alignment/server/internal/store"
	"github.com/google/uuid"
)

// Server represents the main application server
type Server struct {
	supervisor *actors.Supervisor
	wsManager  *comms.WebSocketManager
	datastore  *store.RedisDataStore
	scheduler  *game.Scheduler
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
	wsManager := comms.NewWebSocketManager(actionHandler)

	// Create supervisor
	supervisor := actors.NewSupervisor(datastore, wsManager)

	// Wire dependencies
	actionHandler.supervisor = supervisor

	// Set scheduler callback
	scheduler = game.NewScheduler(func(timer game.Timer) {
		if supervisor != nil {
			handleTimerExpired(supervisor, timer)
		}
	})

	server := &Server{
		supervisor: supervisor,
		wsManager:  wsManager,
		datastore:  datastore,
		scheduler:  scheduler,
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
	supervisor *actors.Supervisor
}

// HandleAction processes game actions from WebSocket clients
func (ah *ActionHandler) HandleAction(action game.Action) error {
	log.Printf("Handling action: %s from player %s for game %s", action.Type, action.PlayerID, action.GameID)

	switch action.Type {
	case game.ActionJoinGame:
		return ah.handleJoinGame(action)
	case game.ActionStartGame:
		return ah.handleStartGame(action)
	default:
		// Forward to game actor
		if actor, exists := ah.supervisor.GetActor(action.GameID); exists {
			actor.SendAction(action)
			return nil
		}
		return fmt.Errorf("game %s not found", action.GameID)
	}
}

func (ah *ActionHandler) handleJoinGame(action game.Action) error {
	gameID := action.GameID

	// Create game if it doesn't exist
	if _, exists := ah.supervisor.GetActor(gameID); !exists {
		err := ah.supervisor.CreateGame(gameID)
		if err != nil {
			return fmt.Errorf("failed to create game: %w", err)
		}
	}

	// Forward action to game actor
	if actor, exists := ah.supervisor.GetActor(gameID); exists {
		actor.SendAction(action)
		return nil
	}

	return fmt.Errorf("failed to get game actor after creation")
}

func (ah *ActionHandler) handleStartGame(action game.Action) error {
	if actor, exists := ah.supervisor.GetActor(action.GameID); exists {
		actor.SendAction(action)
		return nil
	}
	return fmt.Errorf("game %s not found", action.GameID)
}

// HTTP handlers
func (s *Server) setupRoutes() {
	http.HandleFunc("/health", s.healthHandler)
	http.HandleFunc("/ws", s.wsManager.HandleWebSocket)
	http.HandleFunc("/api/games", s.gamesHandler)
	http.HandleFunc("/api/games/create", s.createGameHandler)
	http.HandleFunc("/api/stats", s.statsHandler)
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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// List active games
	games, err := s.datastore.ListActiveGames()
	if err != nil {
		http.Error(w, "Failed to list games", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"games": games,
	})
}

func (s *Server) createGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate new game ID
	gameID := uuid.New().String()

	// Create game actor
	err := s.supervisor.CreateGame(gameID)
	if err != nil {
		http.Error(w, "Failed to create game", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"game_id": gameID,
		"status":  "created",
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
	action := game.Action{
		Type:     timer.Action.Type,
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
	fmt.Println("  GET  /api/games")
	fmt.Println("  POST /api/games/create")
	fmt.Println("  GET  /api/stats")

	log.Printf("Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
