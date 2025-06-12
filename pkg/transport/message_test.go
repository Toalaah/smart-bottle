package transport

import (
	"bytes"
	"testing"
)

func TestMessageMarshaling(t *testing.T) {
	raw := []byte("hello world")
	msg := &Message{Type: WaterLevel}
	expected := append([]byte{byte(WaterLevel), byte(len(raw))}, raw...)
	msg.Load(raw)
	got := msg.MarshalBytes()
	if bytes.Compare(got, expected) != 0 {
		t.Errorf("Expected decrypted plaintext to be '%v', got '%v'", expected, got)
	}
}

func TestMessageUnmarshaling(t *testing.T) {
	raw := []byte("hello world")
	msg := &Message{Type: WaterLevel}
	msg.Load(raw)
	payload := msg.MarshalBytes()

	out := &Message{}
	if err := UnmarshalBytes(out, payload); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(out.Value, raw) != 0 {
		t.Errorf("Expected decrypted plaintext to be '%v', got '%v'", raw, out.Value)
	}
}

func TestMessageUnmarshalingInvalid(t *testing.T) {
	payload := []byte{byte(WaterLevel), byte(10)}
	out := &Message{}
	if err := UnmarshalBytes(out, payload); err == nil {
		t.Fatal(err)
	}
}
