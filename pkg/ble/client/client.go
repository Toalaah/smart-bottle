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
	s.debug("connecting to device", "address", result.Address.String())
	device, err := s.adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		return err
	}

	serviceIDs := []bluetooth.UUID{build.ServiceUUID}
	s.debug("scanning for matching service", "device", device, "serviceIDs", serviceIDs)
	svcs, err := device.DiscoverServices(serviceIDs)
	if err != nil {
		return err
	}
	if len(svcs) == 0 {
		return fmt.Errorf("could not find matching service")
	}

	svc := svcs[0]
	s.debug("found service", "service", svc)
	s.debug("discovering service characteristics", "id", svc.UUID().String())
	characteristicIDs := []bluetooth.UUID{build.CharacteristicUUID}
	s.debug("scanning for matching characteristics", "service", svc.UUID().String(), "characteristicIDs", characteristicIDs)

	chars, err := svc.DiscoverCharacteristics(characteristicIDs)
	if err != nil {
		return err
	}
	if len(chars) == 0 {
		return fmt.Errorf("could not find characteristic")
	}

	for _, char := range chars {
		if char.UUID() == build.CharacteristicUUID {
			s.debug("found characteristic", "characteristicID", char.UUID().String())
			char.EnableNotifications(func(p []byte) {
				msg := transport.Message{}
				if err := transport.UnmarshalBytes(&msg, p); err != nil {
					panic(err)
				}
				s.c <- msg
			})
			s.debug("enabled notifications on characteristic", "characteristicID", char.UUID().String())
			break
		}
	}

	return nil
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
