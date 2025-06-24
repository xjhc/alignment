package events

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestEvent struct {
	Data string
}

func (e TestEvent) EventType() string {
	return "test_event"
}

type AnotherTestEvent struct {
	Value int
}

func (e AnotherTestEvent) EventType() string {
	return "another_test_event"
}

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	receivedEvent := false

	// Create a subscriber channel
	ch := make(chan Event, 1)
	bus.Subscribe("test_event", ch)

	// Listen in a goroutine
	go func() {
		defer wg.Done()
		select {
		case event := <-ch:
			testEvent, ok := event.(TestEvent)
			assert.True(t, ok)
			assert.Equal(t, "hello", testEvent.Data)
			receivedEvent = true
		case <-time.After(1 * time.Second):
			t.Error("Timed out waiting for event")
		}
	}()

	// Publish an event
	bus.Publish(TestEvent{Data: "hello"})

	wg.Wait()
	assert.True(t, receivedEvent, "Subscriber did not receive the event")
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	receivedCount := 0
	var mu sync.Mutex

	// Create two subscriber channels
	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)
	bus.Subscribe("test_event", ch1)
	bus.Subscribe("test_event", ch2)

	// Listen on both channels
	for _, ch := range []chan Event{ch1, ch2} {
		go func(subscriber chan Event) {
			defer wg.Done()
			select {
			case event := <-subscriber:
				testEvent, ok := event.(TestEvent)
				assert.True(t, ok)
				assert.Equal(t, "broadcast", testEvent.Data)
				mu.Lock()
				receivedCount++
				mu.Unlock()
			case <-time.After(1 * time.Second):
				t.Error("Timed out waiting for event")
			}
		}(ch)
	}

	// Publish an event
	bus.Publish(TestEvent{Data: "broadcast"})

	wg.Wait()
	assert.Equal(t, 2, receivedCount, "Both subscribers should receive the event")
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	ch := make(chan Event, 1)
	bus.Subscribe("test_event", ch)
	bus.Unsubscribe("test_event", ch)

	// Publish an event
	bus.Publish(TestEvent{Data: "should_not_receive"})

	// Give it a moment to potentially deliver
	time.Sleep(50 * time.Millisecond)

	select {
	case <-ch:
		t.Error("Should not have received event after unsubscribing")
	default:
		// Expected behavior - no event received
	}
}

func TestEventBus_DifferentEventTypes(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)
	
	bus.Subscribe("test_event", ch1)
	bus.Subscribe("another_test_event", ch2)

	// Publish different event types
	bus.Publish(TestEvent{Data: "first"})
	bus.Publish(AnotherTestEvent{Value: 42})

	// Check first channel receives TestEvent
	select {
	case event := <-ch1:
		testEvent, ok := event.(TestEvent)
		assert.True(t, ok)
		assert.Equal(t, "first", testEvent.Data)
	case <-time.After(100 * time.Millisecond):
		t.Error("First channel should receive TestEvent")
	}

	// Check second channel receives AnotherTestEvent
	select {
	case event := <-ch2:
		anotherEvent, ok := event.(AnotherTestEvent)
		assert.True(t, ok)
		assert.Equal(t, 42, anotherEvent.Value)
	case <-time.After(100 * time.Millisecond):
		t.Error("Second channel should receive AnotherTestEvent")
	}
}

func TestEventBus_Close(t *testing.T) {
	bus := NewEventBus()
	
	ch := make(chan Event, 1)
	bus.Subscribe("test_event", ch)
	
	bus.Close()
	
	// Channel should be closed
	_, ok := <-ch
	assert.False(t, ok, "Channel should be closed after bus.Close()")
}