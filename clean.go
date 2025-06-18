package main

import (
	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type (
	info struct {
		include string
		lib     string
		bin     string
	}
	setup struct {
		findDiaSdk         func()
		findMsvcBuildTools func()
		findWdk            func()
	}
)

func Walk() {
	//var (
	//	user64Includes  []string
	//	user64Libs      []string
	//	user64Bins      []string
	//
	//	user32Includes  []string
	//	user32Libs      []string
	//	user32Bins      []string
	//
	//	kernel64Includes  []string
	//	kernel64Libs      []string
	//	kernel64Bins      []string
	//
	//	kernel32Includes  []string
	//	kernel32Libs      []string
	//	kernel32Bins      []string
	//)

}

func main() {
	//stream.UpdateAllLocalRep()
	root := "V:"
	if stream.IsRunningOnGitHubActions() {
		root = "/mnt/ewdk"
	}
	mylog.Success("root: ", root)
	//const tmp = "tmp"
	const tmp = "ewdk"
	os.RemoveAll(tmp)
	BuildTools := filepath.Join(root, "Program Files", "Microsoft Visual Studio", "2022", "BuildTools")

	fnFixPath := func(path string) string {
		fixPath := strings.TrimPrefix(path, root)
		fixPath = strings.ReplaceAll(fixPath, " ", "_")
		fixPath = filepath.Join(tmp, fixPath)
		fixPath = filepath.ToSlash(fixPath)
		return fixPath
	}

	s := setup{
		findDiaSdk: func() {
			filepath.Walk(filepath.Join(BuildTools, "DIA SDK"), func(path string, info fs.FileInfo, err error) error {
				if strings.Contains(path, "arm") {
					return nil
				}
				if info != nil {
					if info.IsDir() {
						return nil
					}
				}
				fixPath := fnFixPath(path)
				stream.CopyFile(path, fixPath)
				return err
			})
		},
		findMsvcBuildTools: func() {
			msvc := filepath.Join(BuildTools, "VC", "Tools", "MSVC")
			filepath.Walk(msvc, func(path string, info fs.FileInfo, err error) error {
				if filepath.Base(path) == "bin" {
					msvc = filepath.Dir(path)
					return nil
				}
				return err
			})

			msvc64 := info{
				include: filepath.Join(msvc, "include"),
				lib:     filepath.Join(msvc, "lib", "x64"),
				bin:     filepath.Join(msvc, "bin", "Hostx64", "x64"),
			}
			msvc32 := info{
				include: filepath.Join(msvc, "include"),
				lib:     filepath.Join(msvc, "lib", "x86"),
				bin:     filepath.Join(msvc, "bin", "Hostx64", "x86"),
			}
			fixMsvc64 := info{
				include: fnFixPath(msvc64.include),
				lib:     fnFixPath(msvc64.lib),
				bin:     fnFixPath(msvc64.bin),
			}
			fixMsvc32 := info{
				include: fnFixPath(msvc32.include),
				lib:     fnFixPath(msvc32.lib),
				bin:     fnFixPath(msvc32.bin),
			}
			stream.CopyDir(msvc64.include, fixMsvc64.include)
			stream.CopyDir(msvc64.lib, fixMsvc64.lib)
			stream.CopyDir(msvc64.bin, fixMsvc64.bin)
			//stream.CopyDir(msvc32.include, fixMsvc32.include)
			stream.CopyDir(msvc32.lib, fixMsvc32.lib)
			stream.CopyDir(msvc32.bin, fixMsvc32.bin)
		},
		findWdk: func() {
			// V:Program Files\Windows Kits\10\Include\10.0.26100.0\km\ntddk.h
			wdkRoot := filepath.Join(root, "Program Files", "Windows Kits", "10")

			filepath.Walk(filepath.Join(wdkRoot, "Debuggers"), func(path string, info fs.FileInfo, err error) error {
				if strings.Contains(path, "arm") {
					return nil
				}
				if info != nil {
					if info.IsDir() {
						return nil
					}
				}
				fixPath := fnFixPath(path)
				stream.CopyFile(path, fixPath)
				return err
			})

			msvc64 := info{
				include: filepath.Join(wdkRoot, "include"),
				lib:     filepath.Join(wdkRoot, "lib"),
			}
			//msvc32 := info{
			//	include: filepath.Join(wdkRoot, "include"),
			//	lib:     filepath.Join(wdkRoot, "lib"),
			//}
			fixMsvc64 := info{
				include: fnFixPath(msvc64.include),
				lib:     fnFixPath(msvc64.lib),
			}
			//fixMsvc32 := info{
			//	include: fnFixPath(msvc32.include),
			//	lib:     fnFixPath(msvc32.lib),
			//}
			stream.CopyDir(msvc64.include, fixMsvc64.include)
			stream.CopyDir(msvc64.lib, fixMsvc64.lib)
			//stream.CopyDir(msvc32.include, fixMsvc32.include)
			//stream.CopyDir(msvc32.lib, fixMsvc32.lib)

			return
			filepath.Walk(wdkRoot, func(path string, info fs.FileInfo, err error) error {
				mylog.Info(path)
				//if filepath.Base(path) == "include" {
				//	println(path)
				//}
				return err
			})
		},
	}
	s.findDiaSdk()
	s.findMsvcBuildTools()
	s.findWdk()
}
