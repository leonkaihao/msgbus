package inproc

import (
	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"

	log "github.com/sirupsen/logrus"
)

type broker struct {
	*common.BrokerBase
	*TopicMapper
}

func NewBroker(cli model.Client, url string) model.Broker {
	log.Infof("[inproc] broker %v created", url)
	brk := new(broker)
	brk.BrokerBase = common.NewBrokerBase(cli, "inproc-"+url)
	brk.TopicMapper = NewTopicMapper()
	return brk
}

func (brk *broker) Consumer(name string, sub string, group string) model.Consumer {
	csmr := brk.BrokerBase.GetConsumer(name, sub, group)
	if csmr != nil {
		return csmr
	}
	csmr = NewConsumer(name, brk, sub, group)
	brk.BrokerBase.SetConsumer(csmr)
	return csmr
}

func (brk *broker) Producer(name string, topic string) model.Producer {
	prod := brk.BrokerBase.GetProducer(name, topic)
	if prod != nil {
		return prod
	}

	prod = NewProducer(name, brk, topic)
	brk.BrokerBase.SetProducer(prod)
	return prod
}
