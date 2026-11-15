/*

Example of accessing a DualSense/Edge via SDL and passing the real device through to a VIIPER emulated DualSense/Edge via libVIIPER
Build libVIIPER, then run this forwarder, and connect a real DualSense controller.

You can use [HidHide](https://docs.nefarius.at/projects/HidHide/) to hide the real controller from other apps.
CMake fetches SDL3 from the main branch (needs newer than 7b82a40bbf1aad01d9b0f1f3bc18273a1ef9b6b4).
*/

#include "libVIIPER.h"
#include <SDL3/SDL.h>
#include <array>
#include <atomic>
#include <cmath>
#include <csignal>
#include <cstdint>
#include <cstdlib>
#include <format>
#include <iostream>
#include <map>
#include <memory>
#include <numbers>
#include <span>
#include <stdexcept>
#include <string>
#include <string_view>

using Gamepad = std::unique_ptr<SDL_Gamepad, decltype(&SDL_CloseGamepad)>;
using AudioStream = std::unique_ptr<SDL_AudioStream, decltype(&SDL_DestroyAudioStream)>;
using AudioDevices = std::unique_ptr<SDL_AudioDeviceID[], decltype(&SDL_free)>;

constexpr int EVENT_WAIT_TIMEOUT_MS = 16;
constexpr double DEGREES_PER_RADIAN = 180.0 / std::numbers::pi;

constexpr int RECORDING_BUFFER_MAX_MS = 100;
constexpr int RECORDING_BUFFER_MAX_SIZE = RECORDING_BUFFER_MAX_MS * DS_AUDIO_MICROPHONE_PACKET_SIZE;

const std::map<std::string, std::string> SDL_HINTS = {
    {SDL_HINT_JOYSTICK_ALLOW_BACKGROUND_EVENTS, "1"},
    {SDL_HINT_JOYSTICK_HIDAPI, "1"},
    {SDL_HINT_JOYSTICK_HIDAPI_PS5, "1"},
    {SDL_HINT_JOYSTICK_ENHANCED_REPORTS, "1"},
};

struct AudioPassthrough
{
    AudioStream playback = AudioStream(nullptr, SDL_DestroyAudioStream);
    AudioStream recording = AudioStream(nullptr, SDL_DestroyAudioStream);
};

std::atomic<bool> running{true};
SDL_Gamepad *controller = nullptr;
SDL_AudioStream *controller_playback = nullptr;
SDL_AudioStream *controller_recording = nullptr;
std::atomic<DSAudioControlState> audio_control{};

void signal_handler(int)
{
    running = false;
}

void print_server_log(VIIPERLogLevel level, const char *message)
{
    if (level < VIIPER_LOG_INFO)
    {
        return;
    }
    std::cout << "libVIIPER: " << message << "\n";
}

SDL_JoystickID wait_for_dualsense()
{
    while (true)
    {
        SDL_Event ev;
        if (!SDL_WaitEventTimeout(&ev, EVENT_WAIT_TIMEOUT_MS) || ev.type != SDL_EVENT_GAMEPAD_ADDED)
        {
            continue;
        }
        const SDL_JoystickID id = ev.gdevice.which;
        if (SDL_GetGamepadVendorForID(id) != DS_VENDOR_ID)
        {
            continue;
        }
        const Uint16 product = SDL_GetGamepadProductForID(id);
        if (product == DS_PRODUCT_ID || product == DS_PRODUCT_ID_EDGE)
        {
            return id;
        }
    }
}

