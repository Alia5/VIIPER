# VIIPER Client Generator Documentation

## Overview

The VIIPER client generator scans Go source code to extract API routes, device wire formats, and constants; then emits type-safe client SDKs for multiple languages.

**What it extracts:**

- API routes and DTOs from management API handlers  
- Device wire formats from `viiper:wire` comment tags  
- All exported constants from device packages (automatic)

**Output:** Type-safe client SDKs for multiple target languages

!!! note "License"
    All generated client SDKs are licensed under the **MIT License**, providing maximum flexibility for integration into your projects. The core VIIPER server remains under its original license.

## Running the Generator

```bash
cd viiper
go run ./cmd/viiper codegen --lang=all     # Generate all SDKs
go run ./cmd/viiper codegen --lang=c       # Generate C SDK only
go run ./cmd/viiper codegen --lang=csharp  # Generate C# SDK only
go run ./cmd/viiper codegen --lang=typescript # Generate TypeScript SDK only
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

### Constant Export

The generator automatically exports all constants from `pkg/device/*/const.go` for each device type.  
No special tags are required. Exported Go constants are emitted with language-appropriate representations and prefixes.

## Code Generation Flow

**Scan Phase:**  

1. Parse API routes from `internal/server/api/*.go`  
2. Reflect response DTOs from `pkg/apitypes/*.go`  
3. Find device types via `RegisterDevice()` calls  
4. Parse `viiper:wire` comments for packet layouts  
5. Extract all exported constants from `pkg/device/*/const.go` (automatic)

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

## Example: Constant Export

**Go source (`pkg/device/keyboard/const.go`):**

```go
const (
    ModLeftCtrl  = 0x01
    ModLeftShift = 0x02
    KeyA = 0x04
    KeyB = 0x05
    // ...
)
```

**Emitted C header (`viiper_keyboard.h`):**

```c
#define VIIPER_KEYBOARD_MODLEFTCTRL 0x1
#define VIIPER_KEYBOARD_MODLEFTSHIFT 0x2
#define VIIPER_KEYBOARD_KEYA 0x4
#define VIIPER_KEYBOARD_KEYB 0x5
// ...
```

## Regeneration Triggers

Run codegen when any of these change:

- `pkg/apitypes/*.go`: API response structures
- `pkg/device/*/inputstate.go`: Wire tag annotations
- `pkg/device/*/const.go`: Exported constants
- `internal/server/api/*.go`: Route registrations
- `internal/codegen/generator/**/*.go`: Generator templates

## Language-Specific Notes

- **C**: manual memory management for variable-length fields; builds with CMake.  
- **C#**: `ViiperDevice` class with `OnOutput` event; async/await for management API; struct packing via attributes.  
- **TypeScript**: manual byte encoding; ESM/CJS compatible.  

## Current SDK Status

- **C**: ✅ Complete
- **C#**: 🚧 Planned
- **TypeScript**: 🚧 Planned

## Further Reading

- [Design Document](../design.md): Architectural rationale and detailed generation strategy
- [C SDK Documentation](c.md): C-specific usage, build, and examples
- [Go Client Documentation](go.md): Go reference client usage

---

For questions or contributions, see the main VIIPER repository.
