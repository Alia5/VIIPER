# Xbox 360 Controller

The Xbox 360 virtual gamepad emulates an XInput-compatible controller that most
operating systems and games understand out of the box.

Use `xbox360` as the device type when adding a device via the API or client libraries.

## Client Library Support

The wire protocol is abstracted by client libraries.  
The **Go client** includes built-in types (`/device/xbox360`),
and **generated client libraries** provide equivalent structures
with proper packing.

You don't need to manually construct packets, just use the provided types
and send/receive them via the device control and feedback stream.

You can optionally specify a sub type if you wish to emulate a different type of controller.
This is done by specifying it as part of the device options.

For example:

- `{"type":"xbox360", "deviceSpecific": {"subType": 7}}`

### Subtypes

| Subtype                                   | Value |
| ----------------------------------------- | ----- |
| Gamepad                                   | 1     |
| Wheel                                     | 2     |
| Arcade Stick                              | 3     |
| Flight Stick                              | 4     |
| Dance Pad                                 | 5     |
| Guitar                                    | 6     |
| Guitar Alternate                          | 7     |
| Drums                                     | 8     |
| Rock Band Stage Kit                       | 9     |
| Guitar Bass                               | 11    |
| Rock Band Pro Keys                        | 15    |
| Arcade Pad                                | 19    |
| Turntable                                 | 23    |
| Rock Band Pro Guitar                      | 25    |
| Disney Infinity or Lego Dimensions Portal | 33    |
| Skylanders Portal                         | 36    |

See: [API Reference](../api/overview.md)

## (RAW) Streaming protocol

The device stream is a bidirectional, raw TCP connection with fixed-size packets.

### Input State

- 14-byte packets, little-endian layout:
  - Buttons: uint32 (4 bytes, bitfield)
  - Triggers: LT, RT: uint8, uint8 (2 bytes)  
    0-255 (0=not pressed, 255=fully pressed)
  - Sticks: LX, LY, RX, RY: int16 each (8 bytes)  
    0 is center, -32768 is min, 32767 is max

### Rumble Feedback

- 2-byte packets:
  - LeftMotor: uint8, RightMotor: uint8  
    0-255 intensity values

See `/device/xbox360/inputstate.go` for details.

### Button constants

| Button             | Hex Value |
| ------------------ | --------- |
| D-Pad Up           | 0x0001    |
| D-Pad Down         | 0x0002    |
| D-Pad Left         | 0x0004    |
| D-Pad Right        | 0x0008    |
| Start button       | 0x0010    |
| Back button        | 0x0020    |
| Left stick button  | 0x0040    |
| Right stick button | 0x0080    |
| Left bumper        | 0x0100    |
| Right bumper       | 0x0200    |
| Xbox/Guide button  | 0x0400    |
| A button           | 0x1000    |
| B button           | 0x2000    |
| X button           | 0x4000    |
| Y button           | 0x8000    |
