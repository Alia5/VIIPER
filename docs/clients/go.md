# Go Client Documentation

The Go client is the reference implementation for interacting with VIIPER servers. It's included in the repository under `/apiclient` and `/device`.

## Overview

The Go client features:

- **Type-safe API**: Structured request/response types with context support
- **Device streams**: Bidirectional communication using `encoding.BinaryMarshaler`/`BinaryUnmarshaler`
- **Built-in**: No code generation needed; part of the main repository
- **Flexible timeouts**: Configurable connection and I/O timeouts

## Quick Start

```go
package main

import (
  "context"
  "log"
  "time"

  apiclient "github.com/Alia5/VIIPER/apiclient"
  "github.com/Alia5/VIIPER/device"
  "github.com/Alia5/VIIPER/device/keyboard"
)

func main() {
  // Create new Viiper client
  client := apiclient.New("127.0.0.1:3242")
  ctx := context.Background()
  
  // Create or find a bus
  buses, err := client.BusList()
  if err != nil {
    log.Fatal(err)
  }
  
  var busID uint32
  if len(buses) > 0 {
    busID = buses[0]
  } else {
    resp, err := client.BusCreate(nil)
    if err != nil {
      log.Fatal(err)
    }
    busID = resp.BusID
  }
  
  // Add device and connect (optional CreateOptions parameter for VID/PID)
  // Pass nil to use default VID/PID for the device type.
  stream, resp, err := client.AddDeviceAndConnect(ctx, busID, "keyboard", nil)
  if err != nil {
    log.Fatal(err)
  }
  defer stream.Close()
  
  log.Printf("Connected to device %s", resp.ID)
  
  // Send keyboard input
  input := &keyboard.InputState{
    Modifiers: keyboard.ModLeftShift,
  }
  input.SetKey(keyboard.KeyH, true)
  
  if err := stream.WriteBinary(input); err != nil {
    log.Fatal(err)
  }
  
  time.Sleep(100 * time.Millisecond)
  
  // Release
  input = &keyboard.InputState{}
  stream.WriteBinary(input)
}
```

## Device Stream API

### Creating and Connecting

// The simplest way to add a device and open its stream (nil opts):

```go
// Use default VID/PID for the device type
stream, resp, err := client.AddDeviceAndConnect(ctx, busID, "xbox360", nil)
if err != nil {
  log.Fatal(err)
}
defer stream.Close()

log.Printf("Connected to device %s", resp.ID)
```

// Or specify VID/PID using CreateOptions:

```go
opts := &device.CreateOptions{
  IdVendor:  func() *uint16 { v := uint16(0x1234); return &v }(),
  IdProduct: func() *uint16 { p := uint16(0x5678); return &p }(),
}
stream2, resp2, err := client.AddDeviceAndConnect(ctx, busID, "keyboard", opts)
if err != nil {
  log.Fatal(err)
}
defer stream2.Close()

log.Printf("Connected to device %s (custom VID/PID)", resp2.ID)
```

Or connect to an existing device:

```go
stream, err := client.OpenStream(ctx, busID, deviceID)
if err != nil {
  log.Fatal(err)
}
defer stream.Close()
```

### Sending Input

Device input is sent using structs that implement `encoding.BinaryMarshaler`:

```go
import "github.com/Alia5/VIIPER/device/xbox360"

input := &xbox360.InputState{
  Buttons: xbox360.ButtonA,
  LX:      -32768, // Left stick left
  LY:      32767,  // Left stick up
}
if err := stream.WriteBinary(input); err != nil {
  log.Fatal(err)
}
```

### Receiving Output (Callbacks)

For devices that send feedback (rumble, LEDs), use `StartReading` with a decode function:

```go
import (
  "bufio"
  "encoding"
  "io"
  "github.com/Alia5/VIIPER/device/xbox360"
)

// Start async reading for rumble commands
rumbleCh, errCh := stream.StartReading(ctx, 10, func(r *bufio.Reader) (encoding.BinaryUnmarshaler, error) {
  var b [2]byte
  if _, err := io.ReadFull(r, b[:]); err != nil { return nil, err }
  msg := new(xbox360.XRumbleState)
  if err := msg.UnmarshalBinary(b[:]); err != nil { return nil, err }
  return msg, nil
})

go func() {
  for {
    select {
    case msg := <-rumbleCh:
      rumble := msg.(*xbox360.XRumbleState)
      fmt.Printf("Rumble: Left=%d Right=%d\n", rumble.LeftMotor, rumble.RightMotor)
    case err := <-errCh:
      if err != nil { log.Printf("Stream error: %v", err) }
      return
    }
  }
}()
```

### Closing a Stream

```go
stream.Close()
```

## Device-Specific Notes

Each device type has specific wire formats and helper methods. For wire format details and usage patterns, see the [Devices](../devices/) section of the documentation.

The Go client provides device packages under `/device/` with type-safe structs and constants (e.g., `keyboard.InputState`, `keyboard.KeyA`, `mouse.Btn_Left`).

## Configuration and Advanced Usage

### Custom Timeouts

```go
cfg := &apiclient.Config{
  DialTimeout:  2 * time.Second,
  ReadTimeout:  3 * time.Second,
  WriteTimeout: 3 * time.Second,
}
client := apiclient.NewWithConfig("127.0.0.1:3242", cfg)
```

Default timeouts are: Dial 3s, Read/Write 5s.

### Context-Aware Calls

All methods have context-aware variants ending with `Ctx`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

buses, err := client.BusListCtx(ctx)
```

### Error Handling

The server returns errors as `{ "error": "message" }` JSON. The client wraps these as Go errors:

```go
if err != nil {
  log.Printf("request failed: %v", err)
}
```

## Examples

Full working examples are available in the repository:

- **Virtual Mouse**: `examples/go/virtual_mouse/main.go`
- **Virtual Keyboard**: `examples/go/virtual_keyboard/main.go`
- **Virtual Xbox360 Controller**: `examples/go/virtual_x360_pad/main.go`

## See Also

- [Generator Documentation](generator.md): How generated SDKs work
- [C SDK Documentation](c.md): Generated C SDK usage
- [C# SDK Documentation](csharp.md): .NET SDK
- [Rust SDK Documentation](rust.md): Rust SDK
- [TypeScript SDK Documentation](typescript.md): Node.js SDK
- [API Overview](../api/overview.md): Management API reference
