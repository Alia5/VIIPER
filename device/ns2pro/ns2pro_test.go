package ns2pro

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
	"unicode/utf16"

	viiperTesting "github.com/Alia5/VIIPER/_testing"
	"github.com/Alia5/VIIPER/internal/server/api"
	apihandler "github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usbip"
	"github.com/Alia5/VIIPER/viiperclient"
	"github.com/Alia5/VIIPER/virtualbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildReport05(t *testing.T) {
	dev, err := New(nil)
	require.NoError(t, err)

	state := InputState{
		Buttons:       ButtonA | ButtonB | ButtonR | ButtonZR | ButtonMinus | ButtonPlus | ButtonLeftStick | ButtonRightStick | ButtonHome | ButtonCapture | ButtonC | ButtonDown | ButtonRight | ButtonLeft | ButtonUp | ButtonL | ButtonZL | ButtonGR | ButtonGL | ButtonHeadset,
		LX:            0x0123,
		LY:            0x0456,
		RX:            0x0789,
		RY:            0x0ABC,
		AccelX:        0x1122,
		AccelY:        -0x1234,
		AccelZ:        0x3344,
		GyroX:         -0x0102,
		GyroY:         0x5566,
		GyroZ:         -0x0777,
		BatteryLevel:  7,
		Charging:      true,
		ExternalPower: true,
	}
	dev.UpdateInputState(state)

	dev.HandleTransfer(2, usbip.DirOut, featureCommand(0x02, FeatureButtons|FeatureSticks|FeatureIMU|FeatureRumble))
	dev.HandleTransfer(2, usbip.DirOut, featureCommand(0x04, FeatureButtons|FeatureSticks|FeatureIMU|FeatureRumble))
	dev.HandleTransfer(2, usbip.DirOut, selectReportCommand(ReportIDCommon))

	report := dev.HandleTransfer(1, usbip.DirIn, nil)
	require.Len(t, report, InputReportSize)
	assert.Equal(t, byte(ReportIDCommon), report[0])
	assert.Equal(t, uint32(1), binary.LittleEndian.Uint32(report[1:5]))
	assert.Equal(t, []byte{0xCC, 0x7F, 0xCF, 0x13}, report[5:9])

	var leftStick [3]byte
	var rightStick [3]byte
	packStick12(leftStick[:], state.LX, state.LY)
	packStick12(rightStick[:], state.RX, state.RY)
	assert.Equal(t, leftStick[:], report[11:14])
	assert.Equal(t, rightStick[:], report[14:17])
	assert.Equal(t, BatteryVolts, binary.LittleEndian.Uint16(report[0x20:0x22]))
	assert.Equal(t, byte(0x34), report[0x22])
	assert.Equal(t, byte(0x01), report[0x2A])
	assert.Equal(t, uint32(4000), binary.LittleEndian.Uint32(report[0x2B:0x2F]))
	assert.Equal(t, uint16(state.AccelX), binary.LittleEndian.Uint16(report[0x31:0x33]))
	assert.Equal(t, uint16(state.AccelY), binary.LittleEndian.Uint16(report[0x33:0x35]))
	assert.Equal(t, uint16(state.AccelZ), binary.LittleEndian.Uint16(report[0x35:0x37]))
	assert.Equal(t, uint16(state.GyroX), binary.LittleEndian.Uint16(report[0x37:0x39]))
	assert.Equal(t, uint16(state.GyroY), binary.LittleEndian.Uint16(report[0x39:0x3B]))
	assert.Equal(t, uint16(state.GyroZ), binary.LittleEndian.Uint16(report[0x3B:0x3D]))

	next := dev.HandleTransfer(1, usbip.DirIn, nil)
	require.Len(t, next, InputReportSize)
	assert.Equal(t, uint32(2), binary.LittleEndian.Uint32(next[1:5]))
	assert.Equal(t, uint32(8000), binary.LittleEndian.Uint32(next[0x2B:0x2F]))
}

