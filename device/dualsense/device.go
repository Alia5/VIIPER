package dualsense

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"math"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
	"github.com/Alia5/VIIPER/usbip"
)

type DualSense struct {
	ctx context.Context

	inputCh    chan *InputState
	inputState *InputState
	metaState  *MetaState

	outputFunc func(OutputState)
	descriptor usb.Descriptor

	subcommand [2]byte

	edge bool

	seqCounter     uint8
	packetSequence uint32
	nextTouchID    uint8
	touchIDs       [2]uint8
	touchWasActive [2]bool

	touchTimestamp uint8
	timestampBase  time.Time

	mtx sync.Mutex
}

func New(ctx context.Context, o *device.CreateOptions) (*DualSense, error) {
	return new(ctx, o, false)
}
func NewEdge(ctx context.Context, o *device.CreateOptions) (*DualSense, error) {
	return new(ctx, o, true)
}

func new(ctx context.Context, o *device.CreateOptions, edge bool) (*DualSense, error) {
	metaState := &MetaState{
		SerialNumber:       DefaultSerialNumberDS,
		MACAddress:         DefaultMACAddressDS,
		Board:              DefaultBoardStringDS,
		BuildTime:          DefaultBuildTime,
		BatteryStatus:      DefaultBatteryStatus,
		TemperatureCelsius: DefaultTemperature,
		BatteryVoltage:     DefaultVoltage,
		ShellColor:         DefaultShellColor,
	}
	if edge {
		metaState.SerialNumber = DefaultSerialNumberDSEdge
		metaState.MACAddress = DefaultMACAddressDSEdge
		metaState.Board = DefaultBoardStringEdge
	}
	if o != nil && o.DeviceSpecific != "" {
		var newMeta MetaState
		err := json.Unmarshal([]byte(o.DeviceSpecific), &newMeta)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON payload: %w", err)
		}
		if newMeta.SerialNumber != "" {
			metaState.SerialNumber = newMeta.SerialNumber
		}
		if newMeta.MACAddress != "" {
			metaState.MACAddress = newMeta.MACAddress
		}
		if newMeta.Board != "" {
			metaState.Board = newMeta.Board
		}
		if !newMeta.BuildTime.IsZero() {
			metaState.BuildTime = newMeta.BuildTime
		}
		if newMeta.BatteryStatus != 0 {
			metaState.BatteryStatus = newMeta.BatteryStatus
		}
		if newMeta.TemperatureCelsius != 0 {
			metaState.TemperatureCelsius = newMeta.TemperatureCelsius
		}
		if newMeta.BatteryVoltage != 0 {
			metaState.BatteryVoltage = newMeta.BatteryVoltage
		}
		metaState.ShellColor = newMeta.ShellColor
	}

	d := &DualSense{
		ctx:        ctx,
		descriptor: defaultDescriptor,
		metaState:  metaState,
		edge:       edge,
	}
	d.descriptor.Device.IDProduct = DefaultPIDDS
	if edge {
		d.descriptor.Device.IDProduct = DefaultPIDDSEdge
		d.descriptor.Strings = maps.Clone(d.descriptor.Strings)
		d.descriptor.Strings[2] = "DualSense Edge Wireless Controller"

		hidFunction := *d.descriptor.Interfaces[0].HID
		gamepadCollection := hidFunction.ReportDescriptor.Items[2].(hid.Collection)
		gamepadCollection.Items = slices.Clone(gamepadCollection.Items)
		for i, item := range gamepadCollection.Items {
			reportID, isReportID := item.(hid.ReportID)
			if !isReportID {
				continue
			}
			switch reportID.ID {
			case ReportIDOutput:
				gamepadCollection.Items[i+ReportCountOffsetFromReportID] = hid.ReportCount{
					Count: OutputReportCountEdge,
				}
			case 0xF2:
				gamepadCollection.Items[i+ReportCountOffsetFromReportID] = hid.ReportCount{
					Count: FeatureF2CountEdge,
				}
			}
		}
		gamepadCollection.Items = append(gamepadCollection.Items, EdgeOnlyFeatureReports...)
		hidFunction.ReportDescriptor.Items = slices.Clone(hidFunction.ReportDescriptor.Items)
		hidFunction.ReportDescriptor.Items[2] = gamepadCollection
		d.descriptor.Interfaces = slices.Clone(d.descriptor.Interfaces)
		d.descriptor.Interfaces[0].HID = &hidFunction
	}

	if o != nil {
		if o.IDVendor != nil {
			d.descriptor.Device.IDVendor = *o.IDVendor
		}
		if o.IDProduct != nil {
			d.descriptor.Device.IDProduct = *o.IDProduct
		}
	}

	slog.Info("DualSense device instantiated",
		"edge", edge,
		"vid", d.descriptor.Device.IDVendor,
		"pid", d.descriptor.Device.IDProduct,
		"interfaces", len(d.descriptor.Interfaces))

	d.inputState = NewInputState()
	d.inputCh = make(chan *InputState, 1)
	d.inputCh <- d.inputState
	d.timestampBase = time.Now()

	return d, nil
}

