---
name: "ewdk"
description: "Manages EWDK (Enterprise WDK) ISO mounting, environment variable setup, and CMake-based driver/UM build. Invoke when user needs to mount EWDK ISO, setup build environment, build kernel drivers or user-mode binaries with WDK toolchain, or fix EWDK-related env var issues."
---

# EWDK Skill - Windows Driver Kit 环境管理与构建

## 仓库概述

本仓库 (`d:\ewdk`) 是一个 **EWDK (Enterprise Windows Driver Kit) 工具链管理器**，核心目标：**让任何 CMake 工程都能零配置使用 WDK 编译内核驱动和用户态程序**。

### 解决的问题

EWDK 是 ISO 镜像形式发布的，传统使用方式痛苦：
- 每次开机需要手动挂载 ISO
- `SetupBuildEnv.cmd` 设置的环境变量是**临时的**（仅当前 cmd 窗口有效）
- 其他终端/CMake/IDE 无法获取这些变量
- `os.getenv` 不能实时从系统注册表获取环境变量

本仓库通过**挂载 ISO 时由 Go 动态生成完整的 `ewdk.cmake` 文件**来彻底解决这个问题。所有 EWDK 环境变量、编译参数、WDK 库、构建函数全部嵌入单个 cmake 文件，任何 CMake 工程只需 `include(ewdk.cmake)` 即可使用，无需依赖系统环境变量或注册表。

## 核心架构

```
d:\ewdk\
├── main.go             # 全部逻辑：ISO 挂载/卸载、SetupBuildEnv.cmd 执行、ewdk.cmake 生成
│                       #   - ewdkEnv 嵌套结构体 (Common/KM/UM): SetupBuildEnv.cmd 执行结果
│                       #   - mountISO/unmountISO: ISO 挂载卸载
│                       #   - runSetupBuildEnv: 执行 SetupBuildEnv.cmd amd64，解析环境变量
│                       #   - generateEwdkCmake: 生成完整 ewdk.cmake（环境变量+编译参数+函数）
├── ewdk.cmake          # [自动生成] 完整的 WDK 构建模块（环境变量+编译参数+库+函数）
├── CMakeLists.txt      # 根 CMake（仅 include(ewdk.cmake)）
├── build.bat           # 一键构建脚本
├── ninja.exe           # Ninja 构建工具
├── scripts/
│   └── get_ewdk_url.go # 获取 EWDK ISO 下载链接
└── demo/               # 示例工程
    ├── kernel/         # wdk_add_driver → .sys
    ├── um-exe/         # wdk_add_executable → .exe
    ├── um-lib/         # um_library → .lib
    └── um-dll/         # um_dll → .dll
```

## 关键数据结构

### ewdkEnv — 嵌套结构体（Common/KM/UM 分离）

```go
type ewdkCommonEnv struct {
    WDKContentRoot               string
    WindowsTargetPlatformVersion string
    VCToolsInstallDir            string
    WDKBinRoot                   string
    DiaRoot                      string
    VSINSTALLDIR                 string
    BuildLabSetupRoot            string
    CC                           string
    RC                           string
    MT                           string
    SignTool                     string
    NTDDKFile                    string
    NinjaDir                     string
}

type ewdkKMEnv struct {
    IncludeDirs []string
    LibDirs     []string
}

type ewdkUMEnv struct {
    IncludeDirs []string
    LibDirs     []string
}

type ewdkEnv struct {
    Common ewdkCommonEnv
    KM     ewdkKMEnv
    UM     ewdkUMEnv
}
```

## 标准工作流程（main.go 主逻辑）

当用户执行 `go run .` 时，按以下顺序执行：

### Step 1: 挂载 ISO

```go
isoPath := resolveISOPath()
if isMounted() {
    driveLetter = getEwdkDriveLetter()
} else {
    driveLetter = mountISO(isoPath)
}
```

### Step 2: 执行 SetupBuildEnv.cmd（已内置 amd64 参数）

```go
setupEnvCmd := driveLetter + ":\\BuildEnv\\SetupBuildEnv.cmd"
env := runSetupBuildEnv(setupEnvCmd)
```

