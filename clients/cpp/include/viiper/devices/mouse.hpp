

#pragma once

#include "../error.hpp"
#include "../detail/json.hpp"
#include <cstdint>
#include <vector>

namespace viiper
{
    namespace mouse
    {

        // ============================================================================
        // Constants
        // ============================================================================

        constexpr std::uint64_t BtnLeft = 1;
        constexpr std::uint64_t BtnRight = 2;
        constexpr std::uint64_t BtnMiddle = 4;
        constexpr std::uint64_t BtnBack = 8;
        constexpr std::uint64_t BtnForward = 16;

        // ============================================================================
        // Input: Client -> Device
        // ============================================================================

        struct Input
        {
            std::uint8_t buttons = 0;
            std::int16_t dx = 0;
            std::int16_t dy = 0;
            std::int16_t wheel = 0;
            std::int16_t pan = 0;

            [[nodiscard]] std::vector<std::uint8_t> to_bytes() const
            {
                std::vector<std::uint8_t> buf;
                buf.push_back(buttons);
                buf.push_back(static_cast<std::uint8_t>(dx & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((dx >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(dy & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((dy >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(wheel & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((wheel >> 8) & 0xFF));
                buf.push_back(static_cast<std::uint8_t>(pan & 0xFF));
                buf.push_back(static_cast<std::uint8_t>((pan >> 8) & 0xFF));
                return buf;
            }
        };

    } // namespace mouse
} // namespace viiper
