package e2e_bench_test

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/apitypes"
	"github.com/Alia5/VIIPER/device/xbox360"
	"github.com/Alia5/VIIPER/internal/cmd"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/usb"

	_ "github.com/Alia5/VIIPER/internal/registry" // Register all device handlers

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"
)

type TimeWhat int

const (
	TimeWhat_ClientWritePress TimeWhat = iota
	TimeWhat_WaitInput
	TimeWhat_ClientWriteRelease
	TimeWhat_WaitRelease
)

func Benchmark_Xbox360_Delay(b *testing.B) {

	type bench struct {
		name          string
		timeOn        func(tw TimeWhat, b *testing.B)
		useEncryption bool
	}
	benches := []bench{
		{
			name: "1 Go-Client-Write (PLAIN)",
			timeOn: func(tw TimeWhat, b *testing.B) {
				switch tw {
				case TimeWhat_ClientWritePress:
					b.StartTimer()
				case TimeWhat_WaitInput:
				case TimeWhat_ClientWriteRelease:
				case TimeWhat_WaitRelease:
				}
			},
		},
		{
			name: "2 InputDelay-Without-Client (PLAIN)",
			timeOn: func(tw TimeWhat, b *testing.B) {
				switch tw {
				case TimeWhat_ClientWritePress:
				case TimeWhat_WaitInput:
					b.StartTimer()
				case TimeWhat_ClientWriteRelease:
				case TimeWhat_WaitRelease:
				}
			},
		},
		{
			name: "3 E2E-InputDelay (PLAIN)",
			timeOn: func(tw TimeWhat, b *testing.B) {
				switch tw {
				case TimeWhat_ClientWritePress:
					b.StartTimer()
				case TimeWhat_WaitInput:
					b.StartTimer()
				case TimeWhat_ClientWriteRelease:
				case TimeWhat_WaitRelease:
				}
			},
		},
		{
			name: "4 E2E-PressAndRelease (PLAIN)",
			timeOn: func(tw TimeWhat, b *testing.B) {
				switch tw {
				case TimeWhat_ClientWritePress:
					b.StartTimer()
				case TimeWhat_WaitInput:
					b.StartTimer()
				case TimeWhat_ClientWriteRelease:
					b.StartTimer()
				case TimeWhat_WaitRelease:
					b.StartTimer()
				}
			},
		},
		{
			name: "1 Go-Client-Write (ENC)",
			timeOn: func(tw TimeWhat, b *testing.B) {
				switch tw {
				case TimeWhat_ClientWritePress:
					b.StartTimer()
				case TimeWhat_WaitInput:
				case TimeWhat_ClientWriteRelease:
				case TimeWhat_WaitRelease:
				}
			},
			useEncryption: true,
		},
		{
			name: "2 InputDelay-Without-Client (ENC)",
			timeOn: func(tw TimeWhat, b *testing.B) {
				switch tw {
				case TimeWhat_ClientWritePress:
				case TimeWhat_WaitInput:
					b.StartTimer()
				case TimeWhat_ClientWriteRelease:
				case TimeWhat_WaitRelease:
				}
			},
			useEncryption: true,
		},
		{
			name: "3 E2E-InputDelay (ENC)",
			timeOn: func(tw TimeWhat, b *testing.B) {
				switch tw {
				case TimeWhat_ClientWritePress:
					b.StartTimer()
				case TimeWhat_WaitInput:
					b.StartTimer()
				case TimeWhat_ClientWriteRelease:
				case TimeWhat_WaitRelease:
				}
			},
			useEncryption: true,
		},
		{
			name: "4 E2E-PressAndRelease (ENC)",
			timeOn: func(tw TimeWhat, b *testing.B) {
				switch tw {
				case TimeWhat_ClientWritePress:
					b.StartTimer()
				case TimeWhat_WaitInput:
					b.StartTimer()
				case TimeWhat_ClientWriteRelease:
					b.StartTimer()
				case TimeWhat_WaitRelease:
					b.StartTimer()
				}
			},
			useEncryption: true,
		},
	}

	b.SetParallelism(1)

	defer binsdl.Load().Unload()
	defer sdl.Quit()
	sdl.Init(sdl.INIT_GAMEPAD)

	sdl.UpdateGamepads()
	existingGamepads, _ := sdl.GetGamepads()
	existingGamepadSet := make(map[sdl.JoystickID]bool)
	for _, id := range existingGamepads {
		existingGamepadSet[id] = true
	}

	s := cmd.Server{
		UsbServerConfig: usb.ServerConfig{
			Addr:                    ":3244",
			BusCleanupTimeout:       1 * time.Second,
			WriteBatchFlushInterval: 0,
		},
		ApiServerConfig: api.ServerConfig{
			Addr:                        ":3245",
			AutoAttachLocalClient:       true,
			DeviceHandlerConnectTimeout: time.Second * 5,
			Password:                    "testpassword1234",
		},
		ConnectionTimeout: 5 * time.Second,
	}
	logger := slog.Default()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := s.StartServer(ctx, logger, nil); err != nil {
			panic(err)
		}
	}()

	var c *apiclient.Client

	c = apiclient.New("localhost:3245")
	var busResp *apitypes.BusCreateResponse
	var err error
	for range 10 {
		busResp, err = c.BusCreate(1)
		if err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}
	if busResp == nil {
		b.Fatalf("BusCreate failed: %v", err)
	}
	busID := busResp.BusID
	defer c.BusRemove(busID)

	devInfo, err := c.DeviceAdd(busID, "xbox360", nil)
	if err != nil {
		b.Fatalf("DeviceAdd failed: %v", err)
	}

	devStream, err := c.OpenStream(ctx, busID, devInfo.DevId)
	if err != nil {
		b.Fatalf("OpenStream failed: %v", err)
	}
	defer devStream.Close()

	var gamepad *sdl.Gamepad
	for range 10 {
		sdl.UpdateGamepads()
		gIDs, _ := sdl.GetGamepads()
		for _, id := range gIDs {
			if !existingGamepadSet[id] {
				gamepad, err = id.OpenGamepad()
				if err != nil {
					b.Fatalf("OpenGamepad failed: %v", err)
				}
				defer gamepad.Close()
				break
			}
		}
		if gamepad != nil {
			break
		}
		time.Sleep(time.Second * 1)
	}
	if gamepad == nil {
		b.Fatalf("No new gamepad found for testing (expected VIIPER virtual device)")
	}
	padChann := make(chan bool)
	prevPadPressed := false
	go func() {
		defer close(padChann)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			sdl.UpdateGamepads()
			pressed := gamepad.Button(sdl.GAMEPAD_BUTTON_SOUTH)
			if pressed != prevPadPressed {
				padChann <- pressed
				prevPadPressed = pressed
			}
		}
	}()

	for _, bench := range benches {
		if bench.useEncryption {
			c = apiclient.NewWithPassword("localhost:3245", "testpassword1234")
		}
		b.Run(bench.name, func(b *testing.B) {
			for b.Loop() {
				b.StopTimer()
				bench.timeOn(TimeWhat_ClientWritePress, b)
				err = devStream.WriteBinary(&xbox360.InputState{
					Buttons: xbox360.ButtonA,
				})
				b.StopTimer()
				if err != nil {
					b.Fatalf("WriteBinary failed: %v", err)
				}
				timeout := time.After(1 * time.Second)

				bench.timeOn(TimeWhat_WaitInput, b)
				waitForInput(ctx, timeout, padChann, true)

				b.StopTimer()
				bench.timeOn(TimeWhat_ClientWriteRelease, b)
				err = devStream.WriteBinary(&xbox360.InputState{})
				b.StopTimer()
				if err != nil {
					b.Fatalf("WriteBinary failed: %v", err)
				}
				timeout = time.After(10000 * time.Second)
				bench.timeOn(TimeWhat_WaitRelease, b)
				waitForInput(ctx, timeout, padChann, false)

				b.StartTimer()
			}
		})
	}
}

func waitForInput(ctx context.Context, timeout <-chan time.Time, padChann <-chan bool, wantPressed bool) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return context.DeadlineExceeded
		case pressed, ok := <-padChann:
			if !ok {
				return context.Canceled
			}
			if pressed == wantPressed {
				return nil
			}
		}
	}
}
