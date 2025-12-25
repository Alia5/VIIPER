// Package steamdeck provides a minimal Steam Deck (Jupiter/LCD) controller HID implementation.
package steamdeck

import (
	"encoding/binary"
	"sync"
	"sync/atomic"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
	"github.com/Alia5/VIIPER/usbip"
)

type SteamDeck struct {
	tick uint64

	stateMu    sync.Mutex
	inputState *InputState

	featureMu       sync.Mutex
	featureResponse []byte // cached 65-byte (or 64-byte) feature response

	hapticFunc func(HapticState)
	descriptor usb.Descriptor
}

// New returns a new SteamDeck Controller device.
func New(o *device.CreateOptions) *SteamDeck {
	d := &SteamDeck{
		descriptor: defaultDescriptor,
	}
	// Is any1 ever gonna use a patched SDL for this? Ignore custom VID/PIDS for now
	// if o != nil {
	// 	if o.IdVendor != nil {
	// 		d.descriptor.Device.IDVendor = *o.IdVendor
	// 	}
	// 	if o.IdProduct != nil {
	// 		d.descriptor.Device.IDProduct = *o.IdProduct
	// 	}
	// }
	return d
}

// SetRumbleCallback sets a callback that will be invoked when rumble commands arrive.
func (s *SteamDeck) SetRumbleCallback(f func(HapticState)) {
	s.hapticFunc = f
}

// UpdateInputState updates the controller's current input state (thread-safe).
//
// The latest state is used to build the 64-byte interrupt IN report.
func (s *SteamDeck) UpdateInputState(st InputState) {
	s.stateMu.Lock()
	if s.inputState == nil {
		s.inputState = &InputState{}
	}
	*s.inputState = st
	s.stateMu.Unlock()
}

func (s *SteamDeck) getInputStateSnapshot() InputState {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.inputState == nil {
		return InputState{}
	}
	return *s.inputState
}

// HandleTransfer implements interrupt IN for the Steam Deck controller interface.
func (s *SteamDeck) HandleTransfer(ep uint32, dir uint32, _ []byte) []byte {
	if dir != usbip.DirIn {
		return nil
	}
	switch ep {
	case 2: // 0x82 - dummy keyboard interface (iface 0)
		// Keep Windows HID stack happy
		// no actual lizard-mode behavior.
		atomic.AddUint64(&s.tick, 1)
		return make([]byte, 8)
	case 3: // 0x83 - dummy mouse interface (iface 1)
		atomic.AddUint64(&s.tick, 1)
		return make([]byte, 4)
	case 1: // 0x81 - controller input reports
		seq := uint32(atomic.AddUint64(&s.tick, 1))
		st := s.getInputStateSnapshot()
		return buildInReport(seq, st)
	default:
		return nil
	}
}

// HandleControl implements EP0 class requests for HID feature reports.
//
// The Steam Deck Controlelr uses feature reports to disable lizard mode, feed
// a watchdog, and trigger rumble.
func (s *SteamDeck) HandleControl(bmRequestType, bRequest uint8, wValue, wIndex, wLength uint16, data []byte) ([]byte, bool) {
	const (
		hidReqGetReport = 0x01
		hidReqSetReport = 0x09

		hidReqTypeClassInToInterface  = 0xA1
		hidReqTypeClassOutToInterface = 0x21

		hidReportTypeInput   = 0x01
		hidReportTypeOutput  = 0x02
		hidReportTypeFeature = 0x03
	)

	// Steam / SDL expect the controller interface iface 2
	// Our descriptor includes placeholder interfaces for kb/M and the actual controller
	// HID interface at idx 2.
	iface := uint8(wIndex & 0xFF)

	reportType := uint8(wValue >> 8)
	// reportID := uint8(wValue & 0xFF) // We don't use report IDs (but may see 0).
	_ = uint8(wValue & 0xFF)

	// HID-class handling for the dummy interfaces so Windows doesn't flag them as malfunctioning.
	if iface == 0 || iface == 1 {
		want := int(wLength)
		switch {
		case bmRequestType == hidReqTypeClassInToInterface && bRequest == hidReqGetReport:
			if want <= 0 {
				return nil, true
			}
			return make([]byte, want), true
		case bmRequestType == hidReqTypeClassOutToInterface && (bRequest == hidReqSetReport || bRequest == 0x0A || bRequest == 0x0B):
			return nil, true
		case bmRequestType == hidReqTypeClassInToInterface && (bRequest == 0x02 || bRequest == 0x03):
			if want <= 0 {
				return nil, true
			}
			resp := make([]byte, want)
			if bRequest == 0x03 {
				resp[0] = 0x01
			}
			return resp, true
		}
		return nil, false
	}

	if iface != 2 {
		return nil, false
	}

	switch {
	case bmRequestType == hidReqTypeClassInToInterface && bRequest == hidReqGetReport:
		want := int(wLength)
		if want <= 0 {
			return nil, true
		}

		switch reportType {
		case hidReportTypeInput:
			st := s.getInputStateSnapshot()
			report := buildInReport(0, st)
			if want == 65 {
				resp := make([]byte, 65)
				resp[0] = 0x00
				copy(resp[1:], report)
				return resp, true
			}
			return report, true
		case hidReportTypeFeature:
			return s.getFeatureResponse(want), true
		case hidReportTypeOutput:
			return make([]byte, want), true
		default:
			return make([]byte, want), true
		}

	case bmRequestType == hidReqTypeClassOutToInterface && bRequest == hidReqSetReport:
		if reportType == hidReportTypeFeature {
			s.handleFeatureReport(data)
			return nil, true
		}
		// Ignore other report types for now, but ACK them
		return nil, true
	}

	return nil, false
}

