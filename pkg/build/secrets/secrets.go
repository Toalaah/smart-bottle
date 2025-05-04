package secrets

import (
	"crypto/ed25519"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
)

//go:embed bottle-private.pem
var bottlePrivateKeyPEM []byte
var BottlePrivateKey ed25519.PrivateKey
var BottlePublicKey ed25519.PublicKey

func init() {
	priv, pub := parse25519Keypair(bottlePrivateKeyPEM)
	BottlePrivateKey = priv
	BottlePublicKey = pub
}

func parse25519Keypair(privBlock []byte) (ed25519.PrivateKey, ed25519.PublicKey) {
	block, _ := pem.Decode(privBlock)
	if block == nil {
		panic("failed to parse PEM block containing the private key")
	}

	p, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic("failed to parse blocks: " + err.Error())
	}

	priv, ok := p.(ed25519.PrivateKey)
	if !ok {
		panic("failed to cast parsed private key to type ed25519.PrivateKey")
	}

	pub, ok := priv.Public().(ed25519.PublicKey)
	if !ok {
		panic("failed to cast parsed public key to type ed25519.PublicKey")
	}

	return priv, pub
}
