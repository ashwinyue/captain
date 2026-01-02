package streaming

import "errors"

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionClosed   = errors.New("session closed")
	ErrSubscriberFull  = errors.New("subscriber channel full")
)
