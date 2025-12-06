# Xbox 360 Controller

The Xbox 360 virtual gamepad emulates an XInput-compatible controller that most operating systems and games understand out of the box.

- USB IDs: VID 0x045E (Microsoft), PID 0x028E (Xbox 360 Controller)
- Interfaces/Endpoints: single HID interface with one IN interrupt endpoint and one OUT interrupt endpoint for rumble
- Device type id (for API add): `xbox360`

## Client Library Support

The wire protocol is abstracted by client libraries. The **Go client** includes built-in types (`/device/xbox360`), and **generated client libraries** provide equivalent structures with proper packing.  
You don't need to manually construct packets, just use the provided types and send them via the device stream.

See: [Go Client](../clients/go.md), [Generated Client Libraries](../clients/generator.md)

## Adding the device

Use the API to create a bus and add an Xbox 360 controller. Using the raw API (see [API Reference](../api/overview.md) for details):

```bash
# Create a bus
printf "bus/create\0" | nc localhost 3242

# Add xbox360 device with JSON payload
printf 'bus/1/add {"type":"xbox360"}\0' | nc localhost 3242
```

The API returns a Device object with `busId`, `devId`, and other details. Attach it from a USB/IP client, then open a stream to drive input and receive rumble.

Or use one of the [client libraries](../clients/generator.md) which handle the protocol automatically.

## Streaming protocol

The device stream is a bidirectional, raw TCP connection with fixed-size packets.

Direction: client → server (input state)

- 14-byte packets, little-endian layout:
  - Buttons: uint32 (4 bytes)
  - LT, RT: uint8, uint8 (2 bytes)
  - LX, LY, RX, RY: int16 each (8 bytes)

Direction: server → client (rumble feedback)

- 2-byte packets:
  - LeftMotor: uint8, RightMotor: uint8

See `/device/xbox360/inputstate.go` for details.

## Example

A minimal example program that sends input and reads rumble is provided in `examples/`.
