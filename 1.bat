@echo off
setlocal enabledelayedexpansion

:: 先保存原始环境
set > before.txt

:: 运行 EWDK 脚本
call "f:\BuildEnv\SetupBuildEnv.cmd"

:: 保存新环境
set > after.txt

:: 使用 PowerShell 比较差异
powershell -Command "$before = @{}; Get-Content 'before.txt' | ForEach-Object { if ($_ -match '^(.*?)=(.*)$') { $before[$matches[1]] = $matches[2] } }; Get-Content 'after.txt' | ForEach-Object { if ($_ -match '^(.*?)=(.*)$') { if (-not $before.ContainsKey($matches[1])) { $_ } } } | Out-File 'ewdk_env_diff.txt'"

del before.txt after.txt
echo 差异已保存到 ewdk_env_diff.txt
