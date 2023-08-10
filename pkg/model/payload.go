package model

import (
	"encoding/json"
	"fmt"
)

const (
	PLTYPE_RAW     = "raw"
	PLTYPE_MSGBUS  = "msgbus"
	PLTYPE_EDGEBUS = "edgebus"
)

type Package struct {
	Data     []byte            `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type Payload interface {
	Type() string
	Pack(data []byte, metadata map[string]string) ([]byte, error)
	UnPack(data []byte) ([]byte, map[string]string, error)
}

//------------------------------------

type msgbusPayload struct {
}

func NewMsgbusPayload() Payload {
	return &msgbusPayload{}
}

func (pl *msgbusPayload) Type() string {
	return PLTYPE_MSGBUS
}

func (pl *msgbusPayload) Pack(data []byte, metadata map[string]string) ([]byte, error) {
	pkg := &Package{data, metadata}
	return json.Marshal(pkg)
}

func (pl *msgbusPayload) UnPack(buf []byte) ([]byte, map[string]string, error) {
	pkg := &Package{}
	err := json.Unmarshal(buf, pkg)
	return pkg.Data, pkg.Metadata, err
}

//-----------------------------

type edgebusPayload struct {
}

func NewEdgebusPayload() Payload {
	return &edgebusPayload{}
}

func (pl *edgebusPayload) Type() string {
	return PLTYPE_EDGEBUS
}
func (pl *edgebusPayload) Pack(data []byte, metadata map[string]string) ([]byte, error) {
	header := []byte{0xA1, 0x60}

	metabytes := []byte{}
	for k, v := range metadata {
		metabytes = append(metabytes, []byte(k)...)
		metabytes = append(metabytes, byte(0))
		metabytes = append(metabytes, []byte(v)...)
		metabytes = append(metabytes, byte(0))
	}

	metaSize := len(metabytes)
	metaNum := len(metadata)
	metaHeader := []byte{
		byte(metaNum >> 24),
		byte((metaNum >> 16) & 0xff),
		byte((metaNum >> 8) & 0xff),
		byte(metaNum & 0xff),
	}

	dataSize := len(data)
	dataHeader := []byte{
		byte(dataSize >> 24),
		byte((dataSize >> 16) & 0xff),
		byte((dataSize >> 8) & 0xff),
		byte(dataSize & 0xff),
	}

	msgSize := 10 + metaSize + 4 + dataSize
	msgHeader := []byte{
		byte(msgSize >> 24),
		byte((msgSize >> 16) & 0xff),
		byte((msgSize >> 8) & 0xff),
		byte(msgSize & 0xff),
	}
	buf := append(header, msgHeader...)
	buf = append(buf, metaHeader...)
	buf = append(buf, metabytes...)
	buf = append(buf, dataHeader...)
	buf = append(buf, data...)

	return buf, nil
}

func (pl *edgebusPayload) UnPack(buf []byte) ([]byte, map[string]string, error) {
	if len(buf) < 10 {
		return nil, nil, fmt.Errorf("payload %v is not the format of edgebus", buf)
	}
	if buf[0] != 0xA1 || buf[1] != 0x60 {
		return nil, nil, fmt.Errorf("payload header %v... is not the format of edgebus", buf[:2])
	}
	msgSize := (int(buf[2]) << 24) | (int(buf[3]) << 16) | (int(buf[4]) << 8) | int(buf[5])
	if msgSize != len(buf) {
		return nil, nil, fmt.Errorf("expect msgsize %v but got %v", len(buf), msgSize)
	}

	metaNum := (int(buf[6]) << 24) | (int(buf[7]) << 16) | (int(buf[8]) << 8) | int(buf[9])

	metadata := map[string]string{}
	start := 10
	for i := 0; i < metaNum; i++ {
		k, v, end := pl.getMetaItem(buf, start)
		metadata[k] = v
		start = end
	}

	dataHeadOffs := start
	dataSize := (int(buf[dataHeadOffs]) << 24) |
		(int(buf[dataHeadOffs+1]) << 16) |
		(int(buf[dataHeadOffs+2]) << 8) |
		int(buf[dataHeadOffs+3])
	if dataHeadOffs+4+dataSize != len(buf) {
		return nil, nil, fmt.Errorf("payload data part %v... is corrupted", buf[dataHeadOffs:dataHeadOffs+4])
	}
	data := buf[dataHeadOffs+4:]
	return data, metadata, nil
}

func (pl *edgebusPayload) getMetaItem(buf []byte, start int) (string, string, int) {
	var key, val string
	var i int
	for i = start; i < len(buf); i++ {
		if buf[i] == 0 {
			key = string(buf[start:i])
			start = i + 1
			break
		}
	}
	for i = start; i < len(buf); i++ {
		if buf[i] == byte(0) {
			val = string(buf[start:i])
			break
		}
	}
	return key, val, i + 1
}

//-----

type rawPayload struct {
}

func NewRawPayload() Payload {
	return &rawPayload{}
}

func (pl *rawPayload) Type() string {
	return PLTYPE_RAW
}

func (pl *rawPayload) Pack(data []byte, metadata map[string]string) ([]byte, error) {
	return data, nil
}

func (pl *rawPayload) UnPack(buf []byte) ([]byte, map[string]string, error) {
	return buf, map[string]string{}, nil
}

func FindPayloadFromData(buf []byte) (Payload, error) {
	switch {
	case buf[0] == 0xA1 && buf[1] == 0x60:
		return NewEdgebusPayload(), nil
	case buf[0] == '{':
		return NewMsgbusPayload(), nil
	default:
		return NewRawPayload(), nil
	}
}

func FindPayloadByType(plType string) (Payload, error) {
	switch plType {
	case PLTYPE_MSGBUS:
		return NewMsgbusPayload(), nil
	case PLTYPE_EDGEBUS:
		return NewEdgebusPayload(), nil
	case PLTYPE_RAW:
		return NewRawPayload(), nil
	default:
		return nil, fmt.Errorf("unknown payload type %v", plType)
	}
}
