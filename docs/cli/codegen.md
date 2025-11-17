# Code Generation Command

The `codegen` command generates type-safe client SDKs from Go source code annotations.

## Usage

```bash
viiper codegen [flags]
```

## Description

Scans the VIIPER server codebase to extract:

- API routes and response DTOs
- Device wire formats from `viiper:wire` comment tags
- Device constants (keycodes, modifiers, button masks)

Then generates client SDKs with:

- Management API clients
- Device-agnostic stream wrappers
- Per-device encode/decode functions
- Typed constants and enums

## Flags

### `--output`

Output directory for generated SDKs (relative to repository root).

**Default:** `../clients`  
**Environment Variable:** `VIIPER_CODEGEN_OUTPUT`

**Example:**

```bash
viiper codegen --output=../sdk-output
```

### `--lang`

Target language to generate.

**Values:** `c`, `csharp`, `typescript`, `all`  
**Default:** `all`  
**Environment Variable:** `VIIPER_CODEGEN_LANG`

**Examples:**

```bash
# Generate all SDKs
viiper codegen --lang=all

# Generate C SDK only
viiper codegen --lang=c

# Generate C# SDK only
viiper codegen --lang=csharp

# Generate TypeScript SDK only
viiper codegen --lang=typescript
```

## Examples

### Generate All SDKs

```bash
cd viiper
go run ./cmd/viiper codegen
```

### Generate C SDK and Rebuild Examples

```bash
cd viiper
go run ./cmd/viiper codegen --lang=c
cd ../examples/c
cmake --build build --config Release
```

### CI/CD Integration

```yaml
- name: Generate Client SDKs
  run: |
    cd viiper
    go run ./cmd/viiper codegen --lang=all
```

## When to Regenerate

Run codegen when any of these change:

- `pkg/apitypes/*.go`: API response structures
- `pkg/device/*/inputstate.go`: Wire format annotations
- `pkg/device/*/const.go`: Exported constants
- `internal/server/api/*.go`: Route registrations
- Generator templates in `internal/codegen/generator/`

## See Also

- [Generator Documentation](../clients/generator.md): Detailed explanation of tagging system and code generation flow
- [C SDK Documentation](../clients/c.md): C-specific usage and build instructions
- [Configuration](configuration.md): Global configuration options
