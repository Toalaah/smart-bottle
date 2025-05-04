package main

import (
	"fmt"
	"log/slog"
	"machine/usb/cdc"
	"math/rand/v2"
	"os"
	"time"

	"github.com/toalaah/smart-bottle/pkg/build"
)

func main() {
	var l *slog.Logger = nil
	if build.Debug {
		cdc.EnableUSBCDC()
		l = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	time.Sleep(time.Second * 3)
	svc := NewService(
		WithLogger(l),
		WithAdvertisementInterval(1250*time.Millisecond),
	)
	must("initialize BLE service", svc.Init())

	for {
		time.Sleep(time.Second * 5)
		must("send successfully", svc.Send(uint8(rand.IntN(100))))
	}

}

func must(msg string, err error) {
	if err != nil {
		println(fmt.Sprintf("Failed to %s, halting execution: %s", msg, err))
		// By halting execution, we can somewhat cleanly reflash using tinygo's builtin flashing utility without having to manually re-enter BOOTSEL beforehand (assuming that USBCDC is enabled). This speeds up the development workflow significantly.
		select {}
	}
}
