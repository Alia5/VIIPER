package hid

// Common item structs.

// UsagePage sets the current usage page (Global item, tag 0x0).
type UsagePage struct{ Page uint16 }

func (u UsagePage) encode(e *encoder) error {
	return e.short(0x0, ItemTypeGlobal, dataU32(uint32(u.Page)))
}

// Usage sets the current usage (Local item, tag 0x0).
type Usage struct{ Usage uint16 }

func (u Usage) encode(e *encoder) error {
	return e.short(0x0, ItemTypeLocal, dataU32(uint32(u.Usage)))
}

// Collection begins a collection (Main item, tag 0xA) and implicitly ends it.
type Collection struct {
	Kind  CollectionKind
	Items []Item
}

func (c Collection) encode(e *encoder) error {
	if err := e.short(0xA, ItemTypeMain, Data{uint8(c.Kind)}); err != nil {
		return err
	}
	for _, it := range c.Items {
		if err := it.encode(e); err != nil {
			return err
		}
	}
	// End Collection (Main item, tag 0xC) with 0 bytes.
	return e.short(0xC, ItemTypeMain, nil)
}

// UsageMinimum sets the usage minimum (Local item, tag 0x1).
type UsageMinimum struct{ Min uint16 }

func (u UsageMinimum) encode(e *encoder) error {
	return e.short(0x1, ItemTypeLocal, dataU32(uint32(u.Min)))
}

// UsageMaximum sets the usage maximum (Local item, tag 0x2).
type UsageMaximum struct{ Max uint16 }

func (u UsageMaximum) encode(e *encoder) error {
	return e.short(0x2, ItemTypeLocal, dataU32(uint32(u.Max)))
}

// LogicalMinimum sets the logical minimum (Global item, tag 0x1).
type LogicalMinimum struct{ Min int32 }

func (l LogicalMinimum) encode(e *encoder) error {
	return e.short(0x1, ItemTypeGlobal, dataI32(l.Min))
}

// LogicalMaximum sets the logical maximum (Global item, tag 0x2).
type LogicalMaximum struct{ Max int32 }

func (l LogicalMaximum) encode(e *encoder) error {
	return e.short(0x2, ItemTypeGlobal, dataI32(l.Max))
}

// ReportSize sets report size in bits (Global item, tag 0x7).
type ReportSize struct{ Bits uint8 }

func (r ReportSize) encode(e *encoder) error {
	return e.short(0x7, ItemTypeGlobal, Data{r.Bits})
}

// ReportCount sets report count (Global item, tag 0x9).
type ReportCount struct{ Count uint16 }

func (r ReportCount) encode(e *encoder) error {
	return e.short(0x9, ItemTypeGlobal, dataU32(uint32(r.Count)))
}

// Input encodes an Input main item (tag 0x8).
type Input struct{ Flags MainFlags }

func (i Input) encode(e *encoder) error {
	return e.short(0x8, ItemTypeMain, Data{uint8(i.Flags)})
}

// Output encodes an Output main item (tag 0x9).
type Output struct{ Flags MainFlags }

func (o Output) encode(e *encoder) error {
	return e.short(0x9, ItemTypeMain, Data{uint8(o.Flags)})
}

// Feature encodes a Feature main item (tag 0xB).
type Feature struct{ Flags MainFlags }

func (f Feature) encode(e *encoder) error {
	return e.short(0xB, ItemTypeMain, Data{uint8(f.Flags)})
}
