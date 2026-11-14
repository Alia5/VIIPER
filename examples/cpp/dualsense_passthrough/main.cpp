/*

Example of accessing a DualSense/Edge via SDL and passing the real device through to a VIIPER emulated DualSense/Edge via VIIPER TCP Api
Run VIIPER, then this forwarder, and connect a real DualSense controller

You can use [HidHide](https://docs.nefarius.at/projects/HidHide/) to hide the real controller from other apps
CMake fetches SDL3 from the main branch
*/

#define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>
#define VIIPER_JSON_NAMESPACE nlohmann
#define VIIPER_JSON_TYPE json

#include <viiper/viiper.hpp>
#include <SDL3/SDL.h>
#include <algorithm>
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
#include <stop_token>
#include <string>
#include <string_view>
#include <thread>

#ifdef _WIN32
#include <winsock2.h>
#include <ws2tcpip.h>
constexpr int MSG_NOSIGNAL = 0;
#else
#include <netdb.h>
#include <sys/socket.h>
#include <unistd.h>
using SOCKET = int;
constexpr SOCKET INVALID_SOCKET = -1;
constexpr int SD_BOTH = SHUT_RDWR;
constexpr auto closesocket = close;
#endif

namespace dualsense = viiper::dualsense;
namespace dualsense_full = viiper::dualsense_full;

using Gamepad = std::unique_ptr<SDL_Gamepad, decltype(&SDL_CloseGamepad)>;
using AudioStream = std::unique_ptr<SDL_AudioStream, decltype(&SDL_DestroyAudioStream)>;
using AudioDevices = std::unique_ptr<SDL_AudioDeviceID[], decltype(&SDL_free)>;

constexpr int EVENT_WAIT_TIMEOUT_MS = 16;
constexpr double DEGREES_PER_RADIAN = 180.0 / std::numbers::pi;

constexpr std::uint16_t AUDIO_PORT = 5000;

const std::map<std::string, std::string> SDL_HINTS = {
    {SDL_HINT_JOYSTICK_ALLOW_BACKGROUND_EVENTS, "1"},
    {SDL_HINT_JOYSTICK_HIDAPI, "1"},
    {SDL_HINT_JOYSTICK_HIDAPI_PS5, "1"},
    {SDL_HINT_JOYSTICK_ENHANCED_REPORTS, "1"},
};

struct AudioPassthrough
{
    SOCKET conn = INVALID_SOCKET;
    std::jthread playback_thread;
    std::jthread recording_thread;

    ~AudioPassthrough()
    {
        shutdown(conn, SD_BOTH);
        recording_thread.request_stop();
        if (playback_thread.joinable())
        {
            playback_thread.join();
        }
        if (recording_thread.joinable())
        {
            recording_thread.join();
        }
        closesocket(conn);
    }
};

std::atomic<bool> running{true};

void signal_handler(int)
{
    running = false;
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
        if (SDL_GetGamepadVendorForID(id) != dualsense::DefaultVID)
        {
            continue;
        }
        const Uint16 product = SDL_GetGamepadProductForID(id);
        if (product == dualsense::DefaultPIDDS || product == dualsense::DefaultPIDDSEdge)
        {
            return id;
        }
    }
}

