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
- INCLUDE/LIB 等旧变量会干扰 WDK 的正确路径

本仓库通过**将 EWDK 的环境变量永久写入系统注册表**来彻底解决这个问题。

## 核心架构

```
d:\ewdk\
├── env.go              # EnvManager 接口 + RegistryEnvManager 实现
│                       #   - Env* 常量组: 11 个环境变量名常量（消除重复字符串）
│                       #   - List/Delete/Set: 系统环境变量 CRUD
│                       #   - runSetupBuildEnv: 执行 SetupBuildEnv.cmd amd64，返回 ewdkEnv
│                       #   - getEwdkDriveLetter: 读 WDKContentRoot 环境变量获取盘符
│                       #   - MountISO/UnmountISO: ISO 挂载卸载
│                       #   - CleanInvalidVars: 清理无效环境变量
├── main.go             # 主入口：完整的挂载→清理→diff→设置→构建流程
│                       #   - setEwdkEnvToSystem: 将 ewdkEnv 写入系统注册表
├── clean.go            # 从 ISO 提取工具链到 dist/ 目录（CI 打包用）
├── ewdk.cmake          # CMake FindWDK 模块（kernel + user-mode 函数）
├── CMakeLists.txt      # 根 CMake（include ewdk.cmake）
├── build.bat           # 一键构建脚本
├── ninja.exe           # Ninja 构建工具
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
    WindowsTargetPlatformVersion string  // WDK 目标平台版本号
    WDKContentRoot               string  // WDK 内容根目录（如 F:\）
    BuildLabSetupRoot            string  // 构建实验室根目录
    VSINSTALLDIR                 string  // VS 安装目录
    INCLUDE                      []string // 头文件搜索路径（含 DIA/KM/UM/Shared/KMDF）
    LIB                          []string // 库文件搜索路径（含 x64）
    WDKBinRoot                   string  // WDK Bin 根目录
    DiaRoot                      string  // DIA SDK 目录
    VCToolsInstallDir            string  // VC 工具安装目录
    CC                           string  // x64 cl.exe 完整路径（强制覆盖）
}
```

### EnvVar — 环境变量条目（含校验信息）

```go
type EnvVar struct {
    Name       string  // 变量名
    Value      string  // 值
    Type       string  // 注册表类型（SZ, EXPAND_SZ 等）
    Valid      bool    // 路径是否存在
    IsPath     bool    // 是否看起来像路径
    IsCompound bool    // 是否为复合路径（PATH, INCLUDE, LIB 等）
    Reason     string  // 无效原因
}
```

### EnvManager 接口

```go
type EnvManager interface {
    List() (EnvVarList, error)                                 // 列出所有系统环境变量，含类型、路径有效性校验
    Delete(name string) error                                  // 删除指定系统环境变量
    Set(name, value string) error                              // 设置/更新系统环境变量（SZ 类型）
    CreateStartupScript(content, name string) (string, error)  // 将内容写入启动目录生成开机自启脚本
    Expand(value string) (string, error)                       // 展开字符串中的 %VAR% 环境变量引用
    CreateScheduledTask(isoPath string) error                  // 创建/更新 EWDK_Mount 计划任务
    DeleteScheduledTask()                                      // 删除 EWDK_Mount 计划任务
    IsMounted(isoPath string) bool                             // 检测指定 ISO 是否已挂载（内部调用 getEwdkDriveLetter 读取 WDKContentRoot）
    MountISO(isoPath string) (string, error)                   // 挂载 ISO 文件，返回盘符
    UnmountISO(isoPath string) error                           // 卸载指定 ISO
    UnmountAll() error                                         // 卸载所有已挂载的 EWDK ISO
    GetVirtualDiskPhysicalPath(isoPath string) (string, error) // 获取 ISO 的物理磁盘路径
    CleanInvalidVars(shouldDelete func(string) bool) int       // 删除所有无效的环境变量
}
```

