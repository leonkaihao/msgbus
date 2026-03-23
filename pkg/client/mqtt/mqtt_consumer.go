package mqtt

import (
	"fmt"
	"math/rand"

	// "context"

	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type consumer struct {
	id         string
	name       string
	brk        *broker
	topic      string
	msgCount   int64
	chIn       chan MQTT.Message
	chOut      chan model.Messager
	chClose    chan int
	subscribed bool
}

func NewConsumer(name string, brk *broker, topic string) model.Consumer {
	if brk == nil {
		log.Error("[MQTT] broker cannot be Nil.")
		return nil
	}
	csmr := &consumer{
		id:    fmt.Sprintf("%v-%v", name, rand.Intn(1000)),
		name:  name,
		brk:   brk,
		topic: topic,
	}

	log.Infof("[MQTT] consumer(%v) topic %v created.", csmr.ID(), topic)

	return csmr
}

func (csmr *consumer) ID() string {
	return csmr.id
}

func (csmr *consumer) Name() string {
	return csmr.name
}

func (csmr *consumer) Subscribe() (<-chan model.Messager, error) {
	if csmr.subscribed {
		return csmr.chOut, nil
	}
	csmr.chIn = make(chan MQTT.Message, common.RCV_BUF_SIZE)
	csmr.chOut = make(chan model.Messager, common.RCV_BUF_SIZE)

	onMsgReceived := func(c MQTT.Client, msg MQTT.Message) {
		msgr, err := NewMessager(csmr.brk.conn, csmr.msgCount, msg)
		if err != nil {
			log.Warning(err.Error())
		} else {
			csmr.chOut <- msgr
		}
		csmr.msgCount++
	}

	if token := csmr.brk.conn.Subscribe(csmr.topic, 0, onMsgReceived); token.Wait() && token.Error() != nil {
		err := fmt.Errorf("[MQTT] failed to subscribe topic %v: %v", csmr.topic, token.Error())
		return nil, err
	}

	log.Infof("[MQTT] consumer(%v) subscribed to topic %v.", csmr.id, csmr.topic)
	go func(csmr *consumer) {
		csmr.chClose = make(chan int)

		defer func() {
			defer close(csmr.chIn)
			defer close(csmr.chOut)
			csmr.brk.conn.Unsubscribe(csmr.Topic())
			csmr.subscribed = false
		}()
		<-csmr.chClose
	}(csmr)
	csmr.subscribed = true
	return csmr.chOut, nil
}

func (csmr *consumer) Close() error {
	if !csmr.subscribed {
		log.Warningf("[MQTT] close a closed consumer(%v) %v.", csmr.id, csmr.topic)
		return nil
	}
	close(csmr.chClose)
	log.Infof("[MQTT] consumer(%v) %v closed.", csmr.id, csmr.topic)
	return nil
}

func (csmr *consumer) Topic() string {
	return csmr.topic
}

// Group is not supported by MQTT
func (csmr *consumer) Group() string {
	return ""
}
