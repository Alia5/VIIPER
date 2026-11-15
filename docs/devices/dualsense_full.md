# DualSense Controller (full w/ audio)

The DualSense virtual gamepad emulates a _full_ PlayStation 5 DualSense controller connected via USB.  
The DualSense Edge variant is also supported and uses the same wire/state model.

It supports sticks, triggers, D-pad, face/shoulder buttons, PS button, touchpad click,
back paddles/function buttons (Edge variant), mic mute button, IMU (gyro + accelerometer),
touchpad finger coordinates, adaptive triggers and audio.  

If you don't need the full DualSense featureset, but want basic full DualSense / Edge emulation,
omitting audio devices (and thus advanced haptics) see [`dualsense` and `dualsenseedge`](./dualsense.md)

The DualSenses advanced haptics are audio-based and use channels 2 and 3 of the audio output device.
Channels 0 and 1 are for regular audio.  

To use audio / audio-based haptics (with the TCP API), you must specify an audio port when you create the device.  
VIIPER then listens on this TCP port and sends 4 channels (channels 0,1 speaker; channels 2,3 haptics) to your application.
Your application can send 2 microphone channels back via the same socket.  
The Streams are all raw PCM (48 kHz, signed 16-bit little-endian), again 4 channels for output, 2 for mic.  

The usage of TCP and RAW PCM makes it super easy to use with SDLs audio APIs or tools like FFmpeg.  

??? example "FFmpeg example"

    ```sh
    # Play the audio channels
    ffplay -nodisp -f s16le -sample_rate 48000 -ch_layout 4.0 -af "pan=stereo|c0=c0|c1=c1" tcp://localhost:5000

    # Play the haptics channels
    ffplay -nodisp -f s16le -sample_rate 48000 -ch_layout 4.0 -af "pan=stereo|c0=c2|c1=c3" tcp://localhost:5000

    # Send a 440 Hz tone as microphone
    ffmpeg -re -f lavfi -i sine=frequency=440 -ar 48000 -ac 2 -f s16le tcp://localhost:5000

    # Play a file on the haptics channels of the virtual controller
    # You must set your default audio device to the virtual controller for it to work
    ffplay -nodisp -af "pan=quad|c2=c0|c3=c1" music.mp3

    ```

Audio with libVIIPER uses callbacks instead, using the same easy to use, raw PCM format.  
Specifying an audio port when using libVIIPER does nothing(!)

