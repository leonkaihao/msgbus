package common

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/leonkaihao/msgbus/pkg/model"

	log "github.com/sirupsen/logrus"
)

type ProducerBase struct {
	id      string
	seq     int64
	name    string
	topic   string
	payload model.Payload
}

func NewProducerBase(name string, topic string) *ProducerBase {
	prd := &ProducerBase{
		id:      fmt.Sprintf("%v-%v", name, rand.Intn(1000)),
		seq:     0,
		name:    name,
		topic:   topic,
		payload: model.NewMsgbusPayload(),
	}
	return prd
}

func (prd *ProducerBase) PayloadType() string {
	return prd.payload.Type()
}

func (prd *ProducerBase) SetOption(k, v string) error {
	switch k {
	case model.PRD_OPTION_PAYLOAD:
		plType := v
		pl, err := model.FindPayloadByType(plType)
		if err != nil {
			return err
		}
		prd.payload = pl
	default:
		return fmt.Errorf("ProducerBase SetOption: unknown option %v", k)
	}
	return nil
}

func (prd *ProducerBase) Fire(data []byte, metadata map[string]string) error {
	return ErrNotImplemented
}

func (prd *ProducerBase) Request(ctx context.Context, data []byte, metadata map[string]string) ([]byte, error) {
	return nil, ErrNotImplemented
}

func (prd *ProducerBase) Close() error {
	log.Infof("[NATS] ProducerBase(%v) %v closed.", prd.id, prd.topic)
	return nil
}

func (prd *ProducerBase) Topic() string {
	return prd.topic
}

func (prd *ProducerBase) ID() string {
	return prd.id
}

func (prd *ProducerBase) Name() string {
	return prd.name
}

func (prd *ProducerBase) CurrentSeq() int64 {
	return prd.seq
}

func (prd *ProducerBase) SeqInc() {
	prd.seq++
}

func (prd *ProducerBase) PackData(data []byte, metadata map[string]string) []byte {
	metadata[model.KEY_SOURCE] = prd.id
	metadata[model.KEY_SEQ] = strconv.FormatInt(prd.seq, 10)
	buf, _ := prd.payload.Pack(data, metadata)
	return buf
}