bool send_state(SDL_Gamepad *gp, DSDeviceState &s, DSDeviceHandle device)
{
    s.Buttons = 0;
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_SOUTH))
    {
        s.Buttons |= DS_BUTTON_CROSS;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_EAST))
    {
        s.Buttons |= DS_BUTTON_CIRCLE;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_WEST))
    {
        s.Buttons |= DS_BUTTON_SQUARE;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_NORTH))
    {
        s.Buttons |= DS_BUTTON_TRIANGLE;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_LEFT_SHOULDER))
    {
        s.Buttons |= DS_BUTTON_L1;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_RIGHT_SHOULDER))
    {
        s.Buttons |= DS_BUTTON_R1;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_BACK))
    {
        s.Buttons |= DS_BUTTON_CREATE;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_START))
    {
        s.Buttons |= DS_BUTTON_OPTIONS;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_LEFT_STICK))
    {
        s.Buttons |= DS_BUTTON_L3;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_RIGHT_STICK))
    {
        s.Buttons |= DS_BUTTON_R3;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_GUIDE))
    {
        s.Buttons |= DS_BUTTON_PS;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_TOUCHPAD))
    {
        s.Buttons |= DS_BUTTON_TOUCHPAD;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_MISC1))
    {
        s.Buttons |= DS_BUTTON_MIC_MUTE;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_LEFT_PADDLE1))
    {
        s.Buttons |= DS_BUTTON_L4;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_RIGHT_PADDLE1))
    {
        s.Buttons |= DS_BUTTON_R4;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_LEFT_PADDLE2))
    {
        s.Buttons |= DS_BUTTON_LFN;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_RIGHT_PADDLE2))
    {
        s.Buttons |= DS_BUTTON_RFN;
    }

    s.DPad = 0;
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_DPAD_UP))
    {
        s.DPad |= DS_DPAD_UP;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_DPAD_DOWN))
    {
        s.DPad |= DS_DPAD_DOWN;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_DPAD_LEFT))
    {
        s.DPad |= DS_DPAD_LEFT;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_DPAD_RIGHT))
    {
        s.DPad |= DS_DPAD_RIGHT;
    }

    s.LX = static_cast<std::int8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_LEFTX) * INT8_MAX / INT16_MAX);
    s.LY = static_cast<std::int8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_LEFTY) * INT8_MAX / INT16_MAX);
    s.RX = static_cast<std::int8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_RIGHTX) * INT8_MAX / INT16_MAX);
    s.RY = static_cast<std::int8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_RIGHTY) * INT8_MAX / INT16_MAX);
    s.L2 = static_cast<std::uint8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_LEFT_TRIGGER) * UINT8_MAX / INT16_MAX);
    s.R2 = static_cast<std::uint8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_RIGHT_TRIGGER) * UINT8_MAX / INT16_MAX);
    if (s.L2 > 0)
    {
        s.Buttons |= DS_BUTTON_L2;
    }
    if (s.R2 > 0)
    {
        s.Buttons |= DS_BUTTON_R2;
    }

    return SetDualSenseDeviceState(device, s) != 0;
}

std::string to_string(const DSTriggerEffect &effect)
{
    std::string text = std::format("mode={:02x} params=", effect.Mode);
    for (const std::uint8_t param : effect.Params)
    {
        text += std::format("{:02x}", param);
    }
    return text;
}

void store_audio_control(DSDeviceHandle, DSAudioControlState state)
{
    audio_control = state;
}

void forward_audio_output(DSDeviceHandle, const std::uint8_t *pcm, std::uint32_t size)
{
    SDL_PutAudioStreamData(controller_playback, pcm, static_cast<int>(size));
}

void forward_microphone(DSDeviceHandle, std::uint8_t *pcm, std::uint32_t size)
{
    SDL_GetAudioStreamData(controller_recording, pcm, static_cast<int>(size));
}

void limit_recording_buffer(void *, SDL_AudioStream *recording, int, int)
{
    std::array<std::uint8_t, DS_AUDIO_MICROPHONE_PACKET_SIZE> old{};
    while (SDL_GetAudioStreamAvailable(recording) > RECORDING_BUFFER_MAX_SIZE)
    {
        SDL_GetAudioStreamData(recording, old.data(), static_cast<int>(old.size()));
    }
}

