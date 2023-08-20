package common

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/leonkaihao/msgbus/pkg/model"
)

type service struct {
	cb       model.ServiceCallback
	cms      map[string]model.Consumer
	msgChan  chan model.Messager
	exitChan chan bool
	wg       sync.WaitGroup
}

func NewService(cb model.ServiceCallback) model.Service {
	return &service{
		cb:       cb,
		cms:      make(map[string]model.Consumer),
		msgChan:  make(chan model.Messager, RCV_BUF_SIZE),
		exitChan: make(chan bool),
	}
}

func (svc *service) AddConsumers(cms []model.Consumer) {
	for _, csmr := range cms {
		if _, ok := svc.cms[csmr.ID()]; ok {
			return
		}
		svc.wg.Add(1)
		svc.cms[csmr.ID()] = csmr
		go func(cm model.Consumer) {
			defer func() {
				log.Infof("msgbus service: consumer %v exited.", cm.ID())
				svc.wg.Done()
			}()
			chMsg, err := cm.Subscribe()
			if err != nil {
				if svc.cb != nil {
					svc.cb.OnError(cm, err)
				}
				svc.removeConsumer(cm)
				return
			}
			for {
				select {
				case msgr, ok := <-chMsg:
					if !ok {
						svc.cb.OnError(cm, fmt.Errorf("failed to receive a message from channel %v:%v", cm.Sub(), cm.Group()))
						svc.removeConsumer(cm)
						return
					}
					log.Debugf("msgbus service: consumer %v received a message %v", cm.ID(), cm.Sub())
					svc.msgChan <- msgr.WithDest(cm.ID())
				case <-svc.exitChan:
					svc.removeConsumer(cm)
					return
				}
			}
		}(csmr)
	}
}

func (svc *service) removeConsumer(cm model.Consumer) {
	if svc.cb != nil {
		svc.cb.OnDestroy(cm)
	}
	delete(svc.cms, cm.ID())
}

func (svc *service) Serve() error {
	log.Infoln("[msgbus] start serving...")

	svc.wg.Add(1)
	defer svc.wg.Done()
	for {
		select {
		case msgr, ok := <-svc.msgChan:
			if !ok {
				return fmt.Errorf("msg channel is closed")
			}
			msgr.Ack([]byte(model.CONSUMER_ACK))
			if svc.cb != nil {
				svc.cb.OnReceive(svc.cms[msgr.Dest()], msgr)
			}
		case <-svc.exitChan:
			return fmt.Errorf("system exit")
		}
	}
}

func (svc *service) Close() error {
	close(svc.exitChan)
	svc.wg.Wait()
	close(svc.msgChan)
	if svc.cb != nil {
		svc.cb.OnDestroy(svc)
	}
	return nil
}
