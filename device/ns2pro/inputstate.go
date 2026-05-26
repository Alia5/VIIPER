package ns2pro

import (
	"encoding/binary"
	"io"
)

// nolint
// viiper:wire ns2pro c2s buttons:u32 lx:u16 ly:u16 rx:u16 ry:u16 accelX:i16 accelY:i16 accelZ:i16 gyroX:i16 gyroY:i16 gyroZ:i16 batteryLevel:u8 charging:bool externalPower:bool
type InputState struct {
	Buttons uint32

	LX, LY uint16
	RX, RY uint16

	AccelX, AccelY, AccelZ int16
	GyroX, GyroY, GyroZ    int16

	BatteryLevel  uint8
	Charging      bool
	ExternalPower bool
}

func defaultInputState() *InputState {
	return &InputState{
		LX:            StickCenter,
		LY:            StickCenter,
		RX:            StickCenter,
		RY:            StickCenter,
		BatteryLevel:  BatteryMax,
		ExternalPower: true,
	}
}

func (s *InputState) MarshalBinary() ([]byte, error) {
	b := make([]byte, InputWireSize)
	binary.LittleEndian.PutUint32(b[0:4], s.Buttons)
	binary.LittleEndian.PutUint16(b[4:6], s.LX)
	binary.LittleEndian.PutUint16(b[6:8], s.LY)
	binary.LittleEndian.PutUint16(b[8:10], s.RX)
	binary.LittleEndian.PutUint16(b[10:12], s.RY)
	binary.LittleEndian.PutUint16(b[12:14], uint16(s.AccelX))
	binary.LittleEndian.PutUint16(b[14:16], uint16(s.AccelY))
	binary.LittleEndian.PutUint16(b[16:18], uint16(s.AccelZ))
	binary.LittleEndian.PutUint16(b[18:20], uint16(s.GyroX))
	binary.LittleEndian.PutUint16(b[20:22], uint16(s.GyroY))
	binary.LittleEndian.PutUint16(b[22:24], uint16(s.GyroZ))
	b[24] = s.BatteryLevel
	if s.Charging {
		b[25] = 1
	}
	if s.ExternalPower {
		b[26] = 1
	}
	return b, nil
}

func (s *InputState) UnmarshalBinary(data []byte) error {
	if len(data) < InputWireSize {
		return io.ErrUnexpectedEOF
	}
	s.Buttons = binary.LittleEndian.Uint32(data[0:4])
	s.LX = binary.LittleEndian.Uint16(data[4:6])
	s.LY = binary.LittleEndian.Uint16(data[6:8])
	s.RX = binary.LittleEndian.Uint16(data[8:10])
	s.RY = binary.LittleEndian.Uint16(data[10:12])
	s.AccelX = int16(binary.LittleEndian.Uint16(data[12:14]))
	s.AccelY = int16(binary.LittleEndian.Uint16(data[14:16]))
	s.AccelZ = int16(binary.LittleEndian.Uint16(data[16:18]))
	s.GyroX = int16(binary.LittleEndian.Uint16(data[18:20]))
	s.GyroY = int16(binary.LittleEndian.Uint16(data[20:22]))
	s.GyroZ = int16(binary.LittleEndian.Uint16(data[22:24]))
	s.BatteryLevel = data[24]
	s.Charging = data[25] != 0
	s.ExternalPower = data[26] != 0
	return nil
}

// nolint
// viiper:wire ns2pro s2c leftRumble:u8*16 rightRumble:u8*16 flags:u8 playerLedMask:u8
type OutputState struct {
	LeftRumble    [16]byte
	RightRumble   [16]byte
	Flags         uint8
	PlayerLedMask uint8
}

func (o *OutputState) MarshalBinary() ([]byte, error) {
	b := make([]byte, OutputWireSize)
	copy(b[0:16], o.LeftRumble[:])
	copy(b[16:32], o.RightRumble[:])
	b[32] = o.Flags
	b[33] = o.PlayerLedMask
	return b, nil
}

func (o *OutputState) UnmarshalBinary(data []byte) error {
	if len(data) < OutputWireSize {
		return io.ErrUnexpectedEOF
	}
	copy(o.LeftRumble[:], data[0:16])
	copy(o.RightRumble[:], data[16:32])
	o.Flags = data[32]
	o.PlayerLedMask = data[33]
	return nil
}

