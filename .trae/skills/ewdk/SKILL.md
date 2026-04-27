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

本仓库通过**挂载 ISO 时动态生成 `ewdk-env.cmake` 文件**来彻底解决这个问题。所有 EWDK 环境变量直接嵌入 cmake 文件，任何 CMake 工程只需 include 即可使用，无需依赖系统环境变量或注册表。

## 核心架构

```
d:\ewdk\
├── env.go              # 核心逻辑：ISO 挂载/卸载、SetupBuildEnv.cmd 执行、ewdk-env.cmake 生成
│                       #   - ewdkEnv 结构体: SetupBuildEnv.cmd 执行结果
│                       #   - mountISO/unmountISO: ISO 挂载卸载
│                       #   - runSetupBuildEnv: 执行 SetupBuildEnv.cmd amd64，返回 ewdkEnv
│                       #   - generateEwdkCmake: 将 ewdkEnv 写入 ewdk-env.cmake
│                       #   - createScheduledTask/deleteScheduledTask: 开机自动挂载
├── main.go             # 主入口：挂载→执行SetupBuildEnv→生成ewdk-env.cmake→构建
├── ewdk.cmake          # CMake FindWDK 模块（优先使用 EWDK_* 变量，回退到 $ENV{}）
├── ewdk-env.cmake      # [自动生成] 包含所有 EWDK 环境变量的 cmake 文件
├── CMakeLists.txt      # 根 CMake（自动 include ewdk-env.cmake + ewdk.cmake）
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

### ewdkEnv — SetupBuildEnv.cmd 执行结果（已内置 amd64 参数）

```go
type ewdkEnv struct {
    WindowsTargetPlatformVersion string
    WDKContentRoot               string
    BuildLabSetupRoot            string
    VSINSTALLDIR                 string
    INCLUDE                      []string
    LIB                          []string
    WDKBinRoot                   string
    DiaRoot                      string
    VCToolsInstallDir            string
    CC                           string
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
createScheduledTask(isoPath)
```

### Step 2: 执行 SetupBuildEnv.cmd（**已内置 amd64 参数**）

```go
setupEnvCmd := driveLetter + ":\\BuildEnv\\SetupBuildEnv.cmd"
env := runSetupBuildEnv(setupEnvCmd)
```

### Step 3: 生成 ewdk-env.cmake

```go
generateEwdkCmake(env, "ewdk-env.cmake")
```

生成的 `ewdk-env.cmake` 包含：
- `EWDK_WDKContentRoot` / `EWDK_WindowsTargetPlatformVersion` / `EWDK_VCToolsInstallDir` 等 EWDK_* 变量
- `EWDK_INCLUDE` / `EWDK_LIB` 完整路径
- `CMAKE_C_COMPILER` / `CMAKE_CXX_COMPILER` / `CMAKE_RC_COMPILER` 编译器设置
- `CMAKE_INCLUDE_PATH` / `CMAKE_LIBRARY_PATH` 搜索路径
- `CMAKE_PROGRAM_PATH` 含 ninja/cl/rc 路径
- `ENV{WDKContentRoot}` / `ENV{INCLUDE}` / `ENV{LIB}` 环境变量注入

### Step 4: 提示用户构建

```go
mylog.Success("Environment ready. Run build.bat to start building.")
```

ewdk.exe 只负责环境准备，不直接调用构建。用户运行 `build.bat`（或手动 cmake 命令）时，`CMakeLists.txt` 自动 include `ewdk-env.cmake`，所有变量已就绪。

## ewdk.cmake 双模式检测

ewdk.cmake 支持两种模式：

1. **EWDK 模式（优先）**：检测到 `EWDK_WDKContentRoot` 变量时，直接使用 `EWDK_*` 变量，无需系统环境变量
2. **回退模式**：检测 `$ENV{WDKContentRoot}` 或系统安装路径

```cmake
if(DEFINED EWDK_WDKContentRoot)
    # 使用 ewdk-env.cmake 注入的变量
elseif(DEFINED ENV{WDKContentRoot})
    # 回退到系统环境变量
else()
    # 回退到系统安装路径
endif()
```

## ewdk.cmake 提供的函数

| 函数 | 用途 | 输出 |
|------|------|------|
| `wdk_add_driver(target [KMDF ver] srcs...)` | 内核驱动 | `.sys` |
| `wdk_add_library(target [STATIC\|SHARED] [KMDF ver] srcs...)` | 内核库 | `.lib` / `.dll` |
| `wdk_add_executable(target [SUBSYSTEM CONSOLE\|WIN] srcs...)` | 用户态 EXE | `.exe` |
| `um_library(target srcs...)` | 用户态静态库 | `.lib` |
| `um_dll(target srcs...)` | 用户态 DLL | `.dll` |

## ewdk-env.cmake 中设置的变量

| 变量 | 说明 |
|------|------|
| `EWDK_WDKContentRoot` | WDK 内容根目录 |
| `EWDK_WindowsTargetPlatformVersion` | WDK 目标平台版本 |
| `EWDK_VCToolsInstallDir` | VC 工具安装目录 |
| `EWDK_WDKBinRoot` | WDK Bin 根目录 |
| `EWDK_DiaRoot` | DIA SDK 目录 |
| `EWDK_VSINSTALLDIR` | VS 安装目录 |
| `EWDK_BuildLabSetupRoot` | 构建实验室根目录 |
| `EWDK_CC` / `EWDK_CXX` | x64 cl.exe 编译器路径 |
| `EWDK_RC` | rc.exe 资源编译器路径 |
| `EWDK_NINJA_DIR` | ninja.exe 所在目录 |
| `EWDK_INCLUDE` | 完整头文件搜索路径 |
| `EWDK_LIB` | 完整库文件搜索路径 |
| `CMAKE_C_COMPILER` | cl.exe (CACHE FORCE) |
| `CMAKE_CXX_COMPILER` | cl.exe (CACHE FORCE) |
| `CMAKE_RC_COMPILER` | rc.exe (CACHE FORCE) |
| `CMAKE_INCLUDE_PATH` | 头文件搜索路径 (CACHE FORCE) |
| `CMAKE_LIBRARY_PATH` | 库文件搜索路径 (CACHE FORCE) |
| `CMAKE_PROGRAM_PATH` | 程序搜索路径 (含 ninja/cl/rc) |
| `CMAKE_MT` | mt.exe 清单工具路径 (CACHE FORCE) |
| `CMAKE_C_STANDARD_LIBRARIES` | 清空默认 C 标准库（防止内核驱动链接用户态库） |
| `CMAKE_CXX_STANDARD_LIBRARIES` | 清空默认 C++ 标准库（防止内核驱动链接用户态库） |

## CI 流程 (.github/workflows/ci.yml)

1. Checkout → 安装 aria2 → 获取 EWDK ISO URL → 下载 ISO
2. **挂载 ISO + 生成 ewdk-env.cmake** (`go run .`)
3. **构建 Release 和 Debug** (cmake -B build / cmake --build)

## 常见操作命令

```bash
# 一键执行：挂载→生成cmake→构建
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

1. **不再写入注册表**：所有环境变量通过 `ewdk-env.cmake` 传递给 CMake，无需修改系统环境变量
2. **`ewdk-env.cmake` 是自动生成的**：每次挂载 ISO 后重新生成，不应手动编辑或提交到版本控制
3. **`amd64` 已内置于 `runSetupBuildEnv`**：无需调用者手动拼接参数
4. **ewdk.cmake 优先使用 EWDK_* 变量**：当 `ewdk-env.cmake` 被 include 后，`EWDK_WDKContentRoot` 已定义，ewdk.cmake 直接使用
5. **CMakeLists.txt 自动 include**：`if(EXISTS ewdk-env.cmake)` 确保文件存在时才 include，不影响没有该文件时的回退逻辑
6. **开机自动挂载**：通过 Windows 计划任务实现，`go run .` 时自动注册
7. **其他项目使用**：只需将 `ewdk-env.cmake` 和 `ewdk.cmake` 复制到目标项目，在 CMakeLists.txt 中 include 即可