func (d *DualSense) SetMetaState(meta MetaState) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.metaState = &meta
}

func (d *DualSense) SetOutputCallback(f func(OutputState)) {
	d.outputFunc = f
}

func (d *DualSense) UpdateInputState(state *InputState) {
	d.mtx.Lock()
	d.inputState = state
	d.mtx.Unlock()
	select {
	case <-d.inputCh:
	default:
	}
	d.inputCh <- state
}

func (d *DualSense) GetDescriptor() *usb.Descriptor {
	return &d.descriptor
}

func (d *DualSense) GetDeviceSpecificArgs() map[string]any {
	var res map[string]any
	d.mtx.Lock()
	defer d.mtx.Unlock()

	bytes, err := json.Marshal(d.metaState)
	if err != nil {
		return map[string]any{}
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return map[string]any{}
	}
	return res
}

func (d *DualSense) HandleTransfer(ctx context.Context, ep uint32, dir uint32, out []byte) []byte {
	if dir == usbip.DirIn {
		switch ep {
		case 4:
			select {
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					d.mtx.Lock()
					is := d.inputState
					ms := *d.metaState
					d.mtx.Unlock()
					return d.buildUSBInputReport(is, &ms)
				}
				return nil
			case is := <-d.inputCh:
				d.mtx.Lock()
				ms := *d.metaState
				d.mtx.Unlock()
				return d.buildUSBInputReport(is, &ms)
			}
		default:
			return nil
		}
	}

	if dir == usbip.DirOut && ep == 3 {
		if len(out) >= 48 && out[0] == ReportIDOutput {
			if d.outputFunc != nil {
				d.outputFunc(parseOutputReport(out))
			}
		}
	}

	return nil
}

func (d *DualSense) HandleControl(bmRequestType, bRequest uint8, wValue, wIndex, wLength uint16, data []byte) ([]byte, bool) {
	reportType := uint8(wValue >> 8)
	reportID := uint8(wValue & 0xFF)

	switch bmRequestType {
	case hidClassIN:
		switch bRequest {
		case hidGetReport:
			if reportType == reportTypeInput && reportID == ReportIDInput {
				d.mtx.Lock()
				is := *d.inputState
				ms := *d.metaState
				d.mtx.Unlock()
				b := d.buildUSBInputReport(&is, &ms)
				if wLength > 0 && int(wLength) < len(b) {
					b = b[:wLength]
				}
				return b, true
			}
			if reportType == reportTypeFeature {
				if fn, ok := featureGetHandlers[reportID]; ok {
					b := fn(d)
					if wLength > 0 && int(wLength) < len(b) {
						b = b[:wLength]
					}
					return b, true
				}
			}
		case hidGetIdle:
			return []byte{0x00}, true
		case hidGetProtocol:
			return []byte{0x01}, true
		}
	case hidClassOUT:
		if bRequest == hidSetReport {
			switch {
			case reportType == reportTypeFeature && reportID == featureIDCommand && len(data) >= 3:
				d.subcommand[0] = data[1]
				d.subcommand[1] = data[2]
				return nil, true
			case reportType == reportTypeFeature:
				return nil, true
			case reportType == reportTypeOutput && reportID == ReportIDOutput && len(data) >= 48:
				if d.outputFunc != nil {
					d.outputFunc(parseOutputReport(data))
				}
				return nil, true
			}
		}
	}

	slog.Warn("DualSense control request unhandled",
		"bmRequestType", bmRequestType,
		"bRequest", bRequest,
		"reportType", reportType,
		"reportID", reportID,
		"wIndex", wIndex,
		"wLength", wLength,
		"dataLen", len(data))

	return nil, false
}

