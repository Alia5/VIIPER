

#pragma once

#include "../error.hpp"
#include <cstdint>
#include <vector>

namespace viiper
{
    namespace xbone
    {

        // ============================================================================
        // Constants
        // ============================================================================

        constexpr std::size_t OUTPUT_SIZE = 6;

        constexpr std::uint64_t IDVendor = 1118;
        constexpr std::uint64_t IDProduct = 721;
        constexpr std::uint64_t IDProductS = 746;
        constexpr std::uint64_t ButtonMenu = 4;
        constexpr std::uint64_t ButtonView = 8;
        constexpr std::uint64_t ButtonA = 16;
        constexpr std::uint64_t ButtonB = 32;
        constexpr std::uint64_t ButtonX = 64;
        constexpr std::uint64_t ButtonY = 128;
        constexpr std::uint64_t ButtonDPadUp = 256;
        constexpr std::uint64_t ButtonDPadDown = 512;
        constexpr std::uint64_t ButtonDPadLeft = 1024;
        constexpr std::uint64_t ButtonDPadRight = 2048;
        constexpr std::uint64_t ButtonLShoulder = 4096;
        constexpr std::uint64_t ButtonRShoulder = 8192;
        constexpr std::uint64_t ButtonLThumb = 16384;
        constexpr std::uint64_t ButtonRThumb = 32768;
        constexpr std::uint64_t ButtonGuide = 65536;
        constexpr std::uint64_t GIP_CMD_ACKNOWLEDGE = 1;
        constexpr std::uint64_t GIP_CMD_ANNOUNCE = 2;
        constexpr std::uint64_t GIP_CMD_STATUS = 3;
        constexpr std::uint64_t GIP_CMD_IDENTIFY = 4;
        constexpr std::uint64_t GIP_CMD_POWER = 5;
        constexpr std::uint64_t GIP_CMD_AUTHENTICATE = 6;
        constexpr std::uint64_t GIP_CMD_VIRTUAL_KEY = 7;
        constexpr std::uint64_t GIP_CMD_RUMBLE = 9;
        constexpr std::uint64_t GIP_CMD_LED = 10;
        constexpr std::uint64_t GIP_CMD_INPUT = 32;
        constexpr std::uint64_t GIP_OPT_ACKNOWLEDGE = 16;
        constexpr std::uint64_t GIP_OPT_INTERNAL = 32;

        // ============================================================================
        // Input: Client -> Device
        // ============================================================================

        struct Input
        {
            std::uint32_t buttons = 0;
            std::uint16_t lt = 0;
            std::uint16_t rt = 0;
            std::int16_t lx = 0;
            std::int16_t ly = 0;
            std::int16_t rx = 0;
            std::int16_t ry = 0;

            [[nodiscard]] std::vector<std::uint8_t> to_bytes() const
            {
                std::vector<std::uint8_t> buf;
                buf.push_back(static_cast<std::uint8_t>(buttons & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 16) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 24) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(lt & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((lt >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(rt & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((rt >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(lx & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((lx >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(ly & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((ly >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(rx & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((rx >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(ry & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((ry >> 8) & 0xFF));
                return buf;
            }
        };

        // ============================================================================
        // Output: Device -> Client
        // ============================================================================

        struct Output
        {
            std::uint8_t left = 0;
            std::uint8_t right = 0;
            std::uint8_t lefttrigger = 0;
            std::uint8_t righttrigger = 0;
            std::uint8_t duration = 0;
            std::uint8_t delay = 0;

            static Result<Output> from_bytes(const std::uint8_t *data, std::size_t len)
            {
                Output result;
                std::size_t offset = 0;
                if (offset >= len)
                    return Error("buffer too short");
                result.left = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.right = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.lefttrigger = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.righttrigger = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.duration = data[offset++];
                if (offset >= len)
                    return Error("buffer too short");
                result.delay = data[offset++];
                (void)offset; // suppress unused warning
                return result;
            }
        };

    } // namespace xbone
} // namespace viiper
