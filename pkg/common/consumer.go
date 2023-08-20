package common

import (
	"fmt"
	"math/rand"

	"github.com/leonkaihao/msgbus/pkg/model"
)

type ConsumerBase struct {
	id    string
	name  string
	sub   string
	group string
}

func NewConsumerBase(name string, sub string, group string) *ConsumerBase {
	csmr := &ConsumerBase{
		id:    fmt.Sprintf("%v-%v", name, rand.Intn(1000)),
		name:  name,
		sub:   sub,
		group: group,
	}
	return csmr
}

func (csmr *ConsumerBase) ID() string {
	return csmr.id
}

func (csmr *ConsumerBase) Name() string {
	return csmr.name
}

func (csmr *ConsumerBase) Subscribe() (<-chan model.Messager, error) {
	return nil, ErrNotImplemented
}

func (csmr *ConsumerBase) Close() error {
	return ErrNotImplemented
}

func (csmr *ConsumerBase) Sub() string {
	return csmr.sub
}

func (csmr *ConsumerBase) Group() string {
	return csmr.group
}
