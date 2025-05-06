package main

import (
	"fmt"
	"log/slog"
	"machine/usb/cdc"
	"math/rand/v2"
	"os"
	"time"

	"github.com/toalaah/smart-bottle/pkg/build"
	"github.com/toalaah/smart-bottle/pkg/build/secrets"
	"github.com/toalaah/smart-bottle/pkg/crypto"
	"github.com/toalaah/smart-bottle/pkg/transport"
)

func main() {
	var l *slog.Logger = nil
	if build.Debug {
		cdc.EnableUSBCDC()
		l = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	time.Sleep(time.Second * 3)
	svc := NewService(
		WithLogger(l),
		WithAdvertisementInterval(1250*time.Millisecond),
		WithTXBufferSize(64),
	)
	must("initialize BLE service", svc.Init())

	var randomValue uint8
	msg := &transport.Message{
		Type: transport.WaterLevel,
	}
	for {
		time.Sleep(time.Second * 5)
		randomValue = uint8(rand.IntN(100))
		println("New value: ", randomValue)
		// TODO: the public key used here should be the user's not the bottle's. For now, it serves as a placeholder.
		cipher, err := crypto.EncryptEphemeralStaticX25519([]byte{randomValue}, secrets.BottlePublicKey)
		msg.Value = cipher
		msg.Length = uint8(len(cipher))
		must("encrypt", err)
		must("send successfully", svc.SendMessage(msg))
	}

}

func must(msg string, err error) {
	if err != nil {
		println(fmt.Sprintf("Failed to %s, halting execution: %s", msg, err))
		// By halting execution, we can somewhat cleanly reflash using tinygo's builtin flashing utility without having to manually re-enter BOOTSEL beforehand (assuming that USBCDC is enabled). This speeds up the development workflow significantly.
		select {}
	}
}
