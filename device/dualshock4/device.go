package dualshock4

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usbip"
)

type DualShock4 struct {
	inputState *InputState
	metaState  *MetaState

	outputFunc func(OutputState)
	descriptor usb.Descriptor

	probeSelector       [3]byte
	telemetrySubcommand byte

	usbPacketCounter uint32
	lastUSBReportAt  time.Time

	mtx sync.Mutex
}

func New(o *device.CreateOptions) (*DualShock4, error) {
	metaState := &MetaState{
		SerialNumber:       DefaultSerialString,
		Board:              DefaultBoardString,
		BuildTime:          DefaultBuildTime,
		BatteryStatus:      DefaultBatteryStatus,
		TemperatureCelsius: DefaultTemperature,
		BatteryVoltage:     DefaultVoltage,
	}
	if o != nil && o.DeviceSpecific != "" {
		var newMeta MetaState
		err := json.Unmarshal([]byte(o.DeviceSpecific), &newMeta)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON payload: %w", err)
		}
		if newMeta.SerialNumber != "" {
			metaState.SerialNumber = newMeta.SerialNumber
		}
		if newMeta.Board != "" {
			metaState.Board = newMeta.Board
		}
		if !newMeta.BuildTime.IsZero() {
			metaState.BuildTime = newMeta.BuildTime
		}
		if newMeta.BatteryStatus != 0 {
			metaState.BatteryStatus = newMeta.BatteryStatus
		}
		if newMeta.TemperatureCelsius != 0 {
			metaState.TemperatureCelsius = newMeta.TemperatureCelsius
		}
		if newMeta.BatteryVoltage != 0 {
			metaState.BatteryVoltage = newMeta.BatteryVoltage
		}
	}

	d := &DualShock4{
		descriptor: defaultDescriptor,
		metaState:  metaState,
	}
	if o != nil {
		if o.IDVendor != nil {
			d.descriptor.Device.IDVendor = *o.IDVendor
		}
		if o.IDProduct != nil {
			d.descriptor.Device.IDProduct = *o.IDProduct
		}
		if len(d.metaState.SerialNumber) > 0 && len(d.metaState.SerialNumber) <= 16 {
			d.metaState.SerialNumber = fmt.Sprintf("%016s", d.metaState.SerialNumber)
		}
	}

	slog.Info("DS4 device instantiated",
		"vid", d.descriptor.Device.IDVendor,
		"pid", d.descriptor.Device.IDProduct,
		"interfaces", len(d.descriptor.Interfaces))

	d.inputState = NewInputState()

	return d, nil
}

func (d *DualShock4) SetMetaState(meta MetaState) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.metaState = &meta
}

func (d *DualShock4) SetOutputCallback(f func(OutputState)) {
	d.outputFunc = f
}

func (d *DualShock4) UpdateInputState(state *InputState) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.inputState = state
}

func (d *DualShock4) GetDescriptor() *usb.Descriptor {
	return &d.descriptor
}

func (d *DualShock4) GetDeviceSpecificArgs() map[string]any {
	var res map[string]any
	d.mtx.Lock()
	defer d.mtx.Unlock()

	bytes, err := json.Marshal(d.metaState)
	if err != nil {
		return map[string]any{}
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return map[string]any{}
	}
	return res
}

func (d *DualShock4) HandleTransfer(ep uint32, dir uint32, out []byte) []byte {
	if dir == usbip.DirIn {
		switch ep {
		case 4:
			d.mtx.Lock()
			is := *d.inputState
			ms := *d.metaState
			d.mtx.Unlock()
			return d.buildUSBInputReport(&is, &ms)
		default:
			return nil
		}
	}

	if dir == usbip.DirOut && ep == 3 {
		if len(out) >= 11 && out[0] == ReportIDOutput {
			if d.outputFunc != nil {
				d.outputFunc(parseOutputReport(out))
			}
		}
	}

	return nil
}

