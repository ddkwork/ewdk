#.rst:
# FindWDK / EWDK CMake Module
# ----------
#
# This module searches for the installed Windows Development Kit (WDK) and
# exposes commands for creating kernel drivers and kernel libraries.
#
# It supports two modes:
#   1. ewdk-env.cmake injected (preferred): all paths from EWDK_* variables
#   2. Fallback: detect WDK via $ENV{WDKContentRoot} or system paths
#
# Output variables:
# - `WDK_FOUND` -- if false, do not try to use WDK
# - `WDK_ROOT` -- where WDK is installed
# - `WDK_VERSION` -- the version of the selected WDK
# - `WDK_WINVER` -- the WINVER used for kernel drivers and libraries
#        (default value is `0x0601` and can be changed per target or globally)
# - `WDK_NTDDI_VERSION` -- the NTDDI_VERSION used for kernel drivers and libraries,
#                          if not set, the value will be automatically calculated by WINVER
#        (default value is left blank and can be changed per target or globally)

if(DEFINED EWDK_WDKContentRoot)
    set(_WDK_ROOT "${EWDK_WDKContentRoot}")
    set(WDK_VERSION "${EWDK_WindowsTargetPlatformVersion}")
    set(WDK_INC_VERSION "${EWDK_WindowsTargetPlatformVersion}")
    set(WDK_LIB_VERSION "${EWDK_WindowsTargetPlatformVersion}")
    file(GLOB WDK_NTDDK_FILES
        "${_WDK_ROOT}/Include/${WDK_INC_VERSION}/km/ntddk.h"
    )
elseif(DEFINED ENV{WDKContentRoot})
    string(REGEX REPLACE "[\\/]$" "" _WDK_ROOT "$ENV{WDKContentRoot}")
    file(GLOB WDK_NTDDK_FILES
        "${_WDK_ROOT}/Include/*/km/ntddk.h"
        "${_WDK_ROOT}/Include/km/ntddk.h"
    )
else()
    file(GLOB WDK_NTDDK_FILES
        "C:/Program Files*/Windows Kits/*/Include/*/km/ntddk.h"
        "C:/Program Files*/Windows Kits/*/Include/km/ntddk.h"
    )
endif()

if(WDK_NTDDK_FILES)
    if (NOT CMAKE_VERSION VERSION_LESS 3.18.0)
        list(SORT WDK_NTDDK_FILES COMPARE NATURAL)
    endif()
    list(GET WDK_NTDDK_FILES -1 WDK_LATEST_NTDDK_FILE)
endif()

include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(WDK REQUIRED_VARS WDK_LATEST_NTDDK_FILE)

if(NOT WDK_LATEST_NTDDK_FILE)
    return()
endif()

if(NOT DEFINED WDK_VERSION)
    get_filename_component(WDK_ROOT ${WDK_LATEST_NTDDK_FILE} DIRECTORY)
    get_filename_component(WDK_ROOT ${WDK_ROOT} DIRECTORY)
    get_filename_component(WDK_VERSION ${WDK_ROOT} NAME)
    get_filename_component(WDK_ROOT ${WDK_ROOT} DIRECTORY)
    if (NOT WDK_ROOT MATCHES ".*/[0-9][0-9.]*$")
        get_filename_component(WDK_ROOT ${WDK_ROOT} DIRECTORY)
        set(WDK_LIB_VERSION "${WDK_VERSION}")
        set(WDK_INC_VERSION "${WDK_VERSION}")
    else()
        set(WDK_INC_VERSION "")
        foreach(VERSION winv6.3 win8 win7)
            if (EXISTS "${WDK_ROOT}/Lib/${VERSION}/")
                set(WDK_LIB_VERSION "${VERSION}")
                break()
            endif()
        endforeach()
        set(WDK_VERSION "${WDK_LIB_VERSION}")
    endif()
else()
    if(NOT DEFINED WDK_ROOT)
        string(REGEX REPLACE "[\\/]$" "" _WDK_ROOT "${_WDK_ROOT}")
        set(WDK_ROOT "${_WDK_ROOT}")
    endif()
    if(NOT DEFINED WDK_INC_VERSION)
        set(WDK_INC_VERSION "${WDK_VERSION}")
    endif()
    if(NOT DEFINED WDK_LIB_VERSION)
        set(WDK_LIB_VERSION "${WDK_VERSION}")
    endif()
endif()

if(NOT WDK_FIND_QUIETLY)
    message(STATUS "WDK_ROOT: " ${WDK_ROOT})
    message(STATUS "WDK_VERSION: " ${WDK_VERSION})
endif()

set(_WDK_ROOT "${WDK_ROOT}")

set(CMAKE_C_COMPILER_WORKS 1 CACHE INTERNAL "")
set(CMAKE_CXX_COMPILER_WORKS 1 CACHE INTERNAL "")

