package steamdeck_test

import (
	"context"
	"testing"
	"time"

	viiperTesting "github.com/Alia5/VIIPER/_testing"
	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/device/steamdeck"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/virtualbus"
	"github.com/stretchr/testify/assert"

	_ "github.com/Alia5/VIIPER/internal/registry" // Register devices
)

func TestInputReports(t *testing.T) {

	type testCase struct {
		name           string
		inputState     steamdeck.InputState
		expectedReport []byte
	}

	cases := []testCase{
		{
			name:       "no inputs",
			inputState: steamdeck.InputState{},
			expectedReport: []byte{
				0x01, 0x00, 0x09, 0x40, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "buttons a+b",
			inputState: steamdeck.InputState{
				Buttons: steamdeck.ButtonA | steamdeck.ButtonB,
			},
			expectedReport: []byte{
				0x01, 0x00, 0x09, 0x40, 0x00, 0x00, 0x00, 0x00,
				0xA0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "left stick only",
			inputState: steamdeck.InputState{
				LeftStickX: 1234,
				LeftStickY: -2345,
			},
			expectedReport: []byte{
				0x01, 0x00, 0x09, 0x40, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0xD2, 0x04, 0xD7, 0xF6, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "buttons and left stick",
			inputState: steamdeck.InputState{
				Buttons:    steamdeck.ButtonDPadUp | steamdeck.ButtonSteam,
				LeftStickX: -32768,
				LeftStickY: 32767,
			},
			expectedReport: []byte{
				0x01, 0x00, 0x09, 0x40, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x21, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x80, 0xFF, 0x7F, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
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
	stream, _, err := client.AddDeviceAndConnect(context.Background(), b.BusID(), "steamdeck", nil)
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
			assert.Len(t, tc.expectedReport, int(steamdeck.ValveInReportLength))
			assert.Equal(t, byte(steamdeck.ValveInReportMsgVersion), tc.expectedReport[0])
			assert.Equal(t, byte(steamdeck.ValveInReportMsgVersion>>8), tc.expectedReport[1])
			assert.Equal(t, steamdeck.ValveInReportTypeControllerDeckState, tc.expectedReport[2])
			assert.Equal(t, steamdeck.ValveInReportLength, tc.expectedReport[3])

			if !assert.NoError(t, stream.WriteBinary(&tc.inputState)) {
				return
			}

			deadline := time.Now().Add(750 * time.Millisecond)
			var last []byte
			for {
				got, err := usbipClient.ReadInputReport(imp.Conn)
				if !assert.NoError(t, err) {
					return
				}
				last = got
				if steamDeckReportsEqualIgnoringPacketCounter(tc.expectedReport, got) {
					break
				}
				if time.Now().After(deadline) {
					assert.Failf(t, "timed out waiting for matching report", "last=%x want=%x", last, tc.expectedReport)
					return
				}
				time.Sleep(1 * time.Millisecond)
			}
		})
	}

}

func steamDeckReportsEqualIgnoringPacketCounter(want, got []byte) bool {
	if len(want) != len(got) {
		return false
	}
	for i := range want {
		if i >= steamdeck.ValveInReportPacketNumOff && i < steamdeck.ValveInReportPayloadOff {
			continue
		}
		if want[i] != got[i] {
			return false
		}
	}
	return true
}
