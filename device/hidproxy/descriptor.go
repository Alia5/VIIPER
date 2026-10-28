package hidproxy

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Alia5/VIIPER/usb"
)

// Limits on what a client may describe. A real device's descriptors are far smaller; the limits only keep a
// malformed request from making the server hold or serve megabytes.
const (
	maxConfigLength = 4096
	maxReportLength = 4096
	maxStringLength = 126 // a string descriptor holds at most 126 UTF-16 code units
	maxStrings      = 32
)

// Options is the device-specific part of a create request: the descriptors of the device to copy, as the real
// device reports them (hex), so the copy enumerates exactly like it.
type Options struct {
	// Device is the 18-byte device descriptor.
	Device string `json:"device"`
	// Config is the whole configuration descriptor (header, interfaces, class and endpoint descriptors).
	Config string `json:"config"`
	// Strings holds the string descriptors by index, the serial number included if the device has one.
	Strings map[string]string `json:"strings"`
	// Reports holds the HID report descriptor of each HID interface, keyed by interface number.
	Reports map[string]string `json:"reports"`
	// Speed is usbip's speed code (1 low, 2 full, 3 high); 0 means full speed.
	Speed uint32 `json:"speed"`
}

// parseOptions turns the raw descriptors into VIIPER's structured form.
func parseOptions(o Options) (*usb.Descriptor, error) {
	dev, err := hex.DecodeString(o.Device)
	if err != nil || len(dev) != usb.DeviceDescLen || dev[1] != usb.DeviceDescType {
		return nil, fmt.Errorf("device descriptor: want %d bytes of hex", usb.DeviceDescLen)
	}
	cfg, err := hex.DecodeString(o.Config)
	if err != nil || len(cfg) < usb.ConfigDescLen || len(cfg) > maxConfigLength || cfg[1] != usb.ConfigDescType {
		return nil, fmt.Errorf("configuration descriptor: want %d to %d bytes of hex", usb.ConfigDescLen, maxConfigLength)
	}
	reports := map[uint8][]byte{}
	for key, value := range o.Reports {
		number, err := strconv.ParseUint(key, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("report descriptor key %q: want an interface number", key)
		}
		bytes, err := hex.DecodeString(value)
		if err != nil || len(bytes) == 0 || len(bytes) > maxReportLength {
			return nil, fmt.Errorf("report descriptor %d: want 1 to %d bytes of hex", number, maxReportLength)
		}
		reports[uint8(number)] = bytes
	}
	if len(o.Strings) > maxStrings {
		return nil, fmt.Errorf("strings: at most %d", maxStrings)
	}

	speed := o.Speed
	if speed == 0 {
		speed = 2
	}
	if speed > 3 {
		return nil, fmt.Errorf("speed %d: want 1 (low), 2 (full) or 3 (high)", speed)
	}
	d := &usb.Descriptor{
		Device: usb.DeviceDescriptor{
			BcdUSB:             binary.LittleEndian.Uint16(dev[2:]),
			BDeviceClass:       dev[4],
			BDeviceSubClass:    dev[5],
			BDeviceProtocol:    dev[6],
			BMaxPacketSize0:    dev[7],
			IDVendor:           binary.LittleEndian.Uint16(dev[8:]),
			IDProduct:          binary.LittleEndian.Uint16(dev[10:]),
			BcdDevice:          binary.LittleEndian.Uint16(dev[12:]),
			IManufacturer:      dev[14],
			IProduct:           dev[15],
			ISerialNumber:      dev[16],
			BNumConfigurations: dev[17],
			Speed:              speed,
		},
		// Index 0 is the list of languages: US English.
		Strings: map[uint8]string{0: "Љ"},
	}
	for key, value := range o.Strings {
		index, err := strconv.ParseUint(key, 10, 8)
		if err != nil || index == 0 {
			return nil, fmt.Errorf("string key %q: want an index from 1 to 255", key)
		}
		if len([]rune(value)) > maxStringLength {
			return nil, fmt.Errorf("string %d: at most %d characters", index, maxStringLength)
		}
		d.Strings[uint8(index)] = value
	}
	// USB 2.01 makes Windows ask for a BOS descriptor VIIPER doesn't serve; nothing else differs from 2.00.
	if d.Device.BcdUSB > 0x0200 {
		d.Device.BcdUSB = 0x0200
	}
	if err := parseConfiguration(d, cfg, reports); err != nil {
		return nil, err
	}
	return d, nil
}

