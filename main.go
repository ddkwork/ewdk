package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/mylog/pretty"
	"github.com/ddkwork/golibrary/std/stream"
	"golang.org/x/sys/windows/registry"
)

var cmakeEnvVars = map[string]string{
	"CMAKE_INCLUDE_PATH":   "include paths for find_file()",
	"CMAKE_LIBRARY_PATH":   "lib paths for find_library()",
	"CMAKE_PREFIX_PATH":    "prefix paths for find_*",
	"CMAKE_PROGRAM_PATH":   "program paths",
	"CMAKE_FRAMEWORK_PATH": "framework paths (macOS)",
	"CMAKE_APPBUNDLE_PATH": "app bundle paths (macOS)",
	"CC":                   "C compiler",
	"CFLAGS":               "C compiler flags",
	"CXX":                  "C++ compiler",
	"CXXFLAGS":             "C++ compiler flags",
	"LDFLAGS":              "linker flags",
	"ADSP_ROOT":            "Analog Devices SHARC root",
	"CSFLAGS":              "C# compiler flags",
	"CUDACXX":              "CUDA C++ compiler",
	"CUDAFLAGS":            "CUDA flags",
	"CUDAHOSTCXX":          "CUDA host C++ compiler",
	"FC":                   "Fortran compiler",
	"FFLAGS":               "Fortran flags",
	"HIPCXX":               "HIP C++ compiler",
	"HIPFLAGS":             "HIP flags",
	"ISPC":                 "Intel SPMD Program Compiler",
	"RC":                   "Windows resource compiler",
	"RCFLAGS":              "Windows resource compiler flags",
	"SWIFTC":               "Swift compiler",
	"OBJC":                 "Objective-C compiler",
	"OBJCFLAGS":            "Objective-C compiler flags",
	"OBJCXX":               "Objective-C++ compiler",
	"OBJCXXFLAGS":          "Objective-C++ compiler flags",
	"CMAKE_BUILD_TYPE":     "build type (Release/Debug etc.)",
	"CMAKE_GENERATOR":      "generator name",
	"CMAKE_INSTALL_PREFIX": "install prefix",
	"CMAKE_TOOLCHAIN_FILE": "toolchain file path",
	"DSTDIR":               "install destination root",
	"VERBOSE":              "verbose output flag",
}

func main() {
	mgr := NewRegistryEnvManager()

	mgr.CleanInvalidVars(func(key string) bool {

		return false
	})
	//mylog.Struct(mylog.Check2(mgr.List()))

	isoPath := resolveISOPath()
	mylog.Success(isoPath)
	driveLetter := mylog.Check2(mgr.MountISO(isoPath))
	mylog.Success("EWDK mounted at: ", driveLetter)

	setupEnvCmd := driveLetter + ":\\BuildEnv\\SetupBuildEnv.cmd"
	mylog.Success(setupEnvCmd)
	all := mylog.Check2(runSetupBuildEnv(setupEnvCmd))
	mylog.Struct(all)
	setEwdkEnvToSystem(mgr, all)
	syncEwdkEnvToUser(mgr, all)
	saveEnvToFile(mgr)



	mylog.Success("Build complete")
}

func resolveISOPath() string {
	githubWorkspace := os.Getenv("GITHUB_WORKSPACE")
	if githubWorkspace != "" {
		iso := filepath.Join(os.Getenv("TEMP"), "ewdk.iso")
		if !stream.FileExists(iso) {
			panic("请在ci环境中执行下载脚本获取iso下载地址并执行下载操作")
		}
		return iso
	}
	if entries, err := os.ReadDir("."); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "EWDK") && filepath.Ext(entry.Name()) == ".iso" {
				isoPath := mylog.Check2(filepath.Abs(entry.Name()))
				if !stream.FileExists(isoPath) {
					panic("当前路径找到的iso文件不存在，请始终保持一个iso文件存在")
				}
				return isoPath
			}
		}
	}
	panic("Could not find ISO files")
}