### Step 3: 生成 ewdk.cmake

```go
generateEwdkCmake(env, "ewdk.cmake")
```

Go 一次性生成完整的 `ewdk.cmake`，包含：
- **WDK Core**: `WDK_ROOT` / `WDK_VERSION` / `WDK_PLATFORM` 等
- **KM**: `WDK_KM_INCLUDE_DIRS` / `WDK_KM_LIB_DIRS`
- **UM**: `WDK_UM_INCLUDE_DIRS` / `WDK_UM_LIB_DIRS`
- **Compiler**: `CMAKE_C_COMPILER` / `CMAKE_RC_COMPILER` / `CMAKE_MT` 等
- **Settings**: `WDK_COMPILE_FLAGS` / `WDK_COMPILE_DEFINITIONS` / `WDK_LINK_FLAGS`
- **Libraries**: `file(GLOB)` + `foreach` 循环自动注册所有 `WDK::*` 库
- **Functions**: `wdk_add_driver` / `wdk_add_library` / `wdk_add_executable` / `um_library` / `um_dll`
- **SignTool**: `WDK_SIGNTOOL_PATH`（由 Go 动态查找 signtool.exe）

### Step 4: 提示用户构建

```go
mylog.Success("Environment ready. Run build.bat to start building.")
```

ewdk.exe 只负责环境准备，不直接调用构建。用户运行 `build.bat`（或手动 cmake 命令）时，`CMakeLists.txt` 直接 `include(ewdk.cmake)`，所有变量和函数已就绪。

## ewdk.cmake 结构（Go 生成）

生成的 `ewdk.cmake` 分为以下段落：

```cmake
# ---- WDK Core ----
set(WDK_FOUND TRUE)
set(WDK_ROOT "...")
set(WDK_VERSION "...")
set(WDK_PLATFORM "x64")

# ---- KM (Kernel-Mode) ----
set(WDK_KM_INCLUDE_DIRS "...")
set(WDK_KM_LIB_DIRS "...")

# ---- UM (User-Mode) ----
set(WDK_UM_INCLUDE_DIRS "...")
set(WDK_UM_LIB_DIRS "...")

# ---- Compiler ----
set(CMAKE_C_COMPILER "..." CACHE FILEPATH "" FORCE)
set(CMAKE_CXX_COMPILER "..." CACHE FILEPATH "" FORCE)
set(CMAKE_RC_COMPILER "..." CACHE FILEPATH "" FORCE)
set(CMAKE_C_STANDARD_LIBRARIES "")
set(CMAKE_CXX_STANDARD_LIBRARIES "")

# ---- WDK Settings ----
set(WDK_COMPILE_FLAGS ...)
set(WDK_COMPILE_DEFINITIONS "WINNT=1;_AMD64_;AMD64")
set(WDK_LINK_FLAGS ...)

# ---- WDK Libraries ----
file(GLOB WDK_LIB_FILES ".../*.lib")
foreach(LIB_FILE ${WDK_LIB_FILES})
    get_filename_component(LIB_NAME ${LIB_FILE} NAME_WE)
    string(TOUPPER ${LIB_NAME} LIB_NAME)
    add_library(WDK::${LIB_NAME} INTERFACE IMPORTED)
    set_property(TARGET WDK::${LIB_NAME} PROPERTY INTERFACE_LINK_LIBRARIES ${LIB_FILE})
endforeach()

# ---- WDK Functions ----
function(wdk_add_driver ...)
function(wdk_add_library ...)
function(wdk_add_executable ...)
function(um_library ...)
function(um_dll ...)
```

## ewdk.cmake 提供的函数

| 函数 | 用途 | 输出 |
|------|------|------|
| `wdk_add_driver(target [KMDF ver] srcs...)` | 内核驱动 | `.sys` |
| `wdk_add_library(target [KMDF ver] srcs...)` | 内核库 | `.lib` |
| `wdk_add_executable(target [SUBSYSTEM CONSOLE\|WIN] srcs...)` | 用户态 EXE | `.exe` |
| `um_library(target srcs...)` | 用户态静态库 | `.lib` |
| `um_dll(target srcs...)` | 用户态 DLL | `.dll` |

