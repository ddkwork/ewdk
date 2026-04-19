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
│                       #   - List/Delete/Set: 系统环境变量 CRUD
│                       #   - CaptureDiff: 执行 SetupBuildEnv.cmd，返回完整 diff
│                       #   - FillCMake: 从新增变量推断 CMake 相关映射
│                       #   - MountISO/UnmountISO: ISO 挂载卸载
│                       #   - CleanInvalidVars: 清理无效环境变量
├── mount-task.go       # 主入口：完整的挂载→清理→diff→设置→构建流程
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

### EnvDiff — 带完整值的差异结果（不是裸 key 列表）

```go
type EnvVarDelta struct {
    Name     string  // 变量名
    OldValue string  // 改之前的值
    NewValue string  // 改之后的值
}

type EnvDiff struct {
    Added   map[string]string  // name → new_value     新增变量（直接喂 cmake）
    Changed []EnvVarDelta      // 变更详情（old+new）   同步更新时知道改了啥
    Removed map[string]string  // name → old_value     被删变量的最后值
}
```

### EnvManager 接口

```go
type EnvManager interface {
    List() (EnvVarList, error)
    Delete(name string) error
    Set(name, value string) error
    CaptureDiff(setupCmd string) (*EnvDiff, error)
    FillCMake(newVars map[string]string) (map[string]string, error)
    CreateStartupScript(content, name string) (string, error)
    Expand(value string) (string, error)
    GetMountedDriveLetter(isoPath string) string
    MountISO(isoPath string) (string, error)
    UnmountISO(isoPath string) error
    CleanInvalidVars() (int, error)
}
```

## 标准工作流程（mount-task.go 主逻辑）

当用户需要初始化/更新 EWDK 环境时，按以下顺序执行：

### Step 1: 确定 ISO 路径

```go
githubWorkspace := os.Getenv("GITHUB_WORKSPACE")
if githubWorkspace != "" {
    isoPath = filepath.Join(os.Getenv("TEMP"), "ewdk.iso")  // CI 环境
} else {
    isoPath = "d:\\ewdk\\EWDK_br_release_28000_251103-1709.iso"  // 本地开发
}
```

### Step 1: 清理无效环境变量（最优先）

**必须最先执行！** 先扫一遍注册表，把所有指向不存在路径的变量删掉。从 log.log 可以看到注册表里堆积了大量垃圾：

```
"AQUA_VM_OPTIONS", "CLION_VM_OPTIONS", "DATAGRIP_VM_OPTIONS", "GOLAND_VM_OPTIONS",
"IDEA_VM_OPTIONS", "JETBRAINSCLIENT_VM_OPTIONS", "PHPSTORM_VM_OPTIONS",
"PYCHARM_VM_OPTIONS", "RIDER_VM_OPTIONS", "RUBYMINE_VM_OPTIONS", ...
```

这些是 IDE 反复写入的残留，每次挂载/设置都会产生新的一批。先清理再操作，避免垃圾越积越多。

```go
count, _ := mgr.CleanInvalidVars()  // 删除所有指向不存在路径的环境变量
```

### Step 2: 删除已知干扰变量

```go
mgr.Delete("INCLUDE")  // 必须删除，否则 SetupBuildEnv 会追加而非替换
mgr.Delete("LIB")      // 同上
```

### Step 3: 挂载 ISO（如未挂载）

```go
driveLetter, err := mgr.MountISO(isoPath)
// MountISO 内部会：
//   1. 检测是否已挂载（IsMounted + GetMountedDriveLetter）
//   2. 未挂载则调用 Mount-DiskImage
//   3. 自动设置 WDKContentRoot / WDK_ROOT / EWDKSetupEnvCmd
//   4. 创建计划任务实现开机自动挂载（CreateScheduledTask）
```
### Step 5: 执行 SetupBuildEnv.cmd 并捕获 diff
```go
setupEnvCmd, _ := mgr.GetEWDKSetupEnvCmd()  // 如 "F:\BuildEnv\SetupBuildEnv.cmd"
diff, err := mgr.CaptureDiff(setupEnvCmd + " amd64")  // 关键：必须传 amd64，否则默认 x86
// diff.Added   → { "INCLUDE": "F:\\...\\Include\\...", "LIB": "F:\\...\\Lib\\...", "VCToolsInstallDir": "...", ... }
// diff.Changed → [{Name:"PATH", OldValue:"...", NewValue:"...;F:\\...\\bin\\Hostx64\\x64;..."}, ... }
// diff.Removed → 被删的变量
```

