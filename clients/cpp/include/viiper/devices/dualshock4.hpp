

#pragma once

#include "../error.hpp"
#include "../detail/json.hpp"
#include <cstdint>
#include <vector>
#include <string_view>
#include <algorithm>
#include <optional>

namespace viiper
{
    namespace dualshock4
    {

        // ============================================================================
        // Constants
        // ============================================================================

        constexpr std::size_t OUTPUT_SIZE = 8;

        constexpr std::uint64_t DefaultVID = 0x054C;
        constexpr std::uint64_t DefaultPID = 0x09CC;
        constexpr std::uint64_t EndpointIn = 132;
        constexpr std::uint64_t EndpointOut = 3;
        constexpr std::uint64_t ReportIDInput = 1;
        constexpr std::uint64_t ReportIDOutput = 5;
        constexpr std::uint64_t UpdateFlagRumble = 1;
        constexpr std::uint64_t UpdateFlagLED = 2;
        constexpr std::uint64_t UpdateFlagLEDBlink = 4;
        constexpr std::uint64_t InputReportSize = 64;
        constexpr std::uint64_t ButtonSquare = 16;
        constexpr std::uint64_t ButtonCross = 32;
        constexpr std::uint64_t ButtonCircle = 64;
        constexpr std::uint64_t ButtonTriangle = 128;
        constexpr std::uint64_t DPadMask = 15;
        constexpr std::uint64_t ButtonL1 = 256;
        constexpr std::uint64_t ButtonR1 = 512;
        constexpr std::uint64_t ButtonL2 = 1024;
        constexpr std::uint64_t ButtonR2 = 2048;
        constexpr std::uint64_t ButtonShare = 4096;
        constexpr std::uint64_t ButtonOptions = 8192;
        constexpr std::uint64_t ButtonL3 = 16384;
        constexpr std::uint64_t ButtonR3 = 32768;
        constexpr std::uint64_t ButtonPS = 1;
        constexpr std::uint64_t ButtonTouchpadClick = 2;
        constexpr std::uint64_t ButtonPSUSB = 1;
        constexpr std::uint64_t ButtonTouchpadClickUSB = 2;
        constexpr std::uint64_t CounterShift = 2;
        constexpr std::uint64_t DPadUSBUp = 0;
        constexpr std::uint64_t DPadUSBUpRight = 1;
        constexpr std::uint64_t DPadUSBRight = 2;
        constexpr std::uint64_t DPadUSBDownRight = 3;
        constexpr std::uint64_t DPadUSBDown = 4;
        constexpr std::uint64_t DPadUSBDownLeft = 5;
        constexpr std::uint64_t DPadUSBLeft = 6;
        constexpr std::uint64_t DPadUSBUpLeft = 7;
        constexpr std::uint64_t DPadUSBNeutral = 8;
        constexpr std::uint64_t DPadUp = 1;
        constexpr std::uint64_t DPadDown = 2;
        constexpr std::uint64_t DPadLeft = 4;
        constexpr std::uint64_t DPadRight = 8;
        constexpr std::uint64_t DefaultAccelXRaw = 0;
        constexpr std::uint64_t DefaultAccelYRaw = 0;
        constexpr std::uint64_t DefaultAccelZRaw = -8192;
        constexpr std::uint64_t DefaultLedRed = 0;
        constexpr std::uint64_t DefaultLedGreen = 0;
        constexpr std::uint64_t DefaultLedBlue = 64;
        constexpr std::uint64_t TouchpadMinX = 0;
        constexpr std::uint64_t TouchpadMaxX = 1920;
        constexpr std::uint64_t TouchpadMinY = 0;
        constexpr std::uint64_t TouchpadMaxY = 942;
        constexpr std::uint64_t TouchInactiveMask = 128;
        constexpr std::uint64_t BatteryLevelMask = 15;
        constexpr std::uint64_t BatteryChargingFlag = 16;
        constexpr std::uint64_t BatteryFullyCharged = 10;
        constexpr std::uint64_t HardwareVersionMajor = 1;
        constexpr std::uint64_t HardwareVersionMinor = 46080;
        constexpr std::uint64_t SoftwareVersionMajor = 1;
        constexpr std::uint64_t SoftwareVersionMinor = 40971;

        // ============================================================================
        // Device-specific typed helpers
        // ============================================================================

        struct MetaState
        {
            std::string SerialNumber{};
            std::string Board{};
            std::string BuildTime{};
            std::uint8_t BatteryStatus{};
            double TemperatureCelsius{};
            double BatteryVoltage{};

            static MetaState from_map(const json_type &j)
            {
                MetaState result;
                result.SerialNumber = j.value("serial_number", std::string{});
                result.Board = j.value("board", std::string{});
                result.BuildTime = j.value("build_time", std::string{});
                result.BatteryStatus = j.value("battery_status", std::uint8_t{});
                result.TemperatureCelsius = j.value("temperature_celsius", double{});
                result.BatteryVoltage = j.value("battery_voltage", double{});
                return result;
            }

            [[nodiscard]] json_type to_map() const
            {
                json_type j;
                j["serial_number"] = SerialNumber;
                j["board"] = Board;
                j["build_time"] = BuildTime;
                j["battery_status"] = BatteryStatus;
                j["temperature_celsius"] = TemperatureCelsius;
                j["battery_voltage"] = BatteryVoltage;
                return j;
            }
        };

        // ============================================================================
        // Input: Client -> Device
        // ============================================================================

        struct Input
        {
            std::int8_t sticklx = 0;
            std::int8_t stickly = 0;
            std::int8_t stickrx = 0;
            std::int8_t stickry = 0;
            std::uint16_t buttons = 0;
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

        // ============================================================================
        // Output: Device -> Client
        // ============================================================================

        struct Output
        {
            std::uint8_t updateflags = 0;
            std::uint8_t rumblesmall = 0;
            std::uint8_t rumblelarge = 0;
            std::uint8_t ledred = 0;
            std::uint8_t ledgreen = 0;
            std::uint8_t ledblue = 0;
            std::uint8_t flashon = 0;
            std::uint8_t flashoff = 0;

            static Result<Output> from_bytes(const std::uint8_t *data, std::size_t len)
            {
                Output result;
                std::size_t offset = 0;
                if (offset >= len)
                    return Error("buffer too short");
                result.updateflags = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.rumblesmall = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.rumblelarge = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.ledred = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.ledgreen = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.ledblue = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.flashon = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.flashoff = data[offset++];
                (void)offset; // suppress unused warning
                return result;
            }
        };

    } // namespace dualshock4
} // namespace viiper
