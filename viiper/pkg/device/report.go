package device

// ReportBuilder is an interface for device input states that can build USB reports.
type ReportBuilder interface {
	// BuildReport encodes the input state into a byte slice for USB transfer.
	BuildReport() []byte
}
