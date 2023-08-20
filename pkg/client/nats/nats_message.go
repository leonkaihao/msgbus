package nats

import (
	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"

	nats "github.com/nats-io/nats.go"
)

type messager struct {
	*common.MessagerBase
	msg *nats.Msg
}

// NewMessager return wrapped message and
// a flag indicate if the message comes from legacy producer(false)
func NewMessager(id int64, msg *nats.Msg) (model.Messager, error) {
	mb, err := common.NewMessagerBase(id, msg.Subject, msg.Data)
	if err != nil {
		return nil, err
	}
	msgr := &messager{msg: msg, MessagerBase: mb}
	return msgr, nil
}

func (msgr *messager) Ack(data []byte) error {
	return msgr.msg.Respond(data)
}
