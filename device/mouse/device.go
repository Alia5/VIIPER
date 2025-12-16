// Package mouse provides a HID mouse device implementation.
package mouse

import (
	"sync"
	"sync/atomic"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
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
func New(o *device.CreateOptions) *Mouse {
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
	return d
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
var hidReportDescriptor = []byte{
	0x05, 0x01, // Usage Page (Generic Desktop)
	0x09, 0x02, // Usage (Mouse)
	0xA1, 0x01, // Collection (Application)
	0x09, 0x01, //   Usage (Pointer)
	0xA1, 0x00, //   Collection (Physical)
	0x05, 0x09, //     Usage Page (Button)
	0x19, 0x01, //     Usage Minimum (Button 1)
	0x29, 0x05, //     Usage Maximum (Button 5)
	0x15, 0x00, //     Logical Minimum (0)
	0x25, 0x01, //     Logical Maximum (1)
	0x95, 0x05, //     Report Count (5)
	0x75, 0x01, //     Report Size (1)
	0x81, 0x02, //     Input (Data, Variable, Absolute)
	0x95, 0x01, //     Report Count (1)
	0x75, 0x03, //     Report Size (3)
	0x81, 0x01, //     Input - padding
	0x05, 0x01, //     Usage Page (Generic Desktop)
	0x09, 0x30, //     Usage (X)
	0x09, 0x31, //     Usage (Y)
	0x16, 0x00, 0x80, // Logical Minimum (-32768)
	0x26, 0xFF, 0x7F, // Logical Maximum (32767)
	0x75, 0x10, //     Report Size (16)
	0x95, 0x02, //     Report Count (2)
	0x81, 0x06, //     Input (Data, Variable, Relative)
	0x09, 0x38, //     Usage (Wheel)
	0x16, 0x00, 0x80, // Logical Minimum (-32768)
	0x26, 0xFF, 0x7F, // Logical Maximum (32767)
	0x75, 0x10, //     Report Size (16)
	0x95, 0x01, //     Report Count (1)
	0x81, 0x06, //     Input (Data, Variable, Relative)
	0x05, 0x0C, //     Usage Page (Consumer)
	0x0A, 0x38, 0x02, // Usage (AC Pan)
	0x16, 0x00, 0x80, // Logical Minimum (-32768)
	0x26, 0xFF, 0x7F, // Logical Maximum (32767)
	0x75, 0x10, //     Report Size (16)
	0x95, 0x01, //     Report Count (1)
	0x81, 0x06, //     Input (Data, Variable, Relative)
	0xC0, //   End Collection
	0xC0, // End Collection
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
			HIDDescriptor: []byte{
				0x09,       // bLength
				0x21,       // bDescriptorType (HID)
				0x11, 0x01, // bcdHID 1.11
				0x00,                                 // bCountryCode
				0x01,                                 // bNumDescriptors
				0x22,                                 // bDescriptorType (Report)
				byte(len(hidReportDescriptor)), 0x00, // wDescriptorLength
			},
			HIDReport: hidReportDescriptor,
			Endpoints: []usb.EndpointDescriptor{
				{
					BEndpointAddress: 0x81,
					BMAttributes:     0x03,   // Interrupt
					WMaxPacketSize:   0x0010, // 16 bytes (9 needed)
					BInterval:        0x0A,   // 10 ms
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
