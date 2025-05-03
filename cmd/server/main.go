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
	if build.Debug {
		cdc.EnableUSBCDC()
	}

	time.Sleep(time.Second * 3)

	svc := NewService(
		WithLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))),
	)
	must("initialize BLE service", svc.Init())

	t := time.NewTicker(time.Second * 5)
	var randomValue uint8
	for {
		select {
		case <-t.C:
			randomValue = uint8(rand.IntN(100))
			println(fmt.Sprintf("New TX value: %d", randomValue))
			must("send successfully", svc.Send(randomValue))
		}
	}

}

func must(msg string, err error) {
	if err != nil {
		println(fmt.Sprintf("Failed to %s, halting execution: %s", msg, err))
		// By halting execution, we can somewhat cleanly reflash using tinygo's builtin flashing utility without having to manually re-enter BOOTSEL beforehand (assuming that USBCDC is enabled). This speeds up the development workflow significantly.
		select {}
	}
}