void forward_feedback(DSDeviceHandle, DSOutputState f)
{
    const DSAudioControlState audio = audio_control;
    std::string feedback_log = std::format(
        "flags={:02x}/{:02x}/{:02x}",
        f.UpdateFlags1,
        f.UpdateFlags2,
        f.UpdateFlags3);
    if ((f.UpdateFlags1 & (DS_OUTPUT_UPDATE1_RUMBLE | DS_OUTPUT_UPDATE1_HAPTICS_SELECT)) != 0)
    {
        feedback_log += std::format(" rumbleSmall={} rumbleLarge={}", f.RumbleSmall, f.RumbleLarge);
    }
    if ((f.UpdateFlags1 & DS_OUTPUT_UPDATE1_TRIGGER_R2) != 0)
    {
        feedback_log += std::format(" triggerR2=[{}]", to_string(f.TriggerR2));
    }
    if ((f.UpdateFlags1 & DS_OUTPUT_UPDATE1_TRIGGER_L2) != 0)
    {
        feedback_log += std::format(" triggerL2=[{}]", to_string(f.TriggerL2));
    }
    if ((f.UpdateFlags1 & DS_OUTPUT_UPDATE1_SPEAKER_VOLUME) != 0)
    {
        feedback_log += std::format(" speakerVolume={}", audio.SpeakerVolume);
    }
    if ((f.UpdateFlags1 & DS_OUTPUT_UPDATE1_MICROPHONE_VOLUME) != 0)
    {
        feedback_log += std::format(" microphoneVolume={}", audio.MicrophoneVolume);
    }
    if ((f.UpdateFlags1 & DS_OUTPUT_UPDATE1_AUDIO_CONTROL) != 0)
    {
        feedback_log += std::format(" audioControl={:02x}", audio.AudioControl);
    }
    if ((f.UpdateFlags2 & DS_OUTPUT_UPDATE2_MIC_MUTE_LED) != 0)
    {
        feedback_log += std::format(" micMuteLed={}", audio.MicMuteLed);
    }
    if ((f.UpdateFlags2 & DS_OUTPUT_UPDATE2_POWER_SAVE_CONTROL) != 0)
    {
        feedback_log += std::format(" powerSaveControl={:02x}", audio.PowerSaveControl);
    }
    if ((f.UpdateFlags2 & DS_OUTPUT_UPDATE2_AUDIO_CONTROL2) != 0)
    {
        feedback_log += std::format(" audioControl2={:02x}", audio.AudioControl2);
    }
    if ((f.UpdateFlags2 & DS_OUTPUT_UPDATE2_LED) != 0)
    {
        feedback_log += std::format(" led={:02x}{:02x}{:02x}", f.LedRed, f.LedGreen, f.LedBlue);
    }
    if ((f.UpdateFlags2 & DS_OUTPUT_UPDATE2_PLAYER_LEDS) != 0)
    {
        feedback_log += std::format(" playerLeds={:06b}", f.PlayerLeds);
    }
    std::cout << "Received feedback from the VIIPER device " << feedback_log << "\n";

    const auto rumble = static_cast<std::uint8_t>(
        f.UpdateFlags1 & (DS_OUTPUT_UPDATE1_RUMBLE | DS_OUTPUT_UPDATE1_HAPTICS_SELECT));
    const auto triggers = static_cast<std::uint8_t>(
        f.UpdateFlags1 & (DS_OUTPUT_UPDATE1_TRIGGER_R2 | DS_OUTPUT_UPDATE1_TRIGGER_L2));
    const auto player_leds = static_cast<std::uint8_t>(f.UpdateFlags2 & DS_OUTPUT_UPDATE2_PLAYER_LEDS);
    const auto microphone = static_cast<std::uint8_t>(
        f.UpdateFlags2 & (DS_OUTPUT_UPDATE2_MIC_MUTE_LED | DS_OUTPUT_UPDATE2_POWER_SAVE_CONTROL));
    const auto speaker = static_cast<std::uint8_t>(
        f.UpdateFlags1 & (DS_OUTPUT_UPDATE1_SPEAKER_VOLUME | DS_OUTPUT_UPDATE1_AUDIO_CONTROL));
    const auto speaker_gain = static_cast<std::uint8_t>(f.UpdateFlags2 & DS_OUTPUT_UPDATE2_AUDIO_CONTROL2);

    Uint16 rumble_large = 0;
    Uint16 rumble_small = 0;
    if (rumble != 0)
    {
        rumble_large = static_cast<Uint16>(f.RumbleLarge * (UINT16_MAX / UINT8_MAX));
        rumble_small = static_cast<Uint16>(f.RumbleSmall * (UINT16_MAX / UINT8_MAX));
    }
    if (!SDL_RumbleGamepad(controller, rumble_large, rumble_small, UINT32_MAX))
    {
        std::cerr << "Setting the rumble of the controller failed error=" << SDL_GetError() << "\n";
    }

    if ((f.UpdateFlags2 & DS_OUTPUT_UPDATE2_LED) != 0)
    {
        if (!SDL_SetGamepadLED(controller, f.LedRed, f.LedGreen, f.LedBlue))
        {
            std::cerr << "Setting the LED color of the controller failed error=" << SDL_GetError() << "\n";
        }
    }

    if (triggers == 0 && player_leds == 0 && microphone == 0 && speaker == 0 && speaker_gain == 0)
    {
        return;
    }
    DSOutputState effects = f;
    effects.UpdateFlags1 = static_cast<std::uint8_t>(rumble | triggers | speaker);
    effects.UpdateFlags2 = static_cast<std::uint8_t>(player_leds | microphone | speaker_gain);
    effects.UpdateFlags3 &= DS_OUTPUT_UPDATE3_IMPROVED_RUMBLE;
    DSAudioControlState audio_effects = audio;
    audio_effects.AudioControl &= DS_AUDIO_CONTROL_OUTPUT_PATH;
    std::array<std::uint8_t, DS_OUTPUT_REPORT_SIZE> report{};
    DualSenseFullOutputReport(effects, audio_effects, report.data());
    const auto payload = std::span(report).subspan(DS_OUTPUT_REPORT_OFFSET_UPDATE_FLAGS1);
    if (!SDL_SendGamepadEffect(controller, payload.data(), static_cast<int>(payload.size())))
    {
        std::cerr << "Sending the effects to the controller failed error=" << SDL_GetError() << "\n";
    }
}

