package model

import (
	"testing"
)

func Test_payload_Pack(t *testing.T) {
	data := []byte("hello")
	metadata := map[string]string{"name": "bing", "sex": "male"}
	pl := NewMsgbusPayload()
	buf, err := pl.Pack(data, metadata)
	if err != nil {
		t.Error(err)
	}
	data2, metadata2, err := pl.UnPack(buf)
	if err != nil {
		t.Error(err)
	}
	if data2 == nil || string(data2) != "hello" {
		t.Error("wrong data")
	}
	if metadata2 == nil {
		t.Error("loss meta data")
	}
	val, ok := metadata2["name"]
	if !ok || val != "bing" {
		t.Error("wrong meta data")
	}

	pl2 := NewEdgebusPayload()
	buf, err = pl2.Pack(data, metadata)
	if err != nil {
		t.Error(err)
	}
	data2, metadata2, err = pl2.UnPack(buf)
	if err != nil {
		t.Error(err)
	}
	if data2 == nil || string(data2) != "hello" {
		t.Error("wrong data")
	}
	if metadata2 == nil {
		t.Error("loss meta data")
	}
	val, ok = metadata2["sex"]
	if !ok || val != "male" {
		t.Error("wrong meta data")
	}
}
