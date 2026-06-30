#!/usr/bin/env python3
"""自动发现 demo/ 下深度为 1 的子目录（有 CMakeLists.txt），生成 build.cmd。"""
from pathlib import Path

DEMO_DIR = Path(__file__).resolve().parent / "demo"
RELEASE_ONLY = {"qt6", "x64dbg"}

BUILD_CMD_DEBUG = """(cmake -B Debug -G "Ninja" -DCMAKE_BUILD_TYPE=Debug . && cmake --build Debug --config Debug) 2>&1 | powershell -NoProfile -Command "$input | Tee-Object -FilePath build.Debug.log"
"""

BUILD_CMD_RELEASE = """(cmake -B Release -G "Ninja" -DCMAKE_BUILD_TYPE=Release . && cmake --build Release --config Release) 2>&1 | powershell -NoProfile -Command "$input | Tee-Object -FilePath build.Release.log"
"""

def main():
    if not DEMO_DIR.is_dir():
        print(f"[ERR]  {DEMO_DIR} not found")
        return

    entries = sorted(d for d in DEMO_DIR.iterdir() if d.is_dir() and (d / "CMakeLists.txt").exists())

    if not entries:
        print("[WARN] no demo directories found")
        return

    for d in entries:
        name = d.name
        content = BUILD_CMD_RELEASE
        if name not in RELEASE_ONLY:
            content += BUILD_CMD_DEBUG

        (d / "build.cmd").write_text(content, encoding="utf-8")
        print(f"[OK]   {name} build.cmd")


if __name__ == "__main__":
    main()