var dummyKeyboardReportDescriptor = hid.Report{
	Items: []hid.Item{
		hid.UsagePage{Page: hid.UsagePageGenericDesktop},
		hid.Usage{Usage: hid.UsageKeyboard},
		hid.Collection{Kind: hid.CollectionApplication, Items: []hid.Item{
			hid.UsagePage{Page: hid.UsagePageKeyboard},
			hid.UsageMinimum{Min: 0xE0},
			hid.UsageMaximum{Max: 0xE7},
			hid.LogicalMinimum{Min: 0},
			hid.LogicalMaximum{Max: 1},
			hid.ReportSize{Bits: 1},
			hid.ReportCount{Count: 8},
			hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			hid.ReportSize{Bits: 8},
			hid.ReportCount{Count: 1},
			hid.Input{Flags: hid.MainConst},

			hid.LogicalMinimum{Min: 0},
			hid.LogicalMaximum{Max: 255},
			hid.UsageMinimum{Min: 0x00},
			hid.UsageMaximum{Max: 0xFF},
			hid.ReportSize{Bits: 8},
			hid.ReportCount{Count: 6},
			hid.Input{Flags: hid.MainData | hid.MainArray | hid.MainAbs},
		}},
	},
}

var dummyMouseReportDescriptor = hid.Report{
	Items: []hid.Item{
		hid.UsagePage{Page: hid.UsagePageGenericDesktop},
		hid.Usage{Usage: hid.UsageMouse},
		hid.Collection{Kind: hid.CollectionApplication, Items: []hid.Item{
			hid.Usage{Usage: hid.UsagePointer},
			hid.Collection{Kind: hid.CollectionPhysical, Items: []hid.Item{
				hid.UsagePage{Page: hid.UsagePageButton},
				hid.UsageMinimum{Min: 0x01},
				hid.UsageMaximum{Max: 0x03},
				hid.LogicalMinimum{Min: 0},
				hid.LogicalMaximum{Max: 1},
				hid.ReportSize{Bits: 1},
				hid.ReportCount{Count: 3},
				hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
				hid.ReportSize{Bits: 5},
				hid.ReportCount{Count: 1},
				hid.Input{Flags: hid.MainConst},

				hid.UsagePage{Page: hid.UsagePageGenericDesktop},
				hid.Usage{Usage: hid.UsageX},
				hid.Usage{Usage: hid.UsageY},
				hid.Usage{Usage: hid.UsageWheel},
				hid.LogicalMinimum{Min: -127},
				hid.LogicalMaximum{Max: 127},
				hid.ReportSize{Bits: 8},
				hid.ReportCount{Count: 3},
				hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainRel},
			}},
		}},
	},
}

func (s *SteamDeck) getFeatureResponse(want int) []byte {
	if want <= 0 {
		return nil
	}

	s.featureMu.Lock()
	resp := append([]byte(nil), s.featureResponse...)
	s.featureMu.Unlock()

	if len(resp) == 0 {
		return make([]byte, want)
	}

	// If host asks for 64 bytes but we have a 65-byte report (ReportID + payload) return the payload
	if want == 64 && len(resp) == 65 {
		return append([]byte(nil), resp[1:]...)
	}
	if len(resp) >= want {
		return append([]byte(nil), resp[:want]...)
	}
	out := make([]byte, want)
	copy(out, resp)
	return out
}

