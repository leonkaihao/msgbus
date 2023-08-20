package nats

import (
	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"

	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type consumer struct {
	*common.ConsumerBase
	brk        *broker
	msgCount   int64
	chIn       chan *nats.Msg
	chOut      chan model.Messager
	chClose    chan int
	subscribed bool
}

func NewConsumer(name string, brk *broker, sub string, group string) model.Consumer {
	if brk == nil {
		log.Fatalln("[NATS] broker cannot be Nil.")
	}
	cBase := common.NewConsumerBase(name, sub, group)
	csmr := &consumer{
		ConsumerBase: cBase,
		brk:          brk,
	}
	log.Infof("[NATS] consumer(%v) topic group (%v:%v) created.", csmr.ID(), sub, group)
	return csmr
}

func (csmr *consumer) Subscribe() (<-chan model.Messager, error) {
	if csmr.subscribed {
		return csmr.chOut, nil
	}
	csmr.chIn = make(chan *nats.Msg, common.RCV_BUF_SIZE)
	csmr.chOut = make(chan model.Messager, common.RCV_BUF_SIZE)
	sub, err := csmr.brk.conn.QueueSubscribeSyncWithChan(csmr.Sub(), csmr.Group(), csmr.chIn)
	if err != nil {
		log.Errorf("[NATS] consumer(%v) failed to subscribe %v, group %v: %v.", csmr.ID(), csmr.Sub(), csmr.Group(), err)
		return nil, err
	}
	log.Infof("[NATS] consumer(%v) subscribed to topic %v, group %v.", csmr.ID(), csmr.Sub(), csmr.Group())
	go func(csmr *consumer) {
		csmr.chClose = make(chan int)

		defer func() {
			defer close(csmr.chIn)
			defer close(csmr.chOut)
			sub.Unsubscribe()
			csmr.subscribed = false
		}()
		for {
			select {
			case msg, ok := <-csmr.chIn:
				if !ok {
					return
				}
				msgr, err := NewMessager(csmr.msgCount, msg)
				if err != nil {
					log.Error(err)
				} else {
					csmr.chOut <- msgr
				}
				csmr.msgCount++
			case <-csmr.chClose:
				return
			}
		}
	}(csmr)
	csmr.subscribed = true
	return csmr.chOut, nil
}

func (csmr *consumer) Close() error {
	if !csmr.subscribed {
		log.Warnf("[NATS] close a closed consumer(%v) %v:%v.", csmr.ID(), csmr.Sub(), csmr.Group())
		return nil
	}
	close(csmr.chClose)
	log.Infof("[NATS] consumer(%v) %v:%v closed.", csmr.ID(), csmr.Sub(), csmr.Group())
	return nil
}
