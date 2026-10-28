package hidproxy_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	viiperTesting "github.com/Alia5/VIIPER/_testing"
	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/hidproxy"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/viiperclient"
	"github.com/Alia5/VIIPER/virtualbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Through the real server: the host reads the client's report once, and nothing after it (no repeat every 1 ms).
func TestTheHostReadsEachReportOnce(t *testing.T) {
	s := viiperTesting.NewTestServer(t)
	defer s.UsbServer.Close() // nolint
	defer s.ApiServer.Close() // nolint
	r := s.ApiServer.Router()
	r.Register("bus/{id}/add", handler.BusDeviceAdd(s.UsbServer, s.ApiServer))
	r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(s.UsbServer))
	require.NoError(t, s.ApiServer.Start())
	b, err := virtualbus.NewWithBusID(1)
	require.NoError(t, err)
	defer b.Close() // nolint
	require.NoError(t, s.UsbServer.AddBus(b))

	options, err := json.Marshal(hidproxy.Options{
		Device:  "120100020000004001000200000101020301",
		Config:  "09022900010100A032090400000203000000092111010001221400070581034000010705010340000" + "1",
		Strings: map[string]string{"1": "Maker", "2": "Thing", "3": "S1"},
		Reports: map[string]string{"0": "0600FF0901A10185010901150026FF0075089508810285020902B102C0"},
	})
	require.NoError(t, err)
	c := viiperclient.New(s.ApiServer.Addr())
	stream, _, err := c.AddDeviceAndConnect(context.Background(), b.BusID(), "hidproxy", &device.CreateOptions{DeviceSpecific: string(options)})
	require.NoError(t, err)
	defer stream.Close() //nolint:errcheck

	usbip := viiperTesting.NewUsbIpClient(t, s.UsbServer.Addr())
	var devs []viiperTesting.Device
	require.Eventually(t, func() bool {
		devs, err = usbip.ListDevices()
		return err == nil && len(devs) == 1
	}, time.Second, 10*time.Millisecond)
	imp, err := usbip.AttachDevice(devs[0].BusID)
	require.NoError(t, err)
	defer imp.Conn.Close() // nolint

	frame, err := hidproxy.AppendFrame(nil, hidproxy.Frame{Kind: hidproxy.FrameInput, Interface: 0, Payload: []byte{1, 0xAB, 0xCD}})
	require.NoError(t, err)
	_, err = stream.Write(frame)
	require.NoError(t, err)

	report, err := usbip.ReadInputReportWithTimeout(imp.Conn, time.Second)
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 0xAB, 0xCD}, report)

	_, err = usbip.ReadInputReportWithTimeout(imp.Conn, 100*time.Millisecond)
	var timeout net.Error
	assert.True(t, errors.As(err, &timeout) && timeout.Timeout(), "the report came again: %v", err)
}
