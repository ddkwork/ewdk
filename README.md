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
km_sys(target sources... [LIBS libs...] [DEFINES defs...] [INCLUDES dirs...])
km_lib(target sources... [LIBS libs...] [DEFINES defs...] [INCLUDES dirs...])
```

**Example:**
```cmake
km_sys(MyDriver main.c)
km_lib(MyLib lib.c)
km_sys(ZydisWinKernel ZydisWinKernel.c DEFINES ZYAN_NO_LIBC ZYDIS_NO_LIBC LIBS ZydisKm)
```

### User Mode Functions

```cmake
um_exe(target [SUBSYSTEM CONSOLE|WINCON|WINDOWS|WIN]
    [SOURCES ...] [INCLUDES ...] [DEFINES ...] [LIBS ...] [COMPILE_OPTIONS ...] [NOAUTO])
um_lib(target [SOURCES ...] [INCLUDES ...] [DEFINES ...] [LIBS ...] [COMPILE_OPTIONS ...])
um_dll(target [SOURCES ...] [INCLUDES ...] [DEFINES ...] [LIBS ...] [COMPILE_OPTIONS ...])
um_exe_mfc(target [SOURCES ...] [INCLUDES ...] [DEFINES ...] [LIBS ...] [COMPILE_OPTIONS ...])
um_dp64(target [SOURCES ...] [LIBS ...] [DEFINES ...] [INCLUDES ...] [COMPILE_OPTIONS ...] [PLUGINSDK ...])
um_dp86(target [SOURCES ...] [LIBS ...] [DEFINES ...] [INCLUDES ...] [COMPILE_OPTIONS ...] [PLUGINSDK ...])

# x86 cross-compile (Hostx64→x86) variants:
um_exe_x86(target [SOURCES ...] [INCLUDES ...] [DEFINES ...] [LIBS ...] [COMPILE_OPTIONS ...])
um_dll_x86(target [SOURCES ...] [INCLUDES ...] [DEFINES ...] [LIBS ...] [COMPILE_OPTIONS ...])
um_lib_x86(target [SOURCES ...] [INCLUDES ...] [DEFINES ...] [LIBS ...])
um_exe_mfc_x86(target [SOURCES ...] [INCLUDES ...] [DEFINES ...] [LIBS ...])
```

**Examples:**
```cmake
um_exe(MyApp SUBSYSTEM CONSOLE SOURCES main.c LIBS ntdll.lib)
um_exe(NTMemory SUBSYSTEM WINDOWS
    SOURCES main.cpp imgui.cpp
    INCLUDES imgui
    DEFINES UNICODE
    NOAUTO
    LIBS d3d11.lib dxgi.lib
)
um_dp64(MyPlugin
    SOURCES plugin.cpp
    PLUGINSDK x64bridge x64dbg
    LIBS winspool.lib
)
um_dp86(HyperHide32
    SOURCES plugin.cpp
    PLUGINSDK x32bridge x32dbg
    LIBS kernel32.lib
)
```

**x64dbg Plugin Functions** (`um_dp64`/`um_dp86`):
- Create a user-mode DLL with `.dp64` (x64) or `.dp32` (x86) suffix
- `um_dp64` auto-copies `.dp64` to x64dbg's plugins directory
- `um_dp86` (x86 cross-compile) auto-copies `.dp32` to x32dbg's plugins directory
- `PLUGINSDK` keyword: adds `pluginsdk/` include dir, resolves bare names like `x64bridge` to `${CMAKE_CURRENT_SOURCE_DIR}/pluginsdk/x64bridge`
- Variables `X64DBG_X64_DIR` / `X64DBG_X32_DIR` available for custom copy commands

All user mode functions automatically:
- Set correct include paths (WDK UM includes + VC includes)
- Set `_WIN32_WINNT` definition (default `0x0601`)
- Set `MSVC_RUNTIME_LIBRARY` to MultiThreaded (Debug: MultiThreadedDebug)
- Add `/LIBPATH:` for all UM lib directories
- For EXE: link `kernel32.lib` + `user32.lib`
- For DLL: define `_USRDLL` / `_WINDLL`, link `kernel32.lib`

## Unity Build (Automatic)

All user-mode templates automatically enable **Unity Build** (aka "merge compilation" or "single translation unit build"):

- **Native CMake targets** (`um_exe`, `um_lib`, `um_dll`, `um_dp64`, `um_exe_mfc`): `UNITY_BUILD ON` — CMake groups sources into batches and compiles each batch in a single `cl.exe` invocation. Default batch size is 8 files per batch.

- **x86 cross-compile targets** (`um_dp86`, `um_exe_x86`, `um_dll_x86`, `um_lib_x86`, `um_exe_mfc_x86`): A single `_unity_<target>.cpp` is auto-generated in the build directory that `#include`s all source files. Only **one compilation** per target instead of N individual compilations.

**Effect**: Instead of compiling 31 source files × 2 platforms = 62 times, a project needs roughly 4 (x64 batches) + 1 (x86 unity) = 5 compilations total. Build time drops from several minutes to under a minute.

**Footnote**: PCH (precompiled headers) is explicitly superseded by Unity Build — everything compiles at most once per translation unit, making PCH redundant.

### Excluding files from Unity Build

If a source file is incompatible with unity build (e.g., stub implementations with mismatched `extern "C"` signatures), exclude it per-file:

```cmake
set_source_files_properties(${CMAKE_CURRENT_SOURCE_DIR}/callback_stubs.cpp PROPERTIES SKIP_UNITY_BUILD_INCLUSION ON)
```

The excluded file compiles separately as its own `.obj` — linker resolves symbol mismatches by name.

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
