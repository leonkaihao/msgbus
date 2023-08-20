package nats

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
	log.Infof("[NATS] creates producer %v on topic %v\n", name, topic)
	if brk == nil {
		log.Fatalln("[NATS] failed to create producer, broker cannot be nil")
	}
	prd := &producer{
		ProducerBase: common.NewProducerBase(name, topic),
		brk:          brk,
	}
	return prd
}

func (prd *producer) Fire(data []byte, metadata map[string]string) error {
	log.Debugf("producer %v send topic %v, metadata %v, data %v", prd.ID(), prd.Topic(), metadata, string(data))
	err := prd.brk.conn.Publish(prd.Topic(), prd.PackData(data, metadata))
	if err != nil {
		return err
	}
	prd.SeqInc()
	return nil
}

func (prd *producer) Request(ctx context.Context, data []byte, metadata map[string]string) ([]byte, error) {
	resp, err := prd.brk.conn.RequestWithContext(ctx, prd.Topic(), prd.PackData(data, metadata))
	if err != nil {
		return nil, err
	}
	prd.SeqInc()
	return resp.Data, nil
}
