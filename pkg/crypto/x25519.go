package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

var (
	ephemeralKeyBuf []byte = nil
)

func EncryptEphemeralStaticX25519(msg, publicKey []byte) ([]byte, error) {
	// Check public key is valid.
	if len(publicKey) != curve25519.ScalarSize {
		return nil, errors.New("unexpected public key size")
	}

	// Generate a new ephemeral key using a static buffer.
	if ephemeralKeyBuf == nil {
		ephemeralKeyBuf = make([]byte, curve25519.ScalarSize)
	}

	// Make sure we clear out the buffer after we no longer require it.
	defer func() {
		for i := range ephemeralKeyBuf {
			ephemeralKeyBuf[i] = 0
		}
	}()

	if _, err := rand.Read(ephemeralKeyBuf); err != nil {
		return nil, err
	}

	key, err := ComputeSharedSecret(ephemeralKeyBuf, publicKey)
	if err != nil {
		return nil, err
	}

	// Create ChaCha20Poly1305 AEAD cipher.
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Reuse buffer, now contains the corresponding ephemeral public key.
	ephemeralKeyBuf, err = curve25519.X25519(ephemeralKeyBuf, curve25519.Basepoint)
	if err != nil {
		return nil, err
	}

	// Generate nonce. As this is also the output buffer we add additional capacity for cipher overhead, the message itself, and the ephemeral key which we will append to the end of the payload.
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(msg)+aead.Overhead()+len(ephemeralKeyBuf))
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Seal cipher. We also add our ephemeral key to the associated data portion, allowing the recipient to verify that the ephemeral key which we pass alongside the cipher was not tampered.
	cipher := aead.Seal(nonce, nonce, msg, ephemeralKeyBuf)
	cipher = append(cipher, ephemeralKeyBuf...)
	return cipher, nil
}

func DecryptEphemeralStaticX25519(payload, privateKey []byte) ([]byte, error) {
	// Separate payload into cipher and ephemeral key.
	offset := len(payload) - curve25519.ScalarSize
	encryptedMessage, ephemeralKey := payload[:offset], payload[offset:]

	key, err := ComputeSharedSecret(privateKey, ephemeralKey)
	if err != nil {
		return nil, err
	}

	// Create ChaCha20Poly1305 AEAD cipher.
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	nonce, cipher := encryptedMessage[:aead.NonceSize()], encryptedMessage[aead.NonceSize():]
	msg, err := aead.Open(nil, nonce, cipher, ephemeralKey)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func ComputeSharedSecret(private, public []byte) ([]byte, error) {
	// Assert valid key lengths.
	if len(private) != curve25519.ScalarSize {
		return nil, errors.New("unexpected private key size")
	}
	if len(public) != curve25519.ScalarSize {
		return nil, errors.New("unexpected public key size")
	}
	// Compute shared point on curve.
	point, err := curve25519.X25519(private, public)
	if err != nil {
		return nil, err
	}
	// Create key via KDF.
	h := hkdf.New(sha256.New, point, nil, nil)
	key := make([]byte, chacha20poly1305.KeySize)
	if n, err := io.ReadFull(h, key); err != nil {
		return nil, err
	} else if n != chacha20poly1305.KeySize {
		return nil, errors.New("could not read full key size")
	}
	return key, nil
}
