package ns2pro

const (
	DefaultVID    = 0x057E
	DefaultPID    = 0x2069
	DefaultSerial = "00"
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
	InputWireSize    = 27
	OutputRumbleSize = 32
	OutputWireSize   = 34
)

const (
	OutputFlagRumble = 0x01
	OutputFlagLED    = 0x02
)

const (
	StickMin     uint16 = 0
	StickCenter  uint16 = 0x0800
	StickMax     uint16 = 0x0FFF
	BatteryMax   uint8  = 9
	BatteryVolts uint16 = 3800
)

const (
	FeatureButtons = 0x01
	FeatureSticks  = 0x02
	FeatureIMU     = 0x04
	FeatureMouse   = 0x10
	FeatureRumble  = 0x20
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
