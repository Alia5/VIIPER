package steamdeck

import (
	"encoding/binary"
	"io"
)

// InputState is the client-facing input state for a Steam Deck (Jupiter/LCD)
// controller.
//
// This struct mirrors SDL's `SteamDeckStatePacket_t` fields (minus unPacketNum)
//
// Wire format (client -> device stream): fixed 52 bytes, little-endian.
// viiper:wire steamdeck c2s buttons:u64 leftPadX:i16 leftPadY:i16 rightPadX:i16 rightPadY:i16 accelX:i16 accelY:i16 accelZ:i16 gyroX:i16 gyroY:i16 gyroZ:i16 gyroQuatW:i16 gyroQuatX:i16 gyroQuatY:i16 gyroQuatZ:i16 triggerL:u16 triggerR:u16 leftStickX:i16 leftStickY:i16 rightStickX:i16 rightStickY:i16 pressurePadLeft:u16 pressurePadRight:u16
type InputState struct {
	Buttons uint64

	LeftPadX  int16
	LeftPadY  int16
	RightPadX int16
	RightPadY int16

	AccelX int16
	AccelY int16
	AccelZ int16

	GyroX int16
	GyroY int16
	GyroZ int16

	GyroQuatW int16
	GyroQuatX int16
	GyroQuatY int16
	GyroQuatZ int16

	TriggerRawL uint16
	TriggerRawR uint16

	LeftStickX int16
	LeftStickY int16

	RightStickX int16
	RightStickY int16

	PressurePadLeft  uint16
	PressurePadRight uint16
}

// MarshalBinary encodes InputState to the fixed 52-byte wire format.
func (s InputState) MarshalBinary() ([]byte, error) {
	b := make([]byte, InputStateSize)
	o := 0

	binary.LittleEndian.PutUint64(b[o:o+8], s.Buttons)
	o += 8

	putI16 := func(v int16) {
		binary.LittleEndian.PutUint16(b[o:o+2], uint16(v))
		o += 2
	}
	putU16 := func(v uint16) {
		binary.LittleEndian.PutUint16(b[o:o+2], v)
		o += 2
	}

	putI16(s.LeftPadX)
	putI16(s.LeftPadY)
	putI16(s.RightPadX)
	putI16(s.RightPadY)

	putI16(s.AccelX)
	putI16(s.AccelY)
	putI16(s.AccelZ)

	putI16(s.GyroX)
	putI16(s.GyroY)
	putI16(s.GyroZ)

	putI16(s.GyroQuatW)
	putI16(s.GyroQuatX)
	putI16(s.GyroQuatY)
	putI16(s.GyroQuatZ)

	putU16(s.TriggerRawL)
	putU16(s.TriggerRawR)

	putI16(s.LeftStickX)
	putI16(s.LeftStickY)
	putI16(s.RightStickX)
	putI16(s.RightStickY)

	putU16(s.PressurePadLeft)
	putU16(s.PressurePadRight)

	return b, nil
}

// UnmarshalBinary decodes InputState from the fixed 52-byte wire format.
func (s *InputState) UnmarshalBinary(data []byte) error {
	if len(data) < InputStateSize {
		return io.ErrUnexpectedEOF
	}
	o := 0

	s.Buttons = binary.LittleEndian.Uint64(data[o : o+8])
	o += 8

	getI16 := func() int16 {
		v := int16(binary.LittleEndian.Uint16(data[o : o+2]))
		o += 2
		return v
	}
	getU16 := func() uint16 {
		v := binary.LittleEndian.Uint16(data[o : o+2])
		o += 2
		return v
	}

	s.LeftPadX = getI16()
	s.LeftPadY = getI16()
	s.RightPadX = getI16()
	s.RightPadY = getI16()

	s.AccelX = getI16()
	s.AccelY = getI16()
	s.AccelZ = getI16()

	s.GyroX = getI16()
	s.GyroY = getI16()
	s.GyroZ = getI16()

	s.GyroQuatW = getI16()
	s.GyroQuatX = getI16()
	s.GyroQuatY = getI16()
	s.GyroQuatZ = getI16()

	s.TriggerRawL = getU16()
	s.TriggerRawR = getU16()

	s.LeftStickX = getI16()
	s.LeftStickY = getI16()
	s.RightStickX = getI16()
	s.RightStickY = getI16()

	s.PressurePadLeft = getU16()
	s.PressurePadRight = getU16()

	return nil
}
