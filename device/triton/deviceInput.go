package triton

import (
	"encoding/binary"
	"io"
)

// InputState is the client-facing input state for a Valve Steam Triton wired controller.
//
// This struct mirrors SDL's TritonMTUFull_t (TritonMTUNoQuat_t + gyro quaternions)
// from SDL3 `src/joystick/hidapi/steam/controller_structs.h`.
//
// Wire format (client → device stream): fixed 52 bytes, little-endian.
// viiper:wire triton c2s buttons:u32 triggerLeft:i16 triggerRight:i16 leftStickX:i16 leftStickY:i16 rightStickX:i16 rightStickY:i16 leftPadX:i16 leftPadY:i16 pressureLeft:u16 rightPadX:i16 rightPadY:i16 pressureRight:u16 timestamp:u32 accelX:i16 accelY:i16 accelZ:i16 gyroX:i16 gyroY:i16 gyroZ:i16 gyroQuatW:i16 gyroQuatX:i16 gyroQuatY:i16 gyroQuatZ:i16
type InputState struct {
	// Button bitmask — see ButtonXxx constants.
	Buttons uint32

	TriggerLeft  int16
	TriggerRight int16

	LeftStickX  int16
	LeftStickY  int16
	RightStickX int16
	RightStickY int16

	LeftPadX     int16
	LeftPadY     int16
	PressureLeft uint16

	RightPadX     int16
	RightPadY     int16
	PressureRight uint16

	Timestamp uint32

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
}

// MarshalBinary encodes InputState to the fixed 52-byte wire format.
func (s InputState) MarshalBinary() ([]byte, error) {
	b := make([]byte, InputStateSize)
	o := 0

	putU32 := func(v uint32) {
		binary.LittleEndian.PutUint32(b[o:o+4], v)
		o += 4
	}
	putI16 := func(v int16) {
		binary.LittleEndian.PutUint16(b[o:o+2], uint16(v))
		o += 2
	}
	putU16 := func(v uint16) {
		binary.LittleEndian.PutUint16(b[o:o+2], v)
		o += 2
	}

	putU32(s.Buttons)
	putI16(s.TriggerLeft)
	putI16(s.TriggerRight)
	putI16(s.LeftStickX)
	putI16(s.LeftStickY)
	putI16(s.RightStickX)
	putI16(s.RightStickY)
	putI16(s.LeftPadX)
	putI16(s.LeftPadY)
	putU16(s.PressureLeft)
	putI16(s.RightPadX)
	putI16(s.RightPadY)
	putU16(s.PressureRight)
	putU32(s.Timestamp)
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

	return b, nil
}

// UnmarshalBinary decodes InputState from the fixed 52-byte wire format.
func (s *InputState) UnmarshalBinary(data []byte) error {
	if len(data) < InputStateSize {
		return io.ErrUnexpectedEOF
	}
	o := 0

	getU32 := func() uint32 {
		v := binary.LittleEndian.Uint32(data[o : o+4])
		o += 4
		return v
	}
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

	s.Buttons = getU32()
	s.TriggerLeft = getI16()
	s.TriggerRight = getI16()
	s.LeftStickX = getI16()
	s.LeftStickY = getI16()
	s.RightStickX = getI16()
	s.RightStickY = getI16()
	s.LeftPadX = getI16()
	s.LeftPadY = getI16()
	s.PressureLeft = getU16()
	s.RightPadX = getI16()
	s.RightPadY = getI16()
	s.PressureRight = getU16()
	s.Timestamp = getU32()
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

	return nil
}
