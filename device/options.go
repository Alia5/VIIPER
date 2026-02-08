package device

type CreateOptions struct {
	IdVendor       *uint16
	IdProduct      *uint16
	DeviceSpecific map[string]any
}
