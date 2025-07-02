// Package metrics provides Prometheus metrics for the Alignment game server
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Game-related metrics
	GamesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "alignment_games_active",
		Help: "The current number of active games",
	})

	LobbiesActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "alignment_lobbies_active", 
		Help: "The current number of active lobbies",
	})

	PlayersConnected = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "alignment_players_connected",
		Help: "The current number of connected players",
	})

	GamesCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "alignment_games_created_total",
		Help: "The total number of games created",
	})

	PlayersJoinedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "alignment_players_joined_total",
		Help: "The total number of player joins",
	})

	// Action processing metrics
	ActionProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alignment_action_processing_duration_seconds",
		Help:    "Time taken to process game actions",
		Buckets: prometheus.DefBuckets,
	}, []string{"action_type"})

	ActionsProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alignment_actions_processed_total",
		Help: "The total number of actions processed",
	}, []string{"action_type", "status"})

	// WebSocket metrics
	WebSocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "alignment_websocket_connections",
		Help: "The current number of WebSocket connections",
	})

	WebSocketMessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alignment_websocket_messages_total",
		Help: "The total number of WebSocket messages",
	}, []string{"direction"}) // "inbound" or "outbound"

	// Server health metrics
	ServerHealthStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "alignment_server_health_status",
		Help: "Server health status (1=healthy, 0=unhealthy)",
	}, []string{"status"})

	MemoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "alignment_memory_usage_bytes",
		Help: "Current memory usage in bytes",
	})

	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alignment_http_requests_total",
		Help: "The total number of HTTP requests",
	}, []string{"method", "endpoint", "status_code"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alignment_http_request_duration_seconds",
		Help:    "Time taken to process HTTP requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "endpoint"})
)

// RecordGameCreated increments the games created counter and active games gauge
func RecordGameCreated() {
	GamesCreatedTotal.Inc()
	GamesActive.Inc()
}

// RecordGameEnded decrements the active games gauge
func RecordGameEnded() {
	GamesActive.Dec()
}

// RecordLobbyCreated increments the active lobbies gauge
func RecordLobbyCreated() {
	LobbiesActive.Inc()
}

// RecordLobbyDestroyed decrements the active lobbies gauge
func RecordLobbyDestroyed() {
	LobbiesActive.Dec()
}

// RecordPlayerJoined increments the players joined counter and connected players gauge
func RecordPlayerJoined() {
	PlayersJoinedTotal.Inc()
	PlayersConnected.Inc()
}

// RecordPlayerLeft decrements the connected players gauge
func RecordPlayerLeft() {
	PlayersConnected.Dec()
}

// RecordWebSocketConnect increments the WebSocket connections gauge
func RecordWebSocketConnect() {
	WebSocketConnections.Inc()
}

// RecordWebSocketDisconnect decrements the WebSocket connections gauge
func RecordWebSocketDisconnect() {
	WebSocketConnections.Dec()
}

// SetServerHealthy sets the server health status to healthy
func SetServerHealthy() {
	ServerHealthStatus.WithLabelValues("healthy").Set(1)
	ServerHealthStatus.WithLabelValues("overloaded").Set(0)
}

// SetServerOverloaded sets the server health status to overloaded
func SetServerOverloaded() {
	ServerHealthStatus.WithLabelValues("healthy").Set(0)
	ServerHealthStatus.WithLabelValues("overloaded").Set(1)
}