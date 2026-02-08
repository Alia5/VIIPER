package apitypes

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
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
	BusID          uint32         `json:"busId"`
	DevId          string         `json:"devId"`
	Vid            string         `json:"vid"`
	Pid            string         `json:"pid"`
	Type           string         `json:"type"`
	DeviceSpecific map[string]any `json:"deviceSpecific"`
}

type DevicesListResponse struct {
	Devices []Device `json:"devices"`
}

type DeviceRemoveResponse struct {
	BusID uint32 `json:"busId"`
	DevId string `json:"devId"`
}

type DeviceCreateRequest struct {
	Type           *string        `json:"type"`
	IdVendor       *uint16        `json:"idVendor,omitempty"`
	IdProduct      *uint16        `json:"idProduct,omitempty"`
	DeviceSpecific map[string]any `json:"deviceSpecific,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling to accept both uint16 and hex string formats
// for idVendor and idProduct (e.g., "0x12ac" or 4780).
func (d *DeviceCreateRequest) UnmarshalJSON(data []byte) error {
	// Parse into a temporary structure with flexible types
	var raw struct {
		Type           *string        `json:"type"`
		IdVendor       any            `json:"idVendor,omitempty"`
		IdProduct      any            `json:"idProduct,omitempty"`
		DeviceSpecific map[string]any `json:"deviceSpecific,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	d.Type = raw.Type

	if raw.IdVendor != nil {
		val, err := parseNumberOrHex[uint16](raw.IdVendor)
		if err != nil {
			return fmt.Errorf("idVendor: %w", err)
		}
		d.IdVendor = &val
	}

	if raw.IdProduct != nil {
		val, err := parseNumberOrHex[uint16](raw.IdProduct)
		if err != nil {
			return fmt.Errorf("idProduct: %w", err)
		}
		d.IdProduct = &val
	}

	d.DeviceSpecific = raw.DeviceSpecific

	return nil
}

// parseUint16OrHex accepts either a JSON number or a hex string like "0x12ac"
func parseNumberOrHex[N constraints.Integer](v any) (N, error) {
	var zero N
	switch val := v.(type) {
	case float64:
		var minVal, maxVal float64
		switch any(zero).(type) {
		case int8:
			minVal, maxVal = math.MinInt8, math.MaxInt8
		case int16:
			minVal, maxVal = math.MinInt16, math.MaxInt16
		case int32:
			minVal, maxVal = math.MinInt32, math.MaxInt32
		case int64, int:
			minVal, maxVal = math.MinInt64, math.MaxInt64
		case uint8:
			minVal, maxVal = 0, math.MaxUint8
		case uint16:
			minVal, maxVal = 0, math.MaxUint16
		case uint32:
			minVal, maxVal = 0, math.MaxUint32
		case uint64, uint:
			minVal, maxVal = 0, math.MaxUint64
		default:
			return zero, fmt.Errorf("unsupported integer type %T", zero)
		}
		if val < minVal || val > maxVal {
			return zero, fmt.Errorf("value %v out of range for type %T", val, zero)
		}
		return N(val), nil
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
		var bitSize int
		switch any(zero).(type) {
		case int8, uint8:
			bitSize = 8
		case int16, uint16:
			bitSize = 16
		case int32, uint32:
			bitSize = 32
		case int64, uint64, int, uint:
			bitSize = 64
		default:
			return zero, fmt.Errorf("unsupported integer type %T", zero)
		}
		switch any(zero).(type) {
		case int, int8, int16, int32, int64:
			parsed, err := strconv.ParseInt(s, base, bitSize)
			if err != nil {
				return zero, fmt.Errorf("invalid hex/numeric string %q: %w", val, err)
			}
			return N(parsed), nil
		default:
			parsed, err := strconv.ParseUint(s, base, bitSize)
			if err != nil {
				return zero, fmt.Errorf("invalid hex/numeric string %q: %w", val, err)
			}
			return N(parsed), nil
		}
	default:
		return zero, fmt.Errorf("expected number or hex string, got %T", v)
	}
}
