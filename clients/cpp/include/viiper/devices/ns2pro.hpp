

#pragma once

#include "../error.hpp"
#include "../detail/json.hpp"
#include <cstdint>
#include <vector>
#include <array>
#include <string_view>
#include <algorithm>
#include <optional>

namespace viiper
{
    namespace ns2pro
    {

        // ============================================================================
        // Constants
        // ============================================================================

        constexpr std::size_t OUTPUT_SIZE = 34;

        constexpr std::uint64_t DefaultVID = 1406;
        constexpr std::uint64_t DefaultPID = 8297;
        constexpr std::uint64_t DefaultBatteryVolts = 3800;
        constexpr std::uint64_t ButtonB = 1;
        constexpr std::uint64_t ButtonA = 2;
        constexpr std::uint64_t ButtonY = 4;
        constexpr std::uint64_t ButtonX = 8;
        constexpr std::uint64_t ButtonR = 16;
        constexpr std::uint64_t ButtonZR = 32;
        constexpr std::uint64_t ButtonPlus = 64;
        constexpr std::uint64_t ButtonRightStick = 128;
        constexpr std::uint64_t ButtonDown = 256;
        constexpr std::uint64_t ButtonRight = 512;
        constexpr std::uint64_t ButtonLeft = 1024;
        constexpr std::uint64_t ButtonUp = 2048;
        constexpr std::uint64_t ButtonL = 4096;
        constexpr std::uint64_t ButtonZL = 8192;
        constexpr std::uint64_t ButtonMinus = 16384;
        constexpr std::uint64_t ButtonLeftStick = 32768;
        constexpr std::uint64_t ButtonHome = 65536;
        constexpr std::uint64_t ButtonCapture = 131072;
        constexpr std::uint64_t ButtonGR = 262144;
        constexpr std::uint64_t ButtonGL = 524288;
        constexpr std::uint64_t ButtonC = 1048576;
        constexpr std::uint64_t ButtonHeadset = 2097152;
        constexpr std::uint64_t OutputFlagRumble = 1;
        constexpr std::uint64_t OutputFlagLED = 2;
        constexpr std::uint64_t StickMin = 0;
        constexpr std::uint64_t StickCenter = 2048;
        constexpr std::uint64_t StickMax = 4095;
        constexpr std::uint64_t BatteryMax = 9;
        constexpr std::uint64_t EndpointHIDIn = 129;
        constexpr std::uint64_t EndpointHIDOut = 1;
        constexpr std::uint64_t EndpointBulkOut = 2;
        constexpr std::uint64_t EndpointBulkIn = 130;
        constexpr std::uint64_t ReportIDCommon = 5;
        constexpr std::uint64_t ReportIDPro = 9;
        constexpr std::uint64_t ReportIDOutput = 2;
        constexpr std::uint64_t InputReportSize = 64;
        constexpr std::uint64_t OutputReportSize = 64;
        constexpr std::uint64_t InputWireSize = 24;
        constexpr std::uint64_t OutputRumbleSize = 32;
        constexpr std::uint64_t OutputWireSize = 34;
        constexpr std::uint64_t FeatureButtons = 1;
        constexpr std::uint64_t FeatureSticks = 2;
        constexpr std::uint64_t FeatureIMU = 4;
        constexpr std::uint64_t FeatureMouse = 16;
        constexpr std::uint64_t FeatureRumble = 32;

        // ============================================================================
        // Device-specific typed helpers
        // ============================================================================

        struct MetaState
        {
            std::string SerialNumber{};
            std::uint8_t BatteryLevel{};
            bool Charging{};
            bool ExternalPower{};
            std::uint16_t BatteryVolts{};

            static MetaState from_map(const json_type &j)
            {
                MetaState result;
                result.SerialNumber = j.value("serial_number", std::string{});
                result.BatteryLevel = j.value("battery_level", std::uint8_t{});
                result.Charging = j.value("charging", bool{});
                result.ExternalPower = j.value("external_power", bool{});
                result.BatteryVolts = j.value("battery_volts", std::uint16_t{});
                return result;
            }

            [[nodiscard]] json_type to_map() const
            {
                json_type j;
                j["serial_number"] = SerialNumber;
                j["battery_level"] = BatteryLevel;
                j["charging"] = Charging;
                j["external_power"] = ExternalPower;
                j["battery_volts"] = BatteryVolts;
                return j;
            }
        };

        // ============================================================================
        // Input: Client -> Device
        // ============================================================================

        struct Input
        {
            std::uint32_t buttons = 0;
            std::uint16_t lx = 0;
            std::uint16_t ly = 0;
            std::uint16_t rx = 0;
            std::uint16_t ry = 0;
            std::int16_t accelx = 0;
            std::int16_t accely = 0;
            std::int16_t accelz = 0;
            std::int16_t gyrox = 0;
            std::int16_t gyroy = 0;
            std::int16_t gyroz = 0;

            [[nodiscard]] std::vector<std::uint8_t> to_bytes() const
            {
                std::vector<std::uint8_t> buf;
                buf.push_back(static_cast<std::uint8_t>(buttons & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 16) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 24) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(lx & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((lx >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(ly & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((ly >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(rx & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((rx >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(ry & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((ry >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(accelx & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((accelx >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(accely & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((accely >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(accelz & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((accelz >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(gyrox & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((gyrox >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(gyroy & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((gyroy >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(gyroz & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((gyroz >> 8) & 0xFF));
                return buf;
            }
        };

        // ============================================================================
        // Output: Device -> Client
        // ============================================================================

        struct Output
        {
            std::array<std::uint8_t, 16> leftrumble{};
            std::array<std::uint8_t, 16> rightrumble{};
            std::uint8_t flags = 0;
            std::uint8_t playerledmask = 0;

            static Result<Output> from_bytes(const std::uint8_t *data, std::size_t len)
            {
                Output result;
                std::size_t offset = 0;
                for (std::size_t i = 0; i < static_cast<std::size_t>(16); i++)
                {
                    if (offset >= len)
                        return Error("buffer too short");
                    result.leftrumble[i] = data[offset++];
                }
                for (std::size_t i = 0; i < static_cast<std::size_t>(16); i++)
                {
                    if (offset >= len)
                        return Error("buffer too short");
                    result.rightrumble[i] = data[offset++];
                }
                if (offset >= len)
                    return Error("buffer too short");
                result.flags = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.playerledmask = data[offset++];
                (void)offset; // suppress unused warning
                return result;
            }
        };

    } // namespace ns2pro
} // namespace viiper