set(CMAKE_C_STANDARD_LIBRARIES "")
set(CMAKE_CXX_STANDARD_LIBRARIES "")

set(WDK_WINVER "0x0601" CACHE STRING "Default WINVER for WDK targets")
set(WDK_NTDDI_VERSION "" CACHE STRING "Specified NTDDI_VERSION for WDK targets if needed")
set(WDK_TEST_SIGN OFF CACHE BOOL "Enable test signing for drivers")
set(WDK_TEST_SIGN_NAME "HyperDbgTest" CACHE STRING "Certificate name for test signing")

set(WDK_ADDITIONAL_FLAGS_FILE "${CMAKE_CURRENT_BINARY_DIR}${CMAKE_FILES_DIRECTORY}/wdkflags.h")
file(WRITE ${WDK_ADDITIONAL_FLAGS_FILE} "#pragma runtime_checks(\"suc\", off)")

set(WDK_COMPILE_FLAGS
    "/Zp8"
    "/GF"
    "/GR-"
    "/Gz"
    "/kernel"
    "/FIwarning.h"
    "/FI${WDK_ADDITIONAL_FLAGS_FILE}"
	"/Oi"
    )

set(WDK_COMPILE_DEFINITIONS "WINNT=1")
set(WDK_COMPILE_DEFINITIONS_DEBUG "MSC_NOOPT;DEPRECATE_DDK_FUNCTIONS=1;DBG=1")

if(DEFINED CMAKE_SIZEOF_VOID_P)
    if(CMAKE_SIZEOF_VOID_P EQUAL 4)
        list(APPEND WDK_COMPILE_DEFINITIONS "_X86_=1;i386=1;STD_CALL")
        set(WDK_PLATFORM "x86")
    elseif(CMAKE_SIZEOF_VOID_P EQUAL 8 AND CMAKE_CXX_COMPILER_ARCHITECTURE_ID STREQUAL "ARM64")
        list(APPEND WDK_COMPILE_DEFINITIONS "_ARM64_;ARM64;_USE_DECLSPECS_FOR_SAL=1;STD_CALL")
        set(WDK_PLATFORM "arm64")
    elseif(CMAKE_SIZEOF_VOID_P EQUAL 8)
        list(APPEND WDK_COMPILE_DEFINITIONS "_AMD64_;AMD64")
        set(WDK_PLATFORM "x64")
    else()
        message(FATAL_ERROR "Unsupported architecture")
    endif()
endif()
if(NOT WDK_PLATFORM)
    set(WDK_PLATFORM "x64")
    list(APPEND WDK_COMPILE_DEFINITIONS "_AMD64_;AMD64")
endif()

if(DEFINED EWDK_UCRT_INC)
    set(_UCRT_INC "${EWDK_UCRT_INC}")
    set(_UCRT_LIB "${EWDK_UCRT_LIB}")
else()
    set(_UCRT_INC "${_WDK_ROOT}/Include/${WDK_INC_VERSION}/ucrt")
    set(_UCRT_LIB "${_WDK_ROOT}/Lib/${WDK_VERSION}/ucrt/${WDK_PLATFORM}")
endif()
list(APPEND CMAKE_INCLUDE_PATH ${_UCRT_INC})
list(APPEND CMAKE_LIBRARY_PATH ${_UCRT_LIB})

if(DEFINED EWDK_VC_INC_DIRS)
    set(_VC_INC_DIRS "${EWDK_VC_INC_DIRS}")
    set(_VC_LIB_DIRS "${EWDK_VC_LIB_DIRS}")
endif()

string(CONCAT WDK_LINK_FLAGS
    "/MANIFEST:NO "
    "/DRIVER "
    "/OPT:REF "
    "/INCREMENTAL:NO "
    "/OPT:ICF "
    "/SUBSYSTEM:NATIVE "
    "/ENTRY:DriverEntry "
    "/MERGE:_TEXT=.text;_PAGE=PAGE "
    "/NODEFAULTLIB "
    "/SECTION:INIT,d "
    "/VERSION:10.0 "
    )

file(GLOB WDK_LIBRARIES "${WDK_ROOT}/Lib/${WDK_LIB_VERSION}/km/${WDK_PLATFORM}/*.lib")
foreach(LIBRARY IN LISTS WDK_LIBRARIES)
    get_filename_component(LIBRARY_NAME ${LIBRARY} NAME_WE)
    string(TOUPPER ${LIBRARY_NAME} LIBRARY_NAME)
    add_library(WDK::${LIBRARY_NAME} INTERFACE IMPORTED)
    set_property(TARGET WDK::${LIBRARY_NAME} PROPERTY INTERFACE_LINK_LIBRARIES ${LIBRARY})
endforeach(LIBRARY)
unset(WDK_LIBRARIES)

