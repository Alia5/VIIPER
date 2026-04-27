// Package triton provides a minimal Valve Steam Triton Controller (wired, PID 0x1302) HID implementation.
package triton

import (
	"encoding/binary"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
	"github.com/Alia5/VIIPER/usbip"
)

// SteamTriton emulates a wired Valve Steam Triton Controller (VID 0x28DE, PID 0x1302).
//
// Protocol reference: SDL3 src/joystick/hidapi/SDL_hidapi_steam_triton.c and
// steam/controller_structs.h.
type SteamTriton struct {
	tick uint64

	stateMu        sync.Mutex
	inputState     *InputState
	lastReportSent time.Time

	featureMu       sync.Mutex
	featureResponse []byte

	outputFunc func(OutputState)
	descriptor usb.Descriptor
}

const usbSendTimeoutMS = 1000

// New returns a new SteamTriton device.
func New(o *device.CreateOptions) *SteamTriton {
	_ = o
	return &SteamTriton{
		descriptor: defaultDescriptor,
	}
}

// SetOutputCallback registers a callback invoked on haptic output reports.
func (t *SteamTriton) SetOutputCallback(f func(OutputState)) {
	t.outputFunc = f
}

// UpdateInputState updates the controller's current input state (thread-safe).
func (t *SteamTriton) UpdateInputState(st InputState) {
	t.stateMu.Lock()
	newState := st
	t.inputState = &newState
	t.stateMu.Unlock()
}

// HandleTransfer implements interrupt IN/OUT for the Triton controller interface.
//
// Interrupt IN (ep=1, dir=In): returns the 64-byte controller state report.
// Interrupt OUT (ep=1, dir=Out): receives haptic output commands.
func (t *SteamTriton) HandleTransfer(ep uint32, dir uint32, out []byte) []byte {
	switch {
	case ep == 1 && dir == usbip.DirIn:
		t.stateMu.Lock()
		if t.inputState != nil {
			seq := uint8(atomic.AddUint64(&t.tick, 1))
			st := *t.inputState
			t.inputState = nil
			t.lastReportSent = time.Now()
			t.stateMu.Unlock()
			return buildInReport(seq, st)
		}
		if time.Since(t.lastReportSent) > usbSendTimeoutMS*time.Millisecond {
			slog.Debug("SteamTriton input timeout, sending empty report")
			seq := uint8(atomic.AddUint64(&t.tick, 1))
			t.lastReportSent = time.Now()
			t.stateMu.Unlock()
			return buildInReport(seq, InputState{})
		}
		t.stateMu.Unlock()
		return nil

	case ep == 1 && dir == usbip.DirOut:
		// Haptic output report from the host (e.g. rumble via SDL_hid_write).
		t.handleOutputReport(out)
		return nil

	default:
		return nil
	}
}

// HandleControl implements EP0 class requests for HID feature reports.
//
// SDL sends SET_FEATURE with ID_SET_SETTINGS_VALUES to disable lizard mode and
// configure the IMU. We ACK those and respond with dummy attributes on GET_FEATURE.
func (t *SteamTriton) HandleControl(bmRequestType, bRequest uint8, wValue, wIndex, wLength uint16, data []byte) ([]byte, bool) {
	const (
		hidReqGetReport = 0x01
		hidReqSetReport = 0x09

		hidReqTypeClassInToInterface  = 0xA1
		hidReqTypeClassOutToInterface = 0x21

		hidReportTypeFeature = 0x03
	)

	reportType := uint8(wValue >> 8)

	switch {
	case bmRequestType == hidReqTypeClassInToInterface && bRequest == hidReqGetReport:
		want := int(wLength)
		if want <= 0 {
			return nil, true
		}
		if reportType == hidReportTypeFeature {
			return t.getFeatureResponse(want), true
		}
		return make([]byte, want), true

	case bmRequestType == hidReqTypeClassOutToInterface && bRequest == hidReqSetReport:
		if reportType == hidReportTypeFeature {
			t.handleFeatureReport(data)
		}
		return nil, true
	}

	return nil, false
}

