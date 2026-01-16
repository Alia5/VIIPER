# VIIPER Client Generator Documentation

## Overview

The VIIPER client generator scans Go source code to extract API routes, device wire formats, and constants; then emits type-safe client libraries for multiple languages.

**What it extracts:**

- API routes and DTOs from management API handlers  
- Device wire formats from `viiper:wire` comment tags  
- All exported constants from device packages (automatic)

**Output:** Type-safe client libraries for multiple target languages

!!! note "License"
    All generated client libraries are licensed under the **MIT License**, providing maximum flexibility for integration into your projects. The core VIIPER server remains under its original license.

## Running the Generator

```bash
go run ./cmd/viiper codegen --lang=all        # Generate all client libraries
go run ./cmd/viiper codegen --lang=c          # Generate C client library only
go run ./cmd/viiper codegen --lang=csharp     # Generate C# client library only
go run ./cmd/viiper codegen --lang=typescript # Generate TypeScript client library only
```

**Output directory**: `clients/` (relative to repository root)

## Comment Tag System

The generator uses lightweight comment tags placed next to device types and constants.

### `viiper:wire`: Device Stream Formats

**Syntax:**

```go
// viiper:wire <device> <direction> <field1:type> <field2:type> ...
```

**Directions:**  

- `c2s`: Client to server (input)  
- `s2c`: Server to client (output, e.g., rumble, LEDs)

**Field types:**  

- Fixed: `u8`, `i8`, `u16`, `i16`, `u32`, `i32`  
- Variable: `u8*countField` (pointer to count field)

**Example:**

```go
// viiper:wire keyboard c2s modifiers:u8 count:u8 keys:u8*count
type InputState struct { ... }
```

### Constant and Map Export

The generator automatically exports all constants and map literals from `/device/*/const.go` for each device type.  
No special tags are required. Exported Go constants and maps are emitted with language-appropriate representations:

- **Constants**: Grouped into enums (C#/TS) or `#define` macros (C) based on common prefixes
- **Maps**: Converted to Dictionary/Map/lookup functions with helper methods

## Code Generation Flow

**Scan Phase:**  

1. Parse API routes from `internal/server/api/*.go`  
2. Reflect response DTOs from `/apitypes/*.go`  
3. Find device types via `RegisterDevice()` calls  
4. Parse `viiper:wire` comments for packet layouts  
5. Extract all exported constants and map literals from `/device/*/const.go` (automatic)

**Emit Phase:**  
For each language, generate management client, DTO types, device streams, constants, and build configs.

**Post-Process:**  
Optional formatting with `clang-format`, `dotnet format`, or `prettier`.

## Wire Format Mapping Rules

### Fixed-Size Fields

Fixed-size fields are mapped to native integer types in each target language:

- `u8` / `i8`: 8-bit unsigned/signed integers
- `u16` / `i16`: 16-bit unsigned/signed integers
- `u32` / `i32`: 32-bit unsigned/signed integers

### Variable-Length Fields

Variable-length arrays use a **pointer + count** pattern. The field syntax `u8*count` references a count field that determines the array length.

**Wire tag example:**

```go
// viiper:wire keyboard c2s modifiers:u8 count:u8 keys:u8*count
```

Each target language emits appropriate types for dynamic arrays (pointers with counts, managed arrays, or typed arrays depending on the language).

## Struct Packing

For wire compatibility, all device I/O structs are tightly packed (no padding).

- **C:** `#pragma pack(push, 1)` / `#pragma pack(pop)`
- **C#:** `[StructLayout(LayoutKind.Sequential, Pack = 1)]`
- **TypeScript:** Manual byte-level encoding/decoding

## Example: Keyboard Input (Variable-Length)

**Go source with wire tag:**

```go
// viiper:wire keyboard c2s modifiers:u8 count:u8 keys:u8*count
type InputState struct {
    Modifiers uint8
    KeyBitmap [32]uint8  // Internal: 256-bit NKR bitmap
}
```

**Emitted C struct:**

```c
#pragma pack(push, 1)
typedef struct {
    uint8_t modifiers;
    uint8_t count;
    uint8_t* keys;
    size_t keys_count;
} viiper_keyboard_input_t;
#pragma pack(pop)
```

## Example: Constant and Map Export

**Go source (`/device/keyboard/const.go`):**

```go
const (
    ModLeftCtrl  = 0x01
    ModLeftShift = 0x02
    KeyA = 0x04
    KeyB = 0x05
    // ...
)

var CharToKey = map[byte]byte{
    'a': KeyA,
    'b': KeyB,
    '\n': KeyEnter,
    // ...
}
```

**Emitted C header (`viiper_keyboard.h`):**

```c
#define VIIPER_KEYBOARD_MOD_LEFT_CTRL 0x1
#define VIIPER_KEYBOARD_MOD_LEFT_SHIFT 0x2
#define VIIPER_KEYBOARD_KEY_A 0x4
#define VIIPER_KEYBOARD_KEY_B 0x5

// Map lookup function
int viiper_keyboard_char_to_key_lookup(uint8_t key, uint8_t* out_value);
```

**Emitted C# (`KeyboardConstants.cs`):**

```csharp
public enum Mod : uint
{
    LeftCtrl = 0x01,
    LeftShift = 0x02,
    // ...
}

public enum Key : uint
{
    A = 0x04,
    B = 0x05,
    // ...
}

public static class CharToKey
{
    private static readonly Dictionary<byte, Key> _map = new()
    {
        { (byte)'a', Key.A },
        { (byte)'b', Key.B },
        { (byte)'\n', Key.Enter },
        // ...
    };

    public static bool TryGetValue(byte key, out Key value)
    {
        return _map.TryGetValue(key, out value);
    }
}
```

## Regeneration Triggers

Run codegen when any of these change:

- `/apitypes/*.go`: API response structures
- `/device/*/inputstate.go`: Wire tag annotations
- `/device/*/const.go`: Exported constants and map literals
- `internal/server/api/*.go`: Route registrations
- `internal/codegen/generator/**/*.go`: Generator templates
- `internal/codegen/scanner/**/*.go`: Scanner logic (constants, maps, wire tags)

## Language-Specific Notes

- **C**: `#define` macros for constants; switch-based lookup functions for maps; manual memory management for variable-length fields; builds with CMake.  
- **C#**: Enums for constant groups; `Dictionary<K,V>` with static helper methods for maps; `ViiperDevice` class with `OnOutput` event; async/await for management API; struct packing via attributes.  
- **TypeScript**: Enums for constant groups; `Record<K, V>` objects with `Get`/`Has` helper functions for maps; manual byte encoding via `BinaryWriter`/`BinaryReader`; `ViiperDevice` class with EventEmitter for output; `addDeviceAndConnect` convenience method; builds with `tsc`.  

## Further Reading

- [Go Client Documentation](go.md): Go reference client usage
- [C Client Library Documentation](c.md): C-specific usage, build, and examples
- [C# Client Library Documentation](csharp.md): C#-specific usage, async patterns, and map helpers
- [TypeScript Client Library Documentation](typescript.md): TypeScript-specific usage, EventEmitter patterns, and examples
