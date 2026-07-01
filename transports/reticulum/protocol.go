package reticulum

import (
	"sync/atomic"
	"time"
)

const (
	ProtocolVersion = "0.1.0"
	MaxMessageSize  = 65536
)

type MessageType int

const (
	MsgData         MessageType = 0
	MsgAck          MessageType = 1
	MsgHeartbeat    MessageType = 2
	MsgDiscoveryReq MessageType = 3
	MsgDiscoveryResp MessageType = 4
)

var globalSeq atomic.Uint64

type ProtocolMessage struct {
	Type       MessageType `json:"type"`
	Seq        uint64      `json:"seq"`
	Ack        uint64      `json:"ack,omitempty"`
	Source     string      `json:"source"`
	Target     string      `json:"target"`
	Payload    []byte      `json:"payload,omitempty"`
	Timestamp  int64       `json:"timestamp"`
	Encrypted  bool        `json:"encrypted"`
}

func NewProtocolMessage(msgType MessageType, source, target string, payload []byte) ProtocolMessage {
	return ProtocolMessage{
		Type:      msgType,
		Seq:       globalSeq.Add(1),
		Source:    source,
		Target:    target,
		Payload:   payload,
		Timestamp: time.Now().UnixNano(),
	}
}
