package apitypes

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ApiError represents an RFC 7807 (problem+json) error response.
type ApiError struct {
	// Status is the HTTP-style status code (e.g., 400, 404, 500)
	Status int `json:"status"`
	// Title is a short, human-readable summary of the problem type
	Title string `json:"title"`
	// Detail is a human-readable explanation specific to this occurrence
	Detail string `json:"detail"`
}

func (e ApiError) Error() string {
	if e.Status == 0 && e.Title == "" {
		return "unknown error"
	}
	if e.Status == 0 {
		return fmt.Sprintf("%s: %s", e.Title, e.Detail)
	}
	return fmt.Sprintf("%d %s: %s", e.Status, e.Title, e.Detail)
}

// --

type PingResponse struct {
	Server  string `json:"server"`
	Version string `json:"version"`
}

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

type DeviceRemoveResponse struct {
	BusID uint32 `json:"busId"`
	DevId string `json:"devId"`
}

type DeviceCreateRequest struct {
	Type      *string `json:"type"`
	IdVendor  *uint16 `json:"idVendor,omitempty"`
	IdProduct *uint16 `json:"idProduct,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling to accept both uint16 and hex string formats
// for idVendor and idProduct (e.g., "0x12ac" or 4780).
func (d *DeviceCreateRequest) UnmarshalJSON(data []byte) error {
	// Parse into a temporary structure with flexible types
	var raw struct {
		Type      *string `json:"type"`
		IdVendor  any     `json:"idVendor,omitempty"`
		IdProduct any     `json:"idProduct,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	d.Type = raw.Type

	if raw.IdVendor != nil {
		val, err := parseUint16OrHex(raw.IdVendor)
		if err != nil {
			return fmt.Errorf("idVendor: %w", err)
		}
		d.IdVendor = &val
	}

	if raw.IdProduct != nil {
		val, err := parseUint16OrHex(raw.IdProduct)
		if err != nil {
			return fmt.Errorf("idProduct: %w", err)
		}
		d.IdProduct = &val
	}

	return nil
}

// parseUint16OrHex accepts either a JSON number or a hex string like "0x12ac"
func parseUint16OrHex(v any) (uint16, error) {
	switch val := v.(type) {
	case float64:
		if val < 0 || val > 65535 {
			return 0, fmt.Errorf("value %v out of uint16 range", val)
		}
		return uint16(val), nil
	case string:
		s := strings.TrimSpace(val)
		base := 10
		if strings.HasPrefix(strings.ToLower(s), "0x") {
			s = s[2:]
			base = 16
		} else if len(s) > 0 {
			if strings.ContainsAny(s, "abcdefABCDEF") {
				base = 16
			}
		}
		parsed, err := strconv.ParseUint(s, base, 16)
		if err != nil {
			return 0, fmt.Errorf("invalid hex/numeric string %q: %w", val, err)
		}
		return uint16(parsed), nil
	default:
		return 0, fmt.Errorf("expected number or hex string, got %T", v)
	}
}
