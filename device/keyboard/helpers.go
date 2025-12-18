package keyboard

// CharToHID converts an ASCII character to its HID usage code.
// Returns 0 if the character is not supported.
func CharToHID(c byte) uint8 {
	if code, ok := CharToKey[c]; ok {
		return code
	}
	return 0
}

// NeedsShift returns true if the character requires the Shift modifier.
func NeedsShift(c byte) bool {
	return ShiftChars[c]
}

// TypeString converts a string into a sequence of InputState press/release pairs.
// Automatically handles shift modifiers for uppercase letters and symbols.
// Returns a slice of states alternating between press and release.
//
// Example:
//
//	states := TypeString("Hi!")
//	// Returns: [Press Shift+H, Release, Press i, Release, Press Shift+1, Release]
func TypeString(s string) []InputState {
	var states []InputState
	for i := 0; i < len(s); i++ {
		c := s[i]
		press, release := TypeChar(c)
		states = append(states, press, release)
	}
	return states
}

// TypeChar converts a single character to a press/release InputState pair.
// Automatically adds Shift modifier if needed.
func TypeChar(c byte) (press, release InputState) {
	keyCode := CharToHID(c)
	if keyCode == 0 {
		// Unsupported character, return empty states
		return InputState{}, InputState{}
	}

	modifiers := uint8(0)
	if NeedsShift(c) {
		modifiers = ModLeftShift
	}

	press = PressKeyWithMod(modifiers, keyCode)
	release = Release()
	return
}

// PressKey creates an InputState with the specified keys pressed.
// No modifiers are set.
//
// Example:
//
//	state := PressKey(KeyA, KeyB) // Press A and B simultaneously
func PressKey(keys ...uint8) InputState {
	return PressKeyWithMod(0, keys...)
}

// PressKeyWithMod creates an InputState with modifiers and keys pressed.
//
// Example:
//
//	state := PressKeyWithMod(ModLeftCtrl, KeyC) // Ctrl+C
//	state := PressKeyWithMod(ModLeftShift, KeyA) // Shift+A
func PressKeyWithMod(modifiers uint8, keys ...uint8) InputState {
	var state InputState
	state.Modifiers = modifiers

	// Set bits for each key in the bitmap
	for _, key := range keys {
		byteIdx := key / 8
		bitIdx := uint(key % 8)
		state.KeyBitmap[byteIdx] |= 1 << bitIdx
	}

	return state
}

// Release creates an empty InputState with all keys released.
func Release() InputState {
	return InputState{}
}
