package usb

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Alia5/VIIPER/internal/log"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usbip"
	"github.com/Alia5/VIIPER/virtualbus"
)

type batchingWriter struct {
	mu           sync.Mutex
	w            *bufio.Writer
	flushEvery   time.Duration
	flushAtBytes int
	stopCh       chan struct{}
	closeOnce    sync.Once
	err          error
}

const (
	retSubmitHeaderSize = 0x30

	// avoid windows socket overhead while keeping latency very low.
	writeBatcherBufferSize   = 256 * 1024
	writeBatcherFlushEvery   = 1 * time.Millisecond
	writeBatcherFlushAtBytes = 64 * 1024
)

func newBatchingWriter(dst io.Writer, bufSize int, flushEvery time.Duration, flushAtBytes int) *batchingWriter {
	if bufSize <= 0 {
		bufSize = writeBatcherBufferSize
	}
	if flushAtBytes < 0 {
		flushAtBytes = 0
	}
	if flushAtBytes > bufSize {
		flushAtBytes = bufSize
	}
	bw := &batchingWriter{
		w:            bufio.NewWriterSize(dst, bufSize),
		flushEvery:   flushEvery,
		flushAtBytes: flushAtBytes,
		stopCh:       make(chan struct{}),
	}
	if flushEvery > 0 {
		go bw.flushLoop()
	}
	return bw
}

func (b *batchingWriter) flushLoop() {
	t := time.NewTicker(b.flushEvery)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			_ = b.Flush()
		case <-b.stopCh:
			return
		}
	}
}

func (b *batchingWriter) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.err != nil {
		return 0, b.err
	}

	n, err := b.w.Write(p)
	if err != nil {
		b.err = err
		return n, err
	}
	if b.flushAtBytes > 0 && b.w.Buffered() >= b.flushAtBytes {
		if err := b.w.Flush(); err != nil {
			b.err = err
			return n, err
		}
	}
	return n, nil
}

func (b *batchingWriter) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.err != nil {
		return b.err
	}
	if err := b.w.Flush(); err != nil {
		b.err = err
		return err
	}
	return nil
}

func (b *batchingWriter) Close() error {
	b.closeOnce.Do(func() {
		close(b.stopCh)
	})
	return b.Flush()
}

const (
	// USB standard request codes
	usbReqGetStatus        = 0x00
	usbReqClearFeature     = 0x01
	usbReqSetFeature       = 0x03
	usbReqSetAddress       = 0x05
	usbReqGetDescriptor    = 0x06
	usbReqSetDescriptor    = 0x07
	usbReqGetConfiguration = 0x08
	usbReqSetConfiguration = 0x09

	// USB descriptor types
	usbDescTypeDevice        = 0x01
	usbDescTypeConfiguration = 0x02
	usbDescTypeString        = 0x03
	usbDescTypeHID           = 0x21
	usbDescTypeHIDReport     = 0x22

	// USB request types (bmRequestType)
	usbReqTypeStandardToDevice    = 0x00
	usbReqTypeStandardToInterface = 0x81
	usbReqTypeStandardFromDevice  = 0x80

	// USB configuration values
	usbConfigValueDefault   = 1
	usbConfigAttrBusPowered = 0x80
	usbConfigMaxPower100mA  = 50 // In units of 2mA

	// URB header field offsets
	urbHdrSize          = 0x30
	urbHdrOffsetCommand = 0x00
	urbHdrOffsetSeqnum  = 0x04
	urbHdrOffsetDevid   = 0x08
	urbHdrOffsetDir     = 0x0c
	urbHdrOffsetEp      = 0x10
	urbHdrOffsetUnlink  = 0x14
	urbHdrOffsetFlags   = 0x14
	urbHdrOffsetLength  = 0x18
	urbHdrOffsetSetup   = 0x28

	// Standard header peek size
	headerPeekSize = 8

	// BUSID buffer size for import
	busIDSize = 32

	// Error codes
	errConnReset = -104 // -ECONNRESET
)

type Server struct {
	config    *ServerConfig
	logger    *slog.Logger
	rawLogger log.RawLogger
	busses    map[uint32]*virtualbus.VirtualBus
	busesMu   sync.Mutex
	ready     chan struct{}
	readyOnce sync.Once
	ln        net.Listener
}

func New(config ServerConfig, logger *slog.Logger, rawLogger log.RawLogger) *Server {
	return &Server{
		config:    &config,
		logger:    logger,
		rawLogger: rawLogger,
		busses:    make(map[uint32]*virtualbus.VirtualBus),
		ready:     make(chan struct{}),
	}
}

