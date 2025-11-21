package api_test

import (
	"fmt"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"viiper/internal/log"
	"viiper/internal/server/api"
	srvusb "viiper/internal/server/usb"
	th "viiper/internal/testing"
	"viiper/pkg/device"
	"viiper/pkg/device/xbox360"
	pusb "viiper/pkg/usb"
	"viiper/pkg/virtualbus"
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
	mr := th.CreateMockRegistration(t, "xbox360",
		func(o *device.CreateOptions) pusb.Device { return xbox360.New(o) },
		func(conn net.Conn, d *pusb.Device, l *slog.Logger) error { return sentinel },
	)

	api.RegisterDevice("xbox360", mr)
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
