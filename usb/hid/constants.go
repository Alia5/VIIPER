package hid

// Common Usage Pages.
// Values per HID Usage Tables.
const (
	UsagePageGenericDesktop uint16 = 0x01
	UsagePageSimulation     uint16 = 0x02
	UsagePageVR             uint16 = 0x03
	UsagePageSport          uint16 = 0x04
	UsagePageGame           uint16 = 0x05
	UsagePageKeyboard       uint16 = 0x07
	UsagePageLEDs           uint16 = 0x08
	UsagePageButton         uint16 = 0x09
	UsagePageConsumer       uint16 = 0x0C
)

// Generic Desktop usages.
const (
	UsagePointer  uint16 = 0x01
	UsageMouse    uint16 = 0x02
	UsageJoystick uint16 = 0x04
	UsageGamePad  uint16 = 0x05
	UsageKeyboard uint16 = 0x06
	UsageX        uint16 = 0x30
	UsageY        uint16 = 0x31
	UsageZ        uint16 = 0x32
	UsageRx       uint16 = 0x33
	UsageRy       uint16 = 0x34
	UsageRz       uint16 = 0x35
	UsageWheel    uint16 = 0x38
)

// Consumer usages.
const (
	UsageACPan uint16 = 0x0238
)

// CollectionKind values.
type CollectionKind uint8

const (
	CollectionPhysical    CollectionKind = 0x00
	CollectionApplication CollectionKind = 0x01
	CollectionLogical     CollectionKind = 0x02
)

type MainFlags uint8

const (
	MainData  MainFlags = 0x00
	MainConst MainFlags = 0x01

	MainArray MainFlags = 0x00
	MainVar   MainFlags = 0x02

	MainAbs MainFlags = 0x00
	MainRel MainFlags = 0x04

	MainNoWrap MainFlags = 0x00
	MainWrap   MainFlags = 0x08

	MainLinear    MainFlags = 0x00
	MainNonLinear MainFlags = 0x10

	MainPreferredState   MainFlags = 0x00
	MainNoPreferredState MainFlags = 0x20

	MainNoNullPosition MainFlags = 0x00
	MainNullState      MainFlags = 0x40

	MainNonVolatile MainFlags = 0x00
	MainVolatile    MainFlags = 0x80
)
