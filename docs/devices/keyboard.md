# HID Keyboard

A full-featured HID keyboard with N-key rollover using a 256-bit key bitmap,
plus LED status feedback (NumLock, CapsLock, ScrollLock).

Use keyboard as the device type when adding a device via the API or client libraries.

## Client Library Support

The wire protocol is abstracted by client libraries.  
The **Go client** includes built-in types (/device/keyboard),
and **generated client libraries** provide equivalent structures
with proper packing.  

You don't need to manually construct packets, just use the provided types
and send them via the device stream.

See: [API Reference](../api/overview.md)

## (RAW) Streaming protocol

The device stream is a bidirectional, raw TCP connection with variable-size packets.

### Input State

- Variable-length packets:
    - Header: Modifiers (1 byte), KeyCount (1 byte)
    - Followed by KeyCount bytes of HID Usage IDs for currently pressed non-modifier keys

### LED Feedback

- 1-byte packets: LEDs bitfield
    - Bit 0: NumLock
    - Bit 1: CapsLock
    - Bit 2: ScrollLock

See `/device/keyboard/inputstate.go` for details.

## Reference

### Modifiers

| Modifier | Hex Value |
| -------- | ----------- |
| LeftCtrl | 0x01 |
| LeftShift | 0x02 |
| LeftAlt | 0x04 |
| LeftGUI | 0x08 |
| RightCtrl | 0x10 |
| RightShift | 0x20 |
| RightAlt | 0x40 |
| RightGUI | 0x80 |

### Keycodes

HID Usage IDs for keys are available in `/device/keyboard/const.go`,
including standard alphanumeric keys (0x04–0x63)
and media keys (Mute, VolumeUp/Down, PlayPause, Stop, Next, Previous).

Helper functions for common operations are in `/device/keyboard/helpers.go`.
