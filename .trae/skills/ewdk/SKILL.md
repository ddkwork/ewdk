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

本仓库通过**挂载 ISO 时由 Go 动态生成完整的 `ewdk.cmake` 文件**来彻底解决这个问题。所有 EWDK 环境变量、编译参数、WDK 库、构建函数全部嵌入单个 cmake 文件，生成到 `C:/Program Files/CMake/bin/ewdk.cmake`，任何 CMake 工程只需 `include("C:/Program Files/CMake/bin/ewdk.cmake")` 即可使用，无需依赖系统环境变量或注册表。

## 核心架构

```
d:\ewdk\
├── main.go             # 全部逻辑：ISO 挂载/卸载、SetupBuildEnv.cmd 执行、ewdk.cmake 生成
│                       #   - ewdkEnv 嵌套结构体 (Common/KM/UM): SetupBuildEnv.cmd 执行结果
│                       #   - mountISO/unmountISO: ISO 挂载卸载
│                       #   - runSetupBuildEnv: 执行 SetupBuildEnv.cmd amd64，解析环境变量
│                       #   - generateEwdkCmake: 生成完整 ewdk.cmake（环境变量+编译参数+函数）
│                       #   - ensureTestCertificate: 自动创建测试签名证书
│                       #   - cleanGenerated: 清理生成的文件
├── CMakeLists.txt      # 根 CMake（include("C:/Program Files/CMake/bin/ewdk.cmake")）
├── build.bat           # 一键构建脚本（Release + Debug）
├── ninja.exe           # Ninja 构建工具源文件（运行时复制到 CMake bin 目录）
├── go.mod / go.sum     # Go 模块定义（依赖 github.com/ddkwork/golibrary v0.2.2）
├── scripts/
│   └── get_ewdk_url.go # 获取 EWDK ISO 下载链接（从微软官网解析）
├── .github/
│   └── workflows/
│       └── ci.yml      # CI：test → build → release 三阶段
└── demo/               # 示例工程
    ├── kernel/         # km_sys → .sys
    ├── um-exe/         # um_exe → .exe
    ├── um-lib/         # um_lib → .lib
    └── um-dll/         # um_dll → .dll

生成目标（不在仓库内）：
C:/Program Files/CMake/bin/
├── ewdk.cmake          # [自动生成] 完整的 WDK 构建模块
├── ewdk.env.json       # [自动生成] 环境变量 JSON 快照
└── ninja.exe           # [自动复制] 从 d:\ewdk\ninja.exe 复制
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

## 关键常量与外部包

```go
const isoPath = `EWDK_br_release_28000_251103-1709.iso`  // ISO 文件名
const testSignCertName = "WDKTestCert"                    // 测试签名证书名

