package helpers

import (
	"context"
	"testing"
	"time"
)

func TestGoSafe_DoesNotPanicOnNormalExecution(t *testing.T) {
	ctx := context.Background()
	executed := false

	// This should not panic
	GoSafe(ctx, func(ctx context.Context) {
		executed = true
	})

	// Give the goroutine time to execute
	time.Sleep(100 * time.Millisecond)

	if !executed {
		t.Error("Function should have been executed")
	}
}

func TestGoSafe_RecoversPanicAndLogsError(t *testing.T) {
	ctx := context.Background()
	
	// Create a test that will panic
	executed := false
	GoSafe(ctx, func(ctx context.Context) {
		executed = true
		panic("test panic")
	})

	// Give the goroutine time to execute and recover
	time.Sleep(100 * time.Millisecond)

	if !executed {
		t.Error("Function should have been executed before panic")
	}
	// Note: We can't easily test the logging without dependency injection
	// In a real scenario, we would inject the logger as a dependency
}

func TestGoSafe_RespectsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	executed := false
	GoSafe(ctx, func(ctx context.Context) {
		select {
		case <-ctx.Done():
			// Context was cancelled
			return
		default:
			executed = true
		}
	})

	// Give the goroutine time to check context
	time.Sleep(100 * time.Millisecond)

	if executed {
		t.Error("Function should not execute when context is cancelled")
	}
}

func TestGoSafe_MultipleGoroutinesDoNotCrashEachOther(t *testing.T) {
	ctx := context.Background()
	
	normalExecuted := false
	panicExecuted := false
	
	// Start a normal goroutine
	GoSafe(ctx, func(ctx context.Context) {
		time.Sleep(50 * time.Millisecond)
		normalExecuted = true
	})
	
	// Start a panicking goroutine
	GoSafe(ctx, func(ctx context.Context) {
		panicExecuted = true
		panic("test panic")
	})
	
	// Give both goroutines time to execute
	time.Sleep(200 * time.Millisecond)
	
	if !normalExecuted {
		t.Error("Normal goroutine should have executed")
	}
	if !panicExecuted {
		t.Error("Panicking goroutine should have executed before panic")
	}
}

// Integration test that simulates a timer callback panic
func TestGoSafe_TimerCallbackPanicRecovery(t *testing.T) {
	ctx := context.Background()
	
	callbackExecuted := false
	
	// Simulate a timer callback that panics
	timerCallback := func() {
		callbackExecuted = true
		panic("timer callback panic")
	}
	
	// Execute the callback in a GoSafe goroutine (simulating Scheduler behavior)
	GoSafe(ctx, func(_ context.Context) {
		timerCallback()
	})
	
	// Give the goroutine time to execute
	time.Sleep(100 * time.Millisecond)
	
	if !callbackExecuted {
		t.Error("Timer callback should have been executed")
	}
	// If we reach here, it means GoSafe successfully recovered the panic
	// and the test process didn't crash, which is exactly what we want
}