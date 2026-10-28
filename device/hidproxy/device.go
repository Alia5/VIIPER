// Package hidproxy provides a device that copies another one: it enumerates with descriptors the client supplies
// and passes reports between the host and the client, which reads and writes the real device itself and decides
// what passes. VIIPER never talks to the real device: whatever the host sends goes only to the client.
//
// Example use: the real device is hidden from an application (e.g. with HidHide) and the application sees this copy
// instead, so the client decides what reaches it. Nothing in the device is specific to one kind of hardware.
package hidproxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usbip"
)

const (
	classHID     = 0x03
	classCDC     = 0x02
	classCDCData = 0x0A

	requestTypeMask    = 0x60
	requestTypeClass   = 0x20
	recipientMask      = 0x1F
	recipientInterface = 0x01
	recipientEndpoint  = 0x02
	directionIn        = 0x80

	hidGetIdle     = 0x02
	hidGetProtocol = 0x03
	hidSetIdle     = 0x0A
	hidSetProtocol = 0x0B

	cdcSetLineCoding       = 0x20
	cdcGetLineCoding       = 0x21
	cdcSetControlLineState = 0x22
	lineCodingLength       = 7

	inQueueLength = 64
)

var (
	// ControlTimeout is how long a control request waits for a client to attach and answer. Windows gives a control
	// transfer about 5 s; a client that takes longer than this is treated as having refused it.
	ControlTimeout = time.Second
	// WriteTimeout bounds a write to the client, so a client that stops reading can't hold up the device.
	WriteTimeout = time.Second
	// 115200 baud, 1 stop bit, no parity, 8 data bits: what a CDC ACM device reports before the host set anything.
	defaultLineCoding = []byte{0x00, 0xC2, 0x01, 0x00, 0x00, 0x00, 0x08}
)

// HIDProxy is the copy. Create it with New; the stream handler attaches the client.
type HIDProxy struct {
	desc       usb.Descriptor
	inQueues   map[uint32]chan []byte // by IN endpoint number
	inByIface  map[uint8]uint32       // the IN endpoint that carries an interface's input reports
	ifaceByOut map[uint32]uint8       // OUT endpoint number to interface
	classOf    map[uint8]uint8        // interface number to class

	mu         sync.Mutex
	client     net.Conn
	attached   chan struct{} // closed while a client is attached
	writeMu    sync.Mutex
	seq        uint16
	pending    map[uint16]chan controlReply
	lineCoding map[uint8][]byte
}

// New builds a copy from the descriptors in o.
func New(o Options) (*HIDProxy, error) {
	desc, err := parseOptions(o)
	if err != nil {
		return nil, err
	}
	p := &HIDProxy{
		desc:       *desc,
		inQueues:   map[uint32]chan []byte{},
		inByIface:  map[uint8]uint32{},
		ifaceByOut: map[uint32]uint8{},
		classOf:    map[uint8]uint8{},
		attached:   make(chan struct{}),
		pending:    map[uint16]chan controlReply{},
		lineCoding: map[uint8][]byte{},
	}
	for _, iface := range desc.Interfaces {
		number := iface.Descriptor.BInterfaceNumber
		p.classOf[number] = iface.Descriptor.BInterfaceClass
		for _, ep := range iface.Endpoints {
			address := uint32(ep.BEndpointAddress & 0x0F)
			if ep.BEndpointAddress&directionIn == 0 {
				p.ifaceByOut[address] = number
				continue
			}
			if _, ok := p.inQueues[address]; !ok {
				p.inQueues[address] = make(chan []byte, inQueueLength)
			}
			// Input reports go to the interface's interrupt IN endpoint, else to its first IN endpoint.
			if current, ok := p.inByIface[number]; !ok || (ep.BMAttributes&0x03 == 0x03 && !p.isInterrupt(current)) {
				p.inByIface[number] = address
			}
		}
	}
	return p, nil
}

func (p *HIDProxy) isInterrupt(address uint32) bool {
	for _, iface := range p.desc.Interfaces {
		for _, ep := range iface.Endpoints {
			if uint32(ep.BEndpointAddress&0x0F) == address && ep.BEndpointAddress&directionIn != 0 {
				return ep.BMAttributes&0x03 == 0x03
			}
		}
	}
	return false
}

// NoInRepeat tells VIIPER's server not to resend the last input report when none came within bInterval: the copy
// passes the real device's reports, and some of them are events, not state.
func (p *HIDProxy) NoInRepeat() bool { return true }

// GetDescriptor returns the copied descriptors.
func (p *HIDProxy) GetDescriptor() *usb.Descriptor { return &p.desc }

// GetDeviceSpecificArgs returns nothing: the descriptors hold the device's serial number, which the API would
// otherwise echo into responses and logs.
func (p *HIDProxy) GetDeviceSpecificArgs() map[string]any { return map[string]any{} }

// HandleTransfer serves the interrupt and bulk endpoints: IN waits for the client's next report, OUT goes to the
// client (or nowhere without one).
func (p *HIDProxy) HandleTransfer(ctx context.Context, ep uint32, dir uint32, out []byte) []byte {
	if dir == usbip.DirIn {
		queue := p.inQueues[ep]
		if queue == nil {
			<-ctx.Done()
			return nil
		}
		select {
		case report := <-queue:
			return report
		case <-ctx.Done():
			return nil
		}
	}
	iface, ok := p.ifaceByOut[ep]
	if !ok || len(out) == 0 {
		return nil
	}
	_ = p.send(Frame{Kind: FrameOutput, Interface: iface, Payload: append([]byte(nil), out...)})
	return nil
}

