package xbox360

import (
	"encoding/binary"
	"io"
)

// InputState represents the controller state used to build a report.
// Values are more or less XInput's C API
// viiper:wire xbox360 c2s buttons:u32 lt:u8 rt:u8 lx:i16 ly:i16 rx:i16 ry:i16
type InputState struct {
	// Button bitfield (lower 16 bits used typically), higher bits reserved
	Buttons uint32
	// Triggers: 0-255
	LT, RT uint8
	// Sticks: signed 16-bit little endian values
	LX, LY int16
	RX, RY int16
}

// BuildReport encodes an InputState into the 20-byte Xbox 360 wired USB input report.
// Layout (indices in the returned slice):
//
//	 0: 0x00              - Report ID
//	 1: 0x14              - Payload size (20 bytes)
//	 2: Buttons (low byte)
//	 3: Buttons (high byte)
//	 4: LT (0-255)
//	 5: RT (0-255)
//	 6-7: LX (little-endian int16)
//	 8-9: LY (little-endian int16)
//	10-11: RX (little-endian int16)
//	12-13: RY (little-endian int16)
//	14-19: Reserved / zero
func (st InputState) BuildReport() []byte {
	b := make([]byte, 20)
	b[0] = 0x00
	b[1] = 0x14
	binary.LittleEndian.PutUint16(b[2:4], uint16(st.Buttons&0xffff))
	b[4] = st.LT
	b[5] = st.RT
	binary.LittleEndian.PutUint16(b[6:8], uint16(st.LX))
	binary.LittleEndian.PutUint16(b[8:10], uint16(st.LY))
	binary.LittleEndian.PutUint16(b[10:12], uint16(st.RX))
	binary.LittleEndian.PutUint16(b[12:14], uint16(st.RY))
	// Remaining bytes (14-19) are left zeroed
	return b
}

// MarshalBinary encodes InputState to 14 bytes.
func (x *InputState) MarshalBinary() ([]byte, error) {
	b := make([]byte, 14)
	binary.LittleEndian.PutUint32(b[0:4], x.Buttons)
	b[4] = x.LT
	b[5] = x.RT
	binary.LittleEndian.PutUint16(b[6:8], uint16(x.LX))
	binary.LittleEndian.PutUint16(b[8:10], uint16(x.LY))
	binary.LittleEndian.PutUint16(b[10:12], uint16(x.RX))
	binary.LittleEndian.PutUint16(b[12:14], uint16(x.RY))
	return b, nil
}

// UnmarshalBinary decodes 14 bytes into InputState.
func (x *InputState) UnmarshalBinary(data []byte) error {
	if len(data) < 14 {
		return io.ErrUnexpectedEOF
	}
	x.Buttons = binary.LittleEndian.Uint32(data[0:4])
	x.LT = data[4]
	x.RT = data[5]
	x.LX = int16(binary.LittleEndian.Uint16(data[6:8]))
	x.LY = int16(binary.LittleEndian.Uint16(data[8:10]))
	x.RX = int16(binary.LittleEndian.Uint16(data[10:12]))
	x.RY = int16(binary.LittleEndian.Uint16(data[12:14]))
	return nil
}

// XRumbleState is the wire format for rumble/motor commands sent from device to client.
// Total size: 2 bytes (fixed).
// Layout:
//
//	LeftMotor: 1 byte (0-255)
//	RightMotor: 1 byte (0-255)
//
// viiper:wire xbox360 s2c left:u8 right:u8
type XRumbleState struct {
	LeftMotor  uint8
	RightMotor uint8
}

// MarshalBinary encodes XRumbleState to 2 bytes.
func (r *XRumbleState) MarshalBinary() ([]byte, error) {
	return []byte{r.LeftMotor, r.RightMotor}, nil
}

// UnmarshalBinary decodes 2 bytes into XRumbleState.
func (r *XRumbleState) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return io.ErrUnexpectedEOF
	}
	r.LeftMotor = data[0]
	r.RightMotor = data[1]
	return nil
}
