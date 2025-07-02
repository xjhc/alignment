// Package health provides server health monitoring functionality
package health

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/xjhc/alignment/server/internal/logger"
	"github.com/xjhc/alignment/server/internal/metrics"
)

// HealthStatus represents the current health status of the server
type HealthStatus string

const (
	// HealthyStatus indicates the server is operating normally
	HealthyStatus HealthStatus = "HEALTHY"
	// OverloadedStatus indicates the server is under heavy load
	OverloadedStatus HealthStatus = "OVERLOADED"
)

// Config holds configuration for the health monitor
type Config struct {
	// CheckInterval is how often to check server health
	CheckInterval time.Duration
	// MemoryThresholdMB is the memory usage threshold in MB that triggers overload status
	MemoryThresholdMB uint64
	// GCThresholdPercent is the GC CPU percentage threshold that triggers overload status
	GCThresholdPercent float64
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() Config {
	return Config{
		CheckInterval:      5 * time.Second,
		MemoryThresholdMB:  512, // 512MB threshold
		GCThresholdPercent: 30.0, // 30% GC overhead
	}
}

// Monitor monitors server health and maintains status
type Monitor struct {
	config    Config
	status    HealthStatus
	statusMux sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *slog.Logger
}

// NewMonitor creates a new health monitor with the given configuration
func NewMonitor(config Config) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Monitor{
		config: config,
		status: HealthyStatus,
		ctx:    ctx,
		cancel: cancel,
		logger: logger.WithField("component", "health_monitor"),
	}
}

// Start begins health monitoring in a background goroutine
func (m *Monitor) Start() {
	m.logger.Info("Starting health monitor", "check_interval", m.config.CheckInterval)
	
	go m.monitorLoop()
}

// Stop stops the health monitoring
func (m *Monitor) Stop() {
	m.logger.Info("Stopping health monitor")
	m.cancel()
}

// GetStatus returns the current health status
func (m *Monitor) GetStatus() HealthStatus {
	m.statusMux.RLock()
	defer m.statusMux.RUnlock()
	return m.status
}

// IsHealthy returns true if the server is healthy
func (m *Monitor) IsHealthy() bool {
	return m.GetStatus() == HealthyStatus
}

// setStatus updates the health status and metrics
func (m *Monitor) setStatus(newStatus HealthStatus) {
	m.statusMux.Lock()
	oldStatus := m.status
	m.status = newStatus
	m.statusMux.Unlock()

	// Update metrics
	switch newStatus {
	case HealthyStatus:
		metrics.SetServerHealthy()
	case OverloadedStatus:
		metrics.SetServerOverloaded()
	}

	// Log status changes
	if oldStatus != newStatus {
		m.logger.Info("Health status changed", 
			"old_status", oldStatus, 
			"new_status", newStatus)
	}
}

// monitorLoop is the main monitoring loop
func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkHealth()
		}
	}
}

// checkHealth performs a health check and updates the status
func (m *Monitor) checkHealth() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Convert bytes to MB
	currentMemoryMB := memStats.Alloc / 1024 / 1024
	
	// Update memory usage metric
	metrics.MemoryUsageBytes.Set(float64(memStats.Alloc))

	// Calculate GC overhead percentage
	gcOverhead := float64(memStats.GCCPUFraction) * 100

	m.logger.Debug("Health check", 
		"memory_mb", currentMemoryMB,
		"memory_threshold_mb", m.config.MemoryThresholdMB,
		"gc_overhead_percent", gcOverhead,
		"gc_threshold_percent", m.config.GCThresholdPercent)

	// Determine health status based on thresholds
	isOverloaded := currentMemoryMB > m.config.MemoryThresholdMB || 
					gcOverhead > m.config.GCThresholdPercent

	if isOverloaded {
		m.setStatus(OverloadedStatus)
		m.logger.Warn("Server is overloaded", 
			"memory_mb", currentMemoryMB,
			"gc_overhead_percent", gcOverhead)
	} else {
		m.setStatus(HealthyStatus)
	}
}