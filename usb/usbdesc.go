// Package usb contains helpers for building USB descriptors and data.
package usb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf16"

	"github.com/Alia5/VIIPER/usb/hid"
)

// USB descriptor type constants
const (
	DeviceDescType    = 0x01
	ConfigDescType    = 0x02
	InterfaceDescType = 0x04
	EndpointDescType  = 0x05
	IADDescType       = 0x0B
	HIDDescType       = 0x21
	ReportDescType    = 0x22
)

// Descriptor lengths in bytes (fixed values from USB spec)
const (
	DeviceDescLen    = 18
	ConfigDescLen    = 9
	IADDescLen       = 8
	InterfaceDescLen = 9
	EndpointDescLen  = 7
	HIDDescLen       = 9
)

type Data []uint8

// Descriptor holds all static descriptor/config data for a device.
type Descriptor struct {
	Device        DeviceDescriptor
	Configuration ConfigurationDescriptor
	MicrosoftOS10 *MicrosoftOS10Descriptor
	Associations  []InterfaceAssociationDescriptor
	Interfaces    []InterfaceConfig
	Strings       map[uint8]string
}

// MicrosoftOS10Descriptor enables the Microsoft OS 1.0 descriptor probe used by
// Windows to bind vendor-specific interfaces to inbox drivers such as WinUSB.
type MicrosoftOS10Descriptor struct {
	VendorCode          uint8
	InterfaceNumber     uint8
	CompatibleID        string
	SubCompatibleID     string
	DeviceInterfaceGUID string
}

func (d MicrosoftOS10Descriptor) StringDescriptor() []byte {
	return []byte{
		0x12, 0x03,
		'M', 0x00,
		'S', 0x00,
		'F', 0x00,
		'T', 0x00,
		'1', 0x00,
		'0', 0x00,
		'0', 0x00,
		d.EffectiveVendorCode(), 0x00,
	}
}

func (d MicrosoftOS10Descriptor) EffectiveVendorCode() uint8 {
	if d.VendorCode == 0 {
		return 0x20
	}
	return d.VendorCode
}

func (d MicrosoftOS10Descriptor) ControlResponse(wValue, wIndex uint16) ([]byte, bool) {
	switch {
	case wIndex == 0x0004:
		return d.CompatibleIDDescriptor(), true
	case wIndex == 0x0005 || wValue == 0x0005:
		return d.ExtendedPropertiesDescriptor(), true
	default:
		return nil, false
	}
}

func (d MicrosoftOS10Descriptor) CompatibleIDDescriptor() []byte {
	out := make([]byte, 40)
	binary.LittleEndian.PutUint32(out[0:4], uint32(len(out)))
	binary.LittleEndian.PutUint16(out[4:6], 0x0100)
	binary.LittleEndian.PutUint16(out[6:8], 0x0004)
	out[8] = 0x01
	out[16] = d.InterfaceNumber
	copyFixedASCII(out[18:26], d.CompatibleID)
	copyFixedASCII(out[26:34], d.SubCompatibleID)
	return out
}

func (d MicrosoftOS10Descriptor) ExtendedPropertiesDescriptor() []byte {
	if d.DeviceInterfaceGUID == "" {
		out := make([]byte, 10)
		binary.LittleEndian.PutUint32(out[0:4], uint32(len(out)))
		binary.LittleEndian.PutUint16(out[4:6], 0x0100)
		binary.LittleEndian.PutUint16(out[6:8], 0x0005)
		return out
	}

	name := utf16leString("DeviceInterfaceGUID")
	data := utf16leString(d.DeviceInterfaceGUID)

	sectionLen := 4 + 4 + 2 + len(name) + 4 + len(data)
	out := make([]byte, 10+sectionLen)
	binary.LittleEndian.PutUint32(out[0:4], uint32(len(out)))
	binary.LittleEndian.PutUint16(out[4:6], 0x0100)
	binary.LittleEndian.PutUint16(out[6:8], 0x0005)
	binary.LittleEndian.PutUint16(out[8:10], 0x0001)

	off := 10
	binary.LittleEndian.PutUint32(out[off:off+4], uint32(sectionLen))
	off += 4
	binary.LittleEndian.PutUint32(out[off:off+4], 0x00000001) // REG_SZ
	off += 4
	binary.LittleEndian.PutUint16(out[off:off+2], uint16(len(name)))
	off += 2
	copy(out[off:off+len(name)], name)
	off += len(name)
	binary.LittleEndian.PutUint32(out[off:off+4], uint32(len(data)))
	off += 4
	copy(out[off:], data)
	return out
}