var featureGetHandlers = map[byte]func(*DualSense) []byte{
	FeatureIDCalibration:     (*DualSense).featureReportCalibration,
	FeatureIDPairing:         (*DualSense).featureReportPairing,
	FeatureIDFirmware:        (*DualSense).featureReportFirmware,
	featureIDCommandResponse: (*DualSense).featureReportCommandResponse,
}

func parseOutputReport(out []byte) OutputState {
	return OutputState{
		UpdateFlags1: out[1],
		UpdateFlags2: out[2],
		UpdateFlags3: out[39],
		RumbleSmall:  out[3],
		RumbleLarge:  out[4],
		LedRed:       out[45],
		LedGreen:     out[46],
		LedBlue:      out[47],
		PlayerLeds:   out[44],
	}
}

func (d *DualSense) featureReportCalibration() []byte {
	report := make([]byte, 41)
	report[0] = FeatureIDCalibration

	for i, v := range [17]int16{
		0, 0, 0,
		8192, -8192, 8192, -8192, 8192, -8192,
		500, 500,
		8192, -8192, 8192, -8192, 8192, -8192,
	} {
		binary.LittleEndian.PutUint16(report[1+i*2:], uint16(v))
	}

	report[35] = 0x0B // TODO:
	return report
}

func (d *DualSense) featureReportPairing() []byte {
	report := make([]byte, 20)
	report[0] = FeatureIDPairing

	d.mtx.Lock()
	mac := d.metaState.MACAddress
	d.mtx.Unlock()

	if hw, err := net.ParseMAC(mac); err == nil && len(hw) == 6 {
		for i := range 6 {
			report[1+i] = hw[5-i]
		}
	}

	// TODO:
	report[7] = 0x08
	report[8] = 0x25
	report[10] = 0x1E
	report[12] = 0xEE
	report[13] = 0x74
	report[14] = 0xD0
	report[15] = 0xBC
	return report
}

func (d *DualSense) featureReportFirmware() []byte {
	report := make([]byte, 64)
	report[0] = FeatureIDFirmware

	d.mtx.Lock()
	bt := d.metaState.BuildTime
	d.mtx.Unlock()

	copy(report[1:12], bt.Format("Jan 02 2006"))
	copy(report[12:20], bt.Format("15:04:05"))

	report[20] = HardwareType
	report[21] = 0x01 // TODO: unknown
	report[22] = 0x44 // TODO: put in CONST!!! // build revision from real device

	binary.LittleEndian.PutUint32(report[24:28], HwInfo)

	// TODO: unknown
	report[28] = 0x36
	report[31] = 0x01
	report[32] = 0xC1
	report[33] = 0xC8

	binary.LittleEndian.PutUint16(report[44:46], FirmwareVersion)

	// TODO: unknown
	report[48] = 0x14
	report[52] = 0x0B
	report[54] = 0x01
	report[56] = 0x06
	return report
}

func (d *DualSense) featureReportCommandResponse() []byte {
	report := make([]byte, 64)
	report[0] = featureIDCommandResponse

	d.mtx.Lock()
	sub := d.subcommand
	serial := d.metaState.SerialNumber
	voltage := d.metaState.BatteryVoltage
	temp := d.metaState.TemperatureCelsius
	d.mtx.Unlock()

	switch sub[0] {
	case subcmdSerial:
		copy(report[3:21], serial)
	case subcmdStatus:
		// nvs locked
		report[1] = 0x01
		report[4] = 0x01
	case subcmdSensors:
		vRaw := uint16(math.Round(voltage * 1000))
		report[4] = byte(vRaw)
		report[5] = byte(vRaw >> 8)
		tRaw := uint16(math.Max(0, math.Min(4095, math.Round((2470.0-temp*26.0)/0.78125))))
		report[6] = byte(tRaw)
		report[7] = byte(tRaw >> 8)
	default:
		slog.Warn("DualSense: unknown sub-command for featureIDCommandResponse",
			"sub0", sub[0], "sub1", sub[1])
		report[1] = 0x01
	}
	return report
}

