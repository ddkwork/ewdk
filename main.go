package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
)

const (
	ewdkCmakeGenerated              = "ewdk-env.cmake"
	EnvWindowsTargetPlatformVersion = "WindowsTargetPlatformVersion"
	EnvVCToolsInstallDir            = "VCToolsInstallDir"
	EnvWDKContentRoot               = "WDKContentRoot"
	EnvBuildLabSetupRoot            = "BuildLabSetupRoot"
	EnvVSINSTALLDIR                 = "VSINSTALLDIR"
	EnvINCLUDE                      = "INCLUDE"
	EnvLIB                          = "LIB"
	EnvWDKBinRoot                   = "WDKBinRoot"
)

const powershell = "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"

var taskName = "EWDK_Mount"

type ewdkCommonEnv struct {
	WDKContentRoot               string
	WindowsTargetPlatformVersion string
	VCToolsInstallDir            string
	WDKBinRoot                   string
	DiaRoot                      string
	VSINSTALLDIR                 string
	BuildLabSetupRoot            string
	CC                           string
	RC                           string
	MT                           string
	NinjaDir                     string
}

type ewdkKMEnv struct {
	IncludeDirs []string
	LibDirs     []string
}

type ewdkUMEnv struct {
	IncludeDirs []string
	LibDirs     []string
}

type ewdkEnv struct {
	Common ewdkCommonEnv
	KM     ewdkKMEnv
	UM     ewdkUMEnv
	INCLUDE []string
	LIB     []string
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "unmount":
			isoPath := resolveISOPath()
			mylog.Check(unmountISO(isoPath))
			deleteScheduledTask()
			mylog.Success("EWDK unmounted")
			return
		case "clean":
			cleanGenerated()
			return
		}
	}

	isoPath := resolveISOPath()
	mylog.Success("ISO: ", isoPath)

	var driveLetter string
	if isMounted() {
		driveLetter = getEwdkDriveLetter()
		mylog.Success("EWDK already mounted at: ", driveLetter)
	} else {
		driveLetter = mylog.Check2(mountISO(isoPath))
		mylog.Success("EWDK mounted at: ", driveLetter)
	}

	if err := createScheduledTask(isoPath); err != nil {
		fmt.Println("  [WARN] scheduled task:", err)
	}

	setupEnvCmd := driveLetter + ":\\BuildEnv\\SetupBuildEnv.cmd"
	mylog.Success("SetupCmd: ", setupEnvCmd)
	env := mylog.Check2(runSetupBuildEnv(setupEnvCmd))
	mylog.Struct(env)

	cmakePath := filepath.Join(".", ewdkCmakeGenerated)
	mylog.Check(generateEwdkCmake(env, cmakePath))
	mylog.Success("Generated: ", cmakePath)

	mylog.Success("Environment ready. Run build.bat to start building.")
}

const isoPath = `d:\ewdk\EWDK_br_release_28000_251103-1709.iso`

func resolveISOPath() string {
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		iso := filepath.Join(os.Getenv("TEMP"), "ewdk.iso")
		if !stream.FileExists(iso) {
			panic("CI env: ewdk.iso not found in TEMP")
		}
		return iso
	}
	if !stream.FileExists(isoPath) {
		panic("ISO file not found: " + isoPath)
	}
	return isoPath
}

func cleanGenerated() {
	files := []string{ewdkCmakeGenerated}
	for _, f := range files {
		if stream.FileExists(f) {
			mylog.Check(os.Remove(f))
			fmt.Printf("  [OK]   Removed: %s\n", f)
		}
	}
}

func mountISO(isoPath string) (string, error) {
	script := fmt.Sprintf(`(Mount-DiskImage -ImagePath '%s' -PassThru | Get-Volume).DriveLetter`, strings.ReplaceAll(isoPath, "'", "''"))
	output, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).Output()
	if err != nil {
		return "", fmt.Errorf("mount iso: %w", err)
	}
	letter := strings.TrimSpace(string(output))
	if letter == "" || len(letter) != 1 {
		return "", fmt.Errorf("invalid drive letter: %q", letter)
	}
	return letter, nil
}

