cmake_minimum_required(VERSION 3.16)
project(UmKmInfo)

um_exe(UmKmInfo SUBSYSTEM CONSOLE main.c)

target_compile_definitions(UmKmInfo PRIVATE _CRT_SECURE_NO_WARNINGS UNICODE _UNICODE)
target_link_libraries(UmKmInfo ntdll.lib)
set_target_properties(UmKmInfo PROPERTIES LINK_FLAGS "/SUBSYSTEM:CONSOLE /MANIFESTUAC:\"level='requireAdministrator' uiAccess='false'\"")