AudioPassthrough forward_audio(SDL_Gamepad *gamepad, DSDeviceHandle device)
{
    const std::string container_id = SDL_GetStringProperty(
        SDL_GetGamepadProperties(gamepad),
        SDL_PROP_GAMEPAD_CONTAINER_ID_STRING,
        "");
    SDL_AudioDeviceID playback_id = 0;
    SDL_AudioDeviceID recording_id = 0;
    std::string playback_name;
    std::string recording_name;
    int playback_count = 0;
    const AudioDevices playback_devices(SDL_GetAudioPlaybackDevices(&playback_count), SDL_free);
    for (const SDL_AudioDeviceID id : std::span(playback_devices.get(), static_cast<std::size_t>(playback_count)))
    {
        const char *device_container_id = SDL_GetStringProperty(
            SDL_GetAudioDeviceProperties(id),
            SDL_PROP_AUDIO_DEVICE_CONTAINER_ID_STRING,
            "");
        if (container_id != device_container_id)
        {
            continue;
        }
        const char *name = SDL_GetAudioDeviceName(id);
        if (name != nullptr && std::string_view(name).find("DualSense") != std::string_view::npos)
        {
            playback_id = id;
            playback_name = name;
        }
    }
    int recording_count = 0;
    const AudioDevices recording_devices(SDL_GetAudioRecordingDevices(&recording_count), SDL_free);
    for (const SDL_AudioDeviceID id : std::span(recording_devices.get(), static_cast<std::size_t>(recording_count)))
    {
        const char *device_container_id = SDL_GetStringProperty(
            SDL_GetAudioDeviceProperties(id),
            SDL_PROP_AUDIO_DEVICE_CONTAINER_ID_STRING,
            "");
        if (container_id != device_container_id)
        {
            continue;
        }
        const char *name = SDL_GetAudioDeviceName(id);
        if (name != nullptr && std::string_view(name).find("DualSense") != std::string_view::npos)
        {
            recording_id = id;
            recording_name = name;
        }
    }
    if (playback_id == 0 || recording_id == 0)
    {
        throw std::runtime_error("found no playback and recording device of the controller");
    }

    const SDL_AudioSpec playback_spec = {
        .format = SDL_AUDIO_S16LE,
        .channels = DS_AUDIO_OUTPUT_CHANNELS,
        .freq = DS_AUDIO_SAMPLE_RATE,
    };
    AudioPassthrough audio;
    audio.playback.reset(SDL_OpenAudioDeviceStream(playback_id, &playback_spec, nullptr, nullptr));
    if (!audio.playback)
    {
        throw std::runtime_error(SDL_GetError());
    }
    const SDL_AudioSpec recording_spec = {
        .format = SDL_AUDIO_S16LE,
        .channels = DS_AUDIO_MICROPHONE_CHANNELS,
        .freq = DS_AUDIO_SAMPLE_RATE,
    };
    audio.recording.reset(
        SDL_OpenAudioDeviceStream(recording_id, &recording_spec, limit_recording_buffer, nullptr));
    if (!audio.recording)
    {
        throw std::runtime_error(SDL_GetError());
    }
    if (!SDL_ResumeAudioStreamDevice(audio.playback.get()))
    {
        throw std::runtime_error(SDL_GetError());
    }
    if (!SDL_ResumeAudioStreamDevice(audio.recording.get()))
    {
        throw std::runtime_error(SDL_GetError());
    }

    controller_playback = audio.playback.get();
    controller_recording = audio.recording.get();
    SetDualSenseFullAudioOutputCallback(device, forward_audio_output);
    SetDualSenseFullMicrophoneCallback(device, forward_microphone);
    std::cout << std::format(
        "Started the audio passthrough playback=\"{}\" recording=\"{}\"\n",
        playback_name,
        recording_name);
    return audio;
}

