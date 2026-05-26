package ns2pro

func minimalFlashBlock(address uint32) []byte {
	block := make([]byte, 0x40)
	switch address {
	case 0x13000:
		copy(block[2:], []byte("VIIPER-NS2PRO-00"))
	case 0x13080, 0x130C0:
		encodeStickCalibration(block[0x28:], StickCenter, StickCenter, 2047, 2047, 2048, 2048)
	case 0x13040, 0x13100, 0x1FC040, 0x1FC080:
		// Zeroed data is intentional: no gyro/accel bias and no user calibration magic.
	default:
	}
	return block
}

func encodeStickCalibration(out []byte, neutralX, neutralY, maxX, maxY, minX, minY uint16) {
	if len(out) < 9 {
		return
	}
	packStick12(out[0:3], neutralX, neutralY)
	packStick12(out[3:6], maxX, maxY)
	packStick12(out[6:9], minX, minY)
}
