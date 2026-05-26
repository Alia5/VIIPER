# Switch 2 Pro Controller

The `ns2pro` virtual gamepad emulates a Nintendo Switch 2 Pro Controller over USB.
It exposes the Switch 2 HID reports used by SDL, including buttons, sticks,
gyro/accelerometer data, and HD rumble output.

=== "TCP API"

    Use `ns2pro` as the device type when adding a device via the API or client libraries.

    ## Client Library Support

    The Go client can use the built-in types from `/device/ns2pro`.
    Generated client libraries will pick up the `viiper:wire` tags from this package
    the next time codegen is run.

    ## Raw Streaming Protocol

    The device stream is a bidirectional raw TCP connection with fixed-size packets.

    ### Input State

    - 27-byte packets, little-endian layout:
        - Buttons: `uint32` bitfield
        - Sticks: `LX`, `LY`, `RX`, `RY` as raw `uint16` values, clamped to `0..4095`
        - Accelerometer: `AccelX`, `AccelY`, `AccelZ` as raw `int16` report values
        - Gyroscope: `GyroX`, `GyroY`, `GyroZ` as raw `int16` report values
        - Battery: `BatteryLevel` (`0..9`), `Charging`, `ExternalPower`

    ### Feedback

    - 34-byte packets:
        - `LeftRumble`: 16 bytes copied from HID output report `0x02`
        - `RightRumble`: 16 bytes copied from HID output report `0x02`
        - `Flags`: bit 0 = rumble update, bit 1 = player LED update
        - `PlayerLedMask`: SDL/Steam player LED mask from bulk command `0x09/0x07`

    ## Notes

    VIIPER implements the HID and vendor bulk command paths needed by SDL's Switch 2
    driver. The USB identity mirrors a wired Switch 2 Pro Controller closely enough
    for host-side drivers to find the HID interface and vendor bulk interface:
    product string `Switch 2 Pro Controller`, serial `00`, `bcdDevice=0x0200`,
    HID plus vendor bulk interfaces, and Microsoft OS 1.0 compatible ID and
    extended properties descriptors that bind the vendor bulk interface to WinUSB
    on Windows.

    NFC, Bluetooth GATT, and headset audio streaming are not emulated.

    Gyro and accelerometer values are raw report values. Clients that need physical
    units should convert them according to their target host or driver conventions.

    ## Button Constants

    | Button | Constant |
    | --- | --- |
    | B / A / Y / X | `ButtonB`, `ButtonA`, `ButtonY`, `ButtonX` |
    | L / R / ZL / ZR | `ButtonL`, `ButtonR`, `ButtonZL`, `ButtonZR` |
    | Plus / Minus | `ButtonPlus`, `ButtonMinus` |
    | Stick clicks | `ButtonLeftStick`, `ButtonRightStick` |
    | D-pad | `ButtonUp`, `ButtonDown`, `ButtonLeft`, `ButtonRight` |
    | System buttons | `ButtonHome`, `ButtonCapture`, `ButtonC` |
    | Grip buttons | `ButtonGL`, `ButtonGR` |
