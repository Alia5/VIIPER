package api_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	viiperTesting "github.com/Alia5/VIIPER/_testing"
	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/xbox360"
	th "github.com/Alia5/VIIPER/internal/_testing"
	"github.com/Alia5/VIIPER/internal/log"
	_ "github.com/Alia5/VIIPER/internal/registry" // Register devices
	"github.com/Alia5/VIIPER/internal/server/api"
	apierror "github.com/Alia5/VIIPER/internal/server/api/error"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	srvusb "github.com/Alia5/VIIPER/internal/server/usb"
	pusb "github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usbip"
	"github.com/Alia5/VIIPER/virtualbus"
)

func TestAPIServer_StreamHandlerError_ClosesConn(t *testing.T) {
	cfg := srvusb.ServerConfig{Addr: "127.0.0.1:0"}
	usbSrv := srvusb.New(cfg, slog.Default(), log.NewRaw(nil))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	_ = ln.Close()

	apiSrv := api.New(usbSrv, addr, api.ServerConfig{Addr: addr}, slog.Default())
	r := apiSrv.Router()
	r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(usbSrv))
	require.NoError(t, apiSrv.Start())
	defer apiSrv.Close()

	bus, err := virtualbus.NewWithBusId(70002)
	require.NoError(t, err)
	require.NoError(t, usbSrv.AddBus(bus))
	dev := xbox360.New(nil)
	_, err = bus.Add(dev)
	require.NoError(t, err)

	var devID string
	metas := bus.GetAllDeviceMetas()
	require.Greater(t, len(metas), 0)
	for _, m := range metas {
		devID = fmt.Sprintf("%d", m.Meta.DevId)
	}
	require.NotEmpty(t, devID)

	sentinel := fmt.Errorf("boom")
	mr := th.CreateMockRegistration(t, "xbox360_error_stream",
		func(o *device.CreateOptions) pusb.Device { return xbox360.New(o) },
		func(conn net.Conn, d *pusb.Device, l *slog.Logger) error { return sentinel },
	)

	api.RegisterDevice("xbox360_error_stream", mr)
	c, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	_, err = fmt.Fprintf(c, "bus/%d/%s\n", bus.BusID(), devID)
	require.NoError(t, err)

	buf := make([]byte, 1)
	_ = c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, readErr := c.Read(buf)
	require.Error(t, readErr)
	_ = c.Close()
}

func TestAPIServer_WrappedConn(t *testing.T) {

	type testCase struct {
		name             string
		requireLocalAuth bool
		serverPass       string
		clientPass       string
		expectedResponse string
		expectedErr      error

		inputState     xbox360.InputState
		expectedReport []byte
		rumbleState    xbox360.XRumbleState
		outPacket      []byte
	}

	tests := []testCase{
		{
			name:             "SUCCESS unauthenticated",
			requireLocalAuth: false,
			serverPass:       "",
			clientPass:       "",
			expectedResponse: "xbox360",
			inputState: xbox360.InputState{
				Buttons: xbox360.ButtonDPadUp,
			},
			expectedReport: []byte{0x00, 0x14, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			rumbleState: xbox360.XRumbleState{
				LeftMotor:  236,
				RightMotor: 65,
			},
			outPacket: []byte{0x00, 0x08, 0x00, 236, 65, 0x00, 0x00, 0x00},
		},
		{
			name:             "SUCCESS authenticated (required)",
			requireLocalAuth: true,
			serverPass:       "test123",
			clientPass:       "test123",
			expectedResponse: "xbox360",
			inputState: xbox360.InputState{
				Buttons: xbox360.ButtonDPadUp,
			},
			expectedReport: []byte{0x00, 0x14, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			rumbleState: xbox360.XRumbleState{
				LeftMotor:  236,
				RightMotor: 65,
			},
			outPacket: []byte{0x00, 0x08, 0x00, 236, 65, 0x00, 0x00, 0x00},
		},
		{
			name:             "SUCCESS authenticated (optional)",
			requireLocalAuth: false,
			serverPass:       "test123",
			clientPass:       "test123",
			expectedResponse: "xbox360",
			inputState: xbox360.InputState{
				Buttons: xbox360.ButtonDPadUp,
			},
			expectedReport: []byte{0x00, 0x14, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			rumbleState: xbox360.XRumbleState{
				LeftMotor:  236,
				RightMotor: 65,
			},
			outPacket: []byte{0x00, 0x08, 0x00, 236, 65, 0x00, 0x00, 0x00},
		},
		{
			name:             "Invalid password",
			requireLocalAuth: false,
			serverPass:       "test123",
			clientPass:       "wrongpass",
			expectedErr:      apierror.ErrUnauthorized("invalid password"),
		},
		{
			name:             "Auth Required",
			requireLocalAuth: true,
			serverPass:       "test123",
			clientPass:       "",
			expectedErr:      apierror.ErrUnauthorized("authentication required"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testSrvConfig := viiperTesting.TestServerConfig(t)

			testSrvConfig.Server.ApiServerConfig.RequireLocalHostAuth = tc.requireLocalAuth
			testSrvConfig.Server.ApiServerConfig.Password = tc.serverPass

			s := viiperTesting.NewTestServerWithConfig(t, testSrvConfig)
			defer s.ApiServer.Close()
			defer s.UsbServer.Close()

			r := s.ApiServer.Router()
			r.Register("bus/{id}/add", handler.BusDeviceAdd(s.UsbServer, s.ApiServer))
			r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(s.UsbServer))

			if err := s.ApiServer.Start(); err != nil {
				t.Fatalf("Failed to start API server: %v", err)
			}
			time.Sleep(50 * time.Millisecond)

			b, err := virtualbus.NewWithBusId(1)
			if err != nil {
				t.Fatalf("Failed to create virtual bus: %v", err)
			}
			defer b.Close()
			_ = s.UsbServer.AddBus(b)
			time.Sleep(50 * time.Millisecond)

			client := apiclient.NewWithPassword(s.ApiServer.Addr(), tc.clientPass)

			time.Sleep(50 * time.Millisecond)

			resp, err := client.DeviceAdd(b.BusID(), "xbox360", nil)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			} else {
				if !assert.NoError(t, err) {
					return
				}
			}

			assert.Equal(t, tc.expectedResponse, resp.Type)

			time.Sleep(50 * time.Millisecond)

			stream, err := client.OpenStream(context.Background(), b.BusID(), resp.DevId)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			} else {
				if !assert.NoError(t, err) {
					return
				}
			}
			defer stream.Close()
			time.Sleep(50 * time.Millisecond)

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

			time.Sleep(50 * time.Millisecond)

			assert.Equal(t, tc.expectedReport, tc.inputState.BuildReport())
			if !assert.NoError(t, stream.WriteBinary(&tc.inputState)) {
				return
			}

			got, err := usbipClient.PollInputReport(imp.Conn, tc.expectedReport, 750*time.Millisecond)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, tc.expectedReport, got)

			if !assert.NoError(t, usbipClient.Submit(imp.Conn, usbip.DirOut, 1, tc.outPacket, nil)) {
				return
			}
			var buf [2]byte
			_ = stream.SetReadDeadline(time.Now().Add(750 * time.Millisecond))
			_, err = io.ReadFull(stream, buf[:])
			if !assert.NoError(t, err) {
				return
			}
			gotOut := xbox360.XRumbleState{LeftMotor: buf[0], RightMotor: buf[1]}
			assert.Equal(t, tc.rumbleState, gotOut)

		})
	}

}
