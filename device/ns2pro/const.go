package ns2pro

const (
	DefaultVID                 = 0x057E
	DefaultPID                 = 0x2069
	DefaultSerialEnding        = "00"
	DefaultSerial              = "VIIPER-NS2PRO-" + DefaultSerialEnding
	DefaultBatteryVolts uint16 = 3800
)

const (
	ButtonB uint32 = 1 << iota
	ButtonA
	ButtonY
	ButtonX
	ButtonR
	ButtonZR
	ButtonPlus
	ButtonRightStick
	ButtonDown
	ButtonRight
	ButtonLeft
	ButtonUp
	ButtonL
	ButtonZL
	ButtonMinus
	ButtonLeftStick
	ButtonHome
	ButtonCapture
	ButtonGR
	ButtonGL
	ButtonC
	ButtonHeadset
)

const (
	OutputFlagRumble = 0x01
	OutputFlagLED    = 0x02
)

const (
	StickMin    uint16 = 0
	StickCenter uint16 = 0x0800
	StickMax    uint16 = 0x0FFF
	BatteryMax  uint8  = 9
)

const (
	EndpointHIDIn   = 0x81
	EndpointHIDOut  = 0x01
	EndpointBulkOut = 0x02
	EndpointBulkIn  = 0x82
)

const (
	ReportIDCommon = 0x05
	ReportIDPro    = 0x09
	ReportIDOutput = 0x02
)

const (
	InputReportSize  = 64
	OutputReportSize = 64
	InputWireSize    = 24
	OutputRumbleSize = 32
	OutputWireSize   = 34
)

const (
	FeatureButtons = 0x01
	FeatureSticks  = 0x02
	FeatureIMU     = 0x04
	FeatureMouse   = 0x10
	FeatureRumble  = 0x20
)

const (
	hidClassRequestIn  = 0xA1
	hidClassRequestOut = 0x21

	hidGetReport = 0x01
	hidSetReport = 0x09

	reportTypeInput  = 0x01
	reportTypeOutput = 0x02
)

const (
	audioSetCur = 0x01
	audioGetCur = 0x81
	audioGetMin = 0x82
	audioGetMax = 0x83
	audioGetRes = 0x84
)

const (
	requestTypeMask   = 0x60
	requestClass      = 0x20
	recipientMask     = 0x1F
	recipientIface    = 0x01
	recipientEndpoint = 0x02
)

const (
	cmdFlash     = 0x02
	cmdUSB       = 0x03
	cmdPlayerLED = 0x09
	cmdFeature   = 0x0C
)

const (
	subFlashRead = 0x01

	subUSBEnableReports = 0x03
	subUSBSelectReport  = 0x0A
	subUSBStartReports  = 0x0D

	subFeatureInfo    = 0x01
	subFeatureSetMask = 0x02
	subFeatureReset   = 0x03
	subFeatureEnable  = 0x04
	subFeatureDisable = 0x05

	subPlayerLEDSet = 0x07
)

const flashBlockSize = 0x40
