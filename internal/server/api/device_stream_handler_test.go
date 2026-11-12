package api_test

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/keyboard"
	htesting "github.com/Alia5/VIIPER/internal/_testing" // nolint
	th "github.com/Alia5/VIIPER/internal/_testing"       // nolint
	"github.com/Alia5/VIIPER/internal/log"
	"github.com/Alia5/VIIPER/internal/server/api"
	srvusb "github.com/Alia5/VIIPER/internal/server/usb"
	pusb "github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/virtualbus"
)

func TestDeviceStreamHandler_Dispatch(t *testing.T) {
	cfg := srvusb.ServerConfig{Addr: "127.0.0.1:0"}
	srv := srvusb.New(cfg, slog.Default(), log.NewRaw(nil))
	logger := slog.Default()

	bus, err := virtualbus.NewWithBusID(90001)
	require.NoError(t, err)
	require.NoError(t, srv.AddBus(bus))
	dev, err := keyboard.New(t.Context(), nil)
	require.NoError(t, err)
	devCtx, err := bus.Add(dev, func() {})
	require.NoError(t, err)

	meta := device.GetDeviceMeta(devCtx)
	require.NotNil(t, meta)

	var dv interface{}
	metas := bus.GetAllDeviceMetas()
	for _, m := range metas {
		var busid string
		for i, b := range m.Meta.USBBusID {
			if b == 0 {
				busid = string(m.Meta.USBBusID[:i])
				break
			}
		}
		if string(meta.USBBusID[:len(busid)]) == busid {
			dv = m.Dev
			break
		}
	}
	require.NotNil(t, dv)

	handlerCalled := make(chan bool, 1)
	testReg := th.CreateMockRegistration(t, "keyboard",
		func(ctx context.Context, o *device.CreateOptions) (pusb.Device, error) { return keyboard.New(ctx, o) },
		func(conn net.Conn, d *pusb.Device, l *slog.Logger) error {
			handlerCalled <- true
			return nil
		},
	)

	api.RegisterDevice("keyboard", testReg)

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close() //nolint:errcheck

	handler := api.DeviceStreamHandler(srv)
	dvUSB := dv.(pusb.Device)
	go func() {
		err := handler(serverConn, &dvUSB, logger)
		require.NoError(t, err)
	}()

	select {
	case <-handlerCalled:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatal("handler was not called within timeout")
	}
}

func TestAPIServer_StreamRoute_DispatchE2E(t *testing.T) {
	addr, srv, done := htesting.StartAPIServer(t, func(r *api.Router, s *srvusb.Server, apiSrv *api.Server) {
		r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(s))
	})
	defer done()

	bus, err := virtualbus.NewWithBusID(70001)
	require.NoError(t, err)
	require.NoError(t, srv.AddBus(bus))
	dev, err := keyboard.New(t.Context(), nil)
	require.NoError(t, err)
	devCtx, err := bus.Add(dev, func() {})
	require.NoError(t, err)
	meta := device.GetDeviceMeta(devCtx)
	require.NotNil(t, meta)

	var deviceID string
	for i, b := range meta.USBBusID {
		if b == 0 {
			fullId := string(meta.USBBusID[:i])
			splits := strings.Split(fullId, "-")
			deviceID = splits[1]
			break
		}
	}
	require.NotEmpty(t, deviceID)

	handlerCalled := make(chan struct{}, 1)
	testReg := th.CreateMockRegistration(t, "keyboard",
		func(ctx context.Context, o *device.CreateOptions) (pusb.Device, error) { return keyboard.New(ctx, o) },
		func(conn net.Conn, devPtr *pusb.Device, l *slog.Logger) error {
			handlerCalled <- struct{}{}
			return nil
		},
	)
	api.RegisterDevice("keyboard", testReg)

	c, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer c.Close() //nolint:errcheck

	_, err = fmt.Fprintf(c, "bus/%d/%s\x00", bus.BusID(), deviceID)
	require.NoError(t, err)

	select {
	case <-handlerCalled:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatal("stream handler was not called within timeout")
	}
}
