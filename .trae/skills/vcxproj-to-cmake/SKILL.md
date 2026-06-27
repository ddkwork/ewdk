---
name: "vcxproj-to-cmake"
description: "Scans directories for .vcxproj files and generates ewdk-compatible CMakeLists.txt. Invoke when user provides a directory path containing .vcxproj projects and needs CMakeLists.txt created."
---

# vcxproj → CMakeLists.txt 生成技能

根据给定的目录路径，扫描目录下所有 `.vcxproj` 文件，为每个 vcxproj 工程生成对应的 ewdk 体系 CMakeLists.txt。

## 模板来源

所有 ewdk 函数模板定义在以下文件中，生成 CMakeLists.txt 前必须阅读确认：

| 文件 | 用途 |
|------|------|
| `ewdk/main.go`（在工作空间中搜索） | **模板源码**：所有 `um_exe`/`um_dll`/`um_lib`/`km_sys`/`km_lib`/`um_dp64`/`um_dp86`/`um_exe_x86`/`um_dll_x86`/`um_lib_x86`/`um_exe_mfc_x86` 等函数在此生成，可搜索 `function(um_exe` 等定位 |
| `C:/Program Files/CMake/bin/ewdk.cmake` | **生成产物**：main.go 生成的完整 cmake 模块，所有函数实现在此，可搜索 `function(um_exe` 定位参数签名。**生成前必须阅读**，避免重复添加已有的 lib 和设置 |
| `C:/Program Files/CMake/bin/unity.cmake` | **合并编译辅助模块**：提供 `collect_sources()` + `generate_unity()`，用于加速编译。**生成前必须阅读**，了解是否可用 |
| `ewdk/构建.cmd`（在工作空间中搜索） | 编译 main.go 为 `ewdk.exe`（内容：`go build -v .`） |

## 参考写法（内置示例）

开始生成前阅读以下内置示例，避免依赖外部文件路径。

### 示例 A：多目标 + GLOB + asm + DEF（参考 al-khaser 风格）

```cmake
cmake_minimum_required(VERSION 3.30)
include("C:/Program Files/CMake/bin/ewdk.cmake")
project(al-khaser LANGUAGES C CXX ASM_MASM)

# GLOB 收集所有源文件，排除子项目/预编译头/CMakeFiles
file(GLOB_RECURSE alkhaser_SOURCES "*.cpp")
list(FILTER alkhaser_SOURCES EXCLUDE REGEX "/InjectedDLL/|pch\\.cpp$|CMakeFiles/")

# ─── x64 ───
um_exe(al-khaser SUBSYSTEM CONSOLE
    SOURCES
        ${alkhaser_SOURCES}
        int2d_x64.asm
        AntiDisassm_x64.asm
    INCLUDES
        .
        AntiDebug
        AntiVM
        Shared
    DEFINES
        _CONSOLE
        UNICODE
)
um_dll(InjectedDLL
    SOURCES CodeInjection/InjectedDLL/InjectedDLL.cpp
    DEFINES
        _USRDLL
        INJECTEDDLL_EXPORTS
        UNICODE
)
target_link_options(InjectedDLL PRIVATE "/DEF:${CMAKE_CURRENT_SOURCE_DIR}/CodeInjection/InjectedDLL/definitions.def")

um_exe(ATAIdentifyDump SUBSYSTEM CONSOLE
    SOURCES ../Tools/ATAIdentifyDump/ATAIdentifyDump.cpp
    INCLUDES ../Tools/ATAIdentifyDump
    DEFINES _CONSOLE
)

# ─── x86 ───
um_exe_x86(al-khaser_x86
    SOURCES
        ${alkhaser_SOURCES}
        int2d_x86.asm
        AntiDisassm_x86.asm
    INCLUDE_DIRS
        .
        AntiDebug
        AntiVM
        Shared
    DEFINITIONS
        _CONSOLE
        UNICODE
)
um_dll_x86(InjectedDLL_x86
    SOURCES CodeInjection/InjectedDLL/InjectedDLL.cpp
    DEFINITIONS
        _USRDLL
        INJECTEDDLL_EXPORTS
        UNICODE
)
um_exe_x86(ATAIdentifyDump_x86
    SOURCES ../Tools/ATAIdentifyDump/ATAIdentifyDump.cpp
    INCLUDE_DIRS ../Tools/ATAIdentifyDump
    DEFINITIONS _CONSOLE
)
```

