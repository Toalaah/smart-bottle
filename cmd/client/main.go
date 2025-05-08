package main

import (
	"log/slog"
	"os"

	"github.com/toalaah/smart-bottle/pkg/build"
	"github.com/toalaah/smart-bottle/pkg/build/secrets"
	"github.com/toalaah/smart-bottle/pkg/crypto"
	"github.com/toalaah/smart-bottle/pkg/transport"
	"tinygo.org/x/bluetooth"
)

var (
	adapter     = bluetooth.DefaultAdapter
	serviceUUID = build.ServiceUUID
	log         = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
)

func main() {
	log.Debug("enabling")

	must("enable BLE stack", adapter.Enable())

	devices := make(chan bluetooth.ScanResult, 1)
	log.Debug("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if result.LocalName() == "" {
			return
		}
		log.Debug("found device", "name", result.LocalName())
		d := result.ManufacturerData()
		if result.LocalName() == build.ServiceName && len(d) > 0 && d[0].CompanyID == build.ManufacturerUUID {
			log.Debug("device has matching manufacturer UUID", "name", result.LocalName(), "manufacturerData", d[0].CompanyID)
			devices <- result
			adapter.StopScan()
		}
	})

	result := <-devices
	log.Debug("connecting to device", "address", result.Address.String())
	device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	must("connect to device", err)

	serviceIDs := []bluetooth.UUID{build.ServiceUUID}
	log.Debug("scanning for matching service", "device", device, "serviceIDs", serviceIDs)
	svcs, err := device.DiscoverServices(serviceIDs)
	must("discover services", err)

	if len(svcs) == 0 {
		log.Error("could not find any matching service", "device", device, "serviceIDs", serviceIDs)
		os.Exit(1)
	}

	svc := svcs[0]
	log.Debug("found service", "service", svc)
	log.Debug("discovering service characteristics", "id", svc.UUID().String())
	characteristicIDs := []bluetooth.UUID{build.CharacteristicUUID}
	log.Debug("scanning for matching characteristics", "service", svc.UUID().String(), "characteristicIDs", characteristicIDs)
	chars, err := svc.DiscoverCharacteristics(characteristicIDs)
	must("discover characteristics", err)

	if len(chars) == 0 {
		log.Error("could not find characteristic", "service", svc.UUID().String(), "characteristicIDs", characteristicIDs)
		os.Exit(1)
	}

	// Setup RX channel
	rx := make(chan []byte, 1)
	for _, char := range chars {
		if char.UUID() == build.CharacteristicUUID {
			println("found characteristic", char.UUID().String())
			char.EnableNotifications(func(buf []byte) {
				rx <- buf
			})
			break
		}
	}

	// Wait for messages
	msgBuf := &transport.Message{}
	for {
		payload := <-rx
		log.Debug("received message", "data", payload)
		transport.UnmarshalBytes(msgBuf, payload)
		msg, err := crypto.DecryptEphemeralStaticX25519(msgBuf.Value, secrets.UserPrivateKey)
		must("decrypt", err)
		log.Debug("decrypted message", "msg", msg)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
