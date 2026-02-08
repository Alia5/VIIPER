// Package keyboard provides a HID keyboard device implementation with full N-key rollover.
package keyboard

import (
	"sync"
	"sync/atomic"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
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
func New(o *device.CreateOptions) (*Keyboard, error) {
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
	return d, nil
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
var reportDescriptor = hid.Report{
	Items: []hid.Item{
		hid.UsagePage{Page: hid.UsagePageGenericDesktop},
		hid.Usage{Usage: hid.UsageKeyboard},
		hid.Collection{
			Kind: hid.CollectionApplication,
			Items: []hid.Item{
				// Input Report: Modifiers (1 byte)
				hid.UsagePage{Page: hid.UsagePageKeyboard},
				hid.UsageMinimum{Min: 0xE0}, // Left Control
				hid.UsageMaximum{Max: 0xE7}, // Right GUI
				hid.LogicalMinimum{Min: 0},
				hid.LogicalMaximum{Max: 1},
				hid.ReportSize{Bits: 1},
				hid.ReportCount{Count: 8},
				hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

				// Input Report: Reserved byte (1 byte)
				hid.ReportSize{Bits: 8},
				hid.ReportCount{Count: 1},
				hid.Input{Flags: hid.MainConst},

				// Input Report: Key array bitmap (32 bytes = 256 bits)
				hid.UsagePage{Page: hid.UsagePageKeyboard},
				hid.UsageMinimum{Min: 0x00},
				hid.UsageMaximum{Max: 0xFF},
				hid.LogicalMinimum{Min: 0},
				hid.LogicalMaximum{Max: 1},
				hid.ReportSize{Bits: 1},
				hid.ReportCount{Count: 256},
				hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

				// Output Report: LEDs (1 byte)
				hid.UsagePage{Page: hid.UsagePageLEDs},
				hid.UsageMinimum{Min: 0x01}, // Num Lock
				hid.UsageMaximum{Max: 0x05}, // Kana
				hid.LogicalMinimum{Min: 0},
				hid.LogicalMaximum{Max: 1},
				hid.ReportSize{Bits: 1},
				hid.ReportCount{Count: 5},
				hid.Output{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
				hid.ReportSize{Bits: 3},
				hid.ReportCount{Count: 1},
				hid.Output{Flags: hid.MainConst},
			},
		},
	},
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
			HID: &usb.HIDFunction{
				Descriptor: usb.HIDDescriptor{
					BcdHID:       0x0111,
					BCountryCode: 0x00,
					Descriptors: []usb.HIDSubDescriptor{
						{Type: usb.ReportDescType}, // Length auto-filled from Report
					},
				},
				Report: reportDescriptor,
			},
			Endpoints: []usb.EndpointDescriptor{
				{
					BEndpointAddress: 0x81,
					BMAttributes:     0x03, // Interrupt
					WMaxPacketSize:   0x0040,
					BInterval:        0x05, // 5 ms
				},
				{
					BEndpointAddress: 0x01,
					BMAttributes:     0x03, // Interrupt
					WMaxPacketSize:   0x0008,
					BInterval:        0x05, // 5 ms
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

func (x *Keyboard) GetDeviceSpecificArgs() map[string]any {
	return map[string]any{}
}
