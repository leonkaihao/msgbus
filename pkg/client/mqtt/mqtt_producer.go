package mqtt

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/leonkaihao/msgbus/pkg/model"

	MQTT "github.com/eclipse/paho.mqtt.golang"
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
	log.Infof("[MQTT] creates producer %v on topic %v\n", name, topic)
	if brk == nil {
		log.Error("[MQTT] failed to create producer, broker cannot be nil")
		return nil
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

// Fire send topic with original topic defined with producer, no additional msg_id postfix
// consumer side can respond with topic of "/ACK" postfix(just call the Ack() of received messager) but producer won't react to it.
func (prd *producer) Fire(data []byte, metadata map[string]string) error {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	log.Debugf("producer %v send topic %v, metadata %v, data %v", prd.id, prd.topic, metadata, string(data))
	token := prd.brk.conn.Publish(prd.topic, 0, false, prd.packData(data, metadata))
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	prd.seq++
	return nil
}

// Request send topic with additional postfix "/<msg_id>",
// msg_id is globally unique so that the consumer can respond ACK to it with topic of "/ACK" postfix
func (prd *producer) Request(ctx context.Context, data []byte, metadata map[string]string) ([]byte, error) {
	msgID := fmt.Sprintf("%v", rand.Uint32())
	topic := fmt.Sprintf("%v/%v", prd.topic, msgID)
	ackTopic := fmt.Sprintf("%v/%v/ACK", prd.topic, msgID)
	if metadata == nil {
		metadata = make(map[string]string)
	}
	chAck := make(chan []byte)
	defer close(chAck)
	// ACK callback, note that multiple subscribers can trigger the callback multiple times
	onAckReceived := func(c MQTT.Client, msg MQTT.Message) {
		chAck <- msg.Payload()
	}
	if token := prd.brk.conn.Subscribe(ackTopic, 0, onAckReceived); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("[MQTT] failed to subscribe topic %v: %v", prd.topic, token.Error())
	}
	defer prd.brk.conn.Unsubscribe(ackTopic)

	token := prd.brk.conn.Publish(topic, 0, false, prd.packData(data, metadata))
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	var ack []byte
	select {
	case ack = <-chAck:
	case <-time.After(time.Second * 5):
		return nil, fmt.Errorf("[MQTT] failed to get response from subscribed topic %v", prd.topic)
	}
	prd.seq++
	return ack, nil
}

func (prd *producer) Close() error {
	log.Infof("[MQTT] producer(%v) %v closed.", prd.id, prd.topic)
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
	// NOTE: the standard payload has metadata and data field,
	//   but some public messages (like range report type) are raw messages that should keep metadata nil
	if metadata != nil {
		metadata[model.KEY_SOURCE] = prd.id
		metadata[model.KEY_SEQ] = strconv.FormatInt(prd.seq, 10)
	}
	buf, _ := prd.payload.Pack(data, metadata)
	return buf
}
