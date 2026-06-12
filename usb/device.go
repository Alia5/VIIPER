package usb

import "context"

// Device is the minimal interface a device must implement.
// It only handles non-EP0 (interrupt/bulk) transfers.
type Device interface {
	// HandleTransfer processes a non-EP0 transfer (interrupt/bulk).
	// ep is the endpoint number (without direction). dir is usbip.DirIn or usbip.DirOut.
	// For IN transfers the implementation should block until data is available or ctx is
	// cancelled, then return the payload. For OUT transfers, consume 'out' and return nil.
	HandleTransfer(ctx context.Context, ep uint32, dir uint32, out []byte) []byte
	GetDescriptor() *Descriptor
	GetDeviceSpecificArgs() map[string]any
}

// ControlDevice is an optional interface for devices that need to handle
// control transfers on endpoint 0 (EP0).
//
// This is primarily used for class-specific requests that are not covered by
// the server's built-in standard request handling (e.g. HID GET_REPORT/
// SET_REPORT).
type ControlDevice interface {
	// HandleControl handles a control request.
	//
	// - bmRequestType, bRequest, wValue, wIndex, wLength are the raw setup packet fields.
	// - data is the OUT data stage payload (for host-to-device requests), and is nil for
	//   device-to-host requests.
	//
	// If handled is false, the server will fall back to its default behavior.
	// If handled is true, the returned bytes (if any) will be used as the IN data stage.
	HandleControl(bmRequestType, bRequest uint8, wValue, wIndex, wLength uint16, data []byte) (resp []byte, handled bool)
}
