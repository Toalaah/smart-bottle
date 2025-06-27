package secrets

import (
	_ "embed"
	"encoding/pem"

	"golang.org/x/crypto/curve25519"
)

//go:embed bottle-private.pem
var bottlePrivateKeyPEM []byte
var BottlePrivateKey []byte
var BottlePublicKey []byte

//go:embed user-private.pem
var userPrivateKeyPEM []byte
var UserPrivateKey []byte
var UserPublicKey []byte

func init() {
	var err error
	block, _ := pem.Decode(bottlePrivateKeyPEM)
	BottlePrivateKey = block.Bytes[len(block.Bytes)-32:]
	BottlePublicKey, err = curve25519.X25519(BottlePrivateKey, curve25519.Basepoint)
	if err != nil {
		panic(err)
	}

	block, _ = pem.Decode(userPrivateKeyPEM)
	UserPrivateKey = block.Bytes[len(block.Bytes)-32:]
	UserPublicKey, err = curve25519.X25519(UserPrivateKey, curve25519.Basepoint)
	if err != nil {
		panic(err)
	}
}

const BackendUsername = "testuser"
const BackendPassword = "secret"