// AddBus registers a bus with the server. If the bus number is already present,
// an error is returned.
func (s *Server) AddBus(bus *virtualbus.VirtualBus) error {
	s.busesMu.Lock()
	defer s.busesMu.Unlock()
	if bus == nil {
		return fmt.Errorf("bus is nil")
	}
	if _, ok := s.busses[bus.BusID()]; ok {
		return fmt.Errorf("bus %d already registered", bus.BusID())
	}
	s.busses[bus.BusID()] = bus
	return nil
}

// RemoveBus unregisters a bus from the server.
func (s *Server) RemoveBus(busID uint32) error {
	s.busesMu.Lock()
	bus, ok := s.busses[busID]
	if !ok {
		s.busesMu.Unlock()
		return fmt.Errorf("bus %d not found", busID)
	}

	devices := bus.Devices()
	s.busesMu.Unlock()

	if len(devices) > 0 {
		s.logger.Warn(fmt.Sprintf("Removing non-empty bus %d with %d device(s) attached; removing devices", busID, len(devices)))
		for _, dev := range devices {
			_ = bus.Remove(dev)
		}
	}

	s.busesMu.Lock()
	delete(s.busses, busID)
	s.busesMu.Unlock()

	return bus.Close()
}

// RemoveDeviceByID removes a device by busId and cancels its connections.
func (s *Server) RemoveDeviceByID(busID uint32, deviceID string) error {
	s.busesMu.Lock()
	bus, ok := s.busses[busID]
	s.busesMu.Unlock()

	if !ok {
		return fmt.Errorf("bus %d not found", busID)
	}
	err := bus.RemoveDeviceByID(deviceID)
	if err != nil {
		return err
	}

	if emptyCtx := bus.GetBusEmptyContext(); emptyCtx != nil {
		go func() {
			slog.Debug("Started bus cleanup goroutine (RemoveDeviceByID)")
			select {
			case <-emptyCtx.Done():
				// Cancelled - a new device was added
				return
			case <-time.After(s.config.BusCleanupTimeout):
				if b := s.GetBus(busID); b != nil && len(b.Devices()) == 0 {
					if err := s.RemoveBus(busID); err != nil {
						s.logger.Error("timeout: failed to remove empty bus", "busID", busID, "error", err)
					} else {
						s.logger.Info("timeout: removed empty bus", "busID", busID)
					}
				}
			}
		}()
	} else {
		s.logger.Debug("No bus empty context; Cleaning bus immediately")
		if b := s.GetBus(busID); b != nil && len(b.Devices()) == 0 {
			if err := s.RemoveBus(busID); err != nil {
				s.logger.Error("timeout: failed to remove empty bus", "busID", busID, "error", err)
			} else {
				s.logger.Info("timeout: removed empty bus", "busID", busID)
			}
		}
	}

	return nil
}

// ListBuses returns a snapshot of active bus numbers.
func (s *Server) ListBuses() []uint32 {
	s.busesMu.Lock()
	defer s.busesMu.Unlock()
	out := make([]uint32, 0, len(s.busses))
	for k := range s.busses {
		out = append(out, k)
	}
	return out
}

// GetBus returns a bus by ID or nil if not present.
func (s *Server) GetBus(busID uint32) *virtualbus.VirtualBus {
	s.busesMu.Lock()
	defer s.busesMu.Unlock()
	return s.busses[busID]
}

func (s *Server) NextFreeBusID() uint32 {
	s.busesMu.Lock()
	defer s.busesMu.Unlock()
	var id uint32 = 1
	for {
		if _, exists := s.busses[id]; !exists {
			return id
		}
		id++
	}
}

func (s *Server) Addr() string {
	if s.ln != nil {
		return s.ln.Addr().String()
	}
	if s.config != nil {
		return s.config.Addr
	}
	return ""
}

// ListenAndServe starts the USB-IP server and handles incoming connections.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return err
	}
	s.ln = ln
	s.config.Addr = ln.Addr().String()
	s.readyOnce.Do(func() { close(s.ready) })
	s.logger.Info("USBIP server listening", "addr", s.config.Addr)
	for {
		c, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || strings.Contains(strings.ToLower(err.Error()), "use of closed network connection") {
				s.logger.Info("USBIP server stopped")
				return nil
			}
			s.logger.Error("Accept error", "error", err)
			continue
		}
		if tcpConn, ok := c.(*net.TCPConn); ok {
			if err := tcpConn.SetNoDelay(true); err != nil {
				s.logger.Warn("failed to set TCP_NODELAY", "error", err)
			}
		}
		s.logger.Info("Client connected", "remote", c.RemoteAddr())
		go func() {
			if err := s.handleConn(c); err != nil {
				if isClientDisconnect(err) {
					s.logger.Info("Client disconnected", "error", err)
				} else {
					s.logger.Error("Connection handler error", "error", err)
				}
			}
		}()
	}
}

