cmake_minimum_required(VERSION 3.15)
project(virtual_x360_pad C)

set(CMAKE_C_STANDARD 99)

# Paths to the generated C SDK
set(VIIPER_SDK_ROOT "${CMAKE_CURRENT_LIST_DIR}/../../../clients/c")
set(VIIPER_INCLUDE_DIR "${VIIPER_SDK_ROOT}/include")
set(VIIPER_LIB_DIR "${VIIPER_SDK_ROOT}/build/$<CONFIG>")

add_executable(virtual_x360_pad main.c)

target_include_directories(virtual_x360_pad PRIVATE "${VIIPER_INCLUDE_DIR}")

target_link_libraries(virtual_x360_pad PRIVATE "${VIIPER_LIB_DIR}/viiper.lib")

add_custom_command(TARGET virtual_x360_pad POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy_if_different
            "${VIIPER_LIB_DIR}/viiper.dll"
            "$<TARGET_FILE_DIR:virtual_x360_pad>"
)
