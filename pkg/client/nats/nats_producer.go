package nats

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/leonkaihao/msgbus/pkg/model"

	log "github.com/sirupsen/logrus"
)

type producer struct {
	id      string
	seq     int64
	name    string
	brk     *broker
	topic   string
	payload model.Payload
}

func NewProducer(name string, brk *broker, topic string) model.Producer {
	log.Infof("[NATS] creates producer %v on topic %v\n", name, topic)
	if brk == nil {
		log.Fatalln("[NATS] failed to create producer, broker cannot be nil")
	}
	prd := &producer{
		id:      fmt.Sprintf("%v-%v", name, rand.Intn(1000)),
		seq:     0,
		name:    name,
		brk:     brk,
		topic:   topic,
		payload: model.NewMsgbusPayload(),
	}
	return prd
}

func (prd *producer) PayloadType() string {
	return prd.payload.Type()
}

func (prd *producer) SetOption(k, v string) error {
	switch k {
	case model.PRD_OPTION_PAYLOAD:
		plType := v
		pl, err := model.FindPayloadByType(plType)
		if err != nil {
			return err
		}
		prd.payload = pl
	default:
		return fmt.Errorf("producer SetOption: unknown option %v", k)
	}
	return nil
}

func (prd *producer) Fire(data []byte, metadata map[string]string) error {
	log.Debugf("producer %v send topic %v, metadata %v, data %v", prd.id, prd.topic, metadata, string(data))
	err := prd.brk.conn.Publish(prd.topic, prd.packData(data, metadata))
	if err != nil {
		return err
	}
	prd.seq++
	return nil
}

func (prd *producer) Request(ctx context.Context, data []byte, metadata map[string]string) ([]byte, error) {
	resp, err := prd.brk.conn.RequestWithContext(ctx, prd.topic, prd.packData(data, metadata))
	if err != nil {
		return nil, err
	}
	prd.seq++
	return resp.Data, nil
}

func (prd *producer) Close() error {
	log.Infof("[NATS] producer(%v) %v closed.", prd.id, prd.topic)
	return nil
}

func (prd *producer) Topic() string {
	return prd.topic
}

func (prd *producer) ID() string {
	return prd.id
}

func (prd *producer) Name() string {
	return prd.name
}

func (prd *producer) CurrentSeq() int64 {
	return prd.seq
}

func (prd *producer) packData(data []byte, metadata map[string]string) []byte {
	metadata[model.KEY_SOURCE] = prd.id
	metadata[model.KEY_SEQ] = strconv.FormatInt(prd.seq, 10)
	buf, _ := prd.payload.Pack(data, metadata)
	return buf
}
