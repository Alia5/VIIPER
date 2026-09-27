package dualshock4

import "time"

const (
	DefaultVID = 0x054C
	DefaultPID = 0x09CC
)

const (
	DefaultSerialString  = "1111020BF619A500"
	DefaultBoardString   = "JDM-055"
	DefaultTemperature   = 28.0
	DefaultVoltage       = 4.2
	DefaultBatteryStatus = BatteryChargingFlag | BatteryFullyCharged
)

var DefaultBuildTime = time.Date(2021, time.September, 17, 11, 34, 0, 0, time.UTC)

const (
	EndpointIn  = 0x84
	EndpointOut = 0x03
)

const (
	ReportIDInput  = 0x01
	ReportIDOutput = 0x05
)

const (
	UpdateFlagRumble   uint8 = 0x01
	UpdateFlagLED      uint8 = 0x02
	UpdateFlagLEDBlink uint8 = 0x04
)

const (
	InputReportSize = 64
)

const (
	ButtonSquare   uint16 = 0x0010
	ButtonCross    uint16 = 0x0020
	ButtonCircle   uint16 = 0x0040
	ButtonTriangle uint16 = 0x0080

	DPadMask uint8 = 0x0F
)

const (
	ButtonL1      uint16 = 0x0100
	ButtonR1      uint16 = 0x0200
	ButtonL2      uint16 = 0x0400
	ButtonR2      uint16 = 0x0800
	ButtonShare   uint16 = 0x1000
	ButtonOptions uint16 = 0x2000
	ButtonL3      uint16 = 0x4000
	ButtonR3      uint16 = 0x8000

	ButtonPS            uint16 = 0x0001
	ButtonTouchpadClick uint16 = 0x0002
)

const (
	ButtonPSUSB            uint8 = 0x01
	ButtonTouchpadClickUSB uint8 = 0x02

	CounterShift = 2
)

const (
	DPadUSBUp        = 0x00
	DPadUSBUpRight   = 0x01
	DPadUSBRight     = 0x02
	DPadUSBDownRight = 0x03
	DPadUSBDown      = 0x04
	DPadUSBDownLeft  = 0x05
	DPadUSBLeft      = 0x06
	DPadUSBUpLeft    = 0x07
	DPadUSBNeutral   = 0x08
)

const (
	DPadUp    = 0x01
	DPadDown  = 0x02
	DPadLeft  = 0x04
	DPadRight = 0x08
)

// The DS4 USB input report carries gyro/accel as signed int16 values.
// VIIPER's wire protocol keeps them as int16, but interprets them as fixed-point
// physical units to avoid float serialization across clients.
//
// Gyro fields (GyroX/Y/Z): °/s scaled by GyroCountsPerDps.
// Accel fields (AccelX/Y/Z): m/s² scaled by AccelCountsPerMS2.
const (
	// GyroCountsPerDps is the fixed-point scale factor for °/s.
	// resolution is 0.0625 °/s and range is about +-2048 °/s.
	GyroCountsPerDps = 16.0

	// AccelCountsPerMS2 is the fixed-point scale factor for m/s².
	// resolution is ~0.00195 m/s² and range is about +-64 m/s² (~+-6.5 g).
	AccelCountsPerMS2 = 512.0

	StandardGravityMS2 = 9.81
)

// Default accelerometer raw values for a controller lying flat on a table.
const (
	DefaultAccelXRaw int16 = 0
	DefaultAccelYRaw int16 = 0
	// -StandardGravityMS2 * AccelCountsPerMS2 = (-9.81 * 512) = -5023
	DefaultAccelZRaw int16 = -5023
)

const (
	DefaultLedRed   = 0x00
	DefaultLedGreen = 0x00
	DefaultLedBlue  = 0x40
)

const (
	TouchpadMinX uint16 = 0
	TouchpadMaxX uint16 = 1920
	TouchpadMinY uint16 = 0
	TouchpadMaxY uint16 = 942

	TouchInactiveMask uint8 = 0x80
)

const (
	BatteryLevelMask    = 0x0F
	BatteryChargingFlag = 0x10
	BatteryFullyCharged = 0x0A
)
const (
	HardwareVersionMajor uint16 = 0x0001
	HardwareVersionMinor uint16 = 0xB400
	SoftwareVersionMajor uint32 = 0x00000001
	SoftwareVersionMinor uint16 = 0xA00B
)

const (
	hidClassIN  uint8 = 0xA1 // bmRequestType: HID class IN  (device→host)
	hidClassOUT uint8 = 0x21 // bmRequestType: HID class OUT (host→device)
)

const (
	hidGetReport   uint8 = 0x01
	hidGetIdle     uint8 = 0x02
	hidGetProtocol uint8 = 0x03
	hidSetReport   uint8 = 0x09
)

const (
	reportTypeInput   uint8 = 0x01
	reportTypeOutput  uint8 = 0x02
	reportTypeFeature uint8 = 0x03
)

const (
	featureIDCalibration   byte = 0x02
	featureIDCapabilities  byte = 0x03
	featureIDCalibrationBT byte = 0x05
	featureIDProbe         byte = 0x08
	featureIDStatus        byte = 0x10
	featureIDProbeResponse byte = 0x11
	featureIDSerial        byte = 0x12
	featureIDIdentity      byte = 0x81
	featureIDSubcommand    byte = 0xA0
	featureIDBoardInfo     byte = 0xA3
	featureIDTelemetry     byte = 0xA4 // serial or voltage/temperature; content selected by featureIDSubcommand
)
