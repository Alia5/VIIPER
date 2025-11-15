// Package virtualbus manages USB bus topology and auto-assigns device addresses.
package virtualbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"viiper/pkg/device"
	"viiper/pkg/usb"
	"viiper/pkg/usbip"
)

const basepath = "/sys/devices/pci0000:00/0000:00:08.1/0000:00:04:00.3/usb"

var (
	globalBusCounter uint32
	allocatedBusIds  = make(map[uint32]bool)
	globalMutex      sync.Mutex
)

// VirtualBus manages USB bus topology and auto-assigns device addresses.
type VirtualBus struct {
	mutex           sync.Mutex
	busId           uint32
	nextDevID       uint32
	allocatedDevIDs map[uint32]bool
	devices         []busDevice
}

// DeviceMeta exposes a registered device and its metadata for external queries.
type DeviceMeta struct {
	Dev  usb.Device
	Desc DeviceDescriptor
	Meta usbip.ExportMeta
}

// InterfaceConfig holds all descriptors for a single interface for bus management.
type InterfaceConfig struct {
	Descriptor    usb.InterfaceDescriptor
	Endpoints     []usb.EndpointDescriptor
	HIDDescriptor []byte // optional HID class descriptor (0x21)
	HIDReport     []byte // optional HID report descriptor (0x22)
	VendorData    []byte // optional vendor-specific bytes
}

// DeviceDescriptor holds all static descriptor/config data for a device.
type DeviceDescriptor struct {
	Device     usb.DeviceDescriptor
	Interfaces []InterfaceConfig
	Strings    map[uint8][]byte
}

// DescriptorProvider is an optional device interface that exposes a static
// descriptor/config used when registering a device with a VirtualBus via
// `Add(dev)`. Devices that expose this will have their descriptors
// automatically consulted at registration time.
type DescriptorProvider interface {
	GetDeviceDescriptor() DeviceDescriptor
}

// New creates a new VirtualBus instance with a unique auto-assigned bus number.
func New() *VirtualBus {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	busId := globalBusCounter
	if busId == 0 {
		busId = 1
	}
	globalBusCounter = busId + 1
	allocatedBusIds[busId] = true

	return &VirtualBus{
		busId:           busId,
		nextDevID:       0,
		allocatedDevIDs: make(map[uint32]bool),
	}
}

// NewWithBusId creates a new VirtualBus instance starting at a specific bus number.
// Returns an error if the bus number is already allocated.
func NewWithBusId(busId uint32) (*VirtualBus, error) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	if allocatedBusIds[busId] {
		return nil, fmt.Errorf("bus number %d already allocated", busId)
	}
	allocatedBusIds[busId] = true

	return &VirtualBus{
		busId:           busId,
		nextDevID:       0,
		allocatedDevIDs: make(map[uint32]bool),
	}, nil
}

// Add registers a device using a descriptor provider implemented by the device.
// This is a convenience wrapper so callers can simply do "bus.Add(dev)".
// The device must implement a method:
//
//	GetDeviceDescriptor() DeviceDescriptorStruct
//
// which returns a static descriptor that will be used for bus registration.
// Returns a context containing the device's lifecycle and metadata (use GetDeviceMeta to extract).
func (vb *VirtualBus) Add(dev usb.Device) (context.Context, error) {
	if p, ok := dev.(DescriptorProvider); ok {
		desc := p.GetDeviceDescriptor()
		vb.mutex.Lock()
		defer vb.mutex.Unlock()

		for _, d := range vb.devices {
			if d.dev == dev {
				return nil, fmt.Errorf("device already registered on this bus")
			}
		}
		busID := vb.busId
		var devID uint32
		for i := uint32(1); ; i++ {
			if !vb.allocatedDevIDs[i] {
				devID = i
				vb.allocatedDevIDs[i] = true
				break
			}
		}

		busDevID := fmt.Sprintf("%d-%d", busID, devID)
		path := fmt.Sprintf("%s%d/%s", basepath, busID, busDevID)

		var meta usbip.ExportMeta
		copy(meta.Path[:], path)
		copy(meta.USBBusId[:], busDevID)
		meta.BusId = busID
		meta.DevId = devID
		connTimer := time.NewTimer(0)

		ctx, cancel := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, device.ExportMetaKey, &meta)
		ctx = context.WithValue(ctx, device.ConnTimerKey, connTimer)

		vb.devices = append(vb.devices, busDevice{dev: dev, desc: desc, meta: meta, ctx: ctx, cancel: cancel})
		return ctx, nil
	}
	return nil, fmt.Errorf("device does not implement GetDeviceDescriptor")
}

