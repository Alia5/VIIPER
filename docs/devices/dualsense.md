# DualSense Controller

The DualSense virtual gamepad emulates a complete PlayStation 5 DualSense
controller connected via USB.
The DualSense Edge variant is also supported and uses the same wire/state model.

It supports sticks, triggers, D-pad, face/shoulder buttons, PS button,
touchpad click, back paddles/function buttons (Edge variant), mic mute button,
IMU (gyro + accelerometer), and touchpad finger coordinates.

=== "TCP API"

    Use `dualsense` as the default device type when adding a device via the API
    or client libraries.

    Use `dualsenseedge` when you want the Edge variant.

    ## Client Library Support

    The wire protocol is abstracted by client libraries.
    The **Go client** includes built-in types (`/device/dualsense`),
    and **generated client libraries** provide equivalent structures
    with proper packing.

    You don't need to manually construct packets, just use the provided types
    and send/receive them via the device control and feedback stream.

    See: [API Reference](../api/overview.md)

    ## (RAW) Streaming protocol

    The device stream is a bidirectional, raw TCP connection with fixed-size
    packets.

    ### Input State

    - 33-byte packets, little-endian layout:
        - Sticks: StickLX, StickLY, StickRX, StickRY: int8 each (4 bytes)
          -128 to 127 per axis (-128=min, 0=center, 127=max)
        - Buttons: uint32 (4 bytes, bitfield)
        - DPad: uint8 (1 byte, bitfield)
        - Triggers: TriggerL2, TriggerR2: uint8, uint8 (2 bytes)
          0-255 (0=not pressed, 255=fully pressed)
        - Touch1: Touch1X, Touch1Y: uint16 each, Touch1Active: bool (5 bytes)
        - Touch2: Touch2X, Touch2Y: uint16 each, Touch2Active: bool (5 bytes)
        - Gyroscope: GyroX, GyroY, GyroZ: int16 each (6 bytes, raw report
          values)
        - Accelerometer: AccelX, AccelY, AccelZ: int16 each
          (6 bytes, raw report values)

    See `/device/dualsense/state.go` for details.

    ### Feedback (Rumble & LED)

    - 6-byte packets:
        - RumbleSmall: uint8, RumbleLarge: uint8 (2 bytes), 0-255 intensity
          values
        - LED Color: LedRed, LedGreen, LedBlue: uint8 each (3 bytes), 0-255 per
          channel
        - PlayerLeds: uint8 (1 byte), host-controlled player indicator LED mask

    See `/device/dualsense/state.go` for the `OutputState` wire definition.

    ## Reference

    ### Button Constants

    | Button | Hex Value |
    | -------- | ----------- |
    | Square button | 0x00000010 |
    | Cross (X) button | 0x00000020 |
    | Circle button | 0x00000040 |
    | Triangle button | 0x00000080 |
    | L1 (Left bumper) | 0x00000100 |
    | R1 (Right bumper) | 0x00000200 |
    | L2 button | 0x00000400 |
    | R2 button | 0x00000800 |
    | Create button | 0x00001000 |
    | Options button | 0x00002000 |
    | L3 (Left stick button) | 0x00004000 |
    | R3 (Right stick button) | 0x00008000 |
    | PS button | 0x00010000 |
    | Touchpad click | 0x00020000 |
    | Mic mute button | 0x00040000 |
    | Edge Variant only | --- |
    | RFn button | 0x00080000 |
    | LFn button | 0x00100000 |
    | R4 back paddle | 0x00200000 |
    | L4 back paddle | 0x00400000 |

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

    VIIPER clamps touch coordinates to the DualSense range:

    - X: **0..1920**
    - Y: **0..1080**

    See `/device/dualsense/const.go`.

    ### IMU (Gyro + Accelerometer)

    VIIPER exposes DualSense IMU values as raw report-space `int16` values,
    while helper conversions use fixed scale factors.

    Constants (see `/device/dualsense/const.go`):

    - `GyroCountsPerDps = 16.384`
    - `AccelCountsPerMS2 = 835.07`

    Gyro (degrees/second):

        raw_gyro = round(gyro_dps * GyroCountsPerDps)
        gyro_dps = raw_gyro / GyroCountsPerDps

    Accelerometer (m/s2):

        raw_accel = round(accel_ms2 * AccelCountsPerMS2)
        accel_ms2 = raw_accel / AccelCountsPerMS2

    On device creation, VIIPER initializes the accelerometer to a controller
    lying flat with gravity downwards (`AccelZ = -8192`, i.e. roughly -1g).

    Helpers are in `/device/dualsense/helpers.go`.

=== "libVIIPER"

    ## API

    | Function | Description |
    | --- | --- |
    | `CreateDualSenseDevice(...)` | Create a virtual DualSense device |
    | `CreateDualSenseEdgeDevice(...)` | Create a virtual DualSense Edge |
    | `SetDualSenseDeviceState(handle, state)` | Push input state |
    | `SetDualSenseOutputCallback(handle, cb)` | Register output callback |
    | `RemoveDualSenseDevice(handle)` | Remove the device |

    ## Input state

    ```c
    typedef struct {
        int8_t   LX;
        int8_t   LY;
        int8_t   RX;
        int8_t   RY;
        uint32_t Buttons;
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
    } DSDeviceState;
    ```

    ## Meta state

    Optional metadata can be provided during `CreateDualSenseDevice`
    and `CreateDualSenseEdgeDevice`.

    ```c
    typedef struct {
        const char* SerialNumber;       // NULL = use default
        const char* MACAddress;         // NULL = use default
        const char* Board;              // NULL = use default
        uint8_t     BatteryStatus;      // 0 = use default
        double      TemperatureCelsius; // 0 = use default
        double      BatteryVoltage;     // 0 = use default
        const char* ShellColor;         // NULL = use default (e.g. "00", "Z1")
    } DSMetaState;
    ```

    ## Output callback

    Called when the host sends rumble or LED commands to the device.

    ```c
    typedef void (*DSOutputCallback)(
        DSDeviceHandle handle,
        uint8_t rumbleSmall,
        uint8_t rumbleLarge,
        uint8_t ledRed,
        uint8_t ledGreen,
        uint8_t ledBlue,
        uint8_t playerLeds
    );
    ```

    Pass `NULL` to `SetDualSenseOutputCallback` to clear
    a previously registered callback.