// HandleControl answers the class and vendor requests VIIPER's server doesn't. A HID interface's idle rate and
// protocol and a CDC interface's line coding and control line state are the host's own business with the copy and
// are answered here; every other request (HID GET_REPORT and SET_REPORT, other CDC requests such as SEND_BREAK,
// vendor and unknown class requests) goes to the client, which decides whether anything reaches the real device and
// may take it as a sign of a firmware update.
func (p *HIDProxy) HandleControl(bmRequestType, bRequest uint8, wValue, wIndex, wLength uint16, data []byte) ([]byte, bool) {
	if bmRequestType&requestTypeMask == 0 {
		// Standard requests (descriptors, configuration, features) stay VIIPER's.
		return nil, false
	}
	if bmRequestType&requestTypeMask == requestTypeClass && bmRequestType&recipientMask == recipientInterface {
		iface := uint8(wIndex)
		switch p.classOf[iface] {
		case classHID:
			if bRequest == hidGetIdle || bRequest == hidSetIdle || bRequest == hidGetProtocol || bRequest == hidSetProtocol {
				return nil, false
			}
		case classCDC, classCDCData:
			if bRequest == cdcSetLineCoding || bRequest == cdcGetLineCoding || bRequest == cdcSetControlLineState {
				return p.lineControl(iface, bRequest, data), true
			}
		}
	}
	return p.askClient(bmRequestType, bRequest, wValue, wIndex, wLength, data), true
}

func (p *HIDProxy) lineControl(iface uint8, bRequest uint8, data []byte) []byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch bRequest {
	case cdcSetLineCoding:
		if len(data) >= lineCodingLength {
			p.lineCoding[iface] = append([]byte(nil), data[:lineCodingLength]...)
		}
	case cdcGetLineCoding:
		if coding, ok := p.lineCoding[iface]; ok {
			return append([]byte(nil), coding...)
		}
		return append([]byte(nil), defaultLineCoding...)
	case cdcSetControlLineState:
	}
	return nil
}

// askClient sends the request to the client and waits for its answer. No client within ControlTimeout, or a refusal,
// answers with no data: VIIPER's server can't stall a request, so the host sees an empty data stage.
func (p *HIDProxy) askClient(bmRequestType, bRequest uint8, wValue, wIndex, wLength uint16, data []byte) []byte {
	timer := time.NewTimer(ControlTimeout)
	defer timer.Stop()

	p.mu.Lock()
	attached := p.attached
	p.mu.Unlock()
	select {
	case <-attached:
	case <-timer.C:
		return nil
	}

	answer := make(chan controlReply, 1)
	p.mu.Lock()
	p.seq++
	seq := p.seq
	p.pending[seq] = answer
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		delete(p.pending, seq)
		p.mu.Unlock()
	}()

	if err := p.send(controlFrame(seq, bmRequestType, bRequest, wValue, wIndex, wLength, data)); err != nil {
		return nil
	}
	select {
	case reply := <-answer:
		if !reply.done || bmRequestType&directionIn == 0 {
			return nil
		}
		if len(reply.data) > int(wLength) {
			return reply.data[:wLength]
		}
		return reply.data
	case <-timer.C:
		return nil
	}
}

// send writes a frame to the client; without a client the frame is dropped.
func (p *HIDProxy) send(f Frame) error {
	p.mu.Lock()
	client := p.client
	p.mu.Unlock()
	if client == nil {
		return errNoClient
	}
	buffer, err := AppendFrame(nil, f)
	if err != nil {
		return err
	}
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	_ = client.SetWriteDeadline(time.Now().Add(WriteTimeout))
	_, err = client.Write(buffer)
	return err
}

var (
	errNoClient        = errors.New("no client attached")
	errAlreadyAttached = errors.New("hidproxy: a client is already attached")
)

// Serve attaches conn as the client and handles its frames until it disconnects; only one client at a time.
func (p *HIDProxy) Serve(conn net.Conn) error {
	p.mu.Lock()
	if p.client != nil {
		p.mu.Unlock()
		return errAlreadyAttached
	}
	p.client = conn
	close(p.attached)
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.client = nil
		p.attached = make(chan struct{})
		for seq, answer := range p.pending {
			delete(p.pending, seq)
			close(answer)
		}
		p.mu.Unlock()
	}()

	for {
		f, err := ReadFrame(conn)
		if err != nil {
			return err
		}
		switch f.Kind {
		case FrameInput:
			p.queueInput(f.Interface, f.Payload)
		case FrameControlReply:
			reply, err := parseControlReply(f.Payload)
			if err != nil {
				return err
			}
			p.mu.Lock()
			answer := p.pending[reply.seq]
			delete(p.pending, reply.seq)
			p.mu.Unlock()
			if answer != nil {
				answer <- reply
			}
		default:
			return fmt.Errorf("hidproxy: unknown frame kind 0x%02X from the client", f.Kind)
		}
	}
}

// queueInput hands a report to the interface's IN endpoint; when the host doesn't poll fast enough the oldest
// report goes, so the host always gets the newest.
func (p *HIDProxy) queueInput(iface uint8, report []byte) {
	ep, ok := p.inByIface[iface]
	if !ok || len(report) == 0 {
		return
	}
	queue := p.inQueues[ep]
	for {
		select {
		case queue <- report:
			return
		default:
		}
		select {
		case <-queue:
		default:
		}
	}
}
