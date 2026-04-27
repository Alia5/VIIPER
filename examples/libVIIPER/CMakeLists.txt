cmake_minimum_required(VERSION 3.20)
project(libVIIPER_example C)

set(LIB_DIR "${CMAKE_CURRENT_LIST_DIR}/../../dist/libVIIPER")

add_library(libVIIPER SHARED IMPORTED GLOBAL)
set_target_properties(libVIIPER PROPERTIES INTERFACE_INCLUDE_DIRECTORIES "${LIB_DIR}")

if(WIN32)
    if(MSVC)
        set(IMPLIB "${CMAKE_CURRENT_BINARY_DIR}/libVIIPER.lib")
        add_custom_command(OUTPUT "${IMPLIB}"
            COMMAND lib.exe /def:"${LIB_DIR}/libVIIPER.def" /out:"${IMPLIB}" /machine:x64
            WORKING_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}" DEPENDS "${LIB_DIR}/libVIIPER.def" "${LIB_DIR}/libVIIPER.dll")
    else()
        set(IMPLIB "${CMAKE_CURRENT_BINARY_DIR}/libVIIPER.dll.a")
        add_custom_command(OUTPUT "${IMPLIB}"
            COMMAND dlltool -d "${LIB_DIR}/libVIIPER.def" -l "${IMPLIB}"
            WORKING_DIRECTORY "${CMAKE_CURRENT_BINARY_DIR}" DEPENDS "${LIB_DIR}/libVIIPER.def" "${LIB_DIR}/libVIIPER.dll")
    endif()
    add_custom_target(libVIIPER_implib DEPENDS "${IMPLIB}")
    set_target_properties(libVIIPER PROPERTIES IMPORTED_LOCATION "${LIB_DIR}/libVIIPER.dll" IMPORTED_IMPLIB "${IMPLIB}")
else()
    set_target_properties(libVIIPER PROPERTIES IMPORTED_LOCATION "${LIB_DIR}/libVIIPER.so")
endif()

add_executable(libVIIPER_example main.c)
target_link_libraries(libVIIPER_example PRIVATE libVIIPER)

if(WIN32)
    add_dependencies(libVIIPER_example libVIIPER_implib)
    add_custom_command(TARGET libVIIPER_example POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy_if_different "${LIB_DIR}/libVIIPER.dll" "$<TARGET_FILE_DIR:libVIIPER_example>")
endif()
