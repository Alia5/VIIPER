

#pragma once

#include "../error.hpp"
#include "../detail/json.hpp"
#include <array>
#include <cstddef>
#include <cstdint>
#include <vector>
#include <string_view>
#include <algorithm>
#include <optional>

namespace viiper
{
    namespace dualsense
    {
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

        constexpr std::uint64_t DefaultVID = 1356;
        constexpr std::uint64_t DefaultPIDDSEdge = 3570;
        constexpr std::uint64_t DefaultPIDDS = 3302;
        constexpr std::uint64_t EndpointIn = 132;
        constexpr std::uint64_t EndpointOut = 3;
        constexpr std::uint64_t ReportIDInput = 1;
        constexpr std::uint64_t ReportIDOutput = 2;
        constexpr std::uint64_t InputReportSize = 64;
        constexpr std::uint64_t OutputReportSize = 64;
        constexpr std::uint64_t InputStateSize = 33;
        constexpr std::uint64_t OutputStateSize = 31;
        constexpr std::uint64_t UpdateFlag1Rumble = 1;
        constexpr std::uint64_t UpdateFlag1HapticsSelect = 2;
        constexpr std::uint64_t UpdateFlag1TriggerR2 = 4;
        constexpr std::uint64_t UpdateFlag1TriggerL2 = 8;
        constexpr std::uint64_t UpdateFlag2Led = 4;
        constexpr std::uint64_t UpdateFlag2PlayerLeds = 16;
        constexpr std::uint64_t UpdateFlag3ImprovedRumble = 4;
        constexpr std::uint64_t ButtonSquare = 16;
        constexpr std::uint64_t ButtonCross = 32;
        constexpr std::uint64_t ButtonCircle = 64;
        constexpr std::uint64_t ButtonTriangle = 128;
        constexpr std::uint64_t ButtonL1 = 256;
        constexpr std::uint64_t ButtonR1 = 512;
        constexpr std::uint64_t ButtonL2 = 1024;
        constexpr std::uint64_t ButtonR2 = 2048;
        constexpr std::uint64_t ButtonCreate = 4096;
        constexpr std::uint64_t ButtonOptions = 8192;
        constexpr std::uint64_t ButtonL3 = 16384;
        constexpr std::uint64_t ButtonR3 = 32768;
        constexpr std::uint64_t ButtonPS = 65536;
        constexpr std::uint64_t ButtonTouchpad = 131072;
        constexpr std::uint64_t ButtonMicMute = 262144;
        constexpr std::uint64_t ButtonEdgeLFn = 1048576;
        constexpr std::uint64_t ButtonEdgeRFn = 2097152;
        constexpr std::uint64_t ButtonEdgeL4 = 4194304;
        constexpr std::uint64_t ButtonEdgeR4 = 8388608;
        constexpr std::uint64_t DPadUp = 1;
        constexpr std::uint64_t DPadDown = 2;
        constexpr std::uint64_t DPadLeft = 4;
        constexpr std::uint64_t DPadRight = 8;
        constexpr std::uint64_t DPadUSBUp = 0;
        constexpr std::uint64_t DPadUSBUpRight = 1;
        constexpr std::uint64_t DPadUSBRight = 2;
        constexpr std::uint64_t DPadUSBDownRight = 3;
        constexpr std::uint64_t DPadUSBDown = 4;
        constexpr std::uint64_t DPadUSBDownLeft = 5;
        constexpr std::uint64_t DPadUSBLeft = 6;
        constexpr std::uint64_t DPadUSBUpLeft = 7;
        constexpr std::uint64_t DPadUSBNeutral = 8;
        constexpr std::uint64_t DPadMask = 15;
        constexpr std::uint64_t DefaultAccelXRaw = 0;
        constexpr std::uint64_t DefaultAccelYRaw = 0;
        constexpr std::uint64_t DefaultAccelZRaw = -8192;
        constexpr std::uint64_t TouchpadMinX = 0;
        constexpr std::uint64_t TouchpadMinY = 0;
        constexpr std::uint64_t TouchpadMaxX = 1920;
        constexpr std::uint64_t TouchpadMaxY = 1080;
        constexpr std::uint64_t TouchInactiveMask = 128;
        constexpr std::uint64_t DeltaTimeNS = 333;
        constexpr std::uint64_t BatteryFullyCharged = 42;
        constexpr std::uint64_t ConnectStateUSB = 8;
        constexpr std::uint64_t TriggerStatusNeutral = 9;
        constexpr std::uint64_t EdgeActiveProfile = 16;
        constexpr std::uint64_t HardwareType = 3;
        constexpr std::uint64_t HwInfo = 16777736;
        constexpr std::uint64_t FirmwareVersion = 1584;

