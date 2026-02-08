package dualshock4

import (
	"encoding/binary"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
	"github.com/Alia5/VIIPER/usbip"
)

type DualShock4 struct {
	inputState *InputState
	stateMu    sync.Mutex
	outputFunc func(OutputState)
	descriptor usb.Descriptor

	usbReportTimestamp uint32
	usbPacketCounter   uint32
}

func New(o *device.CreateOptions) (*DualShock4, error) {
	d := &DualShock4{
		descriptor: defaultDescriptor,
	}
	if o != nil {
		if o.IdVendor != nil {
			d.descriptor.Device.IDVendor = *o.IdVendor
		}
		if o.IdProduct != nil {
			d.descriptor.Device.IDProduct = *o.IdProduct
		}
	}

	d.inputState = &InputState{
		LX:           0,
		LY:           0,
		RX:           0,
		RY:           0,
		Buttons:      0,
		DPad:         0,
		L2:           0,
		R2:           0,
		Touch1X:      0,
		Touch1Y:      0,
		Touch1Active: false,
		Touch2X:      0,
		Touch2Y:      0,
		Touch2Active: false,
		GyroX:        0,
		GyroY:        0,
		GyroZ:        0,
		AccelX:       DefaultAccelXRaw,
		AccelY:       DefaultAccelYRaw,
		AccelZ:       DefaultAccelZRaw,
	}

	return d, nil
}

func (d *DualShock4) SetOutputCallback(f func(OutputState)) {
	d.outputFunc = f
}

func (d *DualShock4) UpdateInputState(state *InputState) {
	d.stateMu.Lock()
	defer d.stateMu.Unlock()
	d.inputState = state
}

func (d *DualShock4) HandleTransfer(ep uint32, dir uint32, out []byte) []byte {
	if dir == usbip.DirIn {
		switch ep {
		case 4:
			d.stateMu.Lock()
			st := *d.inputState
			d.stateMu.Unlock()
			return d.buildUSBInputReport(st)
		default:
			return nil
		}
	}

	if dir == usbip.DirOut && ep == 3 {
		if len(out) >= 11 && out[OutOffsetReportID] == ReportIDOutput {
			feedback := OutputState{
				RumbleSmall: out[OutOffsetRumbleSmall],
				RumbleLarge: out[OutOffsetRumbleLarge],
				LedRed:      out[OutOffsetLedRed],
				LedGreen:    out[OutOffsetLedGreen],
				LedBlue:     out[OutOffsetLedBlue],
				FlashOn:     out[OutOffsetFlashOn],
				FlashOff:    out[OutOffsetFlashOff],
			}
			if d.outputFunc != nil {
				d.outputFunc(feedback)
			}
		}
	}

	return nil
}

func (d *DualShock4) HandleControl(bmRequestType, bRequest uint8, wValue, _ /* wIndex */, wLength uint16, data []byte) ([]byte, bool) {
	const (
		hidGetReport = 0x01
		hidSetReport = 0x09
	)

	const (
		reportTypeInput   = 0x01
		reportTypeOutput  = 0x02
		reportTypeFeature = 0x03
	)

	reportType := uint8(wValue >> 8)
	reportID := uint8(wValue & 0xFF)

	if bmRequestType == 0xA1 && bRequest == hidGetReport {
		if reportType == reportTypeInput && reportID == ReportIDInput {
			d.stateMu.Lock()
			st := *d.inputState
			d.stateMu.Unlock()
			report := d.buildUSBInputReport(st)
			if wLength > 0 && int(wLength) < len(report) {
				return report[:wLength], true
			}
			return report, true
		}

		if reportType == reportTypeFeature {
			switch reportID {
			case 0x02: // Gyro calibration
				return make([]byte, 37), true
			case 0x03: // Device capabilities
				return make([]byte, 48), true
			case 0x05: // Gyro calibration
				return make([]byte, 41), true
			case 0x12: // Serial number
				return make([]byte, 16), true
			}
		}
	}

	if bmRequestType == 0x21 && bRequest == hidSetReport {
		if reportType == reportTypeOutput && reportID == ReportIDOutput && len(data) >= 11 {
			feedback := OutputState{
				RumbleSmall: data[OutOffsetRumbleSmall],
				RumbleLarge: data[OutOffsetRumbleLarge],
				LedRed:      data[OutOffsetLedRed],
				LedGreen:    data[OutOffsetLedGreen],
				LedBlue:     data[OutOffsetLedBlue],
				FlashOn:     data[OutOffsetFlashOn],
				FlashOff:    data[OutOffsetFlashOff],
			}
			if d.outputFunc != nil {
				d.outputFunc(feedback)
			}
			return nil, true
		}
	}

	slog.Warn("Unsupported control request",
		"bmRequestType", bmRequestType,
		"bRequest", bRequest)

	return nil, false
}

