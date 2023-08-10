package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/leonkaihao/msgbus/pkg/client"
	"github.com/leonkaihao/msgbus/pkg/config"
	"github.com/leonkaihao/msgbus/pkg/model"
)

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

	brk := cli.Broker(cfg.Producer.Endpoint)

	excepSig := make(chan os.Signal, 1)
	signal.Notify(excepSig, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	signalExit := make(chan int)

	for _, event := range cfg.Producer.Events {
		prd := brk.Producer("SimP", event.Topic)
		pl := model.PLTYPE_MSGBUS
		if event.Payload != "" {
			pl = event.Payload
		}
		if err := prd.SetOption(model.PRD_OPTION_PAYLOAD, pl); err != nil {
			log.Error(err)
			continue
		}

		repeatCount := event.Repeat
		delayMili := event.Delay
		sleepMili := event.Sleep
		mode := event.Mode

		eventData := event.Data
		eventMetaData := event.Metadata
		if eventData == nil {
			eventData = make(map[string]interface{})
		}
		if eventMetaData == nil {
			eventMetaData = make(map[string]string)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			var count int64 = 0
			var lastMsgID int64
			tmAfter := time.After(time.Second * 5)
			for i := 0; i < repeatCount; i++ {
				buf, err := json.Marshal(eventData)
				if err != nil {
					log.Errorf("Error: %s\n", err)
					return
				}
				if delayMili != 0 {
					time.Sleep(time.Millisecond * time.Duration(delayMili))
				}

				func() {
					if mode == config.PUB_MODE_REQ {
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
						defer cancel()
						ack, err := prd.Request(ctx, buf, eventMetaData)
						if err != nil {
							log.Errorf("%v failed to request topic %v: %v", prd.Name(), prd.Topic(), err)
							time.Sleep(time.Second * 5)
						}
						log.Infof("send topic %v(%v) got ack: %v", prd.Topic(), count, string(ack))
					} else if mode == config.PUB_MODE_FIRE {
						err := prd.Fire(buf, eventMetaData)
						if err != nil {
							log.Errorf("%v failed to fire topic %v: %v", prd.Name(), prd.Topic(), err)
							time.Sleep(time.Second * 5)
						}
						log.Infof("send topic %v(%v)", prd.Topic(), count)
					}
					count++
				}()

				if sleepMili != 0 {
					time.Sleep(time.Millisecond * time.Duration(sleepMili))
				}

				select {
				case <-tmAfter:
					speed := (count - lastMsgID) / 5
					lastMsgID = count
					log.Infof("Producing rate %d/s, producer %v reaches seq %v\n", speed, prd.ID(), prd.CurrentSeq())
					tmAfter = time.After(time.Second * 5)

				case <-signalExit:
					log.Infof("received exception signal")
					return
				default:
				}
			}
		}()
	}

	<-excepSig
	close(signalExit)
	wg.Wait()
}
