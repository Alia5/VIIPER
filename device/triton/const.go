// Package triton provides a minimal Valve Steam Triton Controller (wired) HID implementation.
package triton

const (
	ValveUSBVID    = 0x28DE
	TritonWiredPID = 0x1302
	TritonBLEPID   = 0x1303
)

// (ApiClient) Stream frame sizes.
const (
	InputStateSize  = 52
	HapticStateSize = 10
)

// HID Input Report IDs (device → host, via interrupt IN endpoint).
// SDL: `controller_structs.h` ETritonReportIDTypes.
const (
	InReportControllerState    uint8 = 0x42
	InReportBatteryStatus      uint8 = 0x43
	InReportControllerStateBLE uint8 = 0x45
	InReportWirelessStatusX    uint8 = 0x46
	InReportWirelessStatus     uint8 = 0x79
)

// HID Output Report IDs (host → device, via interrupt OUT endpoint / hid_write).
// SDL: `controller_structs.h` ValveTritonOutReportMessageIDs.
const (
	OutReportHapticRumble   uint8 = 0x80
	OutReportHapticPulse    uint8 = 0x81
	OutReportHapticCommand  uint8 = 0x82
	OutReportHapticLFOTone  uint8 = 0x83
	OutReportHapticLogSweep uint8 = 0x85
	OutReportHapticScript   uint8 = 0x86
)

// HID Feature Report ID (settings, attributes).
const FeatureReportID uint8 = 0x01

// Feature Report Message IDs (same as Steam Deck / old Steam Controller).
const (
	FeatureIDGetAttributesValues  = 0x83
	FeatureIDSetSettingsValues    = 0x87
	FeatureIDClearDigitalMappings = 0x81
	FeatureIDLoadDefaultSettings  = 0x8E
	FeatureIDGetStringAttribute   = 0xAE
)

// Valve controller attribute tags.
const (
	AttribUniqueID        uint8 = 0x00
	AttribProductID       uint8 = 0x01
	AttribCapabilities    uint8 = 0x02
	AttribFirmwareVersion uint8 = 0x03
	AttribFirmwareBuild   uint8 = 0x04
)

// String attribute tags.
const (
	AttribStrBoardSerial uint8 = 0x00
	AttribStrUnitSerial  uint8 = 0x01
)

// Button bitmask values for TritonMTUNoQuat_t.buttons (uint32).
// SDL: `SDL_hidapi_steam_triton.c` TritonButtons enum.
const (
	ButtonA         uint32 = 0x00000001
	ButtonB         uint32 = 0x00000002
	ButtonX         uint32 = 0x00000004
	ButtonY         uint32 = 0x00000008
	ButtonQAM       uint32 = 0x00000010
	ButtonR3        uint32 = 0x00000020
	ButtonView      uint32 = 0x00000040 // Select / Back
	ButtonR4        uint32 = 0x00000080
	ButtonR5        uint32 = 0x00000100
	ButtonRB        uint32 = 0x00000200
	ButtonDPadDown  uint32 = 0x00000400
	ButtonDPadRight uint32 = 0x00000800
	ButtonDPadLeft  uint32 = 0x00001000
	ButtonDPadUp    uint32 = 0x00002000
	ButtonMenu      uint32 = 0x00004000 // Start
	ButtonL3        uint32 = 0x00008000
	ButtonSteam     uint32 = 0x00010000
	ButtonL4        uint32 = 0x00020000
	ButtonL5        uint32 = 0x00040000
	ButtonLB        uint32 = 0x00080000
)

const HIDFeatureReportBytes = 64