func (t *SteamTriton) getFeatureResponse(want int) []byte {
	if want <= 0 {
		return nil
	}

	t.featureMu.Lock()
	resp := append([]byte(nil), t.featureResponse...)
	t.featureMu.Unlock()

	if len(resp) == 0 {
		out := make([]byte, want)
		if want > 0 {
			out[0] = FeatureReportID
		}
		return out
	}

	if len(resp) >= want {
		return append([]byte(nil), resp[:want]...)
	}
	out := make([]byte, want)
	copy(out, resp)
	return out
}

func (t *SteamTriton) setFeatureResponse(resp []byte) {
	t.featureMu.Lock()
	defer t.featureMu.Unlock()
	if resp == nil {
		t.featureResponse = nil
		return
	}
	t.featureResponse = append([]byte(nil), resp...)
}

func (t *SteamTriton) handleFeatureReport(data []byte) {
	// Strip the leading report ID byte if present.
	if len(data) >= 1 && data[0] == FeatureReportID {
		data = data[1:]
	}
	if len(data) < 2 {
		return
	}

	msgType := data[0]
	switch msgType {
	case FeatureIDGetAttributesValues:
		t.setFeatureResponse(buildGetAttributesResponse())
	case FeatureIDGetStringAttribute:
		var tag uint8 = AttribStrUnitSerial
		if len(data) >= 3 {
			// FeatureReportMsg header is [type, length], first payload byte is data[2]
			tag = data[2]
		}
		t.setFeatureResponse(buildGetStringAttributeResponse(tag))
	case FeatureIDSetSettingsValues, FeatureIDClearDigitalMappings, FeatureIDLoadDefaultSettings:
		// No-op: queue an empty echo so a follow-up GET_FEATURE returns something sensible.
		t.setFeatureResponse(buildEmptyResponse(msgType))
	}
}

func (t *SteamTriton) handleOutputReport(data []byte) {
	if len(data) == 0 {
		return
	}
	if t.outputFunc == nil {
		return
	}
	var out OutputState
	n := copy(out.Payload[:], data)
	_ = n
	t.outputFunc(out)
}

// buildInReport constructs a 64-byte HID interrupt IN report.
//
// Layout: [0x42 (report ID)] [TritonMTUFull_t packed] [zeros to 64 bytes].
// SDL parses from data[1] onward as TritonMTUNoQuat_t (first 45 bytes of the struct).
func buildInReport(seq uint8, st InputState) []byte {
	buf := make([]byte, 64)
	buf[0] = InReportControllerState // Report ID 0x42

	// TritonMTUNoQuat_t / TritonMTUFull_t (packed, little-endian):
	buf[1] = seq // seq_num

	binary.LittleEndian.PutUint32(buf[2:6], st.Buttons)

	binary.LittleEndian.PutUint16(buf[6:8], uint16(st.TriggerLeft))
	binary.LittleEndian.PutUint16(buf[8:10], uint16(st.TriggerRight))

	binary.LittleEndian.PutUint16(buf[10:12], uint16(st.LeftStickX))
	binary.LittleEndian.PutUint16(buf[12:14], uint16(st.LeftStickY))
	binary.LittleEndian.PutUint16(buf[14:16], uint16(st.RightStickX))
	binary.LittleEndian.PutUint16(buf[16:18], uint16(st.RightStickY))

	binary.LittleEndian.PutUint16(buf[18:20], uint16(st.LeftPadX))
	binary.LittleEndian.PutUint16(buf[20:22], uint16(st.LeftPadY))
	binary.LittleEndian.PutUint16(buf[22:24], st.PressureLeft)

	binary.LittleEndian.PutUint16(buf[24:26], uint16(st.RightPadX))
	binary.LittleEndian.PutUint16(buf[26:28], uint16(st.RightPadY))
	binary.LittleEndian.PutUint16(buf[28:30], st.PressureRight)

	// TritonMTUIMUNoQuat_t starts at offset 30:
	binary.LittleEndian.PutUint32(buf[30:34], st.Timestamp)
	binary.LittleEndian.PutUint16(buf[34:36], uint16(st.AccelX))
	binary.LittleEndian.PutUint16(buf[36:38], uint16(st.AccelY))
	binary.LittleEndian.PutUint16(buf[38:40], uint16(st.AccelZ))
	binary.LittleEndian.PutUint16(buf[40:42], uint16(st.GyroX))
	binary.LittleEndian.PutUint16(buf[42:44], uint16(st.GyroY))
	binary.LittleEndian.PutUint16(buf[44:46], uint16(st.GyroZ))

	// TritonMTUIMU_t (full) appends gyro quaternions at offset 46:
	binary.LittleEndian.PutUint16(buf[46:48], uint16(st.GyroQuatW))
	binary.LittleEndian.PutUint16(buf[48:50], uint16(st.GyroQuatX))
	binary.LittleEndian.PutUint16(buf[50:52], uint16(st.GyroQuatY))
	binary.LittleEndian.PutUint16(buf[52:54], uint16(st.GyroQuatZ))

	// Bytes 54–63: zero-padded.
	return buf
}

