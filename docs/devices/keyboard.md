# HID Keyboard

A full-featured HID keyboard with N-key rollover using a 256-bit key bitmap, plus LED status feedback (NumLock, CapsLock, ScrollLock) via an OUT report.

- USB IDs: VID 0x2E8A (Raspberry Pi), PID 0x0010
- Interfaces/Endpoints:
  - IN: 0x81 (keyboard input report)
  - OUT: 0x01 (LED output report)
- Device type id (for API add): `keyboard`

## Client SDK Support

The wire protocol is abstracted by client SDKs. The **Go client** includes built-in types (`pkg/device/keyboard`), and **generated SDKs** provide equivalent structures with proper packing.  
You don't need to manually construct packets, just use the provided types and send them via the device stream.

See: [Go Client](../clients/go.md), [Generated SDKs](../clients/generator.md)

## HID report format (host-facing)

Input (device → host): 34 bytes

- Byte 0: Modifiers bitfield (LeftCtrl, LeftShift, LeftAlt, LeftGUI, RightCtrl, RightShift, RightAlt, RightGUI)
- Byte 1: Reserved (0)
- Bytes 2..33: 256-bit key bitmap (least-significant bit = usage ID 0)

Output (host → device): 1 byte LEDs

- Bit 0 NumLock, Bit 1 CapsLock, Bit 2 ScrollLock (remaining bits reserved)

Note: The HID descriptor uses a long-item Report Count (0x96) to encode 256 for the bitmap.

## Device stream protocol (client-facing)

Wire format from your client into VIIPER:

- Variable-length packets
- Header: [Modifiers (1 byte), KeyCount (1 byte)]
- Followed by KeyCount bytes of HID Usage IDs for the currently pressed non-modifier keys

VIIPER converts this to the bitmap report for the host, so you don’t need to manage the 256-bit array yourself.

Example wire packet to press “A” with LeftShift:

- Modifiers = 0x02 (LeftShift)
- Count = 1
- Keys = [0x04]  // HID usage for “A”

## LEDs feedback

The device sends the current LED state (1 byte) back on the same stream whenever the host changes it. You can use this to update indicators in your client.

## Helpers and keycodes

Convenience helpers and key constants are available in the Go package:

- `pkg/device/keyboard/helpers.go`: TypeString, TypeChar, PressKey, Release, etc.
- `pkg/device/keyboard/const.go`: Modifiers, LED bits, and HID usage IDs, including media keys (Mute, VolumeUp/Down, PlayPause, Stop, Next, Previous)

## Adding the device

Using the raw API (see [API Reference](../api/overview.md) for details):

```bash
# Create a bus
printf "bus/create\n\n" | nc localhost 3242

# Add keyboard device with JSON payload
printf 'bus/1/add {"type":"keyboard"}\n\n' | nc localhost 3242
```

Or use one of the [client SDKs](../clients/generator.md) which handle the protocol automatically.

## Examples

A runnable example that types “Hello!” followed by Enter every few seconds is provided in `examples/virtual_keyboard/`.