func TestHIDReportDescriptorMatchesCapture(t *testing.T) {
	got, err := reportDescriptor.Bytes()
	require.NoError(t, err)
	assert.Equal(t,
		mustHexBytes(t, "05010905a101850505ff0901150026ff00953f750881028509090195028102050919012915250195157501810295017503810305010901a100093009310933093526ff0f9504750c8102c005ff090226ff0095347508810285020901953f9102c0"),
		[]byte(got),
	)
}

func TestDescriptor(t *testing.T) {
	desc := MakeDescriptor()
	assert.Equal(t, uint16(DefaultVID), desc.Device.IDVendor)
	assert.Equal(t, uint16(DefaultPID), desc.Device.IDProduct)
	assert.Equal(t, uint16(0x0200), desc.Device.BcdDevice)
	assert.Equal(t, "Switch 2 Pro Controller", desc.Strings[2])
	assert.Equal(t, DefaultSerial, desc.Strings[3])
	assert.Equal(t, "Nintendo Switch 2 Pro Controller", desc.Strings[4])
	assert.Equal(t, "Nintendo Switch 2 Pro Controller", desc.Strings[5])
	assert.Equal(t, "Pro Controller", desc.Strings[6])
	assert.Equal(t, byte(0x04), desc.Configuration.IConfiguration)
	assert.Equal(t, byte(0xC0), desc.Configuration.BMAttributes)
	assert.Equal(t, byte(0xFA), desc.Configuration.BMaxPower)
	require.NotNil(t, desc.MicrosoftOS10)
	assert.Equal(t, byte(microsoftOS10VendorCode), desc.MicrosoftOS10.VendorCode)
	assert.Equal(t, uint8(2), desc.NumInterfaces())
	require.Empty(t, desc.Associations)
	require.Len(t, desc.Interfaces, 2)

	hidIface := desc.Interfaces[0]
	assert.Equal(t, byte(0x03), hidIface.Descriptor.BInterfaceClass)
	require.NotNil(t, hidIface.HID)
	require.Len(t, hidIface.Endpoints, 2)
	assert.Equal(t, byte(EndpointHIDIn), hidIface.Endpoints[0].BEndpointAddress)
	assert.Equal(t, byte(EndpointHIDOut), hidIface.Endpoints[1].BEndpointAddress)
	report, err := hidIface.HID.ReportBytes()
	require.NoError(t, err)
	assert.Len(t, report, 97)

	bulkIface := desc.Interfaces[1]
	assert.Equal(t, byte(0xFF), bulkIface.Descriptor.BInterfaceClass)
	require.Len(t, bulkIface.Endpoints, 2)
	assert.Equal(t, byte(EndpointBulkOut), bulkIface.Endpoints[0].BEndpointAddress)
	assert.Equal(t, byte(EndpointBulkIn), bulkIface.Endpoints[1].BEndpointAddress)
}

func TestBuildReport09(t *testing.T) {
	dev, err := New(nil)
	require.NoError(t, err)

	state := InputState{
		Buttons:       ButtonB | ButtonA | ButtonX | ButtonZR | ButtonPlus | ButtonRightStick | ButtonDown | ButtonUp | ButtonL | ButtonZL | ButtonMinus | ButtonLeftStick | ButtonHome | ButtonCapture | ButtonGR | ButtonGL | ButtonC,
		LX:            0x0001,
		LY:            0x0002,
		RX:            0x0FFE,
		RY:            0x0FFF,
		BatteryLevel:  5,
		ExternalPower: true,
	}
	dev.UpdateInputState(state)
	dev.HandleTransfer(2, usbip.DirOut, selectReportCommand(ReportIDPro))
	assert.NotEmpty(t, dev.HandleTransfer(2, usbip.DirIn, nil))

	report := dev.HandleTransfer(1, usbip.DirIn, nil)
	require.Len(t, report, InputReportSize)
	assert.Equal(t, byte(ReportIDPro), report[0])
	assert.Equal(t, byte(1), report[1])
	assert.Equal(t, byte(0x15), report[2])
	assert.Equal(t, []byte{0xEB, 0xF9, 0x1F}, report[3:6])
	assert.Equal(t, byte(0x30), report[12])

	var leftStick [3]byte
	var rightStick [3]byte
	packStick12(leftStick[:], state.LX, state.LY)
	packStick12(rightStick[:], state.RX, state.RY)
	assert.Equal(t, leftStick[:], report[6:9])
	assert.Equal(t, rightStick[:], report[9:12])
}

