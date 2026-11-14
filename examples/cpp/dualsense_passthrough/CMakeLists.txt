cmake_minimum_required(VERSION 3.14)
project(viiper-cpp-dualsense-passthrough LANGUAGES C CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

include(FetchContent)
FetchContent_Declare(json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.3
    GIT_SHALLOW TRUE
)
FetchContent_Declare(SDL3
    GIT_REPOSITORY https://github.com/libsdl-org/SDL.git
    GIT_TAG main
    GIT_SHALLOW TRUE
)
set(SDL_TEST_LIBRARY OFF CACHE BOOL "" FORCE)
FetchContent_MakeAvailable(json SDL3)

find_package(OpenSSL REQUIRED)

add_executable(dualsense_passthrough main.cpp)
target_include_directories(dualsense_passthrough PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/../../../clients/cpp/include")
target_link_libraries(dualsense_passthrough PRIVATE
    nlohmann_json::nlohmann_json
    OpenSSL::SSL
    OpenSSL::Crypto
    SDL3::SDL3
)

if(MSVC)
    target_compile_options(dualsense_passthrough PRIVATE /W4)
else()
    target_compile_options(dualsense_passthrough PRIVATE -Wall -Wextra -Wpedantic)
endif()

if(WIN32)
    target_link_libraries(dualsense_passthrough PRIVATE ws2_32)
    add_custom_command(TARGET dualsense_passthrough POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy_if_different
            "$<TARGET_FILE:SDL3::SDL3>"
            "$<TARGET_FILE_DIR:dualsense_passthrough>")
else()
    find_package(Threads REQUIRED)
    target_link_libraries(dualsense_passthrough PRIVATE Threads::Threads)
endif()
