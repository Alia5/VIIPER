package main

import (
	"bufio"
	"context"
	"encoding"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Alia5/VIIPER/device/dualshock4"
	"github.com/Alia5/VIIPER/viiperclient"
)

// Usage:
//
//	virtual_ds4_cli <api_addr>
//
// Example:
//
//	virtual_ds4_cli localhost:3242
//
// Commands (case-insensitive):
//
//	LX=-100
//	R2=82
//	GyroX=12
//	Circle=true
//	Circle=false
//	Triangle=true 12ms        # pulse for 12ms
//	DPadUp=true
//	DPadLeft=false
//	print
//	reset
//	help
//	quit
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: virtual_ds4_cli <api_addr>")
		fmt.Println("Example: virtual_ds4_cli localhost:3242")
		os.Exit(1)
	}

	addr := os.Args[1]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api := viiperclient.New(addr)

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

	stream, addResp, err := api.AddDeviceAndConnect(ctx, busID, "dualshock4", nil)
	if err != nil {
		fmt.Printf("AddDeviceAndConnect error: %v\n", err)
		if createdBus {
			_, _ = api.BusRemoveCtx(ctx, busID)
		}
		os.Exit(1)
	}
	defer stream.Close() //nolint:errcheck

	fmt.Printf("Connected to DualShock 4 device %s on bus %d\n", addResp.DevID, addResp.BusID)

	defer func() {
		if _, err := api.DeviceRemoveCtx(ctx, stream.BusID, stream.DevID); err != nil {
			fmt.Printf("DeviceRemove error: %v\n", err)
		}
		if createdBus {
			_, _ = api.BusRemoveCtx(ctx, busID)
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
				if err != nil {
					fmt.Printf("[Output read error] %v\n", err)
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	type stateBox struct {
		mu     sync.Mutex
		state  dualshock4.InputState
		timers map[string]*time.Timer
	}

	box := &stateBox{
		state: dualshock4.InputState{
			LX: 0, LY: 0, RX: 0, RY: 0,
			Buttons: 0,
			DPad:    0,
			L2:      0,
			R2:      0,
			GyroX:   0,
			GyroY:   0,
			GyroZ:   0,
			AccelX:  0,
			AccelY:  0,
			AccelZ:  0,
		},
		timers: map[string]*time.Timer{},
	}

	sendTicker := time.NewTicker(5 * time.Millisecond)
	defer sendTicker.Stop()

	go func() {
		for {
			select {
			case <-sendTicker.C:
				box.mu.Lock()
				st := box.state
				box.mu.Unlock()
				if err := stream.WriteBinary(&st); err != nil {
					fmt.Printf("Send error: %v\n", err)
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	fmt.Println("DS4 CLI ready. Type 'help' for commands. Ctrl+C to exit.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		select {
		case <-sigCh:
			fmt.Println("\nShutting down...")
			cancel()
			return
		default:
		}

		if !scanner.Scan() {
			cancel()
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		lower := strings.ToLower(line)
		switch lower {
		case "quit", "exit":
			cancel()
			return
		case "help", "?":
			printHelp()
			continue
		case "print":
			box.mu.Lock()
			fmt.Printf("%+v\n", box.state)
			box.mu.Unlock()
			continue
		case "reset":
			box.mu.Lock()
			box.state = dualshock4.InputState{}
			box.mu.Unlock()
			fmt.Println("state reset")
			continue
		}

		key, val, dur, ok, err := parseAssignment(line)
		if err != nil {
			fmt.Printf("parse error: %v\n", err)
			continue
		}
		if !ok {
			fmt.Println("unrecognized command; try 'help'")
			continue
		}

		box.mu.Lock()
		before := box.state
		applyErr := applyKeyValue(&box.state, key, val)
		if applyErr != nil {
			box.mu.Unlock()
			fmt.Printf("apply error: %v\n", applyErr)
			continue
		}

		if dur > 0 {
			id := strings.ToLower(key)
			if t := box.timers[id]; t != nil {
				t.Stop()
			}
			after := box.state
			box.timers[id] = time.AfterFunc(dur, func() {
				box.mu.Lock()
				revertKey(&box.state, id, before, after)
				box.mu.Unlock()
			})
		}
		box.mu.Unlock()
	}
}

func printHelp() {
	fmt.Println("Assignments: Key=Value [duration]")
	fmt.Println("  Example: LX=-100")
	fmt.Println("  Example: Triangle=true 12ms")
	fmt.Println("Keys (case-insensitive):")
	fmt.Println("  Sticks: LX, LY, RX, RY                (int8, -128..127)")
	fmt.Println("  Triggers: L2, R2                      (uint8, 0..255)")
	fmt.Println("  Sensors: GyroX, GyroY, GyroZ          (int16)")
	fmt.Println("           AccelX, AccelY, AccelZ       (int16)")
	fmt.Println("  Buttons (bool): Square, Cross, Circle, Triangle")
	fmt.Println("                 L1, R1, Share, Options, L3, R3")
	fmt.Println("                 PS (aka PlayStation/Guide), Touchpad")
	fmt.Println("  Touchpad:")
	fmt.Printf("    Touch1X, Touch1Y, Touch1Active       (u16, u16, bool; X=%d..%d, Y=%d..%d)\n",
		dualshock4.TouchpadMinX, dualshock4.TouchpadMaxX,
		dualshock4.TouchpadMinY, dualshock4.TouchpadMaxY)
	fmt.Printf("    Touch2X, Touch2Y, Touch2Active       (u16, u16, bool; X=%d..%d, Y=%d..%d)\n",
		dualshock4.TouchpadMinX, dualshock4.TouchpadMaxX,
		dualshock4.TouchpadMinY, dualshock4.TouchpadMaxY)
	fmt.Println("    Touch1=123,456                       (sets Touch1X/Touch1Y + Touch1Active=true)")
	fmt.Println("    Touch2=123,456                       (sets Touch2X/Touch2Y + Touch2Active=true)")
	fmt.Println("  DPad (bool): DPadUp, DPadDown, DPadLeft, DPadRight")
	fmt.Println("Other commands: print | reset | help | quit")
	fmt.Println("NOTE: This is a temporary hacky tool; it only supports what the current wire protocol exposes.")
}

func parseAssignment(line string) (key string, val string, dur time.Duration, ok bool, err error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", "", 0, false, nil
	}
	kv := parts[0]
	eq := strings.IndexByte(kv, '=')
	if eq < 0 {
		return "", "", 0, false, nil
	}
	key = strings.TrimSpace(kv[:eq])
	val = strings.TrimSpace(kv[eq+1:])
	if key == "" {
		return "", "", 0, false, fmt.Errorf("missing key")
	}
	if len(parts) >= 2 {
		d, e := time.ParseDuration(parts[1])
		if e != nil {
			return "", "", 0, false, fmt.Errorf("bad duration %q", parts[1])
		}
		dur = d
	}
	return key, val, dur, true, nil
}

func applyKeyValue(st *dualshock4.InputState, key string, val string) error {
	k := strings.ToLower(strings.TrimSpace(key))
	v := strings.ToLower(strings.TrimSpace(val))

	parseI8 := func() (int8, error) {
		i, err := strconv.ParseInt(val, 10, 8)
		return int8(i), err
	}
	parseU8 := func() (uint8, error) {
		u, err := strconv.ParseUint(val, 10, 8)
		return uint8(u), err
	}
	parseI16 := func() (int16, error) {
		i, err := strconv.ParseInt(val, 10, 16)
		return int16(i), err
	}
	parseU16 := func() (uint16, error) {
		u, err := strconv.ParseUint(val, 10, 16)
		return uint16(u), err
	}
	parseXY := func() (uint16, uint16, error) {
		parts := strings.Split(val, ",")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("expected x,y got %q", val)
		}
		xs := strings.TrimSpace(parts[0])
		ys := strings.TrimSpace(parts[1])
		xu, err := strconv.ParseUint(xs, 10, 16)
		if err != nil {
			return 0, 0, fmt.Errorf("bad x %q: %w", xs, err)
		}
		yu, err := strconv.ParseUint(ys, 10, 16)
		if err != nil {
			return 0, 0, fmt.Errorf("bad y %q: %w", ys, err)
		}
		return uint16(xu), uint16(yu), nil
	}
	parseBool := func() (bool, error) {
		switch v {
		case "1", "true", "t", "yes", "y", "on":
			return true, nil
		case "0", "false", "f", "no", "n", "off":
			return false, nil
		default:
			return false, fmt.Errorf("expected bool, got %q", val)
		}
	}

	setButton := func(mask uint16, on bool) {
		if on {
			st.Buttons |= mask
		} else {
			st.Buttons &^= mask
		}
	}
	setDPad := func(mask uint8, on bool) {
		if on {
			st.DPad |= mask
		} else {
			st.DPad &^= mask
		}
	}
	clampTouchX := func(x uint16) uint16 {
		if x < dualshock4.TouchpadMinX {
			return dualshock4.TouchpadMinX
		}
		if x > dualshock4.TouchpadMaxX {
			return dualshock4.TouchpadMaxX
		}
		return x
	}
	clampTouchY := func(y uint16) uint16 {
		if y < dualshock4.TouchpadMinY {
			return dualshock4.TouchpadMinY
		}
		if y > dualshock4.TouchpadMaxY {
			return dualshock4.TouchpadMaxY
		}
		return y
	}

	switch k {
	case "lx":
		x, err := parseI8()
		if err != nil {
			return err
		}
		st.LX = x
	case "ly":
		x, err := parseI8()
		if err != nil {
			return err
		}
		st.LY = x
	case "rx":
		x, err := parseI8()
		if err != nil {
			return err
		}
		st.RX = x
	case "ry":
		x, err := parseI8()
		if err != nil {
			return err
		}
		st.RY = x

	case "l2":
		x, err := parseU8()
		if err != nil {
			return err
		}
		st.L2 = x
	case "r2":
		x, err := parseU8()
		if err != nil {
			return err
		}
		st.R2 = x

	case "gyrox":
		x, err := parseI16()
		if err != nil {
			return err
		}
		st.GyroX = x
	case "gyroy":
		x, err := parseI16()
		if err != nil {
			return err
		}
		st.GyroY = x
	case "gyroz":
		x, err := parseI16()
		if err != nil {
			return err
		}
		st.GyroZ = x

	case "accelx":
		x, err := parseI16()
		if err != nil {
			return err
		}
		st.AccelX = x
	case "accely":
		x, err := parseI16()
		if err != nil {
			return err
		}
		st.AccelY = x
	case "accelz":
		x, err := parseI16()
		if err != nil {
			return err
		}
		st.AccelZ = x

	case "touch1x":
		x, err := parseU16()
		if err != nil {
			return err
		}
		st.Touch1X = clampTouchX(x)
	case "touch1y":
		y, err := parseU16()
		if err != nil {
			return err
		}
		st.Touch1Y = clampTouchY(y)
	case "touch1active", "touch1down":
		on, err := parseBool()
		if err != nil {
			return err
		}
		st.Touch1Active = on
	case "touch2x":
		x, err := parseU16()
		if err != nil {
			return err
		}
		st.Touch2X = clampTouchX(x)
	case "touch2y":
		y, err := parseU16()
		if err != nil {
			return err
		}
		st.Touch2Y = clampTouchY(y)
	case "touch2active", "touch2down":
		on, err := parseBool()
		if err != nil {
			return err
		}
		st.Touch2Active = on
	case "touch1":
		if v == "false" || v == "off" || v == "0" {
			st.Touch1Active = false
			return nil
		}
		x, y, err := parseXY()
		if err != nil {
			return err
		}
		st.Touch1X, st.Touch1Y = clampTouchX(x), clampTouchY(y)
		st.Touch1Active = true
	case "touch2":
		if v == "false" || v == "off" || v == "0" {
			st.Touch2Active = false
			return nil
		}
		x, y, err := parseXY()
		if err != nil {
			return err
		}
		st.Touch2X, st.Touch2Y = clampTouchX(x), clampTouchY(y)
		st.Touch2Active = true

	case "square":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton((dualshock4.ButtonSquare), on)
	case "cross", "x":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton((dualshock4.ButtonCross), on)
	case "circle":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton((dualshock4.ButtonCircle), on)
	case "triangle":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton((dualshock4.ButtonTriangle), on)

	case "l1":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualshock4.ButtonL1, on)
	case "r1":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualshock4.ButtonR1, on)
	case "share":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualshock4.ButtonShare, on)
	case "options":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualshock4.ButtonOptions, on)
	case "l3":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualshock4.ButtonL3, on)
	case "r3":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualshock4.ButtonR3, on)
	case "ps", "playstation", "guide":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualshock4.ButtonPS, on)
	case "touchpad", "touchpadbutton":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualshock4.ButtonTouchpadClick, on)

	case "dpadup":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setDPad(dualshock4.DPadUp, on)
	case "dpaddown":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setDPad(dualshock4.DPadDown, on)
	case "dpadleft":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setDPad(dualshock4.DPadLeft, on)
	case "dpadright":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setDPad(dualshock4.DPadRight, on)

	default:
		return fmt.Errorf("unknown key %q", key)
	}
	return nil
}