func TestBulkCommands(t *testing.T) {
	dev, err := New(nil)
	require.NoError(t, err)

	dev.HandleTransfer(2, usbip.DirOut, featureCommand(0x02, FeatureButtons|FeatureSticks|FeatureIMU|FeatureRumble))
	assert.NotEmpty(t, dev.HandleTransfer(2, usbip.DirIn, nil))
	dev.HandleTransfer(2, usbip.DirOut, featureCommand(0x04, FeatureButtons|FeatureSticks|FeatureIMU|FeatureRumble))
	assert.NotEmpty(t, dev.HandleTransfer(2, usbip.DirIn, nil))

	dev.HandleTransfer(2, usbip.DirOut, selectReportCommand(ReportIDCommon))
	assert.NotEmpty(t, dev.HandleTransfer(2, usbip.DirIn, nil))
	report := dev.HandleTransfer(1, usbip.DirIn, nil)
	require.Len(t, report, InputReportSize)
	assert.Equal(t, byte(ReportIDCommon), report[0])

	dev.HandleTransfer(2, usbip.DirOut, flashReadCommand(0x13000))
	resp := dev.HandleTransfer(2, usbip.DirIn, nil)
	require.Len(t, resp, 0x50)
	assert.Equal(t, []byte{0x02, 0x01, 0x01, 0x01}, resp[0:4])
	assert.Equal(t, byte(0x40), resp[8])
	assert.Equal(t, uint32(0x13000), binary.LittleEndian.Uint32(resp[12:16]))
	flash := resp[16:]
	require.Len(t, flash, 64)
	assert.Equal(t, "VIIPER-NS2PRO-00", string(flash[2:18]))
}

