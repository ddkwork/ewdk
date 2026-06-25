cmake_minimum_required(VERSION 3.16)

include("C:/Program Files/CMake/bin/ewdk.cmake")

project(UmKmInfo)

um_exe(UmKmInfo
    SUBSYSTEM CONSOLE
    SOURCES main.c
    DEFINES _CRT_SECURE_NO_WARNINGS UNICODE _UNICODE
    LIBS ntdll.lib
)

set_target_properties(UmKmInfo PROPERTIES LINK_FLAGS "/SUBSYSTEM:CONSOLE /MANIFESTUAC:\"level='requireAdministrator' uiAccess='false'\"")
