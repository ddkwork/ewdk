#define _CRT_SECURE_NO_WARNINGS
#define WIN32_LEAN_AND_MEAN
#define NOMINMAX
#include <windows.h>
#include <winternl.h>
#include <stdint.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

#pragma comment(lib, "ntdll.lib")
#pragma comment(lib, "advapi32.lib")

static bool EnableDebugPrivilege(void) {
    HANDLE hToken;
    if (!OpenProcessToken(GetCurrentProcess(), TOKEN_ADJUST_PRIVILEGES | TOKEN_QUERY, &hToken))
        return false;

    TOKEN_PRIVILEGES tp;
    tp.PrivilegeCount = 1;
    tp.Privileges[0].Attributes = SE_PRIVILEGE_ENABLED;

    if (!LookupPrivilegeValueW(NULL, SE_DEBUG_NAME, &tp.Privileges[0].Luid)) {
        CloseHandle(hToken);
        return false;
    }

    BOOL ok = AdjustTokenPrivileges(hToken, FALSE, &tp, sizeof(tp), NULL, NULL);
    DWORD err = GetLastError();
    CloseHandle(hToken);
    return ok && (err == ERROR_SUCCESS);
}

typedef struct _RTL_PROCESS_MODULE_INFORMATION {
    HANDLE Section;
    PVOID MappedBase;
    PVOID ImageBase;
    ULONG ImageSize;
    ULONG Flags;
    USHORT LoadOrderIndex;
    USHORT InitOrderIndex;
    USHORT LoadCount;
    USHORT OffsetToFileName;
    UCHAR FullPathName[256];
} RTL_PROCESS_MODULE_INFORMATION, *PRTL_PROCESS_MODULE_INFORMATION;

typedef struct _RTL_PROCESS_MODULES {
    ULONG NumberOfModules;
    RTL_PROCESS_MODULE_INFORMATION Modules[1];
} RTL_PROCESS_MODULES, *PRTL_PROCESS_MODULES;

typedef struct _KERNEL_DRIVER_INFO {
    char Name[64];
    char FullPath[260];
    uint64_t ImageBase;
    uint32_t ImageSize;
    uint32_t LoadOrder;
    uint16_t OffsetToFileName;
    bool Valid;
} KERNEL_DRIVER_INFO;

typedef struct _KERNEL_DRIVER_LIST {
    KERNEL_DRIVER_INFO* Drivers;
    uint32_t Count;
    uint32_t Capacity;
    uint64_t NtoskrnlBase;
    uint32_t NtoskrnlSize;
    char NtoskrnlName[64];
    uint64_t LastUpdate;
} KERNEL_DRIVER_LIST;

PVOID Core_GetSystemInformation(SYSTEM_INFORMATION_CLASS infoClass) {
    NTSTATUS status;
    PVOID buffer = NULL;
    ULONG bufferSize = 0;
    ULONG returnLength = 0;

    status = NtQuerySystemInformation(infoClass, NULL, 0, &returnLength);
    if (returnLength == 0) {
        returnLength = 0x10000;
    }

    bufferSize = returnLength + 0x1000;
    buffer = LocalAlloc(LMEM_FIXED | LMEM_ZEROINIT, bufferSize);
    if (!buffer) return NULL;

    for (int i = 0; i < 5; i++) {
        status = NtQuerySystemInformation(infoClass, buffer, bufferSize, &returnLength);
        if (status == 0) {
            return buffer;
        }
        if (status == 0xC0000004L) {
            LocalFree(buffer);
            bufferSize = returnLength + 0x1000;
            buffer = LocalAlloc(LMEM_FIXED | LMEM_ZEROINIT, bufferSize);
            if (!buffer) return NULL;
        } else {
            break;
        }
    }

    if (buffer) LocalFree(buffer);
    return NULL;
}

bool Core_RefreshKernelDrivers(KERNEL_DRIVER_LIST* list) {
    if (!list) return false;

    PRTL_PROCESS_MODULES info = (PRTL_PROCESS_MODULES)Core_GetSystemInformation((SYSTEM_INFORMATION_CLASS)11);
    if (!info || info->NumberOfModules == 0) {
        if (info) LocalFree(info);
        return false;
    }

    if (list->Drivers == NULL) {
        list->Capacity = info->NumberOfModules + 32;
        list->Drivers = (KERNEL_DRIVER_INFO*)malloc(list->Capacity * sizeof(KERNEL_DRIVER_INFO));
        if (!list->Drivers) {
            list->Capacity = 0;
            LocalFree(info);
            return false;
        }
    } else if (list->Capacity < info->NumberOfModules) {
        list->Capacity = info->NumberOfModules + 32;
        KERNEL_DRIVER_INFO* newDrivers = (KERNEL_DRIVER_INFO*)realloc(list->Drivers,
            list->Capacity * sizeof(KERNEL_DRIVER_INFO));
        if (!newDrivers) {
            LocalFree(info);
            return false;
        }
        list->Drivers = newDrivers;
    }

    list->Count = 0;

    for (ULONG i = 0; i < info->NumberOfModules; i++) {
        PRTL_PROCESS_MODULE_INFORMATION module = &info->Modules[i];
        KERNEL_DRIVER_INFO* driver = &list->Drivers[list->Count];

        memset(driver, 0, sizeof(KERNEL_DRIVER_INFO));

        strncpy(driver->Name, (char*)(module->FullPathName + module->OffsetToFileName),
            sizeof(driver->Name) - 1);
        strncpy(driver->FullPath, (char*)module->FullPathName, sizeof(driver->FullPath) - 1);

        driver->ImageBase = (uint64_t)module->ImageBase;
        driver->ImageSize = module->ImageSize;
        driver->LoadOrder = module->LoadOrderIndex;
        driver->OffsetToFileName = module->OffsetToFileName;
        driver->Valid = true;

        if (i == 0) {
            list->NtoskrnlBase = driver->ImageBase;
            list->NtoskrnlSize = driver->ImageSize;
            strncpy(list->NtoskrnlName, driver->Name, sizeof(list->NtoskrnlName) - 1);
        }

        list->Count++;
    }

    list->LastUpdate = GetTickCount64();
    LocalFree(info);
    return true;
}

int main(void) {
    if (!EnableDebugPrivilege()) {
        printf("WARNING: Failed to enable SeDebugPrivilege (error %lu)\n", GetLastError());
        printf("Kernel addresses may be zeroed. Run as Administrator.\n\n");
    }

    KERNEL_DRIVER_LIST list = {0};

    if (!Core_RefreshKernelDrivers(&list)) {
        printf("Core_RefreshKernelDrivers failed\n");
        printf("Are you running as Administrator?\n");
        return 1;
    }

    printf("Kernel Drivers: %u\n", list.Count);
    printf("ntoskrnl: %s base=0x%llX size=0x%X\n\n",
        list.NtoskrnlName, list.NtoskrnlBase, list.NtoskrnlSize);

    printf("%-5s %-40s %-18s %-12s %s\n",
        "Order", "Name", "ImageBase", "ImageSize", "FullPath");
    printf("%-5s %-40s %-18s %-12s %s\n",
        "-----", "----------------------------------------", "------------------", "------------", "--------");

    for (uint32_t i = 0; i < list.Count; i++) {
        KERNEL_DRIVER_INFO* d = &list.Drivers[i];
        printf("%-5u %-40s 0x%016llX 0x%-10X %s\n",
            d->LoadOrder, d->Name, d->ImageBase, d->ImageSize, d->FullPath);
    }

    free(list.Drivers);
    return 0;
}
