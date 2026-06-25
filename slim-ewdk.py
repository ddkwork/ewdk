#!/usr/bin/env python3
"""
EWDK ISO Slimmer - extracts only what's needed for CI/offline builds.

Usage:
    python slim-ewdk.py <source> <dest>

    source: EWDK ISO mount path (e.g. E:\\) or unpacked ISO directory
    dest:   output directory for the slimmed EWDK

Keeps: x64 + x86 build tools, ATLMFC, WDK (km/um/shared/ucrt), MSBuild, DIA SDK
Drops: ARM/ARM64/CHPE, Spectre libs, UWP/store libs, debuggers, IDE bloat
"""

import os
import sys
import shutil
import re


def detect_version(dirnames, pattern, fallback):
    """从目录列表中用正则匹配版本号，取最新的。"""
    versions = []
    for d in dirnames:
        m = re.match(pattern, d)
        if m:
            versions.append(d)
    if not versions:
        print(f"  WARN: no version matching '{pattern}' found, using fallback {fallback}")
        return fallback
    versions.sort()
    latest = versions[-1]
    print(f"  Detected: {latest}  (from {len(versions)} candidates)")
    return latest


def copy_filtered(src_dir, dst_dir, patterns=None):
    """Copy files from src_dir to dst_dir, maintaining structure."""
    if not os.path.isdir(src_dir):
        print(f"  SKIP (source not found): {src_dir}")
        return 0

    count = 0
    for root, dirs, files in os.walk(src_dir):
        # Compute relative path
        rel = os.path.relpath(root, src_dir)
        if rel == ".":
            dst_root = dst_dir
        else:
            dst_root = os.path.join(dst_dir, rel)

        os.makedirs(dst_root, exist_ok=True)

        for f in files:
            src_file = os.path.join(root, f)
            dst_file = os.path.join(dst_root, f)
            try:
                shutil.copy2(src_file, dst_file)
                count += 1
            except Exception as e:
                print(f"  ERROR copying {src_file}: {e}")
    return count


def copy_directory(src, dst, description=""):
    """Copy entire directory tree."""
    if description:
        print(f"  [{description}]", end=" ")
    if not os.path.isdir(src):
        print(f"SKIP (not found: {src})")
        return 0
    count = copy_filtered(src, dst)
    print(f"OK ({count} files)")
    return count


def ensure_parent(path):
    os.makedirs(os.path.dirname(path), exist_ok=True)


def copy_single(src, dst, description=""):
    """Copy a single file."""
    if description:
        print(f"  [{description}]", end=" ")
    if not os.path.isfile(src):
        print(f"SKIP (not found: {src})")
        return False
    ensure_parent(dst)
    shutil.copy2(src, dst)
    print(f"OK")
    return True


