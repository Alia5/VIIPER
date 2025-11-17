package keyboard

import (
	"io"
)

// InputState represents the keyboard state used to build a report.
// Internally uses a 256-bit bitmap for N-key rollover support.
// viiper:wire keyboard c2s modifiers:u8 count:u8 keys:u8*count
type InputState struct {
	Modifiers uint8     // bit 0-7: LCtrl, LShift, LAlt, LGui, RCtrl, RShift, RAlt, RGui
	KeyBitmap [32]uint8 // 256 bits for HID usage codes 0x00-0xFF
}

// LEDState represents the state of keyboard LEDs controlled by the host.
// viiper:wire keyboard s2c leds:u8
type LEDState struct {
	NumLock    bool
	CapsLock   bool
	ScrollLock bool
	Compose    bool
	Kana       bool
}

// UnmarshalBinary decodes a 1-byte LED bitmask into LEDState.
// Bits are defined by LEDNumLock, LEDCapsLock, LEDScrollLock, LEDCompose, LEDKana.
func (st *LEDState) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return io.ErrUnexpectedEOF
	}
	b := data[0]
	st.NumLock = b&LEDNumLock != 0
	st.CapsLock = b&LEDCapsLock != 0
	st.ScrollLock = b&LEDScrollLock != 0
	st.Compose = b&LEDCompose != 0
	st.Kana = b&LEDKana != 0
	return nil
}

// BuildReport encodes an InputState into the 34-byte HID keyboard report.
//
// Report layout (34 bytes):
//
//	Byte 0: Modifiers (8 bits)
//	Byte 1: Reserved (0x00)
//	Bytes 2-33: Key bitmap (256 bits, 32 bytes)
func (st InputState) BuildReport() []byte {
	b := make([]byte, 34)
	b[0] = st.Modifiers
	b[1] = 0x00 // Reserved
	copy(b[2:34], st.KeyBitmap[:])
	return b
}

// MarshalBinary encodes InputState to variable-length wire format.
//
// Wire format:
//
//	Byte 0: Modifiers
//	Byte 1: Key count
//	Bytes 2+: Key codes (HID usage codes of pressed keys)
func (st *InputState) MarshalBinary() ([]byte, error) {
	// Count pressed keys
	var keys []uint8
	for i := 0; i < 256; i++ {
		byteIdx := i / 8
		bitIdx := uint(i % 8)
		if st.KeyBitmap[byteIdx]&(1<<bitIdx) != 0 {
			keys = append(keys, uint8(i))
		}
	}

	// Build packet: [modifiers, count, key1, key2, ...]
	b := make([]byte, 2+len(keys))
	b[0] = st.Modifiers
	b[1] = uint8(len(keys))
	copy(b[2:], keys)
	return b, nil
}

// UnmarshalBinary decodes variable-length wire format into InputState.
//
// Wire format:
//
//	Byte 0: Modifiers
//	Byte 1: Key count
//	Bytes 2+: Key codes (HID usage codes of pressed keys)
func (st *InputState) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return io.ErrUnexpectedEOF
	}

	st.Modifiers = data[0]
	keyCount := int(data[1])

	if len(data) < 2+keyCount {
		return io.ErrUnexpectedEOF
	}

	// Clear bitmap
	for i := range st.KeyBitmap {
		st.KeyBitmap[i] = 0
	}

	// Set bits for each key
	for i := 0; i < keyCount; i++ {
		keyCode := data[2+i]
		byteIdx := keyCode / 8
		bitIdx := uint(keyCode % 8)
		st.KeyBitmap[byteIdx] |= 1 << bitIdx
	}

	return nil
}
