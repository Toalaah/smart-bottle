package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"io"
	"math/rand"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

const keySize = chacha20poly1305.KeySize

func NewGCM(key []byte) (cipher.AEAD, error) {
	var aesKey [keySize]byte
	if n, err := io.ReadFull(hkdf.New(sha256.New, key, nil, nil), aesKey[:]); err != nil {
		return nil, err
	} else if n != keySize {
		return nil, errors.New("could not read full key size")
	}
	c, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(c)
}

func EncryptAES(c cipher.AEAD, in, out []byte) error {
	if c == nil {
		return errors.New("aes not initialized")
	}
	s := c.NonceSize()
	nonce := out[:s]
	if n, err := rand.Read(nonce); err != nil {
		return err
	} else if n != s {
		return errors.New("could not read full nonce size")
	}
	c.Seal(out[s:s], nonce, in, nil)
	return nil
}

func DecryptAES(c cipher.AEAD, in []byte, out []byte) error {
	if c == nil {
		return errors.New("aes not initialized")
	}
	s := c.NonceSize()
	if len(in) < s {
		return errors.New("unexpected payload size")
	}
	nonce, cipher := in[:s], in[s:]
	_, err := c.Open(out[:0], nonce, cipher, nil)
	return err
}