关键要点：
- `file(GLOB_RECURSE)` + `list(FILTER EXCLUDE REGEX)` 收集源文件
- x64/x86 **两块独立代码块**，零 if
- asm 文件每行一个，`${SOURCES}` 变量和 asm 文件混排
- DLL 用 `um_dll`/`um_dll_x86`，`.def` 文件用 `target_link_options(... "/DEF:...")`
- 单条目同行，多条目换行缩进

### 示例 B：插件 + 合并编译（参考 x64dbg_mcp 风格）

```cmake
cmake_minimum_required(VERSION 3.30)
include("C:/Program Files/CMake/bin/ewdk.cmake")
project(x64dbg_mcp LANGUAGES C CXX)

# 收集各模块源文件
collect_sources(src/bridge src/tools src/handlers PLUGIN_SOURCES)

# 合并编译加速（如遇 static 冲突则切回逐个文件）
generate_unity(unity ${PLUGIN_SOURCES})

# ─── x64 ───
um_dp64(x64dbg_mcp
    SOURCES ${UNITY_FILE}
    LINK_LIBS x64bridge.lib x64dbg.lib
)

# ─── x86 ───
um_dp86(x64dbg_mcp
    SOURCES ${UNITY_FILE}
    LINK_LIBS x64bridge.lib x64dbg.lib
)
```

关键要点：
- **`unity.cmake` 由 `ewdk.cmake` 自动引入**，CMakeLists.txt 中无需手动 include
- `collect_sources()` 收集多个子目录的源文件
- `generate_unity(unity ${SOURCES})` 生成 `unity.cpp`
- SOURCES 中只用 `${UNITY_FILE}`，替代逐个源文件
- `um_dp64`/`um_dp86` 插件函数，`LINK_LIBS` 传库名

## 更新模板

如果 `main.go` 中的函数模板需要修改（如新增函数、修改参数、修复 bug），按以下步骤执行：

### 修改 main.go

在 `d:\ux\examples\ewdk\main.go` 中找到对应函数模板（搜索 `function(um_exe` 或 `function(km_sys` 定位），修改模板字符串。

### 重新生成 ewdk.cmake

模板修改后，需要重新编译并执行以更新磁盘上的 `ewdk.cmake`：

```bash
cd d:\ux\examples\ewdk
go build -v .
.\ewdk.exe
```

这会重新编译 `ewdk.exe` 并运行，生成最新的 `C:/Program Files/CMake/bin/ewdk.cmake`。

### 验证

在目标项目目录执行构建验证新模板生效：

```cmd
(cmake -B Release -G "Ninja" -DCMAKE_BUILD_TYPE=Release . && cmake --build Release --config Release) 2>&1 | powershell -NoProfile -Command "$input | Tee-Object -FilePath build.Release.log"
(cmake -B Debug -G "Ninja" -DCMAKE_BUILD_TYPE=Debug . && cmake --build Debug --config Debug) 2>&1 | powershell -NoProfile -Command "$input | Tee-Object -FilePath build.Debug.log"
```

## 生成 build.cmd

生成 CMakeLists.txt 后，在**同一目录**创建 `build.cmd`，内容就是上面的验证命令，这样用户一键即可构建 Release + Debug 并记录日志。

## 生成辅助文件

生成 CMakeLists.txt 和 build.cmd 后，根据项目类型在**同一目录**生成以下辅助文件。

### 通用辅助文件（所有项目类型）

以下三个文件在每个项目中都生成，内容直接从技能**自身目录**复制（它们与 `SKILL.md` 在同一目录，永久有效）：

| 文件 | 用途 |
|------|------|
| `.clang-format` | 代码格式化配置（Google 风格缩进 4，列宽 120） |
| `format_all.cmd` | 一键格式化所有 C/C++ 源文件 |
| `convert_encoding.py` | GBK→UTF-8 无 BOM 编码转换 |