// Ready returns a channel that is closed once the server has successfully bound
// to its listen address and is ready to accept connections.
func (s *Server) Ready() <-chan struct{} { return s.ready }

// Close stops the USB server by closing its listener.
func (s *Server) Close() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

// GetListenPort extracts and returns the port number from the server's listen address.
func (s *Server) GetListenPort() uint16 {
	addr := s.Addr()
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0
	}
	return uint16(port)
}

// --

func (s *Server) handleConn(conn net.Conn) error {
	defer conn.Close()
	conn = &logConn{Conn: conn, s: s}
	if err := conn.SetDeadline(time.Now().Add(s.config.ConnectionTimeout)); err != nil {
		s.logger.Warn("Failed to set deadline", "error", err)
	}

	// Peek first 8 bytes to determine management op or URB stream.
	var hdrBuf [headerPeekSize]byte
	if err := usbip.ReadExactly(conn, hdrBuf[:]); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	ver := binary.BigEndian.Uint16(hdrBuf[0:2])
	code := binary.BigEndian.Uint16(hdrBuf[2:4])

	if ver == usbip.Version && (code == usbip.OpReqDevlist || code == usbip.OpReqImport) {
		switch code {
		case usbip.OpReqDevlist:
			s.logger.Info("OP_REQ_DEVLIST")
			return s.handleDevList(conn)
		case usbip.OpReqImport:
			s.logger.Info("OP_REQ_IMPORT")
			dev, err := s.handleImport(conn, hdrBuf[:])
			if err != nil {
				return fmt.Errorf("handle import: %w", err)
			}
			return s.handleUrbStream(conn, dev)
		}
	}

	return fmt.Errorf("protocol violation: client sent URB data without OP_REQ_IMPORT")
}

func (s *Server) handleDevList(conn net.Conn) error {
	_ = conn.SetDeadline(time.Time{})
	var buf bytes.Buffer
	rep := usbip.MgmtHeader{Version: usbip.Version, Command: usbip.OpRepDevlist, Status: 0}
	_ = rep.Write(&buf)
	metas := s.getAllDeviceMetas()
	n := uint32(len(metas))
	dlh := usbip.DevListReplyHeader{NDevices: n}
	_ = dlh.Write(&buf)
	for _, m := range metas {
		desc := m.Dev.GetDescriptor()
		meta := m.Meta

		exp := usbip.ExportedDevice{
			ExportMeta:          meta,
			Speed:               desc.Device.Speed,
			IDVendor:            desc.Device.IDVendor,
			IDProduct:           desc.Device.IDProduct,
			BcdDevice:           desc.Device.BcdDevice,
			BDeviceClass:        desc.Device.BDeviceClass,
			BDeviceSubClass:     desc.Device.BDeviceSubClass,
			BDeviceProtocol:     desc.Device.BDeviceProtocol,
			BConfigurationValue: usbConfigValueDefault,
			BNumConfigurations:  desc.Device.BNumConfigurations,
			BNumInterfaces:      uint8(len(desc.Interfaces)),
		}

		for _, iface := range desc.Interfaces {
			exp.Interfaces = append(exp.Interfaces, usbip.InterfaceDesc{
				Class:    iface.Descriptor.BInterfaceClass,
				SubClass: iface.Descriptor.BInterfaceSubClass,
				Protocol: iface.Descriptor.BInterfaceProtocol,
			})
		}
		_ = exp.WriteDevlist(&buf)
	}
	if _, err := conn.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("write devlist: %w", err)
	}
	return nil
}

