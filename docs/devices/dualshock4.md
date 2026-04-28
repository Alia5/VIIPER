# DualShock 4 Controller

The DualShock 4 virtual gamepad emulates a complete PlayStation 4 Controller (V1)
connected via USB.
It supports sticks, triggers, D-pad, face/shoulder buttons, PS button,
touchpad click, IMU (gyro + accelerometer), and touchpad finger coordinates.

=== "TCP API"

    Use `dualshock4` as the device type when adding a device via the API or client libraries.

    ## Client Library Support

    The wire protocol is abstracted by client libraries.
    The **Go client** includes built-in types (`/device/dualshock4`),
    and **generated client libraries** provide equivalent structures
    with proper packing.

    You don't need to manually construct packets, just use the provided types
    and send/receive them via the device control and feedback stream.

    See: [API Reference](../api/overview.md)

    ## (RAW) Streaming protocol

    The device stream is a bidirectional, raw TCP connection with fixed-size packets.

    ### Input State

    - 31-byte packets, little-endian layout:
        - Sticks: StickLX, StickLY, StickRX, StickRY: int8 each (4 bytes)
          -128 to 127 per axis (-128=min, 0=center, 127=max)
        - Buttons: uint16 (2 bytes, bitfield)
        - DPad: uint8 (1 byte, bitfield)
        - Triggers: TriggerL2, TriggerR2: uint8, uint8 (2 bytes)
          0-255 (0=not pressed, 255=fully pressed)
        - Touch1: Touch1X, Touch1Y: uint16 each, Touch1Active: bool (5 bytes)
        - Touch2: Touch2X, Touch2Y: uint16 each, Touch2Active: bool (5 bytes)
        - Gyroscope: GyroX, GyroY, GyroZ: int16 each (6 bytes, fixed-point deg/s)
        - Accelerometer: AccelX, AccelY, AccelZ: int16 each (6 bytes, fixed-point m/s2)

    See `/device/dualshock4/inputstate.go` for details.

    ### Feedback (Rumble & LED)

    - 7-byte packets:
        - RumbleSmall: uint8, RumbleLarge: uint8 (2 bytes), 0-255 intensity values
        - LED Color: LedRed, LedGreen, LedBlue: uint8 each (3 bytes), 0-255 per channel
        - LED Flash: FlashOn, FlashOff: uint8 each (2 bytes), units of 2.5ms per value

    See `/device/dualshock4/inputstate.go` for the `OutputState` wire definition.

    ## Reference

    ### Button Constants

    | Button | Hex Value |
    | -------- | ----------- |
    | Square button | 0x0010 |
    | Cross (X) button | 0x0020 |
    | Circle button | 0x0040 |
    | Triangle button | 0x0080 |
    | L1 (Left bumper) | 0x0100 |
    | R1 (Right bumper) | 0x0200 |
    | L2 button | 0x0400 |
    | R2 button | 0x0800 |
    | Share button | 0x1000 |
    | Options button | 0x2000 |
    | L3 (Left stick button) | 0x4000 |
    | R3 (Right stick button) | 0x8000 |
    | PS button | 0x0001 |
    | Touchpad click | 0x0002 |

    ### D-Pad Constants

    | D-Pad Direction | Hex Value |
    | --------------- | ----------- |
    | Up | 0x01 |
    | Down | 0x02 |
    | Left | 0x04 |
    | Right | 0x08 |

    ### Touchpad Coordinates

    Touch coordinates are sent as `Touch{1,2}X: uint16` and `Touch{1,2}Y: uint16`
    plus an explicit boolean `Touch{1,2}Active`.

    VIIPER clamps touch coordinates to the DS4 range:

    - X: **0..1920**
    - Y: **0..942**

    See `/device/dualshock4/const.go`.

    ### IMU (Gyro + Accelerometer)

    VIIPER uses fixed-point physical units for IMU values on the wire (stored as `int16`)
    to avoid float serialization differences across client languages.

    Constants (see `/device/dualshock4/const.go`):

    - `GyroCountsPerDps = 16`
    - `AccelCountsPerMS2 = 512`

    Gyro (degrees/second):

        raw_gyro = round(gyro_dps * GyroCountsPerDps)
        gyro_dps = raw_gyro / GyroCountsPerDps

    Accelerometer (m/s2):

        raw_accel = round(accel_ms2 * AccelCountsPerMS2)
        accel_ms2 = raw_accel / AccelCountsPerMS2

    With the default scales:

    - Gyro: resolution `0.0625 deg/s`, max approx `2048 deg/s`
    - Accelerometer: resolution approx `0.00195 m/s2`, max approx `64 m/s2`

    On device creation, VIIPER initializes the accelerometer to a controller lying
    flat with gravity downwards (`AccelZ = -5023`, i.e. `round(-9.81 * 512)`).

    Helpers are in `/device/dualshock4/helpers.go`.

=== "libVIIPER"

    ## API

    | Function | Description |
    | --- | --- |
    | `CreateDS4Device(serverHandle, &handle, busID, autoAttach, vid, pid)` | Create a virtual DualShock 4 controller |
    | `SetDS4DeviceState(handle, state)` | Push an input state to the device |
    | `SetDS4OutputCallback(handle, cb)` | Register a callback for rumble and LED output |
    | `RemoveDS4Device(handle)` | Remove the device |

    ## Input state

    ```c
    typedef struct {
        int8_t   LX;
        int8_t   LY;
        int8_t   RX;
        int8_t   RY;
        uint16_t Buttons;
        uint8_t  DPad;
        uint8_t  L2;
        uint8_t  R2;
        uint16_t Touch1X;
        uint16_t Touch1Y;
        uint8_t  Touch1Active;
        uint16_t Touch2X;
        uint16_t Touch2Y;
        uint8_t  Touch2Active;
        int16_t  GyroX;
        int16_t  GyroY;
        int16_t  GyroZ;
        int16_t  AccelX;
        int16_t  AccelY;
        int16_t  AccelZ;
    } DS4DeviceState;
    ```

    ## Output callback

    Called when the host sends rumble or LED commands to the device.

    ```c
    typedef void (*DS4OutputCallback)(
        DS4DeviceHandle handle,
        uint8_t rumbleSmall,
        uint8_t rumbleLarge,
        uint8_t ledRed,
        uint8_t ledGreen,
        uint8_t ledBlue,
        uint8_t flashOn,
        uint8_t flashOff
    );
    ```

    Pass `NULL` to `SetDS4OutputCallback` to clear a previously registered callback.
