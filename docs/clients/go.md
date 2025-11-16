# Go client usage (timeouts and error handling)

This page summarizes how to configure timeouts and work with errors in the Go client.

## Constructing the client

```go
import (
  "time"
  apiclient "viiper/pkg/apiclient"
)

// Defaults are sensible (Dial 3s, Read/Write 5s)
c := apiclient.New("127.0.0.1:3242")

// Custom timeouts
cfg := &apiclient.Config{DialTimeout: 2 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second}
cc := apiclient.NewWithConfig("127.0.0.1:3242", cfg)
```

## Context-aware calls

All methods have context-aware variants ending with `Ctx`.

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

buses, err := cc.BusListCtx(ctx)
```

## Error handling

The server uses a uniform error envelope `{ "error": "message" }` and the client simply returns a Go `error` with that message.

```go
if err != nil {
  // err contains the server-provided message or a transport error
  log.Printf("request failed: %v", err)
}
```

## Strict JSON decoding

Responses are decoded with `DisallowUnknownFields` to fail fast if the server sends unexpected fields. This helps detect drift early.

## Device streams

Device streams provide bidirectional communication with virtual devices. The client includes native support for device types using `encoding.BinaryMarshaler`/`BinaryUnmarshaler`.

### Connecting to an existing device

```go
import (
  "viiper/pkg/device/xbox360"
)

stream, err := client.OpenStream(ctx, busID, deviceID)
if err != nil {
  log.Fatal(err)
}
defer stream.Close()

// Send input using device structs (client → device)
input := &xbox360.InputState{
  Buttons: 0x0001, // A button
  LX:      -32768, // Left stick left
  LY:      32767,  // Left stick up
}
if err := stream.WriteBinary(input); err != nil {
  log.Fatal(err)
}
```

### Receiving device feedback (event-driven)

For devices that send feedback (rumble, LEDs), you can use `StartReading` with a decode function to avoid polling. The decode function must read exactly one message from a `*bufio.Reader`.

```go
import (
  "bufio"
  "encoding"
  "io"
  "viiper/pkg/device/xbox360"
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

state := &xbox360.InputState{Buttons: 0x0001}
err := stream.WriteBinary(state)
...
```

### Creating and connecting in one step

```go
stream, resp, err := client.AddDeviceAndConnect(ctx, busID, "xbox360")
if err != nil {
  log.Fatal(err)
}
defer stream.Close()

log.Printf("Connected to device %s", resp.ID)
```
