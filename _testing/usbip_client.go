package testing

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Alia5/VIIPER/usbip"
)

type TestUsbIpClient struct {
	address string
	seq     uint32
}

type Device struct {
	Path       string
	BusID      string
	BusNum     uint32
	DeviceNum  uint32
	Speed      uint32
	IDVendor   uint16
	IDProduct  uint16
	BcdDevice  uint16
	Class      uint8
	SubClass   uint8
	Protocol   uint8
	ConfigVal  uint8
	NumConfigs uint8
	NumIfaces  uint8
	Interfaces []usbip.InterfaceDesc
}

type ImportResult struct {
	Conn          net.Conn
	Exported      Device
	RawDescriptor []byte
}

func NewUsbIpClient(t *testing.T, addr string) *TestUsbIpClient {
	t.Helper()

	return &TestUsbIpClient{
		address: addr,
	}
}

func (c *TestUsbIpClient) nextSeq() uint32 {
	// USBIP seqnum only needs to be unique within the session; tests use a single
	// client per test and the server doesn't require a specific starting value.
	return atomic.AddUint32(&c.seq, 1) - 1
}

func (c *TestUsbIpClient) ListDevices() ([]Device, error) {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := (&usbip.MgmtHeader{Version: usbip.Version, Command: usbip.OpReqDevlist}).Write(conn); err != nil {
		return nil, err
	}

	var hdr [12]byte
	if err := usbip.ReadExactly(conn, hdr[:]); err != nil {
		return nil, err
	}

	if v := binary.BigEndian.Uint16(hdr[0:2]); v != usbip.Version {
		return nil, fmt.Errorf("unexpected usbip version %x", v)
	}
	if cmd := binary.BigEndian.Uint16(hdr[2:4]); cmd != usbip.OpRepDevlist {
		return nil, fmt.Errorf("unexpected reply command %x", cmd)
	}

	n := binary.BigEndian.Uint32(hdr[8:12])
	devices := make([]Device, 0, n)
	for i := uint32(0); i < n; i++ {
		dev, err := readExportedDevice(conn)
		if err != nil {
			return nil, err
		}
		devices = append(devices, dev)
	}

	return devices, nil
}

func (c *TestUsbIpClient) AttachDevice(busID string) (*ImportResult, error) {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return nil, err
	}

	if err := (&usbip.MgmtHeader{Version: usbip.Version, Command: usbip.OpReqImport}).Write(conn); err != nil {
		conn.Close()
		return nil, err
	}

	var bus [32]byte
	copy(bus[:], busID)
	if _, err := conn.Write(bus[:]); err != nil {
		conn.Close()
		return nil, err
	}

	var hdr [8]byte
	if err := usbip.ReadExactly(conn, hdr[:]); err != nil {
		conn.Close()
		return nil, err
	}
	if v := binary.BigEndian.Uint16(hdr[0:2]); v != usbip.Version {
		conn.Close()
		return nil, fmt.Errorf("unexpected usbip version %x", v)
	}
	if cmd := binary.BigEndian.Uint16(hdr[2:4]); cmd != usbip.OpRepImport {
		conn.Close()
		return nil, fmt.Errorf("unexpected reply command %x", cmd)
	}

	dev, raw, err := readExportedDeviceImportWithRaw(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &ImportResult{Conn: conn, Exported: dev, RawDescriptor: raw}, nil
}

func readExportedDevice(r net.Conn) (Device, error) {
	dev, _, err := readExportedDeviceWithRaw(r)
	return dev, err
}

func readExportedDeviceImportWithRaw(r net.Conn) (Device, []byte, error) {
	return readExportedDeviceWithRawInternal(r, false)
}

func readExportedDeviceWithRaw(r net.Conn) (Device, []byte, error) {
	return readExportedDeviceWithRawInternal(r, true)
}

func readExportedDeviceWithRawInternal(r net.Conn, readIfaces bool) (Device, []byte, error) {
	var base [312]byte
	if err := usbip.ReadExactly(r, base[:]); err != nil {
		return Device{}, nil, err
	}

	pathField := base[0:256]
	busField := base[256:288]

	pathEnd := bytes.IndexByte(pathField, 0)
	if pathEnd == -1 {
		pathEnd = len(pathField)
	}
	busEnd := bytes.IndexByte(busField, 0)
	if busEnd == -1 {
		busEnd = len(busField)
	}

	busNum := binary.BigEndian.Uint32(base[288:292])
	devNum := binary.BigEndian.Uint32(base[292:296])
	speed := binary.BigEndian.Uint32(base[296:300])
	idVendor := binary.BigEndian.Uint16(base[300:302])
	idProduct := binary.BigEndian.Uint16(base[302:304])
	bcdDevice := binary.BigEndian.Uint16(base[304:306])
	class := base[306]
	subClass := base[307]
	proto := base[308]
	confVal := base[309]
	nConf := base[310]
	nIf := base[311]

	ifaces := make([]usbip.InterfaceDesc, 0, nIf)
	if readIfaces && nIf > 0 {
		ifaceBuf := make([]byte, int(nIf)*4)
		if err := usbip.ReadExactly(r, ifaceBuf); err != nil {
			return Device{}, nil, err
		}
		for i := 0; i < int(nIf); i++ {
			o := i * 4
			ifaces = append(ifaces, usbip.InterfaceDesc{
				Class:    ifaceBuf[o],
				SubClass: ifaceBuf[o+1],
				Protocol: ifaceBuf[o+2],
			})
		}
	}

	return Device{
		Path:       string(pathField[:pathEnd]),
		BusID:      string(busField[:busEnd]),
		BusNum:     busNum,
		DeviceNum:  devNum,
		Speed:      speed,
		IDVendor:   idVendor,
		IDProduct:  idProduct,
		BcdDevice:  bcdDevice,
		Class:      class,
		SubClass:   subClass,
		Protocol:   proto,
		ConfigVal:  confVal,
		NumConfigs: nConf,
		NumIfaces:  nIf,
		Interfaces: ifaces,
	}, base[:], nil
}

