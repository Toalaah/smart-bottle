package client

import (
	"fmt"
	"log/slog"

	"github.com/toalaah/smart-bottle/pkg/build"
	"github.com/toalaah/smart-bottle/pkg/transport"
	"tinygo.org/x/bluetooth"
)

type GattClient struct {
	adapter *bluetooth.Adapter
	logger  *slog.Logger
	c       chan transport.Message

	rxChar, authChar *bluetooth.DeviceCharacteristic
	device           bluetooth.Device
}

func New(opts ...ClientOption) *GattClient {
	s := &GattClient{
		adapter: bluetooth.DefaultAdapter,
		c:       make(chan transport.Message),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *GattClient) Init() error {
	s.debug("enabling adapter")
	if err := s.adapter.Enable(); err != nil {
		return err
	}

	s.debug("scanning...")
	devices := make(chan bluetooth.ScanResult, 1)
	err := s.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if result.LocalName() == "" {
			return
		}
		s.debug("found device", "name", result.LocalName())
		d := result.ManufacturerData()
		if result.LocalName() == build.ServiceName && len(d) > 0 && d[0].CompanyID == build.ManufacturerUUID {
			s.debug("device has matching manufacturer UUID", "name", result.LocalName(), "manufacturerData", d[0].CompanyID)
			devices <- result
			adapter.StopScan()
		}
	})
	if err != nil {
		return err
	}

	result := <-devices
	s.debug("connecting to device", "address", s.device.Address.String())
	s.device, err = s.adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		return err
	}

	serviceIDs := []bluetooth.UUID{build.ServiceUUID}
	s.debug("scanning for matching service", "device", s.device, "serviceIDs", serviceIDs)
	svcs, err := s.device.DiscoverServices(serviceIDs)
	if err != nil {
		return err
	}
	if len(svcs) == 0 {
		return fmt.Errorf("could not find matching service")
	}

	svc := svcs[0]
	s.debug("found service", "service", svc)
	s.debug("discovering service characteristics", "id", svc.UUID().String())
	characteristicIDs := []bluetooth.UUID{build.CharacteristicUUIDFillLevel, build.CharacteristicUUIDAuth}
	s.debug("scanning for matching characteristics", "service", svc.UUID().String(), "characteristicIDs", characteristicIDs)

	chars, err := svc.DiscoverCharacteristics(characteristicIDs)
	if err != nil {
		return err
	}
	if len(chars) == 0 {
		return fmt.Errorf("could not find characteristic")
	}

	for _, char := range chars {
		switch char.UUID() {
		case build.CharacteristicUUIDFillLevel:
			s.debug("found fill level characteristic", "characteristicID", char.UUID().String())
			s.rxChar = &char
		case build.CharacteristicUUIDAuth:
			s.debug("found auth characteristic", "characteristicID", char.UUID().String())
			s.authChar = &char
		}
	}

	if s.rxChar == nil || s.authChar == nil {
		return fmt.Errorf("could not discover all required characteristics")
	}

	s.rxChar.EnableNotifications(func(p []byte) {
		msg := transport.Message{}
		if err := transport.UnmarshalBytes(&msg, p); err != nil {
			panic(err)
		}
		s.c <- msg
	})

	return nil
}

func (s *GattClient) Auth(pin []byte) error {
	s.debug("performing authentication", "pin", fmt.Sprintf("%+v", pin))
	if s.authChar == nil {
		return fmt.Errorf("auth characteristic is nil")
	}
	_, err := s.authChar.WriteWithoutResponse(pin)
	return err
}

func (s *GattClient) Disconnect() error {
	s.debug("performing disconnect", "device", s.device)
	if s.device.Address == (bluetooth.Address{}) {
		return fmt.Errorf("device is nil")
	}
	return s.device.Disconnect()
}

func (s *GattClient) Queue() chan transport.Message {
	return s.c
}

func (s *GattClient) debug(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Debug(msg, args...)
	}
}

type ClientOption func(*GattClient)

func WithLogger(l *slog.Logger) ClientOption {
	return func(c *GattClient) {
		c.logger = l
	}
}
