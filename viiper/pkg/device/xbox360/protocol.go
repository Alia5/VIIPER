package xbox360

import (
	"encoding/binary"
	"io"
)

// XInputState is the wire format for controller inputs sent from client to device.
// Total size: 14 bytes (fixed).
// Layout:
//
//	Buttons: 4 bytes (LE uint32)
//	LT: 1 byte
//	RT: 1 byte
//	LX: 2 bytes (LE int16)
//	LY: 2 bytes (LE int16)
//	RX: 2 bytes (LE int16)
//	RY: 2 bytes (LE int16)
type XInputState struct {
	Buttons uint32
	LT, RT  uint8
	LX, LY  int16
	RX, RY  int16
}

// MarshalBinary encodes XInputState to 12 bytes.
func (x *XInputState) MarshalBinary() ([]byte, error) {
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

// UnmarshalBinary decodes 14 bytes into XInputState.
func (x *XInputState) UnmarshalBinary(data []byte) error {
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

// ToInputState converts XInputState to internal InputState for report building.
func (x *XInputState) ToInputState() InputState {
	return InputState{
		Buttons: x.Buttons,
		LT:      x.LT,
		RT:      x.RT,
		LX:      x.LX,
		LY:      x.LY,
		RX:      x.RX,
		RY:      x.RY,
	}
}

// XRumbleState is the wire format for rumble/motor commands sent from device to client.
// Total size: 2 bytes (fixed).
// Layout:
//
//	LeftMotor: 1 byte (0-255)
//	RightMotor: 1 byte (0-255)
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
