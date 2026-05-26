package main

import (
	"bufio"
	"context"
	"encoding"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alia5/VIIPER/device/ns2pro"
	"github.com/Alia5/VIIPER/viiperclient"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: virtual_ns2pro <api_addr>")
		fmt.Println("Example: virtual_ns2pro localhost:3242")
		os.Exit(1)
	}

	addr := os.Args[1]
	ctx := context.Background()
	api := viiperclient.New(addr)

	busID, createdBus, err := findOrCreateBus(ctx, api)
	if err != nil {
		fmt.Printf("Bus setup error: %v\n", err)
		os.Exit(1)
	}

	stream, addResp, err := api.AddDeviceAndConnect(ctx, busID, "ns2pro", nil)
	if err != nil {
		fmt.Printf("AddDeviceAndConnect error: %v\n", err)
		if createdBus {
			_, _ = api.BusRemoveCtx(ctx, busID)
		}
		os.Exit(1)
	}
	defer stream.Close() // nolint

	fmt.Printf("Created and connected to Switch 2 Pro device %s on bus %d\n", addResp.DevID, addResp.BusID)

	defer func() {
		if _, err := api.DeviceRemoveCtx(ctx, stream.BusID, stream.DevID); err != nil {
			fmt.Printf("DeviceRemove error: %v\n", err)
		} else {
			fmt.Printf("Removed device %d-%s\n", addResp.BusID, addResp.DevID)
		}
		if createdBus {
			if _, err := api.BusRemoveCtx(ctx, busID); err != nil {
				fmt.Printf("BusRemove error: %v\n", err)
			} else {
				fmt.Printf("Removed bus %d\n", busID)
			}
		}
	}()

	rumbleCh, errCh := stream.StartReading(ctx, 10, func(r *bufio.Reader) (encoding.BinaryUnmarshaler, error) {
		var b [ns2pro.OutputWireSize]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, err
		}
		msg := new(ns2pro.OutputState)
		if err := msg.UnmarshalBinary(b[:]); err != nil {
			return nil, err
		}
		return msg, nil
	})

	go func() {
		for {
			select {
			case msg := <-rumbleCh:
				if msg == nil {
					continue
				}
				rumble := msg.(*ns2pro.OutputState)
				fmt.Printf("[Output] HD rumble L=% x R=% x\n", rumble.LeftRumble[:6], rumble.RightRumble[:6])
			case err := <-errCh:
				if err != nil {
					fmt.Printf("[Output read error] %v\n", err)
				}
				return
			}
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	fmt.Println("Switch 2 Pro device active. Press Ctrl+C to exit.")
	fmt.Println("Demo: left stick circles, face buttons rotate, gyro/accel drift gently.")

	var frame uint64
	for {
		select {
		case <-ticker.C:
			frame++
			angle := float64(frame) * 0.045

			state := ns2pro.InputState{
				Buttons:       demoButtons(frame),
				LX:            stickAxis(math.Cos(angle)),
				LY:            stickAxis(math.Sin(angle)),
				RX:            ns2pro.StickCenter,
				RY:            ns2pro.StickCenter,
				AccelX:        int16(800 * math.Sin(angle*0.5)),
				AccelY:        int16(800 * math.Cos(angle*0.5)),
				AccelZ:        -4096,
				GyroX:         int16(500 * math.Sin(angle)),
				GyroY:         int16(500 * math.Cos(angle)),
				GyroZ:         int16(250 * math.Sin(angle*0.33)),
				BatteryLevel:  ns2pro.BatteryMax,
				ExternalPower: true,
			}

			if err := stream.WriteBinary(&state); err != nil {
				fmt.Printf("Send error: %v\n", err)
				return
			}

			if frame%60 == 0 {
				fmt.Printf("→ Sent frame %d buttons=0x%06x LX=%04x LY=%04x Gyro=(%d,%d,%d)\n",
					frame, state.Buttons, state.LX, state.LY, state.GyroX, state.GyroY, state.GyroZ)
			}

		case <-sigCh:
			fmt.Println("\nShutting down...")
			return
		}
	}
}

func findOrCreateBus(ctx context.Context, api *viiperclient.Client) (uint32, bool, error) {
	busesResp, err := api.BusListCtx(ctx)
	if err != nil {
		return 0, false, err
	}

	if len(busesResp.Buses) == 0 {
		r, err := api.BusCreateCtx(ctx, 0)
		if err != nil {
			return 0, false, err
		}
		fmt.Printf("Created bus %d\n", r.BusID)
		return r.BusID, true, nil
	}

	busID := busesResp.Buses[0]
	for _, b := range busesResp.Buses[1:] {
		if b < busID {
			busID = b
		}
	}
	fmt.Printf("Using existing bus %d\n", busID)
	return busID, false, nil
}

func demoButtons(frame uint64) uint32 {
	switch (frame / 60) % 6 {
	case 0:
		return ns2pro.ButtonA
	case 1:
		return ns2pro.ButtonB
	case 2:
		return ns2pro.ButtonX
	case 3:
		return ns2pro.ButtonY
	case 4:
		return ns2pro.ButtonL | ns2pro.ButtonR
	default:
		return ns2pro.ButtonZL | ns2pro.ButtonZR
	}
}

func stickAxis(v float64) uint16 {
	const radius = 1400.0
	raw := float64(ns2pro.StickCenter) + radius*v
	if raw < float64(ns2pro.StickMin) {
		return ns2pro.StickMin
	}
	if raw > float64(ns2pro.StickMax) {
		return ns2pro.StickMax
	}
	return uint16(raw)
}