**复制后必须执行**：
- 运行 `convert_encoding.py`：`python convert_encoding.py`，将目录下所有 GBK/GB2312 编码的源文件转换为 UTF-8 无 BOM
- 运行 `format_all.cmd`：双击或执行 `format_all.cmd`，用 clang-format 格式化所有 C/C++ 源文件
- 这两个文件**不能只是复制过去放着**，必须实际执行以确保代码格式和编码统一

### km_sys 内核驱动专属：install.bat / uninstall.bat

如果项目类型是内核驱动（`km_sys`），额外生成安装/卸载脚本。

#### `install.bat` / `uninstall.bat` — 安装/卸载驱动

内容直接从技能目录复制（与 `SKILL.md` 同目录），**不要修改**。

## 工作流程

### Step 1: 阅读参考文件

开始前必须阅读以下文件，避免重复添加已有 lib/设置，并参考正确写法：

1. `C:/Program Files/CMake/bin/ewdk.cmake` — 了解哪些 lib 已自动链接、有哪些函数可用
2. `C:/Program Files/CMake/bin/unity.cmake` — 了解是否可用合并编译加速
3. `d:\ux\examples\ewdk\debuger\HyperHide\al-khaser\al-khaser\CMakeLists.txt` — al-khaser 完整写法参考
4. `d:\ux\examples\ewdk\debuger\x64dbg_mcp\plugin\CMakeLists.txt` — 插件 + unity 写法参考

### Step 2: 确认目录

用户提供目标目录路径，确认目录存在。

### Step 3: 扫描 vcxproj

用 `Glob` 或 `LS` 查找目标目录下所有 `.vcxproj` 文件：

```
*.vcxproj
```

### Step 4: 解析 vcxproj（读取关键 XML 节点）

每个 `.vcxproj` 是 XML 文件，需要读取以下节点：

#### 3.1 配置平台 (`ProjectConfiguration`)

```xml
<ProjectConfiguration Include="Release|x64">
  <Configuration>Release</Configuration>
  <Platform>x64</Platform>
</ProjectConfiguration>
```

确定项目的目标架构（x64 / x86 / ARM64 等）。

#### 3.2 输出类型 (`ConfigurationType`)

```xml
<ConfigurationType>Application</ConfigurationType>
<!-- Application = .exe -->
<!-- DynamicLibrary = .dll -->
<!-- StaticLibrary = .lib -->
```

决定调用哪个 ewdk 函数：

| ConfigurationType | ewdk 函数 |
|-------------------|-----------|
| `Application` (x64) | `um_exe` → `.exe` |
| `Application` (x86) | `um_exe_x86` → `.exe` |
| `Application` (x86, MFC) | `um_exe_mfc_x86` → `.exe` |
| `DynamicLibrary` (x64) | `um_dll` → `.dll` |
| `DynamicLibrary` (x86) | `um_dll_x86` → `.dll` |
| `StaticLibrary` (x64) | `um_lib` → `.lib` |
| `StaticLibrary` (x86) | `um_lib_x86` → `.lib` |
| `Driver` (kernel) | `km_sys` → `.sys` |
| kernel `StaticLibrary` | `km_lib` → `.lib` |
| x64dbg 插件 (x64) | `um_dp64` → `.dp64` |
| x64dbg 插件 (x86) | `um_dp86` → `.dp32` |

#### 3.3 子系统 (`SubSystem`)

```xml
<SubSystem>Console</SubSystem>
<!-- Console / Windows / Native -->
```

决定 `um_exe` 的 `SUBSYSTEM` 参数。

#### 3.4 源文件 (`ClCompile`)

```xml
<ItemGroup>
  <ClCompile Include="AntiDebug\Interrupt_0x2d.cpp" />
  <ClCompile Include="AntiDebug\Interrupt_3.cpp" />
</ItemGroup>
```

收集所有 `.cpp` / `.c` 源文件路径（相对于 vcxproj 所在目录）。