viiper::Result<void> send_state(SDL_Gamepad *gp, dualsense::Input &s, viiper::ViiperDevice &stream)
{
    s.buttons = 0;
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_SOUTH))
    {
        s.buttons |= dualsense::ButtonCross;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_EAST))
    {
        s.buttons |= dualsense::ButtonCircle;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_WEST))
    {
        s.buttons |= dualsense::ButtonSquare;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_NORTH))
    {
        s.buttons |= dualsense::ButtonTriangle;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_LEFT_SHOULDER))
    {
        s.buttons |= dualsense::ButtonL1;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_RIGHT_SHOULDER))
    {
        s.buttons |= dualsense::ButtonR1;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_BACK))
    {
        s.buttons |= dualsense::ButtonCreate;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_START))
    {
        s.buttons |= dualsense::ButtonOptions;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_LEFT_STICK))
    {
        s.buttons |= dualsense::ButtonL3;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_RIGHT_STICK))
    {
        s.buttons |= dualsense::ButtonR3;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_GUIDE))
    {
        s.buttons |= dualsense::ButtonPS;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_TOUCHPAD))
    {
        s.buttons |= dualsense::ButtonTouchpad;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_MISC1))
    {
        s.buttons |= dualsense::ButtonMicMute;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_LEFT_PADDLE1))
    {
        s.buttons |= dualsense::ButtonEdgeL4;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_RIGHT_PADDLE1))
    {
        s.buttons |= dualsense::ButtonEdgeR4;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_LEFT_PADDLE2))
    {
        s.buttons |= dualsense::ButtonEdgeLFn;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_RIGHT_PADDLE2))
    {
        s.buttons |= dualsense::ButtonEdgeRFn;
    }

    s.dpad = 0;
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_DPAD_UP))
    {
        s.dpad |= dualsense::DPadUp;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_DPAD_DOWN))
    {
        s.dpad |= dualsense::DPadDown;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_DPAD_LEFT))
    {
        s.dpad |= dualsense::DPadLeft;
    }
    if (SDL_GetGamepadButton(gp, SDL_GAMEPAD_BUTTON_DPAD_RIGHT))
    {
        s.dpad |= dualsense::DPadRight;
    }

    s.sticklx = static_cast<std::int8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_LEFTX) * INT8_MAX / INT16_MAX);
    s.stickly = static_cast<std::int8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_LEFTY) * INT8_MAX / INT16_MAX);
    s.stickrx = static_cast<std::int8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_RIGHTX) * INT8_MAX / INT16_MAX);
    s.stickry = static_cast<std::int8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_RIGHTY) * INT8_MAX / INT16_MAX);
    s.triggerl2 = static_cast<std::uint8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_LEFT_TRIGGER) * UINT8_MAX / INT16_MAX);
    s.triggerr2 = static_cast<std::uint8_t>(
        SDL_GetGamepadAxis(gp, SDL_GAMEPAD_AXIS_RIGHT_TRIGGER) * UINT8_MAX / INT16_MAX);
    if (s.triggerl2 > 0)
    {
        s.buttons |= dualsense::ButtonL2;
    }
    if (s.triggerr2 > 0)
    {
        s.buttons |= dualsense::ButtonR2;
    }

    return stream.send(s);
}

std::string to_string(const dualsense::TriggerEffect &effect)
{
    std::string text = std::format("mode={:02x} params=", effect.mode);
    for (const std::uint8_t param : effect.params)
    {
        text += std::format("{:02x}", param);
    }
    return text;
}

