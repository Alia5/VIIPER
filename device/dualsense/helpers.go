package dualsense

import "math"

func GyroDpsToRaw(dps float64) int16 {
	return int16(min(max(math.Round(dps*GyroCountsPerDps), math.MinInt16), math.MaxInt16))
}

func GyroRawToDps(raw int16) float64 {
	return float64(raw) / GyroCountsPerDps
}

func AccelMS2ToRaw(ms2 float64) int16 {
	return int16(min(max(math.Round(ms2*AccelCountsPerMS2), math.MinInt16), math.MaxInt16))
}

func AccelRawToMS2(raw int16) float64 {
	return float64(raw) / AccelCountsPerMS2
}

func DefaultAccelRaw() (x, y, z int16) {
	return DefaultAccelXRaw, DefaultAccelYRaw, DefaultAccelZRaw
}

func encodeTouchCoords(b []byte, x, y uint16) {
	x = min(x, TouchpadMaxX)
	y = min(y, TouchpadMaxY)
	b[0] = uint8(x & 0xFF)
	b[1] = uint8((x>>8)&0x0F) | uint8((y&0x0F)<<4)
	b[2] = uint8(y >> 4)
}
