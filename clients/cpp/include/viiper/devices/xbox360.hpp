

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
    namespace xbox360
    {

        // ============================================================================
        // Constants
        // ============================================================================

        constexpr std::size_t OUTPUT_SIZE = 2;

        constexpr std::uint64_t ButtonDPadUp = 1;
        constexpr std::uint64_t ButtonDPadDown = 2;
        constexpr std::uint64_t ButtonDPadLeft = 4;
        constexpr std::uint64_t ButtonDPadRight = 8;
        constexpr std::uint64_t ButtonStart = 16;
        constexpr std::uint64_t ButtonBack = 32;
        constexpr std::uint64_t ButtonLThumb = 64;
        constexpr std::uint64_t ButtonRThumb = 128;
        constexpr std::uint64_t ButtonLShoulder = 256;
        constexpr std::uint64_t ButtonRShoulder = 512;
        constexpr std::uint64_t ButtonGuide = 1024;
        constexpr std::uint64_t ButtonA = 4096;
        constexpr std::uint64_t ButtonB = 8192;
        constexpr std::uint64_t ButtonX = 16384;
        constexpr std::uint64_t ButtonY = 32768;

        // ============================================================================
        // Device-specific typed helpers
        // ============================================================================

        struct Xbox360CreateOptions
        {
            std::optional<std::uint8_t> SubType{};

            static Xbox360CreateOptions from_map(const json_type &j)
            {
                Xbox360CreateOptions result;
                result.SubType = detail::get_optional_field<std::uint8_t>(j, "subType");
                return result;
            }

            [[nodiscard]] json_type to_map() const
            {
                json_type j;
                if (SubType.has_value())
                {
                    j["subType"] = SubType.value();
                }
                return j;
            }
        };

        // ============================================================================
        // Input: Client -> Device
        // ============================================================================

        struct Input
        {
            std::uint32_t buttons = 0;
            std::uint8_t lt = 0;
            std::uint8_t rt = 0;
            std::int16_t lx = 0;
            std::int16_t ly = 0;
            std::int16_t rx = 0;
            std::int16_t ry = 0;
            std::array<std::uint8_t, 6> reserved{};

            [[nodiscard]] std::vector<std::uint8_t> to_bytes() const
            {
                std::vector<std::uint8_t> buf;
                buf.push_back(static_cast<std::uint8_t>(buttons & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 16) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((buttons >> 24) & 0xFF));
                buf.push_back(lt);
                buf.push_back(rt);
                buf.push_back(static_cast<std::uint8_t>(lx & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((lx >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(ly & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((ly >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(rx & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((rx >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(ry & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((ry >> 8) & 0xFF));
                for (std::size_t i = 0; i < static_cast<std::size_t>(6); i++)
                {
                    const auto v = reserved[i];
                    buf.push_back(static_cast<std::uint8_t>(v));
                }
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
                (void)offset; // suppress unused warning
                return result;
            }
        };

    } // namespace xbox360
} // namespace viiper
