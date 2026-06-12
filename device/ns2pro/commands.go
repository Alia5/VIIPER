package ns2pro

import "encoding/binary"

func (d *NS2Pro) handleBulkOut(out []byte) {
	if len(out) < 8 {
		return
	}
	cmd := out[0]
	seq := out[2]
	sub := out[3]

	switch cmd {
	case cmdFlash:
		d.handleFlashCommand(seq, sub, out)
	case cmdUSB:
		d.handleUSBCommand(seq, sub, out)
	case cmdFeature:
		d.handleFeatureCommand(seq, sub, out)
	case cmdPlayerLED:
		d.handlePlayerLEDCommand(seq, sub, out)
	default:
		d.enqueueResponse(commandHeader(cmd, seq, sub))
	}
}

func (d *NS2Pro) handlePlayerLEDCommand(seq, sub uint8, out []byte) {
	if sub == subPlayerLEDSet && len(out) >= 9 {
		d.emitOutput(OutputState{
			Flags:         OutputFlagLED,
			PlayerLedMask: out[8],
		})
	}
	d.enqueueResponse(commandHeader(cmdPlayerLED, seq, sub))
}

func (d *NS2Pro) handleFlashCommand(seq, sub uint8, out []byte) {
	if sub != subFlashRead || len(out) < 16 {
		d.enqueueResponse(commandHeader(cmdFlash, seq, sub))
		return
	}

	address := binary.LittleEndian.Uint32(out[12:16])
	resp := make([]byte, 16+flashBlockSize)
	copy(resp[0:8], commandHeader(cmdFlash, seq, sub))
	resp[8] = flashBlockSize
	binary.LittleEndian.PutUint32(resp[12:16], address)
	copy(resp[16:], d.minimalFlashBlock(address))
	d.enqueueResponse(resp)
}

func (d *NS2Pro) handleUSBCommand(seq, sub uint8, out []byte) {
	switch sub {
	case subUSBEnableReports:
		if len(out) >= 9 {
			d.protoMu.Lock()
			d.usbReportsEnabled = out[8] != 0
			d.protoMu.Unlock()
		}
		d.enqueueResponse(append(commandHeader(cmdUSB, seq, sub), 0x01, 0x00, 0x00, 0x00))
	case subUSBSelectReport:
		if len(out) >= 9 {
			switch out[8] {
			case ReportIDCommon, ReportIDPro:
				d.protoMu.Lock()
				d.activeReportID = out[8]
				d.protoMu.Unlock()
			}
		}
		d.enqueueResponse(commandHeader(cmdUSB, seq, sub))
	case subUSBStartReports:
		d.protoMu.Lock()
		d.usbReportsEnabled = true
		d.protoMu.Unlock()
		d.enqueueResponse(append(commandHeader(cmdUSB, seq, sub), 0x01, 0x00, 0x00, 0x00))
	default:
		d.enqueueResponse(commandHeader(cmdUSB, seq, sub))
	}
}

func (d *NS2Pro) handleFeatureCommand(seq, sub uint8, out []byte) {
	flags := uint8(0)
	if len(out) >= 9 {
		flags = out[8]
	}

	switch sub {
	case subFeatureInfo:
		payload := make([]byte, 12)
		copy(payload[4:], featureInfo(flags))
		d.enqueueResponse(append(commandHeader(cmdFeature, seq, sub), payload...))
	case subFeatureSetMask:
		d.protoMu.Lock()
		d.featureMask = flags
		d.protoMu.Unlock()
		d.enqueueResponse(append(commandHeader(cmdFeature, seq, sub), 0x00, 0x00, 0x00, 0x00))
	case subFeatureReset:
		d.protoMu.Lock()
		d.featureMask = 0
		d.featureFlags = 0
		d.protoMu.Unlock()
		d.enqueueResponse(append(commandHeader(cmdFeature, seq, sub), 0x00, 0x00, 0x00, 0x00))
	case subFeatureEnable:
		d.protoMu.Lock()
		d.featureFlags |= d.maskedFeatures(flags)
		d.protoMu.Unlock()
		d.enqueueResponse(append(commandHeader(cmdFeature, seq, sub), 0x00, 0x00, 0x00, 0x00))
	case subFeatureDisable:
		d.protoMu.Lock()
		d.featureFlags &^= d.maskedFeatures(flags)
		d.protoMu.Unlock()
		d.enqueueResponse(append(commandHeader(cmdFeature, seq, sub), 0x00, 0x00, 0x00, 0x00))
	default:
		d.enqueueResponse(append(commandHeader(cmdFeature, seq, sub), 0x00, 0x00, 0x00, 0x00))
	}
}

// maskedFeatures must be called with protoMu held.
func (d *NS2Pro) maskedFeatures(flags uint8) uint8 {
	if d.featureMask == 0 {
		return flags
	}
	return flags & d.featureMask
}

func featureInfo(flags uint8) []byte {
	out := make([]byte, 8)
	for _, entry := range featureInfoMap {
		if flags&entry.feature != 0 {
			out[entry.index] = entry.value
		}
	}
	return out
}

var featureInfoMap = []struct {
	feature uint8
	index   int
	value   byte
}{
	{FeatureButtons, 0, 0x07},
	{FeatureSticks, 1, 0x07},
	{FeatureIMU, 2, 0x01},
	{FeatureMouse, 4, 0x03},
	{FeatureRumble, 5, 0x03},
}
