package orchestrator

import (
	"sync"
	"time"

	"kuorum/backend/internal/domain/pipeline"
)

type EventType string

const (
	EventProgress       EventType = "pipeline_progress"
	EventStageCompleted EventType = "stage_completed"
	EventAwaitingInput  EventType = "awaiting_input"
	EventCompleted      EventType = "pipeline_completed"
	EventFailed         EventType = "pipeline_failed"
)

// PipelineEvent represents a real-time event broadcasted to SSE subscribers.
type PipelineEvent struct {
	PipelineID string                  `json:"pipelineId"`
	Type       EventType               `json:"type"`
	Stage      pipeline.Stage          `json:"stage"`
	Message    string                  `json:"message"`
	Progress   int                     `json:"progress"`
	Checkpoint pipeline.CheckpointType `json:"checkpoint,omitempty"`
	Data       any                     `json:"data,omitempty"`
	Step       *pipeline.ThinkingStep  `json:"step,omitempty"`
	Stats      *pipeline.StatsSummary  `json:"stats,omitempty"`
	Timestamp  time.Time               `json:"timestamp"`
}

// EventBroadcaster manages active client SSE subscriber channels.
type EventBroadcaster struct {
	mu          sync.RWMutex
	subscribers map[string][]chan PipelineEvent
	history     map[string][]PipelineEvent
}

func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		subscribers: make(map[string][]chan PipelineEvent),
		history:     make(map[string][]PipelineEvent),
	}
}

// Subscribe registers a new subscriber channel for a pipeline ID and replays past events.
func (b *EventBroadcaster) Subscribe(pipelineID string) (<-chan PipelineEvent, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan PipelineEvent, 50)

	// Replay history if present
	if pastEvents, ok := b.history[pipelineID]; ok {
		for _, ev := range pastEvents {
			select {
			case ch <- ev:
			default:
			}
		}
	}

	b.subscribers[pipelineID] = append(b.subscribers[pipelineID], ch)

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		subs := b.subscribers[pipelineID]
		for i, sub := range subs {
			if sub == ch {
				b.subscribers[pipelineID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
		if len(b.subscribers[pipelineID]) == 0 {
			delete(b.subscribers, pipelineID)
		}
	}

	return ch, unsubscribe
}

// Broadcast distributes an event to all subscribers of the specified pipeline and buffers in history.
func (b *EventBroadcaster) Broadcast(event PipelineEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Append to history buffer (cap at 50 events)
	hist := b.history[event.PipelineID]
	if len(hist) >= 50 {
		hist = hist[1:]
	}
	b.history[event.PipelineID] = append(hist, event)

	subs, ok := b.subscribers[event.PipelineID]
	if !ok {
		return
	}

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// Non-blocking drop if client buffer is saturated
		}
	}
}

// ClearHistory removes cached events for a pipeline.
func (b *EventBroadcaster) ClearHistory(pipelineID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.history, pipelineID)
}
