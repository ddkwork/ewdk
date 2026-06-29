---
name: "bsod-analysis"
description: "Analyzes Windows kernel crash dumps (.dmp) with kd.exe and PDB symbols, identifies root cause, applies minimal source fix. Invoke when user provides .sys, .pdb, .dmp files for driver crash analysis."
---

# BSOD 驱动崩溃分析技能

## 固定环境配置

| 项目 | 路径 |
|------|------|
| kd.exe | `C:\Program Files\WindowsApps\Microsoft.WinDbg_1.2603.20001.0_x64__8wekyb3d8bbwe\amd64\kd.exe` |
| 本地符号路径 | `C:\ProgramData\Dbg\sym` |
| 符号服务器 | `srv*`（自动 fallback） |

用户提供文件：
- `.sys` — 驱动二进制
- `.pdb` — 符号文件（放符号路径下，kd 通过 PE 时间戳自动匹配）
- `.dmp` — 蓝屏转储文件（通常为 `C:\Windows\MEMORY.DMP`）

## 分析步骤

### 1. 加载 dump 获取初步信息

```powershell
& "C:\Program Files\WindowsApps\Microsoft.WinDbg_1.2603.20001.0_x64__8wekyb3d8bbwe\amd64\kd.exe" -z "<dmp_path>" -y "C:\ProgramData\Dbg\sym" -c "!analyze -v;q"
```

### 2. 提取关键信息

从 `!analyze -v` 输出提取：
- **BUGCHECK_CODE** — 如 `0x7E`, `0xD1`, `0x3B`, `0x19`
- **IMAGE_NAME** — 哪个模块崩溃
- **STACK_TEXT** — 完整调用栈
- **FAULTING_SOURCE_LINE** — 源码位置（需 PDB 匹配）
- **FAILURE_BUCKET_ID** — 微软错误桶

### 3. 深度分析

根据栈回溯，用反汇编定位精确崩溃点：

```powershell
# 反汇编关键函数
uf <module>!<function>

# 查看调用栈寄存器
knL
```

### 4. 修复原则

- **最小修复**：只删除/修改直接导致崩溃的代码行，不删日志、不重构、不改其他文件
- 只改源码，不改构建配置或项目文件

## 通用 BSOD 模式对照

| BUGCHECK_CODE | 常见根因 |
|---------------|----------|
| 0x7E | 异常未处理，如 assert 失败 → int 3、空指针、除零 |
| 0xD1 | IRQL 违规，驱动在错误 IRQL 访问分页内存 |
| 0x3B | APC_LEVEL 下 PagedPool 分配 |
| 0x19 | Bad Pool Header，tag 不匹配、double free |

## 输出格式

分析完成后输出：
1. **错误码** + **异常码**
2. **完整调用栈**
3. **根因定位**（函数名 + 源码行号）
4. **修复内容**（改了哪个文件的哪行）
