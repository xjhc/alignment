package mocks

import (
	"sync"
	"github.com/xjhc/alignment/server/internal/events"
)

// MockEventBus is a mock implementation that implements EventBus interface while tracking calls
type MockEventBus struct {
	realEventBus *events.EventBus // Composition instead of embedding
	sync.Mutex

	PublishCalls     []MockPublishCall
	SubscribeCalls   []MockSubscribeCall
	UnsubscribeCalls []MockUnsubscribeCall
	StopCalls        []MockStopCall
}

type MockPublishCall struct {
	Event events.Event
}

type MockSubscribeCall struct {
	EventType string
	Channel   chan events.Event
}

type MockUnsubscribeCall struct {
	EventType string
	Channel   chan events.Event
}

type MockStopCall struct{}

func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		realEventBus:     events.NewEventBus(),
		PublishCalls:     make([]MockPublishCall, 0),
		SubscribeCalls:   make([]MockSubscribeCall, 0),
		UnsubscribeCalls: make([]MockUnsubscribeCall, 0),
		StopCalls:        make([]MockStopCall, 0),
	}
}

// Publish tracks the call and delegates to real implementation
func (m *MockEventBus) Publish(event events.Event) {
	m.Lock()
	m.PublishCalls = append(m.PublishCalls, MockPublishCall{
		Event: event,
	})
	m.Unlock()

	// Call the real implementation
	m.realEventBus.Publish(event)
}

// Subscribe tracks the call and delegates to real implementation
func (m *MockEventBus) Subscribe(eventType string, channel chan events.Event) {
	m.Lock()
	m.SubscribeCalls = append(m.SubscribeCalls, MockSubscribeCall{
		EventType: eventType,
		Channel:   channel,
	})
	m.Unlock()

	// Call the real implementation
	m.realEventBus.Subscribe(eventType, channel)
}

// Unsubscribe tracks the call and delegates to real implementation
func (m *MockEventBus) Unsubscribe(eventType string, channel chan events.Event) {
	m.Lock()
	m.UnsubscribeCalls = append(m.UnsubscribeCalls, MockUnsubscribeCall{
		EventType: eventType,
		Channel:   channel,
	})
	m.Unlock()

	// Call the real implementation
	m.realEventBus.Unsubscribe(eventType, channel)
}

// Close delegates to the real implementation and tracks the call
func (m *MockEventBus) Close() {
	m.Lock()
	m.StopCalls = append(m.StopCalls, MockStopCall{})
	m.Unlock()

	// Call the real implementation
	m.realEventBus.Close()
}