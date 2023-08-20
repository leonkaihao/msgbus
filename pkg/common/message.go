package common

import (
	"fmt"
	"strconv"

	"github.com/leonkaihao/msgbus/pkg/model"
)

type MessagerBase struct {
	id       int64
	topic    string
	data     []byte
	metadata map[string]string
	dest     string
}

func NewMessagerBase(id int64, topic string, data []byte) (*MessagerBase, error) {
	msgr := &MessagerBase{id: id, topic: topic}
	pl, err := model.FindPayloadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to check payload of msg. Topic %v: %v", msgr.Topic(), err)
	}
	d, md, err := pl.UnPack(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack payload(%v). Topic %v: %v", string(data), msgr.Topic(), err)
	}
	msgr.data = d
	msgr.metadata = md
	return msgr, nil
}

func (msgr *MessagerBase) ID() int64 {
	return msgr.id
}

func (msgr *MessagerBase) WithID(id int64) model.Messager {
	msgr.id = id
	return msgr
}

func (msgr *MessagerBase) Topic() string {
	return msgr.topic
}

func (msgr *MessagerBase) Data() []byte {
	return msgr.data
}

func (msgr *MessagerBase) Metadata() map[string]string {
	return msgr.metadata
}

func (msgr *MessagerBase) Ack(data []byte) error {
	return ErrNotImplemented
}

func (msgr *MessagerBase) Source() string {
	val, ok := msgr.metadata[model.KEY_SOURCE]
	if !ok {
		val = ""
	}
	return val
}

func (msgr *MessagerBase) Seq() int64 {
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

func (msgr *MessagerBase) Dest() string {
	return msgr.dest
}

func (msgr *MessagerBase) WithDest(dest string) model.Messager {
	msgr.dest = dest
	return msgr
}
