package device

type CreateOptions struct {
	IDVendor       *uint16
	IDProduct      *uint16
	DeviceSpecific string
}