func (d *DualShock4) HandleControl(bmRequestType, bRequest uint8, wValue, wIndex, wLength uint16, data []byte) ([]byte, bool) {
	reportType := uint8(wValue >> 8)
	reportID := uint8(wValue & 0xFF)

	switch bmRequestType {
	case hidClassIN:
		switch bRequest {
		case hidGetReport:
			if reportType == reportTypeInput && reportID == ReportIDInput {
				d.mtx.Lock()
				is := *d.inputState
				ms := *d.metaState
				d.mtx.Unlock()
				b := d.buildUSBInputReport(&is, &ms)
				if wLength > 0 && int(wLength) < len(b) {
					b = b[:wLength]
				}
				return b, true
			}
			if reportType == reportTypeFeature {
				if fn, ok := featureGetHandlers[reportID]; ok {
					b := fn(d)
					if wLength > 0 && int(wLength) < len(b) {
						b = b[:wLength]
					}
					return b, true
				}
			}
		case hidGetIdle:
			return []byte{0x00}, true
		case hidGetProtocol:
			return []byte{0x01}, true
		case 0x81:
			return []byte{0x00}, true
		case 0x82, 0x83, 0x84:
			return []byte{0x00, 0x00}, true
		}
	case hidClassOUT:
		if bRequest == hidSetReport {
			switch {
			case reportType == reportTypeFeature && reportID == featureIDSubcommand:
				if len(data) >= 2 {
					d.telemetrySubcommand = data[1]
				}
				return nil, true
			case reportType == reportTypeFeature && reportID == featureIDProbe:
				if len(data) >= 4 {
					d.probeSelector[0] = data[1]
					d.probeSelector[1] = data[2]
					d.probeSelector[2] = data[3]
				}
				return nil, true
			case reportType == reportTypeOutput && reportID == ReportIDOutput && len(data) >= 11:
				if d.outputFunc != nil {
					d.outputFunc(parseOutputReport(data))
				}
				return nil, true
			}
		}
	}

	slog.Warn("DS4 control request unhandled",
		"bmRequestType", bmRequestType,
		"bRequest", bRequest,
		"reportType", reportType,
		"reportID", reportID,
		"wIndex", wIndex,
		"wLength", wLength,
		"dataLen", len(data))

	return nil, false
}

// featureGetHandlers maps feature report IDs to their builder functions.
var featureGetHandlers = map[byte]func(*DualShock4) []byte{
	featureIDStatus:        (*DualShock4).featureReportStatus,
	featureIDProbeResponse: (*DualShock4).featureReportProbeResponse,
	featureIDCalibration:   (*DualShock4).featureReportCalibration,
	featureIDCalibrationBT: (*DualShock4).featureReportCalibrationBT,
	featureIDCapabilities:  (*DualShock4).featureReportCapabilities,
	featureIDSerial:        (*DualShock4).featureReportSerial,
	featureIDTelemetry:     (*DualShock4).featureReportTelemetry,
	featureIDIdentity:      (*DualShock4).featureReportIdentity,
	featureIDBoardInfo:     (*DualShock4).featureReportBoardInfo,
}

func parseOutputReport(data []byte) OutputState {
	return OutputState{
		RumbleSmall: data[4],
		RumbleLarge: data[5],
		LedRed:      data[6],
		LedGreen:    data[7],
		LedBlue:     data[8],
		FlashOn:     data[9],
		FlashOff:    data[10],
	}
}

func (d *DualShock4) featureReportTelemetry() []byte {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	s := serialStringToBytes(d.metaState.SerialNumber)
	switch d.telemetrySubcommand {
	case 0x02:
		return []byte{
			featureIDTelemetry,
			s[3], s[2], s[1], s[0], s[7], s[6], s[5], s[4],
			0x00, 0x00, 0x00, 0x00, 0x00,
		}
	case 0x0B:
		return []byte{
			featureIDTelemetry,
			s[3], s[2], s[1], s[0], s[7], s[6], s[5], s[4],
			0xAC, 0xA8, 0x1B,
			0x00, 0x00,
		}
	default:
		volts := telemetryVoltageU16(d.metaState.BatteryVoltage)
		temp := telemetryTemperatureU16(d.metaState.TemperatureCelsius)
		return []byte{
			featureIDTelemetry, d.telemetrySubcommand, 0x03, 0x01, 0x00, 0x04,
			byte(volts), byte(volts >> 8),
			byte(temp), byte(temp >> 8),
			0x00, 0x00, 0x00, 0x00,
		}
	}
}

