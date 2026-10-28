package hidproxy

import (
	"encoding/binary"
	"fmt"
	"io"
)

// The device stream carries frames both ways: a 4-byte header (kind, interface number, payload length as
// little-endian uint16), then the payload.
const (
	// FrameInput (client to VIIPER): an input report for the interface's IN endpoint, as the device sends it over
	// USB (led by the report id if the interface uses ids).
	FrameInput byte = 0x01
	// FrameControlReply (client to VIIPER): the answer to a FrameControl. Payload: sequence (uint16), status (0 = done,
	// anything else = refused), then the data stage for a device-to-host request.
	FrameControlReply byte = 0x02
	// FrameOutput (VIIPER to client): what the host sent to an OUT endpoint of the interface.
	FrameOutput byte = 0x81
	// FrameControl (VIIPER to client): a control request for the client to answer. Payload: sequence (uint16), the
	// 8-byte setup packet as on the wire, then the data stage of a host-to-device request. The interface byte is the
	// low byte of wIndex for requests to an interface or endpoint, 0xFF for requests to the device.
	FrameControl byte = 0x82
)

const (
	frameHeaderSize = 4
	// MaxPayload is the longest payload a frame may carry: a control transfer's data stage is at most 4096 bytes here.
	MaxPayload    = 4096 + 2 + 8
	setupSize     = 8
	noInterface   = 0xFF
	statusDone    = 0
	replyHeadSize = 3
)

// Frame is one message on the device stream.
type Frame struct {
	Kind      byte
	Interface uint8
	Payload   []byte
}

// ReadFrame reads one frame; it fails on a payload longer than MaxPayload.
func ReadFrame(r io.Reader) (Frame, error) {
	var header [frameHeaderSize]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return Frame{}, err
	}
	length := int(binary.LittleEndian.Uint16(header[2:]))
	if length > MaxPayload {
		return Frame{}, fmt.Errorf("frame payload of %d bytes, at most %d", length, MaxPayload)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return Frame{}, err
	}
	return Frame{Kind: header[0], Interface: header[1], Payload: payload}, nil
}

// AppendFrame encodes a frame onto dst.
func AppendFrame(dst []byte, f Frame) ([]byte, error) {
	if len(f.Payload) > MaxPayload {
		return dst, fmt.Errorf("frame payload of %d bytes, at most %d", len(f.Payload), MaxPayload)
	}
	dst = append(dst, f.Kind, f.Interface)
	dst = binary.LittleEndian.AppendUint16(dst, uint16(len(f.Payload)))
	return append(dst, f.Payload...), nil
}

// controlFrame builds the FrameControl for a setup packet and its OUT data.
func controlFrame(seq uint16, bmRequestType, bRequest uint8, wValue, wIndex, wLength uint16, data []byte) Frame {
	payload := make([]byte, 0, 2+setupSize+len(data))
	payload = binary.LittleEndian.AppendUint16(payload, seq)
	payload = append(payload, bmRequestType, bRequest)
	payload = binary.LittleEndian.AppendUint16(payload, wValue)
	payload = binary.LittleEndian.AppendUint16(payload, wIndex)
	payload = binary.LittleEndian.AppendUint16(payload, wLength)
	payload = append(payload, data...)
	iface := uint8(noInterface)
	if recipient := bmRequestType & 0x1F; recipient == recipientInterface || recipient == recipientEndpoint {
		iface = uint8(wIndex)
	}
	return Frame{Kind: FrameControl, Interface: iface, Payload: payload}
}

// controlReply is a client's answer to a FrameControl.
type controlReply struct {
	seq  uint16
	done bool
	data []byte
}

func parseControlReply(payload []byte) (controlReply, error) {
	if len(payload) < replyHeadSize {
		return controlReply{}, fmt.Errorf("control reply of %d bytes, at least %d", len(payload), replyHeadSize)
	}
	return controlReply{
		seq:  binary.LittleEndian.Uint16(payload),
		done: payload[2] == statusDone,
		data: payload[replyHeadSize:],
	}, nil
}
