package main

import (
	"log/slog"
	"time"

	"github.com/toalaah/smart-bottle/pkg/build"
	"tinygo.org/x/bluetooth"
)

var (
	serviceUUID        = build.ServiceUUID
	characteristicUUID = build.CharacteristicUUID
	serviceName        = build.ServiceName
)

type Service struct {
	service *bluetooth.Service
	adapter *bluetooth.Adapter
	logger  *slog.Logger
	conn    chan bluetooth.Device

	txHnd bluetooth.Characteristic

	buf [4]byte
}

func NewService(opts ...ServiceOption) *Service {
	s := &Service{
		adapter: bluetooth.DefaultAdapter,
		txHnd:   bluetooth.Characteristic{},
		logger:  nil,
		conn:    make(chan bluetooth.Device, 1),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) Init() error {
	if err := s.adapter.Enable(); err != nil {
		return err
	}

	services := []bluetooth.Service{
		// Device/vendor information
		bluetooth.Service{
			UUID: bluetooth.ServiceUUIDDeviceInformation,
			Characteristics: []bluetooth.CharacteristicConfig{
				{
					UUID:  bluetooth.CharacteristicUUIDManufacturerNameString,
					Value: []byte(serviceName),
					Flags: bluetooth.CharacteristicReadPermission,
				},
				{
					UUID:  bluetooth.CharacteristicUUIDFirmwareRevisionString,
					Value: []byte("1.0"),
					Flags: bluetooth.CharacteristicReadPermission,
				},
			},
		},
		// Main transport service
		bluetooth.Service{
			UUID: serviceUUID,
			Characteristics: []bluetooth.CharacteristicConfig{
				{
					Handle: &s.txHnd,
					UUID:   characteristicUUID,
					Value:  []byte{0},
					Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicNotifyPermission,
				},
			},
		},
	}

	for _, service := range services {
		if err := s.adapter.AddService(&service); err != nil {
			return err
		}
	}

	adv := s.adapter.DefaultAdvertisement()
	advOpts := bluetooth.AdvertisementOptions{
		LocalName: serviceName,
		Interval:  bluetooth.NewDuration(1285 * time.Millisecond),
		ServiceUUIDs: []bluetooth.UUID{
			serviceUUID,
			bluetooth.ServiceUUIDDeviceInformation,
		},
		ManufacturerData: []bluetooth.ManufacturerDataElement{
			bluetooth.ManufacturerDataElement{
				CompanyID: build.ManufacturerUUID,
			},
		},
	}
	if err := adv.Configure(advOpts); err != nil {
		return err
	}

	if err := adv.Start(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Send(v uint8) error {
	s.buf[0] = v
	if _, err := s.txHnd.Write(s.buf[:1]); err != nil {
		return err
	}
	return nil
}

type ServiceOption func(*Service)

func WithLogger(logger *slog.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}