func (d *DualShock4) featureReportIdentity() []byte {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	serial := serialStringToBytes(d.metaState.SerialNumber)
	firmware := ds4FirmwareVersionString()

	buildDateStr := d.metaState.BuildTime.Format("Jan 02 2006")

	report := make([]byte, 64)
	report[0] = featureIDIdentity
	copy(report[1:9], serial[:])
	copy(report[10:18], serial[:])
	copy(report[18:34], d.metaState.SerialNumber)
	copy(report[34:46], d.metaState.Board)
	copy(report[46:57], buildDateStr)
	copy(report[57:64], firmware[:7])
	return report
}

func (d *DualShock4) featureReportBoardInfo() []byte {
	report := make([]byte, 49)
	report[0] = featureIDBoardInfo

	d.mtx.Lock()
	buildDateStr := d.metaState.BuildTime.Format("Jan 02 2006")
	buildTimeStr := d.metaState.BuildTime.Format("15:04:05")
	d.mtx.Unlock()

	copy(report[1:16], buildDateStr)
	copy(report[16:32], buildTimeStr)
	binary.LittleEndian.PutUint16(report[33:35], HardwareVersionMajor)
	binary.LittleEndian.PutUint16(report[35:37], HardwareVersionMinor)
	binary.LittleEndian.PutUint32(report[37:41], SoftwareVersionMajor)
	binary.LittleEndian.PutUint16(report[41:43], SoftwareVersionMinor)

	report[47] = 1

	return report
}

func (d *DualShock4) featureReportSerial() []byte {
	d.mtx.Lock()
	serial := serialStringToBytes(d.metaState.SerialNumber)
	d.mtx.Unlock()

	report := make([]byte, 16)
	report[0] = featureIDSerial
	report[1] = serial[7]
	report[2] = serial[6]
	report[3] = serial[5]
	report[4] = serial[4]
	report[5] = serial[3]
	report[6] = serial[2]
	report[7] = serial[1]
	copy(report[8:16], serial[:])

	return report
}

func (d *DualShock4) featureReportStatus() []byte {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	report := make([]byte, 5)
	report[0] = featureIDStatus
	report[1] = d.metaState.BatteryStatus & BatteryLevelMask
	report[2] = 12
	binary.LittleEndian.PutUint16(report[3:5], 664)
	return report
}

func (d *DualShock4) featureReportProbeResponse() []byte {
	b1 := d.probeSelector[0]
	b2 := d.probeSelector[1]
	b3 := d.probeSelector[2]

	report := [4]byte{featureIDProbeResponse, b1, b2, b3}

	switch {
	case b1 == 0xFF && b2 == 0x00 && b3 == 0x0C:
		report[1] = 0x01
	}

	return report[:]
}

func (d *DualShock4) featureReportCapabilities() []byte {
	report := make([]byte, 48)
	report[0] = featureIDCapabilities
	report[2] = 0x27

	// Sensor + lightbar + vibration + touchpad capability bits.
	report[4] = 0x02 | 0x04 | 0x08 | 0x40
	report[5] = 0x00 // gamepad

	binary.LittleEndian.PutUint16(report[10:12], 1)
	binary.LittleEndian.PutUint16(report[12:14], 16)
	binary.LittleEndian.PutUint16(report[14:16], 1)
	binary.LittleEndian.PutUint16(report[16:18], 8192)

	return report
}

func (d *DualShock4) featureReportCalibration() []byte {
	return d.buildCalibrationReport(featureIDCalibration)
}

func (d *DualShock4) featureReportCalibrationBT() []byte {
	return d.buildCalibrationReport(featureIDCalibrationBT)
}

func (d *DualShock4) buildCalibrationReport(id byte) []byte {
	report := make([]byte, 37)
	report[0] = id

	// 17 LE int16 fields packed sequentially from offset 1:
	// bias(pitch,yaw,roll) | gyro±(x,y,z) | speed(x,y) | accel±(x,y,z)
	for i, v := range [17]int16{
		0, 0, 0,
		1024, -1024, 1024, -1024, 1024, -1024,
		64, 64,
		8192, -8192, 8192, -8192, 8192, -8192,
	} {
		binary.LittleEndian.PutUint16(report[1+i*2:], uint16(v))
	}

	return report
}

