cmake_minimum_required(VERSION 3.22)

include("C:/Program Files/CMake/bin/ewdk.cmake")

project(ZydisDemo)

include_directories(${CMAKE_CURRENT_SOURCE_DIR})
add_definitions(-DZYCORE_STATIC_BUILD -DZYDIS_STATIC_BUILD)

# --- User-mode static library (Zydis core + compat stubs) ---
um_lib(Zydis Zydis.c ZydisCompat.c)

# --- User-mode examples (decode/encode/formatter) ---
um_exe(ZydisDisassemble       SUBSYSTEM WINCON examples/Disassemble.c)
um_exe(ZydisDisassembleSimple SUBSYSTEM WINCON examples/DisassembleSimple.c)
um_exe(ZydisEncodeMov         SUBSYSTEM WINCON examples/EncodeMov.c)
um_exe(ZydisEncodeFromScratch SUBSYSTEM WINCON examples/EncodeFromScratch.c)
um_exe(ZydisFormatter01       SUBSYSTEM WINCON examples/Formatter01.c)
um_exe(ZydisFormatter02       SUBSYSTEM WINCON examples/Formatter02.c)
um_exe(ZydisFormatter03       SUBSYSTEM WINCON examples/Formatter03.c)
um_exe(ZydisRewriteCode       SUBSYSTEM WINCON examples/RewriteCode.c)

foreach(t
    ZydisDisassemble ZydisDisassembleSimple
    ZydisEncodeMov ZydisEncodeFromScratch
    ZydisFormatter01 ZydisFormatter02 ZydisFormatter03 ZydisRewriteCode)
    target_link_libraries(${t} Zydis)
endforeach()

# --- Kernel-mode static lib (Zydis core only, no compat stubs) ---
km_lib(ZydisKm Zydis.c)

# --- Kernel-mode driver ---
km_sys(ZydisWinKernel examples/ZydisWinKernel.c)
target_include_directories(ZydisWinKernel PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(ZydisWinKernel ZydisKm)
