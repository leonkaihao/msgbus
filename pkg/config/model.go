package config

const (
	PUB_MODE_FIRE = "fire"
	PUB_MODE_REQ  = "request"

	SHOW_METADATA = "metadata"
	SHOW_DATA     = "data"
	SHOW_TOPIC    = "topic"
	SHOW_ALL      = "all"
)

type Event struct {
	Topic    string                 `json:"topic"`
	Mode     string                 `json:"mode"`    // publish mode: fire or request
	Payload  string                 `json:"payload"` // msgbus or edgebus
	Delay    uint                   `json:"delay"`   // wait before sending. millisec
	Sleep    uint                   `json:"sleep"`   // wait after sending. millisec
	Repeat   int                    `json:"repeat"`  // times.
	Data     map[string]interface{} `json:"data"`
	Metadata map[string]string      `json:"metadata"`
}

type Producer struct {
	Endpoint string  `json:"endpoint"`
	Events   []Event `json:"events"`
}
type Topic struct {
	Topic string `json:"topic"`
	Group string `json:"group"`
}
type Consumer struct {
	Endpoint string   `json:"endpoint"`
	Topics   []*Topic `json:"topics"`
	Show     string   `json:"show"` //default value equal to SHOW_ALL
}

type Config struct {
	DeviceId  string      `json:"device_id"`
	Producer  *Producer   `json:"producer"`
	Consumers []*Consumer `json:"consumers"`
}
