package mouse

import (
	"io"
)

// InputState represents the mouse state used to build a report.
// viiper:wire mouse c2s buttons:u8 dx:i16 dy:i16 wheel:i16 pan:i16
type InputState struct {
	// Button bitfield: bit 0=Left, 1=Right, 2=Middle, 3=Back, 4=Forward
	Buttons uint8
	// Delta X/Y: signed 16-bit relative movement
	DX, DY int16
	// Wheel: signed 16-bit vertical scroll
	Wheel int16
	// Pan: signed 16-bit horizontal scroll
	Pan int16
}

// BuildReport encodes an InputState into the 9-byte HID mouse report.
//
// Report layout (9 bytes):
//
//	Byte 0: Button bitfield (bit 0=Left, 1=Right, 2=Middle, 3=Back, 4=Forward, bits 5-7=padding)
//	Bytes 1-2: DX (int16 little-endian, -32768 to +32767)
//	Bytes 3-4: DY (int16 little-endian)
//	Bytes 5-6: Wheel (int16 little-endian)
//	Bytes 7-8: Pan (int16 little-endian)
func (st InputState) BuildReport() []byte {
	b := make([]byte, 9)
	b[0] = st.Buttons & 0x1F // 5 buttons, mask upper bits
	b[1] = byte(st.DX)
	b[2] = byte(st.DX >> 8)
	b[3] = byte(st.DY)
	b[4] = byte(st.DY >> 8)
	b[5] = byte(st.Wheel)
	b[6] = byte(st.Wheel >> 8)
	b[7] = byte(st.Pan)
	b[8] = byte(st.Pan >> 8)
	return b
}

// MarshalBinary encodes InputState to 9 bytes.
func (m *InputState) MarshalBinary() ([]byte, error) {
	b := make([]byte, 9)
	b[0] = m.Buttons
	b[1] = byte(m.DX)
	b[2] = byte(m.DX >> 8)
	b[3] = byte(m.DY)
	b[4] = byte(m.DY >> 8)
	b[5] = byte(m.Wheel)
	b[6] = byte(m.Wheel >> 8)
	b[7] = byte(m.Pan)
	b[8] = byte(m.Pan >> 8)
	return b, nil
}

// UnmarshalBinary decodes 9 bytes into InputState.
func (m *InputState) UnmarshalBinary(data []byte) error {
	if len(data) < 9 {
		return io.ErrUnexpectedEOF
	}
	m.Buttons = data[0]
	m.DX = int16(data[1]) | int16(data[2])<<8
	m.DY = int16(data[3]) | int16(data[4])<<8
	m.Wheel = int16(data[5]) | int16(data[6])<<8
	m.Pan = int16(data[7]) | int16(data[8])<<8
	return nil
}