void forward_feedback(SDL_Gamepad *gamepad, const dualsense_full::Output &f)
{
    std::string feedback_log = std::format(
        "flags={:02x}/{:02x}/{:02x}",
        f.updateflags1,
        f.updateflags2,
        f.updateflags3);
    if ((f.updateflags1 & (dualsense::UpdateFlag1Rumble | dualsense::UpdateFlag1HapticsSelect)) != 0)
    {
        feedback_log += std::format(" rumbleSmall={} rumbleLarge={}", f.rumblesmall, f.rumblelarge);
    }
    if ((f.updateflags1 & dualsense::UpdateFlag1TriggerR2) != 0)
    {
        feedback_log += std::format(" triggerR2=[{}]", to_string(f.trigger_r2()));
    }
    if ((f.updateflags1 & dualsense::UpdateFlag1TriggerL2) != 0)
    {
        feedback_log += std::format(" triggerL2=[{}]", to_string(f.trigger_l2()));
    }
    if ((f.updateflags1 & dualsense::UpdateFlag1SpeakerVolume) != 0)
    {
        feedback_log += std::format(" speakerVolume={}", f.speakervolume);
    }
    if ((f.updateflags1 & dualsense::UpdateFlag1MicrophoneVolume) != 0)
    {
        feedback_log += std::format(" microphoneVolume={}", f.microphonevolume);
    }
    if ((f.updateflags1 & dualsense::UpdateFlag1AudioControl) != 0)
    {
        feedback_log += std::format(" audioControl={:02x}", f.audiocontrol);
    }
    if ((f.updateflags2 & dualsense::UpdateFlag2MicMuteLed) != 0)
    {
        feedback_log += std::format(" micMuteLed={}", f.micmuteled);
    }
    if ((f.updateflags2 & dualsense::UpdateFlag2PowerSaveControl) != 0)
    {
        feedback_log += std::format(" powerSaveControl={:02x}", f.powersavecontrol);
    }
    if ((f.updateflags2 & dualsense::UpdateFlag2AudioControl2) != 0)
    {
        feedback_log += std::format(" audioControl2={:02x}", f.audiocontrol2);
    }
    if ((f.updateflags2 & dualsense::UpdateFlag2Led) != 0)
    {
        feedback_log += std::format(" led={:02x}{:02x}{:02x}", f.ledred, f.ledgreen, f.ledblue);
    }
    if ((f.updateflags2 & dualsense::UpdateFlag2PlayerLeds) != 0)
    {
        feedback_log += std::format(" playerLeds={:06b}", f.playerleds);
    }
    std::cout << "Received feedback from the VIIPER device " << feedback_log << "\n";

    const auto rumble = static_cast<std::uint8_t>(
        f.updateflags1 & (dualsense::UpdateFlag1Rumble | dualsense::UpdateFlag1HapticsSelect));
    const auto triggers = static_cast<std::uint8_t>(
        f.updateflags1 & (dualsense::UpdateFlag1TriggerR2 | dualsense::UpdateFlag1TriggerL2));
    const auto player_leds = static_cast<std::uint8_t>(f.updateflags2 & dualsense::UpdateFlag2PlayerLeds);
    const auto microphone = static_cast<std::uint8_t>(
        f.updateflags2 & (dualsense::UpdateFlag2MicMuteLed | dualsense::UpdateFlag2PowerSaveControl));
    const auto speaker = static_cast<std::uint8_t>(
        f.updateflags1 & (dualsense::UpdateFlag1SpeakerVolume | dualsense::UpdateFlag1AudioControl));
    const auto speaker_gain = static_cast<std::uint8_t>(f.updateflags2 & dualsense::UpdateFlag2AudioControl2);

    Uint16 rumble_large = 0;
    Uint16 rumble_small = 0;
    if (rumble != 0)
    {
        rumble_large = static_cast<Uint16>(f.rumblelarge * (UINT16_MAX / UINT8_MAX));
        rumble_small = static_cast<Uint16>(f.rumblesmall * (UINT16_MAX / UINT8_MAX));
    }
    if (!SDL_RumbleGamepad(gamepad, rumble_large, rumble_small, UINT32_MAX))
    {
        std::cerr << "Setting the rumble of the controller failed error=" << SDL_GetError() << "\n";
    }

    if ((f.updateflags2 & dualsense::UpdateFlag2Led) != 0)
    {
        if (!SDL_SetGamepadLED(gamepad, f.ledred, f.ledgreen, f.ledblue))
        {
            std::cerr << "Setting the LED color of the controller failed error=" << SDL_GetError() << "\n";
        }
    }

    if (triggers == 0 && player_leds == 0 && microphone == 0 && speaker == 0 && speaker_gain == 0)
    {
        return;
    }
    dualsense_full::Output effects = f;
    effects.updateflags1 = static_cast<std::uint8_t>(rumble | triggers | speaker);
    effects.updateflags2 = static_cast<std::uint8_t>(player_leds | microphone | speaker_gain);
    effects.updateflags3 &= dualsense::UpdateFlag3ImprovedRumble;
    // The other AudioControl bits from PS Remote Play silence the microphone of the controller.
    effects.audiocontrol &= dualsense_full::AudioControlOutputPath;
    const auto report = effects.output_report();
    const auto payload = std::span(report).subspan(dualsense::OutputReportOffsetUpdateFlags1);
    if (!SDL_SendGamepadEffect(gamepad, payload.data(), static_cast<int>(payload.size())))
    {
        std::cerr << "Sending the effects to the controller failed error=" << SDL_GetError() << "\n";
    }
}

