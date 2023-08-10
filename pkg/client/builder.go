package client

import (
	"errors"

	"github.com/leonkaihao/msgbus/pkg/client/nats"
	"github.com/leonkaihao/msgbus/pkg/model"
)

const (
	CLI_NATS   = "nats"
	CLI_MQTT3  = "mqtt3"
	CLI_INPROC = "inproc"
)

type Builder interface {
	Build() (model.Client, error)
}

type builder struct {
	clientName string
}

func NewBuilder(clientName string) Builder {
	return &builder{clientName}
}

func (b *builder) Build() (model.Client, error) {
	switch b.clientName {
	case CLI_NATS:
		return nats.NewClient(), nil
	default:
		return nil, errors.New("No client named " + b.clientName)
	}
}