func (d *DualShock4) buildUSBInputReport(s *InputState, m *MetaState) []byte {
	b := make([]byte, InputReportSize)

	b[0] = ReportIDInput

	b[1] = uint8(int16(s.LX) + 128)
	b[2] = uint8(int16(s.LY) + 128)
	b[3] = uint8(int16(s.RX) + 128)
	b[4] = uint8(int16(s.RY) + 128)

	usbDPad := uint8(DPadUSBNeutral)
	switch {
	case s.DPad&DPadUp != 0 && s.DPad&DPadRight != 0:
		usbDPad = DPadUSBUpRight
	case s.DPad&DPadUp != 0 && s.DPad&DPadLeft != 0:
		usbDPad = DPadUSBUpLeft
	case s.DPad&DPadDown != 0 && s.DPad&DPadRight != 0:
		usbDPad = DPadUSBDownRight
	case s.DPad&DPadDown != 0 && s.DPad&DPadLeft != 0:
		usbDPad = DPadUSBDownLeft
	case s.DPad&DPadUp != 0:
		usbDPad = DPadUSBUp
	case s.DPad&DPadDown != 0:
		usbDPad = DPadUSBDown
	case s.DPad&DPadLeft != 0:
		usbDPad = DPadUSBLeft
	case s.DPad&DPadRight != 0:
		usbDPad = DPadUSBRight
	}

	b[5] = (usbDPad & DPadMask) | (uint8(s.Buttons) & 0xF0)
	b[6] = uint8(s.Buttons >> 8)

	counter := atomic.AddUint32(&d.usbPacketCounter, 1) & 0x3F

	psTouch := uint8(0)
	if s.Buttons&ButtonPS != 0 {
		psTouch |= ButtonPSUSB
	}
	if s.Buttons&ButtonTouchpadClick != 0 {
		psTouch |= ButtonTouchpadClickUSB
	}
	b[7] = psTouch | uint8(counter<<CounterShift)

	b[8] = s.L2
	b[9] = s.R2

	ts := d.nextReportTimestamp()
	binary.LittleEndian.PutUint16(b[10:12], uint16(ts))

	binary.LittleEndian.PutUint16(b[13:15], uint16(s.GyroX))
	binary.LittleEndian.PutUint16(b[15:17], uint16(s.GyroY))
	binary.LittleEndian.PutUint16(b[17:19], uint16(s.GyroZ))

	binary.LittleEndian.PutUint16(b[19:21], uint16(s.AccelX))
	binary.LittleEndian.PutUint16(b[21:23], uint16(s.AccelY))
	binary.LittleEndian.PutUint16(b[23:25], uint16(s.AccelZ))

	b[12] = 0x09            // status: touchpad connected, no extension
	b[30] = m.BatteryStatus // low nibble = level, bit4 = cable
	b[33] = 0x01            // nvslocked
	b[34] = 0x01

	touch1Counter := uint8(0)
	if !s.Touch1Active {
		touch1Counter |= TouchInactiveMask
	}
	b[35] = touch1Counter
	encodeTouchCoords(b[36:39], s.Touch1X, s.Touch1Y)

	touch2Counter := uint8(0)
	if !s.Touch2Active {
		touch2Counter |= TouchInactiveMask
	}
	b[39] = touch2Counter
	encodeTouchCoords(b[40:43], s.Touch2X, s.Touch2Y)

	return b
}

func (d *DualShock4) nextReportTimestamp() uint32 {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	now := time.Now()
	if d.lastUSBReportAt.IsZero() {
		d.lastUSBReportAt = now
		return 188
	}
	elapsed := now.Sub(d.lastUSBReportAt).Nanoseconds()
	d.lastUSBReportAt = now
	ts := uint32(elapsed * 3 / 16000)
	if ts == 0 {
		ts = 1
	}
	return ts
}
