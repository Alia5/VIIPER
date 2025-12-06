# Code Generation Command

The `codegen` command generates type-safe client libraries from Go source code annotations.

## Usage

```bash
viiper codegen [flags]
```

## Description

Scans the VIIPER server codebase to extract:

- API routes and response DTOs
- Device wire formats from `viiper:wire` comment tags
- Device constants (keycodes, modifiers, button masks)

Then generates client libraries with:

- Management API clients
- Device-agnostic stream wrappers
- Per-device encode/decode functions
- Typed constants and enums

!!! note "Sourcecode access is required"
    The codegen command requires access to VIIPER source code. Run it from the repository root.

## Flags

### `--output`

Output directory for generated client libraries (relative to repository root).

**Default:** `clients`  
**Environment Variable:** `VIIPER_CODEGEN_OUTPUT`

**Example:**

```bash
viiper codegen --output=../client-libs-output
```

### `--lang`

Target language to generate.

**Values:** `c`, `csharp`, `typescript`, `all`  
**Default:** `all`  
**Environment Variable:** `VIIPER_CODEGEN_LANG`

**Examples:**

```bash
# Generate all client libraries
viiper codegen --lang=all

# Generate C client library only
viiper codegen --lang=c

# Generate C# client library only
viiper codegen --lang=csharp

# Generate TypeScript client library only
viiper codegen --lang=typescript
```

## Examples

### Generate All Client Libraries

```bash
go run ./cmd/viiper codegen
```

### Generate C Client Library and Rebuild Examples

```bash
go run ./cmd/viiper codegen --lang=c
cd examples/c
cmake --build build --config Release
```

## When to Regenerate

Run codegen when any of these change:

- `/apitypes/*.go`: API response structures
- `/device/*/inputstate.go`: Wire format annotations
- `/device/*/const.go`: Exported constants
- `internal/server/api/*.go`: Route registrations
- Generator templates in `internal/codegen/generator/`

## See Also

- [Generator Documentation](../clients/generator.md): Detailed explanation of tagging system and code generation flow
- [C Client Library Documentation](../clients/c.md): C-specific usage and build instructions
- [Configuration](configuration.md): Global configuration options
