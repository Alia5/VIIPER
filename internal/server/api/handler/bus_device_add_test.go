package handler_test

import (
	"encoding/json"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/xbox360"
	th "github.com/Alia5/VIIPER/internal/_testing"
	"github.com/Alia5/VIIPER/internal/log"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/internal/server/usb"
	pusb "github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/virtualbus"
)

func TestBusDeviceAdd(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T, s *usb.Server, as *api.Server)
		pathParams       map[string]string
		payload          any
		expectedResponse string
		extraChecks      func(t *testing.T, response string, srv *usb.Server, apiSrv *api.Server)
	}{
		{
			name: "add device to existing bus",
			setup: func(t *testing.T, s *usb.Server, as *api.Server) {
				b, err := virtualbus.NewWithBusId(80001)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "80001"},
			payload:          `{"type": "xbox360"}`,
			expectedResponse: `{"busId":80001, "devId": "1", "deviceSpecific": {"subType": 1}, "vid":"0x045e", "pid":"0x028e", "type":"xbox360"}`,
		},
		{
			name: "add device to existing bus with device specific args",
			setup: func(t *testing.T, s *usb.Server, as *api.Server) {
				b, err := virtualbus.NewWithBusId(80001)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "80001"},
			payload:          `{"type": "xbox360", "deviceSpecific":{"subType": 7}}`,
			expectedResponse: `{"busId":80001, "devId": "1", "deviceSpecific": {"subType": 7}, "vid":"0x045e", "pid":"0x028e", "type":"xbox360"}`,
		},
		{
			name: "invalid device specific args",
			setup: func(t *testing.T, s *usb.Server, as *api.Server) {
				b, err := virtualbus.NewWithBusId(80001)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "80001"},
			payload:          `{"type": "xbox360", "deviceSpecific":{"subType": "a"}}`,
			expectedResponse: `{"detail":"failed to create device: invalid JSON payload: json: cannot unmarshal string into Go struct field Xbox360CreateOptions.subType of type uint8", "status":400, "title":"Bad Request"}`,
		},
		{
			name:             "add device to non-existing bus",
			setup:            nil,
			pathParams:       map[string]string{"id": "99999"},
			payload:          `{"type": "xbox360"}`,
			expectedResponse: `{"status":404,"title":"Not Found","detail":"bus 99999 not found"}`,
		},
		{
			name:             "invalid bus number",
			setup:            nil,
			pathParams:       map[string]string{"id": "baz"},
			payload:          `{"type": "xbox360"}`,
			expectedResponse: `{"status":400,"title":"Bad Request","detail":"invalid busId: strconv.ParseUint: parsing \"baz\": invalid syntax"}`,
		},
		{
			name: "invalid json",
			setup: func(t *testing.T, s *usb.Server, as *api.Server) {
				b, err := virtualbus.NewWithBusId(2)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "2"},
			payload:          `xbox360`,
			expectedResponse: `{"status":400,"title":"Bad Request","detail":"invalid JSON payload: invalid character 'x' looking for beginning of value"}`,
		},
		{
			name: "invalid payload",
			setup: func(t *testing.T, s *usb.Server, as *api.Server) {
				b, err := virtualbus.NewWithBusId(3)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "3"},
			payload:          `{"tpe": "xbox360"}`,
			expectedResponse: `{"status":400,"title":"Bad Request","detail":"missing device type"}`,
		},
		{
			name: "correct device id after add/remove",
			setup: func(t *testing.T, s *usb.Server, as *api.Server) {
				b, err := virtualbus.NewWithBusId(80005)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
				dev, err := xbox360.New(nil)
				if err != nil {
					t.Fatalf("create device failed: %v", err)
				}
				if _, err := b.Add(dev); err != nil {
					t.Fatalf("add device failed: %v", err)
				}
				if err := b.RemoveDeviceByID("1"); err != nil {
					t.Fatalf("remove device failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "80005"},
			payload:          `{"type": "xbox360"}`,
			expectedResponse: `{"busId":80005, "devId": "1", "deviceSpecific": {"subType":1}, "vid":"0x045e", "pid":"0x028e", "type":"xbox360"}`,
		},
		{
			name: "autoattach fails returns error",
			setup: func(t *testing.T, s *usb.Server, as *api.Server) {
				as.Config().AutoAttachLocalClient = true
				b, err := virtualbus.NewWithBusId(80250)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			pathParams: map[string]string{"id": "80250"},
			payload:    `{"type": "xbox360"}`,
			extraChecks: func(t *testing.T, response string, srv *usb.Server, apiSrv *api.Server) {
				var errResp map[string]interface{}
				err := json.Unmarshal([]byte(response), &errResp)
				require.NoError(t, err)
				require.Equal(t, float64(409), errResp["status"])
				require.Equal(t, "Conflict", errResp["title"])
				detail := errResp["detail"].(string)
				require.Contains(t, detail, "Failed to auto-attach device:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var as *api.Server
			addr, srv, done := th.StartAPIServer(t, func(r *api.Router, s *usb.Server, apiSrv *api.Server) {
				r.Register("bus/create", handler.BusCreate(s))
				r.Register("bus/{id}/add", handler.BusDeviceAdd(s, apiSrv))
				as = apiSrv
			})
			defer done()

			c := apiclient.NewTransport(addr)
			if tt.setup != nil {
				tt.setup(t, srv, as)
			}
			line, err := c.Do("bus/{id}/add", tt.payload, tt.pathParams)
			assert.NoError(t, err)

			if tt.expectedResponse != "" {
				assert.JSONEq(t, tt.expectedResponse, line)
			}

			if tt.extraChecks != nil {
				tt.extraChecks(t, line, srv, as)
			}
		})
	}
}

// Verify that a device added via API is auto-removed if no stream connects within the configured timeout.
func TestBusDeviceAdd_NoConnection_TimeoutCleanup(t *testing.T) {
	usbSrv := usb.New(usb.ServerConfig{
		Addr:              "127.0.0.1:0",
		ConnectionTimeout: time.Millisecond * 500,
		BusCleanupTimeout: time.Millisecond * 500,
	}, slog.Default(), log.NewRaw(nil))

	b, err := virtualbus.NewWithBusId(80100)
	require.NoError(t, err)
	require.NoError(t, usbSrv.AddBus(b))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	_ = ln.Close()

	apiCfg := api.ServerConfig{Addr: addr, DeviceHandlerConnectTimeout: 500 * time.Millisecond}
	apiSrv := api.New(usbSrv, addr, apiCfg, slog.Default())
	r := apiSrv.Router()
	r.Register("bus/{id}/add", handler.BusDeviceAdd(usbSrv, apiSrv))
	r.Register("bus/{id}/list", handler.BusDevicesList(usbSrv))
	require.NoError(t, apiSrv.Start())
	defer apiSrv.Close()

	testReg := th.CreateMockRegistration(t, "xbox360",
		func(o *device.CreateOptions) (pusb.Device, error) { return xbox360.New(o) },
		func(conn net.Conn, devPtr *pusb.Device, l *slog.Logger) error { return nil },
	)

	api.RegisterDevice("xbox360", testReg)

	c := apiclient.New(addr)
	_, err = c.DeviceAdd(80100, "xbox360", nil)
	require.NoError(t, err)

	list, err := c.DevicesList(80100)
	require.NoError(t, err)
	require.Len(t, list.Devices, 1)

	require.Eventually(t, func() bool {
		list, _ := c.DevicesList(80100)
		return list != nil && len(list.Devices) == 0
	}, 3*time.Second, 50*time.Millisecond)

	require.Eventually(t, func() bool {
		return len(usbSrv.ListBuses()) == 0
	}, 3*time.Second, 50*time.Millisecond)
}
