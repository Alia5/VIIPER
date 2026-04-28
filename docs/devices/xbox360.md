# Xbox 360 Controller

The Xbox 360 virtual gamepad emulates an XInput-compatible controller that most
operating systems and games understand out of the box.

=== "TCP API"

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

    - 20-byte packets, little-endian layout:
        - Buttons: uint32 (4 bytes, bitfield)
        - Triggers: LT, RT: uint8, uint8 (2 bytes)  
          0-255 (0=not pressed, 255=fully pressed)
        - Sticks: LX, LY, RX, RY: int16 each (8 bytes)  
          0 is center, -32768 is min, 32767 is max
        - Reserved: there are 6 reserved bytes at the end of the report. For most subtypes, these will be zeroed, but a few subtypes do put data here.

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

=== "libVIIPER"

    ## API

    | Function | Description |
    | --- | --- |
    | `CreateXbox360Device(serverHandle, &handle, busID, autoAttach, vid, pid, subType)` | Create a virtual Xbox 360 controller |
    | `SetXbox360DeviceState(handle, state)` | Push an input state to the device |
    | `SetXbox360RumbleCallback(handle, cb)` | Register a callback for rumble output |
    | `RemoveXbox360Device(handle)` | Remove the device |

    ## Input state

    ```c
    typedef struct {
        uint32_t Buttons;
        uint8_t  LT;
        uint8_t  RT;
        int16_t  LX;
        int16_t  LY;
        int16_t  RX;
        int16_t  RY;
        uint8_t  Reserved[6];
    } Xbox360DeviceState;
    ```

    ### Button flags

    | Constant | Value |
    | --- | --- |
    | `XBOX360_BUTTON_DPAD_UP` | `0x0001` |
    | `XBOX360_BUTTON_DPAD_DOWN` | `0x0002` |
    | `XBOX360_BUTTON_DPAD_LEFT` | `0x0004` |
    | `XBOX360_BUTTON_DPAD_RIGHT` | `0x0008` |
    | `XBOX360_BUTTON_START` | `0x0010` |
    | `XBOX360_BUTTON_BACK` | `0x0020` |
    | `XBOX360_BUTTON_LEFT_THUMB` | `0x0040` |
    | `XBOX360_BUTTON_RIGHT_THUMB` | `0x0080` |
    | `XBOX360_BUTTON_LEFT_SHOULDER` | `0x0100` |
    | `XBOX360_BUTTON_RIGHT_SHOULDER` | `0x0200` |
    | `XBOX360_BUTTON_GUIDE` | `0x0400` |
    | `XBOX360_BUTTON_A` | `0x1000` |
    | `XBOX360_BUTTON_B` | `0x2000` |
    | `XBOX360_BUTTON_X` | `0x4000` |
    | `XBOX360_BUTTON_Y` | `0x8000` |

    ## Rumble callback

    ```c
    typedef void (*Xbox360RumbleCallback)(Xbox360DeviceHandle handle, uint8_t leftMotor, uint8_t rightMotor);
    ```

    Pass `NULL` to `SetXbox360RumbleCallback` to clear a previously registered callback.