func buildEmptyResponse(msgType uint8) []byte {
	resp := make([]byte, HIDFeatureReportBytes)
	resp[0] = FeatureReportID
	resp[1] = msgType
	resp[2] = 0
	return resp
}

func buildFeatureResponse(msgType uint8, payload []byte) []byte {
	resp := make([]byte, HIDFeatureReportBytes)
	resp[0] = FeatureReportID
	resp[1] = msgType
	if payload == nil {
		resp[2] = 0
		return resp
	}
	if len(payload) > 61 { // 64 - 3 (reportID + type + length)
		payload = payload[:61]
	}
	resp[2] = uint8(len(payload))
	copy(resp[3:], payload)
	return resp
}

func buildGetStringAttributeResponse(requestedTag uint8) []byte {
	payload := make([]byte, 21)
	payload[0] = requestedTag
	const serial = "VIIPER000001"
	copy(payload[1:], []byte(serial))
	return buildFeatureResponse(FeatureIDGetStringAttribute, payload)
}

func buildGetAttributesResponse() []byte {
	const (
		productID         = uint32(TritonWiredPID)
		capabilities      = uint32(0x00003FFF) // generic Steam controller caps
		firmwareVersion   = uint32(0x00010001)
		firmwareBuildTime = uint32(0x00010001)
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

// controllerReportDescriptor is the vendor-defined HID report descriptor for
// the Triton controller interface. It uses report IDs to distinguish input,
// output, and feature reports — matching SDL's expectation that data[0] is the
// report type.
var controllerReportDescriptor = hid.Report{
	Items: []hid.Item{
		hid.UsagePage{Page: 0xFF00},
		hid.Usage{Usage: 0x0001},
		hid.Collection{Kind: hid.CollectionApplication, Items: []hid.Item{
			hid.LogicalMinimum{Min: 0},
			hid.LogicalMaximum{Max: 255},
			hid.ReportSize{Bits: 8},

			// Input: controller state (ID_TRITON_CONTROLLER_STATE = 0x42).
			// 63 bytes of payload + 1 byte report ID = 64 bytes on the wire.
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{InReportControllerState}},
			hid.Usage{Usage: uint16(InReportControllerState)},
			hid.ReportCount{Count: 63},
			hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			// Input: battery status (ID_TRITON_BATTERY_STATUS = 0x43).
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{InReportBatteryStatus}},
			hid.Usage{Usage: uint16(InReportBatteryStatus)},
			hid.ReportCount{Count: 9},
			hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			// Input: wireless status (ID_TRITON_WIRELESS_STATUS = 0x79).
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{InReportWirelessStatus}},
			hid.Usage{Usage: uint16(InReportWirelessStatus)},
			hid.ReportCount{Count: 1},
			hid.Input{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			// Output: haptic rumble (ID_OUT_REPORT_HAPTIC_RUMBLE = 0x80).
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{OutReportHapticRumble}},
			hid.Usage{Usage: uint16(OutReportHapticRumble)},
			hid.ReportCount{Count: 9},
			hid.Output{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			// Output: haptic pulse (ID_OUT_REPORT_HAPTIC_PULSE = 0x81).
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{OutReportHapticPulse}},
			hid.Usage{Usage: uint16(OutReportHapticPulse)},
			hid.ReportCount{Count: 9},
			hid.Output{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			// Output: haptic command (ID_OUT_REPORT_HAPTIC_COMMAND = 0x82).
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{OutReportHapticCommand}},
			hid.Usage{Usage: uint16(OutReportHapticCommand)},
			hid.ReportCount{Count: 3},
			hid.Output{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			// Output: haptic LFO tone (ID_OUT_REPORT_HAPTIC_LFO_TONE = 0x83).
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{OutReportHapticLFOTone}},
			hid.Usage{Usage: uint16(OutReportHapticLFOTone)},
			hid.ReportCount{Count: 9},
			hid.Output{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			// Output: haptic log sweep (ID_OUT_REPORT_HAPTIC_LOG_SWEEP = 0x85).
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{OutReportHapticLogSweep}},
			hid.Usage{Usage: uint16(OutReportHapticLogSweep)},
			hid.ReportCount{Count: 8},
			hid.Output{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			// Output: haptic script (ID_OUT_REPORT_HAPTIC_SCRIPT = 0x86).
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{OutReportHapticScript}},
			hid.Usage{Usage: uint16(OutReportHapticScript)},
			hid.ReportCount{Count: 3},
			hid.Output{Flags: hid.MainData | hid.MainVar | hid.MainAbs},

			// Feature: settings / attributes (report ID 0x01).
			hid.AnyItem{Type: hid.ItemTypeGlobal, Tag: 0x8, Data: hid.Data{FeatureReportID}},
			hid.Usage{Usage: uint16(FeatureReportID)},
			hid.ReportCount{Count: 63},
			hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
		}},
	},
}