SOCKET connect_tcp(const std::string &host, std::uint16_t port)
{
#ifdef _WIN32
    WSADATA winsock;
    WSAStartup(WINSOCK_VERSION, &winsock);
#endif
    const addrinfo hints{
        .ai_family = AF_UNSPEC,
        .ai_socktype = SOCK_STREAM,
    };
    addrinfo *resolved = nullptr;
    if (getaddrinfo(host.c_str(), std::to_string(port).c_str(), &hints, &resolved) != 0)
    {
        return INVALID_SOCKET;
    }
    const std::unique_ptr<addrinfo, decltype(&freeaddrinfo)> addresses(resolved, freeaddrinfo);
    for (const addrinfo *address = addresses.get(); address != nullptr; address = address->ai_next)
    {
        const SOCKET conn = socket(address->ai_family, address->ai_socktype, address->ai_protocol);
        if (conn == INVALID_SOCKET)
        {
            continue;
        }
        if (connect(conn, address->ai_addr, static_cast<int>(address->ai_addrlen)) == 0)
        {
            return conn;
        }
        closesocket(conn);
    }
    return INVALID_SOCKET;
}

viiper::Result<void> forward_audio(SDL_Gamepad *gamepad, const std::string &host, AudioPassthrough &audio)
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
        return viiper::Error("found no playback and recording device of the controller");
    }

    const SDL_AudioSpec playback_spec = {
        .format = SDL_AUDIO_S16LE,
        .channels = dualsense_full::AudioOutputChannels,
        .freq = dualsense_full::AudioSampleRate,
    };
    AudioStream playback(
        SDL_OpenAudioDeviceStream(playback_id, &playback_spec, nullptr, nullptr),
        SDL_DestroyAudioStream);
    if (!playback)
    {
        return viiper::Error(SDL_GetError());
    }
    const SDL_AudioSpec recording_spec = {
        .format = SDL_AUDIO_S16LE,
        .channels = dualsense_full::AudioMicrophoneChannels,
        .freq = dualsense_full::AudioSampleRate,
    };
    AudioStream recording(
        SDL_OpenAudioDeviceStream(recording_id, &recording_spec, nullptr, nullptr),
        SDL_DestroyAudioStream);
    if (!recording)
    {
        return viiper::Error(SDL_GetError());
    }
    if (!SDL_ResumeAudioStreamDevice(playback.get()))
    {
        return viiper::Error(SDL_GetError());
    }
    if (!SDL_ResumeAudioStreamDevice(recording.get()))
    {
        return viiper::Error(SDL_GetError());
    }
    audio.conn = connect_tcp(host, AUDIO_PORT);
    if (audio.conn == INVALID_SOCKET)
    {
        return viiper::Error("connecting to the audio port failed");
    }

    audio.playback_thread = std::jthread(
        [playback = std::move(playback), conn = audio.conn]()
        {
            std::array<char, dualsense_full::AudioOutputPacketSize> packet{};
            const int size = static_cast<int>(packet.size());
            while (recv(conn, packet.data(), size, MSG_WAITALL) == size)
            {
                SDL_PutAudioStreamData(playback.get(), packet.data(), size);
            }
        });
    audio.recording_thread = std::jthread(
        [recording = std::move(recording), conn = audio.conn](std::stop_token stop)
        {
            std::array<char, dualsense_full::AudioMicrophonePacketSize> packet{};
            while (!stop.stop_requested())
            {
                const int size = SDL_GetAudioStreamData(
                    recording.get(),
                    packet.data(),
                    static_cast<int>(packet.size()));
                if (size <= 0)
                {
                    SDL_Delay(1);
                    continue;
                }
                if (send(conn, packet.data(), size, MSG_NOSIGNAL) != size)
                {
                    return;
                }
            }
        });
    std::cout << std::format(
        "Started the audio passthrough playback=\"{}\" recording=\"{}\"\n",
        playback_name,
        recording_name);
    return viiper::Result<void>();
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

    viiper::ViiperClient client("localhost");
    auto buses = client.buslist();
    if (buses.is_error())
    {
        std::cerr << "Listing the VIIPER buses failed error=" << buses.error().to_string() << "\n";
        return 1;
    }
    std::uint32_t bus_id = 0;
    bool created_bus = false;
    if (buses.value().buses.empty())
    {
        auto bus = client.buscreate(std::nullopt);
        if (bus.is_error())
        {
            std::cerr << "Creating a VIIPER bus failed error=" << bus.error().to_string() << "\n";
            return 1;
        }
        bus_id = bus.value().busid;
        created_bus = true;
    }
    else
    {
        bus_id = *std::ranges::min_element(buses.value().buses);
    }

    auto addResponse = client.addDeviceAndConnect(
        bus_id,
        {
            .type = std::string(dualsense_full::DEVICE_TYPE_EDGE),
            .devicespecific = dualsense_full::CreateOptions{.AudioPort = AUDIO_PORT}.to_map(),
        });
    if (addResponse.is_error())
    {
        std::cerr << "Creating the VIIPER device failed error=" << addResponse.error().to_string() << "\n";
        if (created_bus)
        {
            client.busremove(bus_id);
        }
        return 1;
    }
    const auto &[device, stream] = addResponse.value();
    std::cout << std::format(
        "Created the VIIPER device type={} bus={} dev={}\n",
        device.type,
        device.busid,
        device.devid);

    stream->on_disconnect(
        []()
        {
            running = false;
        });
    stream->on_error(
        [](const viiper::Error &error)
        {
            std::cerr << "Reading the feedback of the VIIPER device failed error="
                      << error.to_string() << "\n";
        });
    stream->on_output(
        dualsense_full::OUTPUT_SIZE,
        [gamepad = gamepad.get()](const std::uint8_t *bytes, std::size_t size)
        {
            auto output = dualsense_full::Output::from_bytes(bytes, size);
            if (!output.is_error())
            {
                forward_feedback(gamepad, output.value());
            }
        });

    AudioPassthrough audio;
    auto forwarding = forward_audio(gamepad.get(), client.host(), audio);
    if (forwarding.is_error())
    {
        std::cerr << "Starting the audio passthrough failed error=" << forwarding.error().to_string() << "\n";
    }

    std::signal(SIGINT, signal_handler);
    dualsense::Input state{
        .accelx = dualsense::DefaultAccelXRaw,
        .accely = dualsense::DefaultAccelYRaw,
        .accelz = dualsense::DefaultAccelZRaw,
    };
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
                state.gyrox = dualsense::gyro_dps_to_raw(ev.gsensor.data[0] * DEGREES_PER_RADIAN);
                state.gyroy = dualsense::gyro_dps_to_raw(ev.gsensor.data[1] * DEGREES_PER_RADIAN);
                state.gyroz = dualsense::gyro_dps_to_raw(ev.gsensor.data[2] * DEGREES_PER_RADIAN);
                break;
            case SDL_SENSOR_ACCEL:
                state.accelx = dualsense::accel_ms2_to_raw(ev.gsensor.data[0]);
                state.accely = dualsense::accel_ms2_to_raw(ev.gsensor.data[1]);
                state.accelz = dualsense::accel_ms2_to_raw(ev.gsensor.data[2]);
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
                state.touch1active = active;
                if (active)
                {
                    state.touch1x = static_cast<std::uint16_t>(
                        std::round(ev.gtouchpad.x * dualsense::TouchpadMaxX));
                    state.touch1y = static_cast<std::uint16_t>(
                        std::round(ev.gtouchpad.y * dualsense::TouchpadMaxY));
                }
                break;
            case 1:
                state.touch2active = active;
                if (active)
                {
                    state.touch2x = static_cast<std::uint16_t>(
                        std::round(ev.gtouchpad.x * dualsense::TouchpadMaxX));
                    state.touch2y = static_cast<std::uint16_t>(
                        std::round(ev.gtouchpad.y * dualsense::TouchpadMaxY));
                }
                break;
            }
            break;
        }
        case SDL_EVENT_GAMEPAD_UPDATE_COMPLETE:
        {
            auto sent = send_state(gamepad.get(), state, *stream);
            if (sent.is_error())
            {
                std::cerr << "Sending the input state to the VIIPER device failed error="
                          << sent.error().to_string() << "\n";
                running = false;
            }
            break;
        }
        case SDL_EVENT_GAMEPAD_REMOVED:
            if (ev.gdevice.which == SDL_GetGamepadID(gamepad.get()))
            {
                std::cout << "The controller was disconnected\n";
                running = false;
            }
            break;
        }
    }

    stream->stop();
    client.busdeviceremove(device.busid, device.devid);
    SDL_RumbleGamepad(gamepad.get(), 0, 0, 0);
    if (created_bus)
    {
        client.busremove(bus_id);
    }
    return 0;
}