**关键：必须传 `amd64` 参数！** 不传的话 SetupBuildEnv.cmd 默认走 x86，导致：
- PATH 首位是 `HostX86\x86`（32位 cl.exe）
- LIB 混杂 x86 和 x64 路径
- CC/CXX 不会被 EWDK 设置（仍然是 gcc/g++）
- INCLUDE 缺少 WDK 的 um/ucrt/shared/km 路径

### Step 6: 将 diff 写入系统环境变量（永久生效）

```go
// 5a. 设置新增变量到系统注册表
for name, value := range diff.Added {
    mgr.Set(name, value)
}

// 5b. 更新变更变量
for _, delta := range diff.Changed {
    mgr.Set(delta.Name, delta.NewValue)
}
```

### Step 7: 填充 CMake 变量（FillCMake）

```go
cmakeVars, _ := mgr.FillCMake(diff.Added)
for name, value := range cmakeVars {
    mgr.Set(name, value)  // CMAKE_INCLUDE_PATH, CMAKE_LIBRARY_PATH, CC, CXX 等
}
```

`FillCMake` 从 `diff.Added` 中推断 CMake 相关映射：
- `INCLUDE` → `CMAKE_INCLUDE_PATH`
- `LIB` → `CMAKE_LIBRARY_PATH`
- `WDK_ROOT`/`WDKContentRoot` → `CMAKE_PREFIX_PATH`
- `EWDKSetupEnvCmd` → 推断 `CMAKE_TOOLCHAIN_FILE`

### Step 8: 强制设置 x64 编译器（最关键的一步）

EWDK 的 SetupBuildEnv.cmd **不会设置 CC/CXX**，它们在 diff 中可能仍然是 `gcc`/`g++`。必须强制覆盖：

```go
vcToolsDir := addedVars["VCToolsInstallDir"]  // 如 "F:\...\MSVC\14.44.35207\"
clExe := filepath.Join(vcToolsDir, "bin", "Hostx64", "x64", "cl.exe")

mgr.Set("CC", clExe)              // C 编译器
mgr.Set("CXX", clExe)             // C++ 编译器（cl.exe 同时处理两者）
mgr.Set("CMAKE_C_COMPILER", clExe)   // CMake 显式指定
mgr.Set("CMAKE_CXX_COMPILER", clExe) // CMake 显式指定
```

同时清理 PATH，**移除所有 `HostX86\x86` 条目**，确保 x64 工具链优先：

```go
// 从 PATH 中剔除 HostX86\x86，保留 Hostx64\x64
parts := strings.Split(currentPath, ";")
for _, part := range parts {
    if strings.Contains(part, "HostX86\\x86") { continue }  // 删掉 x86
    if strings.Contains(part, "Hostx64\\x64") { keep }      // 保留 x64
}
```

### Step 9: 追加 ninja 到 PATH + 构建

```go
ninjaDir := filepath.Abs(filepath.Dir("ninja.exe"))
// 追加到 PATH 末尾
mgr.Set("PATH", currentPath + ";" + ninjaDir)

// 构建
exec.Command("cmake", "-B", "build", "-G", "Ninja", "-DCMAKE_BUILD_TYPE=Release", ".")
exec.Command("cmake", "--build", "build", "--config", "Release")
```

### Step 10: 最终环境变量健康检查（收尾）

**所有操作完成后必须执行！** 整个流程（清理→挂载→diff→设置→构建）会产生新的垃圾变量或残留。这一步做最终稽查：

```go
func finalCheck(mgr EnvManager) {
    // 1. List() 全量扫描注册表
    allVars, _ := mgr.List()
    
    // 2. 筛选 Invalid && IsPath 的变量
    var invalidVars []EnvVar
    for _, v := range allVars {
        if !v.Valid && v.IsPath {
            invalidVars = append(invalidVars, v)
        }
    }
    
    // 3. 分类处理（大小写不敏感匹配）
    for _, v := range invalidVars {
        if isCompoundPath(v.Name) {
            // 复合路径变量：清理无效子路径，保留有效部分
            cleanedVal := cleanCompoundPath(v.Value)
            if cleanedVal != v.Value {
                mgr.Set(v.Name, cleanedVal)  // [FIX]
            } else if isProtected(v.Name) {
                // 整条都无效但受保护：跳过
                continue  // [SKIP]
            } else {
                mgr.Delete(v.Name)  // [DEL]
            }
        } else if !isProtected(v.Name) {
            mgr.Delete(v.Name)  // [DEL]
        } else {
            continue  // [SKIP]
        }
    }
    
    // 4. 稽查计划任务：删除过期的 EWDK_Mount
    checkScheduledTasks(mgr)
}
```

