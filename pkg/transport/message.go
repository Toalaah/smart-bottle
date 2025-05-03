package transport

type MessageType uint8

const (
	HeartBeat MessageType = 1 << iota
	Battery
	WaterLevel
	ConsumptionRate
)

type Message []byte
