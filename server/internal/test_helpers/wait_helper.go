
package test_helpers

import (
	"sync"
	"testing"
	"time"
)

// WaitWithTimeout is a helper function to wait for a WaitGroup with a timeout.
// It fails the test if the timeout is exceeded.
func WaitWithTimeout(wg *sync.WaitGroup, timeout time.Duration, t *testing.T) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	
	select {
	case <-c:
		// WaitGroup finished successfully
	case <-time.After(timeout):
		t.Fatalf("Test timed out after %v waiting for WaitGroup", timeout)
	}
}

// WaitWithTimeoutNoFatal is like WaitWithTimeout but returns a boolean instead of calling t.Fatalf
func WaitWithTimeoutNoFatal(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return true // Completed successfully
	case <-time.After(timeout):
		return false // Timed out
	}
}