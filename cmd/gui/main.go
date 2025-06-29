package main

import (
	"bytes"
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
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/toalaah/smart-bottle/pkg/ble"
	"github.com/toalaah/smart-bottle/pkg/ble/client"
	"github.com/toalaah/smart-bottle/pkg/build/secrets"
	"github.com/toalaah/smart-bottle/pkg/crypto"
)

var (
	currentFillLevel      float32            = 0
	currentFillPercentage float32            = 0
	l                                        = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	c                     *client.GattClient = nil
	isConnected           bool               = false
	isAuthed              bool               = false
	readings              ReadingsResponse

	connectButton = new(widget.Clickable)
	authKeyBuf    = new(bytes.Buffer)
	editor        = &widget.Editor{
		Submit:     true,
		ReadOnly:   false,
		SingleLine: true,
		Mask:       '*',
	}
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Change according to height of water bottle
var (
	bottleDepthMax float32 = 14.399 // Empty
	bottleDepthMin float32 = 3.700  // Full
)

const (
	title = "Smart Bottle Connect"
)

func main() {
	l.Debug("obtaining access token")
	err := Login()
	if err != nil {
		l.Error("failed to obtain access token", "error", err)
	}

	l.Debug("obtaining readings from api")
	r, err := GetReadings()
	readings = r
	if err != nil {
		l.Error("failed to obtain readings", "error", err)
	}
	l.Debug("got readings", "readings", readings)

	go setupBleClient()
	go func() {
		defer func() {
			if c != nil && isConnected {
				_ = c.Disconnect()
			}
		}()
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
			// ops.Reset()
			gtx := app.NewContext(&ops, e)
			drawLayout(gtx, theme)
			e.Frame(gtx.Ops)
		}
	}
}

func drawLayout(gtx C, th *material.Theme) D {

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(
			func(gtx C) D {
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
			func(gtx C) D {
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
			func(gtx C) D {
				if isAuthed && isConnected {
					return D{}
				}
				for {
					if _, ok := editor.Update(gtx); !ok {
						break
					}
				}
				inset := layout.Inset{Left: unit.Dp(200), Right: unit.Dp(200)}
				e := material.Editor(th, editor, "****")
				e.Font.Style = font.Italic
				border := widget.Border{Color: color.NRGBA{A: 0x7f}, CornerRadius: unit.Dp(0), Width: unit.Dp(2)}
				return inset.Layout(gtx, func(gtx C) D {
					return border.Layout(gtx, e.Layout)
				})
			},
		),
		layout.Rigid(
			// The height of the spacer is 25 Device independent pixels
			layout.Spacer{Height: unit.Dp(25)}.Layout,
		),
		layout.Rigid(
			// The height of the spacer is 25 Device independent pixels
			func(gtx C) D {
				inset := layout.Inset{Left: unit.Dp(200), Right: unit.Dp(200)}
				for connectButton.Clicked(gtx) {
					if isConnected && isAuthed {
						l.Info("disconnecting")
						if c != nil {
							c.Disconnect()
						}
						isConnected = false
						isAuthed = false
						break
					}
					contents := editor.Text()
					authKeyBuf.Reset()
					for _, c := range contents {
						n := uint8(c) - 48
						authKeyBuf.WriteByte(n)
					}
					l.Info("click registered", "key", fmt.Sprintf("%+v", authKeyBuf.Bytes()))
					go authBleClient()
				}
				return inset.Layout(gtx, func(gtx C) D {
					s := "Connect"
					if isConnected && isAuthed {
						s = "Disconnect"
					}
					return material.Button(th, connectButton, s).Layout(gtx)
				})
			},
		),
		layout.Rigid(
			// The height of the spacer is 25 Device independent pixels
			layout.Spacer{Height: unit.Dp(25)}.Layout,
		),
		layout.Rigid(
			func(gtx C) D {
				inset := layout.Inset{Left: 200, Right: 200}
				return inset.Layout(gtx, func(gtx C) D {
					if currentFillPercentage == 0 {
						return D{}
					}
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
		layout.Rigid(
			func(gtx C) D {
				if len(readings.Data) <= 0 {
					return D{}
				}
				list := layout.List{Axis: layout.Vertical, Alignment: layout.Middle}
				hdr := material.H5(th, "Past Fill Levels")
				hdr.Alignment = text.Middle
				hdr.Font.Weight = font.Bold

				inset := layout.Inset{Left: 200, Right: 200}
				return inset.Layout(gtx, func(gtx C) D {
					return list.Layout(gtx, len(readings.Data)+1, func(gtx layout.Context, index int) layout.Dimensions {
						if index == 0 {
							return hdr.Layout(gtx)
						}
						txt := material.H6(th, fmt.Sprintf("%+v", readings.Data[index-1]))
						txt.Alignment = text.Middle
						return txt.Layout(gtx)
					})
				})
			}),
		layout.Rigid(
			// The height of the spacer is 25 Device independent pixels
			layout.Spacer{Height: unit.Dp(250)}.Layout,
		),
	)
}

func authBleClient() {
	for c == nil && !isConnected {
		l.Info("ble client not initialized, waiting to auth")
		time.Sleep(time.Second)
		if isAuthed {
			return
		}
	}
	l.Debug("writing auth token", "pin", fmt.Sprintf("%+v", authKeyBuf.Bytes()))
	if err := c.Auth(authKeyBuf.Bytes()); err != nil {
		l.Error("auth error", "error", err)
	}
	isAuthed = true
}

func setupBleClient() {
	c = ble.NewClient(
		client.WithLogger(l),
	)
	if err := c.Init(); err != nil {
		l.Error("error while setting up ble client", "error", err)
		return
	}

	isConnected = true
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

		l.Debug("posting new reading to api")
		reading := Reading{Timestamp: time.Now(), Value: float64(currentFillLevel)}
		err = PostReading(reading)
		readings.Data = append(readings.Data, reading)
		if err != nil {
			l.Error("failed to post reading", "error", err)
		}
	}
}

func getFillPercentageFromDepth(d float32) float32 {
	if d >= bottleDepthMax {
		return 0
	}
	if d <= bottleDepthMin {
		return 1
	}
	// For the sake of simplicity we assume a linear relationship, that is that the bottle is a perfect cylinder
	return (d - bottleDepthMin) / (bottleDepthMax - bottleDepthMin)
}
