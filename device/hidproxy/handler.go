package hidproxy

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/usb"
)

func init() {
	api.RegisterDevice("hidproxy", &handler{})
}

type handler struct{}

func (h *handler) CreateDevice(o *device.CreateOptions) (usb.Device, error) {
	if o == nil {
		o = &device.CreateOptions{}
	}
	options, err := decodeOptions(o.DeviceSpecific)
	if err != nil {
		return nil, err
	}
	p, err := New(options)
	if err != nil {
		return nil, err
	}
	// The copy keeps the copied device's ids; a create request may still set its own.
	if o.IDVendor != nil {
		p.desc.Device.IDVendor = *o.IDVendor
	}
	if o.IDProduct != nil {
		p.desc.Device.IDProduct = *o.IDProduct
	}
	return p, nil
}

func (h *handler) StreamHandler() api.StreamHandlerFunc {
	return func(conn net.Conn, devPtr *usb.Device, logger *slog.Logger) error {
		if devPtr == nil || *devPtr == nil {
			return fmt.Errorf("nil device")
		}
		p, ok := (*devPtr).(*HIDProxy)
		if !ok {
			return fmt.Errorf("%w: expected hidproxy", device.ErrWrongDeviceType)
		}
		err := p.Serve(conn)
		if errors.Is(err, io.EOF) {
			logger.Info("client disconnected")
			return nil
		}
		return err
	}
}

func (h *handler) UpdateMetaState(meta string, dev *usb.Device) error {
	return fmt.Errorf("hidproxy has no meta state")
}
