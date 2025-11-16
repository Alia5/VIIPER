// Package xbox360 provides an Xbox 360 controller device implementation.
package xbox360

import (
	"sync"
	"sync/atomic"

	"viiper/pkg/usb"
	"viiper/pkg/usbip"
	"viiper/pkg/virtualbus"
)

// Xbox360 implements only the minimal Device interface.
type Xbox360 struct {
	tick       uint64
	inputState *InputState
	stateMu    sync.Mutex
	rumbleFunc func(XRumbleState) // called when rumble commands arrive
}

// New returns a new Xbox360 device.
func New() *Xbox360 {
	return &Xbox360{}
}

// SetRumbleCallback sets a callback that will be invoked when rumble commands arrive.
func (x *Xbox360) SetRumbleCallback(f func(XRumbleState)) {
	x.rumbleFunc = f
}

// UpdateInputState updates the device's current input state (thread-safe).
func (x *Xbox360) UpdateInputState(state InputState) {
	x.stateMu.Lock()
	defer x.stateMu.Unlock()
	x.inputState = &state
}

// HandleTransfer implements interrupt IN/OUT for Xbox360.
func (x *Xbox360) HandleTransfer(ep uint32, dir uint32, out []byte) []byte {
	if dir == usbip.DirIn {
		switch ep {
		case 1: // 0x81 - main input reports
			atomic.AddUint64(&x.tick, 1)

			x.stateMu.Lock()
			var st InputState
			if x.inputState != nil {
				st = *x.inputState
			}
			x.stateMu.Unlock()
			return st.BuildReport()
		default:
			return nil
		}
	}
	if dir == usbip.DirOut && ep == 1 {
		if len(out) >= 8 {
			rumble := XRumbleState{
				LeftMotor:  out[3],
				RightMotor: out[4],
			}
			if x.rumbleFunc != nil {
				x.rumbleFunc(rumble)
			}
		}
	}
	return nil
}

// Static descriptor/config for Xbox360, for registration with the bus.
var Descriptor = virtualbus.DeviceDescriptor{
	Device: usb.DeviceDescriptor{
		BcdUSB:             0x0200,
		BDeviceClass:       0xff,
		BDeviceSubClass:    0xff,
		BDeviceProtocol:    0xff,
		BMaxPacketSize0:    0x08,
		IDVendor:           0x045e,
		IDProduct:          0x028e,
		BcdDevice:          0x0114,
		IManufacturer:      0x01,
		IProduct:           0x02,
		ISerialNumber:      0x03,
		BNumConfigurations: 0x01,
		Speed:              2, // Full speed
	},
	Interfaces: []virtualbus.InterfaceConfig{
		// Interface 0: ff/5d/01 with 2 interrupt endpoints
		{
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x00,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x02,
				BInterfaceClass:    0xff,
				BInterfaceSubClass: 0x5d,
				BInterfaceProtocol: 0x01,
				IInterface:         0x00,
			},
			HIDDescriptor: []byte{0x11, 0x21, 0x00, 0x01, 0x01, 0x25, 0x81, 0x14, 0x00, 0x00, 0x00, 0x00, 0x13, 0x01, 0x08, 0x00, 0x00},
			Endpoints: []usb.EndpointDescriptor{
				{BEndpointAddress: 0x81, BMAttributes: 0x03, WMaxPacketSize: 0x0020, BInterval: 0x04},
				{BEndpointAddress: 0x01, BMAttributes: 0x03, WMaxPacketSize: 0x0020, BInterval: 0x08},
			},
		},
		// Interface 1: ff/5d/03 with 4 interrupt endpoints
		{
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x01,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x04,
				BInterfaceClass:    0xff,
				BInterfaceSubClass: 0x5d,
				BInterfaceProtocol: 0x03,
				IInterface:         0x00,
			},
			HIDDescriptor: []byte{0x1b, 0x21, 0x00, 0x01, 0x01, 0x01, 0x82, 0x40, 0x01, 0x02, 0x20, 0x16, 0x83, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x16, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Endpoints: []usb.EndpointDescriptor{
				{BEndpointAddress: 0x82, BMAttributes: 0x03, WMaxPacketSize: 0x0020, BInterval: 0x02},
				{BEndpointAddress: 0x02, BMAttributes: 0x03, WMaxPacketSize: 0x0020, BInterval: 0x04},
				{BEndpointAddress: 0x83, BMAttributes: 0x03, WMaxPacketSize: 0x0020, BInterval: 0x40},
				{BEndpointAddress: 0x03, BMAttributes: 0x03, WMaxPacketSize: 0x0020, BInterval: 0x10},
			},
		},
		// Interface 2: ff/5d/02 with 1 interrupt endpoint
		{
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x02,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x01,
				BInterfaceClass:    0xff,
				BInterfaceSubClass: 0x5d,
				BInterfaceProtocol: 0x02,
				IInterface:         0x00,
			},
			HIDDescriptor: []byte{0x09, 0x21, 0x00, 0x01, 0x01, 0x22, 0x84, 0x07, 0x00},
			Endpoints: []usb.EndpointDescriptor{
				{BEndpointAddress: 0x84, BMAttributes: 0x03, WMaxPacketSize: 0x0020, BInterval: 0x10},
			},
		},
		// Interface 3: ff/fd/13 with vendor-specific descriptor
		{
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x03,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x00,
				BInterfaceClass:    0xff,
				BInterfaceSubClass: 0xfd,
				BInterfaceProtocol: 0x13,
				IInterface:         0x04,
			},
			VendorData: []byte{0x06, 0x41, 0x00, 0x01, 0x01, 0x03},
		},
	},
	Strings: map[uint8]string{
		0: "\x04\x09", // LangID: en-US (0x0409)
		1: "©Microsoft Corporation",
		2: "Controller",
		3: "296013F",
	},
}

func (x *Xbox360) GetDeviceDescriptor() virtualbus.DeviceDescriptor {
	return Descriptor
}
