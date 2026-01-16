# HID Mouse

A standard 5-button mouse with vertical and horizontal scroll wheels.
Reports relative motion deltas.

Use `mouse` as the device type when adding a device via the API or client libraries.

## Client Library Support

The wire protocol is abstracted by client libraries.  
The **Go client** includes built-in types (/device/mouse),
and **generated client libraries** provide equivalent structures
with proper packing.  

You don't need to manually construct packets, just use the provided types
and send them via the device stream.

See: [API Reference](../api/overview.md)

## (RAW) Streaming protocol

The device stream is a bidirectional, raw TCP connection with fixed-size packets.

### Input State

- 9-byte packets, little-endian layout:
    - Buttons: uint8 (1 byte, bitfield) — bits 0..4 for buttons 1..5
    - X delta: int16 (2 bytes)  
       -32768 to +32767
    - Y delta: int16 (2 bytes)  
       -32768 to +32767
    - Vertical wheel: int16 (2 bytes)  
      positive = up
    - Horizontal wheel/pan: int16 (2 bytes)  
       positive = right

Motion and wheel deltas are consumed after each report and reset;
buttons persist until changed.

See `/device/mouse/inputstate.go` for details.
