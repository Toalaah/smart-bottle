package main

import (
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
)

func main() {
	c := ble.NewClient(
		client.WithLogger(l),
	)
	must("init BLE client", c.Init())
	must("authenticate", c.Auth(secrets.PairingPin[:]))
	for msg := range c.Queue() {
		l.Debug("received message", "msg", msg)
		raw, err := crypto.DecryptEphemeralStaticX25519(msg.Value, secrets.UserPrivateKey)
		must("decrypt", err)
		depth = math.Float32frombits(binary.LittleEndian.Uint32(raw))
		l.Debug("decrypted message", "msg", fmt.Sprintf("%+v", raw), "depth", depth)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
