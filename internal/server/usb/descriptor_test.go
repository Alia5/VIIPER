package usb

import (
	"bytes"
	"encoding/binary"
	"testing"

	usbdesc "github.com/Alia5/VIIPER/usb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfigDescriptorSupportsIADAndAlternateSettings(t *testing.T) {
	desc := &usbdesc.Descriptor{
		Configuration: usbdesc.ConfigurationDescriptor{
			BConfigurationValue: 0x01,
			IConfiguration:      0x04,
			BMAttributes:        0xC0,
			BMaxPower:           0xFA,
		},
		Associations: []usbdesc.InterfaceAssociationDescriptor{
			{BFirstInterface: 0, BInterfaceCount: 1, BFunctionClass: 0x03},
			{BFirstInterface: 1, BInterfaceCount: 2, BFunctionClass: 0x01, BFunctionSubClass: 0x01},
		},
		Interfaces: []usbdesc.InterfaceConfig{
			{
				Descriptor: usbdesc.InterfaceDescriptor{
					BInterfaceNumber: 0, BNumEndpoints: 1, BInterfaceClass: 0x03,
				},
				Endpoints: []usbdesc.EndpointDescriptor{{BEndpointAddress: 0x81, BMAttributes: 0x03, WMaxPacketSize: 64, BInterval: 4}},
			},
			{
				Descriptor: usbdesc.InterfaceDescriptor{
					BInterfaceNumber: 1, BAlternateSetting: 0, BInterfaceClass: 0x01, BInterfaceSubClass: 0x02,
				},
			},
			{
				Descriptor: usbdesc.InterfaceDescriptor{
					BInterfaceNumber: 1, BAlternateSetting: 1, BNumEndpoints: 1, BInterfaceClass: 0x01, BInterfaceSubClass: 0x02,
				},
				Endpoints: []usbdesc.EndpointDescriptor{{
					BEndpointAddress: 0x03,
					BMAttributes:     0x0D,
					WMaxPacketSize:   0x00C0,
					BInterval:        1,
					ClassDescriptors: []usbdesc.ClassSpecificDescriptor{
						{DescriptorType: 0x25, Payload: usbdesc.Data{0x01, 0x00, 0x00, 0x00, 0x00}},
					},
				}},
			},
			{
				Descriptor: usbdesc.InterfaceDescriptor{
					BInterfaceNumber: 2, BAlternateSetting: 0, BInterfaceClass: 0x01, BInterfaceSubClass: 0x02,
				},
			},
		},
	}

	got := (&Server{}).buildConfigDescriptor(desc)
	require.NotEmpty(t, got)
	assert.Equal(t, uint16(len(got)), binary.LittleEndian.Uint16(got[2:4]))
	assert.Equal(t, byte(3), got[4])
	assert.Equal(t, byte(0x01), got[5])
	assert.Equal(t, byte(0x04), got[6])
	assert.Equal(t, byte(0xC0), got[7])
	assert.Equal(t, byte(0xFA), got[8])

	assert.Equal(t, []byte{0x08, 0x0B, 0x00, 0x01, 0x03, 0x00, 0x00, 0x00}, got[9:17])
	assert.True(t, bytes.Contains(got, []byte{0x07, 0x25, 0x01, 0x00, 0x00, 0x00, 0x00}))
}
