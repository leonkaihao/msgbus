package nats

import (
	"github.com/leonkaihao/msgbus/pkg/model"
	log "github.com/sirupsen/logrus"
)

type client struct {
	dict map[string]model.Broker
}

func NewClient() model.Client {
	log.Infoln("[NATS] client created.")
	cli := new(client)
	cli.dict = make(map[string]model.Broker)
	return cli
}

func (cli *client) Broker(url string) model.Broker {
	brk, ok := cli.dict[url]

	if !ok {
		brk = NewBroker(cli, url)
		cli.dict[url] = brk
	}
	return brk
}

func (cli *client) Remove(brk model.Broker) error {
	delete(cli.dict, brk.URL())
	return nil
}

func (cli *client) Close() error {
	for _, v := range cli.dict {
		v.Close()
	}
	log.Infoln("[NATS] client closed.")
	return nil
}
