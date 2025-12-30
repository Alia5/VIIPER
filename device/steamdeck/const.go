package steamdeck

const (
	ValveUSBVID = 0x28DE
	JupiterPID  = 0x1205
)

// (ApiCLient) Stream frame sizes.
const (
	InputStateSize  = 52
	HapticStateSize = 4
)

// Valve/Steam Deck input report (interrupt IN) constants.
//
// SDL`controller_structs.h`.
const (
	ValveInReportMsgVersion uint16 = 0x0001

	ValveInReportTypeControllerDeckState uint8 = 0x09
	ValveInReportLength                  uint8 = 64

	ValveInReportHeaderSize   = 4
	ValveInReportPacketNumOff = 4
	ValveInReportPayloadOff   = 8
)

// Steam Deck feature report sizing.
// SDL uses 65-byte buffers on Windows: ReportID (0) + 64 bytes payload.
const (
	HIDFeatureReportBytes   = 64
	HIDFeatureReportWinSize = 65
)

// Feature Report Message IDs
const (
	FeatureIDSetDigitalMappings        = 0x80
	FeatureIDClearDigitalMappings      = 0x81
	FeatureIDGetAttributesValues       = 0x83
	FeatureIDSetDefaultDigitalMappings = 0x85
	FeatureIDFactoryReset              = 0x86
	FeatureIDSetSettingsValues         = 0x87
	FeatureIDClearSettingsValues       = 0x88
	FeatureIDSetControllerMode         = 0x8D
	FeatureIDLoadDefaultSettings       = 0x8E
	FeatureIDTriggerHapticPulse        = 0x8F
	FeatureIDTurnOffController         = 0x9F
	FeatureIDGetStringAttribute        = 0xAE
	FeatureIDCalibrateGyro             = 0xB5
	FeatureIDPlayAudio                 = 0xB6
	FeatureIDCalibrateJoystick         = 0xBF
	FeatureIDCalibrateAnalogTriggers   = 0xC0
	FeatureIDSetAudioMapping           = 0xC1
	FeatureIDResetIMU                  = 0xCE
	FeatureIDTriggerHapticCmd          = 0xEA
	FeatureIDTriggerRumbleCmd          = 0xEB
)

// Haptic types
const (
	HapticTypeOff      = 0
	HapticTypeTick     = 1
	HapticTypeClick    = 2
	HapticTypeTone     = 3
	HapticTypeRumble   = 4
	HapticTypeNoise    = 5
	HapticTypeScript   = 6
	HapticTypeLogSweep = 7
)

// Haptic sides
const (
	HapticSideLeft  = 0x01
	HapticSideRight = 0x02
	HapticSideBoth  = 0x03
)

// Haptic intensity levels
const (
	HapticIntensitySystem = 0
	HapticIntensityShort  = 1
	HapticIntensityMedium = 2
	HapticIntensityLong   = 3
	HapticIntensityInsane = 4
)

// Haptic pulse priority
const (
	HapticPulseNormal           = 0x00
	HapticPulseHighPriority     = 0x01
	HapticPulseVeryHighPriority = 0x02
	HapticPulseIgnoreUserPrefs  = 0x03
)

// Audio cues
const (
	AudioStartup     = 0
	AudioShutdown    = 1
	AudioPair        = 2
	AudioPairSuccess = 3
	AudioIdentify    = 4
	AudioLizardMode  = 5
	AudioNormalMode  = 6
)

// ControllerAttributes enum values (SDL `ControllerAttributes`).
const (
	AttribUniqueID        = 0
	AttribProductID       = 1
	AttribCapabilities    = 2
	AttribFirmwareVersion = 3
	AttribFirmwareBuild   = 4
)

// ControllerStringAttributes enum values (SDL `ControllerStringAttributes`).
const (
	AttribStrBoardSerial = 0
	AttribStrUnitSerial  = 1
)

// Steam Deck button bitmasks.
//
// Values from SDL's `SDL_hidapi_steamdeck.c`.
const (
	ButtonR2 uint64 = 0x00000001
	ButtonL2 uint64 = 0x00000002
	ButtonRB uint64 = 0x00000004
	ButtonLB uint64 = 0x00000008

	ButtonY uint64 = 0x00000010
	ButtonB uint64 = 0x00000020
	ButtonX uint64 = 0x00000040
	ButtonA uint64 = 0x00000080

	ButtonDPadUp    uint64 = 0x00000100
	ButtonDPadRight uint64 = 0x00000200
	ButtonDPadLeft  uint64 = 0x00000400
	ButtonDPadDown  uint64 = 0x00000800

	ButtonView  uint64 = 0x00001000
	ButtonSteam uint64 = 0x00002000
	ButtonMenu  uint64 = 0x00004000

	ButtonL5 uint64 = 0x00008000
	ButtonR5 uint64 = 0x00010000

	ButtonLeftPadClick  uint64 = 0x00020000
	ButtonRightPadClick uint64 = 0x00040000

	ButtonL3 uint64 = 0x00400000
	ButtonR3 uint64 = 0x04000000

	// High 32-bit button flags (ulButtonsH) shifted into the uint64.
	ButtonL4  uint64 = 0x00000200 << 32
	ButtonR4  uint64 = 0x00000400 << 32
	ButtonQAM uint64 = 0x00040000 << 32
)
