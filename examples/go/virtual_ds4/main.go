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

	"github.com/Alia5/VIIPER/device/dualshock4"
	"github.com/Alia5/VIIPER/viiperclient"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: virtual_ds4 <api_addr>")
		fmt.Println("Example: virtual_ds4 localhost:3242")
		os.Exit(1)
	}

	addr := os.Args[1]
	ctx := context.Background()
	api := viiperclient.New(addr)

	// Find or create a bus
	busesResp, err := api.BusListCtx(ctx)
	if err != nil {
		fmt.Printf("BusList error: %v\n", err)
		os.Exit(1)
	}
	var busID uint32
	createdBus := false
	if len(busesResp.Buses) == 0 {
		r, err := api.BusCreateCtx(ctx, 0)
		if err != nil {
			fmt.Printf("BusCreate failed: %v\n", err)
			os.Exit(1)
		}
		busID = r.BusID
		createdBus = true
		fmt.Printf("Created bus %d\n", busID)
	} else {
		busID = busesResp.Buses[0]
		for _, b := range busesResp.Buses[1:] {
			if b < busID {
				busID = b
			}
		}
		fmt.Printf("Using existing bus %d\n", busID)
	}

	// Add device and connect to stream in one call
	stream, addResp, err := api.AddDeviceAndConnect(ctx, busID, "dualshock4", nil)
	if err != nil {
		fmt.Printf("AddDeviceAndConnect error: %v\n", err)
		if createdBus {
			_, _ = api.BusRemoveCtx(ctx, busID)
		}
		os.Exit(1)
	}
	defer stream.Close() //nolint:errcheck

	fmt.Printf("Created and connected to DualShock 4 device %s on bus %d\n", addResp.DevID, addResp.BusID)

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

	feedbackCh, errCh := stream.StartReading(ctx, 10, func(r *bufio.Reader) (encoding.BinaryUnmarshaler, error) {
		var b [8]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, err
		}
		msg := new(dualshock4.OutputState)
		if err := msg.UnmarshalBinary(b[:]); err != nil {
			return nil, err
		}
		return msg, nil
	})

	go func() {
		for {
			select {
			case feedback := <-feedbackCh:
				f := feedback.(*dualshock4.OutputState)
				fmt.Printf("[Output] Flags=0x%02X, Rumble: S=%d L=%d, LED: R=%d G=%d B=%d, Flash: On=%d Off=%d\n",
					f.UpdateFlags, f.RumbleSmall, f.RumbleLarge, f.LedRed, f.LedGreen, f.LedBlue, f.FlashOn, f.FlashOff)
			case err := <-errCh:
				fmt.Printf("[Output read error] %v\n", err)
				return
			}
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	fmt.Println("DualShock 4 device active. Press Ctrl+C to exit.")
	fmt.Println("Demo: Slowly moving left stick in a circle...")

	angle := 0.0
	for {
		select {
		case <-ticker.C:
			angle += 0.05
			if angle > 6.28 {
				angle = 0
			}

			lx := int8(0x40 * math.Cos(angle))
			ly := int8(0x40 * math.Sin(angle))

			state := dualshock4.InputState{
				LX:           lx,
				LY:           ly,
				RX:           0,
				RY:           0,
				Buttons:      0,
				DPad:         0,
				L2:           0,
				R2:           0,
				Touch1X:      0,
				Touch1Y:      0,
				Touch1Active: false,
				Touch2X:      0,
				Touch2Y:      0,
				Touch2Active: false,
				GyroX:        0,
				GyroY:        0,
				GyroZ:        0,
				AccelX:       0,
				AccelY:       0,
				AccelZ:       0,
			}

			if err := stream.WriteBinary(&state); err != nil {
				fmt.Printf("Send error: %v\n", err)
				return
			}

		case <-sigCh:
			fmt.Println("\nShutting down...")
			return
		}
	}
}
