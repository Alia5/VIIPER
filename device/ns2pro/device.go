// Package ns2pro provides a Nintendo Switch 2 Pro Controller compatible HID device.
package ns2pro

import (
	"sync"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usbip"
)

type NS2Pro struct {
	stateMu        sync.Mutex
	inputState     *InputState
	outputMu       sync.RWMutex
	outputCallback func(OutputState)
	outputVersion  uint64
	descriptor     usb.Descriptor

	batteryVolts uint16

	protoMu           sync.Mutex
	activeReportID    uint8
	featureMask       uint8
	featureFlags      uint8
	usbReportsEnabled bool
	reportCounter32   uint32
	reportCounter8    uint8
	motionTimestamp   uint32
	bulkInQueue       [][]byte
}

func New(o *device.CreateOptions) (*NS2Pro, error) {
	d := &NS2Pro{
		inputState:     defaultInputState(),
		descriptor:     MakeDescriptor(),
		activeReportID: ReportIDPro,
		featureFlags:   FeatureButtons | FeatureSticks,
		batteryVolts:   BatteryVolts,
	}
	if o != nil {
		if o.IDVendor != nil {
			d.descriptor.Device.IDVendor = *o.IDVendor
		}
		if o.IDProduct != nil {
			d.descriptor.Device.IDProduct = *o.IDProduct
		}
	}
	return d, nil
}

func (d *NS2Pro) SetOutputCallback(f func(OutputState)) func() {
	d.outputMu.Lock()
	d.outputVersion++
	version := d.outputVersion
	d.outputCallback = f
	d.outputMu.Unlock()

	return func() {
		d.outputMu.Lock()
		if d.outputVersion == version {
			d.outputCallback = nil
		}
		d.outputMu.Unlock()
	}
}

func (d *NS2Pro) UpdateInputState(state InputState) {
	d.stateMu.Lock()
	defer d.stateMu.Unlock()
	d.inputState = &state
}

func (d *NS2Pro) HandleTransfer(ep uint32, dir uint32, out []byte) []byte {
	switch {
	case dir == usbip.DirIn && ep == 1:
		return d.nextInputReport()
	case dir == usbip.DirIn && ep == 2:
		return d.popBulkIn()
	case dir == usbip.DirOut && ep == 1:
		d.handleOutputReport(out)
	case dir == usbip.DirOut && ep == 2:
		d.handleBulkOut(out)
	}
	return nil
}

func (d *NS2Pro) HandleControl(bmRequestType, bRequest uint8, wValue, wIndex uint16, wLength uint16, data []byte) ([]byte, bool) {
	const (
		hidGetReport = 0x01
		hidSetReport = 0x09
	)
	const (
		reportTypeInput  = 0x01
		reportTypeOutput = 0x02
	)

	reportType := uint8(wValue >> 8)
	reportID := uint8(wValue)

	if bmRequestType == 0xA1 && bRequest == hidGetReport && reportType == reportTypeInput {
		switch reportID {
		case ReportIDCommon, ReportIDPro, 0:
			return d.inputReportForID(reportID), true
		}
	}

	if bmRequestType == 0x21 && bRequest == hidSetReport && reportType == reportTypeOutput && reportID == ReportIDOutput {
		d.handleOutputReport(data)
		return nil, true
	}

	if isAudioClassRequest(bmRequestType) {
		switch bRequest {
		case 0x01: // SET_CUR
			return nil, true
		case 0x81, 0x82, 0x83, 0x84: // GET_CUR/MIN/MAX/RES
			return make([]byte, wLength), true
		}
	}

	return nil, false
}

func isAudioClassRequest(bmRequestType uint8) bool {
	const (
		requestTypeMask   = 0x60
		requestClass      = 0x20
		recipientMask     = 0x1F
		recipientIface    = 0x01
		recipientEndpoint = 0x02
	)
	return bmRequestType&requestTypeMask == requestClass &&
		(bmRequestType&recipientMask == recipientIface || bmRequestType&recipientMask == recipientEndpoint)
}

func (d *NS2Pro) GetDescriptor() *usb.Descriptor {
	return &d.descriptor
}

func (d *NS2Pro) GetDeviceSpecificArgs() map[string]any {
	return nil
}

func (d *NS2Pro) nextInputReport() []byte {
	d.protoMu.Lock()
	reportID := d.activeReportID
	d.protoMu.Unlock()
	return d.inputReportForID(reportID)
}

func (d *NS2Pro) inputReportForID(reportID uint8) []byte {
	d.stateMu.Lock()
	st := *d.inputState
	d.stateMu.Unlock()

	d.protoMu.Lock()
	if reportID == 0 {
		reportID = d.activeReportID
	}
	features := d.featureFlags
	var report []byte
	switch reportID {
	case ReportIDCommon:
		d.reportCounter32++
		if features&FeatureIMU != 0 {
			d.motionTimestamp += 4000
		}
		report = st.buildCommonReport(d.reportCounter32, d.motionTimestamp, features, d.batteryVolts)
	default:
		d.reportCounter8++
		report = st.buildProReport(d.reportCounter8, features)
	}
	d.protoMu.Unlock()
	return report
}

func (d *NS2Pro) handleOutputReport(out []byte) {
	if len(out) == 0 {
		return
	}

	payload := out
	if out[0] == ReportIDOutput {
		payload = out[1:]
	} else if len(out) != OutputRumbleSize {
		return
	}
	if len(payload) < OutputRumbleSize {
		return
	}

	feedback := OutputState{}
	copy(feedback.LeftRumble[:], payload[0:16])
	copy(feedback.RightRumble[:], payload[16:32])
	feedback.Flags = OutputFlagRumble
	d.emitOutput(feedback)
}

func (d *NS2Pro) emitOutput(feedback OutputState) {
	d.outputMu.RLock()
	callback := d.outputCallback
	d.outputMu.RUnlock()
	if callback != nil {
		callback(feedback)
	}
}

func (d *NS2Pro) enqueueResponse(resp []byte) {
	d.protoMu.Lock()
	defer d.protoMu.Unlock()
	d.bulkInQueue = append(d.bulkInQueue, append([]byte(nil), resp...))
}

func (d *NS2Pro) popBulkIn() []byte {
	d.protoMu.Lock()
	defer d.protoMu.Unlock()
	if len(d.bulkInQueue) == 0 {
		return nil
	}
	chunk := d.bulkInQueue[0]
	d.bulkInQueue = d.bulkInQueue[1:]
	return append([]byte(nil), chunk...)
}

func commandHeader(cmd, seq, sub uint8) []byte {
	return []byte{cmd, 0x01, seq, sub, 0x10, 0x78, 0x00, 0x00}
}
