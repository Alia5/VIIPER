package xbox360

import "encoding/binary"

// BuildReport encodes an InputState into the 20-byte Xbox 360 USB report
// layout used by the wired controller. The layout matches the existing
// placeholder used by this project: two-byte header then fields.
//
// Bytes:
//  0: 0x00 (report id/message)
//  1: 0x14 (payload size 20)
//  2-7: buttons/reserved (we use first two bytes for button bitfield)
//  8: LT (0-255)
//  9: RT (0-255)
// 10-11: LX (LE int16)
// 12-13: LY (LE int16)
// 14-15: RX (LE int16)
// 16-17: RY (LE int16)
// 18-19: reserved 0x00
func BuildReport(st InputState) []byte {
	b := make([]byte, 20)
	b[0] = 0x00
	b[1] = 0x14
	binary.LittleEndian.PutUint16(b[2:4], uint16(st.Buttons&0xffff))
	b[8] = st.LT
	b[9] = st.RT
	binary.LittleEndian.PutUint16(b[10:12], uint16(st.LX))
	binary.LittleEndian.PutUint16(b[12:14], uint16(st.LY))
	binary.LittleEndian.PutUint16(b[14:16], uint16(st.RX))
	binary.LittleEndian.PutUint16(b[16:18], uint16(st.RY))
	return b
}