func unmountISO(isoPath string) error {
	if isoPath == "" {
		script := `Get-Volume | Where-Object { $_.DriveType -eq 'CD-ROM' } | ForEach-Object { Dismount-DiskImage -DevicePath "\\.\$($_.DriveLetter):" -ErrorAction SilentlyContinue }`
		output, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).CombinedOutput()
		if err != nil {
			return fmt.Errorf("unmount all: %w, output: %s", err, strings.TrimSpace(string(output)))
		}
		return nil
	}
	script := fmt.Sprintf(`Dismount-DiskImage -ImagePath '%s'`, strings.ReplaceAll(isoPath, "'", "''"))
	if output, err := exec.Command(powershell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script).CombinedOutput(); err != nil {
		return fmt.Errorf("unmount iso: %w, output: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func isMounted() bool {
	return getEwdkDriveLetter() != ""
}

func getEwdkDriveLetter() string {
	for c := 'F'; c >= 'A'; c-- {
		setupCmd := string(c) + ":\\BuildEnv\\SetupBuildEnv.cmd"
		if _, err := os.Stat(setupCmd); err == nil {
			return string(c)
		}
	}
	return ""
}

func createScheduledTask(isoPath string) error {
	deleteScheduledTask()
	script := fmt.Sprintf(`Mount-DiskImage -ImagePath '%s' -PassThru | Out-Null`, strings.ReplaceAll(isoPath, "'", "''"))
	psCmd := fmt.Sprintf(`powershell.exe -ExecutionPolicy Bypass -Command "%s"`, script)
	exec.Command("schtasks", "/Delete", "/TN", taskName, "/F").Run()
	err := exec.Command("schtasks", "/Create",
		"/TN", taskName,
		"/TR", psCmd,
		"/SC", "ONLOGON",
		"/DELAY", "0000:00:10",
		"/F",
	).Run()
	if err != nil {
		return fmt.Errorf("create scheduled task: %w", err)
	}
	fmt.Printf("Task registered: %s\n", taskName)
	return nil
}

func deleteScheduledTask() {
	exec.Command("schtasks", "/Delete", "/TN", taskName, "/F").Run()
}

func runSetupBuildEnv(setupCmd string) (ewdkEnv, error) {
	tmpFile := filepath.Join(os.TempDir(), "ewdk-env-after-setup.txt")
	os.Remove(tmpFile)
	cmd := exec.Command("C:\\Windows\\System32\\cmd.exe", "/c", setupCmd+" amd64 && set > "+tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	mylog.Check(cmd.Run())

	data := mylog.Check2(os.ReadFile(tmpFile))
	mylog.Json(string(data))
	defer os.Remove(tmpFile)

	result := make(map[string]string)
	for s := range strings.Lines(string(data)) {
		key, value, found := strings.Cut(s, "=")
		if found {
			switch key {
			case
				EnvWindowsTargetPlatformVersion,
				EnvVCToolsInstallDir,
				EnvWDKContentRoot,
				EnvBuildLabSetupRoot,
				EnvVSINSTALLDIR,
				EnvINCLUDE,
				EnvLIB,
				EnvWDKBinRoot:
			default:
				continue
			}
			result[key] = strings.TrimSuffix(value, "\r\n")
		}
	}
	DiaRoot := filepath.Join(result[EnvVSINSTALLDIR], "DIA SDK")

	pre := filepath.Join(result[EnvWDKContentRoot], "Include", result[EnvWindowsTargetPlatformVersion])
	ucrtInc := filepath.Join(result[EnvWDKContentRoot], "Include", result[EnvWindowsTargetPlatformVersion], "ucrt")
	ucrtLib := filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "ucrt", "x64")
	umLib := filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "um", "x64")

	vcInc := []string{
		filepath.Join(result[EnvVCToolsInstallDir], "include"),
		filepath.Join(result[EnvVCToolsInstallDir], "ATLMFC", "include"),
		filepath.Join(result[EnvVSINSTALLDIR], "VC", "Auxiliary", "VS", "include"),
		filepath.Join(DiaRoot, "include"),
	}
	vcLib := []string{
		filepath.Join(result[EnvVCToolsInstallDir], "lib", "x64"),
		filepath.Join(result[EnvVCToolsInstallDir], "ATLMFC", "lib", "x64"),
		filepath.Join(DiaRoot, "lib"),
	}

	cc := filepath.Join(result[EnvVCToolsInstallDir], "bin\\Hostx64\\x64\\cl.exe")
	rc := filepath.Join(result[EnvWDKContentRoot], "bin", result[EnvWindowsTargetPlatformVersion], "x64", "rc.exe")
	mt := filepath.Join(result[EnvWDKContentRoot], "bin", result[EnvWindowsTargetPlatformVersion], "x64", "mt.exe")
	ninjaDir, _ := filepath.Abs(filepath.Dir("ninja.exe"))

	common := ewdkCommonEnv{
		WDKContentRoot:               result[EnvWDKContentRoot],
		WindowsTargetPlatformVersion: result[EnvWindowsTargetPlatformVersion],
		VCToolsInstallDir:            result[EnvVCToolsInstallDir],
		WDKBinRoot:                   result[EnvWDKBinRoot],
		DiaRoot:                      DiaRoot,
		VSINSTALLDIR:                 result[EnvVSINSTALLDIR],
		BuildLabSetupRoot:            result[EnvBuildLabSetupRoot],
		CC:                           cc,
		RC:                           rc,
		NinjaDir:                     ninjaDir,
	}
	if _, err := os.Stat(mt); err == nil {
		common.MT = mt
	}

	km := ewdkKMEnv{
		IncludeDirs: []string{
			filepath.Join(pre, "shared"),
			filepath.Join(pre, "km"),
			filepath.Join(pre, "km", "crt"),
		},
		LibDirs: []string{
			filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "km", "x64"),
		},
	}

	um := ewdkUMEnv{
		IncludeDirs: append([]string{
			filepath.Join(pre, "shared"),
			filepath.Join(pre, "um"),
			ucrtInc,
		}, vcInc...),
		LibDirs: append([]string{umLib, ucrtLib}, vcLib...),
	}

	include := strings.Split(result[EnvINCLUDE], ";")
	lib := strings.Split(result[EnvLIB], ";")

	include = append(include,
		filepath.Join(DiaRoot, "include"),
		filepath.Join(pre, "km"),
		filepath.Join(pre, "km", "crt"),
		filepath.Join(pre, "um"),
		filepath.Join(pre, "shared"),
		filepath.Join(result[EnvWDKContentRoot], "Include\\wdf\\kmdf\\1.35"),
	)

	lib = append(lib,
		filepath.Join(DiaRoot, "lib"),
		filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "km", "x64"),
		filepath.Join(result[EnvWDKContentRoot], "Lib", result[EnvWindowsTargetPlatformVersion], "um", "x64"),
	)

	return ewdkEnv{
		Common:  common,
		KM:      km,
		UM:      um,
		INCLUDE: include,
		LIB:     lib,
	}, nil
}