## ewdk.cmake 中设置的变量

| 变量 | 说明 |
|------|------|
| `WDK_ROOT` | WDK 内容根目录 |
| `WDK_VERSION` | WDK 目标平台版本 |
| `WDK_INC_VERSION` | WDK Include 版本 |
| `WDK_LIB_VERSION` | WDK Lib 版本 |
| `WDK_PLATFORM` | 目标平台 (x64) |
| `WDK_KM_INCLUDE_DIRS` | KM 头文件搜索路径 |
| `WDK_KM_LIB_DIRS` | KM 库文件搜索路径 |
| `WDK_UM_INCLUDE_DIRS` | UM 头文件搜索路径 |
| `WDK_UM_LIB_DIRS` | UM 库文件搜索路径 |
| `CMAKE_C_COMPILER` | cl.exe (CACHE FORCE) |
| `CMAKE_CXX_COMPILER` | cl.exe (CACHE FORCE) |
| `CMAKE_RC_COMPILER` | rc.exe (CACHE FORCE) |
| `CMAKE_MT` | mt.exe (CACHE FORCE) |
| `CMAKE_C_STANDARD_LIBRARIES` | 清空（防止内核驱动链接用户态库） |
| `CMAKE_CXX_STANDARD_LIBRARIES` | 清空（防止内核驱动链接用户态库） |
| `CMAKE_INCLUDE_PATH` | 头文件搜索路径 (CACHE FORCE) |
| `CMAKE_LIBRARY_PATH` | 库文件搜索路径 (CACHE FORCE) |
| `CMAKE_PROGRAM_PATH` | 程序搜索路径 (含 ninja/cl/rc) |
| `WDK_SIGNTOOL_PATH` | signtool.exe 路径（Go 动态查找） |
| `WDK_COMPILE_FLAGS` | KM 编译选项 (/Zp8 /GF /GR- /Gz /kernel ...) |
| `WDK_COMPILE_DEFINITIONS` | KM 编译定义 (WINNT=1;_AMD64_;AMD64) |
| `WDK_COMPILE_DEFINITIONS_DEBUG` | Debug 编译定义 (DBG=1;MSC_NOOPT;...) |
| `WDK_LINK_FLAGS` | KM 链接选项 (/DRIVER /SUBSYSTEM:NATIVE /NODEFAULTLIB ...) |

## CI 流程 (.github/workflows/ci.yml)

1. Checkout → 安装 aria2 → 获取 EWDK ISO URL → 下载 ISO
2. **挂载 ISO + 生成 ewdk.cmake** (`go run .`)
3. **构建 Release 和 Debug** (cmake -B build / cmake --build)

## 常见操作命令

```bash
# 一键执行：挂载→生成ewdk.cmake
go run .

# 卸载 ISO
go run . unmount

# 清理生成的文件
go run . clean

# 编译为可执行文件
go build -o ewdk.exe .
./ewdk.exe
```

## 注意事项

1. **不再写入注册表**：所有环境变量通过 `ewdk.cmake` 传递给 CMake，无需修改系统环境变量
2. **`ewdk.cmake` 是自动生成的**：每次挂载 ISO 后由 Go 重新生成，不应手动编辑或提交到版本控制
3. **`amd64` 已内置于 `runSetupBuildEnv`**：无需调用者手动拼接参数
4. **单文件设计**：Go 生成完整的 `ewdk.cmake`，包含环境变量、编译参数、库注册、构建函数，外部只需 `include(ewdk.cmake)`
5. **CMakeLists.txt 极简**：仅需 `include(ewdk.cmake)` 一行
6. **Go 负责所有查找**：signtool.exe、ntddk.h、.lib 文件均由 Go 在生成时定位，CMake 侧无需 file(GLOB) 检测或 fallback 逻辑
7. **KM/UM 完全分离**：`WDK_KM_INCLUDE_DIRS` / `WDK_UM_INCLUDE_DIRS` 独立设置，函数按需引用
8. **其他项目使用**：只需将生成的 `ewdk.cmake` 复制到目标项目，在 CMakeLists.txt 中 `include(ewdk.cmake)` 即可