func TestSDLUSBInitializationSequence(t *testing.T) {
	dev, err := New(nil)
	require.NoError(t, err)

	for _, address := range []uint32{0x13000, 0x13040, 0x13080, 0x130C0, 0x13100, 0x1FC040, 0x1FC080} {
		dev.HandleTransfer(2, usbip.DirOut, flashReadCommand(address))
		resp := dev.HandleTransfer(2, usbip.DirIn, nil)
		require.Len(t, resp, 0x50, "flash block %#x", address)
		assert.Equal(t, []byte{0x02, 0x01, 0x01, 0x01}, resp[0:4])
		assert.Equal(t, byte(0x40), resp[8])
		assert.Equal(t, address, binary.LittleEndian.Uint32(resp[12:16]))
	}

	initSequence := [][]byte{
		{0x07, 0x91, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
		{0x0C, 0x91, 0x00, 0x02, 0x00, 0x04, 0x00, 0x00, 0x27, 0x00, 0x00, 0x00},
		{0x11, 0x91, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
		{0x0A, 0x91, 0x00, 0x08, 0x00, 0x14, 0x00, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x35, 0x00, 0x46, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		{0x0C, 0x91, 0x00, 0x04, 0x00, 0x04, 0x00, 0x00, 0x27, 0x00, 0x00, 0x00},
		{0x01, 0x91, 0x00, 0x0C, 0x00, 0x00, 0x00, 0x00},
		{0x01, 0x91, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
		{0x08, 0x91, 0x00, 0x02, 0x00, 0x04, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
		{0x03, 0x91, 0x00, 0x0A, 0x00, 0x04, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00},
		{0x03, 0x91, 0x00, 0x0D, 0x00, 0x08, 0x00, 0x00, 0x01, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}
	for _, cmd := range initSequence {
		dev.HandleTransfer(2, usbip.DirOut, cmd)
		assert.NotEmpty(t, dev.HandleTransfer(2, usbip.DirIn, nil), "cmd %#x sub %#x", cmd[0], cmd[3])
	}

	dev.UpdateInputState(InputState{BatteryLevel: 9})
	report := dev.HandleTransfer(1, usbip.DirIn, nil)
	require.Len(t, report, InputReportSize)
	assert.Equal(t, byte(ReportIDCommon), report[0])
	assert.Equal(t, uint32(1), binary.LittleEndian.Uint32(report[1:5]))
	assert.Equal(t, uint32(4000), binary.LittleEndian.Uint32(report[0x2B:0x2F]))
}

func TestRumbleOutput(t *testing.T) {
	dev, err := New(nil)
	require.NoError(t, err)

	var got OutputState
	called := false
	dev.SetOutputCallback(func(out OutputState) {
		got = out
		called = true
	})

	packet := make([]byte, OutputReportSize)
	packet[0] = ReportIDOutput
	for i := 0; i < 16; i++ {
		packet[1+i] = byte(i)
		packet[17+i] = byte(0x80 + i)
	}
	dev.HandleTransfer(1, usbip.DirOut, packet)

	require.True(t, called)
	assert.Equal(t, byte(OutputFlagRumble), got.Flags)
	for i := 0; i < 16; i++ {
		assert.Equal(t, byte(i), got.LeftRumble[i])
		assert.Equal(t, byte(0x80+i), got.RightRumble[i])
	}

	called = false
	payload := make([]byte, OutputRumbleSize)
	for i := 0; i < 16; i++ {
		payload[i] = byte(0x40 + i)
		payload[16+i] = byte(0xC0 + i)
	}
	_, handled := dev.HandleControl(0x21, 0x09, 0x0202, 0, 0, payload)
	require.True(t, handled)
	require.True(t, called)
	assert.Equal(t, byte(OutputFlagRumble), got.Flags)
	assert.Equal(t, byte(0x40), got.LeftRumble[0])
	assert.Equal(t, byte(0xC0), got.RightRumble[0])
}

func TestPlayerLEDOutput(t *testing.T) {
	dev, err := New(nil)
	require.NoError(t, err)

	var got OutputState
	called := false
	dev.SetOutputCallback(func(out OutputState) {
		got = out
		called = true
	})

	dev.HandleTransfer(2, usbip.DirOut, []byte{0x09, 0x91, 0x12, 0x07, 0x00, 0x08, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00})
	resp := dev.HandleTransfer(2, usbip.DirIn, nil)

	require.True(t, called)
	assert.Equal(t, byte(OutputFlagLED), got.Flags)
	assert.Equal(t, byte(0x06), got.PlayerLedMask)
	assert.Equal(t, []byte{0x09, 0x01, 0x12, 0x07}, resp[:4])
}

func TestMicrosoftOS10WinUSBDescriptor(t *testing.T) {
	desc := MakeDescriptor()
	require.NotNil(t, desc.MicrosoftOS10)

	resp, handled := desc.MicrosoftOS10.ControlResponse(0, 0x0004)
	require.True(t, handled)
	require.Len(t, resp, 40)
	assert.Equal(t, uint32(40), binary.LittleEndian.Uint32(resp[0:4]))
	assert.Equal(t, uint16(0x0100), binary.LittleEndian.Uint16(resp[4:6]))
	assert.Equal(t, uint16(0x0004), binary.LittleEndian.Uint16(resp[6:8]))
	assert.Equal(t, byte(0x01), resp[8])
	assert.Equal(t, byte(0x01), resp[16])
	assert.Equal(t, []byte("WINUSB\x00\x00"), resp[18:26])
}

func TestMicrosoftOS10ExtendedPropertiesDescriptor(t *testing.T) {
	desc := MakeDescriptor()
	require.NotNil(t, desc.MicrosoftOS10)

	resp, handled := desc.MicrosoftOS10.ControlResponse(0, 0x0005)
	require.True(t, handled)
	require.Len(t, resp, 142)
	assert.Equal(t, uint32(142), binary.LittleEndian.Uint32(resp[0:4]))
	assert.Equal(t, uint16(0x0100), binary.LittleEndian.Uint16(resp[4:6]))
	assert.Equal(t, uint16(0x0005), binary.LittleEndian.Uint16(resp[6:8]))
	assert.Equal(t, uint16(1), binary.LittleEndian.Uint16(resp[8:10]))
	assert.Equal(t, uint32(132), binary.LittleEndian.Uint32(resp[10:14]))
	assert.Equal(t, uint32(1), binary.LittleEndian.Uint32(resp[14:18]))
	assert.Equal(t, uint16(40), binary.LittleEndian.Uint16(resp[18:20]))
	assert.Equal(t, uint32(78), binary.LittleEndian.Uint32(resp[60:64]))
	assert.Contains(t, utf16leToString(resp[20:60]), "DeviceInterfaceGUID")
	assert.Contains(t, utf16leToString(resp[64:]), microsoftOS10DeviceInterfaceGUID)
}

func TestStreamInputAndRumble(t *testing.T) {
	s := viiperTesting.NewTestServer(t)
	defer s.UsbServer.Close()
	defer s.ApiServer.Close()

	r := s.ApiServer.Router()
	r.Register("bus/{id}/add", apihandler.BusDeviceAdd(s.UsbServer, s.ApiServer))
	r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(s.UsbServer))

	require.NoError(t, s.ApiServer.Start())

	b, err := virtualbus.NewWithBusID(1)
	require.NoError(t, err)
	defer b.Close()
	require.NoError(t, s.UsbServer.AddBus(b))

	client := viiperclient.New(s.ApiServer.Addr())
	stream, _, err := client.AddDeviceAndConnect(context.Background(), b.BusID(), "ns2pro", nil)
	require.NoError(t, err)
	defer stream.Close()

	usbipClient := viiperTesting.NewUsbIpClient(t, s.UsbServer.Addr())
	devs, err := usbipClient.ListDevices()
	require.NoError(t, err)
	require.Len(t, devs, 1)
	imp, err := usbipClient.AttachDevice(devs[0].BusID)
	require.NoError(t, err)
	defer imp.Conn.Close()

	productString, err := controlIn(imp.Conn, controlSetup(0x0302, 0, 64))
	require.NoError(t, err)
	assert.Equal(t, usb.EncodeStringDescriptor("Switch 2 Pro Controller"), productString)

	serialString, err := controlIn(imp.Conn, controlSetup(0x0303, 0, 64))
	require.NoError(t, err)
	assert.Equal(t, usb.EncodeStringDescriptor(DefaultSerial), serialString)

	msOSString, err := controlIn(imp.Conn, controlSetup(0x03EE, 0, 18))
	require.NoError(t, err)
	assert.Equal(t, []byte{
		0x12, 0x03,
		'M', 0x00,
		'S', 0x00,
		'F', 0x00,
		'T', 0x00,
		'1', 0x00,
		'0', 0x00,
		'0', 0x00,
		microsoftOS10VendorCode, 0x00,
	}, msOSString)

	config, err := controlIn(imp.Conn, controlSetup(0x0200, 0, 512))
	require.NoError(t, err)
	require.Len(t, config, 64)
	assert.Equal(t, []byte{0x09, 0x02, 0x40, 0x00, 0x02, 0x01, 0x04, 0xC0, 0xFA}, config[:9])
	assert.Equal(t, []byte{0x09, 0x04, 0x00, 0x00, 0x02, 0x03, 0x00, 0x00, 0x05}, config[9:18])
	assert.Equal(t, []byte{0x09, 0x04, 0x01, 0x00, 0x02, 0xFF, 0x00, 0x00, 0x06}, config[41:50])

	require.NoError(t, usbipClient.Submit(imp.Conn, usbip.DirOut, 2, selectReportCommand(ReportIDPro), nil))

	state := InputState{
		Buttons:       ButtonA | ButtonHome | ButtonRight,
		LX:            0x0123,
		LY:            0x0456,
		RX:            0x0789,
		RY:            0x0ABC,
		BatteryLevel:  6,
		ExternalPower: true,
	}
	require.NoError(t, stream.WriteBinary(&state))

	expected := state.buildProReport(0, FeatureButtons|FeatureSticks)
	got := pollInputIgnoringCounter(t, usbipClient, imp.Conn, expected, 750*time.Millisecond)
	require.Len(t, got, InputReportSize)
	got[1] = 0
	assert.Equal(t, expected, got)

	rumble := make([]byte, OutputReportSize)
	rumble[0] = ReportIDOutput
	for i := 0; i < 16; i++ {
		rumble[1+i] = byte(0x10 + i)
		rumble[17+i] = byte(0x90 + i)
	}
	require.NoError(t, usbipClient.Submit(imp.Conn, usbip.DirOut, 1, rumble, nil))

	var outBuf [OutputWireSize]byte
	_ = stream.SetReadDeadline(time.Now().Add(750 * time.Millisecond))
	_, err = io.ReadFull(stream, outBuf[:])
	require.NoError(t, err)
	var out OutputState
	require.NoError(t, out.UnmarshalBinary(outBuf[:]))
	assert.Equal(t, byte(OutputFlagRumble), out.Flags)
	assert.Equal(t, byte(0x10), out.LeftRumble[0])
	assert.Equal(t, byte(0x90), out.RightRumble[0])
}

func featureCommand(sub, flags uint8) []byte {
	return []byte{0x0C, 0x91, 0x00, sub, 0x00, 0x04, 0x00, 0x00, flags, 0x00, 0x00, 0x00}
}

func selectReportCommand(reportID uint8) []byte {
	return []byte{0x03, 0x91, 0x00, 0x0A, 0x00, 0x04, 0x00, 0x00, reportID, 0x00, 0x00, 0x00}
}

func flashReadCommand(address uint32) []byte {
	cmd := []byte{0x02, 0x91, 0x01, 0x01, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	binary.LittleEndian.PutUint32(cmd[12:16], address)
	return cmd
}

func mustHexBytes(t *testing.T, s string) []byte {
	t.Helper()
	out, err := hex.DecodeString(s)
	require.NoError(t, err)
	return out
}

func utf16leToString(b []byte) string {
	units := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		u := binary.LittleEndian.Uint16(b[i : i+2])
		if u == 0 {
			break
		}
		units = append(units, u)
	}
	return string(utf16.Decode(units))
}

func controlSetup(wValue, wIndex, wLength uint16) [8]byte {
	var setup [8]byte
	setup[0] = 0x80 // bmRequestType
	setup[1] = 0x06 // bRequest
	binary.LittleEndian.PutUint16(setup[2:4], wValue)
	binary.LittleEndian.PutUint16(setup[4:6], wIndex)
	binary.LittleEndian.PutUint16(setup[6:8], wLength)
	return setup
}

func controlIn(conn net.Conn, setup [8]byte) ([]byte, error) {
	cmd := usbip.CmdSubmit{
		Basic:             usbip.HeaderBasic{Command: usbip.CmdSubmitCode, Seqnum: 0xC001, Devid: 0, Dir: usbip.DirIn, Ep: 0},
		TransferBufferLen: uint32(binary.LittleEndian.Uint16(setup[6:8])),
		Setup:             setup,
	}
	_ = conn.SetDeadline(time.Now().Add(750 * time.Millisecond))
	defer conn.SetDeadline(time.Time{})
	if err := cmd.Write(conn); err != nil {
		return nil, err
	}

	var retHdr [48]byte
	if err := usbip.ReadExactly(conn, retHdr[:]); err != nil {
		return nil, err
	}
	if gotCmd := binary.BigEndian.Uint32(retHdr[0:4]); gotCmd != usbip.RetSubmitCode {
		return nil, fmt.Errorf("unexpected ret cmd %x", gotCmd)
	}
	if status := int32(binary.BigEndian.Uint32(retHdr[20:24])); status != 0 {
		return nil, fmt.Errorf("ret status %d", status)
	}
	actual := binary.BigEndian.Uint32(retHdr[24:28])
	data := make([]byte, int(actual))
	if actual > 0 {
		if err := usbip.ReadExactly(conn, data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func pollInputIgnoringCounter(t *testing.T, client *viiperTesting.TestUsbIpClient, conn net.Conn, want []byte, timeout time.Duration) []byte {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		got, err := client.ReadInputReport(conn)
		require.NoError(t, err)
		if len(got) == len(want) {
			g := append([]byte(nil), got...)
			w := append([]byte(nil), want...)
			g[1] = 0
			w[1] = 0
			if assert.ObjectsAreEqual(w, g) {
				return got
			}
		}
		if time.Now().After(deadline) {
			return got
		}
		time.Sleep(1 * time.Millisecond)
	}
}