func generateEwdkCmake(env ewdkEnv, outputPath string) error {
	cm := func(p string) string {
		p = strings.ReplaceAll(p, `\`, "/")
		p = strings.TrimRight(p, "/")
		return p
	}

	writeDirs := func(b *strings.Builder, name string, dirs []string) {
		b.WriteString(fmt.Sprintf("\nset(%s \"", name))
		for i, p := range dirs {
			if i > 0 {
				b.WriteString(";")
			}
			b.WriteString(cm(p))
		}
		b.WriteString("\")\n")
	}

	c := env.Common
	var cmake strings.Builder
	cmake.WriteString("# Auto-generated by ewdk toolchain manager - DO NOT EDIT\n")
	cmake.WriteString("# All EWDK environment variables are embedded here for portability\n")

	cmake.WriteString("\n# ---- Common ----\n")
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_WDKContentRoot \"%s\")\n", cm(c.WDKContentRoot)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_WindowsTargetPlatformVersion \"%s\")\n", c.WindowsTargetPlatformVersion))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_VCToolsInstallDir \"%s\")\n", cm(c.VCToolsInstallDir)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_WDKBinRoot \"%s\")\n", cm(c.WDKBinRoot)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_DiaRoot \"%s\")\n", cm(c.DiaRoot)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_VSINSTALLDIR \"%s\")\n", cm(c.VSINSTALLDIR)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_BuildLabSetupRoot \"%s\")\n", cm(c.BuildLabSetupRoot)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_CC \"%s\")\n", cm(c.CC)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_CXX \"%s\")\n", cm(c.CC)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_RC \"%s\")\n", cm(c.RC)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_NINJA_DIR \"%s\")\n", cm(c.NinjaDir)))
	if c.MT != "" {
		cmake.WriteString(fmt.Sprintf("set(EWDK_COMMON_MT \"%s\")\n", cm(c.MT)))
	}

	cmake.WriteString("\n# ---- KM (Kernel-Mode) ----\n")
	writeDirs(&cmake, "EWDK_KM_INCLUDE_DIRS", env.KM.IncludeDirs)
	writeDirs(&cmake, "EWDK_KM_LIB_DIRS", env.KM.LibDirs)

	cmake.WriteString("\n# ---- UM (User-Mode) ----\n")
	writeDirs(&cmake, "EWDK_UM_INCLUDE_DIRS", env.UM.IncludeDirs)
	writeDirs(&cmake, "EWDK_UM_LIB_DIRS", env.UM.LibDirs)

	cmake.WriteString("\n# ---- Legacy env vars ----\n")
	writeDirs(&cmake, "EWDK_INCLUDE", env.INCLUDE)
	writeDirs(&cmake, "EWDK_LIB", env.LIB)

	cmake.WriteString("\n# ---- Compiler / Linker ----\n")
	cmake.WriteString("set(CMAKE_C_COMPILER \"${EWDK_COMMON_CC}\" CACHE FILEPATH \"\" FORCE)\n")
	cmake.WriteString("set(CMAKE_CXX_COMPILER \"${EWDK_COMMON_CXX}\" CACHE FILEPATH \"\" FORCE)\n")
	cmake.WriteString("set(CMAKE_RC_COMPILER \"${EWDK_COMMON_RC}\" CACHE FILEPATH \"\" FORCE)\n")
	cmake.WriteString("set(CMAKE_INCLUDE_PATH \"${EWDK_INCLUDE}\" CACHE STRING \"\" FORCE)\n")
	cmake.WriteString("set(CMAKE_LIBRARY_PATH \"${EWDK_LIB}\" CACHE STRING \"\" FORCE)\n")

	cmake.WriteString("\nlist(APPEND CMAKE_PROGRAM_PATH \"${EWDK_COMMON_NINJA_DIR}\")\n")
	cmake.WriteString(fmt.Sprintf("list(APPEND CMAKE_PROGRAM_PATH \"%s\")\n", cm(filepath.Dir(c.CC))))
	if _, err := os.Stat(c.RC); err == nil {
		cmake.WriteString(fmt.Sprintf("list(APPEND CMAKE_PROGRAM_PATH \"%s\")\n", cm(filepath.Dir(c.RC))))
	}
	if c.MT != "" {
		cmake.WriteString(fmt.Sprintf("set(CMAKE_MT \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(c.MT)))
	}

	cmake.WriteString("\nset(ENV{WDKContentRoot} \"${EWDK_COMMON_WDKContentRoot}\")\n")
	cmake.WriteString("set(ENV{INCLUDE} \"${EWDK_INCLUDE}\")\n")
	cmake.WriteString("set(ENV{LIB} \"${EWDK_LIB}\")\n")

	return os.WriteFile(outputPath, []byte(cmake.String()), 0644)
}
