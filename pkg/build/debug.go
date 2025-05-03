package build

import "tinygo.org/x/bluetooth"

const (
	Debug       = true
	ServiceName = "Smart Flask"
)

var (
	ServiceUUID               = bluetooth.New32BitUUID(0xdeadbeef)
	CharacteristicUUID        = bluetooth.New32BitUUID(0xcafebabe)
	ManufacturerUUID   uint16 = 0xc001
)