func appendPATH(mgr EnvManager, elem string) {
	key, err := openEnvKey(registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer key.Close()

	currentPath, _, err := key.GetStringValue("PATH")
	if err != nil {
		return
	}

	if strings.Contains(currentPath, elem) {
		fmt.Printf("  [SKIP] already in PATH: %s\n", elem)
		return
	}

	newPath := currentPath + ";" + elem
	if err := mgr.Set("PATH", newPath); err != nil {
		fmt.Printf("  [FAIL] Append to PATH: %v\n", err)
	} else {
		fmt.Printf("  [OK]   PATH += %s\n", elem)
	}
}

func setEwdkEnvToSystem(mgr EnvManager, env ewdkEnv) {
	mgr.Set("CMAKE_C_COMPILER", env.CC)
	mgr.Set("CMAKE_CXX_COMPILER", env.CC)
	if ninjaDir, err := filepath.Abs(filepath.Dir("ninja.exe")); err == nil {
		appendPATH(mgr, ninjaDir)
	}
	rc := filepath.Join(env.WDKContentRoot, "bin", env.WindowsTargetPlatformVersion, "x64", "rc.exe")
	if stream.FileExists(rc) {
		mgr.Set("RC", rc)
		appendPATH(mgr, filepath.Dir(rc))
	}
	mgr.Set("CMAKE_INCLUDE_PATH", strings.Join(env.INCLUDE, ";"))
	mgr.Set("CMAKE_LIBRARY_PATH", strings.Join(env.LIB, ";"))
	appendPATH(mgr, filepath.Dir(env.CC))

	pairs := map[string]string{
		EnvWindowsTargetPlatformVersion: env.WindowsTargetPlatformVersion,
		EnvWDKContentRoot:               env.WDKContentRoot,
		EnvBuildLabSetupRoot:            env.BuildLabSetupRoot,
		EnvVSINSTALLDIR:                 env.VSINSTALLDIR,
		EnvWDKBinRoot:                   env.WDKBinRoot,
		EnvDiaRoot:                      env.DiaRoot,
		EnvVCToolsInstallDir:            env.VCToolsInstallDir,
		EnvCC:                           env.CC,
		EnvCXX:                          env.CC,
		EnvINCLUDE:                      strings.Join(env.INCLUDE, ";"),
		EnvLIB:                          strings.Join(env.LIB, ";"),
	}
	for name, value := range pairs {
		if value == "" {
			continue
		}
		if err := mgr.Set(name, value); err != nil {
			fmt.Printf("  [FAIL] Set %s: %v\n", name, err)
		} else {
			fmt.Printf("  [OK]   %s=%s\n", name, value)
		}
	}
}

func syncEwdkEnvToUser(mgr EnvManager, env ewdkEnv) {
	userKey, err := openUserEnvKey(registry.SET_VALUE)
	if err != nil {
		fmt.Printf("  [FAIL] open user env: %v\n", err)
		return
	}
	defer userKey.Close()

	systemVars, err := mgr.List()
	if err != nil {
		fmt.Printf("  [FAIL] list system env: %v\n", err)
		return
	}

	for _, ev := range systemVars {
		if ev.Type != "SZ" && ev.Type != "EXPAND_SZ" {
			continue
		}
		_, _, userErr := userKey.GetStringValue(ev.Name)
		if userErr == nil {
			continue
		}
		if err := userKey.SetStringValue(ev.Name, ev.Value); err != nil {
			fmt.Printf("  [FAIL] UserEnv Set %s: %v\n", ev.Name, err)
		} else {
			fmt.Printf("  [OK]   UserEnv %s=%s\n", ev.Name, ev.Value)
		}
	}
}

func saveEnvToFile(mgr EnvManager) {
	systemVars, err := mgr.List()
	if err != nil {
		fmt.Printf("  [FAIL] list system env: %v\n", err)
		return
	}
	userVars, err := mgr.ListUser()
	if err != nil {
		fmt.Printf("  [FAIL] list user env: %v\n", err)
		return
	}

	filename := filepath.Join(".", "ewdk-env-"+strings.ReplaceAll(time.Now().Format("20060102-150405"), ":", "")+".txt")
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("  [FAIL] create file: %v\n", err)
		return
	}
	defer f.Close()

	fmt.Fprintln(f, "=== SYSTEM ENV ===")
	pretty.PrintTo(f, systemVars, true)
	fmt.Fprintln(f, "\n=== USER ENV ===")
	pretty.PrintTo(f, userVars, true)

	fmt.Printf("  [OK]   env saved to: %s\n", filename)
}