func revertKey(st *dualshock4.InputState, key string, before dualshock4.InputState, after dualshock4.InputState) {
	switch key {
	case "lx":
		st.LX = before.LX
	case "ly":
		st.LY = before.LY
	case "rx":
		st.RX = before.RX
	case "ry":
		st.RY = before.RY
	case "l2":
		st.L2 = before.L2
	case "r2":
		st.R2 = before.R2
	case "gyrox":
		st.GyroX = before.GyroX
	case "gyroy":
		st.GyroY = before.GyroY
	case "gyroz":
		st.GyroZ = before.GyroZ
	case "accelx":
		st.AccelX = before.AccelX
	case "accely":
		st.AccelY = before.AccelY
	case "accelz":
		st.AccelZ = before.AccelZ
	case "touch1x":
		st.Touch1X = before.Touch1X
	case "touch1y":
		st.Touch1Y = before.Touch1Y
	case "touch1active", "touch1down", "touch1":
		st.Touch1Active = before.Touch1Active
		st.Touch1X = before.Touch1X
		st.Touch1Y = before.Touch1Y
	case "touch2x":
		st.Touch2X = before.Touch2X
	case "touch2y":
		st.Touch2Y = before.Touch2Y
	case "touch2active", "touch2down", "touch2":
		st.Touch2Active = before.Touch2Active
		st.Touch2X = before.Touch2X
		st.Touch2Y = before.Touch2Y
	case "square", "cross", "x", "circle", "triangle", "l1", "r1", "share", "options", "l3", "r3", "ps", "playstation", "guide", "touchpad", "touchpadbutton":
		st.Buttons = before.Buttons
	case "dpadup", "dpaddown", "dpadleft", "dpadright":
		st.DPad = before.DPad
	default:
		_ = after
	}
}
