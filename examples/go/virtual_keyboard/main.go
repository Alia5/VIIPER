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

	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/device/keyboard"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: virtual_keyboard <api_addr>")
		fmt.Println("Example: virtual_keyboard localhost:3242")
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
	stream, addResp, err := api.AddDeviceAndConnect(ctx, busID, "keyboard", nil)
	if err != nil {
		fmt.Printf("AddDeviceAndConnect error: %v\n", err)
		if createdBus {
			_, _ = api.BusRemoveCtx(ctx, busID)
		}
		os.Exit(1)
	}
	defer stream.Close()

	fmt.Printf("Created and connected to device %s on bus %d\n", addResp.DevId, addResp.BusID)

	// Cleanup on exit
	defer func() {
		if _, err := api.DeviceRemoveCtx(ctx, stream.BusID, stream.DevID); err != nil {
			fmt.Printf("DeviceRemove error: %v\n", err)
		} else {
			fmt.Printf("Removed device %d-%s\n", addResp.BusID, addResp.DevId)
		}
		if createdBus {
			if _, err := api.BusRemoveCtx(ctx, busID); err != nil {
				fmt.Printf("BusRemove error: %v\n", err)
			} else {
				fmt.Printf("Removed bus %d\n", busID)
			}
		}
	}()

	// Start reading LED feedback (1 byte per LED state change) using StartReading
	ledCh, ledErrCh := stream.StartReading(ctx, 10, func(r *bufio.Reader) (encoding.BinaryUnmarshaler, error) {
		var b [1]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, err
		}
		st := new(keyboard.LEDState)
		if err := st.UnmarshalBinary(b[:]); err != nil {
			return nil, err
		}
		return st, nil
	})

	go func() {
		for {
			select {
			case m := <-ledCh:
				if m == nil {
					continue
				}
				lm := m.(*keyboard.LEDState)
				fmt.Printf("→ LEDs: Num=%v Caps=%v Scroll=%v Compose=%v Kana=%v\n",
					lm.NumLock, lm.CapsLock, lm.ScrollLock, lm.Compose, lm.Kana)
			case err := <-ledErrCh:
				if err != nil {
					fmt.Printf("LED read error: %v\n", err)
				}
				return
			}
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Every 5s: type 'Hello!' + Enter. Press Ctrl+C to stop.")
	for {
		select {
		case <-ticker.C:
			// Type "Hello!" character by character
			states := keyboard.TypeString("Hello!")
			for _, state := range states {
				if err := stream.WriteBinary(&state); err != nil {
					fmt.Printf("Write error: %v\n", err)
					return
				}
				time.Sleep(100 * time.Millisecond)
			}

			// Press and release Enter
			time.Sleep(100 * time.Millisecond)
			enterPress := keyboard.PressKey(keyboard.KeyEnter)
			if err := stream.WriteBinary(&enterPress); err != nil {
				fmt.Printf("Write error (enter): %v\n", err)
				return
			}

			time.Sleep(100 * time.Millisecond)
			enterRelease := keyboard.Release()
			if err := stream.WriteBinary(&enterRelease); err != nil {
				fmt.Printf("Write error (release): %v\n", err)
				return
			}

			fmt.Println("→ Typed: Hello!")
		case <-sigCh:
			fmt.Println("Signal received, stopping…")
			return
		}
	}
}
