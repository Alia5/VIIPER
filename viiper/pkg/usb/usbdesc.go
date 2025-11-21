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

// Descriptor holds all static descriptor/config data for a device.
type Descriptor struct {
	Device     DeviceDescriptor
	Interfaces []InterfaceConfig
	Strings    map[uint8]string
}

// InterfaceConfig holds all descriptors for a single interface for bus management.
type InterfaceConfig struct {
	Descriptor    InterfaceDescriptor
	Endpoints     []EndpointDescriptor
	HIDDescriptor []byte // optional HID class descriptor (0x21)
	HIDReport     []byte // optional HID report descriptor (0x22)
	VendorData    []byte // optional vendor-specific bytes
}

// EncodeStringDescriptor converts a UTF-8 string to a USB string descriptor byte array.
// The resulting descriptor has the format:
//
//	Byte 0: bLength (total descriptor length)
//	Byte 1: bDescriptorType (0x03 for string)
//	Bytes 2+: UTF-16LE encoded string
func EncodeStringDescriptor(s string) []byte {
	runes := []rune(s)
	buf := make([]byte, 2+len(runes)*2)
	buf[0] = uint8(len(buf)) // bLength
	buf[1] = 0x03            // bDescriptorType (STRING)
	for i, r := range runes {
		buf[2+i*2] = uint8(r)
		buf[2+i*2+1] = uint8(r >> 8)
	}
	return buf
}

// DeviceDescriptor represents the standard USB device descriptor.
// BLength is computed dynamically; BDescriptorType is implied DeviceDescType.
type DeviceDescriptor struct {
	BcdUSB             uint16 // LE
	BDeviceClass       uint8
	BDeviceSubClass    uint8
	BDeviceProtocol    uint8
	BMaxPacketSize0    uint8
	IDVendor           uint16 // LE; may get overridden
	IDProduct          uint16 // LE; may get overridden
	BcdDevice          uint16 // LE
	IManufacturer      uint8
	IProduct           uint8
	ISerialNumber      uint8
	BNumConfigurations uint8
	Speed              uint32 // USB speed: 1=low, 2=full, 3=high, 4=super
}

// Bytes returns the binary representation of the DeviceDescriptor with BLength auto-filled.
func (d Descriptor) Bytes() []byte {
	var b bytes.Buffer
	b.WriteByte(DeviceDescLen)
	b.WriteByte(DeviceDescType)
	_ = binary.Write(&b, binary.LittleEndian, d.Device.BcdUSB)
	b.WriteByte(d.Device.BDeviceClass)
	b.WriteByte(d.Device.BDeviceSubClass)
	b.WriteByte(d.Device.BDeviceProtocol)
	b.WriteByte(d.Device.BMaxPacketSize0)
	_ = binary.Write(&b, binary.LittleEndian, d.Device.IDVendor)
	_ = binary.Write(&b, binary.LittleEndian, d.Device.IDProduct)
	_ = binary.Write(&b, binary.LittleEndian, d.Device.BcdDevice)
	b.WriteByte(d.Device.IManufacturer)
	b.WriteByte(d.Device.IProduct)
	b.WriteByte(d.Device.ISerialNumber)
	b.WriteByte(d.Device.BNumConfigurations)
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