// github.com/ddkwork/golibrary/cmake 包提供：
const BinDir = "C:/Program Files/CMake/bin"
var EwdkCmakeFile = filepath.Join(BinDir, "ewdk.cmake")  // "C:/Program Files/CMake/bin/ewdk.cmake"
var EwdkEnvFile   = filepath.Join(BinDir, "ewdk.env.json") // "C:/Program Files/CMake/bin/ewdk.env.json"
func Module() string  // 返回 CMake Modules 目录路径
```

## 标准工作流程（main.go 主逻辑）

当用户执行 `go run .` 时，按以下顺序执行：

### Step 0: 检测 CMake 模块目录

```go
info := cmake.Module()  // 查找 C:/Program Files/CMake/share/cmake-xxx/Modules
mylog.Success(info)
```

### Step 1: 处理命令行参数（如有）

```go
// "unmount" → 卸载 ISO
// "clean"   → 清理生成的 ewdk.cmake
// 无参数    → 继续以下步骤
```

### Step 2: 挂载 ISO

```go
isoPath := resolveISOPath()  // CI 环境从 TEMP/ewdk.iso 读取，本地从常量 isoPath 读取
if isMounted() {
    driveLetter = getEwdkDriveLetter()  // 从 F: 到 A: 扫描 SetupBuildEnv.cmd
} else {
    driveLetter = mountISO(isoPath)     // PowerShell Mount-DiskImage
}
```

### Step 3: 执行 SetupBuildEnv.cmd（已内置 amd64 参数）

```go
setupEnvCmd := driveLetter + ":\\BuildEnv\\SetupBuildEnv.cmd"
env := runSetupBuildEnv(setupEnvCmd)
// 内部：cmd /c setupCmd amd64 && set > tempfile
// 解析 8 个关键环境变量：WindowsTargetPlatformVersion, VCToolsInstallDir,
//   WDKContentRoot, BuildLabSetupRoot, VSINSTALLDIR, INCLUDE, LIB, WDKBinRoot
```

### Step 4: 生成 ewdk.cmake

```go
generateEwdkCmake(env, cmake.EwdkCmakeFile)
// 生成到 C:/Program Files/CMake/bin/ewdk.cmake
```

Go 一次性生成完整的 `ewdk.cmake`，包含：
- **WDK Core**: `WDK_FOUND` / `WDK_ROOT` / `WDK_VERSION` / `WDK_PLATFORM` 等
- **KM**: `WDK_KM_INCLUDE_DIRS` / `WDK_KM_LIB_DIRS`
- **UM**: `WDK_UM_INCLUDE_DIRS` / `WDK_UM_LIB_DIRS`
- **Compiler**: `CMAKE_C_COMPILER` / `CMAKE_CXX_COMPILER` / `CMAKE_RC_COMPILER` / `CMAKE_MT` 等
- **Environment**: `ENV{WDKContentRoot}` / `ENV{INCLUDE}` / `ENV{LIB}`
- **Settings**: `KM_COMPILE_FLAGS` / `KM_COMPILE_DEFINITIONS` / `KM_LINK_FLAGS` / `WDK_WINVER` / `KM_TEST_SIGN`
- **Libraries**: `file(GLOB)` + `foreach` 循环自动注册所有 `WDK::*` 库
- **Functions**: `km_sys` / `km_lib` / `um_exe` / `um_lib` / `um_dll`
- **SignTool**: `KM_SIGNTOOL_PATH`（由 Go 动态查找 signtool.exe）

### Step 5: 复制 ninja.exe 到 CMake bin 目录

```go
stream.CopyFile("ninja.exe", filepath.Join(cmake.BinDir, "ninja.exe"))
// 从 d:\ewdk\ninja.exe 复制到 C:/Program Files/CMake/bin/ninja.exe
```

### Step 6: 确保测试签名证书存在

```go
ensureTestCertificate()
// PowerShell 检查 Cert:\CurrentUser\My 中是否存在 CN=WDKTestCert
// 不存在则自动创建 New-SelfSignedCertificate -Type CodeSigningCert
```

### Step 7: 生成 ewdk.env.json

```go
envData := json.MarshalIndent(env, "", "  ")
os.WriteFile(cmake.EwdkEnvFile, envData, 0644)
// 生成到 C:/Program Files/CMake/bin/ewdk.env.json
```

### Step 8: 提示用户构建

```go
mylog.Success("Environment ready. Run build.bat to start building.")
```

ewdk.exe 只负责环境准备，不直接调用构建。用户运行 `build.bat`（或手动 cmake 命令）时，`CMakeLists.txt` 直接 `include("C:/Program Files/CMake/bin/ewdk.cmake")`，所有变量和函数已就绪。

## ewdk.cmake 结构（Go 生成）

生成的 `ewdk.cmake`（位于 `C:/Program Files/CMake/bin/ewdk.cmake`）分为以下段落：

```cmake
# ---- WDK Core ----
set(WDK_FOUND TRUE)
set(WDK_ROOT "...")
set(WDK_VERSION "...")
set(WDK_INC_VERSION "...")
set(WDK_LIB_VERSION "...")
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
set(CMAKE_C_COMPILER_WORKS 1 CACHE INTERNAL "")
set(CMAKE_CXX_COMPILER_WORKS 1 CACHE INTERNAL "")
set(CMAKE_C_STANDARD_LIBRARIES "")
set(CMAKE_CXX_STANDARD_LIBRARIES "")
set(CMAKE_INCLUDE_PATH "..." CACHE STRING "" FORCE)
set(CMAKE_LIBRARY_PATH "..." CACHE STRING "" FORCE)
list(APPEND CMAKE_PROGRAM_PATH "...")
set(CMAKE_MT "..." CACHE FILEPATH "" FORCE)

# ---- Environment Variables ----
set(ENV{WDKContentRoot} "${WDK_ROOT}")
set(ENV{INCLUDE} "${WDK_KM_INCLUDE_DIRS};${WDK_UM_INCLUDE_DIRS}")
set(ENV{LIB} "${WDK_KM_LIB_DIRS};${WDK_UM_LIB_DIRS}")

