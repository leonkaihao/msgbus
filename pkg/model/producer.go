package model

import "context"

const (
	PRD_OPTION_PAYLOAD = "payload"
)

type Producer interface {
	// ID is the id of this producer
	ID() string
	// Name is assigned by user for categorizing
	Name() string
	// CurrentSeq is the current sequence number of sending payload
	CurrentSeq() int64
	// Topic return the topic string that bind to the producer
	Topic() string
	// Fire is a fire-and-forget mode publishing, and does not guarantee that consumer receives the message.
	Fire(data []byte, metadata map[string]string) error
	// Request is a request-and-response mode publishing, it waits consumer's response and finish.
	// If a consumer got killed, the method may get blocked because no response data is collected and return.
	// Define the context with timeout can avoid block.
	Request(context context.Context, data []byte, metadata map[string]string) ([]byte, error)
	// PayloadType has const definitions(PLTYPE_...) in payload.go
	PayloadType() string
	// SetOption set options like payload type
	SetOption(k, v string) error
	// Close releases the resources owned by Producer
	Close() error
}
