package inproc

import (
	"context"

	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"

	log "github.com/sirupsen/logrus"
)

type producer struct {
	*common.ProducerBase
	brk *broker
}

func NewProducer(name string, brk *broker, topic string) model.Producer {
	log.Infof("[inproc] creates producer %v on topic %v\n", name, topic)
	if brk == nil {
		log.Fatalln("[inproc] failed to create producer, broker cannot be nil")
	}
	prd := &producer{
		ProducerBase: common.NewProducerBase(name, topic),
		brk:          brk,
	}
	return prd
}

func (prd *producer) Fire(data []byte, metadata map[string]string) error {
	log.Debugf("producer %v send topic %v, metadata %v, data %v", prd.ID(), prd.Topic(), metadata, string(data))
	msg, err := NewMessager(prd.CurrentSeq(), prd.Topic(), prd.PackData(data, metadata))
	if err != nil {
		return err
	}
	err = prd.brk.Dispatch(msg)
	if err != nil {
		return err
	}
	prd.SeqInc()
	return nil
}

func (prd *producer) Request(ctx context.Context, data []byte, metadata map[string]string) ([]byte, error) {
	err := prd.Fire(data, metadata)
	return []byte("ACK"), err
}
