package mqtt

import (
	"fmt"

	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type messager struct {
	*common.MessagerBase
	conn MQTT.Client
	msg  MQTT.Message
}

// NewMessager return wrapped message and
// a flag indicate if the message comes from legacy producer(false)
func NewMessager(conn MQTT.Client, id int64, msg MQTT.Message) (model.Messager, error) {
	mb, err := common.NewMessagerBase(id, msg.Topic(), msg.Payload())
	if err != nil {
		return nil, err
	}
	msgr := &messager{conn: conn, msg: msg, MessagerBase: mb}
	return msgr, nil
}

func (msgr *messager) Ack(data []byte) error {
	ackTopic := fmt.Sprintf("%v/ACK", msgr.Topic())
	token := msgr.conn.Publish(ackTopic, 0, false, data)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}
