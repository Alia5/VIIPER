package triton

// OutputState is the device-facing output state for a Valve Steam Triton controller.
//
// The Triton uses HID output reports (via hid_write / interrupt OUT endpoint)
// for haptic commands. The first byte is the output report ID (e.g. OutReportHapticRumble).
//
// viiper:wire triton s2c payload:byte*10
type OutputState struct {
	Payload [HapticStateSize]byte
}

func (s OutputState) MarshalBinary() ([]byte, error) {
	b := make([]byte, HapticStateSize)
	copy(b, s.Payload[:])
	return b, nil
}