**注意**：只收集 `ClCompile`，不处理 `ClInclude`（头文件不需要列在 SOURCES 中）。

#### 3.5 MASM 汇编文件 (`MASM`)

```xml
<ItemGroup>
  <MASM Include="AntiDebug\int2d_x64.asm">
    <ExcludedFromBuild Condition="'$(Configuration)|$(Platform)'=='Debug|x86'">true</ExcludedFromBuild>
    <ExcludedFromBuild Condition="'$(Configuration)|$(Platform)'=='Release|x86'">true</ExcludedFromBuild>
  </MASM>
  <MASM Include="AntiDebug\int2d_x86.asm">
    <ExcludedFromBuild Condition="'$(Configuration)|$(Platform)'=='Debug|x64'">true</ExcludedFromBuild>
    <ExcludedFromBuild Condition="'$(Configuration)|$(Platform)'=='Release|x64'">true</ExcludedFromBuild>
  </MASM>
</ItemGroup>
```

根据 `ExcludedFromBuild` 判断架构归属：
- 条件含 `x86` 的 ExcludedFromBuild → 该 asm 文件是 x64（x86 排除）
- 条件含 `x64` 的 ExcludedFromBuild → 该 asm 文件是 x86（x64 排除）
- 无 ExcludedFromBuild → 两架构共有

#### 3.6 包含目录 (`AdditionalIncludeDirectories`)

```xml
<AdditionalIncludeDirectories>%(AdditionalIncludeDirectories);$(SolutionDir)Shared;$(SolutionDir)AntiDebug</AdditionalIncludeDirectories>
```

收集 include 目录路径。

#### 3.7 预处理器定义 (`PreprocessorDefinitions`)

```xml
<PreprocessorDefinitions>_CONSOLE;UNICODE;_UNICODE;%(PreprocessorDefinitions)</PreprocessorDefinitions>
```

收集 define 宏。

#### 3.8 链接库 (`AdditionalDependencies`)

```xml
<AdditionalDependencies>ole32.lib;oleaut32.lib;%(AdditionalDependencies)</AdditionalDependencies>
```

收集额外链接库。

> **注意**：通用系统 lib 和 WDK 专有 lib 已集成到 `ewdk.cmake` 的 `WDK_UM_SDK_LIBS` 和 `WDK_UM_EXTRA_LIBS` 变量中（搜索这两个变量名确认具体列表）。生成前先对比 vcxproj 的 `AdditionalDependencies` 和这两个变量，**只有 ewdk.cmake 未包含的 lib** 才写入 `LIBS`/`LINK_LIBS` 参数。

### Step 4: 生成 CMakeLists.txt

按以下模板生成。

#### 4.0 格式化规则

- **单条目**：与关键字同行（`SOURCES file.cpp`、`DEFINES _CONSOLE`、`INCLUDES dir`）
- **多条目**：每个条目单独一行，缩进 8 空格（如上例 SOURCES 的写法）
- **asm 文件**：`${SOURCES}` 变量和 asm 文件混排，每个一行

#### 4.1 基础模板

```cmake
cmake_minimum_required(VERSION 3.30)
include("C:/Program Files/CMake/bin/ewdk.cmake")
project(<target-name> LANGUAGES C CXX ASM_MASM)

# ─── x64 ───
um_exe(<target> SUBSYSTEM CONSOLE
    SOURCES
        file1.cpp
        file2.cpp
        x64_asm.asm
    INCLUDES
        dir1
        dir2
    DEFINES
        _CONSOLE
        UNICODE
)

# ─── x86 ───
um_exe_x86(<target>_x86
    SOURCES
        file1.cpp
        file2.cpp
        x86_asm.asm
    INCLUDE_DIRS
        dir1
        dir2
    DEFINITIONS
        _CONSOLE
        UNICODE
)
```

#### 4.2 架构规则

ewdk 已经分好了两套架构模板，**禁止使用 `if(TARGET_ARCH ...)` 或任何架构判断条件**。直接写两块：

