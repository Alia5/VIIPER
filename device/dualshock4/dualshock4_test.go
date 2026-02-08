package dualshock4_test

import (
	"context"
	"encoding/binary"
	"io"
	"testing"
	"time"

	viiperTesting "github.com/Alia5/VIIPER/_testing"
	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/device/dualshock4"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/usbip"
	"github.com/Alia5/VIIPER/virtualbus"
	"github.com/stretchr/testify/assert"

	_ "github.com/Alia5/VIIPER/internal/registry" // Register devices
)

func TestInputReports(t *testing.T) {
	type testCase struct {
		name           string
		inputState     dualshock4.InputState
		expectedReport []byte
	}

	cases := []testCase{
		{
			name: "neutral defaults",
			inputState: dualshock4.InputState{
				LX:           0,
				LY:           0,
				RX:           0,
				RY:           0,
				Buttons:      0,
				DPad:         0,
				L2:           0,
				R2:           0,
				Touch1X:      0,
				Touch1Y:      0,
				Touch1Active: false,
				Touch2X:      0,
				Touch2Y:      0,
				Touch2Active: false,
				GyroX:        0,
				GyroY:        0,
				GyroZ:        0,
				AccelX:       0,
				AccelY:       0,
				AccelZ:       0,
			},
			expectedReport: []byte{
				0x01,
				0x80, 0x80, 0x80, 0x80,
				0x08,
				0x00,
				0x00,
				0x00, 0x00,
				0x00, 0x00,
				0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00,
				0x0b,
				0x00, 0x00, 0x00, 0x00,
				0x80, 0x00, 0x00, 0x00,
				0x80, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00,
			},
		},
		{
			name: "dpad up",
			inputState: dualshock4.InputState{
				LX:           0,
				LY:           0,
				RX:           0,
				RY:           0,
				Buttons:      0,
				DPad:         dualshock4.DPadUp,
				Touch1Active: false,
				Touch2Active: false,
			},
			expectedReport: func() []byte {
				b := make([]byte, dualshock4.InputReportSize)
				b[0] = 0x01
				b[1], b[2], b[3], b[4] = 0x80, 0x80, 0x80, 0x80
				b[5] = 0x00
				b[30] = 0x0b
				b[35] = 0x80
				b[39] = 0x80
				return b
			}(),
		},
		{
			name: "buttons - square",
			inputState: dualshock4.InputState{
				LX:           0,
				LY:           0,
				RX:           0,
				RY:           0,
				Buttons:      uint16(dualshock4.ButtonSquare),
				DPad:         0,
				Touch1Active: false,
				Touch2Active: false,
			},
			expectedReport: func() []byte {
				b := make([]byte, dualshock4.InputReportSize)
				b[0] = 0x01
				b[1], b[2], b[3], b[4] = 0x80, 0x80, 0x80, 0x80
				b[5] = 0x18
				b[30] = 0x0b
				b[35] = 0x80
				b[39] = 0x80
				return b
			}(),
		},
		{
			name: "buttons - ps",
			inputState: dualshock4.InputState{
				LX:           0,
				LY:           0,
				RX:           0,
				RY:           0,
				Buttons:      dualshock4.ButtonPS,
				DPad:         0,
				Touch1Active: false,
				Touch2Active: false,
			},
			expectedReport: func() []byte {
				b := make([]byte, dualshock4.InputReportSize)
				b[0] = 0x01
				b[1], b[2], b[3], b[4] = 0x80, 0x80, 0x80, 0x80
				b[5] = 0x08
				b[7] = 0x01
				b[30] = 0x0b
				b[35] = 0x80
				b[39] = 0x80
				return b
			}(),
		},
		{
			name: "triggers - l2/r2",
			inputState: dualshock4.InputState{
				LX:           0,
				LY:           0,
				RX:           0,
				RY:           0,
				Buttons:      0,
				DPad:         0,
				L2:           0x12,
				R2:           0xFE,
				Touch1Active: false,
				Touch2Active: false,
			},
			expectedReport: func() []byte {
				b := make([]byte, dualshock4.InputReportSize)
				b[0] = 0x01
				b[1], b[2], b[3], b[4] = 0x80, 0x80, 0x80, 0x80
				b[5] = 0x08
				b[8] = 0x12
				b[9] = 0xFE
				b[30] = 0x0b
				b[35] = 0x80
				b[39] = 0x80
				return b
			}(),
		},
		{
			name: "touch1 active with coords",
			inputState: dualshock4.InputState{
				LX:           0,
				LY:           0,
				RX:           0,
				RY:           0,
				Buttons:      0,
				DPad:         0,
				Touch1X:      123,
				Touch1Y:      456,
				Touch1Active: true,
				Touch2Active: false,
			},
			expectedReport: func() []byte {
				b := make([]byte, dualshock4.InputReportSize)
				b[0] = 0x01
				b[1], b[2], b[3], b[4] = 0x80, 0x80, 0x80, 0x80
				b[5] = 0x08
				b[30] = 0x0b
				b[35] = 0x00
				b[36] = 0x7b
				b[37] = 0x80
				b[38] = 0x1c
				b[39] = 0x80
				return b
			}(),
		},
		{
			name: "sensors",
			inputState: dualshock4.InputState{
				LX:           0,
				LY:           0,
				RX:           0,
				RY:           0,
				Buttons:      0,
				DPad:         0,
				GyroX:        1234,
				GyroY:        -2345,
				GyroZ:        3456,
				AccelX:       -111,
				AccelY:       222,
				AccelZ:       -333,
				Touch1Active: false,
				Touch2Active: false,
			},
			expectedReport: func() []byte {
				b := make([]byte, dualshock4.InputReportSize)
				b[0] = 0x01
				b[1], b[2], b[3], b[4] = 0x80, 0x80, 0x80, 0x80
				b[5] = 0x08
				b[13], b[14] = 0xD2, 0x04
				b[15], b[16] = 0xD7, 0xF6
				b[17], b[18] = 0x80, 0x0D
				b[19], b[20] = 0x91, 0xFF
				b[21], b[22] = 0xDE, 0x00
				b[23], b[24] = 0xB3, 0xFE
				b[30] = 0x0b
				b[35] = 0x80
				b[39] = 0x80
				return b
			}(),
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
	stream, _, err := client.AddDeviceAndConnect(context.Background(), b.BusID(), "dualshock4", nil)
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

	var seq uint32
	readInputReport := func(timeout time.Duration) ([]byte, error) {
		seq++
		cmd := usbip.CmdSubmit{
			Basic:             usbip.HeaderBasic{Command: usbip.CmdSubmitCode, Seqnum: seq, Devid: 0, Dir: usbip.DirIn, Ep: 4},
			TransferFlags:     0,
			TransferBufferLen: 255,
			StartFrame:        0,
			NumberOfPackets:   0,
			Interval:          0,
			Setup:             [8]byte{},
		}
		_ = imp.Conn.SetDeadline(time.Now().Add(timeout))
		if err := cmd.Write(imp.Conn); err != nil {
			return nil, err
		}
		var retHdr [48]byte
		if err := usbip.ReadExactly(imp.Conn, retHdr[:]); err != nil {
			return nil, err
		}
		if gotCmd := binary.BigEndian.Uint32(retHdr[0:4]); gotCmd != usbip.RetSubmitCode {
			return nil, io.ErrUnexpectedEOF
		}
		status := int32(binary.BigEndian.Uint32(retHdr[20:24]))
		actual := binary.BigEndian.Uint32(retHdr[24:28])
		if status != 0 {
			return nil, io.ErrUnexpectedEOF
		}
		data := make([]byte, int(actual))
		if actual > 0 {
			if err := usbip.ReadExactly(imp.Conn, data); err != nil {
				return nil, err
			}
		}
		_ = imp.Conn.SetDeadline(time.Time{})
		return data, nil
	}

	pollInputReport := func(want []byte, timeout time.Duration) ([]byte, error) {
		deadline := time.Now().Add(timeout)
		var last []byte
		for {
			got, err := readInputReport(250 * time.Millisecond)
			if err != nil {
				return nil, err
			}
			last = got
			if len(got) == len(want) {
				gg := append([]byte(nil), got...)
				ww := append([]byte(nil), want...)
				gg[7] &= 0x03
				ww[7] &= 0x03
				gg[10], gg[11] = 0, 0
				ww[10], ww[11] = 0, 0
				if assert.ObjectsAreEqual(ww, gg) {
					return got, nil
				}
			}
			if time.Now().After(deadline) {
				return last, nil
			}
			time.Sleep(1 * time.Millisecond)
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dev, err := dualshock4.New(nil)
			if !assert.NoError(t, err) {
				return
			}
			dev.UpdateInputState(&tc.inputState)
			built := dev.HandleTransfer(4, usbip.DirIn, nil)
			bb := append([]byte(nil), built...)
			exp := append([]byte(nil), tc.expectedReport...)
			bb[7] &= 0x03
			exp[7] &= 0x03
			bb[10], bb[11] = 0, 0
			exp[10], exp[11] = 0, 0
			assert.Equal(t, exp, bb)

			if !assert.NoError(t, stream.WriteBinary(&tc.inputState)) {
				return
			}
			got, err := pollInputReport(tc.expectedReport, 750*time.Millisecond)
			if !assert.NoError(t, err) {
				return
			}
			if !assert.Len(t, got, dualshock4.InputReportSize) {
				return
			}
			gg := append([]byte(nil), got...)
			gg[7] &= 0x03
			gg[10], gg[11] = 0, 0
			assert.Equal(t, exp, gg)
		})
	}
}

func TestFeedback(t *testing.T) {
	type testCase struct {
		name        string
		outputState dualshock4.OutputState
		outPacket   []byte
	}

	cases := []testCase{
		{
			name: "off",
			outputState: dualshock4.OutputState{
				RumbleSmall: 0,
				RumbleLarge: 0,
				LedRed:      0,
				LedGreen:    0,
				LedBlue:     0,
				FlashOn:     0,
				FlashOff:    0,
			},
			outPacket: []byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name: "rumble + led + flash",
			outputState: dualshock4.OutputState{
				RumbleSmall: 0x12,
				RumbleLarge: 0xFE,
				LedRed:      0x01,
				LedGreen:    0x02,
				LedBlue:     0x03,
				FlashOn:     0x04,
				FlashOff:    0x05,
			},
			outPacket: []byte{0x05, 0x00, 0x00, 0x00, 0x12, 0xFE, 0x01, 0x02, 0x03, 0x04, 0x05},
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
	stream, _, err := client.AddDeviceAndConnect(context.Background(), b.BusID(), "dualshock4", nil)
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
			if !assert.NoError(t, usbipClient.Submit(imp.Conn, usbip.DirOut, 3, tc.outPacket, nil)) {
				return
			}
			var buf [7]byte
			_ = stream.SetReadDeadline(time.Now().Add(750 * time.Millisecond))
			_, err := io.ReadFull(stream, buf[:])
			if !assert.NoError(t, err) {
				return
			}
			got := dualshock4.OutputState{}
			if !assert.NoError(t, got.UnmarshalBinary(buf[:])) {
				return
			}
			assert.Equal(t, tc.outputState, got)
		})
	}
}
