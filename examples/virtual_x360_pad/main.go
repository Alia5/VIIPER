package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
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
	api := apiclient.New(addr)

	busesResp, err := api.BusList()
	if err != nil {
		fmt.Printf("BusList error: %v\n", err)
		os.Exit(1)
	}
	var busID uint32
	createdBus := false
	if len(busesResp.Buses) == 0 {
		var createErr error
		for try := uint32(1); try <= 100; try++ {
			if r, err := api.BusCreate(try); err == nil {
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

	addResp, err := api.DeviceAdd(busID, "xbox360")
	if err != nil {
		fmt.Printf("DeviceAdd error: %v\n", err)
		if createdBus {
			_, _ = api.BusRemove(busID)
		}
		os.Exit(1)
	}
	deviceBusId := addResp.ID
	devId := deviceBusId
	if i := strings.Index(deviceBusId, "-"); i >= 0 && i+1 < len(deviceBusId) {
		devId = deviceBusId[i+1:]
	}
	createdDevice := true
	fmt.Printf("Created device %s on bus %d\n", devId, busID)

	defer func() {
		if createdDevice {
			if _, err := api.DeviceRemove(busID, devId); err != nil {
				fmt.Printf("DeviceRemove error: %v\n", err)
			} else {
				fmt.Printf("Removed device %s\n", deviceBusId)
			}
		}
		if createdBus {
			if _, err := api.BusRemove(busID); err != nil {
				fmt.Printf("BusRemove error: %v\n", err)
			} else {
				fmt.Printf("Removed bus %d\n", busID)
			}
		}
	}()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Stream dial error: %v\n", err)
		return
	}
	defer conn.Close()
	if _, err := fmt.Fprintf(conn, "bus/%d/%s\n", busID, devId); err != nil {
		fmt.Printf("Stream activate error: %v\n", err)
		return
	}
	fmt.Printf("Stream activated for %s\n", devId)

	stop := make(chan struct{})
	go func() {
		buf := make([]byte, 2)
		for {
			if _, err := io.ReadFull(conn, buf); err != nil {
				return
			}
			fmt.Printf("← Rumble: Left=%d, Right=%d\n", buf[0], buf[1])
		}
	}()

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
				buttons = 0x0001 // A
			case 1:
				buttons = 0x0002 // B
			case 2:
				buttons = 0x0004 // X
			default:
				buttons = 0x0008 // Y
			}
			state := xbox360.InputState{
				Buttons: buttons,
				LT:      uint8((frame * 2) % 256),
				RT:      uint8((frame * 3) % 256),
				LX:      int16(20000.0 * 0.7071),
				LY:      int16(20000.0 * 0.7071),
				RX:      0,
				RY:      0,
			}
			pkt, _ := state.MarshalBinary()
			if _, err := conn.Write(pkt); err != nil {
				fmt.Printf("Write error: %v\n", err)
				close(stop)
				return
			}
			if frame%60 == 0 {
				fmt.Printf("→ Sent input (frame %d): buttons=0x%04x, LT=%d, RT=%d\n", frame, state.Buttons, state.LT, state.RT)
			}
		case <-sigCh:
			fmt.Println("Signal received, stopping…")
			return
		case <-stop:
			return
		}
	}
}
