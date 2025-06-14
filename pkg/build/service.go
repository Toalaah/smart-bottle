package build

import (
	"tinygo.org/x/bluetooth"
)

const (
	ServiceName    = "Smart Flask"
	ServiceVersion = "0.1"
)

var (
	ServiceUUID                 = bluetooth.New32BitUUID(0xdeadbeef)
	CharacteristicUUIDFillLevel = bluetooth.New32BitUUID(0xcafebabe)
	CharacteristicUUIDAuth      = bluetooth.New32BitUUID(0xfefefefe)
	ManufacturerUUID            = uint16(0xc001)
	UserPin                     = []byte{1, 3, 3, 7}
)
