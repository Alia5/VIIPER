package apiclient_test

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"testing"
	"time"

	apiclient "github.com/Alia5/VIIPER/apiclient"
	apitypes "github.com/Alia5/VIIPER/apitypes"
	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/xbox360"
	"github.com/Alia5/VIIPER/internal/log"
	api "github.com/Alia5/VIIPER/internal/server/api"
	handler "github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/internal/server/usb"
	htesting "github.com/Alia5/VIIPER/internal/testing"
	pusb "github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/virtualbus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenStream_NotSupportedWithMockTransport(t *testing.T) {
	c := testClient(map[string]string{}, nil)
	_, err := c.OpenStream(context.Background(), 1, "1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported with mock transport")
}

func TestAddDeviceAndConnect(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(responses map[string]string) error
		wantDevice    *apitypes.Device
		wantErrSubstr string
	}{
		{
			name: "success parse then stream error",
			setup: func(responses map[string]string) error {
				responses["bus/{id}/add"] = `{"busId":42,"devId":"7","vid":"0x1234","pid":"0xabcd","type":"test"}`
				return nil
			},
			wantDevice:    &apitypes.Device{BusID: 42, DevId: "7", Vid: "0x1234", Pid: "0xabcd", Type: "test"},
			wantErrSubstr: "not supported with mock transport",
		},
		{
			name:          "transport dial error",
			setup:         func(responses map[string]string) error { return errors.New("dial fail") },
			wantErrSubstr: "dial fail",
		},
		{
			name:          "blank response error",
			setup:         func(responses map[string]string) error { return nil }, // no key => blank
			wantErrSubstr: "empty response",
		},
		{
			name: "api error response",
			setup: func(responses map[string]string) error {
				responses["bus/{id}/add"] = `{"status":404,"title":"Not Found","detail":"bus 42 not found"}`
				return nil
			},
			wantErrSubstr: "bus 42 not found",
		},
		{
			name: "strict JSON decode error (extra field)",
			setup: func(responses map[string]string) error {
				responses["bus/{id}/add"] = `{"busId":42,"devId":"7","vid":"0x1234","pid":"0xabcd","type":"test","extra":true}`
				return nil
			},
			wantErrSubstr: "decode:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responses := map[string]string{}
			errInject := error(nil)
			if e := tt.setup(responses); e != nil {
				errInject = e
			}
			c := testClient(responses, errInject)
			stream, resp, err := c.AddDeviceAndConnect(context.Background(), 42, "test", nil)
			if tt.wantDevice != nil {
				assert.Nil(t, stream)
				require.NotNil(t, resp, "device response should be parsed")
				assert.Equal(t, tt.wantDevice.DevId, resp.DevId)
				assert.Equal(t, tt.wantDevice.BusID, resp.BusID)
				assert.Equal(t, tt.wantDevice.Vid, resp.Vid)
				assert.Equal(t, tt.wantDevice.Pid, resp.Pid)
				assert.Equal(t, tt.wantDevice.Type, resp.Type)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrSubstr)
				return
			}
			assert.Nil(t, resp)
			assert.Nil(t, stream)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrSubstr)
		})
	}
}

func TestDeviceStream_Operations(t *testing.T) {
	type operation func(t *testing.T, stream *apiclient.DeviceStream)

	tests := []struct {
		name               string
		busID              uint32
		customRegistration bool
		op                 operation
	}{
		{
			name:  "read deadline timeout",
			busID: 201,
			op: func(t *testing.T, stream *apiclient.DeviceStream) {
				// Force immediate timeout by setting deadline in the past.
				require.NoError(t, stream.SetReadDeadline(time.Now().Add(-10*time.Millisecond)))
				buf := make([]byte, 2)
				_, readErr := stream.Read(buf)
				assert.Error(t, readErr)
				if ne, ok := readErr.(net.Error); ok {
					assert.True(t, ne.Timeout(), "expected timeout error")
				} else {
					assert.Fail(t, "expected net.Error timeout, got %v", readErr)
				}
				_ = stream.Close()
			},
		},
		{
			name:               "closed stream read/write errors",
			busID:              202,
			customRegistration: true,
			op: func(t *testing.T, stream *apiclient.DeviceStream) {
				require.NoError(t, stream.Close())
				buf := make([]byte, 1)
				_, rErr := stream.Read(buf)
				assert.Error(t, rErr)
				assert.Contains(t, rErr.Error(), "stream closed")
				_, wErr := stream.Write([]byte{0x01})
				assert.Error(t, wErr)
				assert.Contains(t, wErr.Error(), "stream closed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usbSrv := usb.New(usb.ServerConfig{Addr: "127.0.0.1:0"}, slog.Default(), log.NewRaw(nil))
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			addr := ln.Addr().String()
			_ = ln.Close()
			apiCfg := api.ServerConfig{Addr: addr, DeviceHandlerConnectTimeout: 500 * time.Millisecond}
			apiSrv := api.New(usbSrv, addr, apiCfg, slog.Default())
			r := apiSrv.Router()
			if tt.customRegistration {
				testReg := htesting.CreateMockRegistration(t, "xbox360",
					func(o *device.CreateOptions) pusb.Device { return xbox360.New(o) },
					func(conn net.Conn, devPtr *pusb.Device, l *slog.Logger) error {
						<-time.After(50 * time.Millisecond)
						return nil
					},
				)
				api.RegisterDevice("xbox360", testReg)
			}
			r.Register("bus/{id}/add", handler.BusDeviceAdd(usbSrv, apiSrv))
			r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(usbSrv))
			require.NoError(t, apiSrv.Start())
			defer apiSrv.Close()

			b, err := virtualbus.NewWithBusId(tt.busID)
			require.NoError(t, err)
			require.NoError(t, usbSrv.AddBus(b))

			c := apiclient.New(addr)
			stream, devResp, err := c.AddDeviceAndConnect(context.Background(), tt.busID, "xbox360", nil)
			require.NoError(t, err)
			require.NotNil(t, devResp)
			require.NotNil(t, stream)

			tt.op(t, stream)
		})
	}
}
