package nats

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"

	nats "github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type consumer struct {
	id         string
	name       string
	brk        *broker
	topic      string
	group      string
	msgCount   int64
	chIn       chan *nats.Msg
	chOut      chan model.Messager
	chClose    chan int
	wg         sync.WaitGroup
	subscribed bool
}

func NewConsumer(name string, brk *broker, topic string, group string) model.Consumer {
	if brk == nil {
		log.Fatalln("[NATS] broker cannot be Nil.")
	}
	csmr := &consumer{
		id:    fmt.Sprintf("%v-%v", name, rand.Intn(1000)),
		name:  name,
		brk:   brk,
		topic: topic,
		group: group,
	}
	log.Infof("[NATS] consumer(%v) topic group (%v:%v) created.", csmr.ID(), topic, group)
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
	csmr.chIn = make(chan *nats.Msg, common.RCV_BUF_SIZE)
	csmr.chOut = make(chan model.Messager, common.RCV_BUF_SIZE)
	sub, err := csmr.brk.conn.QueueSubscribeSyncWithChan(csmr.topic, csmr.group, csmr.chIn)
	if err != nil {
		log.Errorf("[NATS] consumer(%v) failed to subscribe topic %v, group %v: %v.", csmr.id, csmr.topic, csmr.group, err)
		return nil, err
	}
	log.Infof("[NATS] consumer(%v) subscribed to topic %v, group %v.", csmr.id, csmr.topic, csmr.group)
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
		log.Warnf("[NATS] close a closed consumer(%v) %v:%v.", csmr.id, csmr.topic, csmr.group)
		return nil
	}
	close(csmr.chClose)
	csmr.wg.Wait()
	log.Infof("[NATS] consumer(%v) %v:%v closed.", csmr.id, csmr.topic, csmr.group)
	return nil
}

func (csmr *consumer) Topic() string {
	return csmr.topic
}

func (csmr *consumer) Group() string {
	return csmr.group
}
