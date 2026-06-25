#ifndef ZYDIS_COMPAT_H
#define ZYDIS_COMPAT_H

// Include the amalgamated Zydis header
#include "Zydis.h"

// Standard library headers (replaces Zycore/LibC.h)
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

// ---------------------------------------------------------------------------
// Zycore/LibC.h compatibility macros
// ---------------------------------------------------------------------------
#define ZYAN_STDERR     stderr
#define ZYAN_STDOUT     stdout
#define ZYAN_STDIN      stdin
#define ZYAN_FPRINTF    fprintf
#define ZYAN_PRINTF     printf
#define ZYAN_PUTS       puts
#define ZYAN_FPUTS      fputs
#define ZYAN_STRCMP     strcmp
#define ZYAN_STRLEN     strlen
#define ZYAN_ERRNO      errno
#define ZYAN_MEMSET     memset
#define ZYAN_MEMCPY     memcpy
#define ZYAN_MEMMOVE    memmove
#define ZYAN_MEMCMP     memcmp
#define ZYAN_GETENV     getenv
#define ZYAN_CALLOC     calloc
#define ZYAN_MALLOC     malloc
#define ZYAN_REALLOC    realloc
#define ZYAN_FREE       free
#define ZYAN_SSCANF     sscanf
#define ZYAN_VSNPRINTF  vsnprintf

// ---------------------------------------------------------------------------
// Zycore/API/Terminal.h - VT100 SGR sequences
// ---------------------------------------------------------------------------
#define ZYAN_VT100SGR_RESET             "\033[0m"
#define ZYAN_VT100SGR_FG_DEFAULT        "\033[39m"
#define ZYAN_VT100SGR_FG_BLACK          "\033[30m"
#define ZYAN_VT100SGR_FG_RED            "\033[31m"
#define ZYAN_VT100SGR_FG_GREEN          "\033[32m"
#define ZYAN_VT100SGR_FG_YELLOW         "\033[33m"
#define ZYAN_VT100SGR_FG_BLUE           "\033[34m"
#define ZYAN_VT100SGR_FG_MAGENTA        "\033[35m"
#define ZYAN_VT100SGR_FG_CYAN           "\033[36m"
#define ZYAN_VT100SGR_FG_WHITE          "\033[37m"
#define ZYAN_VT100SGR_FG_BRIGHT_BLACK   "\033[90m"
#define ZYAN_VT100SGR_FG_BRIGHT_RED     "\033[91m"
#define ZYAN_VT100SGR_FG_BRIGHT_GREEN   "\033[92m"
#define ZYAN_VT100SGR_FG_BRIGHT_YELLOW  "\033[93m"
#define ZYAN_VT100SGR_FG_BRIGHT_BLUE    "\033[94m"
#define ZYAN_VT100SGR_FG_BRIGHT_MAGENTA "\033[95m"
#define ZYAN_VT100SGR_FG_BRIGHT_CYAN    "\033[96m"
#define ZYAN_VT100SGR_FG_BRIGHT_WHITE   "\033[97m"

// ---------------------------------------------------------------------------
// Zycore/API/Terminal.h - Standard stream enum
// ---------------------------------------------------------------------------
typedef enum ZyanStandardStream_
{
    ZYAN_STDSTREAM_IN,
    ZYAN_STDSTREAM_OUT,
    ZYAN_STDSTREAM_ERR
} ZyanStandardStream;

// ---------------------------------------------------------------------------
// Zycore/API/Terminal.h - Function declarations
// ---------------------------------------------------------------------------
ZyanStatus ZyanTerminalEnableVT100(ZyanStandardStream stream);
ZyanStatus ZyanTerminalIsTTY(ZyanStandardStream stream);

// ---------------------------------------------------------------------------
// Zycore/API/Memory.h - Page protection enum
// ---------------------------------------------------------------------------
typedef enum ZyanMemoryPageProtection_
{
    ZYAN_PAGE_READONLY          = 0x02,
    ZYAN_PAGE_READWRITE         = 0x04,
    ZYAN_PAGE_EXECUTE           = 0x10,
    ZYAN_PAGE_EXECUTE_READ      = 0x20,
    ZYAN_PAGE_EXECUTE_READWRITE = 0x40
} ZyanMemoryPageProtection;

// ---------------------------------------------------------------------------
// Zycore/API/Memory.h - Function declaration
// ---------------------------------------------------------------------------
ZyanStatus ZyanMemoryVirtualProtect(void* address, ZyanUSize size,
    ZyanMemoryPageProtection protection);

#endif // ZYDIS_COMPAT_H
