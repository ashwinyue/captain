package memory

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/cloudwego/eino/schema"
)

// Store persists and restores conversation history.
type Store interface {
	// Write stores messages for a session
	Write(ctx context.Context, sessionID string, msgs []*schema.Message) error
	// Read returns messages for a session
	Read(ctx context.Context, sessionID string) ([]*schema.Message, error)
	// Append adds messages to a session
	Append(ctx context.Context, sessionID string, msgs ...*schema.Message) error
	// Delete removes a session's messages
	Delete(ctx context.Context, sessionID string) error
}

// EncodeMessages serializes messages using Gob
func EncodeMessages(msgs []*schema.Message) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(msgs); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeMessages deserializes messages
func DecodeMessages(b []byte) ([]*schema.Message, error) {
	if len(b) == 0 {
		return nil, nil
	}
	dec := gob.NewDecoder(bytes.NewReader(b))
	var msgs []*schema.Message
	if err := dec.Decode(&msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}