# ---- WDK Settings ----
set(WDK_WINVER "0x0601" CACHE STRING "")
set(WDK_NTDDI_VERSION "" CACHE STRING "")
set(KM_TEST_SIGN ON/OFF CACHE BOOL "")  # CI 中 OFF，本地 ON
set(KM_TEST_SIGN_NAME "WDKTestCert" CACHE STRING "")
set(KM_ADDITIONAL_FLAGS_FILE "...")     # 禁用 runtime_checks 的临时头文件
set(KM_COMPILE_FLAGS ...)
set(KM_COMPILE_DEFINITIONS "WINNT=1;_AMD64_;AMD64")
set(KM_COMPILE_DEFINITIONS_DEBUG "MSC_NOOPT;DEPRECATE_DDK_FUNCTIONS=1;DBG=1")
set(KM_LINK_FLAGS ...)

# ---- KM Libraries ----
file(GLOB KM_LIB_FILES ".../*.lib")
foreach(LIB_FILE ${KM_LIB_FILES})
    get_filename_component(LIB_NAME ${LIB_FILE} NAME_WE)
    string(TOUPPER ${LIB_NAME} LIB_NAME)
    add_library(WDK::${LIB_NAME} INTERFACE IMPORTED)
    set_property(TARGET WDK::${LIB_NAME} PROPERTY INTERFACE_LINK_LIBRARIES ${LIB_FILE})
endforeach()

# ---- SignTool ----
set(KM_SIGNTOOL_PATH "...")

# ---- KM/UM Functions ----
function(km_sys ...)
function(km_lib ...)
function(um_exe ...)
function(um_lib ...)
function(um_dll ...)
```

## ewdk.cmake 提供的函数

| 函数 | 用途 | 输出 | 关键选项 |
|------|------|------|----------|
| `km_sys(target [KMDF ver] srcs...)` | 内核驱动 | `.sys` | `KMDF` / `WINVER` / `NTDDI_VERSION` |
| `km_lib(target [KMDF ver] srcs...)` | 内核库 | `.lib` | `KMDF` / `WINVER` / `NTDDI_VERSION` |
| `um_exe(target [SUBSYSTEM s] srcs...)` | 用户态 EXE | `.exe` | `SUBSYSTEM CONSOLE\|WINCON\|WINDOWS\|WIN` / `WINVER` / `NTDDI_VERSION` |
| `um_lib(target srcs...)` | 用户态静态库 | `.lib` | `WINVER` / `NTDDI_VERSION` |
| `um_dll(target srcs...)` | 用户态 DLL | `.dll` | `WINVER` / `NTDDI_VERSION` |

### 函数行为细节

**km_sys**: 自动链接 `WDK::NTOSKRNL` `WDK::HAL` `WDK::WMILIB` + `WDK::BUFFEROVERFLOWK`（或 `BUFFEROVERFLOWFASTFAILK`）。KMDF 模式下额外链接 WdfDriverEntry.lib/WdfLdr.lib 并修改入口为 FxDriverEntry。非 KMDF 入口为 GsDriverEntry。`KM_TEST_SIGN=ON` 时自动用 signtool 签名。

**um_exe**: 默认 SUBSYSTEM=CONSOLE。自动链接 kernel32.lib + user32.lib。设置 MSVC_RUNTIME_LIBRARY 为 MultiThreaded（Debug 为 MultiThreadedDebug）。

**um_lib**: 纯静态库，设置 MSVC 运行时库。

**um_dll**: SHARED 库，自动定义 `_USRDLL` `_WINDLL`，链接 kernel32.lib。

## ewdk.cmake 中设置的变量

| 变量 | 说明 |
|------|------|
| `WDK_FOUND` | 始终为 TRUE |
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
| `CMAKE_C_COMPILER_WORKS` | 1 (CACHE INTERNAL，跳过编译器检测) |
| `CMAKE_CXX_COMPILER_WORKS` | 1 (CACHE INTERNAL，跳过编译器检测) |
| `CMAKE_MT` | mt.exe (CACHE FORCE) |
| `CMAKE_C_STANDARD_LIBRARIES` | 清空（防止内核驱动链接用户态库） |
| `CMAKE_CXX_STANDARD_LIBRARIES` | 清空（防止内核驱动链接用户态库） |
| `CMAKE_INCLUDE_PATH` | 头文件搜索路径 (CACHE FORCE) |
| `CMAKE_LIBRARY_PATH` | 库文件搜索路径 (CACHE FORCE) |
| `CMAKE_PROGRAM_PATH` | 程序搜索路径 (含 cl/rc 目录) |
| `WDK_WINVER` | 目标 Windows 版本，默认 0x0601 (Win7) |
| `WDK_NTDDI_VERSION` | 可选 NTDDI_VERSION 定义 |
| `KM_TEST_SIGN` | 是否自动测试签名（CI=OFF，本地=ON） |
| `KM_TEST_SIGN_NAME` | 测试签名证书名 (WDKTestCert) |
| `KM_ADDITIONAL_FLAGS_FILE` | 临时头文件，禁用 runtime_checks |
| `KM_SIGNTOOL_PATH` | signtool.exe 路径（Go 动态查找） |
| `KM_COMPILE_FLAGS` | KM 编译选项 (/Zp8 /GF /GR- /Gz /kernel /Oi ...) |
| `KM_COMPILE_DEFINITIONS` | KM 编译定义 (WINNT=1;_AMD64_;AMD64) |
| `KM_COMPILE_DEFINITIONS_DEBUG` | Debug 编译定义 (MSC_NOOPT;DEPRECATE_DDK_FUNCTIONS=1;DBG=1) |
| `KM_LINK_FLAGS` | KM 链接选项 (/MANIFEST:NO /DRIVER /SUBSYSTEM:NATIVE /NODEFAULTLIB ...) |

## CI 流程 (.github/workflows/ci.yml)

三阶段流水线：

### Job 1: test
- Checkout

### Job 2: build（依赖 test）
1. Checkout
2. 安装 aria2（多线程下载工具）
3. 缓存/下载 EWDK ISO（key 基于 `scripts/get_ewdk_url.go` 哈希）
4. **挂载 ISO + 生成 ewdk.cmake** (`go run .`)
5. **构建 Release 和 Debug** (`build.bat`)

### Job 3: release（依赖 build）
1. 删除旧 release（tag: `ci-release`）
2. 创建新 tag 并推送
3. 创建源码归档并发布 GitHub Release

## 常见操作命令

```bash
# 一键执行：挂载→生成ewdk.cmake→复制ninja→创建证书→生成env.json
go run .

