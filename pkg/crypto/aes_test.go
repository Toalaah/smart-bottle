package crypto

import (
	"bytes"
	"testing"
)

func TestAES(t *testing.T) {
	var (
		key       = []byte("randomkeymaterialfromdhkex")
		msg       = []byte("hello world")
		out       = make([]byte, 39)
		recovered = make([]byte, len(msg))
	)

	ciphEnc, err := NewGCM(key)
	if err != nil {
		t.Fatalf("Expected nil error during aes init, got %s", err)
	}

	ciphDec, err := NewGCM(key)
	if err != nil {
		t.Fatalf("Expected nil error during aes init, got %s", err)
	}

	if err := EncryptAES(ciphEnc, msg, out); err != nil {
		t.Fatalf("Expected nil error while encrypting, got %s", err)
	}

	if err := DecryptAES(ciphDec, out, recovered); err != nil {
		t.Fatalf("Expected nil error while decrypting, got %s", err)
	}

	if bytes.Compare(msg, recovered) != 0 {
		t.Fatalf("Expected decrypted plaintext to be '%s', got '%s'", string(msg), string(recovered))
	}
}
