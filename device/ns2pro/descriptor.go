package ns2pro

import (
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
)

const microsoftOS10VendorCode = 0x20

const microsoftOS10DeviceInterfaceGUID = "{7D8F1E5C-9D89-4E0D-A2F0-6D10F1F89A8F}"

func MakeDescriptor() usb.Descriptor {
	return usb.Descriptor{
		Device: usb.DeviceDescriptor{
			BcdUSB:             0x0200,
			BDeviceClass:       0xEF,
			BDeviceSubClass:    0x02,
			BDeviceProtocol:    0x01,
			BMaxPacketSize0:    0x40,
			IDVendor:           DefaultVID,
			IDProduct:          DefaultPID,
			BcdDevice:          0x0200,
			IManufacturer:      0x01,
			IProduct:           0x02,
			ISerialNumber:      0x03,
			BNumConfigurations: 0x01,
			Speed:              2,
		},
		Configuration: usb.ConfigurationDescriptor{
			BConfigurationValue: 0x01,
			IConfiguration:      0x04,
			BMAttributes:        0xC0,
			BMaxPower:           0xFA,
		},
		MicrosoftOS10: &usb.MicrosoftOS10Descriptor{
			VendorCode:          microsoftOS10VendorCode,
			InterfaceNumber:     0x01,
			CompatibleID:        "WINUSB",
			DeviceInterfaceGUID: microsoftOS10DeviceInterfaceGUID,
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
					IInterface:         0x05,
				},
				HID: &usb.HIDFunction{
					Descriptor: usb.HIDDescriptor{
						BcdHID:       0x0111,
						BCountryCode: 0x00,
						Descriptors: []usb.HIDSubDescriptor{
							{Type: usb.ReportDescType},
						},
					},
					ReportDescriptor: reportDescriptor,
				},
				Endpoints: []usb.EndpointDescriptor{
					{BEndpointAddress: EndpointHIDIn, BMAttributes: 0x03, WMaxPacketSize: 64, BInterval: 4},
					{BEndpointAddress: EndpointHIDOut, BMAttributes: 0x03, WMaxPacketSize: 64, BInterval: 4},
				},
			},
			{
				Descriptor: usb.InterfaceDescriptor{
					BInterfaceNumber:   0x01,
					BAlternateSetting:  0x00,
					BNumEndpoints:      0x02,
					BInterfaceClass:    0xFF,
					BInterfaceSubClass: 0x00,
					BInterfaceProtocol: 0x00,
					IInterface:         0x06,
				},
				Endpoints: []usb.EndpointDescriptor{
					{BEndpointAddress: EndpointBulkOut, BMAttributes: 0x02, WMaxPacketSize: 64, BInterval: 0},
					{BEndpointAddress: EndpointBulkIn, BMAttributes: 0x02, WMaxPacketSize: 64, BInterval: 0},
				},
			},
		},
		Strings: map[uint8]string{
			0: "\u0409",
			1: "Nintendo",
			2: "Switch 2 Pro Controller",
			3: DefaultSerial,
			4: "Nintendo Switch 2 Pro Controller",
			5: "Nintendo Switch 2 Pro Controller",
			6: "Pro Controller",
		},
	}
}

var reportDescriptor = hid.ReportDescriptor{
	Items: []hid.Item{
		hidShort(hid.ItemTypeGlobal, 0x0, 0x01),
		hidShort(hid.ItemTypeLocal, 0x0, 0x05),
		hidShort(hid.ItemTypeMain, 0xA, 0x01),
		hidShort(hid.ItemTypeGlobal, 0x8, ReportIDCommon),
		hidShort(hid.ItemTypeGlobal, 0x0, 0xFF),
		hidShort(hid.ItemTypeLocal, 0x0, 0x01),
		hidShort(hid.ItemTypeGlobal, 0x1, 0x00),
		hidShort(hid.ItemTypeGlobal, 0x2, 0xFF, 0x00),
		hidShort(hid.ItemTypeGlobal, 0x9, 0x3F),
		hidShort(hid.ItemTypeGlobal, 0x7, 0x08),
		hidShort(hid.ItemTypeMain, 0x8, 0x02),
		hidShort(hid.ItemTypeGlobal, 0x8, ReportIDPro),
		hidShort(hid.ItemTypeLocal, 0x0, 0x01),
		hidShort(hid.ItemTypeGlobal, 0x9, 0x02),
		hidShort(hid.ItemTypeMain, 0x8, 0x02),
		hidShort(hid.ItemTypeGlobal, 0x0, 0x09),
		hidShort(hid.ItemTypeLocal, 0x1, 0x01),
		hidShort(hid.ItemTypeLocal, 0x2, 0x15),
		hidShort(hid.ItemTypeGlobal, 0x2, 0x01),
		hidShort(hid.ItemTypeGlobal, 0x9, 0x15),
		hidShort(hid.ItemTypeGlobal, 0x7, 0x01),
		hidShort(hid.ItemTypeMain, 0x8, 0x02),
		hidShort(hid.ItemTypeGlobal, 0x9, 0x01),
		hidShort(hid.ItemTypeGlobal, 0x7, 0x03),
		hidShort(hid.ItemTypeMain, 0x8, 0x03),
		hidShort(hid.ItemTypeGlobal, 0x0, 0x01),
		hidShort(hid.ItemTypeLocal, 0x0, 0x01),
		hidShort(hid.ItemTypeMain, 0xA, 0x00),
		hidShort(hid.ItemTypeLocal, 0x0, 0x30),
		hidShort(hid.ItemTypeLocal, 0x0, 0x31),
		hidShort(hid.ItemTypeLocal, 0x0, 0x33),
		hidShort(hid.ItemTypeLocal, 0x0, 0x35),
		hidShort(hid.ItemTypeGlobal, 0x2, 0xFF, 0x0F),
		hidShort(hid.ItemTypeGlobal, 0x9, 0x04),
		hidShort(hid.ItemTypeGlobal, 0x7, 0x0C),
		hidShort(hid.ItemTypeMain, 0x8, 0x02),
		hidShort(hid.ItemTypeMain, 0xC),
		hidShort(hid.ItemTypeGlobal, 0x0, 0xFF),
		hidShort(hid.ItemTypeLocal, 0x0, 0x02),
		hidShort(hid.ItemTypeGlobal, 0x2, 0xFF, 0x00),
		hidShort(hid.ItemTypeGlobal, 0x9, 0x34),
		hidShort(hid.ItemTypeGlobal, 0x7, 0x08),
		hidShort(hid.ItemTypeMain, 0x8, 0x02),
		hidShort(hid.ItemTypeGlobal, 0x8, ReportIDOutput),
		hidShort(hid.ItemTypeLocal, 0x0, 0x01),
		hidShort(hid.ItemTypeGlobal, 0x9, 0x3F),
		hidShort(hid.ItemTypeMain, 0x9, 0x02),
		hidShort(hid.ItemTypeMain, 0xC),
	}}

func hidShort(itemType hid.ItemType, tag uint8, data ...uint8) hid.AnyItem {
	return hid.AnyItem{Type: itemType, Tag: tag, Data: hid.Data(data)}
}
