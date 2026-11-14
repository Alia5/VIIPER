cmake_minimum_required(VERSION 3.20)
project(viiper-libviiper-dualsense-passthrough LANGUAGES C CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

include(FetchContent)
FetchContent_Declare(SDL3
    GIT_REPOSITORY https://github.com/libsdl-org/SDL.git
    GIT_TAG main
    GIT_SHALLOW TRUE
)
set(SDL_TEST_LIBRARY OFF CACHE BOOL "" FORCE)
FetchContent_MakeAvailable(SDL3)

set(LIB_DIR "${CMAKE_CURRENT_LIST_DIR}/../../../../dist/libVIIPER")

add_library(libVIIPER SHARED IMPORTED GLOBAL)
set_target_properties(libVIIPER PROPERTIES INTERFACE_INCLUDE_DIRECTORIES "${LIB_DIR}")

if(WIN32)
    if(MSVC)
        set(IMPLIB "${CMAKE_CURRENT_BINARY_DIR}/libVIIPER.lib")
        add_custom_command(OUTPUT "${IMPLIB}"
            COMMAND lib.exe /def:"${LIB_DIR}/libVIIPER.def" /out:"${IMPLIB}" /machine:x64
            WORKING_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}"
            DEPENDS "${LIB_DIR}/libVIIPER.def" "${LIB_DIR}/libVIIPER.dll")
    else()
        set(IMPLIB "${CMAKE_CURRENT_BINARY_DIR}/libVIIPER.dll.a")
        add_custom_command(OUTPUT "${IMPLIB}"
            COMMAND dlltool -d "${LIB_DIR}/libVIIPER.def" -l "${IMPLIB}"
            WORKING_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}"
            DEPENDS "${LIB_DIR}/libVIIPER.def" "${LIB_DIR}/libVIIPER.dll")
    endif()
    add_custom_target(libVIIPER_implib DEPENDS "${IMPLIB}")
    set_target_properties(libVIIPER PROPERTIES
        IMPORTED_LOCATION "${LIB_DIR}/libVIIPER.dll"
        IMPORTED_IMPLIB   "${IMPLIB}")
else()
    set_target_properties(libVIIPER PROPERTIES
        IMPORTED_LOCATION "${LIB_DIR}/libVIIPER.so")
endif()

add_executable(dualsense_passthrough main.cpp)
target_link_libraries(dualsense_passthrough PRIVATE libVIIPER SDL3::SDL3)

if(MSVC)
    target_compile_options(dualsense_passthrough PRIVATE /W4)
else()
    target_compile_options(dualsense_passthrough PRIVATE -Wall -Wextra -Wpedantic)
endif()

if(WIN32)
    add_dependencies(dualsense_passthrough libVIIPER_implib)
    target_link_libraries(dualsense_passthrough PRIVATE ws2_32)
    add_custom_command(TARGET dualsense_passthrough POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy_if_different
            "${LIB_DIR}/libVIIPER.dll"
            "$<TARGET_FILE:SDL3::SDL3>"
            "$<TARGET_FILE_DIR:dualsense_passthrough>")
else()
    find_package(Threads REQUIRED)
    target_link_libraries(dualsense_passthrough PRIVATE Threads::Threads)
endif()
