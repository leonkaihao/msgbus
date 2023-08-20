package nats

import (
	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"

	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type broker struct {
	*common.BrokerBase
	conn *nats.Conn
}

func NewBroker(cli model.Client, url string) model.Broker {
	log.Infof("[NATS] broker for %v created", url)
	brk := new(broker)
	brk.BrokerBase = common.NewBrokerBase(cli, url)
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

func (brk *broker) Close() error {
	brk.BrokerBase.Close()
	brk.conn.Close()
	return nil
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