func (s *Server) handleImport(conn net.Conn, first8 []byte) (usb.Device, error) {
	var rest [busIDSize]byte
	if err := usbip.ReadExactly(conn, rest[:]); err != nil {
		return nil, fmt.Errorf("read import busid: %w", err)
	}
	reqBus := string(rest[:bytes.IndexByte(rest[:], 0)])
	s.logger.Info("Import request", "busid", reqBus)
	var chosen usb.Device
	var chosenMeta *usbip.ExportMeta
	var chosenDesc *usb.Descriptor
	for _, m := range s.getAllDeviceMetas() {
		meta := m.Meta
		end := bytes.IndexByte(meta.USBBusId[:], 0)
		bid := string(meta.USBBusId[:end])
		if bid == reqBus {
			chosen = m.Dev
			chosenMeta = &meta
			chosenDesc = m.Dev.GetDescriptor()
			break
		}
	}
	if chosen == nil || chosenMeta == nil || chosenDesc == nil {
		return nil, fmt.Errorf("no device matches busid %s", reqBus)
	}
	var buf bytes.Buffer
	rep := usbip.MgmtHeader{Version: usbip.Version, Command: usbip.OpRepImport, Status: 0}
	_ = rep.Write(&buf)
	exp := usbip.ExportedDevice{
		ExportMeta:          *chosenMeta,
		Speed:               chosenDesc.Device.Speed,
		IDVendor:            chosenDesc.Device.IDVendor,
		IDProduct:           chosenDesc.Device.IDProduct,
		BcdDevice:           chosenDesc.Device.BcdDevice,
		BDeviceClass:        chosenDesc.Device.BDeviceClass,
		BDeviceSubClass:     chosenDesc.Device.BDeviceSubClass,
		BDeviceProtocol:     chosenDesc.Device.BDeviceProtocol,
		BConfigurationValue: usbConfigValueDefault,
		BNumConfigurations:  chosenDesc.Device.BNumConfigurations,
		BNumInterfaces:      uint8(len(chosenDesc.Interfaces)),
	}
	for _, iface := range chosenDesc.Interfaces {
		exp.Interfaces = append(exp.Interfaces, usbip.InterfaceDesc{
			Class:    iface.Descriptor.BInterfaceClass,
			SubClass: iface.Descriptor.BInterfaceSubClass,
			Protocol: iface.Descriptor.BInterfaceProtocol,
		})
	}
	_ = exp.WriteImport(&buf)
	if _, err := conn.Write(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("write import reply failed: %w", err)
	}
	return chosen, nil
}

// getAllDeviceMetas aggregates device metas from all registered busses.
func (s *Server) getAllDeviceMetas() []virtualbus.DeviceMeta {
	s.busesMu.Lock()
	defer s.busesMu.Unlock()
	out := []virtualbus.DeviceMeta{}
	for _, b := range s.busses {
		out = append(out, b.GetAllDeviceMetas()...)
	}
	return out
}

type readBufferConn struct {
	net.Conn
	buf []byte
}

func (r *readBufferConn) Read(p []byte) (int, error) {
	if len(r.buf) > 0 {
		n := copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}
	return r.Conn.Read(p)
}

type logConn struct {
	net.Conn
	s *Server
}

func (lc *logConn) Read(p []byte) (int, error) {
	n, err := lc.Conn.Read(p)
	if n > 0 && lc.s.rawLogger != nil {
		lc.s.rawLogger.Log(true, p[:n])
	}
	return n, err
}

func (lc *logConn) Write(p []byte) (int, error) {
	n, err := lc.Conn.Write(p)
	if n > 0 && lc.s.rawLogger != nil {
		lc.s.rawLogger.Log(false, p[:n])
	}
	return n, err
}

