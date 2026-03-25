package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leonkaihao/msgbus/pkg/client"
	"github.com/leonkaihao/msgbus/pkg/model"
	log "github.com/sirupsen/logrus"
)

func main() {

	cli, err := client.NewBuilder(client.CLI_MQTT3).Build()
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer cli.Close()

	brk := cli.Broker("tcp://0.0.0.0:1883")
	chClose := make(chan bool)

	// Consumer-PING
	csmrPing := brk.Consumer("Consumer-PING", "ping/+", "")
	chPing, err := csmrPing.Subscribe()
	if err != nil {
		log.Error(err.Error())
		return
	}
	go func() {
		for {
			select {
			case data, ok := <-chPing:
				if !ok {
					return
				}
				log.Infof("%v received %v from %v, topic %v", csmrPing.ID(), string(data.Data()), data.Source(), data.Topic())
				data.Ack([]byte(csmrPing.ID()))
			case <-chClose:
				return
			}
		}
	}()

	// Consumer-PONG
	csmrPong := brk.Consumer("Consumer-PONG", "pong", "")
	chPong, err := csmrPong.Subscribe()
	if err != nil {
		log.Error(err.Error())
		return
	}
	go func() {
		for {
			select {
			case data, ok := <-chPong:
				if !ok {
					return
				}
				log.Infof("%v received %v from %v, topic %v", csmrPong.ID(), string(data.Data()), data.Source(), data.Topic())
				data.Ack([]byte(csmrPong.ID()))
			case <-chClose:
				return
			}
		}
	}()

	prdcPing := brk.Producer("Producer-PING", "ping")
	prdcPing.SetOption(model.PRD_OPTION_PAYLOAD, model.PLTYPE_EDGEBUS)
	prdcPong := brk.Producer("Producer-PONG", "pong")
	prdcPong.SetOption(model.PRD_OPTION_PAYLOAD, model.PLTYPE_RAW)

	// Producer-PING
	go func() {
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				data := "ping!!"
				log.Infof("%v sends request %v to broker, topic %v", prdcPing.ID(), data, prdcPing.Topic())
				resp, err := prdcPing.Request(context.Background(), []byte(data), nil)
				if err != nil {
					log.Warning(err.Error())
					break
				}
				log.Infof("%v gets response from %v, ACK %v", prdcPing.ID(), prdcPing.Topic(), string(resp))
			case <-chClose:
				return
			}
		}
	}()

	//Producer-PONG
	go func() {
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				data := "pong!!"
				log.Infof("%v fire message %v to broker, topic %v", prdcPong.ID(), data, prdcPong.Topic())
				err := prdcPong.Fire([]byte("pong!!"), nil)
				if err != nil {
					log.Warning(err.Error())
				}
			case <-chClose:
				return
			}
		}
	}()

	excepSig := make(chan os.Signal, 1)
	signal.Notify(excepSig, os.Interrupt, syscall.SIGTERM)
	<-excepSig
}
