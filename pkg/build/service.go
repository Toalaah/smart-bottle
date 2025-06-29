package build

import (
	"tinygo.org/x/bluetooth"
)

const (
	ServiceName    = "Smart Flask"
	ServiceVersion = "0.1"
	BackendAddr    = "http://localhost:8000"
	NonceLen       = 4
)

var (
	ServiceUUID                 = bluetooth.New32BitUUID(0xdeadbeef)
	CharacteristicUUIDFillLevel = bluetooth.New32BitUUID(0xcafebabe)
	CharacteristicUUIDAuth      = bluetooth.New32BitUUID(0xfefefefe)
	CharacteristicUUIDNonce     = bluetooth.New32BitUUID(0xf00dbabe)
	ManufacturerUUID            = uint16(0xc001)
)
