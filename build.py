#!/usr/bin/env python3
"""批量执行各 demo 的 build.cmd。"""
import subprocess
import sys
from pathlib import Path
from datetime import datetime

PROJECT = Path(__file__).resolve().parent
DEMO_DIR = PROJECT / "demo"


def run(cmd_path: Path) -> bool:
    print(f"\n>>> [{cmd_path.parent.name}] {cmd_path}")
    with subprocess.Popen(
        [str(cmd_path)],
        cwd=cmd_path.parent,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        encoding="utf-8",
    ) as p:
        for line in p.stdout:
            print(line, end="", flush=True)
    return p.returncode == 0


def main():
    print("=" * 50)
    print("EWDK Build All")
    print(datetime.now().strftime("%Y-%m-%d %H:%M:%S"))
    print("=" * 50)

    scripts = sorted(DEMO_DIR.glob("*/build.cmd"))
    if not scripts:
        print("No build.cmd found under demo/")
        sys.exit(1)

    failed = []
    for s in scripts:
        if not run(s):
            failed.append(s.parent.name)
        print()

    if failed:
        print(f"FAILED: {', '.join(failed)}")
        sys.exit(1)
    else:
        print("All demos built successfully!")
        sys.exit(0)


if __name__ == "__main__":
    main()
