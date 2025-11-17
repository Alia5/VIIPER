package main

import (
	"bufio"
	"context"
	"encoding"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"viiper/pkg/apiclient"
	"viiper/pkg/device/xbox360"
)

// Minimal example: ensure a bus, create an xbox360 device, stream inputs, read rumble, clean up on exit.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: xbox360_client <api_addr>")
		fmt.Println("Example: xbox360_client localhost:3242")
		os.Exit(1)
	}

	addr := os.Args[1]
	ctx := context.Background()
	api := apiclient.New(addr)

	// Find or create a bus
	busesResp, err := api.BusListCtx(ctx)
	if err != nil {
		fmt.Printf("BusList error: %v\n", err)
		os.Exit(1)
	}
	var busID uint32
	createdBus := false
	if len(busesResp.Buses) == 0 {
		var createErr error
		for try := uint32(1); try <= 100; try++ {
			if r, err := api.BusCreateCtx(ctx, try); err == nil {
				busID = r.BusID
				createdBus = true
				break
			}
			createErr = err
		}
		if busID == 0 {
			fmt.Printf("BusCreate failed: %v\n", createErr)
			os.Exit(1)
		}
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
	stream, addResp, err := api.AddDeviceAndConnect(ctx, busID, "xbox360")
	if err != nil {
		fmt.Printf("AddDeviceAndConnect error: %v\n", err)
		if createdBus {
			_, _ = api.BusRemoveCtx(ctx, busID)
		}
		os.Exit(1)
	}
	defer stream.Close()

	deviceBusId := addResp.ID
	fmt.Printf("Created and connected to device %s on bus %d\n", deviceBusId, busID)

	// Cleanup on exit
	defer func() {
		if _, err := api.DeviceRemoveCtx(ctx, stream.BusID, stream.DevID); err != nil {
			fmt.Printf("DeviceRemove error: %v\n", err)
		} else {
			fmt.Printf("Removed device %s\n", deviceBusId)
		}
		if createdBus {
			if _, err := api.BusRemoveCtx(ctx, busID); err != nil {
				fmt.Printf("BusRemove error: %v\n", err)
			} else {
				fmt.Printf("Removed bus %d\n", busID)
			}
		}
	}()

	// Start event-driven rumble reading
	rumbleCh, errCh := stream.StartReading(ctx, 10, func(r *bufio.Reader) (encoding.BinaryUnmarshaler, error) {
		var b [2]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, err
		}
		msg := new(xbox360.XRumbleState)
		if err := msg.UnmarshalBinary(b[:]); err != nil {
			return nil, err
		}
		return msg, nil
	})

	go func() {
		for {
			select {
			case msg := <-rumbleCh:
				if msg != nil {
					rumble := msg.(*xbox360.XRumbleState)
					fmt.Printf("← Rumble: Left=%d, Right=%d\n", rumble.LeftMotor, rumble.RightMotor)
				}
			case err := <-errCh:
				if err != nil {
					fmt.Printf("Stream read error: %v\n", err)
				}
				return
			}
		}
	}()

	// Send controller inputs
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	var frame uint64
	for {
		select {
		case <-ticker.C:
			frame++
			var buttons uint32
			switch (frame / 60) % 4 {
			case 0:
				buttons = xbox360.ButtonA
			case 1:
				buttons = xbox360.ButtonB
			case 2:
				buttons = xbox360.ButtonX
			default:
				buttons = xbox360.ButtonY
			}
			state := &xbox360.InputState{
				Buttons: buttons,
				LT:      uint8((frame * 2) % 256),
				RT:      uint8((frame * 3) % 256),
				LX:      int16(20000.0 * 0.7071),
				LY:      int16(20000.0 * 0.7071),
				RX:      0,
				RY:      0,
			}
			if err := stream.WriteBinary(state); err != nil {
				fmt.Printf("Write error: %v\n", err)
				return
			}
			if frame%60 == 0 {
				fmt.Printf("→ Sent input (frame %d): buttons=0x%04x, LT=%d, RT=%d\n", frame, state.Buttons, state.LT, state.RT)
			}
		case <-sigCh:
			fmt.Println("Signal received, stopping…")
			return
		}
	}
}