| 块 | 函数命名 | 参数名差异 |
|----|----------|-----------|
| `# ─── x64 ───` | `um_exe` / `um_lib` / `um_dll` | `INCLUDES` / `DEFINES` / `LIBS` |
| `# ─── x86 ───` | `um_exe_x86` / `um_lib_x86` / `um_dll_x86` | `INCLUDE_DIRS` / `DEFINITIONS` / `LINK_LIBS` |

#### 4.3 仅 x64 项目

如果 vcxproj 只有 x64 配置，只生成 `# ─── x64 ───` 块，不生成 x86。

#### 4.4 架构特定 asm 文件

asm 文件**不要用 `file(GLOB)` 收集**，直接列在对应架构的 `SOURCES` 中：

```cmake
# x64 SOURCES
    ${SOURCES}
    int2d_x64.asm
    AntiDisassm_x64.asm

# x86 SOURCES
    ${SOURCES}
    int2d_x86.asm
    AntiDisassm_x86.asm
```

#### 4.5 空 SOURCES 处理

如果没有源文件（如只有非 ClCompile 内容），跳过该架构块。

### Step 5: 源文件收集 —— 优先 GLOB

对于已有大量文件的成熟项目，**优先使用 `file(GLOB_RECURSE`** 收集源码，然后手动列出架构特定的 asm 文件：

```cmake
file(GLOB_RECURSE target_SOURCES "*.cpp")
list(FILTER target_SOURCES EXCLUDE REGEX "/path/to/exclude/|pch\\.cpp$|CMakeFiles/")

um_exe(target SUBSYSTEM CONSOLE
    SOURCES
        ${target_SOURCES}
        arch_x64.asm
    INCLUDES ...
    DEFINES ...
)
```

对排除目录：
- 用 `/specific_dir_name/` 格式排除子目录
- 排除 `pch.cpp` / `stdafx.h` 等预编译头文件

### Step 6: 优先合并编译（大量源文件时）

如果项目源文件较多（如 50+ 个 .cpp），第一次编译会很慢。**优先考虑 `unity.cmake` 的合并编译方案**来加速：

```cmake
include("C:/Program Files/CMake/bin/ewdk.cmake")

file(GLOB_RECURSE target_SOURCES "*.cpp")
list(FILTER target_SOURCES EXCLUDE REGEX "pch\\.cpp$|CMakeFiles/")

# 合并编译加速
collect_sources(src/subdir1 src/subdir2 COLLECTED_SOURCES)
generate_unity(unity ${COLLECTED_SOURCES})

um_exe(target SUBSYSTEM CONSOLE
    SOURCES ${UNITY_FILE}
    ...
)
```

注意：
- `collect_sources()` 收集多个子目录，`generate_unity()` 生成 `unity.cpp`
- 如果遇到 `static` 变量/函数冲突，**不要排除文件**，直接放弃合并编译，退回逐个文件编译
- **不要使用 CMake 原生 `UNITY_BUILD`**，用 `unity.cmake` 手动控制
- x86 交叉编译模式也支持合并编译

### Step 7: 多项目结构（一个目录多 proj）

如果目录包含多个 vcxproj，在单个 CMakeLists.txt 中并列多个目标：

```cmake
cmake_minimum_required(VERSION 3.30)
include("C:/Program Files/CMake/bin/ewdk.cmake")
project(multi-project LANGUAGES C CXX ASM_MASM)

# ─── x64 ───
um_exe(proj1 SUBSYSTEM CONSOLE
    SOURCES ...
    INCLUDES ...
    DEFINES ...
)

um_dll(dll1
    SOURCES ...
    DEFINES ...
)

# ─── x86 ───
um_exe_x86(proj1_x86
    SOURCES ...
    INCLUDE_DIRS ...
    DEFINITIONS ...
)

um_dll_x86(dll1_x86
    SOURCES ...
    DEFINITIONS ...
)
```

### Step 8: 特殊处理

