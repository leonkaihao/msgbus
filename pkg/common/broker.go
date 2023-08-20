package common

import (
	"github.com/leonkaihao/msgbus/pkg/model"

	log "github.com/sirupsen/logrus"
)

type BrokerBase struct {
	cli       model.Client
	url       string
	consumers map[string]model.Consumer
	producers map[string]model.Producer
}

func NewBrokerBase(cli model.Client, url string) *BrokerBase {
	brk := &BrokerBase{
		cli:       cli,
		url:       url,
		consumers: make(map[string]model.Consumer),
		producers: make(map[string]model.Producer),
	}
	return brk
}

func (brk *BrokerBase) URL() string {
	return brk.url
}

func (brk *BrokerBase) Consumer(name string, sub string, group string) model.Consumer {
	// need to be implemented by successor
	return nil
}

func (brk *BrokerBase) Producer(name string, topic string) model.Producer {
	// need to be implemented by successor
	return nil
}

func (brk *BrokerBase) GetConsumer(name string, sub string, group string) model.Consumer {
	consumerKey := brk.makeConsumerKey(name, sub, group)
	return brk.consumers[consumerKey]
}

func (brk *BrokerBase) GetProducer(name string, topic string) model.Producer {
	producerKey := brk.makeProducerKey(name, topic)
	return brk.producers[producerKey]
}

func (brk *BrokerBase) SetConsumer(csmr model.Consumer) {
	consumerKey := brk.makeConsumerKey(csmr.Name(), csmr.Sub(), csmr.Group())
	brk.consumers[consumerKey] = csmr
}

func (brk *BrokerBase) SetProducer(prod model.Producer) {
	producerKey := brk.makeProducerKey(prod.Name(), prod.Topic())
	brk.producers[producerKey] = prod
}

func (brk *BrokerBase) Close() error {
	brk.closeAllConsumer()
	brk.closeAllProducer()
	log.Infof("[NATS] BrokerBase %v closed.", brk.url)
	return nil
}

func (brk *BrokerBase) makeConsumerKey(name string, group string, topic string) string {
	return name + ":" + topic + ":" + group
}

func (brk *BrokerBase) makeProducerKey(name string, topic string) string {
	return name + ":" + topic
}

func (brk *BrokerBase) RemoveConsumer(csmr model.Consumer) error {
	if csmr == nil {
		return nil
	}
	csmr.Close()
	delete(brk.consumers, brk.makeConsumerKey(csmr.Name(), csmr.Sub(), csmr.Group()))
	return nil
}

func (brk *BrokerBase) RemoveProducer(prd model.Producer) error {
	if prd == nil {
		return nil
	}
	prd.Close()
	delete(brk.producers, brk.makeProducerKey(prd.Name(), prd.Topic()))
	return nil
}

func (brk *BrokerBase) closeAllConsumer() {
	for _, v := range brk.consumers {
		v.Close()
	}
	brk.consumers = map[string]model.Consumer{}
}

func (brk *BrokerBase) closeAllProducer() {
	for _, v := range brk.producers {
		v.Close()
	}
	brk.producers = map[string]model.Producer{}
}
