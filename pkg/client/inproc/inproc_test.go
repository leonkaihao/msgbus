package inproc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	cli := NewClient()
	brk := cli.Broker("1")
	prd1 := brk.Producer("prod1", "tpc1")
	csmr1 := brk.Consumer("cmsr1", "tpc1", "grp1")
	msgChan, err := csmr1.Subscribe()
	assert.NoError(t, err)
	closeChan := make(chan int)
	defer close(closeChan)
	go func() {
		for {
			select {
			case msg := <-msgChan:
				msg.Ack(msg.Data())
			case <-closeChan:
				return
			}
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := prd1.Request(ctx, []byte("hello"), map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, string(resp), "hello")
	resp, err = prd1.Request(ctx, []byte("world"), map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, string(resp), "world")
}
