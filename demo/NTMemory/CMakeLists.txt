cmake_minimum_required(VERSION 3.16)

include("C:/Program Files/CMake/bin/ewdk.cmake")

project(NTMemory)

set(IMGUI_DIR ${CMAKE_CURRENT_SOURCE_DIR}/Imgui)

um_exe(NTMemory
    SUBSYSTEM WINDOWS
    SOURCES
        main.cpp
        core.c
        ${IMGUI_DIR}/imgui.cpp
        ${IMGUI_DIR}/imgui_demo.cpp
        ${IMGUI_DIR}/imgui_draw.cpp
        ${IMGUI_DIR}/imgui_impl_dx11.cpp
        ${IMGUI_DIR}/imgui_impl_win32.cpp
        ${IMGUI_DIR}/imgui_tables.cpp
        ${IMGUI_DIR}/imgui_widgets.cpp
    INCLUDES ${IMGUI_DIR}
    DEFINES _CRT_SECURE_NO_WARNINGS UNICODE _UNICODE
    NOAUTO
    LIBS d3d11.lib dxgi.lib ntdll.lib psapi.lib advapi32.lib
)

set_target_properties(NTMemory PROPERTIES LINK_FLAGS "/SUBSYSTEM:WINDOWS /MANIFESTUAC:\"level='requireAdministrator' uiAccess='false'\"")
