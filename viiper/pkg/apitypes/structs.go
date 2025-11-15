package apitypes

// Shared API response structs used by both handlers and clients.

type BusListResponse struct {
	Buses []uint32 `json:"buses"`
}

type BusCreateResponse struct {
	BusID uint32 `json:"busId"`
}

type BusRemoveResponse struct {
	BusID uint32 `json:"busId"`
}

type Device struct {
	BusID uint32 `json:"busId"`
	DevId string `json:"devId"`
	Vid   string `json:"vid"`
	Pid   string `json:"pid"`
	Type  string `json:"type"`
}

type DevicesListResponse struct {
	Devices []Device `json:"devices"`
}

type DeviceAddResponse struct {
	ID string `json:"id"` // Format: "<busId>-<devId>"
}

type DeviceRemoveResponse struct {
	BusID uint32 `json:"busId"`
	DevId string `json:"devId"`
}