func (d *DualSense) buildUSBInputReport(s *InputState, m *MetaState) []byte {
	b := make([]byte, InputReportSize)

	b[0] = ReportIDInput

	b[1] = uint8(int16(s.LX) + 128)
	b[2] = uint8(int16(s.LY) + 128)
	b[3] = uint8(int16(s.RX) + 128)
	b[4] = uint8(int16(s.RY) + 128)

	b[5] = s.L2
	b[6] = s.R2

	d.seqCounter++
	b[7] = d.seqCounter

	d.packetSequence++
	binary.LittleEndian.PutUint32(b[12:16], d.packetSequence)

	usbDPad := uint8(DPadUSBNeutral)
	switch {
	case s.DPad&DPadUp != 0 && s.DPad&DPadRight != 0:
		usbDPad = DPadUSBUpRight
	case s.DPad&DPadUp != 0 && s.DPad&DPadLeft != 0:
		usbDPad = DPadUSBUpLeft
	case s.DPad&DPadDown != 0 && s.DPad&DPadRight != 0:
		usbDPad = DPadUSBDownRight
	case s.DPad&DPadDown != 0 && s.DPad&DPadLeft != 0:
		usbDPad = DPadUSBDownLeft
	case s.DPad&DPadUp != 0:
		usbDPad = DPadUSBUp
	case s.DPad&DPadDown != 0:
		usbDPad = DPadUSBDown
	case s.DPad&DPadLeft != 0:
		usbDPad = DPadUSBLeft
	case s.DPad&DPadRight != 0:
		usbDPad = DPadUSBRight
	}
	b[8] = (usbDPad & DPadMask) | (uint8(s.Buttons) & 0xF0)
	b[9] = uint8(s.Buttons >> 8)
	b[10] = uint8(s.Buttons >> 16)

	binary.LittleEndian.PutUint16(b[16:18], uint16(s.GyroX))
	binary.LittleEndian.PutUint16(b[18:20], uint16(s.GyroY))
	binary.LittleEndian.PutUint16(b[20:22], uint16(s.GyroZ))

	binary.LittleEndian.PutUint16(b[22:24], uint16(s.AccelX))
	binary.LittleEndian.PutUint16(b[24:26], uint16(s.AccelY))
	binary.LittleEndian.PutUint16(b[26:28], uint16(s.AccelZ))

	ts := uint32(time.Since(d.timestampBase).Microseconds() * 3)
	binary.LittleEndian.PutUint32(b[28:32], ts)

	touchedInLastReport := d.touchWasActive[0] || d.touchWasActive[1]

	b[33] = d.touchContactByte(0, s.Touch1Active)
	encodeTouchCoords(b[34:37], s.Touch1X, s.Touch1Y)

	b[37] = d.touchContactByte(1, s.Touch2Active)
	encodeTouchCoords(b[38:41], s.Touch2X, s.Touch2Y)

	if s.Touch1Active || s.Touch2Active || touchedInLastReport {
		lastTouchScan := time.Since(d.timestampBase).Truncate(touchScanInterval)
		d.touchTimestamp = uint8(lastTouchScan / touchTimestampUnit)
	}
	b[41] = d.touchTimestamp

	b[42] = TriggerStatusNeutral
	b[43] = TriggerStatusNeutral

	if d.edge {
		b[49] = EdgeActiveProfile
	} else {
		binary.LittleEndian.PutUint32(b[49:53], ts)
	}
	b[53] = m.BatteryStatus
	b[54] = ConnectStateUSB

	return b
}

func (d *DualSense) touchContactByte(finger int, active bool) uint8 {
	if active && !d.touchWasActive[finger] {
		d.touchIDs[finger] = d.nextTouchID
		d.nextTouchID = (d.nextTouchID + 1) & touchIDMask
	}
	d.touchWasActive[finger] = active
	if active {
		return d.touchIDs[finger]
	}
	return d.touchIDs[finger] | TouchInactiveMask
}
