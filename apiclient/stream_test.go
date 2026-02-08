package apiclient_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	apiclient "github.com/Alia5/VIIPER/apiclient"
	apitypes "github.com/Alia5/VIIPER/apitypes"
	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/xbox360"
	htesting "github.com/Alia5/VIIPER/internal/_testing"
	"github.com/Alia5/VIIPER/internal/log"
	api "github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/auth"
	handler "github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/internal/server/usb"
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
					func(o *device.CreateOptions) (pusb.Device, error) { return xbox360.New(o) },
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

func TestEncryptedStream(t *testing.T) {
	type testCase struct {
		name          string
		password      string
		serverHandler func(t *testing.T, conn net.Conn)
		streamPath    string
		expectedErr   error
	}

	echoStreamHandler := func(t *testing.T, conn net.Conn) {
		defer conn.Close()
		r := bufio.NewReader(conn)

		key, err := auth.DeriveKey("test123")
		assert.NoError(t, err)

		clientNonce, serverNonce, err := auth.HandleAuthHandshake(r, conn, key, false)
		if err != nil {
			var apiErr apitypes.ApiError
			if errors.As(err, &apiErr) {
				b, err := json.Marshal(apiErr)
				if err != nil {
					slog.Error("failed to marshal api error", "error", err)
					return
				}
				_, _ = conn.Write(append(b, '\n'))
				return
			}
			return
		}

		sessionKey := auth.DeriveSessionKey(key, serverNonce, clientNonce)
		secureConn, err := auth.WrapConn(conn, sessionKey)
		assert.NoError(t, err)

		rr := bufio.NewReader(secureConn)
		line, err := rr.ReadString('\x00')
		if err != nil {
			return
		}

		assert.Equal(t, "bus/1/1\x00", line)
	}

	cases := []testCase{
		{
			name:          "success",
			password:      "test123",
			serverHandler: echoStreamHandler,
			streamPath:    "bus/1/1",
		},
		{
			name:          "wrong password",
			password:      "wrongpass",
			serverHandler: echoStreamHandler,
			expectedErr:   errors.New("401 Unauthorized: invalid password"),
		},
		{
			name:     "bad handshake response",
			password: "test123",
			serverHandler: func(t *testing.T, conn net.Conn) {
				defer conn.Close()
				_, _ = conn.Write([]byte("NO\x00" + strings.Repeat("x", 32)))
			},
			expectedErr: errors.New(""),
		},
		{
			name:     "server closes early",
			password: "test123",
			serverHandler: func(t *testing.T, conn net.Conn) {
				_ = conn.Close()
			},
			expectedErr: errors.New(""),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			assert.NoError(t, err)
			defer ln.Close()

			go func() {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				tc.serverHandler(t, conn)
			}()

			cfg := &apiclient.Config{
				DialTimeout:  3 * time.Second,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 5 * time.Second,
				Password:     tc.password,
			}
			client := apiclient.NewWithConfig(ln.Addr().String(), cfg)
			stream, err := client.OpenStream(context.Background(), 1, "1")

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, stream)
			_ = stream.Close()
		})
	}
}
