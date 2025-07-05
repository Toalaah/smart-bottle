package main

import (
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"log/slog"
	"machine/usb/cdc"
	"math"
	"os"
	"time"

	"github.com/toalaah/smart-bottle/pkg/ble"
	"github.com/toalaah/smart-bottle/pkg/ble/service"
	"github.com/toalaah/smart-bottle/pkg/build"
	"github.com/toalaah/smart-bottle/pkg/crypto"
	"github.com/toalaah/smart-bottle/pkg/sensor"
	"github.com/toalaah/smart-bottle/pkg/transport"
)

var (
	l               *slog.Logger = nil
	fillLevel       float32
	err             error
	fillBuf         = [4]byte{}
	out             = [32]byte{}
	publishInterval = time.Second * 3
	msg             = &transport.Message{Type: transport.WaterLevel}
	gcm             cipher.AEAD
)

func main() {
	if build.Debug {
		cdc.EnableUSBCDC()
		l = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	time.Sleep(time.Second * 5)

	svc := ble.NewService(
		service.WithLogger(l),
		service.WithAdvertisementInterval(1250*time.Millisecond),
		service.WithTXBufferSize(34), // Type + length + 32 bytes payload
		service.WithAuth(true),
	)
	must("initialize BLE service", svc.Init())

	depthSensor := sensor.NewDepthSensorService(
		sensor.WithLogger(l),
		sensor.WithMaxReadAttempts(200), // Total read timeout of 20 seconds.
		sensor.WithRetryDelay(time.Millisecond*100),
	)
	must("initialize depth sensor", depthSensor.Init())

	// Perform a budget "TLS" connection (basically just a DH handshake).
	// Wait until client has paired and authenticated.
	key := svc.GetPairingKeyBlocking()
	// Derive symmetric encryption channel.
	gcm, err = crypto.NewGCM(key)
	must("init gcm", err)

	for {
		time.Sleep(publishInterval)

		fillLevel, err = depthSensor.Read()
		if err != nil {
			l.Error("error reading fill level", "error", err)
		} else {
			l.Debug("read fill level", "level", fillLevel)
		}

		binary.LittleEndian.PutUint32(fillBuf[:], math.Float32bits(fillLevel))
		must("encrypt successfully", crypto.EncryptAES(gcm, fillBuf[:], out[:]))

		msg.Load(out[:])
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
