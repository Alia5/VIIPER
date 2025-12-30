package steamdeck

// OutputState is the device-facing output state for a Steam Deck (Jupiter/LCD)
// viiper:wire steamdeck s2c payload:byte*64
type OutputState struct {
	Payload [64]byte // Just do raw forwarding for now...
}

func (s OutputState) MarshalBinary() ([]byte, error) {
	b := make([]byte, 64)
	copy(b[0:64], s.Payload[0:64])
	return b, nil
}
