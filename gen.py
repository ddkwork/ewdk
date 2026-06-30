#!/usr/bin/env python3
"""遍历工作区覆写已有 build.cmd / build.bat，并为 demo/ 下新目录生成 build.cmd。"""
import shutil
from pathlib import Path

WORKSPACE = Path(__file__).resolve().parent
DEMO_DIR = WORKSPACE / "demo"
RELEASE_ONLY = {"qt6", "x64dbg"}

# 遍历时跳过的目录名（大小写敏感）
_EXCLUDE_DIRS = frozenset({
    "CMakeFiles", "Debug", "Release", "MinSizeRel", "RelWithDebInfo",
    "_deps", ".git", ".vscode", "packages", "x64", "x86", "Win32",
})

BUILD_CMD_DEBUG = """(cmake -B Debug -G "Ninja" -DCMAKE_BUILD_TYPE=Debug . && cmake --build Debug --config Debug) 2>&1 | powershell -NoProfile -Command "$input | Tee-Object -FilePath build.Debug.log"
"""

BUILD_CMD_RELEASE = """(cmake -B Release -G "Ninja" -DCMAKE_BUILD_TYPE=Release . && cmake --build Release --config Release) 2>&1 | powershell -NoProfile -Command "$input | Tee-Object -FilePath build.Release.log"
"""


def _build_content(dir_name: str) -> str:
    """根据目录名决定写入的内容。"""
    if dir_name in RELEASE_ONLY:
        return BUILD_CMD_RELEASE
    return BUILD_CMD_DEBUG + BUILD_CMD_RELEASE


def _skip_rel_parts(rel_parts):
    """判断相对路径的父级目录是否应跳过。"""
    return any(
        part in _EXCLUDE_DIRS or part.startswith(".")
        for part in rel_parts[:-1]
    )


def overwrite_existing() -> tuple[int, int]:
    """遍历工作区，覆写已有 build.cmd，重命名 build.bat → build.cmd。"""
    count_cmd = 0
    count_bat = 0

    for p in WORKSPACE.rglob("build.cmd"):
        rel = p.relative_to(WORKSPACE)
        if _skip_rel_parts(rel.parts):
            continue
        p.write_text(_build_content(p.parent.name), encoding="utf-8")
        count_cmd += 1
        print(f"[OVR]  {p}")

    for p in WORKSPACE.rglob("build.bat"):
        rel = p.relative_to(WORKSPACE)
        if _skip_rel_parts(rel.parts):
            continue
        new = p.with_suffix(".cmd")
        shutil.move(str(p), str(new))
        new.write_text(_build_content(new.parent.name), encoding="utf-8")
        count_bat += 1
        print(f"[REN]  {p} → {new.name}")

    return count_cmd, count_bat


def main():
    # 第一步：覆写工作区已有的构建脚本
    c1, c2 = overwrite_existing()
    if c1 or c2:
        print(f"[STA]  overwritten {c1} build.cmd, renamed {c2} build.bat")
    else:
        print("[STA]  no existing build script found")

    # 第二步：为 demo/ 下有 CMakeLists.txt 的目录生成
    if not DEMO_DIR.is_dir():
        print(f"[ERR]  {DEMO_DIR} not found")
        return

    entries = sorted(d for d in DEMO_DIR.iterdir() if d.is_dir() and (d / "CMakeLists.txt").exists())

    if not entries:
        print("[WARN] no demo directories found")
        return

    for d in entries:
        name = d.name
        (d / "build.cmd").write_text(_build_content(name), encoding="utf-8")
        print(f"[GEN]  {name} build.cmd")


if __name__ == "__main__":
    main()