# 卸载 ISO
go run . unmount

# 清理生成的文件（仅删除 ewdk.cmake）
go run . clean

# 编译为可执行文件
go build -o ewdk.exe .
./ewdk.exe

# 获取 EWDK ISO 下载链接
go run scripts/get_ewdk_url.go

# 一键构建（Release + Debug）
build.bat
```

## 注意事项

1. **不再写入注册表**：所有环境变量通过 `ewdk.cmake` 传递给 CMake，无需修改系统环境变量
2. **`ewdk.cmake` 是自动生成的**：每次挂载 ISO 后由 Go 重新生成到 `C:/Program Files/CMake/bin/ewdk.cmake`，不应手动编辑或提交到版本控制
3. **`amd64` 已内置于 `runSetupBuildEnv`**：无需调用者手动拼接参数
4. **单文件设计**：Go 生成完整的 `ewdk.cmake`，包含环境变量、编译参数、库注册、构建函数，外部只需 `include("C:/Program Files/CMake/bin/ewdk.cmake")`
5. **CMakeLists.txt 极简**：仅需 `include("C:/Program Files/CMake/bin/ewdk.cmake")` 一行
6. **Go 负责所有查找**：signtool.exe、ntddk.h、.lib 文件均由 Go 在生成时定位，CMake 侧无需 file(GLOB) 检测或 fallback 逻辑
7. **KM/UM 完全分离**：`WDK_KM_INCLUDE_DIRS` / `WDK_UM_INCLUDE_DIRS` 独立设置，函数按需引用
8. **其他项目使用**：在 CMakeLists.txt 中 `include("C:/Program Files/CMake/bin/ewdk.cmake")` 即可
9. **ninja.exe 自动部署**：运行时从项目根目录复制到 `C:/Program Files/CMake/bin/ninja.exe`
10. **测试签名自动管理**：`ensureTestCertificate()` 自动检测并创建 `WDKTestCert` 代码签名证书
11. **环境快照**：`ewdk.env.json` 记录完整的 ewdkEnv 结构体，便于调试和审计
12. **ISO 路径解析**：CI 环境从 `%TEMP%/ewdk.iso` 读取，本地从硬编码常量 `EWDK_br_release_28000_251103-1709.iso` 读取
