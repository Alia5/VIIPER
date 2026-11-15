# DualSense Controller (partial)

The DualSense virtual gamepad emulates a _partial_ PlayStation 5 DualSense controller connected via USB.  
The DualSense Edge variant is also supported and uses the same wire/state model.  

It supports sticks, triggers, D-pad, face/shoulder buttons, PS button,
touchpad click, back paddles/function buttons (Edge variant), mic mute button,
IMU (gyro + accelerometer), touchpad finger coordinates and adaptive-triggers.  

The normal `dualsense` and `dualsenseedge` devices only emulate the
"normal controller" portion and **do not** feature the audio devices.  

If you want to get full DualSense / Edge emulation including Audio devices
(and advanced haptics via Audio) see [`dualsense_full` and `dualsenseedge_full`](./dualsense_full.md)

=== "TCP API"

    Use `dualsense` as the default device type when adding a device via the API
    or client libraries.

    Use `dualsenseedge` when you want the Edge variant.

    ## Client Library Support

    The wire protocol is abstracted by client libraries.  
    The **Go client** includes built-in types (`/device/dualsense`),
    and the other **client libraries** provide equivalent structures
    with proper packing.  

    You don't need to manually construct packets, just use the provided types
    and send/receive them via the device control and feedback stream.  

    You can optionally specify metadata of the controller (serial number, shell color, battery, ...)  
    This is done by specifying it as part of the device options.

    For example:

    - `{"type":"dualsense", "deviceSpecific": {"shell_color": "Z1"}}`

    ### Metadata

    | Field                  | Type   | Description                             |
    | ---------------------- | ------ | --------------------------------------- |
    | `serial_number`        | string | Serial number                           |
    | `mac_address`          | string | `"XX:XX:XX:XX:XX:XX"`                   |
    | `board`                | string | Board name                              |
    | `build_time`           | string | RFC 3339 time                           |
    | `battery_status`       | number | Battery status byte                     |
    | `temperature_celsius`  | number | Temperature                             |
    | `battery_voltage`      | number | Battery voltage                         |
    | `shell_color`          | string | Color code, for example `"00"`, `"Z1"`  |
    | `headphones_connected` | bool   | Headphones are in the audio jack        |
    | `microphone_connected` | bool   | Headset microphone is in the audio jack |

    See: [API Reference](../api/overview.md)

    ## (RAW) Streaming protocol

    The device stream is a bidirectional, raw TCP connection with fixed-size packets.

    ### Input State

    - 33-byte packets, little-endian layout:
        - Sticks: `StickLX`, `StickLY`, `StickRX`, `StickRY`: `int8` each (4 bytes)  
          -128 to 127 per axis (-128=min, 0=center, 127=max)
        - Buttons: `uint32` (4 bytes, bitfield)
        - DPad: `uint8` ( bitfield)
        - Triggers: `TriggerL2`, `TriggerR2`: `uint8`, `uint8` (2 bytes)
          0-255 (0=not pressed, 255=fully pressed)
        - Touch1: `Touch1X`, `Touch1Y`: `uint16` each, `Touch1Active`: bool (5 bytes)
        - Touch2: `Touch2X`, `Touch2Y`: `uint16` each, `Touch2Active`: bool (5 bytes)
        - Gyroscope: `GyroX`, `GyroY`, `GyroZ`: `int16` each (6 bytes, raw report values)
        - Accelerometer: `AccelX`, `AccelY`, `AccelZ`: `int16` each (6 bytes, raw report values)

    See `/device/dualsense/state.go` for details.

    ### Feedback (Rumble, LED & Adaptive Triggers)

    - 31-byte packets:
        - UpdateFlags1: `uint8` (raw flag byte 1 of the output report)
            - `UpdateFlag1Rumble`: 0x01
            - `UpdateFlag1HapticsSelect`: 0x02
            - `UpdateFlag1TriggerR2`: 0x04
            - `UpdateFlag1TriggerL2`: 0x08
        - UpdateFlags2: `uint8` (raw flag byte 2 of the output report)
            - `UpdateFlag2Led`: 0x04
            - `UpdateFlag2PlayerLeds`: 0x10
        - UpdateFlags3: `uint8` (raw flag byte 39 of the output report)
            - `UpdateFlag3ImprovedRumble`: 0x04
        - RumbleSmall: `uint8`, RumbleLarge: `uint8` (2 bytes), 0-255 intensity values
        - LED Color: Red, Green, Blue: `uint8` each (3 bytes), 0-255 per channel
        - PlayerLeds: `uint8`, host-controlled player indicator LED mask
        - TriggerR2Mode: `uint8`, effect mode of the R2 trigger
        - TriggerR2Params: `uint8` * 10, raw effect parameters
        - TriggerL2Mode: `uint8`, effect mode of the L2 trigger
        - TriggerL2Params: `uint8` * 10, raw effect parameters

    The data fields are always copied from the output report.
    Use the flags to find which fields the host set:

    - Rumble is active when `UpdateFlags1` has `UpdateFlag1Rumble` or `UpdateFlag1HapticsSelect`.  
      When the two bits are clear, rumble is off.
    - With `UpdateFlag3ImprovedRumble` the rumble values are full scale.  
      Without it, hosts usually send half-scale values.
    - The LED color is valid when `UpdateFlags2` has `UpdateFlag2Led`.
    - `PlayerLeds` is valid when `UpdateFlags2` has `UpdateFlag2PlayerLeds`.
    - The R2 trigger fields are valid when `UpdateFlags1` has `UpdateFlag1TriggerR2`.
    - The L2 trigger fields are valid when `UpdateFlags1` has `UpdateFlag1TriggerL2`.

    The flag bytes are the raw bytes of the output report, so they can have more bits than the ones listed here.  
    The audio flags are described in [`dualsense_full`](./dualsense_full.md)

    VIIPER sets the adaptive trigger status in the USB input report of the virtual device.  
    It uses the active trigger effect and the `TriggerL2` / `TriggerR2` values of the input state.  
    The client does not send it.

    See `/device/dualsense/state.go` for the `OutputState` wire definition.

    ### Trigger effects

    The Go client decodes the raw parameters with `OutputState.TriggerR2()` and `OutputState.TriggerL2()`  
    They return a `TriggerEffect` (see `/device/dualsense/trigger_effect.go`; The other client libraries have the same decoder)  
    

    | Mode                         | Hex Value |
    | ---------------------------- | --------- |
    | `TriggerModeSimpleFeedback`  | 0x01      |
    | `TriggerModeSimpleWeapon`    | 0x02      |
    | `TriggerModeOff`             | 0x05      |
    | `TriggerModeSimpleVibration` | 0x06      |
    | `TriggerModeLimitedFeedback` | 0x11      |
    | `TriggerModeLimitedWeapon`   | 0x12      |
    | `TriggerModeFeedback`        | 0x21      |
    | `TriggerModeBow`             | 0x22      |
    | `TriggerModeGalloping`       | 0x23      |
    | `TriggerModeWeapon`          | 0x25      |
    | `TriggerModeVibration`       | 0x26      |
    | `TriggerModeMachine`         | 0x27      |
    | `TriggerModeCalibration`     | 0xFC      |
    | `TriggerModeDebugFD`         | 0xFD      |
    | `TriggerModeDebugFE`         | 0xFE      |

    Decoded `TriggerEffect` fields for each mode:

    - Off: none
    - Feedback: `ZoneStrengths` (10 zones, 0 = zone not active, 1-8)
    - Weapon: `StartZone`, `EndZone`, `Strength` (1-8)
    - Vibration: `ZoneStrengths` (amplitude, 0 or 1-8), `Frequency` (Hz)
    - Bow: `StartZone`, `EndZone`, `Strength` (1-8), `SnapForce` (1-8)
    - Galloping: `StartZone`, `EndZone`, `FirstFoot`, `SecondFoot`, `Frequency` (Hz)
    - Machine: `StartZone`, `EndZone`, `AmplitudeA`, `AmplitudeB`, `Frequency` (Hz), `Period`
    - Simple Feedback: `StartPosition`, `Strength` (raw values)
    - Simple Weapon: `StartPosition`, `EndPosition`, `Strength` (raw values)
    - Simple Vibration: `StartPosition`, `Strength` (amplitude), `Frequency` (Hz)
    - Limited Feedback: `StartPosition`, `Strength` (0-10)
    - Limited Weapon: `StartPosition`, `EndPosition`, `Strength` (0-10)
    - Calibration, DebugFD, DebugFE: not decoded

    Raw parameter layout (parameter byte index 0-9):

    - Feedback, Vibration:  
      Bytes 0-1 are a little-endian bitmask of the active zones.  
      Bytes 2-5 hold one 3-bit value (strength - 1) for each zone.  
      Vibration has the frequency in byte 8.
    - Weapon, Bow, Galloping, Machine:  
      Bytes 0-1 are a bitmask with the start zone bit and the end zone bit.  
      Byte 2 holds two 3-bit values:
        - Weapon: strength - 1
        - Bow: strength - 1, snap force - 1
        - Galloping: second foot, first foot
        - Machine: amplitude A, amplitude B

      Galloping and Machine have the frequency in byte 3.  
      Machine has the period in byte 4.
    - Simple and Limited Feedback:
        - Byte 0: start position
        - Byte 1: strength
    - Simple and Limited Weapon:
        - Byte 0: start position
        - Byte 1: end position
        - Byte 2: strength
    - Simple Vibration:
        - Byte 0: frequency
        - Byte 1: amplitude
        - Byte 2: start position

    ## Reference

    ### Button Constants

    | Button                  | Hex Value  |
    | ----------------------- | ---------- |
    | Square button           | 0x00000010 |
    | Cross (X) button        | 0x00000020 |
    | Circle button           | 0x00000040 |
    | Triangle button         | 0x00000080 |
    | L1 (Left bumper)        | 0x00000100 |
    | R1 (Right bumper)       | 0x00000200 |
    | L2 button               | 0x00000400 |
    | R2 button               | 0x00000800 |
    | Create button           | 0x00001000 |
    | Options button          | 0x00002000 |
    | L3 (Left stick button)  | 0x00004000 |
    | R3 (Right stick button) | 0x00008000 |
    | PS button               | 0x00010000 |
    | Touchpad click          | 0x00020000 |
    | Mic mute button         | 0x00040000 |
    | Edge Variant only       | ---        |
    | RFn button              | 0x00200000 |
    | LFn button              | 0x00100000 |
    | R4 back paddle          | 0x00800000 |
    | L4 back paddle          | 0x00400000 |

    ### D-Pad Constants

    | D-Pad Direction | Hex Value |
    | --------------- | --------- |
    | Up              | 0x01      |
    | Down            | 0x02      |
    | Left            | 0x04      |
    | Right           | 0x08      |

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

    | Function                                 | Description                                       |
    | ---------------------------------------- | ------------------------------------------------- |
    | `CreateDualSense(...)`                   | Create a virtual DualSense device                 |
    | `CreateDualSenseEdge(...)`               | Create a virtual DualSense Edge                   |
    | `SetDualSenseDeviceState(handle, state)` | Push input state                                  |
    | `SetDualSenseOutputCallback(handle, cb)` | Register output callback                          |
    | `RemoveDualSense(handle)`                | Remove the device                                 |
    | `DualSenseOutputReport(state, report)`   | Build the USB output report from an output state  |
    | `DualSenseGyroDpsToRaw(dps)`             | Convert degrees/second to the raw gyroscope value |
    | `DualSenseAccelMS2ToRaw(ms2)`            | Convert m/s² to the raw accelerometer value       |

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

    Optional metadata can be provided during `CreateDualSense` and `CreateDualSenseEdge`.

    ```c
    typedef struct {
        const char* SerialNumber;        // NULL = use default
        const char* MACAddress;          // NULL = use default
        const char* Board;               // NULL = use default
        uint8_t     BatteryStatus;       // 0 = use default
        double      TemperatureCelsius;  // 0 = use default
        double      BatteryVoltage;      // 0 = use default
        const char* ShellColor;          // NULL = use default (e.g. "00", "Z1")
        uint8_t     HeadphonesConnected; // 0 = no headphones in the audio jack
        uint8_t     MicrophoneConnected; // 0 = no headset microphone in the audio jack
    } DSMetaState;
    ```

    ## Trigger effect

    ```c
    typedef struct {
        uint8_t Mode;          // DS_TRIGGER_MODE_*
        uint8_t Params[DS_TRIGGER_PARAMS_SIZE];
        uint8_t ZoneStrengths[DS_TRIGGER_ZONE_COUNT];
        uint8_t StartZone;     // 0-9
        uint8_t EndZone;       // 0-9
        uint8_t StartPosition; // raw position
        uint8_t EndPosition;   // raw position
        uint8_t Strength;
        uint8_t SnapForce;     // 1-8
        uint8_t AmplitudeA;    // 0-7
        uint8_t AmplitudeB;    // 0-7
        uint8_t FirstFoot;     // 0-6
        uint8_t SecondFoot;    // 1-7
        uint8_t Frequency;     // Hz
        uint8_t Period;        // tenths of a second
    } DSTriggerEffect;
    ```

    libVIIPER decodes the effect parameters.  
    `Params` holds the raw bytes for a client that forwards the effect to a physical controller.  
    The modes and the fields that each mode uses are the same as in the "Trigger effects" section of the TCP API.

    ## Output callback

    Called when the host sends rumble, LED or adaptive trigger commands to the device.

    ```c
    typedef struct {
        uint8_t         UpdateFlags1; // DS_OUTPUT_UPDATE1_*
        uint8_t         UpdateFlags2; // DS_OUTPUT_UPDATE2_*
        uint8_t         UpdateFlags3; // DS_OUTPUT_UPDATE3_*
        uint8_t         RumbleSmall;
        uint8_t         RumbleLarge;
        uint8_t         LedRed;
        uint8_t         LedGreen;
        uint8_t         LedBlue;
        uint8_t         PlayerLeds;
        DSTriggerEffect TriggerR2;
        DSTriggerEffect TriggerL2;
    } DSOutputState;

    typedef void (*DSOutputCallback)(DSDeviceHandle handle, DSOutputState state);
    ```

    The fields are always copied from the output report.
    Use the flags to find which fields the host set:

    - Rumble is active when `UpdateFlags1` has `DS_OUTPUT_UPDATE1_RUMBLE` or `DS_OUTPUT_UPDATE1_HAPTICS_SELECT`.
    - With `DS_OUTPUT_UPDATE3_IMPROVED_RUMBLE` the rumble values are full scale.
    - The LED color is valid when `UpdateFlags2` has `DS_OUTPUT_UPDATE2_LED`.
    - `PlayerLeds` is valid when `UpdateFlags2` has `DS_OUTPUT_UPDATE2_PLAYER_LEDS`.
    - `TriggerR2` is valid when `UpdateFlags1` has `DS_OUTPUT_UPDATE1_TRIGGER_R2`.
    - `TriggerL2` is valid when `UpdateFlags1` has `DS_OUTPUT_UPDATE1_TRIGGER_L2`.

    Pass `NULL` to `SetDualSenseOutputCallback` to clear a previously registered callback.

    ## Forwarding feedback to a physical controller

    `DualSenseOutputReport(state, report)` builds the USB output report (report ID `0x02`) of a physical DualSense from an output state.
    `report` is a buffer of `DS_OUTPUT_REPORT_SIZE` bytes.
    The `DS_OUTPUT_REPORT_OFFSET_*` defines give the position of each field in that report.
