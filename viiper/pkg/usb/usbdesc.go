// Package usb contains helpers for building USB descriptors and data.
package usb

import (
	"bytes"
	"encoding/binary"
)

// USB descriptor type constants
const (
	DeviceDescType    = 0x01
	ConfigDescType    = 0x02
	InterfaceDescType = 0x04
	EndpointDescType  = 0x05
	HIDDescType       = 0x21
	ReportDescType    = 0x22
)

// Descriptor lengths in bytes (fixed values from USB spec)
const (
	DeviceDescLen    = 18
	ConfigDescLen    = 9
	InterfaceDescLen = 9
	EndpointDescLen  = 7
	HIDDescLen       = 9
)

// DeviceDescriptor represents the standard USB device descriptor.
// BLength is computed dynamically; BDescriptorType is implied DeviceDescType.
type DeviceDescriptor struct {
	BcdUSB             uint16 // LE
	BDeviceClass       uint8
	BDeviceSubClass    uint8
	BDeviceProtocol    uint8
	BMaxPacketSize0    uint8
	IDVendor           uint16 // LE
	IDProduct          uint16 // LE
	BcdDevice          uint16 // LE
	IManufacturer      uint8
	IProduct           uint8
	ISerialNumber      uint8
	BNumConfigurations uint8
	Speed              uint32 // USB speed: 1=low, 2=full, 3=high, 4=super
}

// Bytes returns the binary representation of the DeviceDescriptor with BLength auto-filled.
func (d DeviceDescriptor) Bytes() []byte {
	var b bytes.Buffer
	b.WriteByte(DeviceDescLen)
	b.WriteByte(DeviceDescType)
	_ = binary.Write(&b, binary.LittleEndian, d.BcdUSB)
	b.WriteByte(d.BDeviceClass)
	b.WriteByte(d.BDeviceSubClass)
	b.WriteByte(d.BDeviceProtocol)
	b.WriteByte(d.BMaxPacketSize0)
	_ = binary.Write(&b, binary.LittleEndian, d.IDVendor)
	_ = binary.Write(&b, binary.LittleEndian, d.IDProduct)
	_ = binary.Write(&b, binary.LittleEndian, d.BcdDevice)
	b.WriteByte(d.IManufacturer)
	b.WriteByte(d.IProduct)
	b.WriteByte(d.ISerialNumber)
	b.WriteByte(d.BNumConfigurations)
	return b.Bytes()
}

// ConfigHeader represents the USB configuration descriptor header (9 bytes).
type ConfigHeader struct {
	WTotalLength        uint16 // LE, to be patched after building
	BNumInterfaces      uint8
	BConfigurationValue uint8
	IConfiguration      uint8
	BMAttributes        uint8
	BMaxPower           uint8
}

func (h ConfigHeader) Write(b *bytes.Buffer) {
	b.WriteByte(ConfigDescLen)
	b.WriteByte(ConfigDescType)
	_ = binary.Write(b, binary.LittleEndian, h.WTotalLength)
	b.WriteByte(h.BNumInterfaces)
	b.WriteByte(h.BConfigurationValue)
	b.WriteByte(h.IConfiguration)
	b.WriteByte(h.BMAttributes)
	b.WriteByte(h.BMaxPower)

}

// InterfaceDescriptor (9 bytes) for each interface altsetting.
type InterfaceDescriptor struct {
	BInterfaceNumber   uint8
	BAlternateSetting  uint8
	BNumEndpoints      uint8
	BInterfaceClass    uint8
	BInterfaceSubClass uint8
	BInterfaceProtocol uint8
	IInterface         uint8
}

func (i InterfaceDescriptor) Write(b *bytes.Buffer) {
	b.WriteByte(InterfaceDescLen)
	b.WriteByte(InterfaceDescType)
	b.WriteByte(i.BInterfaceNumber)
	b.WriteByte(i.BAlternateSetting)
	b.WriteByte(i.BNumEndpoints)
	b.WriteByte(i.BInterfaceClass)
	b.WriteByte(i.BInterfaceSubClass)
	b.WriteByte(i.BInterfaceProtocol)
	b.WriteByte(i.IInterface)

}

// EndpointDescriptor (7 bytes) for each endpoint.
type EndpointDescriptor struct {
	BEndpointAddress uint8
	BMAttributes     uint8
	WMaxPacketSize   uint16 // LE
	BInterval        uint8
}

func (e EndpointDescriptor) Write(b *bytes.Buffer) {
	b.WriteByte(EndpointDescLen)
	b.WriteByte(EndpointDescType)
	b.WriteByte(e.BEndpointAddress)
	b.WriteByte(e.BMAttributes)
	_ = binary.Write(b, binary.LittleEndian, e.WMaxPacketSize)
	b.WriteByte(e.BInterval)

}

// HIDDescriptor (class descriptor, 0x21) with one subordinate report descriptor (0x22).
type HIDDescriptor struct {
	BcdHID            uint16 // LE
	BCountryCode      uint8
	BNumDescriptors   uint8
	ClassDescType     uint8  // 0x22 (report)
	WDescriptorLength uint16 // LE, report descriptor length
}

func (h HIDDescriptor) Write(b *bytes.Buffer) {
	b.WriteByte(HIDDescLen)
	b.WriteByte(HIDDescType)
	_ = binary.Write(b, binary.LittleEndian, h.BcdHID)
	b.WriteByte(h.BCountryCode)
	b.WriteByte(h.BNumDescriptors)
	b.WriteByte(h.ClassDescType)
	_ = binary.Write(b, binary.LittleEndian, h.WDescriptorLength)

}

// ReportDescriptor is a container for HID report descriptor bytes (0x22).
// Builders can populate Data to emit via Bytes().
type ReportDescriptor struct {
	Data []byte
}

func (r ReportDescriptor) Bytes() []byte {
	return r.Data
}
