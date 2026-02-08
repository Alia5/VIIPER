// Package mouse provides a HID mouse device implementation.
package mouse

import (
	"sync"
	"sync/atomic"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
	"github.com/Alia5/VIIPER/usbip"
)

// Mouse implements the minimal Device interface for a 5-button HID mouse
// with vertical and horizontal wheels.
type Mouse struct {
	tick       uint64
	inputState *InputState
	stateMu    sync.Mutex
	descriptor usb.Descriptor
}

// New returns a new Mouse device.
func New(o *device.CreateOptions) (*Mouse, error) {
	d := &Mouse{
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
	return d, nil
}

// UpdateInputState updates the device's current input state (thread-safe).
func (m *Mouse) UpdateInputState(state InputState) {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()
	m.inputState = &state
}

// HandleTransfer implements interrupt IN for Mouse.
func (m *Mouse) HandleTransfer(ep uint32, dir uint32, out []byte) []byte {
	if dir == usbip.DirIn {
		switch ep {
		case 1: // 0x81 - main input reports
			atomic.AddUint64(&m.tick, 1)

			m.stateMu.Lock()
			var st InputState
			if m.inputState != nil {
				// Snapshot current state
				st = *m.inputState
				// Consume relative deltas so they are one-shot per poll cycle.
				// Buttons persist until explicitly changed by the client.
				m.inputState.DX = 0
				m.inputState.DY = 0
				m.inputState.Wheel = 0
				m.inputState.Pan = 0
			}
			m.stateMu.Unlock()
			return st.BuildReport()
		default:
			return nil
		}
	}
	return nil
}

// HID Report Descriptor for a 5-button mouse with vertical and horizontal wheels.
// Boot protocol compatible.
var reportDescriptor = hid.Report{
	Items: []hid.Item{
		hid.UsagePage{Page: hid.UsagePageGenericDesktop},
		hid.Usage{Usage: hid.UsageMouse},
		hid.Collection{Kind: hid.CollectionApplication, Items: []hid.Item{
			hid.Usage{Usage: hid.UsagePointer},
			hid.Collection{
				Kind: hid.CollectionPhysical,
				Items: []hid.Item{
					hid.UsagePage{Page: hid.UsagePageButton},
					hid.UsageMinimum{Min: 0x01}, // Button 1
					hid.UsageMaximum{Max: 0x05}, // Button 5
					hid.LogicalMinimum{Min: 0},
					hid.LogicalMaximum{Max: 1},
					hid.ReportCount{Count: 5},
					hid.ReportSize{Bits: 1},
					hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
					hid.ReportCount{Count: 1},
					hid.ReportSize{Bits: 3},
					hid.Input{Flags: hid.MainConst},
					hid.UsagePage{Page: hid.UsagePageGenericDesktop},
					hid.Usage{Usage: hid.UsageX},
					hid.Usage{Usage: hid.UsageY},
					hid.LogicalMinimum{Min: -32768},
					hid.LogicalMaximum{Max: 32767},
					hid.ReportSize{Bits: 16},
					hid.ReportCount{Count: 2},
					hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainRel},
					hid.Usage{Usage: hid.UsageWheel},
					hid.LogicalMinimum{Min: -32768},
					hid.LogicalMaximum{Max: 32767},
					hid.ReportSize{Bits: 16},
					hid.ReportCount{Count: 1},
					hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainRel},
					hid.UsagePage{Page: hid.UsagePageConsumer},
					hid.Usage{Usage: hid.UsageACPan},
					hid.LogicalMinimum{Min: -32768},
					hid.LogicalMaximum{Max: 32767},
					hid.ReportSize{Bits: 16},
					hid.ReportCount{Count: 1},
					hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainRel},
				},
			},
		}},
	},
}

// Descriptor defines the static USB descriptor for the mouse.
var defaultDescriptor = usb.Descriptor{
	Device: usb.DeviceDescriptor{
		BcdUSB:             0x0200,
		BDeviceClass:       0x00,
		BDeviceSubClass:    0x00,
		BDeviceProtocol:    0x00,
		BMaxPacketSize0:    0x40, // 64 bytes
		IDVendor:           0x2E8A,
		IDProduct:          0x0011,
		BcdDevice:          0x0100,
		IManufacturer:      0x01,
		IProduct:           0x02,
		ISerialNumber:      0x03,
		BNumConfigurations: 0x01,
		Speed:              2, // Full speed
	},
	Interfaces: []usb.InterfaceConfig{
		{
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x00,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x01,
				BInterfaceClass:    0x03, // HID
				BInterfaceSubClass: 0x01, // Boot Interface
				BInterfaceProtocol: 0x02, // Mouse
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
				Report: reportDescriptor,
			},
			Endpoints: []usb.EndpointDescriptor{
				{
					BEndpointAddress: 0x81,
					BMAttributes:     0x03,   // Interrupt
					WMaxPacketSize:   0x0010, // 16 bytes (9 needed)
					BInterval:        0x05,   // 5 ms
				},
			},
		},
	},
	Strings: map[uint8]string{
		0: "\x04\x09", // LangID: en-US (0x0409)
		1: "VIIPER",
		2: "HID Mouse",
		3: "1337",
	},
}

func (m *Mouse) GetDescriptor() *usb.Descriptor {
	return &m.descriptor
}

func (x *Mouse) GetDeviceSpecificArgs() map[string]any {
	return map[string]any{}
}
