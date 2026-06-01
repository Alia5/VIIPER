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

	"github.com/Alia5/VIIPER/device/dualsense"
	"github.com/Alia5/VIIPER/viiperclient"
)

// Usage:
//
//	virtual_ds4_cli <api_addr> <edge?>
//
// Example:
//
//	virtual_ds4_cli localhost:3242 edge
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
		fmt.Println("Usage: virtual_dsedge_cli <api_addr> <edge?>")
		fmt.Println("Example: virtual_dsedge_cli localhost:3242 edge")
		os.Exit(1)
	}

	addr := os.Args[1]
	edge := len(os.Args) > 2 && strings.ToLower(os.Args[2]) == "edge"

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

	deviceType := "dualsense"
	if edge {
		deviceType = "dualsenseedge"
	}

	stream, addResp, err := api.AddDeviceAndConnect(ctx, busID, deviceType, nil)
	if err != nil {
		fmt.Printf("AddDeviceAndConnect error: %v\n", err)
		if createdBus {
			_, _ = api.BusRemoveCtx(ctx, busID)
		}
		os.Exit(1)
	}
	defer stream.Close() //nolint:errcheck

	fmt.Printf("Connected to %s device %s on bus %d\n", deviceType, addResp.DevID, addResp.BusID)

	defer func() {
		if _, err := api.DeviceRemoveCtx(ctx, stream.BusID, stream.DevID); err != nil {
			fmt.Printf("DeviceRemove error: %v\n", err)
		}
		if createdBus {
			_, _ = api.BusRemoveCtx(ctx, busID)
		}
	}()

	feedbackCh, errCh := stream.StartReading(ctx, 10, func(r *bufio.Reader) (encoding.BinaryUnmarshaler, error) {
		var b [7]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, err
		}
		msg := new(dualsense.OutputState)
		if err := msg.UnmarshalBinary(b[:]); err != nil {
			return nil, err
		}
		return msg, nil
	})

	go func() {
		for {
			select {
			case feedback := <-feedbackCh:
				f := feedback.(*dualsense.OutputState)
				fmt.Printf("[Output] Rumble: S=%d L=%d, LED: R=%d G=%d B=%d, Player LEDs: %d\n",
					f.RumbleSmall, f.RumbleLarge, f.LedRed, f.LedGreen, f.LedBlue, f.PlayerLeds)
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
		state  dualsense.InputState
		timers map[string]*time.Timer
	}

	box := &stateBox{
		state: dualsense.InputState{
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

	fmt.Printf("%s CLI ready. Type 'help' for commands. Ctrl+C to exit.\n", deviceType)

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
			box.state = dualsense.InputState{}
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
	fmt.Println("                 L4, R4, LFn, RFn, MicMute")
	fmt.Println("  Touchpad:")
	fmt.Printf("    Touch1X, Touch1Y, Touch1Active       (u16, u16, bool; X=%d..%d, Y=%d..%d)\n",
		dualsense.TouchpadMinX, dualsense.TouchpadMaxX,
		dualsense.TouchpadMinY, dualsense.TouchpadMaxY)
	fmt.Printf("    Touch2X, Touch2Y, Touch2Active       (u16, u16, bool; X=%d..%d, Y=%d..%d)\n",
		dualsense.TouchpadMinX, dualsense.TouchpadMaxX,
		dualsense.TouchpadMinY, dualsense.TouchpadMaxY)
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
	before, after, ok0 := strings.Cut(kv, "=")
	if !ok0 {
		return "", "", 0, false, nil
	}
	key = strings.TrimSpace(before)
	val = strings.TrimSpace(after)
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

func applyKeyValue(st *dualsense.InputState, key string, val string) error {
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

	setButton := func(mask uint32, on bool) {
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
		if x < dualsense.TouchpadMinX {
			return dualsense.TouchpadMinX
		}
		if x > dualsense.TouchpadMaxX {
			return dualsense.TouchpadMaxX
		}
		return x
	}
	clampTouchY := func(y uint16) uint16 {
		if y < dualsense.TouchpadMinY {
			return dualsense.TouchpadMinY
		}
		if y > dualsense.TouchpadMaxY {
			return dualsense.TouchpadMaxY
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
		setButton((dualsense.ButtonSquare), on)
	case "cross", "x":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton((dualsense.ButtonCross), on)
	case "circle":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton((dualsense.ButtonCircle), on)
	case "triangle":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton((dualsense.ButtonTriangle), on)

	case "l1":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonL1, on)
	case "r1":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonR1, on)
	case "share":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonCreate, on)
	case "options":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonOptions, on)
	case "l3":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonL3, on)
	case "r3":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonR3, on)
	case "ps", "playstation", "guide":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonPS, on)
	case "touchpad", "touchpadbutton":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonTouchpad, on)

	case "dpadup":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setDPad(dualsense.DPadUp, on)
	case "dpaddown":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setDPad(dualsense.DPadDown, on)
	case "dpadleft":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setDPad(dualsense.DPadLeft, on)
	case "dpadright":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setDPad(dualsense.DPadRight, on)

	case "r4":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonEdgeR4, on)
	case "rfn":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonEdgeRFn, on)
	case "l4":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonEdgeL4, on)
	case "lfn":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonEdgeLFn, on)
	case "micmute":
		on, err := parseBool()
		if err != nil {
			return err
		}
		setButton(dualsense.ButtonMicMute, on)

	default:
		return fmt.Errorf("unknown key %q", key)
	}
	return nil
}

func revertKey(st *dualsense.InputState, key string, before dualsense.InputState, after dualsense.InputState) {
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
	case "square", "cross", "x", "circle", "triangle", "l1", "r1", "share", "options", "l3", "r3", "ps", "playstation", "guide", "touchpad", "touchpadbutton", "r4", "rfn", "l4", "lfn", "micmute":
		st.Buttons = before.Buttons
	case "dpadup", "dpaddown", "dpadleft", "dpadright":
		st.DPad = before.DPad
	default:
		_ = after
	}
}
