package dualsense

import "time"

const (
	DefaultVID       = 0x054C
	DefaultPIDDSEdge = 0x0DF2
	DefaultPIDDS     = 0x0CE6
)

const (
	DefaultMACAddressDSEdge   = "A5:FE:9C:CF:92:00" // Steam reads this as serial? // TODO: not detected by all apps
	DefaultSerialNumberDSEdge = "E55E00GTD1190A500" // Byte 6 (00) is "color code" will be replaced by MetaState
	DefaultBoardStringEdge    = "HMB-010"
)

const (
	DefaultMACAddressDS   = "A5:FA:9C:CF:92:00" // Steam reads this as serial? // TODO: not detected by all apps
	DefaultSerialNumberDS = "E55700GTD1190A500" // Byte 6 (00) is "color code" will be replaced by MetaState
	DefaultBoardStringDS  = "BDM-050"
)
const (
	DefaultBatteryStatus = BatteryFullyCharged
	DefaultTemperature   = 28.0
	DefaultVoltage       = 3.8
	DefaultShellColor    = ShellColorBlack
)

var DefaultBuildTime = time.Date(2025, time.July, 4, 10, 10, 32, 0, time.UTC)

const (
	EndpointIn  = 0x84
	EndpointOut = 0x03
)

const (
	ReportIDInput  = 0x01
	ReportIDOutput = 0x02
)

const (
	InputReportSize  = 64
	OutputReportSize = 64
	InputStateSize   = 33
	OutputStateSize  = 6
)

const (
	ButtonSquare   uint32 = 0x0010
	ButtonCross    uint32 = 0x0020
	ButtonCircle   uint32 = 0x0040
	ButtonTriangle uint32 = 0x0080

	ButtonL1      uint32 = 0x0100
	ButtonR1      uint32 = 0x0200
	ButtonL2      uint32 = 0x0400
	ButtonR2      uint32 = 0x0800
	ButtonCreate  uint32 = 0x1000
	ButtonOptions uint32 = 0x2000
	ButtonL3      uint32 = 0x4000
	ButtonR3      uint32 = 0x8000

	ButtonPS       uint32 = 0x00010000
	ButtonTouchpad uint32 = 0x00020000
	ButtonMicMute  uint32 = 0x00040000

	ButtonEdgeLFn uint32 = 0x00100000
	ButtonEdgeRFn uint32 = 0x00200000
	ButtonEdgeL4  uint32 = 0x00400000
	ButtonEdgeR4  uint32 = 0x00800000
)

const (
	DPadUp    = 0x01
	DPadDown  = 0x02
	DPadLeft  = 0x04
	DPadRight = 0x08
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

const DPadMask uint8 = 0x0F

// Gyro/Accel scale factors matching the USB report domain.
//
//	Gyro: BMI323 ±2000 dps passthrough = 16.384 LSB/dps
//	Accel: BMI323 4096 LSB/g × ScaleAccel(×2) = 8192 LSB/g = 835.07 LSB/(m/s²)
const (
	GyroCountsPerDps  = 16.384
	AccelCountsPerMS2 = 835.07 // 8192 / 9.81
)

const (
	DefaultAccelXRaw int16 = 0
	DefaultAccelYRaw int16 = 0
	DefaultAccelZRaw int16 = -8192 // -1g (8192 counts/g)
)

// Touchpad dimensions.
const (
	TouchpadMinX uint16 = 0
	TouchpadMinY uint16 = 0
	TouchpadMaxX uint16 = 1920
	TouchpadMaxY uint16 = 1080

	TouchInactiveMask uint8 = 0x80
)

const DeltaTimeNS = 333

const (
	BatteryFullyCharged = 0x2A // Status=0x2 (Full), Level=0xA (100%)
)

const (
	hidClassIN  uint8 = 0xA1
	hidClassOUT uint8 = 0x21
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
	featureIDCalibration     uint8 = 0x05
	featureIDPairing         uint8 = 0x09
	featureIDFirmware        uint8 = 0x20
	featureIDCommand         uint8 = 0x80
	featureIDCommandResponse uint8 = 0x81
)

const (
	subcmdSerial  uint8 = 0x01
	subcmdStatus  uint8 = 0x03
	subcmdSensors uint8 = 0x04
)

const (
	HardwareType    uint8  = 0x03
	HwInfo          uint32 = 0x01000208
	FirmwareVersion uint16 = 0x0630
)

const (
	ShellColorWhite                  = "00"
	ShellColorBlack                  = "01"
	ShellColorCosmicRed              = "02"
	ShellColorNovaPink               = "03"
	ShellColorGalacticPurple         = "04"
	ShellColorStarlightBlue          = "05"
	ShellColorGreyCamouflage         = "06"
	ShellColorVolcanicRed            = "07"
	ShellColorSterlingSilver         = "08"
	ShellColorCobaltBlue             = "09"
	ShellColorChromaTeal             = "10"
	ShellColorChromaIndigo           = "11"
	ShellColorChromaPearl            = "12"
	ShellColorAnniversary30th        = "30"
	ShellColorGodOfWarRagnarok       = "Z1"
	ShellColorSpiderMan2             = "Z2"
	ShellColorAstroBot               = "Z3"
	ShellColorFortnite               = "Z4"
	ShellColorMonsterHunterWilds     = "Z5"
	ShellColorTheLastOfUs            = "Z6"
	ShellColorGhostOfYotei           = "Z7"
	ShellColorIconBlueLimitedEdition = "ZB"
	ShellColorAstroBotJoyfulEdition  = "ZC"
	ShellColorGenshinImpact          = "ZE"
)
