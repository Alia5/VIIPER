# HID Mouse

A standard 5-button mouse with vertical and horizontal scroll wheels. Reports relative motion deltas and supports up to five buttons.

- USB IDs: VID 0x2E8A (Raspberry Pi), PID 0x0011
- Interface/Endpoint: IN 0x81 (mouse input report)
- Device type id (for API add): `mouse`

## Client Library Support

The wire protocol is abstracted by client libraries. The **Go client** includes built-in types (`/device/mouse`), and **generated client libraries** provide equivalent structures with proper packing.  
You don't need to manually construct packets, just use the provided types and send them via the device stream.

See: [Go Client](../clients/go.md), [Generated Client Libraries](../clients/generator.md)

## HID report format (host-facing)

Input (device → host): 9 bytes

- Byte 0: Buttons bitfield (bits 0..4 for buttons 1..5)
- Bytes 1-2: X delta (int16 little-endian, -32768 to +32767)
- Bytes 3-4: Y delta (int16 little-endian, -32768 to +32767)
- Bytes 5-6: Vertical wheel (int16 little-endian; positive up)
- Bytes 7-8: Horizontal wheel/pan (int16 little-endian; positive right)

Deltas are consumed after each IN report so motion is truly relative and not repeated across host polls.

## Device stream protocol (client-facing)

Wire format from your client into VIIPER:

- Fixed 9-byte packets matching the HID report layout:
  [Buttons, dX_lo, dX_hi, dY_lo, dY_hi, Wheel_lo, Wheel_hi, Pan_lo, Pan_hi]

Buttons persist until changed; motion/wheel deltas are applied once and reset.

## Adding the device

Using the raw API (see [API Reference](../api/overview.md) for details):

```bash
# Create a bus
printf "bus/create\0" | nc localhost 3242

# Add mouse device with JSON payload
printf 'bus/1/add {"type":"mouse"}\0' | nc localhost 3242
```

Or use one of the [client libraries](../clients/generator.md) which handle the protocol automatically.

## Examples

A runnable example that periodically moves the mouse a short distance, clicks, and scrolls is provided in `examples/virtual_mouse/`.