func (s *SteamDeck) setFeatureResponse(resp []byte) {
	s.featureMu.Lock()
	defer s.featureMu.Unlock()
	if resp == nil {
		s.featureResponse = nil
		return
	}
	s.featureResponse = append([]byte(nil), resp...)
}

func (s *SteamDeck) handleFeatureReport(data []byte) {
	if len(data) == 65 && data[0] == 0x00 {
		data = data[1:]
	}
	if len(data) < 2 {
		return
	}

	msgType := data[0]
	// msgLen := data[1]
	_ = data[1]

	// If this request expects a response, queue it so the following GET_REPORT(feature)
	// returns exactly what SDL's ReadResponse() expects.
	switch msgType {
	case FeatureIDGetAttributesValues:
		s.setFeatureResponse(buildGetAttributesResponse())
	case FeatureIDGetStringAttribute:
		var tag uint8 = AttribStrUnitSerial
		if len(data) >= 3 {
			// FeatureReportMsg header is [type,length], so first payload byte is data[2]
			tag = data[2]
		}
		s.setFeatureResponse(buildGetStringAttributeResponse(tag))
	case FeatureIDClearDigitalMappings, FeatureIDSetSettingsValues, FeatureIDLoadDefaultSettings:
		// No-op but report protocol-level success; queue an empty response matching
		// the request ID so a follow-up GET_FEATURE doesn't fuck up SDL.
		s.setFeatureResponse(buildEmptyResponse(msgType))
	}

	if msgType != FeatureIDTriggerRumbleCmd {
		return
	}

	// FeatureReportMsg:
	//   [0]=type, [1]=length, [2..]=payload
	// MsgSimpleRumbleCmd payload starts at offset 2:
	//   [2]=unRumbleType
	//   [3:5]=unIntensity
	//   [5:7]=unLeftMotorSpeed
	//   [7:9]=unRightMotorSpeed
	if len(data) < 9 {
		return
	}
	left := binary.LittleEndian.Uint16(data[5:7])
	right := binary.LittleEndian.Uint16(data[7:9])

	if s.hapticFunc != nil {
		s.hapticFunc(HapticState{LeftMotor: left, RightMotor: right})
	}
}

func buildEmptyResponse(msgType uint8) []byte {
	resp := make([]byte, 65)
	resp[0] = 0x00
	resp[1] = msgType
	resp[2] = 0 // header.length
	return resp
}

func buildFeatureResponse(msgType uint8, payload []byte) []byte {
	resp := make([]byte, 65)
	resp[0] = 0x00
	resp[1] = msgType
	if payload == nil {
		resp[2] = 0
		return resp
	}
	if len(payload) > 62 {
		payload = payload[:62]
	}
	resp[2] = uint8(len(payload))
	copy(resp[3:], payload)
	return resp
}

func buildGetAttributesResponse() []byte {
	// copied from real device
	const (
		productID         = uint32(0x1205)
		capabilities      = uint32(0x0160BFFF)
		firmwareVersion   = uint32(1752616979)
		firmwareBuildTime = uint32(1752616979)
	)

	payload := make([]byte, 0, 5*5)
	appendAttr := func(tag uint8, value uint32) {
		b := make([]byte, 5)
		b[0] = tag
		binary.LittleEndian.PutUint32(b[1:5], value)
		payload = append(payload, b...)
	}
	appendAttr(AttribUniqueID, 0)
	appendAttr(AttribProductID, productID)
	appendAttr(AttribCapabilities, capabilities)
	appendAttr(AttribFirmwareVersion, firmwareVersion)
	appendAttr(AttribFirmwareBuild, firmwareBuildTime)

	return buildFeatureResponse(FeatureIDGetAttributesValues, payload)
}

func buildGetStringAttributeResponse(requestedTag uint8) []byte {
	// MsgGetStringAttribute: { uint8 attributeTag; char attributeValue[20]; }
	payload := make([]byte, 21)
	payload[0] = requestedTag

	const serial = "VIIPER000001"
	copy(payload[1:], []byte(serial))

	return buildFeatureResponse(FeatureIDGetStringAttribute, payload)
}

