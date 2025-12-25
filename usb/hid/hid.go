// Package hid provides a structured representation of HID report descriptors.
//
// A HID report descriptor is a byte-coded DSL. This package models it as a tree
// of Go structs (including nested collections) and encodes it to the exact
// descriptor byte stream.
package hid

import (
	"fmt"
)

// Data is a strongly-typed byte slice used for HID report descriptor payloads.
//
// It exists to avoid exposing raw []byte fields on report descriptor models.
// The underlying representation is still bytes because that is what the USB/HID
// specification ultimately requires.
type Data []uint8

// ItemType is the HID short item "type" field.
// See HID 1.11 spec: Main=0, Global=1, Local=2, Reserved=3.
type ItemType uint8

const (
	ItemTypeMain     ItemType = 0
	ItemTypeGlobal   ItemType = 1
	ItemTypeLocal    ItemType = 2
	ItemTypeReserved ItemType = 3
)

// Item is one node in a HID report descriptor.
type Item interface {
	encode(e *encoder) error
}

// Report is a complete HID report descriptor (type 0x22).
type Report struct {
	Items []Item
}

// Bytes encodes the report descriptor.
func (r Report) Bytes() (Data, error) {
	e := &encoder{}
	for _, it := range r.Items {
		if it == nil {
			return nil, fmt.Errorf("hid: nil item")
		}
		if err := it.encode(e); err != nil {
			return nil, err
		}
	}
	return Data(e.buf), nil
}

// AnyItem is an escape hatch for rarely used or vendor-defined items.
//
// For short items, Data must have length 0, 1, 2, or 4.
// For other sizes, use LongItem.
type AnyItem struct {
	Type ItemType
	Tag  uint8
	Data Data
}

func (a AnyItem) encode(e *encoder) error {
	n := len(a.Data)
	var sizeCode uint8
	switch n {
	case 0:
		sizeCode = 0
	case 1:
		sizeCode = 1
	case 2:
		sizeCode = 2
	case 4:
		sizeCode = 3
	default:
		return fmt.Errorf("hid: AnyItem short item data must be 0/1/2/4 bytes, got %d", n)
	}
	header := (a.Tag << 4) | (uint8(a.Type) << 2) | sizeCode
	e.buf = append(e.buf, header)
	e.buf = append(e.buf, a.Data...)
	return nil
}

// LongItem encodes a HID long item (rare). Format: 0xFE, len, tag, data...
type LongItem struct {
	Tag  uint8
	Data Data
}

func (l LongItem) encode(e *encoder) error {
	if len(l.Data) > 255 {
		return fmt.Errorf("hid: long item too large: %d", len(l.Data))
	}
	e.buf = append(e.buf, 0xFE, uint8(len(l.Data)), l.Tag)
	e.buf = append(e.buf, l.Data...)
	return nil
}

type encoder struct {
	buf []byte
}

func (e *encoder) short(tag uint8, typ ItemType, data Data) error {
	n := len(data)
	var sizeCode uint8
	switch n {
	case 0:
		sizeCode = 0
	case 1:
		sizeCode = 1
	case 2:
		sizeCode = 2
	case 4:
		sizeCode = 3
	default:
		return fmt.Errorf("hid: short item data must be 0/1/2/4 bytes, got %d", n)
	}
	header := (tag << 4) | (uint8(typ) << 2) | sizeCode
	e.buf = append(e.buf, header)
	e.buf = append(e.buf, data...)
	return nil
}

func dataU32(v uint32) Data {
	if v <= 0xFF {
		return Data{uint8(v)}
	}
	if v <= 0xFFFF {
		return Data{uint8(v), uint8(v >> 8)}
	}
	return Data{uint8(v), uint8(v >> 8), uint8(v >> 16), uint8(v >> 24)}
}

func dataI32(v int32) Data {
	if v >= -128 && v <= 127 {
		return Data{uint8(v)}
	}
	if v >= -32768 && v <= 32767 {
		uv := uint16(int16(v))
		return Data{uint8(uv), uint8(uv >> 8)}
	}
	uv := uint32(v)
	return Data{uint8(uv), uint8(uv >> 8), uint8(uv >> 16), uint8(uv >> 24)}
}
