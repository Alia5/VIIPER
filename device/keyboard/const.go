package keyboard

// Modifier key bitmasks
const (
	ModLeftCtrl   = 0x01
	ModLeftShift  = 0x02
	ModLeftAlt    = 0x04
	ModLeftGUI    = 0x08 // Windows/Command key
	ModRightCtrl  = 0x10
	ModRightShift = 0x20
	ModRightAlt   = 0x40
	ModRightGUI   = 0x80
)

// LED bitmasks
const (
	LEDNumLock    = 0x01
	LEDCapsLock   = 0x02
	LEDScrollLock = 0x04
	LEDCompose    = 0x08
	LEDKana       = 0x10
)

// HID Usage codes for keyboard keys (USB HID Keyboard/Keypad usage page)
const (
	// Letters A-Z
	KeyA = 0x04
	KeyB = 0x05
	KeyC = 0x06
	KeyD = 0x07
	KeyE = 0x08
	KeyF = 0x09
	KeyG = 0x0A
	KeyH = 0x0B
	KeyI = 0x0C
	KeyJ = 0x0D
	KeyK = 0x0E
	KeyL = 0x0F
	KeyM = 0x10
	KeyN = 0x11
	KeyO = 0x12
	KeyP = 0x13
	KeyQ = 0x14
	KeyR = 0x15
	KeyS = 0x16
	KeyT = 0x17
	KeyU = 0x18
	KeyV = 0x19
	KeyW = 0x1A
	KeyX = 0x1B
	KeyY = 0x1C
	KeyZ = 0x1D

	// Numbers 1-0 (top row)
	Key1 = 0x1E
	Key2 = 0x1F
	Key3 = 0x20
	Key4 = 0x21
	Key5 = 0x22
	Key6 = 0x23
	Key7 = 0x24
	Key8 = 0x25
	Key9 = 0x26
	Key0 = 0x27

	// Special keys
	KeyEnter      = 0x28
	KeyEscape     = 0x29
	KeyBackspace  = 0x2A
	KeyTab        = 0x2B
	KeySpace      = 0x2C
	KeyMinus      = 0x2D // - and _
	KeyEqual      = 0x2E // = and +
	KeyLeftBrace  = 0x2F // [ and {
	KeyRightBrace = 0x30 // ] and }
	KeyBackslash  = 0x31 // \ and |
	KeyNonUSHash  = 0x32 // Non-US # and ~
	KeySemicolon  = 0x33 // ; and :
	KeyApostrophe = 0x34 // ' and "
	KeyGrave      = 0x35 // ` and ~
	KeyComma      = 0x36 // , and <
	KeyPeriod     = 0x37 // . and >
	KeySlash      = 0x38 // / and ?
	KeyCapsLock   = 0x39

	// Function keys
	KeyF1  = 0x3A
	KeyF2  = 0x3B
	KeyF3  = 0x3C
	KeyF4  = 0x3D
	KeyF5  = 0x3E
	KeyF6  = 0x3F
	KeyF7  = 0x40
	KeyF8  = 0x41
	KeyF9  = 0x42
	KeyF10 = 0x43
	KeyF11 = 0x44
	KeyF12 = 0x45

	// Control keys
	KeyPrintScreen = 0x46
	KeyScrollLock  = 0x47
	KeyPause       = 0x48
	KeyInsert      = 0x49
	KeyHome        = 0x4A
	KeyPageUp      = 0x4B
	KeyDelete      = 0x4C
	KeyEnd         = 0x4D
	KeyPageDown    = 0x4E

	// Arrow keys
	KeyRight = 0x4F
	KeyLeft  = 0x50
	KeyDown  = 0x51
	KeyUp    = 0x52

	// Numpad
	KeyNumLock    = 0x53
	KeyKpSlash    = 0x54 // Keypad /
	KeyKpAsterisk = 0x55 // Keypad *
	KeyKpMinus    = 0x56 // Keypad -
	KeyKpPlus     = 0x57 // Keypad +
	KeyKpEnter    = 0x58 // Keypad Enter
	KeyKp1        = 0x59 // Keypad 1 and End
	KeyKp2        = 0x5A // Keypad 2 and Down
	KeyKp3        = 0x5B // Keypad 3 and PageDn
	KeyKp4        = 0x5C // Keypad 4 and Left
	KeyKp5        = 0x5D // Keypad 5
	KeyKp6        = 0x5E // Keypad 6 and Right
	KeyKp7        = 0x5F // Keypad 7 and Home
	KeyKp8        = 0x60 // Keypad 8 and Up
	KeyKp9        = 0x61 // Keypad 9 and PageUp
	KeyKp0        = 0x62 // Keypad 0 and Insert
	KeyKpDot      = 0x63 // Keypad . and Delete

	// Additional keys
	KeyNonUSBackslash = 0x64 // Non-US \ and |
	KeyApplication    = 0x65 // Application (Windows Menu key)
	KeyPower          = 0x66 // Power (not commonly used)
	KeyKpEqual        = 0x67 // Keypad =

	// Extended function keys
	KeyF13 = 0x68
	KeyF14 = 0x69
	KeyF15 = 0x6A
	KeyF16 = 0x6B
	KeyF17 = 0x6C
	KeyF18 = 0x6D
	KeyF19 = 0x6E
	KeyF20 = 0x6F
	KeyF21 = 0x70
	KeyF22 = 0x71
	KeyF23 = 0x72
	KeyF24 = 0x73

	// Execution keys
	KeyExecute    = 0x74
	KeyHelp       = 0x75
	KeyMenu       = 0x76
	KeySelect     = 0x77
	KeyStop       = 0x78
	KeyAgain      = 0x79 // Redo
	KeyUndo       = 0x7A
	KeyCut        = 0x7B
	KeyCopy       = 0x7C
	KeyPaste      = 0x7D
	KeyFind       = 0x7E
	KeyMute       = 0x7F
	KeyVolumeUp   = 0x80
	KeyVolumeDown = 0x81

	// Media control keys
	KeyMediaPlayPause = 0xE8 // Play/Pause
	KeyMediaStop      = 0xE9 // Stop
	KeyMediaNext      = 0xEB // Next Track
	KeyMediaPrevious  = 0xEC // Previous Track
)