func (s *Server) handleUrbStream(conn net.Conn, dev usb.Device) error {
	_ = conn.SetDeadline(time.Time{})

	bw := newBatchingWriter(conn, writeBatcherBufferSize, writeBatcherFlushEvery, writeBatcherFlushAtBytes)
	defer func() { _ = bw.Close() }()

	var owningBus *virtualbus.VirtualBus
	for _, b := range s.busses {
		devices := b.Devices()
		for _, d := range devices {
			if d == dev {
				owningBus = b
				break
			}
		}
		if owningBus != nil {
			break
		}
	}
	if owningBus == nil {
		return fmt.Errorf("device does not belong to any bus")
	}

	ctx := owningBus.GetDeviceContext(dev)
	if ctx == nil {
		return fmt.Errorf("no device context available from bus")
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("device removed, closing URB stream")
			busID := owningBus.BusID()
			if emptyCtx := owningBus.GetBusEmptyContext(); emptyCtx != nil {
				go func() {
					slog.Debug("Started bus cleanup goroutine (HandleUrbStream ctx.Done)")
					select {
					case <-emptyCtx.Done():
						// Cancelled - a new device was added
						return
					case <-time.After(s.config.BusCleanupTimeout):
						if b := s.GetBus(busID); b != nil && len(b.Devices()) == 0 {
							if err := s.RemoveBus(busID); err != nil {
								s.logger.Error("timeout: failed to remove empty bus", "busID", busID, "error", err)
							} else {
								s.logger.Info("timeout: removed empty bus", "busID", busID)
							}
						}
					}
				}()
			} else {
				s.logger.Debug("No bus empty context; Cleaning bus immediately")
				if b := s.GetBus(busID); b != nil && len(b.Devices()) == 0 {
					if err := s.RemoveBus(busID); err != nil {
						s.logger.Error("timeout: failed to remove empty bus", "busID", busID, "error", err)
					} else {
						s.logger.Info("timeout: removed empty bus", "busID", busID)
					}
				}
			}
			return nil
		default:
		}

		var hdr [urbHdrSize]byte
		if err := usbip.ReadExactly(conn, hdr[:]); err != nil {
			return fmt.Errorf("read URB header: %w", err)
		}
		cmd := binary.BigEndian.Uint32(hdr[urbHdrOffsetCommand : urbHdrOffsetCommand+4])
		seq := binary.BigEndian.Uint32(hdr[urbHdrOffsetSeqnum : urbHdrOffsetSeqnum+4])
		devid := binary.BigEndian.Uint32(hdr[urbHdrOffsetDevid : urbHdrOffsetDevid+4])
		dir := binary.BigEndian.Uint32(hdr[urbHdrOffsetDir : urbHdrOffsetDir+4])
		ep := binary.BigEndian.Uint32(hdr[urbHdrOffsetEp : urbHdrOffsetEp+4])
		if cmd == usbip.CmdUnlinkCode {
			unlinkSeq := binary.BigEndian.Uint32(hdr[urbHdrOffsetUnlink : urbHdrOffsetUnlink+4])
			s.logger.Debug("USBIP_CMD_UNLINK", "seq", seq, "unlink", unlinkSeq)
			// Reply with -ECONNRESET
			ret := usbip.RetUnlink{Basic: usbip.HeaderBasic{Command: usbip.RetUnlinkCode, Seqnum: seq, Devid: 0, Dir: 0, Ep: 0}, Status: errConnReset}
			_ = ret.Write(bw)
			continue
		}
		if cmd != usbip.CmdSubmitCode {
			return fmt.Errorf("unsupported cmd %d (seq=%d, devid=%d)", cmd, seq, devid)
		}
		xferFlags := binary.BigEndian.Uint32(hdr[urbHdrOffsetFlags : urbHdrOffsetFlags+4])
		xferLen := binary.BigEndian.Uint32(hdr[urbHdrOffsetLength : urbHdrOffsetLength+4])
		setup := hdr[urbHdrOffsetSetup:urbHdrSize]

		var outPayload []byte
		if dir == usbip.DirOut && xferLen > 0 {
			outPayload = make([]byte, xferLen)
			if err := usbip.ReadExactly(conn, outPayload); err != nil {
				return fmt.Errorf("read OUT payload: %w", err)
			}
		}

		respData := s.processSubmit(dev, ep, dir, setup, outPayload)

		ret := usbip.RetSubmit{
			Basic:           usbip.HeaderBasic{Command: usbip.RetSubmitCode, Seqnum: seq, Devid: 0, Dir: 0, Ep: 0},
			Status:          0,
			ActualLength:    uint32(len(respData)),
			StartFrame:      0,
			NumberOfPackets: 0,
			ErrorCount:      0,
		}
		var out bytes.Buffer
		out.Grow(retSubmitHeaderSize)
		if err := ret.Write(&out); err != nil {
			return fmt.Errorf("build RET_SUBMIT header: %w", err)
		}
		if _, err := bw.Write(out.Bytes()); err != nil {
			return fmt.Errorf("write RET_SUBMIT: %w", err)
		}
		if len(respData) > 0 {
			if _, err := bw.Write(respData); err != nil {
				return fmt.Errorf("write RET_SUBMIT payload: %w", err)
			}
		}
		_ = xferFlags
		_ = devid
	}
}

