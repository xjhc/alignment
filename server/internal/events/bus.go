package events

import (
	"log"
	"sync"
)

// Event represents a generic event that can be published to the bus
type Event interface {
	EventType() string
}

// EventBus provides a simple in-memory event publishing and subscription system
type EventBus struct {
	subscribers map[string][]chan Event
	mu          sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

// Subscribe registers a channel to receive events of a specific type
func (eb *EventBus) Subscribe(eventType string, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	log.Printf("EventBus: Subscribed to event type '%s'", eventType)
}

// Unsubscribe removes a channel from receiving events of a specific type
func (eb *EventBus) Unsubscribe(eventType string, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	subs := eb.subscribers[eventType]
	for i, subscriber := range subs {
		if subscriber == ch {
			eb.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

// Publish sends an event to all subscribers of that event type
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	subscribers := eb.subscribers[event.EventType()]
	eb.mu.RUnlock()
	
	// Publish to all subscribers in separate goroutines to prevent blocking
	for _, ch := range subscribers {
		go func(subscriber chan Event) {
			select {
			case subscriber <- event:
				// Event delivered successfully
			default:
				// Channel is full or closed, skip this subscriber
				log.Printf("EventBus: Failed to deliver event %s to subscriber (channel full/closed)", event.EventType())
			}
		}(ch)
	}
	
	log.Printf("EventBus: Published event %s to %d subscribers", event.EventType(), len(subscribers))
}

// Close shuts down the event bus
func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	// Close all subscriber channels
	for eventType, subs := range eb.subscribers {
		for _, ch := range subs {
			close(ch)
		}
		delete(eb.subscribers, eventType)
	}
	
	log.Println("EventBus: Closed")
}