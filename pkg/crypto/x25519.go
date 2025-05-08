package crypto

import (
	"crypto/sha256"
	"errors"
	"io"
	"math/rand"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

var (
	ephemeralKeyBuf = make([]byte, curve25519.ScalarSize)
	sharedKeyBuf    = make([]byte, chacha20poly1305.KeySize)
)

func EncryptEphemeralStaticX25519(msg, publicKey []byte) ([]byte, error) {
	// Check public key is valid.
	if len(publicKey) != curve25519.ScalarSize {
		return nil, errors.New("unexpected public key size")
	}
	if _, err := rand.Read(ephemeralKeyBuf); err != nil {
		return nil, err
	}

	_, err := ComputeSharedSecret(sharedKeyBuf, ephemeralKeyBuf, publicKey)
	if err != nil {
		return nil, err
	}

	// Create ChaCha20Poly1305 AEAD cipher.
	aead, err := chacha20poly1305.New(sharedKeyBuf)
	if err != nil {
		return nil, err
	}

	// Reuse buffer, now contains the corresponding ephemeral public key.
	ephemeralKeyBuf, err = curve25519.X25519(ephemeralKeyBuf, curve25519.Basepoint)
	if err != nil {
		return nil, err
	}

	// Generate nonce. As this is also the output buffer we add additional capacity for cipher overhead, the message itself, and the ephemeral key which we will append to the end of the payload.
	// nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(msg)+aead.Overhead()+len(ephemeralKeyBuf))
	nonce := make([]byte, 128)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Seal cipher. We also add our ephemeral key to the associated data portion, allowing the recipient to verify that the ephemeral key which we pass alongside the cipher was not tampered.
	cipher := aead.Seal(nonce[:aead.NonceSize()], nonce[:aead.NonceSize()], msg, ephemeralKeyBuf)
	cipher = append(cipher, ephemeralKeyBuf...)
	return cipher, nil
}

func DecryptEphemeralStaticX25519(payload, privateKey []byte) ([]byte, error) {
	// Separate payload into cipher and ephemeral key.
	offset := len(payload) - curve25519.ScalarSize
	encryptedMessage, ephemeralKey := payload[:offset], payload[offset:]

	key := make([]byte, chacha20poly1305.KeySize)
	_, err := ComputeSharedSecret(key, privateKey, ephemeralKey)
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

func ComputeSharedSecret(out, private, public []byte) ([]byte, error) {
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
	if len(out) < chacha20poly1305.KeySize {
		return nil, errors.New("out buffer is too small")
	}
	if n, err := io.ReadFull(hkdf.New(sha256.New, point, nil, nil), out); err != nil {
		return nil, err
	} else if n != chacha20poly1305.KeySize {
		return nil, errors.New("could not read full key size")
	}
	return out, nil
}
