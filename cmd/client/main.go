package main

import (
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"os"

	"github.com/toalaah/smart-bottle/pkg/ble"
	"github.com/toalaah/smart-bottle/pkg/ble/client"
	"github.com/toalaah/smart-bottle/pkg/build"
	"github.com/toalaah/smart-bottle/pkg/build/secrets"
	"github.com/toalaah/smart-bottle/pkg/crypto"
	"tinygo.org/x/bluetooth"
)

var (
	adapter     = bluetooth.DefaultAdapter
	serviceUUID = build.ServiceUUID
	l           = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	depth       float32
	gcm         cipher.AEAD
	buf         = [4]byte{}
)

func main() {
	c := ble.NewClient(
		client.WithLogger(l),
	)
	must("init BLE client", c.Init())
	key, err := c.Auth(secrets.PairingPin[:])
	must("authenticate", err)

	gcm, err := crypto.NewGCM(key)
	must("init gcm", err)

	for msg := range c.Queue() {
		l.Debug("received message", "msg", msg)
		err := crypto.DecryptAES(gcm, msg.Value, buf[:])
		if err != nil {
			l.Error("failed to decrypt", "error", err)
		}
		depth = math.Float32frombits(binary.LittleEndian.Uint32(buf[:]))
		l.Debug("decrypted message", "msg", fmt.Sprintf("%+v", buf[:]), "depth", depth)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
