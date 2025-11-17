cmake_minimum_required(VERSION 3.15)
project(virtual_keyboard C)

# Include generated SDK headers
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../../clients/c/include)

add_executable(virtual_keyboard main.c)

# Link against generated viiper library
if (WIN32)
  set(VIIPER_LIB ${CMAKE_CURRENT_SOURCE_DIR}/../../../clients/c/build/Release/viiper.lib)
else()
  set(VIIPER_LIB ${CMAKE_CURRENT_SOURCE_DIR}/../../../clients/c/build/libviiper.a)
endif()

target_link_libraries(virtual_keyboard PRIVATE ${VIIPER_LIB})

# Copy runtime DLL next to executable on Windows so running from build dir works
if (WIN32)
  add_custom_command(TARGET virtual_keyboard POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy_if_different
      "${VIIPER_LIB}" "$<TARGET_FILE_DIR:virtual_keyboard>"
    COMMAND ${CMAKE_COMMAND} -E copy_if_different
      "${CMAKE_CURRENT_SOURCE_DIR}/../../../clients/c/build/Release/viiper.dll" "$<TARGET_FILE_DIR:virtual_keyboard>"
  )
endif()
