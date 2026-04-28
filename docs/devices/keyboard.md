# HID Keyboard

A full-featured HID keyboard with N-key rollover using a 256-bit key bitmap,
plus LED status feedback (NumLock, CapsLock, ScrollLock).

=== "TCP API"

    Use `keyboard` as the device type when adding a device via the API or client libraries.

    ## Client Library Support

    The wire protocol is abstracted by client libraries.  
    The **Go client** includes built-in types (`/device/keyboard`),
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
        - Followed by KeyCount bytes of HID Usage IDs for pressed non-modifier keys

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

    Helper functions are in `/device/keyboard/helpers.go`.

=== "libVIIPER"

    ## API

    | Function | Description |
    | --- | --- |
    | `CreateKeyboardDevice(serverHandle, &handle, busID, autoAttach, vid, pid)` | Create a virtual HID keyboard |
    | `SetKeyboardDeviceState(handle, state)` | Push an input state to the device |
    | `SetKeyboardLEDCallback(handle, cb)` | Register a callback for LED state changes |
    | `RemoveKeyboardDevice(handle)` | Remove the device |

    ## Input state

    ```c
    typedef struct {
        uint8_t Modifiers;
        uint8_t KeyBitmap[32]; /* 256-bit bitmap, one bit per HID key code */
    } KeyboardDeviceState;
    ```

    ### Modifier flags

    | Constant | Value | Key |
    | --- | --- | --- |
    | `KB_MOD_LEFT_CTRL` | `0x01` | Left Control |
    | `KB_MOD_LEFT_SHIFT` | `0x02` | Left Shift |
    | `KB_MOD_LEFT_ALT` | `0x04` | Left Alt |
    | `KB_MOD_LEFT_GUI` | `0x08` | Left GUI (Win/Cmd) |
    | `KB_MOD_RIGHT_CTRL` | `0x10` | Right Control |
    | `KB_MOD_RIGHT_SHIFT` | `0x20` | Right Shift |
    | `KB_MOD_RIGHT_ALT` | `0x40` | Right Alt |
    | `KB_MOD_RIGHT_GUI` | `0x80` | Right GUI (Win/Cmd) |

    Key codes in `KeyBitmap` follow the [USB HID Usage Tables](https://usb.org/sites/default/files/hut1_5.pdf) (page 83, Keyboard/Keypad page).

    ## LED callback

    Called when the host changes keyboard LED state.

    ```c
    typedef void (*KeyboardLEDCallback)(KeyboardDeviceHandle handle, uint8_t leds);
    ```

    ### LED flags

    | Constant | Value |
    | --- | --- |
    | `KB_LED_NUM_LOCK` | `0x01` |
    | `KB_LED_CAPS_LOCK` | `0x02` |
    | `KB_LED_SCROLL_LOCK` | `0x04` |
    | `KB_LED_COMPOSE` | `0x08` |
    | `KB_LED_KANA` | `0x10` |

    Pass `NULL` to `SetKeyboardLEDCallback` to clear a previously registered callback.
