#pragma once

#include "../error.hpp"
#include "dualsense.hpp"
#include <array>
#include <cstddef>
#include <cstdint>
#include <string_view>

namespace viiper
{
    namespace dualsense_full
    {

        constexpr std::string_view DEVICE_TYPE = "dualsense_full";
        constexpr std::string_view DEVICE_TYPE_EDGE = "dualsenseedge_full";

        constexpr std::size_t OUTPUT_SIZE = 31;
        constexpr std::size_t TriggerParamsSize = 10;
        constexpr std::size_t TriggerZoneCount = 10;

        constexpr std::uint8_t TriggerModeSimpleFeedback = 0x01;
        constexpr std::uint8_t TriggerModeSimpleWeapon = 0x02;
        constexpr std::uint8_t TriggerModeOff = 0x05;
        constexpr std::uint8_t TriggerModeSimpleVibration = 0x06;
        constexpr std::uint8_t TriggerModeLimitedFeedback = 0x11;
        constexpr std::uint8_t TriggerModeLimitedWeapon = 0x12;
        constexpr std::uint8_t TriggerModeFeedback = 0x21;
        constexpr std::uint8_t TriggerModeBow = 0x22;
        constexpr std::uint8_t TriggerModeGalloping = 0x23;
        constexpr std::uint8_t TriggerModeWeapon = 0x25;
        constexpr std::uint8_t TriggerModeVibration = 0x26;
        constexpr std::uint8_t TriggerModeMachine = 0x27;
        constexpr std::uint8_t TriggerModeCalibration = 0xFC;
        constexpr std::uint8_t TriggerModeDebugFD = 0xFD;
        constexpr std::uint8_t TriggerModeDebugFE = 0xFE;

        using Input = dualsense::Input;
        using MetaState = dualsense::MetaState;

        struct TriggerEffect
        {
            std::uint8_t mode = 0;
            std::array<std::uint8_t, TriggerParamsSize> params{};
            std::array<std::uint8_t, TriggerZoneCount> zone_strengths{};
            std::uint8_t start_zone = 0;
            std::uint8_t end_zone = 0;
            std::uint8_t start_position = 0;
            std::uint8_t end_position = 0;
            std::uint8_t strength = 0;
            std::uint8_t snap_force = 0;
            std::uint8_t amplitude_a = 0;
            std::uint8_t amplitude_b = 0;
            std::uint8_t first_foot = 0;
            std::uint8_t second_foot = 0;
            std::uint8_t frequency = 0;
            std::uint8_t period = 0;

            [[nodiscard]] static TriggerEffect decode(std::uint8_t mode, const std::array<std::uint8_t, TriggerParamsSize> &params)
            {
                constexpr std::size_t bits_per_strength = 3;
                constexpr std::uint8_t strength_mask = 0x07;
                constexpr std::size_t frequency_param = 8;

                TriggerEffect effect;
                effect.mode = mode;
                effect.params = params;
                const auto zones = static_cast<std::uint16_t>(params[0] | params[1] << 8);

                switch (mode)
                {
                case TriggerModeFeedback:
                case TriggerModeVibration:
                {
                    const std::uint32_t packed_strengths =
                        static_cast<std::uint32_t>(params[2]) |
                        static_cast<std::uint32_t>(params[3]) << 8 |
                        static_cast<std::uint32_t>(params[4]) << 16 |
                        static_cast<std::uint32_t>(params[5]) << 24;
                    for (std::size_t zone = 0; zone < TriggerZoneCount; zone++)
                    {
                        if (zones & (1u << zone))
                        {
                            effect.zone_strengths[zone] = static_cast<std::uint8_t>(((packed_strengths >> (bits_per_strength * zone)) & strength_mask) + 1);
                        }
                    }
                    if (mode == TriggerModeVibration)
                    {
                        effect.frequency = params[frequency_param];
                    }
                    break;
                }
                case TriggerModeWeapon:
                case TriggerModeBow:
                case TriggerModeGalloping:
                case TriggerModeMachine:
                {
                    for (std::size_t zone = TriggerZoneCount; zone-- > 0;)
                    {
                        if (zones & (1u << zone))
                        {
                            effect.start_zone = static_cast<std::uint8_t>(zone);
                        }
                    }
                    for (std::size_t zone = 0; zone < TriggerZoneCount; zone++)
                    {
                        if (zones & (1u << zone))
                        {
                            effect.end_zone = static_cast<std::uint8_t>(zone);
                        }
                    }
                    const auto first = static_cast<std::uint8_t>(params[2] & strength_mask);
                    const auto second = static_cast<std::uint8_t>((params[2] >> bits_per_strength) & strength_mask);
                    if (mode == TriggerModeWeapon)
                    {
                        effect.strength = static_cast<std::uint8_t>(first + 1);
                    }
                    if (mode == TriggerModeBow)
                    {
                        effect.strength = static_cast<std::uint8_t>(first + 1);
                        effect.snap_force = static_cast<std::uint8_t>(second + 1);
                    }
                    if (mode == TriggerModeGalloping)
                    {
                        effect.second_foot = first;
                        effect.first_foot = second;
                        effect.frequency = params[3];
                    }
                    if (mode == TriggerModeMachine)
                    {
                        effect.amplitude_a = first;
                        effect.amplitude_b = second;
                        effect.frequency = params[3];
                        effect.period = params[4];
                    }
                    break;
                }
                case TriggerModeSimpleFeedback:
                case TriggerModeLimitedFeedback:
                    effect.start_position = params[0];
                    effect.strength = params[1];
                    break;
                case TriggerModeSimpleWeapon:
                case TriggerModeLimitedWeapon:
                    effect.start_position = params[0];
                    effect.end_position = params[1];
                    effect.strength = params[2];
                    break;
                case TriggerModeSimpleVibration:
                    effect.frequency = params[0];
                    effect.strength = params[1];
                    effect.start_position = params[2];
                    break;
                default:
                    break;
                }
                return effect;
            }
        };

        struct Output : dualsense::Output
        {
            std::uint8_t trigger_r2_mode = 0;
            std::array<std::uint8_t, TriggerParamsSize> trigger_r2_params{};
            std::uint8_t trigger_l2_mode = 0;
            std::array<std::uint8_t, TriggerParamsSize> trigger_l2_params{};

            [[nodiscard]] TriggerEffect trigger_r2() const
            {
                return TriggerEffect::decode(trigger_r2_mode, trigger_r2_params);
            }

            [[nodiscard]] TriggerEffect trigger_l2() const
            {
                return TriggerEffect::decode(trigger_l2_mode, trigger_l2_params);
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
                result.trigger_r2_mode = data[offset++];
                for (std::size_t i = 0; i < TriggerParamsSize; i++)
                {
                    result.trigger_r2_params[i] = data[offset++];
                }
                result.trigger_l2_mode = data[offset++];
                for (std::size_t i = 0; i < TriggerParamsSize; i++)
                {
                    result.trigger_l2_params[i] = data[offset++];
                }
                return result;
            }
        };

    } // namespace dualsense_full
} // namespace viiper