> **注意**: `GetMountedDriveLetter` 已从接口中移除。盘符检测由 [`getEwdkDriveLetter()`](file:///d:/ewdk/env.go) 直接读取 `WDKContentRoot` 环境变量完成，`IsMounted()` 和 `MountISO()` 内部调用它。

## 标准工作流程（main.go 主逻辑）

当用户执行 `go run .` 时，按以下顺序执行：

### Step 1: 清理无效环境变量（最优先）

**必须最先执行！** 先扫一遍注册表，把所有指向不存在路径的变量删掉：

```go
mgr := NewRegistryEnvManager()
mgr.CleanInvalidVars(func(key string) bool {
    // 受保护变量白名单（保留不删除）
    switch {
    case strings.Contains(key, "_VM_OPTIONS"): return true
    case strings.HasPrefix(key, "VSCODE_GIT"): return true
    case strings.HasPrefix(key, "TRAE_"): return true
    case strings.HasPrefix(key, "VSCMD_"): return true
    // ... 更多白名单规则
    }
})
```

### Step 2: 挂载 ISO

```go
isoPath := resolveISOPath()  // 自动查找当前目录下的 EWDK*.iso 或 CI 环境的 TEMP\ewdk.iso
driveLetter, _ := mgr.MountISO(isoPath)
// MountISO 内部：
//   1. IsMounted() → getEwdkDriveLetter() 读 WDKContentRoot 检测是否已挂载
//   2. 未挂载则 PowerShell Mount-DiskImage
//   3. CreateScheduledTask() 创建开机自动挂载任务
```

### Step 3: 执行 SetupBuildEnv.cmd（**已内置 amd64 参数**）

```go
setupEnvCmd := driveLetter + ":\\BuildEnv\\SetupBuildEnv.cmd"
all := runSetupBuildEnv(setupEnvCmd)
// 内部执行: cmd /c "F:\BuildEnv\SetupBuildEnv.cmd amd64 && set > tmpFile"
// 返回 ewdkEnv 结构体，包含所有关键环境变量
```

> **关键设计决策**: `amd64` 参数在 [runSetupBuildEnv](file:///d:/ewdk/env.go) 函数内部硬编码，调用者无需关心。
> 不传 `amd64` 会导致 SetupBuildEnv.cmd 默认走 x86，PATH 首位是 32 位工具链。

### Step 4: 将 ewdkEnv 写入系统环境变量（永久生效）

```go
setEwdkEnvToSystem(mgr, all)
```

[`setEwdkEnvToSystem()`](file:///d:/ewdk/main.go) 将以下 11 个变量写入系统注册表：

| 变量名 | 来源 | 说明 |
|--------|------|------|
| `WindowsTargetPlatformVersion` | `env.WindowsTargetPlatformVersion` | WDK 目标平台版本 |
| `WDKContentRoot` | `env.WDKContentRoot` | WDK 内容根目录 |
| `BuildLabSetupRoot` | `env.BuildLabSetupRoot` | 构建实验室根目录 |
| `VSINSTALLDIR` | `env.VSINSTALLDIR` | VS 安装目录 |
| `WDKBinRoot` | `env.WDKBinRoot` | WDK Bin 根目录 |
| `DiaRoot` | `env.DiaRoot` | DIA SDK 目录 |
| `VCToolsInstallDir` | `env.VCToolsInstallDir` | VC 工具安装目录 |
| **`CC`** | `filepath.Join(VCToolsInstallDir, "bin\\Hostx64\\x64\\cl.exe")` | 强制 x64 C 编译器 |
| **`CXX`** | 同 CC | C++ 编译器（cl.exe 同时处理） |
| `INCLUDE` | `strings.Join(env.INCLUDE, ";")` | 头文件搜索路径（含 DIA/KM/UM/Shared/KMDF 1.35） |
| `LIB` | `strings.Join(env.LIB, ";")` | 库文件搜索路径（含 x64） |

每个变量写入时带 `[OK]`/`[FAIL]` 状态输出。

### Step 5: 追加 ninja 到 PATH

```go
ninjaDir, _ := filepath.Abs(filepath.Dir("ninja.exe"))
appendNinjaToPATH(mgr, ninjaDir)  // 追加到 PATH 末尾，去重检查
```

### Step 6: 构建项目

```go
if err := stream.RunCommands("build.bat"); err != nil {
    mylog.Check(err)  // 构建失败时 panic 报错
}
```

## runSetupBuildEnv 内部细节

位于 [env.go](file:///d:/ewdk/env.go)，核心逻辑：

1. 创建临时文件 `ewdk-env-after-setup.txt`
2. 执行 `cmd /c "setupCmd amd64 && set > tmpFile"` — **amd64 已硬编码**
3. 解析输出，提取关键字段：`WindowsTargetPlatformVersion`、`VCToolsInstallDir`、`WDKContentRoot`、`VSINSTALLDIR`、`INCLUDE`、`LIB`、`WDKBinRoot`
4. 补充额外路径到 INCLUDE：
   - `DIA SDK\include`
   - `WDKBinRoot\km` / `km\crt` / `um` / `shared`
   - `WDKContentRoot\Include\wdf\kmdf\1.35`
5. 补充额外路径到 LIB：
   - `DIA SDK\lib`
   - `WDKBinRoot\km\x64` / `um\x64`
6. 从 `VCToolsInstallDir` 拼接 x64 `cl.exe` 路径作为 `CC`
7. 返回完整的 `ewdkEnv` 结构体

## CMake 使用方式

### 用户工程只需 3 行 CMakeLists.txt

```cmake
cmake_minimum_required(VERSION 3.16)
project(MyDriver)

# 引入 ewdk.cmake（依赖 $ENV{WDKContentRoot} 已设好）
list(APPEND CMAKE_MODULE_PATH "<path_to_ewdk_repo>")
include(ewdk)

wdk_add_driver(MyDriver KMDF 1.15 main.c)
```

### ewdk.cmake 提供的函数

| 函数 | 用途 | 输出 |
|------|------|------|
| `wdk_add_driver(target [KMDF ver] srcs...)` | 内核驱动 | `.sys` |
| `wdk_add_library(target [STATIC\|SHARED] [KMDF ver] srcs...)` | 内核库 | `.lib` / `.dll` |
| `wdk_add_executable(target [SUBSYSTEM CONSOLE\|WIN] srcs...)` | 用户态 EXE | `.exe` |
| `um_library(target srcs...)` | 用户态静态库 | `.lib` |
| `um_dll(target srcs...)` | 用户态 DLL | `.dll` |

### ewdk.cmake 依赖的环境变量

| 变量 | 来源 | 用途 |
|------|------|------|
| `WDKContentRoot` | setEwdkEnvToSystem 设置 | 定位 WDK 头文件/库/签名工具 |
| `INCLUDE` | setEwdkEnvToSystem 设置 | 含完整 WDK 路径的头文件搜索路径 |
| `LIB` | setEwdkEnvToSystem 设置 | 含 x64 的库文件搜索路径 |
| `CC` / `CXX` | setEwdkEnvToSystem 设置 | x64 cl.exe 编译器 |

## CI 流程 (.github/workflows/ci.yml)

1. Checkout → 安装 Ninja → 安装 aria2 → 获取 EWDK ISO URL → 下载 ISO
2. **挂载 ISO + 设置环境变量 + 构建** (`go run .` 即 main.go，内部依次执行：清理→MountISO→runSetupBuildEnv→setEwdkEnvToSystem→appendNinjaToPATH→build.bat)

## 常见操作命令

```bash
# 构建整个项目（包括 demo）
go build -o ewdk.exe .
./ewdk.exe                    # 执行主流程：清理→挂载→setupEnv→设置变量→构建

# 仅运行测试
go test -v ./...

# 手动挂载/卸载
go run . mount  <iso_path>    # 挂载并设置环境变量
go run . unmount <iso_path>   # 卸载并清理环境变量

# 清理无效变量
go run . clean                # 删除所有指向不存在路径的环境变量
```

## 注意事项

1. **`amd64` 已内置于 `runSetupBuildEnv`**：无需调用者手动拼接参数，函数内部硬编码 `setupCmd + " amd64"`
2. **EWDK 不设 CC/CXX**：`runSetupBuildEnv` 内部从 `VCToolsInstallDir` 自动拼接 `bin\Hostx64\x64\cl.exe` 作为 CC 和 CXX
3. **`GetMountedDriveLetter` 已移除**：盘符检测由 `getEwdkDriveLetter()` 读取 `WDKContentRoot` 环境变量完成，不再需要 Win32 API 遍历驱动器
4. **环境变量写入的是系统级注册表**（`HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`），重启/新终端均生效
5. **ninja.exe 在项目根目录**，`appendNinjaToPATH` 会将其目录追加到 PATH（去重检查）
6. **build.bat 应尽可能简化**，因为环境变量已由 Go 程序永久设好，bat 只需负责 cmake 调用
7. **`setEwdkEnvToSystem` 写入 11 个变量**，空值自动跳过，每条带 `[OK]`/`[FAIL]` 状态输出
8. **环境变量名使用 `Env*` 常量**（定义于 [env.go](file:///d:/ewdk/env.go)）：`EnvWDKContentRoot`、`EnvVCToolsInstallDir`、`EnvINCLUDE` 等 11 个，`runSetupBuildEnv`、`setEwdkEnvToSystem`、`getEwdkDriveLetter` 共享同一套常量，消除重复字符串