func (s InputState) buildCommonReport(counter, motionTimestamp uint32, features uint8, batteryVolts uint16) []byte {
	b := make([]byte, InputReportSize)
	b[0] = ReportIDCommon
	binary.LittleEndian.PutUint32(b[1:5], counter)

	buttons := s.commonButtonBytes()
	copy(b[5:9], buttons[:])
	packStick12(b[11:14], s.LX, s.LY)
	packStick12(b[14:17], s.RX, s.RY)

	binary.LittleEndian.PutUint16(b[0x20:0x22], batteryVolts)
	b[0x22] = chargingState(s)
	b[0x2A] = 0x01

	if features&FeatureIMU != 0 {
		binary.LittleEndian.PutUint32(b[0x2B:0x2F], motionTimestamp)
		binary.LittleEndian.PutUint16(b[0x31:0x33], uint16(s.AccelX))
		binary.LittleEndian.PutUint16(b[0x33:0x35], uint16(s.AccelY))
		binary.LittleEndian.PutUint16(b[0x35:0x37], uint16(s.AccelZ))
		binary.LittleEndian.PutUint16(b[0x37:0x39], uint16(s.GyroX))
		binary.LittleEndian.PutUint16(b[0x39:0x3B], uint16(s.GyroY))
		binary.LittleEndian.PutUint16(b[0x3B:0x3D], uint16(s.GyroZ))
	}

	return b
}

func (s InputState) buildProReport(counter uint8, features uint8) []byte {
	b := make([]byte, InputReportSize)
	b[0] = ReportIDPro
	b[1] = counter
	b[2] = powerInfo(s)

	buttons := s.proButtonBytes()
	copy(b[3:6], buttons[:])
	packStick12(b[6:9], s.LX, s.LY)
	packStick12(b[9:12], s.RX, s.RY)

	if features&FeatureRumble != 0 {
		b[12] = 0x38
	} else {
		b[12] = 0x30
	}
	b[13] = 0x00
	b[14] = 0x00
	b[15] = 0x00
	return b
}

func (s InputState) commonButtonBytes() [4]byte {
	var out [4]byte
	encodeButtonMap(s.Buttons, commonButtonMap, out[:])
	return out
}

func (s InputState) proButtonBytes() [3]byte {
	var out [3]byte
	encodeButtonMap(s.Buttons, proButtonMap, out[:])
	return out
}

type buttonReportBit struct {
	button uint32
	index  int
	mask   byte
}

func encodeButtonMap(buttons uint32, mapping []buttonReportBit, out []byte) {
	for _, bit := range mapping {
		if buttons&bit.button != 0 {
			out[bit.index] |= bit.mask
		}
	}
}

var commonButtonMap = []buttonReportBit{
	{ButtonY, 0, 0x01},
	{ButtonX, 0, 0x02},
	{ButtonB, 0, 0x04},
	{ButtonA, 0, 0x08},
	{ButtonR, 0, 0x40},
	{ButtonZR, 0, 0x80},
	{ButtonMinus, 1, 0x01},
	{ButtonPlus, 1, 0x02},
	{ButtonRightStick, 1, 0x04},
	{ButtonLeftStick, 1, 0x08},
	{ButtonHome, 1, 0x10},
	{ButtonCapture, 1, 0x20},
	{ButtonC, 1, 0x40},
	{ButtonDown, 2, 0x01},
	{ButtonUp, 2, 0x02},
	{ButtonRight, 2, 0x04},
	{ButtonLeft, 2, 0x08},
	{ButtonL, 2, 0x40},
	{ButtonZL, 2, 0x80},
	{ButtonGR, 3, 0x01},
	{ButtonGL, 3, 0x02},
	{ButtonHeadset, 3, 0x10},
}

var proButtonMap = []buttonReportBit{
	{ButtonB, 0, 0x01},
	{ButtonA, 0, 0x02},
	{ButtonY, 0, 0x04},
	{ButtonX, 0, 0x08},
	{ButtonR, 0, 0x10},
	{ButtonZR, 0, 0x20},
	{ButtonPlus, 0, 0x40},
	{ButtonRightStick, 0, 0x80},
	{ButtonDown, 1, 0x01},
	{ButtonRight, 1, 0x02},
	{ButtonLeft, 1, 0x04},
	{ButtonUp, 1, 0x08},
	{ButtonL, 1, 0x10},
	{ButtonZL, 1, 0x20},
	{ButtonMinus, 1, 0x40},
	{ButtonLeftStick, 1, 0x80},
	{ButtonHome, 2, 0x01},
	{ButtonCapture, 2, 0x02},
	{ButtonGR, 2, 0x04},
	{ButtonGL, 2, 0x08},
	{ButtonC, 2, 0x10},
}

func packStick12(out []byte, x, y uint16) {
	if len(out) < 3 {
		return
	}
	x = clampStick(x)
	y = clampStick(y)
	out[0] = byte(x)
	out[1] = byte((x>>8)&0x0F) | byte((y&0x0F)<<4)
	out[2] = byte(y >> 4)
}

func clampStick(v uint16) uint16 {
	if v > StickMax {
		return StickMax
	}
	return v
}

func powerInfo(s InputState) uint8 {
	level := s.BatteryLevel
	if level > BatteryMax {
		level = BatteryMax
	}
	out := (level & 0x0F) << 2
	if s.ExternalPower {
		out |= 0x01
	}
	if s.Charging {
		out |= 0x02
	}
	return out
}

func chargingState(s InputState) uint8 {
	if s.Charging {
		return 0x34
	}
	return 0x20
}