func copyFixedASCII(dst []byte, s string) {
	for i := range dst {
		dst[i] = 0
	}
	copy(dst, []byte(s))
}

func utf16leString(s string) []byte {
	units := utf16.Encode([]rune(s + "\x00"))
	out := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(out[i*2:i*2+2], u)
	}
	return out
}

// NumInterfaces returns the number of distinct interface numbers in the active
// configuration. Alternate settings share the same interface number and only
// count once in the USB configuration descriptor header.
func (d Descriptor) NumInterfaces() uint8 {
	seen := map[uint8]struct{}{}
	for _, iface := range d.Interfaces {
		seen[iface.Descriptor.BInterfaceNumber] = struct{}{}
	}
	return uint8(len(seen))
}

// Interface returns the first matching interface descriptor for an interface
// number, preferring alternate setting zero when available.
func (d Descriptor) Interface(number uint8) (InterfaceConfig, bool) {
	var found InterfaceConfig
	ok := false
	for _, iface := range d.Interfaces {
		if iface.Descriptor.BInterfaceNumber != number {
			continue
		}
		if iface.Descriptor.BAlternateSetting == 0 {
			return iface, true
		}
		if !ok {
			found = iface
			ok = true
		}
	}
	return found, ok
}

// InterfaceConfig holds all descriptors for a single interface for bus management.
type InterfaceConfig struct {
	Descriptor InterfaceDescriptor
	Endpoints  []EndpointDescriptor

	// HID describes a HID-class interface (bInterfaceClass=0x03).
	// If set, the server will emit the HID descriptor (0x21) in the configuration
	// descriptor and serve the report descriptor (0x22) via GET_DESCRIPTOR.
	HID *HIDFunction

	// ClassDescriptors are additional interface-level class-specific descriptors
	// emitted as part of the configuration descriptor (after the interface descriptor
	// and before the endpoints).
	//
	// This is also used for vendor-specific interfaces that need to expose opaque
	// descriptors (e.g. type 0x21 blobs on Xbox360).
	ClassDescriptors []ClassSpecificDescriptor
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

// ConfigurationDescriptor contains the non-derived fields from the USB
// configuration descriptor. Zero values keep the server defaults.
type ConfigurationDescriptor struct {
	BConfigurationValue uint8
	IConfiguration      uint8
	BMAttributes        uint8
	BMaxPower           uint8
}

// InterfaceAssociationDescriptor describes a USB Interface Association
// Descriptor (IAD), used by composite devices to group related interfaces.
type InterfaceAssociationDescriptor struct {
	BFirstInterface   uint8
	BInterfaceCount   uint8
	BFunctionClass    uint8
	BFunctionSubClass uint8
	BFunctionProtocol uint8
	IFunction         uint8
}

func (i InterfaceAssociationDescriptor) Write(b *bytes.Buffer) {
	b.WriteByte(IADDescLen)
	b.WriteByte(IADDescType)
	b.WriteByte(i.BFirstInterface)
	b.WriteByte(i.BInterfaceCount)
	b.WriteByte(i.BFunctionClass)
	b.WriteByte(i.BFunctionSubClass)
	b.WriteByte(i.BFunctionProtocol)
	b.WriteByte(i.IFunction)
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

	// ClassDescriptors are optional endpoint-level class-specific descriptors
	// emitted immediately after this endpoint descriptor.
	ClassDescriptors []ClassSpecificDescriptor
}

func (e EndpointDescriptor) Write(b *bytes.Buffer) {
	b.WriteByte(EndpointDescLen)
	b.WriteByte(EndpointDescType)
	b.WriteByte(e.BEndpointAddress)
	b.WriteByte(e.BMAttributes)
	_ = binary.Write(b, binary.LittleEndian, e.WMaxPacketSize)
	b.WriteByte(e.BInterval)

}

// HIDSubDescriptor is one subordinate descriptor entry in the HID class descriptor.
//
// Type is typically ReportDescType (0x22). If Type==ReportDescType and Length==0,
// the server will auto-fill Length from the associated HID report descriptor at
// serialization time.
type HIDSubDescriptor struct {
	Type   uint8
	Length uint16 // LE
}

// HIDDescriptor is the HID class descriptor (0x21) for HID-class interfaces.
//
// bDescriptorType is fixed to HIDDescType (0x21).
// bLength is auto-calculated as: 6 + 3*len(Descriptors).
type HIDDescriptor struct {
	BcdHID       uint16 // LE
	BCountryCode uint8
	Descriptors  []HIDSubDescriptor
}

func (h HIDDescriptor) IsZero() bool {
	return h.BcdHID == 0 && h.BCountryCode == 0 && len(h.Descriptors) == 0
}

func (h HIDDescriptor) Write(b *bytes.Buffer, reportLen uint16) error {
	if len(h.Descriptors) == 0 {
		return fmt.Errorf("usb: HIDDescriptor has no subordinate descriptors")
	}
	b.WriteByte(uint8(6 + 3*len(h.Descriptors)))
	b.WriteByte(HIDDescType)
	_ = binary.Write(b, binary.LittleEndian, h.BcdHID)
	b.WriteByte(h.BCountryCode)
	b.WriteByte(uint8(len(h.Descriptors)))
	for _, sd := range h.Descriptors {
		b.WriteByte(sd.Type)
		l := sd.Length
		if sd.Type == ReportDescType && l == 0 {
			l = reportLen
		}
		_ = binary.Write(b, binary.LittleEndian, l)
	}
	return nil
}

// ClassSpecificDescriptor represents an opaque class-specific interface descriptor.
//
// It auto-calculates bLength and hardcodes bDescriptorType. Payload contains all bytes
// after the (bLength,bDescriptorType) header.
type ClassSpecificDescriptor struct {
	DescriptorType uint8
	Payload        Data
}

func (d ClassSpecificDescriptor) Bytes() Data {
	out := make([]uint8, 0, 2+len(d.Payload))
	out = append(out, uint8(2+len(d.Payload)))
	out = append(out, d.DescriptorType)
	out = append(out, d.Payload...)
	return Data(out)
}

// HIDFunction bundles the HID class descriptor (0x21) and the report descriptor (0x22)
// for a HID-class interface.
type HIDFunction struct {
	Descriptor       HIDDescriptor
	ReportDescriptor hid.ReportDescriptor
	// ReportDescriptorBytes, when non-empty, is returned verbatim as the HID report
	// descriptor (0x22) and used for HID report length calculation. This is useful for
	// complex, vendor-specific descriptors that are easier to provide as raw bytes.
	ReportDescriptorBytes Data
}

func (f HIDFunction) reportLen() (uint16, error) {
	if len(f.ReportDescriptorBytes) > 0 {
		if len(f.ReportDescriptorBytes) > 0xFFFF {
			return 0, fmt.Errorf("usb: HID raw report descriptor too large: %d", len(f.ReportDescriptorBytes))
		}
		return uint16(len(f.ReportDescriptorBytes)), nil
	}

	rb, err := f.ReportDescriptor.Bytes()
	if err != nil {
		return 0, err
	}
	if len(rb) > 0xFFFF {
		return 0, fmt.Errorf("usb: HID report descriptor too large: %d", len(rb))
	}
	return uint16(len(rb)), nil
}

// DescriptorBytes returns the HID class descriptor (0x21) bytes.
func (f HIDFunction) DescriptorBytes() (Data, error) {
	rl, err := f.reportLen()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := f.Descriptor.Write(&b, rl); err != nil {
		return nil, err
	}
	return Data(b.Bytes()), nil
}

// ReportBytes returns the HID report descriptor (0x22) bytes.
func (f HIDFunction) ReportBytes() (Data, error) {
	if len(f.ReportDescriptorBytes) > 0 {
		out := make([]uint8, len(f.ReportDescriptorBytes))
		copy(out, f.ReportDescriptorBytes)
		return Data(out), nil
	}

	rb, err := f.ReportDescriptor.Bytes()
	if err != nil {
		return nil, err
	}
	return Data(rb), nil
}
