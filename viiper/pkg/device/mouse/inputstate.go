package mouse

import (
	"io"
)

// InputState represents the mouse state used to build a report.
// viiper:wire mouse c2s buttons:u8 dx:i8 dy:i8 wheel:i8 pan:i8
type InputState struct {
	// Button bitfield: bit 0=Left, 1=Right, 2=Middle, 3=Back, 4=Forward
	Buttons uint8
	// Delta X/Y: signed 8-bit relative movement
	DX, DY int8
	// Wheel: signed 8-bit vertical scroll
	Wheel int8
	// Pan: signed 8-bit horizontal scroll
	Pan int8
}

// BuildReport encodes an InputState into the 5-byte HID mouse report.
//
// Report layout (5 bytes):
//
//	Byte 0: Button bitfield (bit 0=Left, 1=Right, 2=Middle, 3=Back, 4=Forward, bits 5-7=padding)
//	Byte 1: DX (int8, -127 to +127)
//	Byte 2: DY (int8)
//	Byte 3: Wheel (int8)
//	Byte 4: Pan (int8)
func (st InputState) BuildReport() []byte {
	b := make([]byte, 5)
	b[0] = st.Buttons & 0x1F // 5 buttons, mask upper bits
	b[1] = byte(st.DX)
	b[2] = byte(st.DY)
	b[3] = byte(st.Wheel)
	b[4] = byte(st.Pan)
	return b
}

// MarshalBinary encodes InputState to 5 bytes.
func (m *InputState) MarshalBinary() ([]byte, error) {
	b := make([]byte, 5)
	b[0] = m.Buttons
	b[1] = byte(m.DX)
	b[2] = byte(m.DY)
	b[3] = byte(m.Wheel)
	b[4] = byte(m.Pan)
	return b, nil
}

// UnmarshalBinary decodes 5 bytes into InputState.
func (m *InputState) UnmarshalBinary(data []byte) error {
	if len(data) < 5 {
		return io.ErrUnexpectedEOF
	}
	m.Buttons = data[0]
	m.DX = int8(data[1])
	m.DY = int8(data[2])
	m.Wheel = int8(data[3])
	m.Pan = int8(data[4])
	return nil
}
