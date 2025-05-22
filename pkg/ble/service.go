package ble

import (
	"log/slog"
	"time"

	"github.com/toalaah/smart-bottle/pkg/build"
	"github.com/toalaah/smart-bottle/pkg/transport"
	"tinygo.org/x/bluetooth"
)

type GattService struct {
	service     *bluetooth.Service
	adapter     *bluetooth.Adapter
	logger      *slog.Logger
	advInterval time.Duration

	txHnd     bluetooth.Characteristic
	txBufSize uint32
}

func NewService(opts ...ServiceOption) *GattService {
	s := &GattService{
		adapter:     bluetooth.DefaultAdapter,
		txHnd:       bluetooth.Characteristic{},
		logger:      nil,
		advInterval: 1000 * time.Millisecond,
		txBufSize:   128,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *GattService) Init() error {
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
					Value:  make([]byte, s.txBufSize),
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

func (s *GattService) SendMessage(m *transport.Message) error {
	s.debug("writing value", "handle", s.txHnd, "length", m.Length)
	if _, err := s.txHnd.Write(append([]byte{m.Length}, m.Value...)); err != nil {
		return err
	}
	return nil
}

func (s *GattService) Send(payload []byte) error {
	s.debug("writing value", "handle", s.txHnd, "length", len(payload))
	if _, err := s.txHnd.Write(payload); err != nil {
		return err
	}
	return nil
}

func (s *GattService) debug(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Debug(msg, args...)
	}
}

type ServiceOption func(*GattService)

func WithLogger(logger *slog.Logger) ServiceOption {
	return func(s *GattService) {
		s.logger = logger
	}
}

func WithAdvertisementInterval(interval time.Duration) ServiceOption {
	return func(s *GattService) {
		s.advInterval = interval
	}
}

func WithTXBufferSize(n uint32) ServiceOption {
	return func(s *GattService) {
		s.txBufSize = n
	}
}
