package hidproxy

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usbip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The Steam Controller puck (28DE:1304) as read from the real device: CDC ACM on
// interfaces 0 and 1, four controller slots (HID 2-5), the dongle (HID 6).
func puckOptions(t *testing.T) Options {
	t.Helper()
	raw, err := os.ReadFile("testdata/puck.json")
	require.NoError(t, err)
	var o Options
	require.NoError(t, json.Unmarshal(raw, &o))
	o.Strings["3"] = "SERIAL0000001"
	return o
}

// A small HID device: one interface with report ids, interrupt IN 0x81 and OUT 0x01, both 1 ms.
func smallOptions() Options {
	return Options{
		Device: "12010002000000400100020000010102030" + "1",
		Config: "09022900010100A032" +
			"090400000203000000" +
			"0921110100012214" + "00" +
			"0705810340000" + "1" +
			"0705010340000" + "1",
		Strings: map[string]string{"1": "Maker", "2": "Thing", "3": "S1"},
		Reports: map[string]string{"0": "0600FF0901A10185010901150026FF0075089508810285020902B102C0"},
	}
}

func TestPuckDescriptorsEnumerateLikeThePuck(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	d := p.GetDescriptor()
	assert.Equal(t, uint16(0x28DE), d.Device.IDVendor)
	assert.Equal(t, uint16(0x1304), d.Device.IDProduct)
	assert.Equal(t, uint16(0x0200), d.Device.BcdUSB, "2.01 would make Windows ask for a BOS descriptor")
	assert.Equal(t, "SERIAL0000001", d.Strings[d.Device.ISerialNumber])
	assert.Len(t, d.Interfaces, 7)
	assert.Len(t, d.Associations, 1)
	for number := uint8(2); number <= 6; number++ {
		iface, ok := d.Interface(number)
		require.True(t, ok)
		require.NotNil(t, iface.HID, "interface %d", number)
		assert.NotEmpty(t, iface.HID.ReportDescriptorBytes)
	}
	assert.Equal(t, uint32(3), p.inByIface[2], "slot 1 reports on its interrupt IN endpoint 0x83")
	assert.Equal(t, uint32(1), p.inByIface[0], "CDC: the interrupt endpoint")
	assert.Equal(t, uint8(1), p.ifaceByOut[1], "CDC data's bulk OUT")
	assert.Empty(t, p.GetDeviceSpecificArgs(), "the serial number must not be echoed")
}

func TestBadOptionsAreRefused(t *testing.T) {
	cases := map[string]func(o *Options){
		"device not hex":          func(o *Options) { o.Device = "zz" },
		"device too short":        func(o *Options) { o.Device = "1201" },
		"device wrong type":       func(o *Options) { o.Device = "12020002000000400100020000010102030" + "1" },
		"config too short":        func(o *Options) { o.Config = "0902" },
		"config bad length":       func(o *Options) { o.Config = "09022900010100A032" + "FF04" },
		"config truncated":        func(o *Options) { o.Config = "09022900010100A032" + "09" },
		"config no interfaces":    func(o *Options) { o.Config = "09020900000100A032" },
		"report missing":          func(o *Options) { o.Reports = map[string]string{} },
		"report key not a number": func(o *Options) { o.Reports["x"] = "00" },
		"report empty":            func(o *Options) { o.Reports["0"] = "" },
		"string index 0":          func(o *Options) { o.Strings["0"] = "x" },
		"string too long":         func(o *Options) { o.Strings["1"] = string(make([]rune, 127)) },
		"speed unknown":           func(o *Options) { o.Speed = 5 },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			o := smallOptions()
			change(&o)
			_, err := New(o)
			assert.Error(t, err)
		})
	}
	_, err := decodeOptions("")
	assert.Error(t, err, "no descriptors")
	_, err = decodeOptions("{")
	assert.Error(t, err)
}

// attach serves a client over a pipe and returns the client's end.
func attach(t *testing.T, p *HIDProxy) (net.Conn, <-chan error) {
	t.Helper()
	server, client := net.Pipe()
	done := make(chan error, 1)
	go func() { done <- p.Serve(server) }()
	require.Eventually(t, func() bool {
		p.mu.Lock()
		defer p.mu.Unlock()
		return p.client != nil
	}, time.Second, time.Millisecond)
	t.Cleanup(func() { _ = client.Close() })
	return client, done
}

