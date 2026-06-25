#include "ZydisCompat.h"
#include <windows.h>
#include <stdarg.h>

ZyanStatus ZyanTerminalEnableVT100(ZyanStandardStream stream)
{
    HANDLE h = (stream == ZYAN_STDSTREAM_OUT) ? GetStdHandle(STD_OUTPUT_HANDLE)
              : (stream == ZYAN_STDSTREAM_ERR) ? GetStdHandle(STD_ERROR_HANDLE)
              : INVALID_HANDLE_VALUE;
    if (h == INVALID_HANDLE_VALUE || h == NULL)
        return ZYAN_STATUS_SUCCESS;

    DWORD mode;
    if (!GetConsoleMode(h, &mode))
        return ZYAN_STATUS_SUCCESS;
    mode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING;
    SetConsoleMode(h, mode);
    return ZYAN_STATUS_SUCCESS;
}

ZyanStatus ZyanTerminalIsTTY(ZyanStandardStream stream)
{
    HANDLE h = (stream == ZYAN_STDSTREAM_OUT) ? GetStdHandle(STD_OUTPUT_HANDLE)
              : (stream == ZYAN_STDSTREAM_ERR) ? GetStdHandle(STD_ERROR_HANDLE)
              : GetStdHandle(STD_INPUT_HANDLE);
    if (h == INVALID_HANDLE_VALUE || h == NULL)
        return ZYAN_STATUS_FALSE;

    DWORD mode;
    if (!GetConsoleMode(h, &mode))
        return ZYAN_STATUS_FALSE;
    return ZYAN_STATUS_TRUE;
}

ZyanStatus ZyanMemoryVirtualProtect(void* address, ZyanUSize size,
    ZyanMemoryPageProtection protection)
{
    DWORD old;
    if (VirtualProtect(address, size, (DWORD)protection, &old))
        return ZYAN_STATUS_SUCCESS;
    return ZYAN_STATUS_FALSE;
}

// Minimal ZyanStringAppendFormat — only needed by Formatter01/02 examples.
ZyanStatus ZyanStringAppendFormat(ZyanString* string, const char* format, ...)
{
    if (!string || !format)
        return ZYAN_STATUS_INVALID_ARGUMENT;

    // Current string length (excluding null terminator)
    ZyanUSize len = string->vector.size > 0 ? string->vector.size - 1 : 0;

    va_list ap;
    va_start(ap, format);
    ZyanI32 needed = vsnprintf(NULL, 0, format, ap);
    va_end(ap);
    if (needed < 0)
        return ZYAN_STATUS_FAILED;

    ZyanUSize new_total = len + (ZyanUSize)needed + 1; // +1 for null
    if (new_total > string->vector.capacity)
    {
        if (string->flags & ZYAN_STRING_HAS_FIXED_CAPACITY)
            return ZYAN_STATUS_INSUFFICIENT_BUFFER_SIZE;

        void* new_data = realloc(string->vector.data, new_total);
        if (!new_data)
            return ZYAN_STATUS_NOT_ENOUGH_MEMORY;
        string->vector.data = new_data;
        string->vector.capacity = new_total;
    }

    va_start(ap, format);
    vsnprintf((char*)string->vector.data + len, (size_t)(needed + 1), format, ap);
    va_end(ap);

    string->vector.size = new_total;
    return ZYAN_STATUS_SUCCESS;
}
