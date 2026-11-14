#pragma once

#include "../detail/json.hpp"
#include "../error.hpp"
#include "dualsense.hpp"
#include <array>
#include <cstddef>
#include <cstdint>
#include <optional>
#include <string_view>

namespace viiper
{
    namespace dualsense_full
    {

        constexpr std::string_view DEVICE_TYPE = "dualsense_full";
        constexpr std::string_view DEVICE_TYPE_EDGE = "dualsenseedge_full";

        constexpr std::size_t OUTPUT_SIZE = dualsense::OUTPUT_SIZE + 6;

        constexpr std::uint64_t PowerSaveControlMicrophoneMute = 16;
        constexpr std::uint8_t AudioControlOutputPath = 48;

        constexpr std::size_t OutputReportOffsetSpeakerVolume = 6;
        constexpr std::size_t OutputReportOffsetMicrophoneVolume = 7;
        constexpr std::size_t OutputReportOffsetAudioControl = 8;
        constexpr std::size_t OutputReportOffsetMicMuteLed = 9;
        constexpr std::size_t OutputReportOffsetPowerSaveControl = 10;
        constexpr std::size_t OutputReportOffsetAudioControl2 = 38;

        constexpr int AudioSampleRate = 48000;
        constexpr int AudioBytesPerSample = 2;
        constexpr int AudioOutputChannels = 4;
        constexpr int AudioMicrophoneChannels = 2;
        constexpr int AudioOutputPacketSize = AudioSampleRate / 1000 * AudioOutputChannels * AudioBytesPerSample;
        constexpr int AudioMicrophonePacketSize = AudioSampleRate / 1000 * AudioMicrophoneChannels * AudioBytesPerSample;

        using Input = dualsense::Input;
        using MetaState = dualsense::MetaState;
        using TriggerEffect = dualsense::TriggerEffect;

        struct Output : dualsense::Output
        {
            std::uint8_t speakervolume = 0;
            std::uint8_t microphonevolume = 0;
            std::uint8_t audiocontrol = 0;
            std::uint8_t micmuteled = 0;
            std::uint8_t powersavecontrol = 0;
            std::uint8_t audiocontrol2 = 0;

            [[nodiscard]] std::array<std::uint8_t, dualsense::OutputReportSize> output_report() const
            {
                auto report = dualsense::Output::output_report();
                report[OutputReportOffsetSpeakerVolume] = speakervolume;
                report[OutputReportOffsetMicrophoneVolume] = microphonevolume;
                report[OutputReportOffsetAudioControl] = audiocontrol;
                report[OutputReportOffsetMicMuteLed] = micmuteled;
                report[OutputReportOffsetPowerSaveControl] = powersavecontrol;
                report[OutputReportOffsetAudioControl2] = audiocontrol2;
                return report;
            }

            static Result<Output> from_bytes(const std::uint8_t *data, std::size_t len)
            {
                if (len < OUTPUT_SIZE)
                {
                    return Error("buffer too short");
                }
                auto base = dualsense::Output::from_bytes(data, len);
                if (base.is_error())
                {
                    return base.error();
                }

                Output result;
                static_cast<dualsense::Output &>(result) = base.value();
                std::size_t offset = dualsense::OUTPUT_SIZE;
                result.speakervolume = data[offset++];
                result.microphonevolume = data[offset++];
                result.audiocontrol = data[offset++];
                result.micmuteled = data[offset++];
                result.powersavecontrol = data[offset++];
                result.audiocontrol2 = data[offset++];
                return result;
            }
        };

        struct CreateOptions
        {
            std::optional<MetaState> Meta{};
            std::optional<std::uint16_t> AudioPort{};

            [[nodiscard]] json_type to_map() const
            {
                json_type j;
                if (Meta.has_value())
                {
                    j["Meta"] = Meta->to_map();
                }
                if (AudioPort.has_value())
                {
                    j["audio_port"] = AudioPort.value();
                }
                return j;
            }
        };

    } // namespace dualsense_full
} // namespace viiper
