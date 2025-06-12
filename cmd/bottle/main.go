package main

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"machine/usb/cdc"
	"math"
	"os"
	"time"

	"github.com/toalaah/smart-bottle/pkg/ble"
	"github.com/toalaah/smart-bottle/pkg/build"
	"github.com/toalaah/smart-bottle/pkg/build/secrets"
	"github.com/toalaah/smart-bottle/pkg/crypto"
	"github.com/toalaah/smart-bottle/pkg/sensor"
	"github.com/toalaah/smart-bottle/pkg/transport"
)

var (
	l               *slog.Logger = nil
	fillLevel       float32
	err             error
	buf             = make([]byte, 4)
	publishInterval = time.Second * 3
	msg             = &transport.Message{Type: transport.WaterLevel}
)

func main() {
	if build.Debug {
		cdc.EnableUSBCDC()
		l = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	time.Sleep(publishInterval)

	svc := ble.NewService(
		ble.WithLogger(l),
		ble.WithAdvertisementInterval(1250*time.Millisecond),
		ble.WithTXBufferSize(66), // type + length + 64 bytes payload
	)
	must("initialize BLE service", svc.Init())

	depthSensor := sensor.NewDepthSensorService(
		sensor.WithLogger(l),
	)
	must("initialize depth sensor", depthSensor.Init())

	for {
		time.Sleep(publishInterval)

		fillLevel, err = depthSensor.Read()
		if err != nil {
			l.Error("error reading fill level", "error", err)
		} else {
			l.Debug("read fill level", "level", fillLevel)
		}

		binary.LittleEndian.PutUint32(buf, math.Float32bits(fillLevel))
		cipher, err := crypto.EncryptEphemeralStaticX25519(buf, secrets.UserPublicKey)
		must("encrypt", err)

		msg.Load(cipher)
		must("send successfully", svc.SendMessage(msg))
	}
}

func must(msg string, err error) {
	if err != nil {
		println(fmt.Sprintf("failed to %s, halting execution: %s", msg, err))
		// By halting execution, we can somewhat cleanly reflash using tinygo's builtin flashing utility without having to manually re-enter BOOTSEL beforehand (assuming that USBCDC is enabled). This speeds up the development workflow significantly.
		select {}
	}
}
