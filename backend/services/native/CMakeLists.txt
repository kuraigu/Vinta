cmake_minimum_required(VERSION 3.20)
project(vintaservices LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Enable asynchronous exception handling on MSVC for C++/WinRT coroutines and COM
if(MSVC)
    add_compile_options(/EHa /O2 /W4 /wd4100 /wd4201 /wd4458 /wd4505 /wd4244)
endif()

add_library(vintaservices SHARED
    winrt_bridge.cpp
    winrt_bridge.h
    notification_service.cpp
    notification_service.h
    media_service.cpp
    media_service.h
    audio_peak_service.cpp
    audio_peak_service.h
    radio_service.cpp
    radio_service.h
    icon_service.cpp
    icon_service.h
    winrt_common.h
)

target_link_libraries(vintaservices PRIVATE
    WindowsApp.lib
    Ole32.lib
    OleAut32.lib
    RuntimeObject.lib
    Advapi32.lib
    User32.lib
    Shell32.lib
    Gdi32.lib
    Windowscodecs.lib
)

# Ensure DLL is copied to backend/services and build/bin on build completion
add_custom_command(TARGET vintaservices POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy_if_different
        $<TARGET_FILE:vintaservices>
        ${CMAKE_CURRENT_SOURCE_DIR}/../vintaservices.dll
    COMMAND ${CMAKE_COMMAND} -E make_directory
        ${CMAKE_CURRENT_SOURCE_DIR}/../../../build/bin
    COMMAND ${CMAKE_COMMAND} -E copy_if_different
        $<TARGET_FILE:vintaservices>
        ${CMAKE_CURRENT_SOURCE_DIR}/../../../build/bin/vintaservices.dll
    COMMENT "Copying vintaservices.dll to backend/services/ and build/bin/"
)
