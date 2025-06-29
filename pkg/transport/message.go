package transport

import (
	"fmt"
)

type MessageType uint8

const (
	HeartBeat MessageType = 1 << iota
	WaterLevel
	Nonce
)

type Message struct {
	Type   MessageType
	Length uint8
	Value  []byte
}

func (m *Message) Load(b []byte) {
	m.Value = b
	m.Length = uint8(len(b))
}

func (m *Message) MarshalBytes() []byte {
	return append([]byte{byte(m.Type), m.Length}, m.Value...)
}

func UnmarshalBytes(m *Message, b []byte) error {
	if len(b) < 2 {
		return fmt.Errorf("expected slice of at least length 2")
	}
	m.Type = MessageType(b[0])
	m.Length = b[1]
	rest := b[2:]
	if len(rest) < int(m.Length) {
		return fmt.Errorf("expected payload value to have length of at least %d, got %d", len(rest), m.Length)
	}
	m.Value = rest[:m.Length]
	return nil
}
