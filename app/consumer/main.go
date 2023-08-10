package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/leonkaihao/msgbus/pkg/client"
	"github.com/leonkaihao/msgbus/pkg/common"
	"github.com/leonkaihao/msgbus/pkg/config"
	"github.com/leonkaihao/msgbus/pkg/model"
)

type serviceCb struct {
	countMap map[string]int64 // (producerid: count)
	showCfg  map[model.Consumer]string
}

func NewServiceCB() *serviceCb {
	return &serviceCb{
		countMap: map[string]int64{},
		showCfg:  map[model.Consumer]string{},
	}
}
func (svccb *serviceCb) OnReceive(csmr model.Consumer, msg model.Messager) {
	meta, _ := json.Marshal(msg.Metadata())
	prodId := msg.Source()
	if prodId == "" {
		prodId = "unknown"
	}
	cnt, ok := svccb.countMap[prodId]
	if !ok {
		svccb.countMap[prodId] = 0
	}
	cnt++
	svccb.countMap[prodId] = cnt

	var logShowStr string
	show, _ := svccb.showCfg[csmr]
	switch show {
	case config.SHOW_TOPIC:
		logShowStr = fmt.Sprintf("topic: %v", msg.Topic())
	case config.SHOW_METADATA:
		logShowStr = fmt.Sprintf("metadata: %v", string(meta))
	case config.SHOW_DATA:
		logShowStr = fmt.Sprintf("data: %v", string(msg.Data()))
	default:
		fallthrough
	case config.SHOW_ALL:
		logShowStr = fmt.Sprintf("topic: %v, METADATA %v. data %v", msg.Topic(), string(meta), string(msg.Data()))
	}
	log.Infof("consumer %v receives message from %v, seq %v, count %v, {%v}", csmr.ID(), msg.Source(), msg.Seq(), cnt, logShowStr)
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
	svccb := NewServiceCB()
	svc := common.NewService(svccb)
	csmrs := []model.Consumer{}
	for _, cfgConsumer := range cfg.Consumers {
		brk := cli.Broker(cfgConsumer.Endpoint)
		for _, cfgTopic := range cfgConsumer.Topics {
			csmr := brk.Consumer("SimC", cfgTopic.Topic, cfgTopic.Group)
			csmrs = append(csmrs, csmr)
			svccb.showCfg[csmr] = cfgConsumer.Show
		}
	}
	svc.AddConsumers(csmrs)
	defer svc.Close()
	go svc.Serve()

	excepSig := make(chan os.Signal, 1)
	signal.Notify(excepSig, os.Interrupt, syscall.SIGTERM)
	<-excepSig

}