**三类处理策略（按优先级）：**

| 类型 | 示例 | 策略 | 说明 |
|------|------|------|------|
| **复合路径变量** | PATH, LIB, INCLUDE, PSModulePath, CMAKE_*, SAFE_RM_* | `cleanCompoundPath()` 逐段校验，删无效子路径，保留有效部分 | 绝不能整条删除 |
| **受保护变量** | HOMEPATH, HOME, USERPROFILE, SystemRoot, windir | 跳过不操作 | Windows 系统必需或相对路径 |
| **普通垃圾变量** | ToolsPathARCH, TRAE_*, *_VM_OPTIONS, VSCMD_ARG_* | 直接 Delete | 无残留价值 |

**为什么需要这一步：**
- SetupBuildEnv.cmd 可能写入 `ToolsPathARCH`、`Platform=x86`、`VSCMD_ARG_*` 等一次性变量
- IDE 反复写入的 `*_VM_OPTIONS` 残留（路径已不存在）
- `F:\Program Files\Windows Kits\10\Tools\10.0.28000.0\x64` 这类路径可能在 ISO 卸载后失效
- 旧的 `EWDK_Mount` 计划任务可能指向错误的 ISO 路径
- Trae IDE 写入的临时变量（TRAE_JWT_TOKEN_PATH 等）

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
| `WDKContentRoot` | mount-task.go 设置 | 定位 WDK 头文件/库/签名工具 |
| `PATH` | diff 中 Added/Changed | 包含 `cl.exe`, `link.exe`, `lib.exe` 等工具路径 |

## CI 流程 (.github/workflows/ci.yml)

1. Checkout → 安装 Ninja → 获取 EWDK ISO URL → 下载 ISO
2. **挂载 ISO** (`mount-task.ps1`)
3. **CMake 构建**（此时环境变量已由 mount-task.ps1 设好）

## 常见操作命令

```bash
# 构建整个项目（包括 demo）
go build -o ewdk.exe .
./ewdk.exe                    # 执行主流程：挂载→清理→diff→设置→构建

# 仅运行测试
go test -v ./...

# 手动挂载/卸载
go run . mount  <iso_path>    # 挂载并设置环境变量
go run . unmount <iso_path>   # 卸载并清理环境变量

# 清理无效变量
go run . clean                # 删除所有指向不存在路径的环境变量
```

## 注意事项

1. **必须先 Delete("INCLUDE") 和 Delete("LIB")** 再执行 CaptureDiff，否则 SetupBuildEnv.cmd 会将新路径**追加**到旧值后面，导致路径混乱
2. **CaptureDiff 必须传 `amd64` 参数**：`CaptureDiff(setupCmd + " amd64")`，不传默认 x86，CC/CXX 不会被设置，PATH 首位是 32 位工具
3. **EWDK 不设 CC/CXX**：SetupBuildEnv.cmd 执行后 CC/CXX 可能仍是 gcc/g++，必须从 `VCToolsInstallDir` 拼出 x64 cl.exe 路径强制覆盖
4. **必须清理 PATH 中的 HostX86\x86**：amd64 模式下 PATH 仍可能包含 x86 工具链路径，需剔除以确保 cmake 找到正确的 64 位 cl.exe
5. **CaptureDiff 返回的是带值的 EnvDiff**，不是裸 key 列表。Added 直接喂给 FillCMake/系统注册表，Changed 包含 old/new 用于审计
6. **环境变量写入的是系统级注册表**（`HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`），重启/新终端均生效
7. **ninja.exe 在项目根目录**，设置环境变量时应将其目录追加到 PATH
8. **build.bat 应尽可能简化**，因为环境变量已由 Go 程序永久设好，bat 只需负责 cmake 调用
9. **最终必须执行 finalCheck()**：全量 List() 扫描注册表，删除所有 Invalid+IsPath 的变量 + 清理过期计划任务。整个流程会产生新垃圾（ToolsPathARCH、Platform、VSCMD_ARG_* 等），不收尾会越积越多
