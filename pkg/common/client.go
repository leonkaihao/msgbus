package common

import (
	"github.com/leonkaihao/msgbus/pkg/model"
	log "github.com/sirupsen/logrus"
)

type ClientBase struct {
	dict map[string]model.Broker
}

func NewClientBase() *ClientBase {
	log.Infoln("[NATS] client created.")
	cli := new(ClientBase)
	cli.dict = make(map[string]model.Broker)
	return cli
}

func (cli *ClientBase) Broker(url string) model.Broker {
	brk, ok := cli.dict[url]
	if !ok {
		return nil
	}
	return brk
}

func (cli *ClientBase) WithBroker(url string, broker model.Broker) *ClientBase {
	cli.dict[url] = broker
	return cli
}

func (cli *ClientBase) Remove(brk model.Broker) error {
	delete(cli.dict, brk.URL())
	return nil
}

func (cli *ClientBase) Close() error {
	for _, v := range cli.dict {
		v.Close()
	}
	log.Infoln("[NATS] client closed.")
	return nil
}
