cmake_minimum_required(VERSION 3.14)
project(viiper-cpp-examples LANGUAGES CXX)

# Require C++20 (SDK uses concepts)
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Strict warnings
if(MSVC)
    add_compile_options(/W4)
else()
    add_compile_options(-Wall -Wextra -Wpedantic)
endif()

# VIIPER SDK is header-only - just add include path
set(VIIPER_INCLUDE_DIR "${CMAKE_CURRENT_SOURCE_DIR}/../../clients/cpp/include")

# Fetch nlohmann_json via FetchContent
include(FetchContent)
FetchContent_Declare(json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.3
    GIT_SHALLOW TRUE
)
FetchContent_MakeAvailable(json)

# Create interface library for VIIPER includes
add_library(viiper_sdk INTERFACE)
target_include_directories(viiper_sdk INTERFACE ${VIIPER_INCLUDE_DIR})
target_link_libraries(viiper_sdk INTERFACE nlohmann_json::nlohmann_json)

# Example: virtual_keyboard
add_executable(virtual_keyboard virtual_keyboard.cpp)
target_link_libraries(virtual_keyboard PRIVATE viiper_sdk)

# Example: virtual_mouse
add_executable(virtual_mouse virtual_mouse.cpp)
target_link_libraries(virtual_mouse PRIVATE viiper_sdk)

# Example: virtual_x360_pad
add_executable(virtual_x360_pad virtual_x360_pad.cpp)
target_link_libraries(virtual_x360_pad PRIVATE viiper_sdk)

# Platform-specific libraries
if(WIN32)
    target_link_libraries(virtual_keyboard PRIVATE ws2_32)
    target_link_libraries(virtual_mouse PRIVATE ws2_32)
    target_link_libraries(virtual_x360_pad PRIVATE ws2_32)
else()
    find_package(Threads REQUIRED)
    target_link_libraries(virtual_keyboard PRIVATE Threads::Threads)
    target_link_libraries(virtual_mouse PRIVATE Threads::Threads)
    target_link_libraries(virtual_x360_pad PRIVATE Threads::Threads)
endif()
