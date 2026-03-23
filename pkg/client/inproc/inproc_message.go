package inproc

import (
	"context"
	"fmt"
	"time"

	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"
)

type messager struct {
	*common.MessagerBase
	respChan chan []byte
}

// NewMessager return wrapped message and
// a flag indicate if the message comes from legacy producer(false)
func NewMessager(id int64, topic string, data []byte) (model.Messager, error) {
	mb, err := common.NewMessagerBase(id, topic, data)
	if err != nil {
		return nil, err
	}
	msgr := &messager{MessagerBase: mb, respChan: make(chan []byte)}
	return msgr, nil
}

func (msgr *messager) Ack(data []byte) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to ack to inproc message, %v", r)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// we don't care if there is any receiver for the ack
	// but after ACK by any msg consumer, the channel in msgr should be closed to avoid accumulated acks.
	select {
	case msgr.respChan <- data:
	case <-ctx.Done():
	}
	close(msgr.respChan)
	return
}
