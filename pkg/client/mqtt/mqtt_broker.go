package mqtt

import (
	"time"

	"github.com/leonkaihao/msgbus/pkg/model"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type broker struct {
	cli       model.Client
	url       string
	conn      MQTT.Client
	consumers map[string]model.Consumer
	producers map[string]model.Producer
}

func NewBroker(cli model.Client, url string) model.Broker {
	log.Infof("[MQTT] broker for %v created", url)
	brk := new(broker)
	brk.cli = cli
	brk.url = url
	brk.consumers = make(map[string]model.Consumer)
	brk.producers = make(map[string]model.Producer)
	brk.conn = MQTT.NewClient(connectOptions(url))
	if token := brk.conn.Connect(); token.Wait() && token.Error() != nil {
		log.Errorf("[MQTT] failed to start mqtt client, %v", token.Error())
		return nil
	}

	return brk
}

func (brk *broker) URL() string {
	return brk.url
}

// Consumer does not support group, default empty
func (brk *broker) Consumer(name string, topic string, group string) model.Consumer {
	consumerKey := brk.makeConsumerKey(name, topic)
	if csmr, ok := brk.consumers[consumerKey]; ok {
		return csmr
	}
	csmr := NewConsumer(name, brk, topic)
	brk.consumers[consumerKey] = csmr
	return csmr
}

// StreamConsumer is not supported by MQTT, just return normal consumer
func (brk *broker) StreamConsumer(name string, topic string, group string, streamName string, streamConsName string) model.Consumer {
	return brk.Consumer(name, topic, group)
}

func (brk *broker) Producer(name string, topic string) model.Producer {
	producerKey := brk.makeProducerKey(name, topic)
	if prod, ok := brk.producers[producerKey]; ok {
		return prod
	}

	prod := NewProducer(name, brk, topic)
	brk.producers[producerKey] = prod
	return prod
}

func (brk *broker) Close() error {
	brk.closeAllConsumer()
	brk.closeAllProducer()
	brk.conn.Disconnect(500)
	log.Infof("[MQTT] broker %v closed.", brk.url)
	return nil
}

func (brk *broker) makeConsumerKey(name string, topic string) string {
	return name + ":" + topic
}

func (brk *broker) makeProducerKey(name string, topic string) string {
	return name + ":" + topic
}

func (brk *broker) RemoveConsumer(csmr model.Consumer) error {
	if csmr != nil {
		csmr.Close()
		delete(brk.consumers, brk.makeConsumerKey(csmr.Name(), csmr.Topic()))
	}
	return nil
}

func (brk *broker) RemoveProducer(prd model.Producer) error {
	if prd != nil {
		prd.Close()
		delete(brk.producers, brk.makeProducerKey(prd.Name(), prd.Topic()))
	}
	return nil
}

func (brk *broker) closeAllConsumer() {
	for _, v := range brk.consumers {
		v.Close()
	}
	brk.consumers = map[string]model.Consumer{}
}

func (brk *broker) closeAllProducer() {
	for _, v := range brk.producers {
		v.Close()
	}
	brk.producers = map[string]model.Producer{}
}

func connectOptions(url string) *MQTT.ClientOptions {

	connOpts := MQTT.NewClientOptions()
	connOpts.AddBroker(url)
	connOpts.SetAutoReconnect(true)
	connOpts.SetCleanSession(true)
	connOpts.SetKeepAlive(60 * time.Second)
	connOpts.SetConnectTimeout(time.Duration(2) * time.Second)

	connOpts.OnConnect = func(c MQTT.Client) {
		log.Infof("[MQTT] connected: %v", url)
	}

	connOpts.OnConnectionLost = func(c MQTT.Client, err error) {
		log.Infof("[MQTT] connection lost: %s", err.Error())
	}

	connOpts.OnReconnecting = func(c MQTT.Client, opts *MQTT.ClientOptions) {
		log.Infof("[MQTT] reconnecting to: %v", connOpts.Servers)
	}

	connOpts.DefaultPublishHandler = func(c MQTT.Client, msg MQTT.Message) {
		log.Debugf("[MQTT] received Message, topic: %s, payload: %v", msg.Topic(), msg.Payload())
	}

	return connOpts
}
