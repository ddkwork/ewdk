cmake_minimum_required(VERSION 3.28)

# ── MSVC x86 (WDK) ──
# _asm{} 内联汇编仅支持 32 位，故使用 um_exe_x86() 通过 X86_CL/X86_LINK 编译
//include(ewdk.cmake)

project(asm C)

set(CMAKE_C_STANDARD 11)

um_exe_x86(asm
    SOURCES main.c
)
