package model

const (
	KEY_SOURCE = "msgsrc"
	KEY_SEQ    = "msgseq"
)

type Messager interface {
	// ID is an increasing number for identifing a message
	ID() int64
	WithID(int64) Messager
	// Full topic string of the message
	// need this because we may use wildcard to filter messages.
	Topic() string
	// Data is Real data in bytes
	Data() []byte
	// Metadata keeps K:V description of the data
	Metadata() map[string]string
	// Source is an id represent producer
	Source() string
	// Seq is the serial number of this payload from producer
	Seq() int64
	// Dest is an id of a consumer
	Dest() string
	WithDest(string) Messager
	// Ack is for responding to message producer
	Ack(data []byte) error
}