// isClientDisconnect tests whether an error represents a normal client
// disconnect (EOF, ECONNRESET, broken pipe, or the Windows WSAECONNRESET
// translated error). We treat those as normal client disconnects and log
// them at Info level instead of Error.
func isClientDisconnect(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		switch t := opErr.Err.(type) {
		case syscall.Errno:
			if t == syscall.ECONNRESET || t == syscall.EPIPE {
				return true
			}
		}
	}
	e := strings.ToLower(err.Error())
	if strings.Contains(e, "connection reset by peer") || strings.Contains(e, "forcibly closed") || strings.Contains(e, "an existing connection was forcibly closed") || strings.Contains(e, "aborted") {
		return true
	}
	return false
}

func (s *Server) processSubmit(dev usb.Device, ep uint32, dir uint32, setup []byte, out []byte) []byte {
	if ep != 0 {
		return dev.HandleTransfer(ep, dir, out)
	}
	if len(setup) != 8 {
		return nil
	}
	bm := setup[0]
	breq := setup[1]
	wValue := binary.LittleEndian.Uint16(setup[2:4])
	wIndex := binary.LittleEndian.Uint16(setup[4:6])
	wLength := binary.LittleEndian.Uint16(setup[6:8])

	if breq == usbReqSetAddress && bm == usbReqTypeStandardToDevice {
		return nil
	}
	if breq == usbReqSetConfiguration && bm == usbReqTypeStandardToDevice {
		return nil
	}
	if breq == usbReqGetConfiguration && bm == usbReqTypeStandardFromDevice {
		return []byte{0x01}
	}

	desc := dev.GetDescriptor()

	if breq == usbReqGetDescriptor && bm == usbReqTypeStandardFromDevice {
		dtype := uint8(wValue >> 8)
		dindex := uint8(wValue & 0xff)
		var data []byte
		switch dtype {
		case usbDescTypeDevice:
			data = desc.Bytes()
		case usbDescTypeConfiguration:
			data = s.buildConfigDescriptor(desc)
		case usbDescTypeString:
			if s, ok := desc.Strings[dindex]; ok {
				data = usb.EncodeStringDescriptor(s)
			}
		}
		if len(data) == 0 {
			return nil
		}
		if int(wLength) < len(data) {
			return data[:wLength]
		}
		return data
	}
	if breq == usbReqGetDescriptor && bm == usbReqTypeStandardToInterface {
		dtype := uint8(wValue >> 8)
		iface := uint8(wIndex & 0xff)
		var data []byte
		if int(iface) < len(desc.Interfaces) {
			ifaceConf := desc.Interfaces[iface]
			if ifaceConf.HID != nil {
				switch dtype {
				case usbDescTypeHID:
					d, err := ifaceConf.HID.DescriptorBytes()
					if err != nil {
						s.logger.Error("failed to build HID descriptor", "iface", iface, "error", err)
						return nil
					}
					data = []byte(d)
				case usbDescTypeHIDReport:
					d, err := ifaceConf.HID.ReportBytes()
					if err != nil {
						s.logger.Error("failed to build HID report descriptor", "iface", iface, "error", err)
						return nil
					}
					data = []byte(d)
				}
			}
			if len(data) == 0 {
				for _, cd := range ifaceConf.ClassDescriptors {
					if cd.DescriptorType == dtype {
						data = []byte(cd.Bytes())
						break
					}
				}
			}
		}
		if len(data) == 0 {
			return nil
		}
		if int(wLength) < len(data) {
			return data[:wLength]
		}
		return data
	}
	_ = wIndex
	_ = out
	return nil
}

func (s *Server) buildConfigDescriptor(desc *usb.Descriptor) []byte {
	var b bytes.Buffer
	h := usb.ConfigHeader{
		WTotalLength:        0, // to be patched
		BNumInterfaces:      uint8(len(desc.Interfaces)),
		BConfigurationValue: usbConfigValueDefault,
		IConfiguration:      0,
		BMAttributes:        usbConfigAttrBusPowered,
		BMaxPower:           usbConfigMaxPower100mA,
	}
	h.Write(&b)
	for _, iface := range desc.Interfaces {
		iface.Descriptor.Write(&b)
		if iface.HID != nil {
			hd, err := iface.HID.DescriptorBytes()
			if err != nil {
				s.logger.Error("failed to build HID descriptor", "iface", iface.Descriptor.BInterfaceNumber, "error", err)
				// Stall/return minimal config descriptor.
				return nil
			}
			b.Write([]byte(hd))
		}
		for _, cd := range iface.ClassDescriptors {
			b.Write([]byte(cd.Bytes()))
		}
		for _, ep := range iface.Endpoints {
			ep.Write(&b)
		}
	}

	data := b.Bytes()
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(data)))
	return data
}