=== "TCP API"

    Use `dualsense_full` as the device type when adding a device via the API
    or client libraries.

    Use `dualsenseedge_full` when you want the Edge variant.

    ## Client Library Support

    The wire protocol is abstracted by client libraries.  
    The **Go client** includes built-in types (`/device/dualsense/full`),
    and the other **client libraries** provide equivalent structures with proper packing.  

    You don't need to manually construct packets,
    just use the provided types and send/receive them via the device control and feedback stream.  

    You can specify an audio port and the optional `MetaState` of the controller (serial number, shell color, battery, ...)  
    This is done by specifying it as part of the device options.

    Example:

    - `{"type":"dualsense_full", "deviceSpecific": { "audio_port": 5000, "Meta": { "shell_color": "Z1" }}}`

    ### Audio port

    `audio_port` is the TCP port for audio and audio-based haptics.  
    VIIPER will start to listen on this port, once the device is added or returns an appropriate error if the port can't be used.  

    Without `audio_port` the host still sees the audio devices of the controller,
    but the audio output (and with it the haptics) is dropped and the microphone is silent.  

    ### MetaState

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

    The input state is the same 33-byte packet as for `dualsense`.  
    See [Input State](./dualsense.md#input-state).

    ### Feedback (Rumble, LED, Adaptive Triggers & Audio Control)

    - 37-byte packets:
        - Bytes 0-30: the 31-byte feedback packet of `dualsense`  
          See [Feedback](./dualsense.md#feedback-rumble-led-adaptive-triggers)
        - SpeakerVolume: `uint8`
        - MicrophoneVolume: `uint8`
        - AudioControl: `uint8`
            - `AudioControlOutputPath`: 0x30 (bits 4-5)
        - MicMuteLed: `uint8`
        - PowerSaveControl: `uint8`
            - `PowerSaveControlMicrophoneMute`: 0x10
        - AudioControl2: `uint8`  
          Bits 0-2: gain of the speaker pre-amplifier

    The full variant uses these additional flags:

    - UpdateFlags1
        - `UpdateFlag1SpeakerVolume`: 0x20
        - `UpdateFlag1MicrophoneVolume`: 0x40
        - `UpdateFlag1AudioControl`: 0x80
    - UpdateFlags2
        - `UpdateFlag2MicMuteLed`: 0x01
        - `UpdateFlag2PowerSaveControl`: 0x02
        - `UpdateFlag2AudioControl2`: 0x80

    A field is valid when `UpdateFlags1` / `UpdateFlags2` has its flag.

    `AudioControlOutputPath` selects where audio channels 0 and 1 play:

    | Value | Headphones               | Speaker       |
    | ----- | ------------------------ | ------------- |
    | 0     | Left, right              | Off           |
    | 1     | Left channel, both sides | Off           |
    | 2     | Left channel, both sides | Right channel |
    | 3     | Off                      | Right channel |

    See `/device/dualsense/full/state.go` for the `OutputState` wire definition.

    ### Trigger effects

    The trigger effects are the same as for `dualsense`.  
    See [Trigger effects](./dualsense.md#trigger-effects).

    ## Reference

    Button, D-Pad, touchpad and IMU constants are the same as for `dualsense`.
    See [Reference](./dualsense.md#reference).

    ## Examples

    Examples of passthrough via re-emulation of a DualSense controller:
        
    - [Go](https://github.com/Alia5/VIIPER/tree/main/examples/go/dualsense_passthrough)
    - [C++](https://github.com/Alia5/VIIPER/tree/main/examples/cpp/dualsense_passthrough)

=== "libVIIPER"

    ## API

    | Function                                           | Description                                        |
    | -------------------------------------------------- | -------------------------------------------------- |
    | `CreateDualSenseFull(...)`                         | Create a virtual DualSense with audio              |
    | `CreateDualSenseEdgeFull(...)`                     | Create a virtual DualSense Edge with audio         |
    | `SetDualSenseFullAudioOutputCallback(handle, cb)`  | Register audio output callback (PCM)               |
    | `SetDualSenseFullMicrophoneCallback(handle, cb)`   | Register microphone callback (PCM)                 |
    | `SetDualSenseFullAudioControlCallback(handle, cb)` | Register audio control callback                    |
    | `DualSenseFullOutputReport(state, audio, report)`  | Build the USB output report from the output states |

    All other functions are the same as for `dualsense`
    (`SetDualSenseDeviceState`, `SetDualSenseOutputCallback`, `RemoveDualSense`, ...)  
    See [DualSense](./dualsense.md).

    The create functions have the same parameters as `CreateDualSense`.

    ## Input state, meta state, trigger effect and output callback

    `DSDeviceState`, `DSMetaState`, `DSTriggerEffect`, `DSOutputState` and `DSOutputCallback`
    are the same as for `dualsense`.  
    See [DualSense](./dualsense.md).

    ## Audio output and microphone callbacks

    ```c
    typedef void (*DSAudioOutputCallback)(DSDeviceHandle handle, const uint8_t* pcm, uint32_t size);
    typedef void (*DSMicrophoneCallback)(DSDeviceHandle handle, uint8_t* pcm, uint32_t size);
    ```

    - `DSAudioOutputCallback` is called with the audio output of the host.  
      `pcm` has 4 channels (channels 0,1 speaker; channels 2,3 haptics).
    - `DSMicrophoneCallback` is called when the host reads the microphone.  
      Fill `pcm` with 2 channels. The buffer is silence before the call.

    `pcm` is raw PCM (48 kHz, signed 16-bit little-endian) and only valid during the call.  
    The `DS_AUDIO_*` defines give the sample rate, the channel counts and the packet sizes.

    Without callbacks the host still sees the audio devices of the controller,
    but the audio output (and with it the haptics) is dropped and the microphone is silent.  

    Pass `NULL` to `SetDualSenseFullAudioOutputCallback` / `SetDualSenseFullMicrophoneCallback`
    to clear a previously registered callback.

    ## Audio control callback

    Called when the host sends volume, audio control or microphone mute commands to the device.  
    It is called before the output callback of the same output report.

    ```c
    typedef struct {
        uint8_t UpdateFlags1;     // DS_OUTPUT_UPDATE1_*
        uint8_t UpdateFlags2;     // DS_OUTPUT_UPDATE2_*
        uint8_t SpeakerVolume;
        uint8_t MicrophoneVolume;
        uint8_t AudioControl;     // DS_AUDIO_CONTROL_*
        uint8_t MicMuteLed;
        uint8_t PowerSaveControl; // DS_POWER_SAVE_CONTROL_*
        uint8_t AudioControl2;    // bits 0-2: gain of the speaker pre-amplifier
    } DSAudioControlState;

    typedef void (*DSAudioControlCallback)(DSDeviceHandle handle, DSAudioControlState state);
    ```

    Use the flags to find which fields the host set:

    - `SpeakerVolume` is valid when `UpdateFlags1` has `DS_OUTPUT_UPDATE1_SPEAKER_VOLUME`.
    - `MicrophoneVolume` is valid when `UpdateFlags1` has `DS_OUTPUT_UPDATE1_MICROPHONE_VOLUME`.
    - `AudioControl` is valid when `UpdateFlags1` has `DS_OUTPUT_UPDATE1_AUDIO_CONTROL`.
    - `MicMuteLed` is valid when `UpdateFlags2` has `DS_OUTPUT_UPDATE2_MIC_MUTE_LED`.
    - `PowerSaveControl` is valid when `UpdateFlags2` has `DS_OUTPUT_UPDATE2_POWER_SAVE_CONTROL`.
    - `AudioControl2` is valid when `UpdateFlags2` has `DS_OUTPUT_UPDATE2_AUDIO_CONTROL2`.

    `DS_AUDIO_CONTROL_OUTPUT_PATH` (0x30) and `DS_POWER_SAVE_CONTROL_MICROPHONE_MUTE` (0x10)
    have the same meaning as in the "Feedback" section of the TCP API.

    Pass `NULL` to `SetDualSenseFullAudioControlCallback` to clear a previously registered callback.

    ## Forwarding feedback to a physical controller

    `DualSenseFullOutputReport(state, audio, report)` builds the USB output report (report ID `0x02`)
    of a physical DualSense from an output state and an audio control state.  
    `report` is a buffer of `DS_OUTPUT_REPORT_SIZE` bytes.

    ## Examples

    Example of passthrough via re-emulation of a DualSense controller: [C++](https://github.com/Alia5/VIIPER/tree/main/examples/libVIIPER/cpp/dualsense_passthrough)