func write(t *testing.T, conn net.Conn, f Frame) {
	t.Helper()
	buffer, err := AppendFrame(nil, f)
	require.NoError(t, err)
	_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
	_, err = conn.Write(buffer)
	require.NoError(t, err)
}

func read(t *testing.T, conn net.Conn) Frame {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	f, err := ReadFrame(conn)
	require.NoError(t, err)
	return f
}

func in(p *HIDProxy, ep uint32, wait time.Duration) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	return p.HandleTransfer(ctx, ep, usbip.DirIn, nil)
}

func TestInputReportsReachTheirInterfacesEndpoint(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, _ := attach(t, p)
	write(t, client, Frame{Kind: FrameInput, Interface: 3, Payload: []byte{0x42, 1, 2}})
	assert.Equal(t, []byte{0x42, 1, 2}, in(p, 4, time.Second), "slot 2 is interface 3, endpoint 0x84")
	assert.Nil(t, in(p, 3, 20*time.Millisecond), "slot 1 got nothing")
}

func TestASlowHostGetsTheNewestReports(t *testing.T) {
	p, err := New(smallOptions())
	require.NoError(t, err)
	for i := 0; i < inQueueLength+5; i++ {
		p.queueInput(0, []byte{1, byte(i)})
	}
	assert.Equal(t, []byte{1, 5}, in(p, 1, time.Second), "the five oldest were dropped")
}

func TestInputForAnUnknownInterfaceIsDropped(t *testing.T) {
	p, err := New(smallOptions())
	require.NoError(t, err)
	p.queueInput(9, []byte{1})
	p.queueInput(0, nil)
	assert.Nil(t, in(p, 1, 20*time.Millisecond))
}

func TestOutputGoesToTheClient(t *testing.T) {
	p, err := New(smallOptions())
	require.NoError(t, err)
	client, _ := attach(t, p)
	go p.HandleTransfer(context.Background(), 1, usbip.DirOut, []byte{0x80, 7})
	f := read(t, client)
	assert.Equal(t, FrameOutput, f.Kind)
	assert.Equal(t, uint8(0), f.Interface)
	assert.Equal(t, []byte{0x80, 7}, f.Payload)
}

func TestOutputWithoutAClientGoesNowhere(t *testing.T) {
	p, err := New(smallOptions())
	require.NoError(t, err)
	assert.Nil(t, p.HandleTransfer(context.Background(), 1, usbip.DirOut, []byte{0x80, 7}))
	assert.Equal(t, errNoClient, p.send(Frame{Kind: FrameOutput}))
}

// answer replies to the next control request with data and returns the request.
func answer(t *testing.T, client net.Conn, status byte, data []byte) Frame {
	t.Helper()
	f := read(t, client)
	require.Equal(t, FrameControl, f.Kind)
	reply := binary.LittleEndian.AppendUint16(nil, binary.LittleEndian.Uint16(f.Payload))
	reply = append(reply, status)
	reply = append(reply, data...)
	write(t, client, Frame{Kind: FrameControlReply, Interface: f.Interface, Payload: reply})
	return f
}

func TestGetReportIsAnsweredByTheClient(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, _ := attach(t, p)
	go answer(t, client, 0, []byte{1, 0x83, 9, 9, 9})
	resp, handled := p.HandleControl(0xA1, 0x01, 0x0301, 2, 4, nil)
	assert.True(t, handled)
	assert.Equal(t, []byte{1, 0x83, 9, 9}, resp, "cut to wLength")
}

func TestTheControlFrameCarriesTheSetupPacketAndData(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, _ := attach(t, p)
	got := make(chan Frame, 1)
	go func() { got <- answer(t, client, 0, []byte{0xEE}) }()
	resp, handled := p.HandleControl(0x21, 0x09, 0x0301, 4, 3, []byte{1, 0x87, 5})
	assert.True(t, handled)
	assert.Nil(t, resp, "host to device: no data stage back")
	f := <-got
	assert.Equal(t, uint8(4), f.Interface)
	assert.Equal(t, []byte{0x21, 0x09, 0x01, 0x03, 4, 0, 3, 0, 1, 0x87, 5}, f.Payload[2:])
}

