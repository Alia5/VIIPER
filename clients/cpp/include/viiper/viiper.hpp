

#pragma once

// ============================================================================
// VIIPER C++ Client Library - Header-Only Library
// ============================================================================
//
// Version: 0.9.0
//
// Before including this file, define your JSON library configuration:
//   #define VIIPER_JSON_INCLUDE <your/json/library.hpp>
//   #define VIIPER_JSON_NAMESPACE your_namespace
//   #define VIIPER_JSON_TYPE your_json_type
//
// See config.hpp for JSON library requirements.
//
// Example with nlohmann::json:
//   #define VIIPER_JSON_INCLUDE <nlohmann/json.hpp>
//   #define VIIPER_JSON_NAMESPACE nlohmann
//   #define VIIPER_JSON_TYPE json
//   #include <viiper/viiper.hpp>
//
// ============================================================================

#include "config.hpp"
#include "error.hpp"
#include "types.hpp"
#include "client.hpp"
#include "device.hpp"

// Device-specific headers

#include "devices/xbox360.hpp"
#include "devices/dualsense.hpp"
#include "devices/dualsense_full.hpp"
#include "devices/dualshock4.hpp"
#include "devices/keyboard.hpp"
#include "devices/mouse.hpp"
#include "devices/ns2pro.hpp"

namespace viiper
{

    // Version information
    constexpr int VERSION_MAJOR = 0;
    constexpr int VERSION_MINOR = 9;
    constexpr int VERSION_PATCH = 0;

    inline std::string version()
    {
        return std::to_string(VERSION_MAJOR) + "." +
               std::to_string(VERSION_MINOR) + "." +
               std::to_string(VERSION_PATCH);
    }

} // namespace viiper