func (c *TestUsbIpClient) Submit(conn net.Conn, dir uint32, ep uint32, outPayload []byte, setup *[8]byte) error {
	return c.SubmitWithTimeout(conn, dir, ep, outPayload, setup, 750*time.Millisecond)
}

func (c *TestUsbIpClient) SubmitWithTimeout(conn net.Conn, dir uint32, ep uint32, outPayload []byte, setup *[8]byte, timeout time.Duration) error {
	if conn == nil {
		return io.ErrUnexpectedEOF
	}

	var setupBytes [8]byte
	if setup != nil {
		setupBytes = *setup
	}

	cur := c.nextSeq()

	cmd := usbip.CmdSubmit{
		Basic:             usbip.HeaderBasic{Command: usbip.CmdSubmitCode, Seqnum: cur, Devid: 0, Dir: dir, Ep: ep},
		TransferFlags:     0,
		TransferBufferLen: uint32(len(outPayload)),
		StartFrame:        0,
		NumberOfPackets:   0,
		Interval:          0,
		Setup:             setupBytes,
	}

	_ = conn.SetDeadline(time.Now().Add(timeout))
	if err := cmd.Write(conn); err != nil {
		return err
	}
	if len(outPayload) > 0 {
		if _, err := conn.Write(outPayload); err != nil {
			return err
		}
	}

	var retHdr [48]byte
	if err := usbip.ReadExactly(conn, retHdr[:]); err != nil {
		return err
	}
	if gotCmd := binary.BigEndian.Uint32(retHdr[0:4]); gotCmd != usbip.RetSubmitCode {
		return fmt.Errorf("unexpected ret cmd %x", gotCmd)
	}
	status := int32(binary.BigEndian.Uint32(retHdr[20:24]))
	actual := binary.BigEndian.Uint32(retHdr[24:28])
	if status != 0 {
		return fmt.Errorf("ret status %d", status)
	}

	if dir == usbip.DirIn && actual > 0 {
		discard := make([]byte, int(actual))
		if err := usbip.ReadExactly(conn, discard); err != nil {
			return err
		}
	}
	_ = conn.SetDeadline(time.Time{})
	return nil
}

func (c *TestUsbIpClient) ReadInputReport(conn net.Conn) ([]byte, error) {
	return c.ReadInputReportWithTimeout(conn, 250*time.Millisecond)
}

func (c *TestUsbIpClient) ReadInputReportWithTimeout(conn net.Conn, timeout time.Duration) ([]byte, error) {
	if conn == nil {
		return nil, io.ErrUnexpectedEOF
	}
	cur := c.nextSeq()

	// Request a buffer large enough for all current VIIPER HID devices.
	// (Keyboard reports are 34 bytes; mouse/xbox360 are smaller.)
	const inMax = 255

	cmd := usbip.CmdSubmit{
		Basic:             usbip.HeaderBasic{Command: usbip.CmdSubmitCode, Seqnum: cur, Devid: 0, Dir: usbip.DirIn, Ep: 1},
		TransferFlags:     0,
		TransferBufferLen: inMax,
		StartFrame:        0,
		NumberOfPackets:   0,
		Interval:          0,
		Setup:             [8]byte{},
	}
	_ = conn.SetDeadline(time.Now().Add(timeout))
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
	status := int32(binary.BigEndian.Uint32(retHdr[20:24]))
	actual := binary.BigEndian.Uint32(retHdr[24:28])
	if status != 0 {
		return nil, fmt.Errorf("ret status %d", status)
	}
	data := make([]byte, int(actual))
	if actual > 0 {
		if err := usbip.ReadExactly(conn, data); err != nil {
			return nil, err
		}
	}
	_ = conn.SetDeadline(time.Time{})
	return data, nil
}

func (c *TestUsbIpClient) PollInputReport(conn net.Conn, want []byte, timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)
	var last []byte
	for {
		got, err := c.ReadInputReport(conn)
		if err != nil {
			return nil, err
		}
		last = got
		if len(got) == len(want) {
			eq := true
			for i := range want {
				if want[i] != got[i] {
					eq = false
					break
				}
			}
			if eq {
				return got, nil
			}
		}
		if time.Now().After(deadline) {
			return last, nil
		}
		time.Sleep(1 * time.Millisecond)
	}
}
