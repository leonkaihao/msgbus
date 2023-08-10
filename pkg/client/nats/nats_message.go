package nats

import (
	"fmt"
	"strconv"

	"github.com/leonkaihao/msgbus/pkg/model"

	nats "github.com/nats-io/nats.go"
)

type messager struct {
	id       int64
	topic    string
	msg      *nats.Msg
	data     []byte
	metadata map[string]string
	dest     string
}

// NewMessager return wrapped message and
// a flag indicate if the message comes from legacy producer(false)
func NewMessager(id int64, msg *nats.Msg) (model.Messager, error) {
	msgr := &messager{id: id, msg: msg, topic: msg.Subject}
	pl, err := model.FindPayloadFromData(msg.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to check payload of msg. Topic %v: %v", msgr.Topic(), err)
	}
	d, md, err := pl.UnPack(msg.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack payload(%v). Topic %v: %v", string(msg.Data), msgr.Topic(), err)
	}
	msgr.data = d
	msgr.metadata = md
	return msgr, nil
}

func (msgr *messager) ID() int64 {
	return msgr.id
}

func (msgr *messager) WithID(id int64) model.Messager {
	msgr.id = id
	return msgr
}

func (msgr *messager) Topic() string {
	return msgr.topic
}

func (msgr *messager) Data() []byte {
	return msgr.data
}

func (msgr *messager) Metadata() map[string]string {
	return msgr.metadata
}

func (msgr *messager) Ack(data []byte) error {
	return msgr.msg.Respond(data)
}

func (msgr *messager) Source() string {
	val, ok := msgr.metadata[model.KEY_SOURCE]
	if !ok {
		val = ""
	}
	return val
}

func (msgr *messager) Seq() int64 {
	str, ok := msgr.metadata[model.KEY_SEQ]
	if !ok || str == "" {
		str = "-1"
	}
	val, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		return -1
	}
	return val
}

func (msgr *messager) Dest() string {
	return msgr.dest
}

func (msgr *messager) WithDest(dest string) model.Messager {
	msgr.dest = dest
	return msgr
}
