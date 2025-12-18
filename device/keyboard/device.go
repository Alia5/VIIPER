// Package keyboard provides a HID keyboard device implementation with full N-key rollover.
package keyboard

import (
	"sync"
	"sync/atomic"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usbip"
)

// Keyboard implements the Device interface for a full HID keyboard with LED support.
type Keyboard struct {
	tick        uint64
	inputState  *InputState
	stateMu     sync.Mutex
	ledState    uint8
	ledCallback func(LEDState)
	descriptor  usb.Descriptor
}

// New returns a new Keyboard device.
func New(o *device.CreateOptions) *Keyboard {
	d := &Keyboard{
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

// SetLEDCallback sets a callback that will be invoked when LED state changes.
func (k *Keyboard) SetLEDCallback(f func(LEDState)) {
	k.ledCallback = f
}

// GetLEDState returns the current LED state from the host.
func (k *Keyboard) GetLEDState() LEDState {
	k.stateMu.Lock()
	defer k.stateMu.Unlock()
	return LEDState{
		NumLock:    k.ledState&LEDNumLock != 0,
		CapsLock:   k.ledState&LEDCapsLock != 0,
		ScrollLock: k.ledState&LEDScrollLock != 0,
		Compose:    k.ledState&LEDCompose != 0,
		Kana:       k.ledState&LEDKana != 0,
	}
}

// UpdateInputState updates the device's current input state (thread-safe).
func (k *Keyboard) UpdateInputState(state InputState) {
	k.stateMu.Lock()
	defer k.stateMu.Unlock()
	k.inputState = &state
}

// HandleTransfer implements interrupt IN/OUT for Keyboard.
func (k *Keyboard) HandleTransfer(ep uint32, dir uint32, out []byte) []byte {
	if dir == usbip.DirIn {
		switch ep {
		case 1: // 0x81 - keyboard input reports
			atomic.AddUint64(&k.tick, 1)

			k.stateMu.Lock()
			var st InputState
			if k.inputState != nil {
				st = *k.inputState
			}
			k.stateMu.Unlock()
			return st.BuildReport()
		default:
			return nil
		}
	}
	if dir == usbip.DirOut && ep == 1 {
		// 0x01 - LED state from host
		if len(out) >= 1 {
			k.stateMu.Lock()
			k.ledState = out[0]
			k.stateMu.Unlock()

			if k.ledCallback != nil {
				k.ledCallback(LEDState{
					NumLock:    out[0]&LEDNumLock != 0,
					CapsLock:   out[0]&LEDCapsLock != 0,
					ScrollLock: out[0]&LEDScrollLock != 0,
					Compose:    out[0]&LEDCompose != 0,
					Kana:       out[0]&LEDKana != 0,
				})
			}
		}
	}
	return nil
}

// HID Report Descriptor for a full keyboard with 256-bit key bitmap and LED output.
var hidReportDescriptor = []byte{
	0x05, 0x01, // Usage Page (Generic Desktop)
	0x09, 0x06, // Usage (Keyboard)
	0xA1, 0x01, // Collection (Application)

	// Input Report: Modifiers (1 byte)
	0x05, 0x07, //   Usage Page (Keyboard/Keypad)
	0x19, 0xE0, //   Usage Minimum (Left Control)
	0x29, 0xE7, //   Usage Maximum (Right GUI)
	0x15, 0x00, //   Logical Minimum (0)
	0x25, 0x01, //   Logical Maximum (1)
	0x75, 0x01, //   Report Size (1)
	0x95, 0x08, //   Report Count (8)
	0x81, 0x02, //   Input (Data, Variable, Absolute) - Modifier byte

	// Input Report: Reserved byte (1 byte)
	0x75, 0x08, //   Report Size (8)
	0x95, 0x01, //   Report Count (1)
	0x81, 0x01, //   Input (Constant) - Reserved byte

	// Input Report: Key array bitmap (32 bytes = 256 bits)
	0x05, 0x07, //   Usage Page (Keyboard/Keypad)
	0x19, 0x00, //   Usage Minimum (0x00)
	0x29, 0xFF, //   Usage Maximum (0xFF)
	0x15, 0x00, //   Logical Minimum (0)
	0x25, 0x01, //   Logical Maximum (1)
	0x75, 0x01, //   Report Size (1)
	0x96, 0x00, 0x01, // Report Count (256) - long item (0x96 for 2-byte count)
	0x81, 0x02, //   Input (Data, Variable, Absolute) - Key bitmap

	// Output Report: LEDs (1 byte)
	0x05, 0x08, //   Usage Page (LEDs)
	0x19, 0x01, //   Usage Minimum (Num Lock)
	0x29, 0x05, //   Usage Maximum (Kana)
	0x15, 0x00, //   Logical Minimum (0)
	0x25, 0x01, //   Logical Maximum (1)
	0x75, 0x01, //   Report Size (1)
	0x95, 0x05, //   Report Count (5)
	0x91, 0x02, //   Output (Data, Variable, Absolute) - LED bits
	0x75, 0x03, //   Report Size (3)
	0x95, 0x01, //   Report Count (1)
	0x91, 0x01, //   Output (Constant) - LED padding

	0xC0, // End Collection
}

// Descriptor defines the static USB descriptor for the keyboard.
var defaultDescriptor = usb.Descriptor{
	Device: usb.DeviceDescriptor{
		BcdUSB:             0x0200,
		BDeviceClass:       0x00,
		BDeviceSubClass:    0x00,
		BDeviceProtocol:    0x00,
		BMaxPacketSize0:    0x40, // 64 bytes
		IDVendor:           0x2E8A,
		IDProduct:          0x0010,
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
				BNumEndpoints:      0x02,
				BInterfaceClass:    0x03, // HID
				BInterfaceSubClass: 0x00, // No Subclass
				BInterfaceProtocol: 0x00, // None
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
					BMAttributes:     0x03, // Interrupt
					WMaxPacketSize:   0x0040,
					BInterval:        0x0A, // 10 ms
				},
				{
					BEndpointAddress: 0x01,
					BMAttributes:     0x03, // Interrupt
					WMaxPacketSize:   0x0008,
					BInterval:        0x0A, // 10 ms
				},
			},
		},
	},
	Strings: map[uint8]string{
		0: "\x04\x09", // LangID: en-US (0x0409)
		1: "VIIPER",
		2: "HID Keyboard",
		3: "1337",
	},
}

func (k *Keyboard) GetDescriptor() *usb.Descriptor {
	return &k.descriptor
}
