package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"viiper/pkg/apiclient"
	"viiper/pkg/device/mouse"
)

// Minimal example: ensure a bus, create a mouse device, stream inputs, clean up on exit.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: virtual_mouse <api_addr>")
		fmt.Println("Example: virtual_mouse localhost:3242")
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
	stream, addResp, err := api.AddDeviceAndConnect(ctx, busID, "mouse")
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

	// Send a short movement once every 3 seconds for easy local testing.
	// Followed by a short click and a single scroll notch.
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Alternate direction to keep the pointer near its origin.
	dir := int8(1)
	const step = int8(50) // move diagonally by 50 px in X and Y
	fmt.Println("Every 3s: move diagonally by 50px (X and Y), then click and scroll. Press Ctrl+C to stop.")
	for {
		select {
		case <-ticker.C:
			// Move diagonally: (+step,+step) then (-step,-step) next tick
			dx := step * dir
			dy := step * dir
			dir *= -1

			// One-shot movement report (diagonal)
			move := &mouse.InputState{DX: dx, DY: dy}
			if err := stream.WriteBinary(move); err != nil {
				fmt.Printf("Write error (move): %v\n", err)
				return
			}
			fmt.Printf("→ Moved mouse dx=%d dy=%d\n", dx, dy)

			// Zero state shortly after to keep movement one-shot (harmless safety)
			time.Sleep(30 * time.Millisecond)
			zero := &mouse.InputState{}
			if err := stream.WriteBinary(zero); err != nil {
				fmt.Printf("Write error (zero after move): %v\n", err)
				return
			}

			// Simulate a short left click: press then release
			time.Sleep(50 * time.Millisecond)
			press := &mouse.InputState{Buttons: mouse.Btn_Left}
			if err := stream.WriteBinary(press); err != nil {
				fmt.Printf("Write error (press): %v\n", err)
				return
			}
			time.Sleep(60 * time.Millisecond)
			rel := &mouse.InputState{Buttons: 0x00}
			if err := stream.WriteBinary(rel); err != nil {
				fmt.Printf("Write error (release): %v\n", err)
				return
			}
			fmt.Printf("→ Clicked (left)\n")

			// Simulate a short scroll: one notch upwards
			time.Sleep(50 * time.Millisecond)
			scr := &mouse.InputState{Wheel: 1}
			if err := stream.WriteBinary(scr); err != nil {
				fmt.Printf("Write error (scroll): %v\n", err)
				return
			}
			time.Sleep(30 * time.Millisecond)
			scr0 := &mouse.InputState{}
			if err := stream.WriteBinary(scr0); err != nil {
				fmt.Printf("Write error (zero after scroll): %v\n", err)
				return
			}
			fmt.Printf("→ Scrolled (wheel=+1)\n")
		case <-sigCh:
			fmt.Println("Signal received, stopping…")
			return
		}
	}
}
