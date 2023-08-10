package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/leonkaihao/msgbus/pkg/client"
	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/config"
	"github.com/leonkaihao/msgbus/pkg/model"
)

type serviceCb struct {
}

func (svccb *serviceCb) OnReceive(csmr model.Consumer, msg model.Messager) {
	log.Infof("consumer %v receives data: %v from source: %v\n", csmr.ID(), string(msg.Data()), msg.Metadata()["source"])
}

func (svccb *serviceCb) OnDestroy(inst interface{}) {
	switch inst.(type) {
	case model.Service:
		log.Infoln("service was removed")
	case model.Consumer:
		log.Infof("consumer(%v) was removed from service", inst.(model.Consumer).ID())
	}
}

func (svccb *serviceCb) OnError(inst interface{}, err error) {
	switch inst.(type) {
	case model.Service:
		log.Infoln("service closed")
	case model.Consumer:
		log.Errorf("consumer(%v) error: %v", inst.(model.Consumer).ID(), err)
	}
}

func main() {
	log.Info("simulator start")
	args := os.Args
	if len(args) != 2 {
		log.Error("Need the argument of config json file.")
		return
	}
	cfg, err := config.LoadConfig(os.Args[1])
	if err != nil {
		log.Errorf("Error: %v", err)
		return
	}

	cli, err := client.NewBuilder(client.CLI_NATS).Build()
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	brk := cli.Broker(cfg.Consumers[0].Endpoint)

	csmr1 := brk.Consumer("SimC1", "topic1", "group1")
	csmr2 := brk.Consumer("SimC1", "topic1", "group2")
	csmr3 := brk.Consumer("SimC2", "topic2", "group2")
	csmr4 := brk.Consumer("SimC2", "topic3", "group3")
	prd1 := brk.Producer("SimP1", "topic1")
	prd2 := brk.Producer("SimP2", "topic2")
	go func() {
		for {
			time.Sleep(time.Second)
			prd1.Fire([]byte("hello1"), map[string]string{"source": prd1.ID()})
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				prd2.Request(ctx, []byte("hello2"), map[string]string{"source": prd2.ID()})
			}()
		}
	}()
	svc := common.NewService(&serviceCb{})
	svc.AddConsumers([]model.Consumer{csmr1, csmr2, csmr3, csmr4})
	defer svc.Close()
	go svc.Serve()

	excepSig := make(chan os.Signal, 1)
	signal.Notify(excepSig, os.Interrupt, syscall.SIGTERM)
	<-excepSig

}
