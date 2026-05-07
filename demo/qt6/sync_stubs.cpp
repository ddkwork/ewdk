#include <windows.h>
#include <synchapi.h>

extern "C" {

BOOL WINAPI WaitOnAddress(volatile void* addr, PVOID compare, SIZE_T size, DWORD ms) {
    static auto fn = (BOOL(WINAPI*)(volatile void*, PVOID, SIZE_T, DWORD))GetProcAddress(GetModuleHandleW(L"kernel32.dll"), "WaitOnAddress");
    return fn ? fn(addr, compare, size, ms) : FALSE;
}

VOID WINAPI WakeByAddressAll(PVOID addr) {
    static auto fn = (VOID(WINAPI*)(PVOID))GetProcAddress(GetModuleHandleW(L"kernel32.dll"), "WakeByAddressAll");
    if (fn) fn(addr);
}

VOID WINAPI WakeByAddressSingle(PVOID addr) {
    static auto fn = (VOID(WINAPI*)(PVOID))GetProcAddress(GetModuleHandleW(L"kernel32.dll"), "WakeByAddressSingle");
    if (fn) fn(addr);
}

}
