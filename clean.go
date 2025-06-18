package main

import (
	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
	"github.com/ddkwork/golibrary/std/stream/net/httpClient"
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

func getIsoLink() {
	b := httpClient.New().SetDebug(true).Get().Url("https://learn.microsoft.com/en-us/legal/windows/hardware/enterprise-wdk-license-2022").Request().Buffer
	//latestEWDK := "https://download.microsoft.com/download/65cf0837-67ba-4070-9081-dea2b75be528/iso_EWDK/EWDK_ge_release_svc_prod3_26100_250523-0801.iso"
	latestEWDK := ""
	for s := range strings.Lines(b.String()) {
		if strings.Contains(s, "Accept license terms") {
			before, after, found := strings.Cut(s, `" data-linktype`)
			if found {
				before, after, found = strings.Cut(before, `href="`)
				latestEWDK = after
				break
			}
		}
	}

	println(latestEWDK)
}

func main() {
	//stream.UpdateAllLocalRep()
	//https://go.microsoft.com/fwlink/?linkid=2324618
	//https://download.microsoft.com/download/65cf0837-67ba-4070-9081-dea2b75be528/iso_EWDK/EWDK_ge_release_svc_prod3_26100_250523-0801.iso
	root := "V:"
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
				switch {
				case info.IsDir():
					return nil
				case strings.Contains(path, "arm"):
					return nil
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
				switch {
				case info.IsDir():
					return nil
				case strings.Contains(path, "arm"):
					return nil
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
