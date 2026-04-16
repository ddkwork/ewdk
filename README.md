# EWDK Demo Project

Windows Driver Kit (WDK) demo project with CMake build system.

## Features

- **Kernel Driver**: `KernelDriverDemo.sys` - Simple kernel driver
- **User-mode EXE**: `UmExeDemo.exe` - Console application
- **Static Library**: `UmLibDemo.lib` - User-mode static library
- **Dynamic Library**: `UmDllDemo.dll` - User-mode DLL

## Prerequisites

- Windows Driver Kit (WDK) installed or EWDK ISO mounted
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
wdk_add_driver(target [KMDF version] sources...)
wdk_add_library(target [STATIC|SHARED] [KMDF version] sources...)
```

**Example:**
```cmake
wdk_add_driver(MyDriver KMDF 1.15 main.c)
wdk_add_library(MyLib STATIC main.c)
```

### User Mode Functions

```cmake
wdk_add_executable(target [SUBSYSTEM CONSOLE|WINDOWS] sources...)
um_library(target sources...)
um_dll(target sources...)
```

**Examples:**
```cmake
# Console application
wdk_add_executable(MyApp SUBSYSTEM WINCON main.c)

# Windows GUI application
wdk_add_executable(MyGuiApp SUBSYSTEM WIN main.cpp)

# Static library
um_library(MyLib utils.c)

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
