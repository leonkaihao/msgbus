package nats

import (
	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"
	log "github.com/sirupsen/logrus"
)

type client struct {
	*common.ClientBase
}

func NewClient() model.Client {
	log.Infoln("[NATS] client created.")
	cli := &client{
		common.NewClientBase(),
	}
	return cli
}

func (cli *client) Broker(url string) model.Broker {
	if cli.ClientBase.Broker(url) == nil {
		cli.ClientBase.WithBroker(url, NewBroker(cli, url))
	}
	return cli.ClientBase.Broker(url)
}