// GetAllDeviceMetas returns a copy of all registered devices with their descriptors and export metadata.
func (vb *VirtualBus) GetAllDeviceMetas() []DeviceMeta {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()
	out := make([]DeviceMeta, 0, len(vb.devices))
	for _, d := range vb.devices {
		out = append(out, DeviceMeta{Dev: d.dev, Desc: d.desc, Meta: d.meta})
	}
	return out
}

// BusID returns the bus number for this VirtualBus.
func (vb *VirtualBus) BusID() uint32 {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()
	return vb.busId
}

// GetDeviceDescriptor looks up a device's descriptor/config by device reference.
// Returns nil if the device is not found.
func (vb *VirtualBus) GetDeviceDescriptor(dev usb.Device) *DeviceDescriptor {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()
	for i := range vb.devices {
		if vb.devices[i].dev == dev {
			return &vb.devices[i].desc
		}
	}
	return nil
}

// Devices returns all devices currently attached to this bus.
func (vb *VirtualBus) Devices() []usb.Device {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()
	out := make([]usb.Device, 0, len(vb.devices))
	for _, d := range vb.devices {
		out = append(out, d.dev)
	}
	return out
}

// RemoveDeviceByID removes a device by its  ID (e.g., "1").
// Returns error if not found.
func (vb *VirtualBus) RemoveDeviceByID(deviceID string) error {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()
	for i, d := range vb.devices {
		if fmt.Sprintf("%d", d.meta.DevId) == deviceID {
			if d.cancel != nil {
				d.cancel()
			}
			delete(vb.allocatedDevIDs, d.meta.DevId)
			vb.devices = append(vb.devices[:i], vb.devices[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("device with id %s not found on bus %d", deviceID, vb.busId)
}

// Remove unregisters a device from the bus.
// This removes the device from the internal list; it does not currently free
// the global bus number. Removal should be used for dynamic device teardown
// during runtime.
func (vb *VirtualBus) Remove(dev usb.Device) error {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()
	for i, d := range vb.devices {
		if d.dev == dev {
			if d.cancel != nil {
				d.cancel()
			}
			delete(vb.allocatedDevIDs, d.meta.DevId)
			vb.devices = append(vb.devices[:i], vb.devices[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("device not found")
}

// Close frees the bus number allocated to this VirtualBus, allowing it to be
// reused. After calling Close, this VirtualBus instance should not be used.
func (vb *VirtualBus) Close() error {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()

	for i := range vb.devices {
		if vb.devices[i].cancel != nil {
			vb.devices[i].cancel()
		}
		vb.devices[i].ctx = nil
		vb.devices[i].cancel = nil
	}

	globalMutex.Lock()
	defer globalMutex.Unlock()

	delete(allocatedBusIds, vb.busId)
	return nil
}

// Note: Contexts are owned by the bus and created at Add(). They are cancelled
// when a device is removed or the bus is closed.

// GetDeviceContext returns the context for a specific device.
// Returns nil if the device is not found or has no active context.
func (vb *VirtualBus) GetDeviceContext(dev usb.Device) context.Context {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()
	for i := range vb.devices {
		if vb.devices[i].dev == dev {
			return vb.devices[i].ctx
		}
	}
	return nil
}

type busDevice struct {
	dev    usb.Device
	desc   DeviceDescriptor
	meta   usbip.ExportMeta
	ctx    context.Context
	cancel context.CancelFunc
}
