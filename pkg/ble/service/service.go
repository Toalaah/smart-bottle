package service

import (
	"bytes"
	"fmt"
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

	txHnd, authHnd               bluetooth.Characteristic
	txBufSize, authKeySize       uint32
	authEnabled, didAuthenticate bool

	connectedDevice chan bluetooth.Device
}

func New(opts ...ServiceOption) *GattService {
	s := &GattService{
		adapter:         bluetooth.DefaultAdapter,
		txHnd:           bluetooth.Characteristic{},
		logger:          nil,
		advInterval:     1000 * time.Millisecond,
		txBufSize:       128,
		authKeySize:     uint32(len(build.UserPin)),
		authEnabled:     false,
		didAuthenticate: false,
		connectedDevice: make(chan bluetooth.Device, 1),
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

	s.adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		if connected {
			s.debug("new device connection", "state", connected, "device", device)
			select {
			case s.connectedDevice <- device:
			default:
			}
		} else {
			s.debug("resetting auth")
			s.didAuthenticate = false
		}
	})

	go func() {
		for device := range s.connectedDevice {
			time.Sleep(time.Second * 5)
			device.RequestConnectionParams(bluetooth.ConnectionParams{
				MinInterval: bluetooth.NewDuration(495 * time.Millisecond),
				MaxInterval: bluetooth.NewDuration(510 * time.Millisecond),
				Timeout:     bluetooth.NewDuration(5 * time.Second),
			})
		}
	}()

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
					UUID:   build.CharacteristicUUIDFillLevel,
					Value:  make([]byte, s.txBufSize),
					Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicNotifyPermission,
				},
			},
		},
	}

	if s.authEnabled {
		services[1].Characteristics = append(services[1].Characteristics, bluetooth.CharacteristicConfig{
			Handle: &s.authHnd,
			UUID:   build.CharacteristicUUIDAuth,
			Value:  make([]byte, s.authKeySize),
			Flags:  bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
			WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
				s.debug("received write event", "value", fmt.Sprintf("%+v", value))
				if l := len(value); l != int(s.authKeySize) {
					s.debug("auth key has unexpected length", "length", l)
				}
				if bytes.Compare(value, build.UserPin) == 0 {
					s.debug("auth succeeded")
					s.didAuthenticate = true
				}
			},
		})
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
	if !s.didAuthenticate && s.authEnabled {
		s.debug("no authentication handshake has taken place, skipping sending message")
		return nil
	}
	s.debug("writing value", "handle", s.txHnd, "length", 2+m.Length)
	if _, err := s.txHnd.Write(append([]byte{uint8(m.Type), m.Length}, m.Value...)); err != nil {
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

func WithLogger(l *slog.Logger) ServiceOption {
	return func(s *GattService) {
		s.logger = l.With("service", "gatt")
	}
}

func WithAdvertisementInterval(d time.Duration) ServiceOption {
	return func(s *GattService) {
		s.advInterval = d
	}
}

func WithTXBufferSize(n uint32) ServiceOption {
	return func(s *GattService) {
		s.txBufSize = n
	}
}

func WithAuth(enable bool) ServiceOption {
	return func(s *GattService) {
		s.authEnabled = enable
	}
}

func WithAuthKeySize(n uint32) ServiceOption {
	return func(s *GattService) {
		s.authKeySize = n
	}
}
