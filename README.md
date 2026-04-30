# EWDK Demo Project

Windows Driver Kit (WDK) demo project with CMake build system.

## Features

- **Kernel Driver**: `KernelDriverDemo.sys` - Simple kernel driver (auto-signed with test certificate)
- **User-mode EXE**: `UmExeDemo.exe` - Console application
- **Static Library**: `UmLibDemo.lib` - User-mode static library
- **Dynamic Library**: `UmDllDemo.dll` - User-mode DLL

## Prerequisites

- Windows 10/11
- CMake 3.22+ installed at `C:/Program Files/CMake/`
- EWDK ISO file (e.g., `EWDK_br_release_28000_251103-1709.iso`) in project root
- Go 1.26+ (for running `main.go`)

## Quick Start

### Step 1: Setup Environment

```batch
go run .
```

This will:
1. Mount the EWDK ISO
2. Execute `SetupBuildEnv.cmd amd64` and parse environment variables
3. Generate `C:/Program Files/CMake/bin/ewdk.cmake` (complete WDK build module)
4. Copy `ninja.exe` to `C:/Program Files/CMake/bin/`
5. Create test signing certificate (`WDKTestCert`) if not exists
6. Generate `C:/Program Files/CMake/bin/ewdk.env.json` (environment snapshot)

### Step 2: Build

```batch
build.bat
```

This will compile all demos in both Release and Debug configurations, output to `build/` and `build_debug/` directories.

## Project Structure

```
d:\ewdk\
├── CMakeLists.txt          # Main CMake (includes ewdk.cmake from CMake bin dir)
├── main.go                 # EWDK toolchain manager (mount/generate/build-env)
├── go.mod / go.sum         # Go module (github.com/ddkwork/golibrary v0.2.2)
├── build.bat               # One-click build script (Release + Debug)
├── ninja.exe               # Ninja build tool (copied to CMake bin dir at runtime)
├── scripts/
│   └── get_ewdk_url.go     # Fetch EWDK ISO download URL from Microsoft
├── .github/
│   └── workflows/
│       └── ci.yml          # CI: test → build → release
└── demo/
    ├── kernel/             # Kernel driver demo
    │   ├── CMakeLists.txt  # km_sys(KernelDriverDemo main.c)
    │   └── main.c
    ├── um-exe/             # User-mode executable demo
    │   ├── CMakeLists.txt  # um_exe(UmExeDemo SUBSYSTEM WINCON main.c)
    │   └── main.c
    ├── um-lib/             # Static library demo
    │   ├── CMakeLists.txt  # um_lib(UmLibDemo lib.c)
    │   └── lib.c
    └── um-dll/             # DLL demo
        ├── CMakeLists.txt  # um_dll(UmDllDemo dllmain.c)
        └── dllmain.c

Generated at runtime (not in repo):
C:/Program Files/CMake/bin/
├── ewdk.cmake              # Complete WDK build module (auto-generated)
├── ewdk.env.json           # Environment snapshot (auto-generated)
└── ninja.exe               # Ninja build tool (copied from project root)
```

## Available CMake Functions

The `ewdk.cmake` module (generated at `C:/Program Files/CMake/bin/ewdk.cmake`) provides:

### Kernel Mode Functions

```cmake
km_sys(target [KMDF version] sources...)
km_lib(target [KMDF version] sources...)
```

**Example:**
```cmake
km_sys(MyDriver main.c)
km_sys(MyKmdfDriver KMDF 1.15 main.c)
km_lib(MyLib main.c)
```

### User Mode Functions

```cmake
um_exe(target [SUBSYSTEM CONSOLE|WINCON|WINDOWS|WIN] sources...)
um_lib(target sources...)
um_dll(target sources...)
```

**Examples:**
```cmake
um_exe(MyApp SUBSYSTEM CONSOLE main.c)
um_exe(MyGuiApp SUBSYSTEM WIN main.cpp)
um_lib(MyLib utils.c)
um_dll(MyPlugin plugin.c)
```

All user mode functions automatically:
- Set correct include paths (WDK UM includes + VC includes)
- Set `_WIN32_WINNT` definition (default `0x0601`)
- Set `MSVC_RUNTIME_LIBRARY` to MultiThreaded (Debug: MultiThreadedDebug)
- Add `/LIBPATH:` for all UM lib directories
- For EXE: link `kernel32.lib` + `user32.lib`
- For DLL: define `_USRDLL` / `_WINDLL`, link `kernel32.lib`

## Configuration Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WDK_WINVER` | `0x0601` | Target Windows version (Win7) |
| `WDK_NTDDI_VERSION` | `""` | Optional NTDDI_VERSION definition |
| `KM_TEST_SIGN` | `ON` (local) / `OFF` (CI) | Enable automatic test signing for kernel drivers |
| `KM_TEST_SIGN_NAME` | `WDKTestCert` | Certificate name for test signing |

## Commands

```bash
# Setup: mount ISO + generate ewdk.cmake + copy ninja + create cert
go run .

# Unmount ISO
go run . unmount

# Clean generated ewdk.cmake
go run . clean

# Get EWDK ISO download URL
go run scripts/get_ewdk_url.go

# Build as standalone executable
go build -o ewdk.exe .
```

## Output Files

After running `build.bat`:

| File | Type | Config | Location |
|------|------|--------|----------|
| `KernelDriverDemo.sys` | SYS | Release | `build/demo/kernel/` |
| `UmExeDemo.exe` | EXE | Release | `build/demo/um-exe/` |
| `UmLibDemo.lib` | LIB | Release | `build/demo/um-lib/` |
| `UmDllDemo.dll` | DLL | Release | `build/demo/um-dll/` |
| `KernelDriverDemo.sys` | SYS | Debug | `build_debug/demo/kernel/` |
| `UmExeDemo.exe` | EXE | Debug | `build_debug/demo/um-exe/` |
| `UmLibDemo.lib` | LIB | Debug | `build_debug/demo/um-lib/` |
| `UmDllDemo.dll` | DLL | Debug | `build_debug/demo/um-dll/` |

## License

Based on [FindWDK](https://github.com/podobry/FindWDK) by Sergey Podobry (BSD 3-Clause License).
