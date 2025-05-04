package main

import (
	"log/slog"
	"time"

	"github.com/toalaah/smart-bottle/pkg/build"
	"tinygo.org/x/bluetooth"
)

type Service struct {
	service     *bluetooth.Service
	adapter     *bluetooth.Adapter
	logger      *slog.Logger
	advInterval time.Duration

	txHnd bluetooth.Characteristic
	buf   [4]byte
}

func NewService(opts ...ServiceOption) *Service {
	s := &Service{
		adapter:     bluetooth.DefaultAdapter,
		txHnd:       bluetooth.Characteristic{},
		logger:      nil,
		advInterval: 1000 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) Init() error {
	s.debug("enabling adapter")
	if err := s.adapter.Enable(); err != nil {
		return err
	}
	mac, err := s.adapter.Address()
	if err != nil {
		return err
	}
	s.debug("have adapter address", "address", mac)

	services := []bluetooth.Service{
		// Device/vendor information
		bluetooth.Service{
			UUID: bluetooth.ServiceUUIDDeviceInformation,
			Characteristics: []bluetooth.CharacteristicConfig{
				{
					UUID:  bluetooth.CharacteristicUUIDManufacturerNameString,
					Value: []byte(build.ServiceName),
					Flags: bluetooth.CharacteristicReadPermission,
				},
				{
					UUID:  bluetooth.CharacteristicUUIDFirmwareRevisionString,
					Value: []byte(build.ServiceVersion),
					Flags: bluetooth.CharacteristicReadPermission,
				},
			},
		},
		// Main transport service
		bluetooth.Service{
			UUID: build.ServiceUUID,
			Characteristics: []bluetooth.CharacteristicConfig{
				{
					Handle: &s.txHnd,
					UUID:   build.CharacteristicUUID,
					Value:  []byte{0},
					Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicNotifyPermission,
				},
			},
		},
	}

	for _, service := range services {
		s.debug("adding service", "id", service.UUID)
		if err := s.adapter.AddService(&service); err != nil {
			return err
		}
	}

	adv := s.adapter.DefaultAdvertisement()
	advOpts := bluetooth.AdvertisementOptions{
		LocalName: build.ServiceName,
		Interval:  bluetooth.NewDuration(s.advInterval),
		ServiceUUIDs: []bluetooth.UUID{
			build.ServiceUUID,
			bluetooth.ServiceUUIDDeviceInformation,
		},
		ManufacturerData: []bluetooth.ManufacturerDataElement{
			bluetooth.ManufacturerDataElement{
				CompanyID: build.ManufacturerUUID,
			},
		},
	}
	s.debug("configuring advertisement", "name", advOpts.LocalName, "address", mac)
	if err := adv.Configure(advOpts); err != nil {
		return err
	}

	s.debug("starting advertisement", "name", advOpts.LocalName, "address", mac)
	if err := adv.Start(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Send(v uint8) error {
	s.debug("writing value", "value", v, "handle", s.txHnd)
	s.buf[0] = v
	if _, err := s.txHnd.Write(s.buf[:1]); err != nil {
		return err
	}
	return nil
}

func (s *Service) debug(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Debug(msg, args...)
	}
}

type ServiceOption func(*Service)

func WithLogger(logger *slog.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = logger
	}
}

func WithAdvertisementInterval(interval time.Duration) ServiceOption {
	return func(s *Service) {
		s.advInterval = interval
	}
}
