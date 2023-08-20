package inproc

import (
	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/model"

	log "github.com/sirupsen/logrus"
)

type consumer struct {
	*common.ConsumerBase
	brk        *broker
	msgCount   int64
	chIn       chan model.Messager
	chOut      chan model.Messager
	chClose    chan int
	subscribed bool
}

func NewConsumer(name string, brk *broker, sub string, group string) model.Consumer {
	if brk == nil {
		log.Fatalln("[inproc] broker cannot be Nil.")
	}
	cBase := common.NewConsumerBase(name, sub, group)
	csmr := &consumer{
		ConsumerBase: cBase,
		brk:          brk,
	}
	log.Infof("[inproc] consumer(%v) topic group (%v:%v) created.", csmr.ID(), sub, group)
	return csmr
}

func (csmr *consumer) Subscribe() (<-chan model.Messager, error) {
	if csmr.subscribed {
		return csmr.chOut, nil
	}
	csmr.chIn = make(chan model.Messager, common.RCV_BUF_SIZE)
	csmr.chOut = make(chan model.Messager, common.RCV_BUF_SIZE)
	csmr.brk.Subscribe(csmr, csmr.chIn)
	log.Infof("[inproc] consumer(%v) subscribed to topic %v, group %v.", csmr.ID(), csmr.Sub(), csmr.Group())
	go func(csmr *consumer) {
		csmr.chClose = make(chan int)

		defer func() {
			csmr.brk.UnSubscribe(csmr)
			close(csmr.chOut)
			close(csmr.chIn)
			csmr.subscribed = false
		}()
		for {
			select {
			case msg, ok := <-csmr.chIn:
				if !ok {
					return
				}
				csmr.chOut <- msg
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
		log.Warnf("[inproc] close a closed consumer(%v) %v:%v.", csmr.ID(), csmr.Sub(), csmr.Group())
		return nil
	}
	close(csmr.chClose)
	log.Infof("[inproc] consumer(%v) %v:%v closed.", csmr.ID(), csmr.Sub(), csmr.Group())
	return nil
}
