package model

const (
	CONSUMER_ACK string = "ACK"
)

type Consumer interface {
	// ID is the id of this consumer
	ID() string
	// Name is assigned by user for categorizing
	Name() string
	// Sub return the subscribe string
	Sub() string
	// Group return the conusmer group string, used by Subscribe
	Group() string
	// Subscribe the topic within the consumer group
	Subscribe() (<-chan Messager, error)
	// Close release the resource of its own
	Close() error
}
