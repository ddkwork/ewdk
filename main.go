package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
)

const ewdkCmakeGenerated = "ewdk-env.cmake"

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

func resolveISOPath() string {
	githubWorkspace := os.Getenv("GITHUB_WORKSPACE")
	if githubWorkspace != "" {
		iso := filepath.Join(os.Getenv("TEMP"), "ewdk.iso")
		if !stream.FileExists(iso) {
			panic("CI env: ewdk.iso not found in TEMP")
		}
		return iso
	}
	if entries, err := os.ReadDir("."); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "EWDK") && filepath.Ext(entry.Name()) == ".iso" {
				isoPath := mylog.Check2(filepath.Abs(entry.Name()))
				if !stream.FileExists(isoPath) {
					panic("ISO file not found: " + isoPath)
				}
				return isoPath
			}
		}
	}
	panic("Could not find EWDK ISO file")
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
