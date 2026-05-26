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
	case 0x02:
		d.handleFlashCommand(seq, sub, out)
	case 0x03:
		d.handleUSBCommand(seq, sub, out)
	case 0x0C:
		d.handleFeatureCommand(seq, sub, out)
	case 0x09:
		d.handlePlayerLEDCommand(seq, sub, out)
	default:
		d.enqueueResponse(commandHeader(cmd, seq, sub))
	}
}

func (d *NS2Pro) handlePlayerLEDCommand(seq, sub uint8, out []byte) {
	if sub == 0x07 && len(out) >= 9 {
		d.emitOutput(OutputState{
			Flags:         OutputFlagLED,
			PlayerLedMask: out[8],
		})
	}
	d.enqueueResponse(commandHeader(0x09, seq, sub))
}

func (d *NS2Pro) handleFlashCommand(seq, sub uint8, out []byte) {
	if sub != 0x01 || len(out) < 16 {
		d.enqueueResponse(commandHeader(0x02, seq, sub))
		return
	}

	address := binary.LittleEndian.Uint32(out[12:16])
	resp := make([]byte, 0x50)
	copy(resp[0:8], commandHeader(0x02, seq, sub))
	resp[8] = 0x40
	binary.LittleEndian.PutUint32(resp[12:16], address)
	copy(resp[16:], minimalFlashBlock(address))
	d.enqueueResponse(resp)
}

func (d *NS2Pro) handleUSBCommand(seq, sub uint8, out []byte) {
	switch sub {
	case 0x03:
		if len(out) >= 9 {
			d.usbReportsEnabled = out[8] != 0
		}
		d.enqueueResponse(append(commandHeader(0x03, seq, sub), 0x01, 0x00, 0x00, 0x00))
	case 0x0A:
		if len(out) >= 9 {
			switch out[8] {
			case ReportIDCommon, ReportIDPro:
				d.activeReportID = out[8]
			}
		}
		d.enqueueResponse(commandHeader(0x03, seq, sub))
	case 0x0D:
		d.usbReportsEnabled = true
		d.enqueueResponse(append(commandHeader(0x03, seq, sub), 0x01, 0x00, 0x00, 0x00))
	default:
		d.enqueueResponse(commandHeader(0x03, seq, sub))
	}
}

func (d *NS2Pro) handleFeatureCommand(seq, sub uint8, out []byte) {
	flags := uint8(0)
	if len(out) >= 9 {
		flags = out[8]
	}

	switch sub {
	case 0x01:
		payload := make([]byte, 12)
		copy(payload[4:], featureInfo(flags))
		d.enqueueResponse(append(commandHeader(0x0C, seq, sub), payload...))
	case 0x02:
		d.featureMask = flags
		d.enqueueResponse(append(commandHeader(0x0C, seq, sub), 0x00, 0x00, 0x00, 0x00))
	case 0x03:
		d.featureMask = 0
		d.featureFlags = 0
		d.enqueueResponse(append(commandHeader(0x0C, seq, sub), 0x00, 0x00, 0x00, 0x00))
	case 0x04:
		d.featureFlags |= d.maskedFeatures(flags)
		d.enqueueResponse(append(commandHeader(0x0C, seq, sub), 0x00, 0x00, 0x00, 0x00))
	case 0x05:
		d.featureFlags &^= d.maskedFeatures(flags)
		d.enqueueResponse(append(commandHeader(0x0C, seq, sub), 0x00, 0x00, 0x00, 0x00))
	default:
		d.enqueueResponse(append(commandHeader(0x0C, seq, sub), 0x00, 0x00, 0x00, 0x00))
	}
}

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
