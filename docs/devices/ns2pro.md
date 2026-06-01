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
    | Headset | `ButtonHeadset` |

=== "libVIIPER"

    ## API

    | Function | Description |
    | --- | --- |
    | `CreateNS2ProDevice(...)` | Create a virtual Switch 2 Pro Controller |
    | `SetNS2ProDeviceState(handle, state)` | Push input state |
    | `SetNS2ProOutputCallback(handle, cb)` | Register output (rumble/LED) callback |
    | `RemoveNS2ProDevice(handle)` | Remove the device |

    ## Input state

    ```c
    typedef struct {
        uint32_t Buttons;
        uint16_t LX;
        uint16_t LY;
        uint16_t RX;
        uint16_t RY;
        int16_t  AccelX;
        int16_t  AccelY;
        int16_t  AccelZ;
        int16_t  GyroX;
        int16_t  GyroY;
        int16_t  GyroZ;
    } NS2ProDeviceState;
    ```

    Stick values are in the range `0..0x0FFF` (`NS2PRO_STICK_MIN` / `NS2PRO_STICK_CENTER` / `NS2PRO_STICK_MAX`).

    ## Meta state

    Optional metadata passed to `CreateNS2ProDevice`. Controls battery reporting and serial number.

    ```c
    typedef struct {
        const char* SerialNumber;  // NULL = use default
        uint8_t     BatteryLevel;  // 0-9; 0 = use default (9 = full)
        uint8_t     Charging;      // 0 = not charging
        uint8_t     ExternalPower; // 0 = battery only
        uint16_t    BatteryVolts;  // mV; 0 = use default (3800)
    } NS2ProMetaState;
    ```

    ## Output callback

    Called when the host sends HD rumble or player LED commands to the device.

    ```c
    typedef struct {
        uint8_t LeftRumble[16];
        uint8_t RightRumble[16];
        uint8_t Flags;        // bit 0 = rumble update, bit 1 = player LED update
        uint8_t PlayerLedMask;
    } NS2ProOutputState;

    typedef void (*NS2ProOutputCallback)(NS2ProDeviceHandle handle, NS2ProOutputState output);
    ```

    Pass `NULL` to `SetNS2ProOutputCallback` to clear a previously registered callback.

    ## Button constants

    | Constant | Hex Value |
    | --- | --- |
    | `NS2PRO_BUTTON_B` | 0x00000001 |
    | `NS2PRO_BUTTON_A` | 0x00000002 |
    | `NS2PRO_BUTTON_Y` | 0x00000004 |
    | `NS2PRO_BUTTON_X` | 0x00000008 |
    | `NS2PRO_BUTTON_R` | 0x00000010 |
    | `NS2PRO_BUTTON_ZR` | 0x00000020 |
    | `NS2PRO_BUTTON_PLUS` | 0x00000040 |
    | `NS2PRO_BUTTON_RIGHT_STICK` | 0x00000080 |
    | `NS2PRO_BUTTON_DOWN` | 0x00000100 |
    | `NS2PRO_BUTTON_RIGHT` | 0x00000200 |
    | `NS2PRO_BUTTON_LEFT` | 0x00000400 |
    | `NS2PRO_BUTTON_UP` | 0x00000800 |
    | `NS2PRO_BUTTON_L` | 0x00001000 |
    | `NS2PRO_BUTTON_ZL` | 0x00002000 |
    | `NS2PRO_BUTTON_MINUS` | 0x00004000 |
    | `NS2PRO_BUTTON_LEFT_STICK` | 0x00008000 |
    | `NS2PRO_BUTTON_HOME` | 0x00010000 |
    | `NS2PRO_BUTTON_CAPTURE` | 0x00020000 |
    | `NS2PRO_BUTTON_GR` | 0x00040000 |
    | `NS2PRO_BUTTON_GL` | 0x00080000 |
    | `NS2PRO_BUTTON_C` | 0x00100000 |
    | `NS2PRO_BUTTON_HEADSET` | 0x00200000 |

    ## Feature flags

    Passed to `CreateNS2ProDevice` to declare which features the device exposes.

    | Constant | Hex Value | Description |
    | --- | --- | --- |
    | `NS2PRO_FEATURE_BUTTONS` | 0x01 | Button input |
    | `NS2PRO_FEATURE_STICKS` | 0x02 | Analog sticks |
    | `NS2PRO_FEATURE_IMU` | 0x04 | Gyro + accelerometer |
    | `NS2PRO_FEATURE_MOUSE` | 0x10 | Mouse mode |
    | `NS2PRO_FEATURE_RUMBLE` | 0x20 | HD rumble output |
