# EWDK Demo Project

Windows Driver Kit (WDK) demo project with CMake build system.

## Features

- **Kernel Driver**: `KernelDriverDemo.sys` - Simple kernel driver (auto-signed with test certificate)
- **User-mode EXE**: `UmExeDemo.exe` - Console application
- **Static Library**: `UmLibDemo.lib` - User-mode static library
- **Dynamic Library**: `UmDllDemo.dll` - User-mode DLL

## Prerequisites

- Windows Driver Kit (WDK) installed (EWDK ISO mount handled by `mount-task.ps1`)
- CMake 3.22+
- Ninja build tool (if not in PATH, place `ninja.exe` in project root)
- Visual Studio Build Tools (included in EWDK)

## Quick Start

### One-click Build

```batch
build.bat
```

This will compile all demos and output to `build\demo\` directory.

## Project Structure

```
d:\ewdk\
├── CMakeLists.txt          # Main CMake configuration
├── ewdk.cmake              # WDK module (kernel + user-mode functions)
├── build.bat               # One-click build script
├── ninja.exe               # Ninja build tool (if not in PATH)
├── demo/
│   ├── kernel/             # Kernel driver demo
│   │   ├── CMakeLists.txt
│   │   └── main.c
│   ├── um-exe/             # User-mode executable demo
│   │   ├── CMakeLists.txt
│   │   └── main.c
│   ├── um-lib/             # Static library demo
│   │   ├── CMakeLists.txt
│   │   └── lib.c
│   └── um-dll/             # DLL demo
│       ├── CMakeLists.txt
│       └── dllmain.c
└── build/                  # Output directory (generated)
    └── demo/
        ├── um-exe/UmExeDemo.exe
        ├── um-lib/UmLibDemo.lib
        ├── um-dll/UmDllDemo.dll
        └── kernel/KernelDriverDemo.sys
```

## Available CMake Functions

The `ewdk.cmake` module provides:

### Kernel Mode Functions

```cmake
km_sys(target [KMDF version] sources...)
km_lib(target [KMDF version] sources...)
```

**Example:**
```cmake
km_sys(MyDriver KMDF 1.15 main.c)
km_lib(MyLib main.c)
```

### User Mode Functions

```cmake
um_exe(target [SUBSYSTEM CONSOLE|WINDOWS] sources...)
um_lib(target sources...)
um_dll(target sources...)
```

**Examples:**
```cmake
# Console application
um_exe(MyApp SUBSYSTEM CONSOLE main.c)

# Windows GUI application
um_exe(MyGuiApp SUBSYSTEM WIN main.cpp)

# Static library
um_lib(MyLib utils.c)

# DLL
um_dll(MyPlugin plugin.c)
```

All user mode functions automatically:
- Set correct include paths (`shared/`, `um/`)
- Set `_WIN32_WINNT` definition
- Link required libraries (`kernel32.lib`, `user32.lib`)
- For DLLs: set `_USRDLL`, `_WINDLL` definitions

## Configuration Variables

Set in environment or CMake cache:

| Variable | Default | Description |
|----------|---------|-------------|
| `WDKContentRoot` | Auto-detected | WDK installation root |
| `WDK_WINVER` | `0x0601` | Target Windows version (Win7) |
| `KM_TEST_SIGN` | `ON` | Enable automatic test signing for kernel drivers |
| `KM_TEST_SIGN_NAME` | `WDKTestCert` | Certificate name for test signing |

## Output Files

After running `build.bat`:

| File | Type | Size | Location |
|------|------|------|----------|
| `UmExeDemo.exe` | EXE | ~10 KB | `build/demo/um-exe/` |
| `UmLibDemo.lib` | LIB | ~1 KB | `build/demo/um-lib/` |
| `UmDllDemo.dll` | DLL | ~9 KB | `build/demo/um-dll/` |
| `KernelDriverDemo.sys` | SYS | ~6 KB | `build/demo/kernel/` |

## License

Based on [FindWDK](https://github.com/podobry/FindWDK) by Sergey Podobry (BSD 3-Clause License).