// parseConfiguration walks the configuration descriptor: associations, interfaces with their HID descriptor and
// endpoints, and keeps any other descriptor as an opaque class descriptor where it stood.
func parseConfiguration(d *usb.Descriptor, cfg []byte, reports map[uint8][]byte) error {
	d.Configuration = usb.ConfigurationDescriptor{
		BConfigurationValue: cfg[5],
		IConfiguration:      cfg[6],
		BMAttributes:        cfg[7],
		BMaxPower:           cfg[8],
	}
	var current *usb.InterfaceConfig
	for at := int(cfg[0]); at < len(cfg); {
		if at+1 >= len(cfg) {
			return fmt.Errorf("configuration descriptor: truncated at %d", at)
		}
		length, kind := int(cfg[at]), cfg[at+1]
		if length < 2 || at+length > len(cfg) {
			return fmt.Errorf("configuration descriptor: bad length %d at %d", length, at)
		}
		body := cfg[at : at+length]
		switch {
		case kind == usb.IADDescType && length >= usb.IADDescLen:
			d.Associations = append(d.Associations, usb.InterfaceAssociationDescriptor{
				BFirstInterface: body[2], BInterfaceCount: body[3], BFunctionClass: body[4],
				BFunctionSubClass: body[5], BFunctionProtocol: body[6], IFunction: body[7],
			})
		case kind == usb.InterfaceDescType && length >= usb.InterfaceDescLen:
			d.Interfaces = append(d.Interfaces, usb.InterfaceConfig{Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber: body[2], BAlternateSetting: body[3], BNumEndpoints: body[4],
				BInterfaceClass: body[5], BInterfaceSubClass: body[6], BInterfaceProtocol: body[7], IInterface: body[8],
			}})
			current = &d.Interfaces[len(d.Interfaces)-1]
		case kind == usb.HIDDescType && current != nil && current.Descriptor.BInterfaceClass == classHID && length >= 6:
			hid := usb.HIDDescriptor{BcdHID: binary.LittleEndian.Uint16(body[2:]), BCountryCode: body[4]}
			for i := 0; i < int(body[5]) && 6+3*i+2 < length; i++ {
				sub := usb.HIDSubDescriptor{Type: body[6+3*i], Length: binary.LittleEndian.Uint16(body[7+3*i:])}
				if sub.Type == usb.ReportDescType {
					// Length 0: VIIPER fills in the length of the report descriptor it serves.
					sub.Length = 0
				}
				hid.Descriptors = append(hid.Descriptors, sub)
			}
			report, ok := reports[current.Descriptor.BInterfaceNumber]
			if !ok {
				return fmt.Errorf("no report descriptor for HID interface %d", current.Descriptor.BInterfaceNumber)
			}
			current.HID = &usb.HIDFunction{Descriptor: hid, ReportDescriptorBytes: report}
		case kind == usb.EndpointDescType && current != nil && length >= usb.EndpointDescLen:
			current.Endpoints = append(current.Endpoints, usb.EndpointDescriptor{
				BEndpointAddress: body[2], BMAttributes: body[3],
				WMaxPacketSize: binary.LittleEndian.Uint16(body[4:]), BInterval: body[6],
			})
		case current != nil && len(current.Endpoints) > 0:
			last := &current.Endpoints[len(current.Endpoints)-1]
			last.ClassDescriptors = append(last.ClassDescriptors, usb.ClassSpecificDescriptor{DescriptorType: kind, Payload: append([]byte(nil), body[2:]...)})
		case current != nil:
			current.ClassDescriptors = append(current.ClassDescriptors, usb.ClassSpecificDescriptor{DescriptorType: kind, Payload: append([]byte(nil), body[2:]...)})
		}
		at += length
	}
	if len(d.Interfaces) == 0 {
		return fmt.Errorf("configuration descriptor has no interfaces")
	}
	return nil
}

// decodeOptions reads the device-specific JSON of a create request.
func decodeOptions(deviceSpecific string) (Options, error) {
	var o Options
	if deviceSpecific == "" {
		return o, fmt.Errorf("hidproxy needs the descriptors of the device to copy in deviceSpecific")
	}
	if err := json.Unmarshal([]byte(deviceSpecific), &o); err != nil {
		return o, fmt.Errorf("invalid device specific JSON: %w", err)
	}
	return o, nil
}
