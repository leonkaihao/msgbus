package inproc

import (
	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"
)

type messager struct {
	*common.MessagerBase
}

// NewMessager return wrapped message and
// a flag indicate if the message comes from legacy producer(false)
func NewMessager(id int64, topic string, data []byte) (model.Messager, error) {
	mb, err := common.NewMessagerBase(id, topic, data)
	if err != nil {
		return nil, err
	}
	msgr := &messager{MessagerBase: mb}
	return msgr, nil
}

func (msgr *messager) Ack(data []byte) error {
	return nil
}
