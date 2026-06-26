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
4. Generate `C:/Program Files/CMake/bin/unity.cmake` (source collection + optional unity build helpers)
5. Copy `ninja.exe` to `C:/Program Files/CMake/bin/`
6. Create test signing certificate (`WDKTestCert`) if not exists
7. Generate `C:/Program Files/CMake/bin/ewdk.env.json` (environment snapshot)

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
├── unity.cmake             # Source collection + unity build helpers (auto-generated)
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

## Unity Build (Manual, via unity.cpp)

Unity Build (aka "merge compilation") 默认关闭，原因是 CMake 原生 `UNITY_BUILD` 有以下问题：
1. 不同源文件中的同名 `static` 变量/函数会冲突（如 `SwallowedException` / `VectoredHandler`）
2. CMake 自动分桶策略不可控，排查问题困难
3. x86 交叉编译（custom command mode）不支持 CMake 原生 unity

如需合并编译加速，项目中提供了 `generate_unity()` 函数，手动生成 `unity.cpp` 文件 `#include` 所有源文件，由开发者自行控制编译方式：

```cmake
# 在 CMakeLists.txt 中调用
collect_sources(
    src/subdir1 src/subdir2
    MY_SOURCES
)

# 生成 unity.cpp 并按需编译
generate_unity(unity ${MY_SOURCES})
# 然后用 unity.cpp 替换 SOURCES 中的文件列表
```

遇到 static 变量冲突时，切回逐个文件编译即可。

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
