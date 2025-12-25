package mouse_test

import (
	"context"
	"testing"
	"time"

	viiperTesting "github.com/Alia5/VIIPER/_testing"
	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/device/mouse"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/virtualbus"
	"github.com/stretchr/testify/assert"

	_ "github.com/Alia5/VIIPER/internal/registry" // Register devices
)

func TestInputReports(t *testing.T) {
	type testCase struct {
		name           string
		inputState     mouse.InputState
		expectedReport []byte
	}

	cases := []testCase{
		{
			name: "No movement, no buttons",
			inputState: mouse.InputState{
				Buttons: 0,
				DX:      0,
				DY:      0,
				Pan:     0,
				Wheel:   0,
			},
			expectedReport: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "Left down",
			inputState: mouse.InputState{
				Buttons: mouse.Btn_Left,
				DX:      0,
				DY:      0,
				Pan:     0,
				Wheel:   0,
			},
			expectedReport: []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "right down",
			inputState: mouse.InputState{
				Buttons: mouse.Btn_Right,
				DX:      0,
				DY:      0,
				Pan:     0,
				Wheel:   0,
			},
			expectedReport: []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "all down",
			inputState: mouse.InputState{
				Buttons: mouse.Btn_Right | mouse.Btn_Left | mouse.Btn_Middle | mouse.Btn_Back | mouse.Btn_Forward,
				DX:      0,
				DY:      0,
				Pan:     0,
				Wheel:   0,
			},
			expectedReport: []byte{0x1f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "Move 50 xy",
			inputState: mouse.InputState{
				Buttons: 0,
				DX:      50,
				DY:      50,
				Pan:     0,
				Wheel:   0,
			},
			expectedReport: []byte{0x00, 0x32, 0x00, 0x32, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "Move -50 xy",
			inputState: mouse.InputState{
				Buttons: 0,
				DX:      -50,
				DY:      -50,
				Pan:     0,
				Wheel:   0,
			},
			expectedReport: []byte{0x00, 0xce, 0xff, 0xce, 0xff, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "Wheel up 1",
			inputState: mouse.InputState{
				Buttons: 0,
				DX:      0,
				DY:      0,
				Pan:     0,
				Wheel:   1,
			},
			expectedReport: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		},
		{
			name: "Wheel down 1",
			inputState: mouse.InputState{
				Buttons: 0,
				DX:      0,
				DY:      0,
				Pan:     0,
				Wheel:   -1,
			},
			expectedReport: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00},
		},
		{
			name: "Pan right 1",
			inputState: mouse.InputState{
				Buttons: 0,
				DX:      0,
				DY:      0,
				Pan:     1,
				Wheel:   0,
			},
			expectedReport: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00},
		},
		{
			name: "Move right 100, down 50, left button",
			inputState: mouse.InputState{
				Buttons: mouse.Btn_Left,
				DX:      100,
				DY:      50,
				Pan:     0,
				Wheel:   0,
			},
			expectedReport: []byte{0x01, 0x64, 0x00, 0x32, 0x00, 0x00, 0x00, 0x00, 0x00},
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
	stream, _, err := client.AddDeviceAndConnect(context.Background(), b.BusID(), "mouse", nil)
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
