package keyboard_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/device/keyboard"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	viiperTesting "github.com/Alia5/VIIPER/testing"
	"github.com/Alia5/VIIPER/usbip"
	"github.com/Alia5/VIIPER/virtualbus"
	"github.com/stretchr/testify/assert"

	_ "github.com/Alia5/VIIPER/internal/registry" // Register devices
)

func TestInputReports(t *testing.T) {
	type testCase struct {
		name           string
		inputState     keyboard.InputState
		expectedReport []byte
	}

	cases := []testCase{
		{
			name: "No keys, no modifiers",
			inputState: keyboard.InputState{
				Modifiers: 0,
				KeyBitmap: [32]uint8{},
			},
			expectedReport: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			name:           "C",
			inputState:     keyboard.PressKey(keyboard.KeyC),
			expectedReport: []byte{0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:           "CTRL+C",
			inputState:     keyboard.PressKeyWithMod(keyboard.ModLeftCtrl, keyboard.KeyC),
			expectedReport: []byte{0x01, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:           "SHIFT+C",
			inputState:     keyboard.PressKeyWithMod(keyboard.ModLeftShift, keyboard.KeyC),
			expectedReport: []byte{0x02, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:           "ALT+C",
			inputState:     keyboard.PressKeyWithMod(keyboard.ModLeftAlt, keyboard.KeyC),
			expectedReport: []byte{0x04, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:           "WASD",
			inputState:     keyboard.PressKey(keyboard.KeyW, keyboard.KeyA, keyboard.KeyS, keyboard.KeyD),
			expectedReport: []byte{0x00, 0x00, 0x90, 0x00, 0x40, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	s := viiperTesting.NewTestServer(t)
	defer s.UsbServer.Close()
	defer s.ApiServer.Close()

	r := s.ApiServer.Router()
	r.Register("bus/{id}/add", handler.BusDeviceAdd(s.UsbServer, s.ApiServer))
	r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(s.UsbServer))

	if err := s.ApiServer.Start(); err != nil {
		t.Fatalf("Failed to start API server: %v", err)
	}

	b, err := virtualbus.NewWithBusId(1)
	if err != nil {
		t.Fatalf("Failed to create virtual bus: %v", err)
	}
	defer b.Close()
	_ = s.UsbServer.AddBus(b)

	client := apiclient.New(s.ApiServer.Addr())
	stream, _, err := client.AddDeviceAndConnect(context.Background(), b.BusID(), "keyboard", nil)
	if !assert.NoError(t, err) {
		return
	}
	defer stream.Close()

	usbipClient := viiperTesting.NewUsbIpClient(t, s.UsbServer.Addr())
	devs, err := usbipClient.ListDevices()
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Len(t, devs, 1) {
		return
	}
	imp, err := usbipClient.AttachDevice(devs[0].BusID)
	if !assert.NoError(t, err) {
		return
	}
	if imp != nil && imp.Conn != nil {
		defer imp.Conn.Close()
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedReport, tc.inputState.BuildReport())
			if !assert.NoError(t, stream.WriteBinary(&tc.inputState)) {
				return
			}
			got, err := usbipClient.PollInputReport(imp.Conn, tc.expectedReport, 750*time.Millisecond)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tc.expectedReport, got)
		})
	}
}

func TestLEDs(t *testing.T) {
	type testCase struct {
		name      string
		ledMask   byte
		outPacket []byte
	}

	cases := []testCase{
		{
			name:      "off",
			ledMask:   0x00,
			outPacket: []byte{0x00},
		},
		{
			name:      "numlock",
			ledMask:   keyboard.LEDNumLock,
			outPacket: []byte{keyboard.LEDNumLock},
		},
		{
			name:      "capslock",
			ledMask:   keyboard.LEDCapsLock,
			outPacket: []byte{keyboard.LEDCapsLock},
		},
		{
			name:      "scrolllock",
			ledMask:   keyboard.LEDScrollLock,
			outPacket: []byte{keyboard.LEDScrollLock},
		},
		{
			name:    "all",
			ledMask: keyboard.LEDNumLock | keyboard.LEDCapsLock | keyboard.LEDScrollLock | keyboard.LEDCompose | keyboard.LEDKana,
			outPacket: []byte{
				keyboard.LEDNumLock | keyboard.LEDCapsLock | keyboard.LEDScrollLock | keyboard.LEDCompose | keyboard.LEDKana,
			},
		},
	}

	s := viiperTesting.NewTestServer(t)
	defer s.UsbServer.Close()
	defer s.ApiServer.Close()

	r := s.ApiServer.Router()
	r.Register("bus/{id}/add", handler.BusDeviceAdd(s.UsbServer, s.ApiServer))
	r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(s.UsbServer))

	if err := s.ApiServer.Start(); err != nil {
		t.Fatalf("Failed to start API server: %v", err)
	}

	b, err := virtualbus.NewWithBusId(1)
	if err != nil {
		t.Fatalf("Failed to create virtual bus: %v", err)
	}
	defer b.Close()
	_ = s.UsbServer.AddBus(b)

	client := apiclient.New(s.ApiServer.Addr())
	stream, _, err := client.AddDeviceAndConnect(context.Background(), b.BusID(), "keyboard", nil)
	if !assert.NoError(t, err) {
		return
	}
	defer stream.Close()

	usbipClient := viiperTesting.NewUsbIpClient(t, s.UsbServer.Addr())
	devs, err := usbipClient.ListDevices()
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Len(t, devs, 1) {
		return
	}
	imp, err := usbipClient.AttachDevice(devs[0].BusID)
	if !assert.NoError(t, err) {
		return
	}
	if imp != nil && imp.Conn != nil {
		defer imp.Conn.Close()
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !assert.NoError(t, usbipClient.Submit(imp.Conn, usbip.DirOut, 1, tc.outPacket, nil)) {
				return
			}
			var buf [1]byte
			_ = stream.SetReadDeadline(time.Now().Add(750 * time.Millisecond))
			_, err := io.ReadFull(stream, buf[:])
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tc.ledMask, buf[0])
		})
	}
}