// KeyName maps HID usage codes to human-readable key names.
var KeyName = map[uint8]string{
	// Letters
	KeyA: "A", KeyB: "B", KeyC: "C", KeyD: "D", KeyE: "E", KeyF: "F", KeyG: "G",
	KeyH: "H", KeyI: "I", KeyJ: "J", KeyK: "K", KeyL: "L", KeyM: "M", KeyN: "N",
	KeyO: "O", KeyP: "P", KeyQ: "Q", KeyR: "R", KeyS: "S", KeyT: "T", KeyU: "U",
	KeyV: "V", KeyW: "W", KeyX: "X", KeyY: "Y", KeyZ: "Z",

	// Numbers
	Key1: "1", Key2: "2", Key3: "3", Key4: "4", Key5: "5",
	Key6: "6", Key7: "7", Key8: "8", Key9: "9", Key0: "0",

	// Special keys
	KeyEnter:      "Enter",
	KeyEscape:     "Escape",
	KeyBackspace:  "Backspace",
	KeyTab:        "Tab",
	KeySpace:      "Space",
	KeyMinus:      "Minus",
	KeyEqual:      "Equal",
	KeyLeftBrace:  "LeftBrace",
	KeyRightBrace: "RightBrace",
	KeyBackslash:  "Backslash",
	KeySemicolon:  "Semicolon",
	KeyApostrophe: "Apostrophe",
	KeyGrave:      "Grave",
	KeyComma:      "Comma",
	KeyPeriod:     "Period",
	KeySlash:      "Slash",
	KeyCapsLock:   "CapsLock",

	// Function keys
	KeyF1: "F1", KeyF2: "F2", KeyF3: "F3", KeyF4: "F4", KeyF5: "F5", KeyF6: "F6",
	KeyF7: "F7", KeyF8: "F8", KeyF9: "F9", KeyF10: "F10", KeyF11: "F11", KeyF12: "F12",
	KeyF13: "F13", KeyF14: "F14", KeyF15: "F15", KeyF16: "F16", KeyF17: "F17", KeyF18: "F18",
	KeyF19: "F19", KeyF20: "F20", KeyF21: "F21", KeyF22: "F22", KeyF23: "F23", KeyF24: "F24",

	// Control keys
	KeyPrintScreen: "PrintScreen",
	KeyScrollLock:  "ScrollLock",
	KeyPause:       "Pause",
	KeyInsert:      "Insert",
	KeyHome:        "Home",
	KeyPageUp:      "PageUp",
	KeyDelete:      "Delete",
	KeyEnd:         "End",
	KeyPageDown:    "PageDown",

	// Arrow keys
	KeyRight: "Right",
	KeyLeft:  "Left",
	KeyDown:  "Down",
	KeyUp:    "Up",

	// Numpad
	KeyNumLock:    "NumLock",
	KeyKpSlash:    "Kp/",
	KeyKpAsterisk: "Kp*",
	KeyKpMinus:    "Kp-",
	KeyKpPlus:     "Kp+",
	KeyKpEnter:    "KpEnter",
	KeyKp1:        "Kp1",
	KeyKp2:        "Kp2",
	KeyKp3:        "Kp3",
	KeyKp4:        "Kp4",
	KeyKp5:        "Kp5",
	KeyKp6:        "Kp6",
	KeyKp7:        "Kp7",
	KeyKp8:        "Kp8",
	KeyKp9:        "Kp9",
	KeyKp0:        "Kp0",
	KeyKpDot:      "Kp.",

	// Additional
	KeyApplication: "Application",
	KeyMute:        "Mute",
	KeyVolumeUp:    "VolumeUp",
	KeyVolumeDown:  "VolumeDown",

	// Media control
	KeyMediaPlayPause: "MediaPlayPause",
	KeyMediaStop:      "MediaStop",
	KeyMediaNext:      "MediaNext",
	KeyMediaPrevious:  "MediaPrevious",
}

