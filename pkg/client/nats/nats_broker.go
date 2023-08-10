package nats

import (
	"github.com/leonkaihao/msgbus/pkg/model"

	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type broker struct {
	cli       model.Client
	url       string
	conn      *nats.Conn
	consumers map[string]model.Consumer
	producers map[string]model.Producer
}

func NewBroker(cli model.Client, url string) model.Broker {
	log.Infof("[NATS] broker for %v created", url)
	brk := new(broker)
	brk.cli = cli
	brk.url = url
	brk.consumers = make(map[string]model.Consumer)
	brk.producers = make(map[string]model.Producer)
	var err error
	// DefaultURL                = "nats://127.0.0.1:4222"
	// DefaultPort               = 4222
	// DefaultMaxReconnect       = 60
	// DefaultReconnectWait      = 2 * time.Second
	// DefaultReconnectJitter    = 100 * time.Millisecond
	// DefaultReconnectJitterTLS = time.Second
	// DefaultTimeout            = 2 * time.Second
	// DefaultPingInterval       = 2 * time.Minute
	// DefaultMaxPingOut         = 2
	// DefaultMaxChanLen         = 64 * 1024       // 64k
	// DefaultReconnectBufSize   = 8 * 1024 * 1024 // 8MB
	// RequestChanLen            = 8
	// DefaultDrainTimeout       = 30 * time.Second
	brk.conn, err = nats.Connect(url, connectOptions()...)

	if err != nil {
		log.Fatalf("[NATS] unable to connnect nats broker %v, reason: %v.", url, err)
	}
	return brk
}

func (brk *broker) URL() string {
	return brk.url
}

func (brk *broker) Consumer(name string, topic string, group string) model.Consumer {
	consumerKey := brk.makeConsumerKey(name, topic, group)
	if csmr, ok := brk.consumers[consumerKey]; ok {
		return csmr
	}
	csmr := NewConsumer(name, brk, topic, group)
	brk.consumers[consumerKey] = csmr
	return csmr
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
	brk.conn.Close()
	log.Infof("[NATS] broker %v closed.", brk.url)
	return nil
}

func (brk *broker) makeConsumerKey(name string, group string, topic string) string {
	return name + ":" + topic + ":" + group
}

func (brk *broker) makeProducerKey(name string, topic string) string {
	return name + ":" + topic
}

func (brk *broker) RemoveConsumer(csmr model.Consumer) error {
	if csmr == nil {
		return nil
	}
	csmr.Close()
	delete(brk.consumers, brk.makeConsumerKey(csmr.Name(), csmr.Topic(), csmr.Group()))
	return nil
}

func (brk *broker) RemoveProducer(prd model.Producer) error {
	if prd == nil {
		return nil
	}
	prd.Close()
	delete(brk.producers, brk.makeProducerKey(prd.Name(), prd.Topic()))
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

func natsErrHandler(nc *nats.Conn, sub *nats.Subscription, natsErr error) {
	log.Errorf("natsErrHandler: %v\n", natsErr)
	if natsErr == nats.ErrSlowConsumer {
		pendingMsgs, _, err := sub.Pending()
		if err != nil {
			log.Errorf("natsErrHandler: couldn't get pending messages: %v", err)
			return
		}
		log.Infof("natsErrHandler: Falling behind with %d pending messages on subject %q.\n",
			pendingMsgs, sub.Subject)
		// Log error, notify operations...
	}
	// check for other errors
}

func connectOptions() []nats.Option {
	return []nats.Option{
		nats.NoEcho(),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(30 * 60 * 24 * 2),  // 2 days reconnection attempts.
		nats.MaxPingsOutstanding(30 * 24 * 2), // 2 days ping before closing connection
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Infof("[NATS] disconnected! reason: %v.", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Infof("[NATS] reconnected to %v.", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Infof("[NATS] connection closed. reason: %v.", nc.LastError())
		}),
		nats.ErrorHandler(natsErrHandler),
	}
}
