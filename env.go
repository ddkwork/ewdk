package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ddkwork/golibrary/std/mylog"
)

const (
	EnvWindowsTargetPlatformVersion = "WindowsTargetPlatformVersion"
	EnvVCToolsInstallDir            = "VCToolsInstallDir"
	EnvWDKContentRoot               = "WDKContentRoot"
	EnvBuildLabSetupRoot            = "BuildLabSetupRoot"
	EnvVSINSTALLDIR                 = "VSINSTALLDIR"
	EnvINCLUDE                      = "INCLUDE"
	EnvLIB                          = "LIB"
	EnvWDKBinRoot                   = "WDKBinRoot"
	EnvDiaRoot                      = "DiaRoot"
	EnvCC                           = "CC"
	EnvCXX                          = "CXX"
)

type ewdkEnv struct {
	WindowsTargetPlatformVersion string
	WDKContentRoot               string
	BuildLabSetupRoot            string
	VSINSTALLDIR                 string
	INCLUDE                      []string
	LIB                          []string
	WDKBinRoot                   string
	DiaRoot                      string
	VCToolsInstallDir            string
	CC                           string
}

const powershell = "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"

var taskName = "EWDK_Mount"

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
	letter := getEwdkDriveLetter()
	return letter != ""
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

	include := strings.Split(result[EnvINCLUDE], ";")
	lib := strings.Split(result[EnvLIB], ";")

	pre := filepath.Join(result[EnvWDKContentRoot], "Include", result[EnvWindowsTargetPlatformVersion])
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
		WindowsTargetPlatformVersion: result[EnvWindowsTargetPlatformVersion],
		WDKContentRoot:               result[EnvWDKContentRoot],
		BuildLabSetupRoot:            result[EnvBuildLabSetupRoot],
		VSINSTALLDIR:                 result[EnvVSINSTALLDIR],
		INCLUDE:                      include,
		LIB:                          lib,
		WDKBinRoot:                   result[EnvWDKBinRoot],
		DiaRoot:                      DiaRoot,
		VCToolsInstallDir:            result[EnvVCToolsInstallDir],
		CC:                           filepath.Join(result[EnvVCToolsInstallDir], "bin\\Hostx64\\x64\\cl.exe"),
	}, nil
}

func generateEwdkCmake(env ewdkEnv, outputPath string) error {
	ninjaDir, _ := filepath.Abs(filepath.Dir("ninja.exe"))
	rc := filepath.Join(env.WDKContentRoot, "bin", env.WindowsTargetPlatformVersion, "x64", "rc.exe")

	cm := func(p string) string { return strings.ReplaceAll(strings.ReplaceAll(p, `\`, "/"), "//", "/") }

	var cmake strings.Builder
	cmake.WriteString("# Auto-generated by ewdk toolchain manager - DO NOT EDIT\n")
	cmake.WriteString("# All EWDK environment variables are embedded here for portability\n\n")

	cmake.WriteString(fmt.Sprintf("set(EWDK_WDKContentRoot \"%s\")\n", cm(env.WDKContentRoot)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_WindowsTargetPlatformVersion \"%s\")\n", env.WindowsTargetPlatformVersion))
	cmake.WriteString(fmt.Sprintf("set(EWDK_VCToolsInstallDir \"%s\")\n", cm(env.VCToolsInstallDir)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_WDKBinRoot \"%s\")\n", cm(env.WDKBinRoot)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_DiaRoot \"%s\")\n", cm(env.DiaRoot)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_VSINSTALLDIR \"%s\")\n", cm(env.VSINSTALLDIR)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_BuildLabSetupRoot \"%s\")\n", cm(env.BuildLabSetupRoot)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_CC \"%s\")\n", cm(env.CC)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_CXX \"%s\")\n", cm(env.CC)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_RC \"%s\")\n", cm(rc)))
	cmake.WriteString(fmt.Sprintf("set(EWDK_NINJA_DIR \"%s\")\n", cm(ninjaDir)))

	cmake.WriteString("\nset(EWDK_INCLUDE \"")
	for i, p := range env.INCLUDE {
		if i > 0 {
			cmake.WriteString(";")
		}
		cmake.WriteString(cm(p))
	}
	cmake.WriteString("\")\n")

	cmake.WriteString("\nset(EWDK_LIB \"")
	for i, p := range env.LIB {
		if i > 0 {
			cmake.WriteString(";")
		}
		cmake.WriteString(cm(p))
	}
	cmake.WriteString("\")\n")

	cmake.WriteString("\nset(CMAKE_C_COMPILER \"${EWDK_CC}\" CACHE FILEPATH \"\" FORCE)\n")
	cmake.WriteString("set(CMAKE_CXX_COMPILER \"${EWDK_CXX}\" CACHE FILEPATH \"\" FORCE)\n")
	cmake.WriteString("set(CMAKE_RC_COMPILER \"${EWDK_RC}\" CACHE FILEPATH \"\" FORCE)\n")
	cmake.WriteString("set(CMAKE_INCLUDE_PATH \"${EWDK_INCLUDE}\" CACHE STRING \"\" FORCE)\n")
	cmake.WriteString("set(CMAKE_LIBRARY_PATH \"${EWDK_LIB}\" CACHE STRING \"\" FORCE)\n")

	cmake.WriteString("\nlist(APPEND CMAKE_PROGRAM_PATH \"${EWDK_NINJA_DIR}\")\n")
	cmake.WriteString(fmt.Sprintf("list(APPEND CMAKE_PROGRAM_PATH \"%s\")\n", cm(filepath.Dir(env.CC))))
	if _, err := os.Stat(rc); err == nil {
		cmake.WriteString(fmt.Sprintf("list(APPEND CMAKE_PROGRAM_PATH \"%s\")\n", cm(filepath.Dir(rc))))
	}

	mt := filepath.Join(env.WDKContentRoot, "bin", env.WindowsTargetPlatformVersion, "x64", "mt.exe")
	if _, err := os.Stat(mt); err == nil {
		cmake.WriteString(fmt.Sprintf("\nset(CMAKE_MT \"%s\" CACHE FILEPATH \"\" FORCE)\n", cm(mt)))
	}

	cmake.WriteString("\nset(ENV{WDKContentRoot} \"${EWDK_WDKContentRoot}\")\n")
	cmake.WriteString("set(ENV{INCLUDE} \"${EWDK_INCLUDE}\")\n")
	cmake.WriteString("set(ENV{LIB} \"${EWDK_LIB}\")\n")

	return os.WriteFile(outputPath, []byte(cmake.String()), 0644)
}