        struct MetaState
        {
            std::string SerialNumber{};
            std::string MACAddress{};
            std::string Board{};
            std::string BuildTime{};
            std::uint8_t BatteryStatus{};
            double TemperatureCelsius{};
            double BatteryVoltage{};
            std::string ShellColor{};
            bool HeadphonesConnected{};
            bool MicrophoneConnected{};

            static MetaState from_map(const json_type &j)
            {
                MetaState result;
                result.SerialNumber = j.value("serial_number", std::string{});
                result.MACAddress = j.value("mac_address", std::string{});
                result.Board = j.value("board", std::string{});
                result.BuildTime = j.value("build_time", std::string{});
                result.BatteryStatus = j.value("battery_status", std::uint8_t{});
                result.TemperatureCelsius = j.value("temperature_celsius", double{});
                result.BatteryVoltage = j.value("battery_voltage", double{});
                result.ShellColor = j.value("shell_color", std::string{});
                result.HeadphonesConnected = j.value("headphones_connected", bool{});
                result.MicrophoneConnected = j.value("microphone_connected", bool{});
                return result;
            }

            [[nodiscard]] json_type to_map() const
            {
                json_type j;
                j["serial_number"] = SerialNumber;
                j["mac_address"] = MACAddress;
                j["board"] = Board;
                j["build_time"] = BuildTime;
                j["battery_status"] = BatteryStatus;
                j["temperature_celsius"] = TemperatureCelsius;
                j["battery_voltage"] = BatteryVoltage;
                j["shell_color"] = ShellColor;
                j["headphones_connected"] = HeadphonesConnected;
                j["microphone_connected"] = MicrophoneConnected;
                return j;
            }
        };
        struct Input
        {
            std::int8_t sticklx = 0;
            std::int8_t stickly = 0;
            std::int8_t stickrx = 0;
            std::int8_t stickry = 0;
            std::uint32_t buttons = 0;
            std::uint8_t dpad = 0;
            std::uint8_t triggerl2 = 0;
            std::uint8_t triggerr2 = 0;
            std::uint16_t touch1x = 0;
            std::uint16_t touch1y = 0;
            bool touch1active = 0;
            std::uint16_t touch2x = 0;
            std::uint16_t touch2y = 0;
            bool touch2active = 0;
            std::int16_t gyrox = 0;
            std::int16_t gyroy = 0;
            std::int16_t gyroz = 0;
            std::int16_t accelx = 0;
            std::int16_t accely = 0;
            std::int16_t accelz = 0;

            [[nodiscard]] std::vector<std::uint8_t> to_bytes() const
            {
                std::vector<std::uint8_t> buf;
                buf.push_back(static_cast<std::uint8_t>(sticklx));
                buf.push_back(static_cast<std::uint8_t>(stickly));
                buf.push_back(static_cast<std::uint8_t>(stickrx));
                buf.push_back(static_cast<std::uint8_t>(stickry));
                buf.push_back(static_cast<std::uint8_t>(buttons & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 16) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 24) & 0xFF));
                buf.push_back(dpad);
                buf.push_back(triggerl2);
                buf.push_back(triggerr2);
                buf.push_back(static_cast<std::uint8_t>(touch1x & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((touch1x >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(touch1y & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((touch1y >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(touch2x & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((touch2x >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(touch2y & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((touch2y >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(gyrox & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((gyrox >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(gyroy & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((gyroy >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(gyroz & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((gyroz >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(accelx & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((accelx >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(accely & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((accely >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(accelz & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((accelz >> 8) & 0xFF));
                return buf;
            }
        };

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

        struct Output
        {
            std::uint8_t updateflags1 = 0;
            std::uint8_t updateflags2 = 0;
            std::uint8_t updateflags3 = 0;
            std::uint8_t rumblesmall = 0;
            std::uint8_t rumblelarge = 0;
            std::uint8_t ledred = 0;
            std::uint8_t ledgreen = 0;
            std::uint8_t ledblue = 0;
            std::uint8_t playerleds = 0;
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
                Output result;
                std::size_t offset = 0;
                result.updateflags1 = data[offset++];
                result.updateflags2 = data[offset++];
                result.updateflags3 = data[offset++];
                result.rumblesmall = data[offset++];
                result.rumblelarge = data[offset++];
                result.ledred = data[offset++];
                result.ledgreen = data[offset++];
                result.ledblue = data[offset++];
                result.playerleds = data[offset++];
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

    } // namespace dualsense
} // namespace viiper