function(wdk_add_driver _target)
    cmake_parse_arguments(WDK "" "KMDF;WINVER;NTDDI_VERSION" "" ${ARGN})

    add_executable(${_target} ${WDK_UNPARSED_ARGUMENTS})

    set_target_properties(${_target} PROPERTIES SUFFIX ".sys")
    set_target_properties(${_target} PROPERTIES COMPILE_OPTIONS "${WDK_COMPILE_FLAGS}")
    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "${WDK_COMPILE_DEFINITIONS};$<$<CONFIG:Debug>:${WDK_COMPILE_DEFINITIONS_DEBUG}>;_WIN32_WINNT=${WDK_WINVER}"
        )
    set_target_properties(${_target} PROPERTIES LINK_FLAGS "${WDK_LINK_FLAGS}")
    set_target_properties(${_target} PROPERTIES LINK_INTERFACE_LIBRARIES "")
    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE
        "${WDK_ROOT}/Include/${WDK_INC_VERSION}/shared"
        "${WDK_ROOT}/Include/${WDK_INC_VERSION}/km"
        "${WDK_ROOT}/Include/${WDK_INC_VERSION}/km/crt"
        )

    target_link_libraries(${_target} WDK::NTOSKRNL WDK::HAL WDK::WMILIB)

    if(WDK::BUFFEROVERFLOWK)
        target_link_libraries(${_target} WDK::BUFFEROVERFLOWK)
    else()
        target_link_libraries(${_target} WDK::BUFFEROVERFLOWFASTFAILK)
    endif()

    if(CMAKE_CXX_COMPILER_ARCHITECTURE_ID STREQUAL "ARM64")
        target_link_libraries(${_target} "arm64rt.lib")
    endif()

    if(CMAKE_SIZEOF_VOID_P EQUAL 4)
        target_link_libraries(${_target} WDK::MEMCMP)
    endif()

    if(DEFINED WDK_KMDF)
        target_include_directories(${_target} PRIVATE "${WDK_ROOT}/Include/wdf/kmdf/${WDK_KMDF}")
        target_link_libraries(${_target}
            "${WDK_ROOT}/Lib/wdf/kmdf/${WDK_PLATFORM}/${WDK_KMDF}/WdfDriverEntry.lib"
            "${WDK_ROOT}/Lib/wdf/kmdf/${WDK_PLATFORM}/${WDK_KMDF}/WdfLdr.lib"
            )

        if(CMAKE_SIZEOF_VOID_P EQUAL 4)
            set_property(TARGET ${_target} APPEND_STRING PROPERTY LINK_FLAGS "/ENTRY:FxDriverEntry@8")
        elseif(CMAKE_SIZEOF_VOID_P EQUAL 8)
            set_property(TARGET ${_target} APPEND_STRING PROPERTY LINK_FLAGS "/ENTRY:FxDriverEntry")
        endif()
    else()
        if(CMAKE_SIZEOF_VOID_P EQUAL 4)
            set_property(TARGET ${_target} APPEND_STRING PROPERTY LINK_FLAGS "/ENTRY:GsDriverEntry@8")
        elseif(CMAKE_SIZEOF_VOID_P EQUAL 8)
            set_property(TARGET ${_target} APPEND_STRING PROPERTY LINK_FLAGS "/ENTRY:GsDriverEntry")
        endif()
    endif()

    if(WDK_TEST_SIGN)
        if(CMAKE_SIZEOF_VOID_P EQUAL 4)
            set(WDK_SIGNTOOL_PATH "${EWDK_WDKContentRoot}/bin/${WDK_VERSION}/x86/signtool.exe")
        else()
            set(WDK_SIGNTOOL_PATH "${EWDK_WDKContentRoot}/bin/${WDK_VERSION}/x64/signtool.exe")
        endif()
        add_custom_command(TARGET ${_target} POST_BUILD
            COMMAND ${WDK_SIGNTOOL_PATH} sign /fd SHA256 /s My /n ${WDK_TEST_SIGN_NAME} /t http://timestamp.digicert.com $<TARGET_FILE:${_target}>
            WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
            COMMENT "Signing driver with test certificate: ${WDK_TEST_SIGN_NAME}"
            VERBATIM
        )
    endif()
endfunction()

