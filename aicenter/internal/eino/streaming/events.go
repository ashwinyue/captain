package streaming

import (
	"time"
)

// EventType represents the type of streaming event
type EventType string

const (
	EventTypeMessage    EventType = "message"
	EventTypeToolCall   EventType = "tool_call"
	EventTypeToolResult EventType = "tool_result"
	EventTypeTransfer   EventType = "transfer"
	EventTypeError      EventType = "error"
	EventTypeComplete   EventType = "complete"
	EventTypePing       EventType = "ping"
)

// Event represents a streaming event
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	AgentName string                 `json:"agent_name,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewMessageEvent creates a new message event
func NewMessageEvent(agentName, content string) Event {
	return Event{
		Type:      EventTypeMessage,
		Timestamp: time.Now(),
		AgentName: agentName,
		Content:   content,
	}
}

// NewToolCallEvent creates a new tool call event
func NewToolCallEvent(agentName, toolName string, args map[string]interface{}) Event {
	return Event{
		Type:      EventTypeToolCall,
		Timestamp: time.Now(),
		AgentName: agentName,
		Data: map[string]interface{}{
			"tool_name": toolName,
			"arguments": args,
		},
	}
}

// NewToolResultEvent creates a new tool result event
func NewToolResultEvent(agentName, toolName, result string) Event {
	return Event{
		Type:      EventTypeToolResult,
		Timestamp: time.Now(),
		AgentName: agentName,
		Data: map[string]interface{}{
			"tool_name": toolName,
			"result":    result,
		},
	}
}

// NewTransferEvent creates a new transfer event
func NewTransferEvent(fromAgent, toAgent string) Event {
	return Event{
		Type:      EventTypeTransfer,
		Timestamp: time.Now(),
		AgentName: fromAgent,
		Data: map[string]interface{}{
			"to_agent": toAgent,
		},
	}
}

// NewErrorEvent creates a new error event
func NewErrorEvent(agentName, err string) Event {
	return Event{
		Type:      EventTypeError,
		Timestamp: time.Now(),
		AgentName: agentName,
		Error:     err,
	}
}

// NewCompleteEvent creates a complete event
func NewCompleteEvent(agentName string) Event {
	return Event{
		Type:      EventTypeComplete,
		Timestamp: time.Now(),
		AgentName: agentName,
	}
}

// NewPingEvent creates a ping event for keep-alive
func NewPingEvent() Event {
	return Event{
		Type:      EventTypePing,
		Timestamp: time.Now(),
	}
}