func TestARefusedRequestAnswersWithoutData(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, _ := attach(t, p)
	go answer(t, client, 1, []byte{1, 2})
	resp, handled := p.HandleControl(0xA1, 0x01, 0x0301, 2, 64, nil)
	assert.True(t, handled)
	assert.Nil(t, resp)
}

func TestVendorAndUnknownRequestsGoToTheClient(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, _ := attach(t, p)
	for _, setup := range [][2]uint8{{0x40, 0x01}, {0xC0, 0x02}, {0x41, 0x05}, {0x21, 0x42}, {0x22, 0x01}} {
		got := make(chan Frame, 1)
		go func() { got <- answer(t, client, 0, nil) }()
		_, handled := p.HandleControl(setup[0], setup[1], 0, 2, 0, nil)
		assert.True(t, handled)
		f := <-got
		assert.Equal(t, setup[0], f.Payload[2], "bmRequestType 0x%02X", setup[0])
	}
}

func TestOtherCdcRequestsGoToTheClient(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, _ := attach(t, p)

	// SEND_BREAK (0x23) and an encapsulated command (0x00) to the CDC interfaces: not line housekeeping, so the
	// client decides (it may take them as a sign of a firmware update).
	for _, setup := range [][2]uint8{{0x23, 0}, {0x00, 1}} {
		got := make(chan Frame, 1)
		go func() { got <- answer(t, client, 0, nil) }()
		_, handled := p.HandleControl(0x21, setup[0], 0xFFFF, uint16(setup[1]), 0, nil)
		assert.True(t, handled)
		select {
		case f := <-got:
			assert.Equal(t, setup[0], f.Payload[3], "bRequest 0x%02X", setup[0])
		case <-time.After(2 * time.Second):
			t.Fatalf("bRequest 0x%02X never reached the client", setup[0])
		}
	}
}

func TestHidAndCdcHousekeepingNeverReachesTheClient(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, _ := attach(t, p)

	for _, request := range []uint8{hidSetIdle, hidGetIdle, hidSetProtocol, hidGetProtocol} {
		_, handled := p.HandleControl(0x21|request&0x80, request, 0, 2, 0, nil)
		assert.False(t, handled, "HID request 0x%02X: VIIPER's default answer", request)
	}

	_, handled := p.HandleControl(0x21, cdcSetLineCoding, 0, 0, 7, []byte{0x80, 0x25, 0, 0, 0, 0, 8})
	assert.True(t, handled)
	resp, _ := p.HandleControl(0xA1, cdcGetLineCoding, 0, 0, 7, nil)
	assert.Equal(t, []byte{0x80, 0x25, 0, 0, 0, 0, 8}, resp)
	resp, _ = p.HandleControl(0xA1, cdcGetLineCoding, 0, 1, 7, nil)
	assert.Equal(t, defaultLineCoding, resp, "interface 1 was never set")
	_, handled = p.HandleControl(0x21, cdcSetControlLineState, 3, 0, 0, nil)
	assert.True(t, handled)

	_, handled = p.HandleControl(0x00, 0x09, 1, 0, 0, nil)
	assert.False(t, handled, "standard requests stay VIIPER's")

	_ = client.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	_, err = ReadFrame(client)
	var timeout net.Error
	assert.True(t, errors.As(err, &timeout) && timeout.Timeout(), "nothing was sent to the client: %v", err)
}

func TestWithoutAClientAControlRequestTimesOut(t *testing.T) {
	old := ControlTimeout
	ControlTimeout = 30 * time.Millisecond
	defer func() { ControlTimeout = old }()
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	started := time.Now()
	resp, handled := p.HandleControl(0xA1, 0x01, 0x0301, 2, 64, nil)
	assert.True(t, handled)
	assert.Nil(t, resp)
	assert.Less(t, time.Since(started), time.Second)
}

