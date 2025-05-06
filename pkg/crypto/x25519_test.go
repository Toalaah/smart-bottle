package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/curve25519"
)

func TestEncryptionDecryption(t *testing.T) {
	private := make([]byte, curve25519.ScalarSize)
	if _, err := rand.Read(private); err != nil {
		t.Fatal(err)
	}
	public, err := curve25519.X25519(private, curve25519.Basepoint)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("hello world")
	cipher, err := EncryptEphemeralStaticX25519(msg, public)
	if err != nil {
		t.Fatal(err)
	}

	recovered, err := DecryptEphemeralStaticX25519(cipher, private)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(msg, recovered) != 0 {
		t.Errorf("Expected decrypted plaintext to be '%s', got '%s'", string(msg), string(recovered))
	}
}