var defaultDescriptor = usb.Descriptor{
	Device: usb.DeviceDescriptor{
		BcdUSB:             0x0200,
		BDeviceClass:       0x00,
		BDeviceSubClass:    0x00,
		BDeviceProtocol:    0x00,
		BMaxPacketSize0:    0x40,
		IDVendor:           ValveUSBVID,
		IDProduct:          TritonWiredPID,
		BcdDevice:          0x0100,
		IManufacturer:      0x01,
		IProduct:           0x02,
		ISerialNumber:      0x03,
		BNumConfigurations: 0x01,
		Speed:              2, // Full speed
	},
	// The Triton wired controller exposes a single HID interface.
	// SDL's Triton driver (SDL_hidapi_steam_triton.c) does not filter by interface
	// number, so a single-interface design is correct.
	Interfaces: []usb.InterfaceConfig{
		{
			Descriptor: usb.InterfaceDescriptor{
				BInterfaceNumber:   0x00,
				BAlternateSetting:  0x00,
				BNumEndpoints:      0x02, // interrupt IN + interrupt OUT
				BInterfaceClass:    0x03, // HID
				BInterfaceSubClass: 0x00,
				BInterfaceProtocol: 0x00,
				IInterface:         0x00,
			},
			HID: &usb.HIDFunction{
				Descriptor: usb.HIDDescriptor{
					BcdHID:       0x0111,
					BCountryCode: 0x00,
					Descriptors:  []usb.HIDSubDescriptor{{Type: usb.ReportDescType}},
				},
				Report: controllerReportDescriptor,
			},
			Endpoints: []usb.EndpointDescriptor{
				{
					BEndpointAddress: 0x81, // interrupt IN
					BMAttributes:     0x03,
					WMaxPacketSize:   0x0040, // 64 bytes
					BInterval:        0x01,   // 1ms (1 kHz)
				},
				{
					BEndpointAddress: 0x01, // interrupt OUT
					BMAttributes:     0x03,
					WMaxPacketSize:   0x0040, // 64 bytes
					BInterval:        0x01,
				},
			},
		},
	},
	Strings: map[uint8]string{
		0: "\x04\x09", // LangID: en-US (0x0409)
		1: "Valve Software",
		2: "Steam Controller",
		3: "VIIPER-Triton",
	},
}

func (t *SteamTriton) GetDescriptor() *usb.Descriptor {
	return &t.descriptor
}

func (t *SteamTriton) GetDeviceSpecificArgs() map[string]any {
	return map[string]any{}
}
