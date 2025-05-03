package main

import (
	"fmt"

	"github.com/toalaah/smart-bottle/pkg/build"
	"tinygo.org/x/bluetooth"
)

var (
	adapter     = bluetooth.DefaultAdapter
	serviceUUID = build.ServiceUUID
)

func main() {
	println("enabling")

	must("enable BLE stack", adapter.Enable())

	ch := make(chan bluetooth.ScanResult, 1)
	println("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		println("found device:", result.LocalName())
		d := result.ManufacturerData()
		if result.LocalName() == build.ServiceName && len(d) > 0 && d[0].CompanyID == build.ManufacturerUUID {
			fmt.Printf("device %s has matching manufacturer UUID, stopping scan\n", result.LocalName())
			ch <- result
			adapter.StopScan()
		}
	})

	result := <-ch
	println("connecting to ", result.Address.String())
	device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	must("connect to device", err)

	println("discovering services")
	svcs, err := device.DiscoverServices([]bluetooth.UUID{build.ServiceUUID})
	must("discover services", err)

	if len(svcs) == 0 {
		panic("could not find matching service")
	}
	svc := svcs[0]
	println("found service ", svc.UUID().String())

	println("discovering characteristics")
	chars, err := svc.DiscoverCharacteristics([]bluetooth.UUID{build.CharacteristicUUID})
	must("discover characteristics", err)

	if len(chars) == 0 {
		panic("could not find characteristic")
	}

	// Setup RX channel
	c := make(chan []byte, 1)
	for _, char := range chars {
		if char.UUID() == build.CharacteristicUUID {
			println("found characteristic", char.UUID().String())
			char.EnableNotifications(func(buf []byte) {
				c <- buf
			})
			break
		}
	}

	// Wait for messages
	for {
		msg := <-c
		fmt.Printf("Data: %+v\n", msg)
	}
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