def main():
    if len(sys.argv) != 3:
        print(__doc__)
        sys.exit(1)

    src = os.path.abspath(sys.argv[1])
    dst = os.path.abspath(sys.argv[2])

    if not os.path.isdir(src):
        print(f"ERROR: source '{src}' does not exist")
        sys.exit(1)

    os.makedirs(dst, exist_ok=True)

    total_files = 0
    print(f"\n{'='*60}")
    print(f" Slimming EWDK ISO")
    print(f"  Source: {src}")
    print(f"  Dest:   {dst}")
    print(f"{'='*60}\n")

        # Auto-detect VS version (2022/2029/2035...) — don't hardcode
    VS_VER = detect_version(
        os.listdir(os.path.join(src, "Program Files", "Microsoft Visual Studio")),
        r"^\d{4}$",  # matches "2022", "2029", "2035" etc.
        "2022"
    )
    SRC_VSROOT = os.path.join(src, "Program Files", "Microsoft Visual Studio", VS_VER, "BuildTools")
    DST_VSROOT = os.path.join(dst, "Program Files", "Microsoft Visual Studio", VS_VER, "BuildTools")
    print(f"  VS version:           {VS_VER}")

    # Auto-detect MSVC toolchain version (14.44.35207, 14.48.xxxxx...)
    MSVC_VER = detect_version(
        os.listdir(os.path.join(SRC_VSROOT, "VC", "Tools", "MSVC")),
        r"^\d+\.\d+\.\d+$",
        "14.44.35207"
    )

    # Auto-detect WDK version (10.0.28000.0, 10.0.31000.0...)
    WDK_VER = detect_version(
        os.listdir(os.path.join(src, "Program Files", "Windows Kits", "10", "Include")),
        r"^\d+\.\d+\.\d+\.\d+$",
        "10.0.28000.0"
    )

    SRC_MSVC = os.path.join(SRC_VSROOT, "VC", "Tools", "MSVC", MSVC_VER)
    DST_MSVC = os.path.join(dst, "Program Files", "Microsoft Visual Studio", VS_VER, "BuildTools", "VC", "Tools", "MSVC", MSVC_VER)

    SRC_KITS = os.path.join(src, "Program Files", "Windows Kits", "10")
    DST_KITS = os.path.join(dst, "Program Files", "Windows Kits", "10")

    # =========================================================================
    # 1. BuildEnv scripts
    # =========================================================================
    print("\n--- BuildEnv ---")
    total_files += copy_directory(
        os.path.join(src, "BuildEnv"),
        os.path.join(dst, "BuildEnv"),
        "BuildEnv scripts")

    # =========================================================================
    # 2. MSVC Toolchain
    # =========================================================================
    print("\n--- MSVC Toolchain ---")

    # 2a. compilers/linkers (Hostx64 -> x64 + x86 targets)
    for host, target in [("x64", "x64"), ("x64", "x86")]:
        total_files += copy_directory(
            os.path.join(SRC_MSVC, "bin", f"Host{host}", target),
            os.path.join(DST_MSVC, "bin", f"Host{host}", target),
            f"bin/Host{host}/{target}")

    # also copy Hostx64/1033 (common UI resources used by tools)
    total_files += copy_directory(
        os.path.join(SRC_MSVC, "bin", "Hostx64", "x64", "1033"),
        os.path.join(DST_MSVC, "bin", "Hostx64", "x64", "1033"),
        "bin/Hostx64/x64/1033 (tool resources)")
    total_files += copy_directory(
        os.path.join(SRC_MSVC, "bin", "Hostx64", "x86", "1033"),
        os.path.join(DST_MSVC, "bin", "Hostx64", "x86", "1033"),
        "bin/Hostx64/x86/1033 (tool resources)")

    # rc.exe from WDK (used by both x64 and x86 builds)
    # will be copied in WDK bin section

    # 2b. MSVC headers
    total_files += copy_directory(
        os.path.join(SRC_MSVC, "include"),
        os.path.join(DST_MSVC, "include"),
        "include (MSVC headers)")

    # 2c. MSVC libs (x64 + x86 only)
    for arch in ("x64", "x86"):
        total_files += copy_directory(
            os.path.join(SRC_MSVC, "lib", arch),
            os.path.join(DST_MSVC, "lib", arch),
            f"lib/{arch}")

    # 2d. CRT source (i386 needed for __allmul etc.)
    total_files += copy_directory(
        os.path.join(SRC_MSVC, "crt", "src", "i386"),
        os.path.join(DST_MSVC, "crt", "src", "i386"),
        "crt/src/i386 (asm helpers)")

    # 2e. version.txt
    copy_single(
        os.path.join(SRC_VSROOT, "version.txt"),
        os.path.join(DST_VSROOT, "version.txt"),
        "version.txt")

    # 2f. Common7/Tools (needed for vsdevcmd.bat)
    total_files += copy_directory(
        os.path.join(SRC_VSROOT, "Common7", "Tools"),
        os.path.join(DST_VSROOT, "Common7", "Tools"),
        "Common7/Tools (vsdevcmd)")

    # =========================================================================
    # 3. ATLMFC
    # =========================================================================
    print("\n--- ATLMFC ---")
    total_files += copy_directory(
        os.path.join(SRC_MSVC, "atlmfc", "include"),
        os.path.join(DST_MSVC, "atlmfc", "include"),
        "atlmfc/include")

    for arch in ("x64", "x86"):
        total_files += copy_directory(
            os.path.join(SRC_MSVC, "atlmfc", "lib", arch),
            os.path.join(DST_MSVC, "atlmfc", "lib", arch),
            f"atlmfc/lib/{arch}")

    # =========================================================================
    # 4. DIA SDK
    # =========================================================================
    print("\n--- DIA SDK ---")
    total_files += copy_directory(
        os.path.join(SRC_VSROOT, "DIA SDK", "include"),
        os.path.join(DST_VSROOT, "DIA SDK", "include"),
        "DIA SDK/include")
    total_files += copy_directory(
        os.path.join(SRC_VSROOT, "DIA SDK", "lib", "amd64"),
        os.path.join(DST_VSROOT, "DIA SDK", "lib", "amd64"),
        "DIA SDK/lib/amd64")
    total_files += copy_directory(
        os.path.join(SRC_VSROOT, "DIA SDK", "bin", "amd64"),
        os.path.join(DST_VSROOT, "DIA SDK", "bin", "amd64"),
        "DIA SDK/bin/amd64")

    # =========================================================================
    # 5. MSBuild (needed for .vcxproj targets, C++ build props)
    # =========================================================================
    print("\n--- MSBuild ---")
    msbuild_src = os.path.join(SRC_VSROOT, "MSBuild", "Current")
    msbuild_dst = os.path.join(DST_VSROOT, "MSBuild", "Current")
    total_files += copy_directory(
        os.path.join(msbuild_src, "Bin"),
        os.path.join(msbuild_src, "Bin", "amd64"),
        "MSBuild/Current/Bin (core + amd64)")

    # Keep only the English locale
    # Actually just keep all of Bin - it's not that large
    total_files += copy_directory(
        os.path.join(msbuild_src, "Bin"),
        os.path.join(msbuild_dst, "Bin"),
        "MSBuild/Current/Bin")

    # Keep Roslyn (needed for C# projects in Qt/MSBuild scenarios)
    total_files += copy_directory(
        os.path.join(SRC_MSVC, "bin", "Hostx64", "x64", "1033"),
        os.path.join(DST_MSVC, "bin", "Hostx64", "x64", "1033"),
        "MSVC/bin ui resources")

    # =========================================================================
    # 6. Windows Kits (WDK + SDK)
    # =========================================================================
    print("\n--- Windows Kits ---")

    # 6a. Include directories (keep ALL - headers are not architecture specific)
    # km/crt is required for C++ kernel drivers (new, eh.h, etc.)
    for subdir in ("km", "um", "shared", "ucrt"):
        total_files += copy_directory(
            os.path.join(SRC_KITS, "Include", WDK_VER, subdir),
            os.path.join(DST_KITS, "Include", WDK_VER, subdir),
            f"Include/{WDK_VER}/{subdir}")

    # WDF include (needed for KMDF drivers)
    total_files += copy_directory(
        os.path.join(SRC_KITS, "Include", "wdf"),
        os.path.join(DST_KITS, "Include", "wdf"),
        "Include/wdf")

    # 6b. Lib directories (x64 + x86 only)
    for subdir in ("km",):
        for arch in ("x64",):
            total_files += copy_directory(
                os.path.join(SRC_KITS, "Lib", WDK_VER, subdir, arch),
                os.path.join(DST_KITS, "Lib", WDK_VER, subdir, arch),
                f"Lib/{WDK_VER}/{subdir}/{arch}")

    for subdir in ("ucrt", "um"):
        for arch in ("x64", "x86"):
            total_files += copy_directory(
                os.path.join(SRC_KITS, "Lib", WDK_VER, subdir, arch),
                os.path.join(DST_KITS, "Lib", WDK_VER, subdir, arch),
                f"Lib/{WDK_VER}/{subdir}/{arch}")

    # 6c. WDF libs
    total_files += copy_directory(
        os.path.join(SRC_KITS, "Lib", "wdf"),
        os.path.join(DST_KITS, "Lib", "wdf"),
        "Lib/wdf")

    # 6d. WDK binaries (tools)
    for arch in ("x64",):
        total_files += copy_directory(
            os.path.join(SRC_KITS, "bin", WDK_VER, arch),
            os.path.join(DST_KITS, "bin", WDK_VER, arch),
            f"bin/{WDK_VER}/{arch}")

    # Also copy x86 tools (some may be needed for x86 builds)
    # Keep minimal: rc.exe, mt.exe, midl.exe, mc.exe
    src_wdk_bin_x86 = os.path.join(SRC_KITS, "bin", WDK_VER, "x86")
    dst_wdk_bin_x86 = os.path.join(DST_KITS, "bin", WDK_VER, "x86")
    TOOLS_ESSENTIAL = (
        "rc.exe", "rcdll.dll", "mt.exe", "midl.exe", "midlc.exe",
        "mc.exe", "signtool.exe", "makecat.exe", "certmgr.exe",
        "inf*.exe",
    )
    if os.path.isdir(src_wdk_bin_x86):
        os.makedirs(dst_wdk_bin_x86, exist_ok=True)
        for f in os.listdir(src_wdk_bin_x86):
            src_file = os.path.join(src_wdk_bin_x86, f)
            if os.path.isfile(src_file):
                name_lower = f.lower()
                if any(name_lower.startswith(pattern.replace("*", "").lower()) for pattern in TOOLS_ESSENTIAL):
                    shutil.copy2(src_file, os.path.join(dst_wdk_bin_x86, f))
                    total_files += 1
                    print(f"  [bin/{WDK_VER}/x86] {f}")

    # 6e. WDK other infrastructure
    total_files += copy_directory(
        os.path.join(SRC_KITS, "bin", WDK_VER, "WppConfig"),
        os.path.join(DST_KITS, "bin", WDK_VER, "WppConfig"),
        f"bin/{WDK_VER}/WppConfig")

    # =========================================================================
    # 7. Root files
    # =========================================================================
    print("\n--- Root files ---")
    for f in ("LaunchBuildEnv.cmd", "LICENSE.rtf"):
        copy_single(
            os.path.join(src, f),
            os.path.join(dst, f),
            f)

    # =========================================================================
    # Summary
    # =========================================================================
    print(f"\n{'='*60}")
    print(f" Done! {total_files} files copied")
    print(f" Destination: {dst}")
    print(f"{'='*60}")

    # Estimate size
    total_size = 0
    for root, dirs, files in os.walk(dst):
        for f in files:
            try:
                total_size += os.path.getsize(os.path.join(root, f))
            except:
                pass
    print(f" Total size: {total_size / (1024**3):.1f} GB ({total_size / (1024**2):.0f} MB)")
    print()


if __name__ == "__main__":
    main()