package handler_test

import (
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"viiper/internal/log"
	"viiper/internal/server/api"
	"viiper/internal/server/api/handler"
	"viiper/internal/server/usb"
	handlerTest "viiper/internal/testing"
	"viiper/pkg/apiclient"
	"viiper/pkg/device/xbox360"
	pusb "viiper/pkg/usb"
	"viiper/pkg/virtualbus"
)

// testRegistration implements api.DeviceRegistration for tests
type testRegistration struct {
	creator func() pusb.Device
	handler api.StreamHandlerFunc
}

func (t *testRegistration) CreateDevice() pusb.Device            { return t.creator() }
func (t *testRegistration) StreamHandler() api.StreamHandlerFunc { return t.handler }

func TestBusDeviceAdd(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T, s *usb.Server)
		pathParams       map[string]string
		payload          any
		expectedResponse string
	}{
		{
			name: "add device to existing bus",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(80001)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "80001"},
			payload:          "xbox360",
			expectedResponse: `{"id":"80001-1"}`,
		},
		{
			name:             "add device to non-existing bus",
			setup:            nil,
			pathParams:       map[string]string{"id": "99999"},
			payload:          "xbox360",
			expectedResponse: `{"error":"unknown bus"}`,
		},
		{
			name:             "invalid bus number",
			setup:            nil,
			pathParams:       map[string]string{"id": "baz"},
			payload:          "xbox360",
			expectedResponse: `{"error":"strconv.ParseUint: parsing \"baz\": invalid syntax"}`,
		},
		{
			name: "correct device id after add/remove",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(80005)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
				if _, err := b.Add(xbox360.New()); err != nil {
					t.Fatalf("add device failed: %v", err)
				}
				if err := b.RemoveDeviceByID("1"); err != nil {
					t.Fatalf("remove device failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "80005"},
			payload:          "xbox360",
			expectedResponse: `{"id":"80005-1"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, srv, done := handlerTest.StartAPIServer(t, func(r *api.Router, s *usb.Server, apiSrv *api.Server) {
				r.Register("bus/create", handler.BusCreate(s))
				r.Register("bus/{id}/add", handler.BusDeviceAdd(s, apiSrv))
			})
			defer done()

			c := apiclient.NewTransport(addr)
			if tt.setup != nil {
				tt.setup(t, srv)
			}
			line, err := c.Do("bus/{id}/add", tt.payload, tt.pathParams)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResponse, line)
		})
	}
}

// Verify that a device added via API is auto-removed if no stream connects within the configured timeout.
func TestBusDeviceAdd_NoConnection_TimeoutCleanup(t *testing.T) {
	// We need to control API DeviceHandlerConnectTimeout, so set up API server manually (not via StartAPIServer).
	usbSrv := usb.New(usb.ServerConfig{Addr: "127.0.0.1:0"}, slog.Default(), log.NewRaw(nil))

	// Create a bus directly on the USB server.
	b, err := virtualbus.NewWithBusId(80100)
	require.NoError(t, err)
	require.NoError(t, usbSrv.AddBus(b))

	// Choose a free TCP address for API server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	_ = ln.Close()

	// Start API server with a very short timeout
	apiCfg := api.ServerConfig{Addr: addr, DeviceHandlerConnectTimeout: 200 * time.Millisecond}
	apiSrv := api.New(usbSrv, addr, apiCfg, slog.Default())
	r := apiSrv.Router()
	r.Register("bus/{id}/add", handler.BusDeviceAdd(usbSrv, apiSrv))
	r.Register("bus/{id}/devices", handler.BusDevicesList(usbSrv))
	require.NoError(t, apiSrv.Start())
	defer apiSrv.Close()

	// Register a minimal device registration for xbox360 that creates a real device
	api.RegisterDevice("xbox360", &testRegistration{
		creator: func() pusb.Device { return xbox360.New() },
		handler: func(conn net.Conn, dev *pusb.Device, logger *slog.Logger) error { return nil },
	})

	// Use API client to add device, then wait beyond timeout and verify removal
	c := apiclient.New(addr)
	_, err = c.DeviceAdd(80100, "xbox360")
	require.NoError(t, err)

	// Immediately after add, the device should be present
	list, err := c.DevicesList(80100)
	require.NoError(t, err)
	require.Len(t, list.Devices, 1)

	// Wait slightly beyond timeout to allow auto-removal
	time.Sleep(350 * time.Millisecond)

	list2, err := c.DevicesList(80100)
	require.NoError(t, err)
	assert.Len(t, list2.Devices, 0)
}
