package build

import (
	"tinygo.org/x/bluetooth"
)

const (
	ServiceName    = "Smart Flask"
	ServiceVersion = "0.1"
)

var (
	ServiceUUID               = bluetooth.New32BitUUID(0xdeadbeef)
	CharacteristicUUID        = bluetooth.New32BitUUID(0xcafebabe)
	ManufacturerUUID   uint16 = 0xc001
)
