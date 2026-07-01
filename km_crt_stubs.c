#include <ntddk.h>
#include <stdarg.h>  /* va_list, va_arg */

#pragma function(memset, memcpy)

/* ---- internal helper: format unsigned integer ---- */
__forceinline static char* fmt_uint64(char* dst, unsigned long long v, int base, int upper, int width, int zp)
{
    char buf[24];
    char* p = buf + sizeof(buf);
    const char* dig = upper ? "0123456789ABCDEF" : "0123456789abcdef";
    *--p = 0;
    if (!v) { *--p = '0'; }
    else { while (v) { *--p = dig[v % (unsigned)base]; v /= (unsigned)base; } }
    int len = (int)((buf + sizeof(buf)) - p - 1);
    while (len < width && p > buf) { *--p = (char)(zp ? '0' : ' '); len++; }
    while (*p) { *dst++ = *p++; }
    return dst;
}

/* ---- core formatter: called by __stdio_common_vsprintf family ---- */
static int vsprintf_core(char* Buffer, size_t BufferCount, const char* Format, va_list ArgList)
{
    if (!Buffer || !Format) return -1;
    if (!BufferCount) return 0;

    char* d = Buffer;
    const char* f = Format;
    size_t space = BufferCount;
    int written = 0;

#define PUT(c)  do { if (space > 1) { *d++ = (c); space--; written++; } } while(0)
#define PUTS(s) do { const char* _p = (s); while (_p && *_p) { PUT(*_p++); } } while(0)

    while (*f && space > 1)
    {
        if (*f != '%') { PUT(*f++); continue; }
        f++;

        /* flags */
        int zp = 0;
        for (;;) {
            if (*f == '-') f++;
            else if (*f == '0') { zp = 1; f++; }
            else break;
        }

        /* width */
        int w = 0;
        while (*f >= '0' && *f <= '9') { w = w * 10 + (*f - '0'); f++; }

        /* precision (skip) */
        if (*f == '.') { f++; while (*f >= '0' && *f <= '9') f++; }

        /* length */
        int ll = 0;
        if (*f == 'h') { f++; if (*f == 'h') f++; }
        else if (*f == 'l') { ll = 1; f++; if (*f == 'l') { ll = 2; f++; } }
        else if (*f == 'z') { ll = 1; f++; }
        else if (*f == 't') { f++; }
        else if (*f == 'w') {
            /* %wZ - UNICODE_STRING */
            if (*(f+1) == 'Z') {
                f += 2;
                PCUNICODE_STRING us = va_arg(ArgList, PCUNICODE_STRING);
                if (us && us->Buffer) {
                    for (int i = 0; i < (int)(us->Length / sizeof(WCHAR)); i++) {
                        PUT((char)us->Buffer[i]);
                    }
                } else { PUTS("(null)"); }
            }
            continue;
        }

        switch (*f) {
            case 's': {
                const char* s = va_arg(ArgList, const char*);
                PUTS(s ? s : "(null)");
                break;
            }
            case 'S': {
                const wchar_t* ws = va_arg(ArgList, const wchar_t*);
                if (ws) { while (*ws) { PUT((char)*ws++); } }
                else { PUTS("(null)"); }
                break;
            }
            case 'd': case 'i': {
                long long v;
                if (ll >= 2) v = va_arg(ArgList, long long);
                else if (ll == 1) v = va_arg(ArgList, long);
                else v = va_arg(ArgList, int);
                if (v < 0) { PUT('-'); v = -v; }
                d = fmt_uint64(d, (unsigned long long)v, 10, 0, w, zp);
                written = (int)(d - Buffer);
                space = BufferCount - written;
                break;
            }
            case 'u': {
                unsigned long long v;
                if (ll >= 2) v = va_arg(ArgList, unsigned long long);
                else if (ll == 1) v = va_arg(ArgList, unsigned long);
                else v = va_arg(ArgList, unsigned int);
                d = fmt_uint64(d, v, 10, 0, w, zp);
                written = (int)(d - Buffer);
                space = BufferCount - written;
                break;
            }
            case 'x': {
                unsigned long long v;
                if (ll >= 2) v = va_arg(ArgList, unsigned long long);
                else if (ll == 1) v = va_arg(ArgList, unsigned long);
                else v = va_arg(ArgList, unsigned int);
                d = fmt_uint64(d, v, 16, 0, w, zp);
                written = (int)(d - Buffer);
                space = BufferCount - written;
                break;
            }
            case 'X': {
                unsigned long long v;
                if (ll >= 2) v = va_arg(ArgList, unsigned long long);
                else if (ll == 1) v = va_arg(ArgList, unsigned long);
                else v = va_arg(ArgList, unsigned int);
                d = fmt_uint64(d, v, 16, 1, w, zp);
                written = (int)(d - Buffer);
                space = BufferCount - written;
                break;
            }
            case 'p': {
                void* p = va_arg(ArgList, void*);
                PUT('0'); PUT('x');
                d = fmt_uint64(d, (unsigned long long)(ULONG_PTR)p, 16, 1, 0, 1);
                written = (int)(d - Buffer);
                space = BufferCount - written;
                break;
            }
            case 'c': {
                int ch = va_arg(ArgList, int);
                PUT((char)ch);
                break;
            }
            case '%': { PUT('%'); break; }
            default: { PUT(*f); break; }
        }
        if (*f) f++;
    }
    *d = 0;
#undef PUT
#undef PUTS
    return written;
}

/* ---- UCRT stubs: __stdio_common_vsprintf family ---- */
int __cdecl __stdio_common_vsprintf_s(unsigned __int64 Options, char* Buffer, size_t BufferCount, const char* Format, _locale_t Locale, char* ArgList)
{
    (void)Options; (void)Locale;
    return vsprintf_core(Buffer, BufferCount, Format, (va_list)ArgList);
}

int __cdecl __stdio_common_vsprintf(unsigned __int64 Options, char* Buffer, size_t BufferCount, const char* Format, _locale_t Locale, char* ArgList)
{
    (void)Options; (void)Locale;
    return vsprintf_core(Buffer, BufferCount, Format, (va_list)ArgList);
}

int __cdecl __stdio_common_vsnprintf_s(unsigned __int64 Options, char* Buffer, size_t BufferCount, size_t MaxCount, const char* Format, _locale_t Locale, char* ArgList)
{
    (void)Options; (void)Locale; (void)MaxCount;
    size_t cb = (BufferCount < MaxCount) ? BufferCount : MaxCount;
    return vsprintf_core(Buffer, cb, Format, (va_list)ArgList);
}