package main

import (
	"errors"
	"image/color"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/toalaah/smart-bottle/pkg/build"

	"log/slog"

	"tinygo.org/x/bluetooth"
)

var (
	currentFillLevel float32 = 0
	fillLevelChan            = make(chan []byte, 1)
	log                      = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
)

const (
	title = "Smart Bottle Connect"
)

func setupBleClient() error {
	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return err
	}
	log.Info("enabled adapter", "adapter", adapter)

	devices := make(chan bluetooth.ScanResult, 1)
	log.Debug("scanning...")
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if result.LocalName() == "" {
			return
		}
		log.Debug("found device", "name", result.LocalName())
		d := result.ManufacturerData()
		if result.LocalName() == build.ServiceName /* && len(d) > 0 && d[0].CompanyID == build.ManufacturerUUID */ {
			log.Debug("device has matching manufacturer UUID", "name", result.LocalName(), "manufacturerData", d[0].CompanyID)
			devices <- result
			adapter.StopScan()
		}
	})

	result := <-devices
	log.Debug("connecting to device", "address", result.Address.String())
	device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		return err
	}

	serviceIDs := []bluetooth.UUID{bluetooth.ServiceUUIDBloodPressure}
	log.Debug("scanning for matching service", "device", device, "serviceIDs", serviceIDs)
	svcs, err := device.DiscoverServices(serviceIDs)
	if err != nil {
		return err
	}
	log.Debug("got services", "device", device, "services", svcs)

	if len(svcs) == 0 {
		return errors.New("could not find matching service")
	}

	svc := svcs[0]
	log.Debug("found service", "service", svc)
	log.Debug("discovering service characteristics", "id", svc.UUID().String())
	characteristicIDs := []bluetooth.UUID{bluetooth.CharacteristicUUIDBloodPressureFeature}
	log.Debug("scanning for matching characteristics", "service", svc.UUID().String(), "characteristicIDs", characteristicIDs)
	chars, err := svc.DiscoverCharacteristics(characteristicIDs)
	if err != nil {
		return err
	}

	if len(chars) == 0 {
		return errors.New("could not find matching characteristic")
	}

	for _, char := range chars {
		if char.UUID() == build.CharacteristicUUID {
			log.Info("found characteristic", "uuid", char.UUID().String())
			char.EnableNotifications(func(buf []byte) {
				fillLevelChan <- buf
			})
			break
		}
	}

	return nil
}

func main() {
	go func() {
		for {
			time.Sleep(time.Millisecond * 100)
			currentFillLevel += 0.01
			if currentFillLevel > 1 {
				currentFillLevel = 0
			}
		}
	}()
	go func() {
		window := new(app.Window)
		window.Option(app.Title(title))
		if err := setupBleClient(); err != nil {
			log.Error("error while setting up ble client", "error", err)
			os.Exit(1)
		}
		if err := run(window); err != nil {
			log.Error("error while running main loop", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window) error {
	theme := material.NewTheme()
	var ops op.Ops
	for {
		event := window.Event()
		switch e := event.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// Handle key events
			if e, ok := e.Source.Event(key.Filter{}); ok {
				switch e := e.(type) {
				case key.Event:
					if e.State == key.Press {
						switch e.Name {
						case "Q", key.NameEscape:
							log.Info("exiting")
							return nil
						}
					}
				}
			}
			ops.Reset()
			gtx := app.NewContext(&ops, e)
			drawLayout(gtx, theme)
			e.Frame(gtx.Ops)
		}
	}
}

func drawLayout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(
			func(gtx layout.Context) layout.Dimensions {
				title := material.H2(th, "Smart Bottle Tracker")
				black := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
				title.Color = black
				title.Alignment = text.Middle
				return title.Layout(gtx)
			},
		),
		layout.Rigid(
			// The height of the spacer is 25 Device independent pixels
			layout.Spacer{Height: unit.Dp(25)}.Layout,
		),
		layout.Rigid(
			func(gtx layout.Context) layout.Dimensions {
				fillWidget := material.ProgressCircle(th, currentFillLevel)
				inv := op.InvalidateCmd{At: gtx.Now.Add(time.Second / 25)}
				gtx.Execute(inv)
				return fillWidget.Layout(gtx)
			},
		),
	)
}