func (d *DualShock4) GetDescriptor() *usb.Descriptor {
	return &d.descriptor
}

func (x *DualShock4) GetDeviceSpecificArgs() map[string]any {
	return map[string]any{}
}

func (d *DualShock4) buildUSBInputReport(s InputState) []byte {
	b := make([]byte, InputReportSize)

	b[0] = ReportIDInput

	b[1] = uint8(int16(s.LX) + 128)
	b[2] = uint8(int16(s.LY) + 128)
	b[3] = uint8(int16(s.RX) + 128)
	b[4] = uint8(int16(s.RY) + 128)

	usbDPad := uint8(DPadUSBNeutral)
	if s.DPad&DPadUp != 0 && s.DPad&DPadRight != 0 {
		usbDPad = DPadUSBUpRight
	} else if s.DPad&DPadUp != 0 && s.DPad&DPadLeft != 0 {
		usbDPad = DPadUSBUpLeft
	} else if s.DPad&DPadDown != 0 && s.DPad&DPadRight != 0 {
		usbDPad = DPadUSBDownRight
	} else if s.DPad&DPadDown != 0 && s.DPad&DPadLeft != 0 {
		usbDPad = DPadUSBDownLeft
	} else if s.DPad&DPadUp != 0 {
		usbDPad = DPadUSBUp
	} else if s.DPad&DPadDown != 0 {
		usbDPad = DPadUSBDown
	} else if s.DPad&DPadLeft != 0 {
		usbDPad = DPadUSBLeft
	} else if s.DPad&DPadRight != 0 {
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

	ts := atomic.AddUint32(&d.usbReportTimestamp, 1)
	binary.LittleEndian.PutUint16(b[10:12], uint16(ts))

	b[12] = 0x00

	binary.LittleEndian.PutUint16(b[13:15], uint16(s.GyroX))
	binary.LittleEndian.PutUint16(b[15:17], uint16(s.GyroY))
	binary.LittleEndian.PutUint16(b[17:19], uint16(s.GyroZ))

	binary.LittleEndian.PutUint16(b[19:21], uint16(s.AccelX))
	binary.LittleEndian.PutUint16(b[21:23], uint16(s.AccelY))
	binary.LittleEndian.PutUint16(b[23:25], uint16(s.AccelZ))

	b[30] = BatteryFullyCharged

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

func encodeTouchCoords(b []byte, x, y uint16) {
	if x > TouchpadMaxX {
		x = TouchpadMaxX
	}
	if y > TouchpadMaxY {
		y = TouchpadMaxY
	}

	b[0] = uint8(x & 0xFF)
	b[1] = uint8((x>>8)&0x0F) | uint8((y&0x0F)<<4)
	b[2] = uint8(y >> 4)
}

var defaultDescriptor = usb.Descriptor{
	Device: usb.DeviceDescriptor{
		BcdUSB:             0x0200,
		BDeviceClass:       0x00,
		BDeviceSubClass:    0x00,
		BDeviceProtocol:    0x00,
		BMaxPacketSize0:    0x40,
		IDVendor:           DefaultVID,
		IDProduct:          DefaultPID,
		BcdDevice:          0x0100,
		IManufacturer:      0x01,
		IProduct:           0x02,
		ISerialNumber:      0x00,
		BNumConfigurations: 0x01,
		Speed:              2,
	},
	Interfaces: []usb.InterfaceConfig{
		{
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x00,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x02,
				BInterfaceClass:    0x03,
				BInterfaceSubClass: 0x00,
				BInterfaceProtocol: 0x00,
				IInterface:         0x00,
			},
			HID: &usb.HIDFunction{
				Descriptor: usb.HIDDescriptor{
					BcdHID:       0x0111,
					BCountryCode: 0x00,
					Descriptors: []usb.HIDSubDescriptor{
						{Type: usb.ReportDescType},
					},
				},
				Report: hid.Report{
					Items: []hid.Item{
						hid.UsagePage{Page: hid.UsagePageGenericDesktop},
						hid.Usage{Usage: hid.UsageGamePad},
						hid.Collection{Kind: hid.CollectionApplication, Items: []hid.Item{

							hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x08, Data: hid.Data{0x01}},

							hid.UsagePage{Page: hid.UsagePageGenericDesktop},
							hid.Usage{Usage: hid.UsageX},
							hid.Usage{Usage: hid.UsageY},
							hid.Usage{Usage: hid.UsageZ},
							hid.Usage{Usage: hid.UsageRz},
							hid.LogicalMinimum{Min: 0},
							hid.LogicalMaximum{Max: 255},
							hid.ReportSize{Bits: 8},
							hid.ReportCount{Count: 4},
							hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

							hid.UsagePage{Page: hid.UsagePageGenericDesktop},
							hid.Usage{Usage: 0x39},
							hid.LogicalMinimum{Min: 0},
							hid.LogicalMaximum{Max: 7},
							hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x3, Data: hid.Data{0x00}},
							hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x4, Data: hid.Data{0x3B, 0x01}},
							hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x6, Data: hid.Data{0x14}},
							hid.ReportSize{Bits: 4},
							hid.ReportCount{Count: 1},
							hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs | hid.MainNullState},
							hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x6, Data: hid.Data{0x00}},

							hid.UsagePage{Page: hid.UsagePageButton},
							hid.UsageMinimum{Min: 0x01},
							hid.UsageMaximum{Max: 0x0E},
							hid.LogicalMinimum{Min: 0},
							hid.LogicalMaximum{Max: 1},
							hid.ReportCount{Count: 14},
							hid.ReportSize{Bits: 1},
							hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

							hid.UsagePage{Page: 0xFF00},
							hid.Usage{Usage: 0x20},
							hid.ReportSize{Bits: 6},
							hid.ReportCount{Count: 1},
							hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

							hid.UsagePage{Page: hid.UsagePageGenericDesktop},
							hid.Usage{Usage: 0x32},
							hid.Usage{Usage: 0x35},
							hid.LogicalMinimum{Min: 0},
							hid.LogicalMaximum{Max: 255},
							hid.ReportSize{Bits: 8},
							hid.ReportCount{Count: 2},
							hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

							hid.UsagePage{Page: 0xFF00},
							hid.Usage{Usage: 0x20},
							hid.LogicalMinimum{Min: 0},
							hid.LogicalMaximum{Max: 255},
							hid.ReportSize{Bits: 8},
							hid.ReportCount{Count: 54},
							hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

							hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x08, Data: hid.Data{0x05}},

							hid.UsagePage{Page: 0xFF00},
							hid.Usage{Usage: 0x21},
							hid.LogicalMinimum{Min: 0},
							hid.LogicalMaximum{Max: 255},
							hid.ReportSize{Bits: 8},
							hid.ReportCount{Count: 31},
							hid.Output{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
						}},
					},
				},
			},
			Endpoints: []usb.EndpointDescriptor{
				{
					BEndpointAddress: EndpointIn,
					BMAttributes:     0x03,
					WMaxPacketSize:   64,
					BInterval:        5,
				},
				{
					BEndpointAddress: EndpointOut,
					BMAttributes:     0x03,
					WMaxPacketSize:   64,
					BInterval:        5,
				},
			},
		},
	},
	Strings: map[uint8]string{
		0: "\x04\x09",
		1: "Sony Interactive Entertainment",
		2: "Wireless Controller",
	},
}