int main()
{
    for (const auto &[hint, value] : SDL_HINTS)
    {
        if (!SDL_SetHint(hint.c_str(), value.c_str()))
        {
            std::cerr << "Setting an SDL hint failed hint=" << hint << " error=" << SDL_GetError() << "\n";
            return 1;
        }
    }
    if (!SDL_InitSubSystem(SDL_INIT_GAMEPAD | SDL_INIT_SENSOR | SDL_INIT_AUDIO))
    {
        std::cerr << "Initializing SDL failed error=" << SDL_GetError() << "\n";
        return 1;
    }
    std::atexit(SDL_Quit);

    std::cout << "Waiting for a DualSense controller\n";
    const Gamepad gamepad(SDL_OpenGamepad(wait_for_dualsense()), SDL_CloseGamepad);
    if (!gamepad)
    {
        std::cerr << "Opening the controller failed error=" << SDL_GetError() << "\n";
        return 1;
    }
    if (!SDL_SetGamepadSensorEnabled(gamepad.get(), SDL_SENSOR_GYRO, true))
    {
        std::cerr << "Enabling the gyroscope failed error=" << SDL_GetError() << "\n";
    }
    if (!SDL_SetGamepadSensorEnabled(gamepad.get(), SDL_SENSOR_ACCEL, true))
    {
        std::cerr << "Enabling the accelerometer failed error=" << SDL_GetError() << "\n";
    }
    const char *name = SDL_GetGamepadName(gamepad.get());
    const char *serial = SDL_GetGamepadSerial(gamepad.get());
    const char *path = SDL_GetGamepadPath(gamepad.get());
    std::cout << std::format(
        "Opened the controller name=\"{}\" serial={} path={}\n",
        name != nullptr ? name : "",
        serial != nullptr ? serial : "",
        path != nullptr ? path : "");

    std::string server_addr = "localhost:3245";
    USBServerConfig config{
        .addr = server_addr.data(),
    };
    USBServerHandle server = 0;
    if (!NewUSBServer(&config, &server, print_server_log))
    {
        std::cerr << "Creating the USB server failed\n";
        return 1;
    }
    std::uint32_t bus_id = 0;
    if (!CreateUSBBus(server, &bus_id))
    {
        std::cerr << "Creating a USB bus failed\n";
        CloseUSBServer(server);
        return 1;
    }
    DSDeviceHandle device = 0;
    if (!CreateDualSenseEdgeFull(server, &device, bus_id, true, 0, 0, nullptr))
    {
        std::cerr << "Creating the VIIPER device failed\n";
        CloseUSBServer(server);
        return 1;
    }
    std::cout << std::format("Created the VIIPER device bus={}\n", bus_id);

    controller = gamepad.get();
    SetDualSenseFullAudioControlCallback(device, store_audio_control);
    SetDualSenseOutputCallback(device, forward_feedback);

    AudioPassthrough audio;
    try
    {
        audio = forward_audio(gamepad.get(), device);
    }
    catch (const std::runtime_error &error)
    {
        std::cerr << "Starting the audio passthrough failed error=" << error.what() << "\n";
    }

    std::signal(SIGINT, signal_handler);
    DSDeviceState state{};
    while (running)
    {
        SDL_Event ev;
        if (!SDL_WaitEventTimeout(&ev, EVENT_WAIT_TIMEOUT_MS))
        {
            continue;
        }
        switch (ev.type)
        {
        case SDL_EVENT_QUIT:
            running = false;
            break;
        case SDL_EVENT_GAMEPAD_SENSOR_UPDATE:
            switch (ev.gsensor.sensor)
            {
            case SDL_SENSOR_GYRO:
                state.GyroX = DualSenseGyroDpsToRaw(ev.gsensor.data[0] * DEGREES_PER_RADIAN);
                state.GyroY = DualSenseGyroDpsToRaw(ev.gsensor.data[1] * DEGREES_PER_RADIAN);
                state.GyroZ = DualSenseGyroDpsToRaw(ev.gsensor.data[2] * DEGREES_PER_RADIAN);
                break;
            case SDL_SENSOR_ACCEL:
                state.AccelX = DualSenseAccelMS2ToRaw(ev.gsensor.data[0]);
                state.AccelY = DualSenseAccelMS2ToRaw(ev.gsensor.data[1]);
                state.AccelZ = DualSenseAccelMS2ToRaw(ev.gsensor.data[2]);
                break;
            }
            break;
        case SDL_EVENT_GAMEPAD_TOUCHPAD_DOWN:
        case SDL_EVENT_GAMEPAD_TOUCHPAD_MOTION:
        case SDL_EVENT_GAMEPAD_TOUCHPAD_UP:
        {
            const bool active = ev.type != SDL_EVENT_GAMEPAD_TOUCHPAD_UP;
            switch (ev.gtouchpad.finger)
            {
            case 0:
                state.Touch1Active = active;
                if (active)
                {
                    state.Touch1X = static_cast<std::uint16_t>(
                        std::round(ev.gtouchpad.x * DS_TOUCHPAD_MAX_X));
                    state.Touch1Y = static_cast<std::uint16_t>(
                        std::round(ev.gtouchpad.y * DS_TOUCHPAD_MAX_Y));
                }
                break;
            case 1:
                state.Touch2Active = active;
                if (active)
                {
                    state.Touch2X = static_cast<std::uint16_t>(
                        std::round(ev.gtouchpad.x * DS_TOUCHPAD_MAX_X));
                    state.Touch2Y = static_cast<std::uint16_t>(
                        std::round(ev.gtouchpad.y * DS_TOUCHPAD_MAX_Y));
                }
                break;
            }
            break;
        }
        case SDL_EVENT_GAMEPAD_UPDATE_COMPLETE:
            if (!send_state(gamepad.get(), state, device))
            {
                std::cerr << "Sending the input state to the VIIPER device failed\n";
                running = false;
            }
            break;
        case SDL_EVENT_GAMEPAD_REMOVED:
            if (ev.gdevice.which == SDL_GetGamepadID(gamepad.get()))
            {
                std::cout << "The controller was disconnected\n";
                running = false;
            }
            break;
        }
    }

    CloseUSBServer(server);
    SDL_RumbleGamepad(gamepad.get(), 0, 0, 0);
    return 0;
}