- **`.def` 文件**（DLL 导出定义）：用 `target_link_options(target PRIVATE "/DEF:${path}")`
- **预编译头**（`pch.cpp`、`stdafx.cpp`）：从 SOURCES 排除（ewdk 在 CI 下不支持 PCH）
- **多配置**（Release/Debug）：CMake 本身管理配置，CMakeLists.txt 只列源文件、include 和 lib
- **um_dp64 / um_dp86**：x64dbg 插件用这两个函数，`SOURCES` 是唯一必需的参数，lib 用 `LINK_LIBS`

## 注意事项

1. **零 if**：架构区分用两块独立代码，不用 `if(TARGET_ARCH ...)` 或 `if(WIN32)` 等
2. **切勿修改 build.cmd / build.bat**：只生成/修改 CMakeLists.txt
3. **保留原始注释**：生成或修改文件时，不得删除用户原有的任何注释代码（包括被注释掉的代码段），仅可在必要时添加新注释或调整代码结构
4. **切勿修改源文件**（`.cpp`/`.h`）：编译问题通过 CMakeLists.txt 解决，不碰源代码
5. **不要添加 `set_target_properties(... UNITY_BUILD ON/OFF)`**：ewdk 不使用 CMake 原生 unity build
6. **x86 是交叉编译**：x86 函数名带 `_x86` 后缀，参数名与 x64 不同（`INCLUDE_DIRS` / `DEFINITIONS` / `LINK_LIBS`）
7. **通用系统 lib 不用写**：`WDK_UM_SDK_LIBS` + `WDK_UM_EXTRA_LIBS` 已自动链接
8. **只用工程需要的 lib**：如果 vcxproj 引用了非默认的 lib（如 `ntdll.lib`、`crypt32.lib`、`winhttp.lib`、`dbghelp.lib` 等），仍然需要在 `LIBS`/`LINK_LIBS` 中显式列出

## 自动构建与修复

生成 CMakeLists.txt 和 build.cmd 后，必须自动执行构建并修复问题直到编译通过。

### 执行构建

在目标目录下运行 build.cmd：

```cmd
build.cmd
```

或单独运行构建命令（同时观察输出）：

```cmd
cmake -B Release -G "Ninja" -DCMAKE_BUILD_TYPE=Release . && cmake --build Release --config Release
```

### 检查编译输出

- 通过 CMake 错误信息定位问题（语法错误、缺少文件、链接找不到符号等）
- 通过日志文件 `build.Release.log` 和 `build.Debug.log` 查看完整编译输出
- 标记为 Binary file 的日志文件没法读，在目标目录直接用 `Read` 工具读取也不可行，应直接在终端执行 cmake --build 命令看**实时输出**

### 常见编译错误修复

| 错误类型 | 原因 | 修复 |
|----------|------|------|
| `CXX_STANDARD is set to invalid value` | CMake 限制标准值 | 不要手动设置 `CMAKE_CXX_STANDARD`，ewdk.cmake 用编译选项控制标准 |
| `fatal error C1083: Cannot open include file` | include 路径缺失 | 在 `INCLUDES`/`INCLUDE_DIRS` 中添加缺失目录 |
| `unresolved external symbol` | 缺少链接库 | 在 `LIBS`/`LINK_LIBS` 中添加对应 lib（先确认不在 WDK_UM_SDK_LIBS/EXTRA_LIBS 中） |
| `LNK2001 unresolved external symbol _main` | 子系统不匹配 | `SUBSYSTEM CONSOLE` 或 `SUBSYSTEM WINDOWS` |
| `warning STL4038` / D9025 | C++ 标准冲突 | native 目标用 `CXX_STANDARD 17`，x86 custom command 用 `/std:c++latest` |
| `fatal error LNK1120` | .def 文件未找到 | 检查 `target_link_options(... "/DEF:${path}")` 路径是否正确 |
| asm 编译错误 | ml.exe 找不到 | 确认 x86 asm 在 x86 块中，x64 asm 在 x64 块中，as 文件用 `file(GLOB_RECURSE` 收集了 *.asm 再 FILTER 或手动列出 |

### 修复循环

1. 读取 CMake 错误信息
2. 修复 CMakeLists.txt
3. 重新执行构建命令
4. 重复直到 Release 和 Debug 均编译通过（零错误，零警告）