function(wdk_add_library _target)
    cmake_parse_arguments(WDK "" "KMDF;WINVER;NTDDI_VERSION" "" ${ARGN})

    add_library(${_target} ${WDK_UNPARSED_ARGUMENTS})

    set_target_properties(${_target} PROPERTIES COMPILE_OPTIONS "${WDK_COMPILE_FLAGS}")
    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "${WDK_COMPILE_DEFINITIONS};$<$<CONFIG:Debug>:${WDK_COMPILE_DEFINITIONS_DEBUG};>_WIN32_WINNT=${WDK_WINVER}"
        )
    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE
        "${WDK_ROOT}/Include/${WDK_INC_VERSION}/shared"
        "${WDK_ROOT}/Include/${WDK_INC_VERSION}/km"
        "${WDK_ROOT}/Include/${WDK_INC_VERSION}/km/crt"
        )

    if(DEFINED WDK_KMDF)
        target_include_directories(${_target} PRIVATE "${WDK_ROOT}/Include/wdf/kmdf/${WDK_KMDF}")
    endif()
endfunction()

function(wdk_add_executable _target)
    cmake_parse_arguments(WDK "" "SUBSYSTEM;WINVER;NTDDI_VERSION" "" ${ARGN})

    if(NOT WDK_SUBSYSTEM)
        set(WDK_SUBSYSTEM "CONSOLE")
    endif()
    if(NOT WDK_WINVER)
        set(WDK_WINVER "${WDK_WINVER}")
    endif()

    add_executable(${_target} ${WDK_UNPARSED_ARGUMENTS})

    string(TOUPPER "${WDK_SUBSYSTEM}" _subsystem_upper)

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER}")

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE
        "${_WDK_ROOT}/Include/${WDK_INC_VERSION}/shared"
        "${_WDK_ROOT}/Include/${WDK_INC_VERSION}/um"
        ${_UCRT_INC}
        ${_VC_INC_DIRS}
        )

    if(_subsystem_upper STREQUAL "CONSOLE" OR _subsystem_upper STREQUAL "WINCON")
        set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:CONSOLE")
    elseif(_subsystem_upper STREQUAL "WINDOWS" OR _subsystem_upper STREQUAL "WIN")
        set_target_properties(${_target} PROPERTIES LINK_FLAGS "/SUBSYSTEM:WINDOWS")
    endif()

    target_link_options(${_target} PRIVATE "/LIBPATH:${_WDK_ROOT}/Lib/${WDK_VERSION}/um/${WDK_PLATFORM}")
    target_link_options(${_target} PRIVATE "/LIBPATH:${_UCRT_LIB}")
    foreach(_vc_lib_dir ${_VC_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_vc_lib_dir}")
    endforeach()
    target_link_libraries(${_target} kernel32.lib user32.lib)
endfunction()

function(um_library _target)
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION" "" ${ARGN})

    if(NOT WDK_WINVER)
        set(WDK_WINVER "${WDK_WINVER}")
    endif()

    add_library(${_target} ${WDK_UNPARSED_ARGUMENTS})

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS "_WIN32_WINNT=${WDK_WINVER}")

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE
        "${_WDK_ROOT}/Include/${WDK_INC_VERSION}/shared"
        "${_WDK_ROOT}/Include/${WDK_INC_VERSION}/um"
        ${_UCRT_INC}
        ${_VC_INC_DIRS}
        )
    target_link_options(${_target} PRIVATE "/LIBPATH:${_WDK_ROOT}/Lib/${WDK_VERSION}/um/${WDK_PLATFORM}")
    target_link_options(${_target} PRIVATE "/LIBPATH:${_UCRT_LIB}")
    foreach(_vc_lib_dir ${_VC_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_vc_lib_dir}")
    endforeach()
endfunction()

function(um_dll _target)
    cmake_parse_arguments(WDK "" "WINVER;NTDDI_VERSION" "" ${ARGN})

    if(NOT WDK_WINVER)
        set(WDK_WINVER "${WDK_WINVER}")
    endif()

    add_library(${_target} SHARED ${WDK_UNPARSED_ARGUMENTS})

    set_target_properties(${_target} PROPERTIES COMPILE_DEFINITIONS
        "_WIN32_WINNT=${WDK_WINVER};_USRDLL;_WINDLL"
        )

    if(WDK_NTDDI_VERSION)
        target_compile_definitions(${_target} PRIVATE NTDDI_VERSION=${WDK_NTDDI_VERSION})
    endif()

    target_include_directories(${_target} PRIVATE
        "${_WDK_ROOT}/Include/${WDK_INC_VERSION}/shared"
        "${_WDK_ROOT}/Include/${WDK_INC_VERSION}/um"
        ${_UCRT_INC}
        ${_VC_INC_DIRS}
        )
    target_link_options(${_target} PRIVATE "/LIBPATH:${_WDK_ROOT}/Lib/${WDK_VERSION}/um/${WDK_PLATFORM}")
    target_link_options(${_target} PRIVATE "/LIBPATH:${_UCRT_LIB}")
    foreach(_vc_lib_dir ${_VC_LIB_DIRS})
        target_link_options(${_target} PRIVATE "/LIBPATH:${_vc_lib_dir}")
    endforeach()
    target_link_libraries(${_target} kernel32.lib)
endfunction()
