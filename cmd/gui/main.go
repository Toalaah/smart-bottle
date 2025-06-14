package main

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"math"
	"os"
	"time"

	"log/slog"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/toalaah/smart-bottle/pkg/ble"
	"github.com/toalaah/smart-bottle/pkg/ble/client"
	"github.com/toalaah/smart-bottle/pkg/build"
	"github.com/toalaah/smart-bottle/pkg/build/secrets"
	"github.com/toalaah/smart-bottle/pkg/crypto"
)

var (
	currentFillLevel      float32            = 0
	currentFillPercentage float32            = 0
	l                                        = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	c                     *client.GattClient = nil
)

// Change according to height of water bottle
var (
	bottleDepthMax float32 = 7.4 // Empty
	bottleDepthMin float32 = 3.2 // Full
)

const (
	title = "Smart Bottle Connect"
)

func main() {
	go setupBleClient()
	go func() {
		window := new(app.Window)
		window.Option(app.Title(title))
		if err := run(window); err != nil {
			l.Error("error while running main loop", "error", err)
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
							l.Info("exiting")
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
				txt := material.H3(th, fmt.Sprintf("Fill level: %.2f%%", currentFillPercentage*100))
				txt.Alignment = text.Middle
				txt.Font.Weight = font.Bold
				return txt.Layout(gtx)
			},
		),
		layout.Rigid(
			// The height of the spacer is 25 Device independent pixels
			layout.Spacer{Height: unit.Dp(25)}.Layout,
		),
		layout.Rigid(
			func(gtx layout.Context) layout.Dimensions {
				inset := layout.Inset{Left: 200, Right: 200}
				return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					fillWidget := material.ProgressCircle(th, currentFillPercentage)
					inv := op.InvalidateCmd{At: gtx.Now.Add(time.Second / 25)}
					gtx.Execute(inv)
					return fillWidget.Layout(gtx)
				})
			}),
		layout.Rigid(
			// The height of the spacer is 25 Device independent pixels
			layout.Spacer{Height: unit.Dp(250)}.Layout,
		),
	)
}

func setupBleClient() {
	c = ble.NewClient(
		client.WithLogger(l),
	)
	if err := c.Init(); err != nil {
		l.Error("error while setting up ble client", "error", err)
		os.Exit(1)
	}
	// TODO: make this an interactive entry in the GUI
	if err := c.Auth(build.UserPin); err != nil {
		l.Error("auth error", "error", err)
		os.Exit(1)
	}
	for msg := range c.Queue() {
		l.Debug("received message", "msg", msg)
		raw, err := crypto.DecryptEphemeralStaticX25519(msg.Value, secrets.UserPrivateKey)
		if err != nil {
			l.Error("error while decrypting payload", "error", err, "msg", msg)
			continue
		}
		d := math.Float32frombits(binary.LittleEndian.Uint32(raw))
		l.Debug("decrypted message", "msg", fmt.Sprintf("%+v", raw), "fillLevel", d)
		currentFillPercentage = getFillPercentageFromDepth(d)
		currentFillLevel = d
	}
}

func getFillPercentageFromDepth(d float32) float32 {
	if d < bottleDepthMin || d > bottleDepthMax { // Assume failed/invalid reading, if so we just reuse the last fill level
		d = currentFillLevel
	}
	if d == 0 {
		return 0
	}
	// For the sake of simplicity we assume a linear relationship, that is that the bottle is a perfect cylinder
	return (d - bottleDepthMin) / (bottleDepthMax - bottleDepthMin)
}