// CharToKey maps ASCII characters to their corresponding HID usage codes.
// For shifted characters (uppercase, symbols), use with NeedsShift().
var CharToKey = map[byte]uint8{
	// Lowercase letters
	'a': KeyA, 'b': KeyB, 'c': KeyC, 'd': KeyD, 'e': KeyE, 'f': KeyF, 'g': KeyG,
	'h': KeyH, 'i': KeyI, 'j': KeyJ, 'k': KeyK, 'l': KeyL, 'm': KeyM, 'n': KeyN,
	'o': KeyO, 'p': KeyP, 'q': KeyQ, 'r': KeyR, 's': KeyS, 't': KeyT, 'u': KeyU,
	'v': KeyV, 'w': KeyW, 'x': KeyX, 'y': KeyY, 'z': KeyZ,

	// Uppercase letters (same keys, need shift)
	'A': KeyA, 'B': KeyB, 'C': KeyC, 'D': KeyD, 'E': KeyE, 'F': KeyF, 'G': KeyG,
	'H': KeyH, 'I': KeyI, 'J': KeyJ, 'K': KeyK, 'L': KeyL, 'M': KeyM, 'N': KeyN,
	'O': KeyO, 'P': KeyP, 'Q': KeyQ, 'R': KeyR, 'S': KeyS, 'T': KeyT, 'U': KeyU,
	'V': KeyV, 'W': KeyW, 'X': KeyX, 'Y': KeyY, 'Z': KeyZ,

	// Numbers (top row)
	'1': Key1, '2': Key2, '3': Key3, '4': Key4, '5': Key5,
	'6': Key6, '7': Key7, '8': Key8, '9': Key9, '0': Key0,

	// Shifted number row symbols
	'!': Key1, '@': Key2, '#': Key3, '$': Key4, '%': Key5,
	'^': Key6, '&': Key7, '*': Key8, '(': Key9, ')': Key0,

	// Unshifted symbols
	'-':  KeyMinus,
	'=':  KeyEqual,
	'[':  KeyLeftBrace,
	']':  KeyRightBrace,
	'\\': KeyBackslash,
	';':  KeySemicolon,
	'\'': KeyApostrophe,
	'`':  KeyGrave,
	',':  KeyComma,
	'.':  KeyPeriod,
	'/':  KeySlash,

	// Shifted symbols
	'_': KeyMinus,
	'+': KeyEqual,
	'{': KeyLeftBrace,
	'}': KeyRightBrace,
	'|': KeyBackslash,
	':': KeySemicolon,
	'"': KeyApostrophe,
	'~': KeyGrave,
	'<': KeyComma,
	'>': KeyPeriod,
	'?': KeySlash,

	// Whitespace
	' ':  KeySpace,
	'\n': KeyEnter,
	'\r': KeyEnter,
	'\t': KeyTab,
}

// ShiftChars defines which characters require the Shift modifier.
var ShiftChars = map[byte]bool{
	// Uppercase letters
	'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true, 'G': true,
	'H': true, 'I': true, 'J': true, 'K': true, 'L': true, 'M': true, 'N': true,
	'O': true, 'P': true, 'Q': true, 'R': true, 'S': true, 'T': true, 'U': true,
	'V': true, 'W': true, 'X': true, 'Y': true, 'Z': true,

	// Shifted number row
	'!': true, '@': true, '#': true, '$': true, '%': true,
	'^': true, '&': true, '*': true, '(': true, ')': true,

	// Shifted symbols
	'_': true, '+': true, '{': true, '}': true, '|': true,
	':': true, '"': true, '~': true, '<': true, '>': true, '?': true,
}
