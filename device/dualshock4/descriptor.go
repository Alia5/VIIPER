package dualshock4

import (
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
)

var defaultDescriptor = usb.Descriptor{
	Device: usb.DeviceDescriptor{
		BcdUSB:             0x0200,
		BDeviceClass:       0x00,
		BDeviceSubClass:    0x00,
		BDeviceProtocol:    0x00,
		BMaxPacketSize0:    0x40,
		IDVendor:           DefaultVID,
		IDProduct:          DefaultPID,
		BcdDevice:          0x0100,
		IManufacturer:      0x01,
		IProduct:           0x02,
		ISerialNumber:      0x00,
		BNumConfigurations: 0x01,
		Speed:              2,
	},
	Interfaces: []usb.InterfaceConfig{
		{
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x00,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x02,
				BInterfaceClass:    0x03,
				BInterfaceSubClass: 0x00,
				BInterfaceProtocol: 0x00,
				IInterface:         0x00,
			},
			HID: &usb.HIDFunction{
				Descriptor: usb.HIDDescriptor{
					BcdHID:       0x0111,
					BCountryCode: 0x00,
					Descriptors: []usb.HIDSubDescriptor{
						{Type: usb.ReportDescType},
					},
				},
				ReportDescriptor: hid.ReportDescriptor{Items: []hid.Item{
					hid.UsagePage{Page: hid.UsagePageGenericDesktop},
					hid.Usage{Usage: hid.UsageGamePad},
					hid.Collection{Kind: hid.CollectionApplication, Items: []hid.Item{

						hid.ReportID{ID: ReportIDInput},
						hid.UsagePage{Page: hid.UsagePageGenericDesktop},
						hid.Usage{Usage: hid.UsageX},
						hid.Usage{Usage: hid.UsageY},
						hid.Usage{Usage: hid.UsageZ},
						hid.Usage{Usage: hid.UsageRz},
						hid.LogicalMinimum{Min: 0},
						hid.LogicalMaximum{Max: 255},
						hid.ReportSize{Bits: 8},
						hid.ReportCount{Count: 4},
						hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.Usage{Usage: 0x39},
						hid.LogicalMinimum{Min: 0},
						hid.LogicalMaximum{Max: 7},
						hid.PhysicalMinimum{Min: 0},
						hid.PhysicalMaximum{Max: 315},
						hid.Unit{Value: 0x14},
						hid.ReportSize{Bits: 4},
						hid.ReportCount{Count: 1},
						hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs | hid.MainNullState},
						hid.Unit{Value: 0},

						hid.UsagePage{Page: hid.UsagePageButton},
						hid.UsageMinimum{Min: 0x01},
						hid.UsageMaximum{Max: 0x0E},
						hid.LogicalMinimum{Min: 0},
						hid.LogicalMaximum{Max: 1},
						hid.ReportSize{Bits: 1},
						hid.ReportCount{Count: 14},
						hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.UsagePage{Page: 0xFF00},
						hid.Usage{Usage: 0x20},
						hid.ReportSize{Bits: 6},
						hid.ReportCount{Count: 1},
						hid.LogicalMinimum{Min: 0},
						hid.LogicalMaximum{Max: 0x7F},
						hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.UsagePage{Page: hid.UsagePageGenericDesktop},
						hid.Usage{Usage: hid.UsageRx},
						hid.Usage{Usage: hid.UsageRy},
						hid.LogicalMinimum{Min: 0},
						hid.LogicalMaximum{Max: 255},
						hid.ReportSize{Bits: 8},
						hid.ReportCount{Count: 2},
						hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.UsagePage{Page: 0xFF00},
						hid.Usage{Usage: 0x21},
						hid.ReportCount{Count: 54},
						hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: ReportIDOutput},
						hid.Usage{Usage: 0x22},
						hid.ReportCount{Count: 31},
						hid.Output{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x04},
						hid.Usage{Usage: 0x23},
						hid.ReportCount{Count: 36},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: featureIDCalibration},
						hid.Usage{Usage: 0x24},
						hid.ReportCount{Count: 36},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: featureIDProbe},
						hid.Usage{Usage: 0x25},
						hid.ReportCount{Count: 3},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: featureIDStatus},
						hid.Usage{Usage: 0x26},
						hid.ReportCount{Count: 4},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: featureIDProbeResponse},
						hid.Usage{Usage: 0x27},
						hid.ReportCount{Count: 2},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: featureIDSerial},
						hid.UsagePage{Page: 0xFF02},
						hid.Usage{Usage: 0x21},
						hid.ReportCount{Count: 15},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x13},
						hid.Usage{Usage: 0x22},
						hid.ReportCount{Count: 22},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x14},
						hid.UsagePage{Page: 0xFF05},
						hid.Usage{Usage: 0x20},
						hid.ReportCount{Count: 16},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x15},
						hid.Usage{Usage: 0x21},
						hid.ReportCount{Count: 44},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.UsagePage{Page: 0xFF80},

						hid.ReportID{ID: 0x80},
						hid.Usage{Usage: 0x20},
						hid.ReportCount{Count: 6},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: featureIDIdentity},
						hid.Usage{Usage: 0x21},
						hid.ReportCount{Count: 6},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x82},
						hid.Usage{Usage: 0x22},
						hid.ReportCount{Count: 5},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x83},
						hid.Usage{Usage: 0x23},
						hid.ReportCount{Count: 1},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x84},
						hid.Usage{Usage: 0x24},
						hid.ReportCount{Count: 4},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x85},
						hid.Usage{Usage: 0x25},
						hid.ReportCount{Count: 6},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x86},
						hid.Usage{Usage: 0x26},
						hid.ReportCount{Count: 6},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x87},
						hid.Usage{Usage: 0x27},
						hid.ReportCount{Count: 35},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x88},
						hid.Usage{Usage: 0x28},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x89},
						hid.Usage{Usage: 0x29},
						hid.ReportCount{Count: 2},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x90},
						hid.Usage{Usage: 0x30},
						hid.ReportCount{Count: 5},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x91},
						hid.Usage{Usage: 0x31},
						hid.ReportCount{Count: 3},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x92},
						hid.Usage{Usage: 0x32},
						hid.ReportCount{Count: 3},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x93},
						hid.Usage{Usage: 0x33},
						hid.ReportCount{Count: 12},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0x94},
						hid.Usage{Usage: 0x34},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: featureIDSubcommand},
						hid.Usage{Usage: 0x40},
						hid.ReportCount{Count: 6},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xA1},
						hid.Usage{Usage: 0x41},
						hid.ReportCount{Count: 1},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xA2},
						hid.Usage{Usage: 0x42},
						hid.ReportCount{Count: 1},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: featureIDBoardInfo},
						hid.Usage{Usage: 0x43},
						hid.ReportCount{Count: 48},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: featureIDTelemetry},
						hid.Usage{Usage: 0x44},
						hid.ReportCount{Count: 13},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xF0},
						hid.Usage{Usage: 0x47},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xF1},
						hid.Usage{Usage: 0x48},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xF2},
						hid.Usage{Usage: 0x49},
						hid.ReportCount{Count: 15},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xA7},
						hid.Usage{Usage: 0x4A},
						hid.ReportCount{Count: 1},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xA8},
						hid.Usage{Usage: 0x4B},
						hid.ReportCount{Count: 1},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xA9},
						hid.Usage{Usage: 0x4C},
						hid.ReportCount{Count: 8},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xAA},
						hid.Usage{Usage: 0x4E},
						hid.ReportCount{Count: 1},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xAB},
						hid.Usage{Usage: 0x4F},
						hid.ReportCount{Count: 57},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xAC},
						hid.Usage{Usage: 0x50},
						hid.ReportCount{Count: 57},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xAD},
						hid.Usage{Usage: 0x51},
						hid.ReportCount{Count: 11},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xAE},
						hid.Usage{Usage: 0x52},
						hid.ReportCount{Count: 1},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xAF},
						hid.Usage{Usage: 0x53},
						hid.ReportCount{Count: 2},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xB0},
						hid.Usage{Usage: 0x54},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xE0},
						hid.Usage{Usage: 0x57},
						hid.ReportCount{Count: 2},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xB3},
						hid.Usage{Usage: 0x55},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xB4},
						hid.Usage{Usage: 0x55},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xB5},
						hid.Usage{Usage: 0x56},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xD0},
						hid.Usage{Usage: 0x58},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

						hid.ReportID{ID: 0xD4},
						hid.Usage{Usage: 0x59},
						hid.ReportCount{Count: 63},
						hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
					}},
				}},
			},
			Endpoints: []usb.EndpointDescriptor{
				{
					BEndpointAddress: EndpointIn,
					BMAttributes:     0x03,
					WMaxPacketSize:   64,
					BInterval:        4,
				},
				{
					BEndpointAddress: EndpointOut,
					BMAttributes:     0x03,
					WMaxPacketSize:   64,
					BInterval:        4,
				},
			},
		},
	},
	Strings: map[uint8]string{
		0: "\u0409", // LangID: en-US (0x0409)
		1: "Sony Interactive Entertainment",
		2: "Wireless Controller",
	},
}