func buildInReport(packetNum uint32, st InputState) []byte {
	buf := make([]byte, 64)
	// ValveInReportHeader_t (packed):
	//   uint16 unReportVersion (LE)
	//   uint8  ucType
	//   uint8  ucLength
	buf[0] = byte(ValveInReportMsgVersion)
	buf[1] = byte(ValveInReportMsgVersion >> 8)
	buf[2] = ValveInReportTypeControllerDeckState
	buf[3] = ValveInReportLength
	binary.LittleEndian.PutUint32(buf[4:8], packetNum)

	payload, _ := st.MarshalBinary() // fixed-size, cannot fail
	copy(buf[8:], payload)
	return buf
}

var defaultDescriptor = usb.Descriptor{
	Device: usb.DeviceDescriptor{
		BcdUSB:             0x0200,
		BDeviceClass:       0x00,
		BDeviceSubClass:    0x00,
		BDeviceProtocol:    0x00,
		BMaxPacketSize0:    0x40,
		IDVendor:           ValveUSBVID,
		IDProduct:          JupiterPID,
		BcdDevice:          0x0200,
		IManufacturer:      0x01,
		IProduct:           0x02,
		ISerialNumber:      0x03,
		BNumConfigurations: 0x01,
		Speed:              2, // Full speed
	},
	// Steam's SDL Steam Controller driver expects the controller interface to be
	// interface 2.
	// add two dummy interfaces and ONLY emulate the actual
	// "controller" *cough* device at iface 2
	// We can SKIP the actual "lizard-mode kb/m (dummies)"
	// We can also SKIP the CDC interfaces
	// they are (AFAIK) only used for FW updates
	Interfaces: []usb.InterfaceConfig{
		{
			// Placeholder interface 0
			// IS needed for steam to open the device
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x00,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x01,
				BInterfaceClass:    0x03,
				BInterfaceSubClass: 0x01,
				BInterfaceProtocol: 0x01,
				IInterface:         0x00,
			},
			HID: &usb.HIDFunction{
				Descriptor: usb.HIDDescriptor{
					BcdHID:       0x0111,
					BCountryCode: 0x00,
					Descriptors:  []usb.HIDSubDescriptor{{Type: usb.ReportDescType}},
				},
				Report: dummyKeyboardReportDescriptor,
			},
			Endpoints: []usb.EndpointDescriptor{
				{
					BEndpointAddress: 0x82,
					BMAttributes:     0x03,
					WMaxPacketSize:   0x0008,
					BInterval:        0x0A,
				},
			},
		},
		{
			// Placeholder interface 1
			// IS needed for steam to open the device
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x01,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x01,
				BInterfaceClass:    0x03,
				BInterfaceSubClass: 0x01,
				BInterfaceProtocol: 0x02,
				IInterface:         0x00,
			},
			HID: &usb.HIDFunction{
				Descriptor: usb.HIDDescriptor{
					BcdHID:       0x0111,
					BCountryCode: 0x00,
					Descriptors:  []usb.HIDSubDescriptor{{Type: usb.ReportDescType}},
				},
				Report: dummyMouseReportDescriptor,
			},
			Endpoints: []usb.EndpointDescriptor{
				{
					BEndpointAddress: 0x83,
					BMAttributes:     0x03,
					WMaxPacketSize:   0x0004,
					BInterval:        0x0A,
				},
			},
		},
		// Actual "controller" descriptor
		{
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x02,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x01,
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
				Report: hid.Report{
					Items: []hid.Item{
						hid.UsagePage{Page: 0xFFFF},
						hid.Usage{Usage: 0x0001},
						hid.Collection{Kind: hid.CollectionApplication, Items: []hid.Item{
							hid.LogicalMinimum{Min: 0},
							hid.LogicalMaximum{Max: 255},
							hid.ReportSize{Bits: 8},
							hid.ReportCount{Count: 64},
							hid.Usage{Usage: 0x0001},
							hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
							hid.Usage{Usage: 0x0001},
							hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
						}},
					},
				},
			},
			Endpoints: []usb.EndpointDescriptor{
				{
					BEndpointAddress: 0x81,
					BMAttributes:     0x03,
					WMaxPacketSize:   0x0040,
					BInterval:        0x01,
				},
			},
		},
	},
	Strings: map[uint8]string{
		0: "\x04\x09", // LangID: en-US (0x0409)
		1: "Valve Software",
		2: "Steam Deck Controller",
		3: "VIIPER-Deck-Jupiter",
	},
}

func (s *SteamDeck) GetDescriptor() *usb.Descriptor {
	return &s.descriptor
}