func TestAClientAttachingLateStillGetsTheRequest(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	result := make(chan []byte, 1)
	go func() {
		resp, _ := p.HandleControl(0xA1, 0x01, 0x0301, 2, 64, nil)
		result <- resp
	}()
	time.Sleep(20 * time.Millisecond)
	client, _ := attach(t, p)
	answer(t, client, 0, []byte{1, 2, 3})
	assert.Equal(t, []byte{1, 2, 3}, <-result)
}

func TestAClientThatNeverAnswersTimesOut(t *testing.T) {
	old := ControlTimeout
	ControlTimeout = 30 * time.Millisecond
	defer func() { ControlTimeout = old }()
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, _ := attach(t, p)
	go func() { _, _ = ReadFrame(client) }()
	resp, _ := p.HandleControl(0xA1, 0x01, 0x0301, 2, 64, nil)
	assert.Nil(t, resp)
}

func TestALateOrForeignReplyIsIgnored(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, _ := attach(t, p)
	write(t, client, Frame{Kind: FrameControlReply, Payload: []byte{0x34, 0x12, 0, 9}})
	go answer(t, client, 0, []byte{7})
	resp, _ := p.HandleControl(0xA1, 0x01, 0x0301, 2, 64, nil)
	assert.Equal(t, []byte{7}, resp)
}

func TestADisconnectEndsAWaitingRequest(t *testing.T) {
	p, err := New(puckOptions(t))
	require.NoError(t, err)
	client, done := attach(t, p)
	go func() {
		_, _ = ReadFrame(client)
		_ = client.Close()
	}()
	started := time.Now()
	resp, _ := p.HandleControl(0xA1, 0x01, 0x0301, 2, 64, nil)
	assert.Nil(t, resp)
	assert.Less(t, time.Since(started), ControlTimeout)
	assert.ErrorIs(t, <-done, io.EOF)
}

func TestOnlyOneClientAtATime(t *testing.T) {
	p, err := New(smallOptions())
	require.NoError(t, err)
	attach(t, p)
	server, _ := net.Pipe()
	assert.ErrorIs(t, p.Serve(server), errAlreadyAttached)
}

func TestAClientCanAttachAgainAfterADisconnect(t *testing.T) {
	p, err := New(smallOptions())
	require.NoError(t, err)
	client, done := attach(t, p)
	_ = client.Close()
	<-done
	client, _ = attach(t, p)
	write(t, client, Frame{Kind: FrameInput, Interface: 0, Payload: []byte{1, 1}})
	assert.Equal(t, []byte{1, 1}, in(p, 1, time.Second))
}

func TestBadFramesEndTheStream(t *testing.T) {
	p, err := New(smallOptions())
	require.NoError(t, err)
	client, done := attach(t, p)
	write(t, client, Frame{Kind: 0x7F})
	assert.Error(t, <-done)

	client, done = attach(t, p)
	write(t, client, Frame{Kind: FrameControlReply, Payload: []byte{1}})
	assert.Error(t, <-done, "a reply too short")

	client, done = attach(t, p)
	_, _ = client.Write([]byte{FrameInput, 0, 0xFF, 0xFF})
	assert.Error(t, <-done, "a payload too long")
}

func TestFramesRoundTrip(t *testing.T) {
	buffer, err := AppendFrame(nil, Frame{Kind: FrameOutput, Interface: 5, Payload: []byte{1, 2, 3}})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x81, 5, 3, 0, 1, 2, 3}, buffer)
	_, err = AppendFrame(nil, Frame{Payload: make([]byte, MaxPayload+1)})
	assert.Error(t, err)
	assert.Equal(t, uint8(noInterface), controlFrame(1, 0x40, 1, 0, 2, 0, nil).Interface, "a request to the device names no interface")
}

func TestTheCopyAsksNotToRepeatInput(t *testing.T) {
	p, err := New(smallOptions())
	require.NoError(t, err)
	var device usb.Device = p
	repeat, ok := device.(usb.NoInRepeatDevice)
	assert.True(t, ok && repeat.NoInRepeat())
}
