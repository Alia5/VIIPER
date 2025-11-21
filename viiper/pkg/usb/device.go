package usb

// Device is the minimal interface a device must implement.
// It only handles non-EP0 (interrupt/bulk) transfers.
type Device interface {
	// HandleTransfer processes a non-EP0 transfer (interrupt/bulk).
	// ep is the endpoint number (without direction). dir is protocol.DirIn or protocol.DirOut.
	// For IN transfers, return the payload to send; for OUT, consume 'out' and return nil.
	HandleTransfer(ep uint32, dir uint32, out []byte) []byte
	GetDescriptor() *Descriptor
}
