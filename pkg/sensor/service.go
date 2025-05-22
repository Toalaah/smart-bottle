package sensor

import (
	"fmt"
	"log/slog"
	"machine"
	"time"
)

type DepthSensorService struct {
	logger *slog.Logger
	u      *machine.UART
	d      time.Duration
	buf    []byte
}

func NewDepthSensorService(opts ...DepthSensorServiceOption) *DepthSensorService {
	s := &DepthSensorService{
		logger: nil,
		u:      machine.DefaultUART,
		d:      time.Millisecond * 100,
		buf:    make([]byte, 4),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *DepthSensorService) Init() error {
	s.debug("init depth sensor")
	s.u.Configure(machine.UARTConfig{BaudRate: 9600})
	return nil
}

func (s *DepthSensorService) Read() (float32, error) {
	s.debug("reading from depth sensor")
	// read 4 bytes (w/ offset marker of 0xff)
	for {
		s.u.Read(s.buf)
		b, _ := s.u.ReadByte()
		if b != 0xff && s.buf[0] == 0xff {
			break
		}
		time.Sleep(s.d)
	}

	s.debug("read packet", "value", s.buf)
	sum := (s.buf[0] + s.buf[1] + s.buf[2]) & 0xff
	if sum != s.buf[3] {
		return 0, fmt.Errorf("incorrect checksum")
	}

	// distance = s.buf[1] * 256 + s.buf[2]
	d := float32(s.buf[1]<<8 + s.buf[2])
	// min range of 3cm
	if d > 30 {
		d = d / 10
		return d, nil
	}
	return 0, nil
}

func (s *DepthSensorService) debug(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Debug(msg, args...)
	}
}

type DepthSensorServiceOption func(*DepthSensorService)

func WithRetryDelay(d time.Duration) DepthSensorServiceOption {
	return func(s *DepthSensorService) {
		s.d = d
	}
}

func WithLogger(logger *slog.Logger) DepthSensorServiceOption {
	return func(s *DepthSensorService) {
		s.logger = logger
	}
}
