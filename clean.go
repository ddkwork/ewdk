package main

import (
	_ "embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
)

func main() {
	Walk()
}

const (
	kmdfVersion = "1.35"
	wdkVersion  = "10.0.26100.0"
	vcVersion   = "14.41.34120"
	vsVersion   = "2022"
)

var (
	//go:embed wdk.cmake
	wdkCmake string

	//go:embed sdk.cmake
	sdkCmake string
)

func Walk() Config {
	cfg := Config{
		userIncludes:   make([]string, 0),
		user64Libs:     make([]string, 0),
		user32Libs:     make([]string, 0),
		kernelIncludes: make([]string, 0),
		kernel64Libs:   make([]string, 0),
		Bins64: bin{
			cc:   "",
			lib:  "",
			link: "",
			asm:  "",
		},
		Bins32: bin{
			cc:   "",
			lib:  "",
			link: "",
			asm:  "",
		},
		msdia140: "",
	}

	//stream.UpdateAllLocalRep()
	root := "V:"
	if stream.IsRunningOnGitHubActions() {
		root = "/mnt/ewdk" //linux 下root前面被自动加上了空格，日了狗了
	}
	mylog.Success("root: ", root)
	outDir := "D:/ewdk/dist"
	if stream.IsRunningOnGitHubActions() {
		outDir = "dist"
	}
	mylog.Check(os.RemoveAll(outDir))
	BuildTools := filepath.Join(root, "Program Files", "Microsoft Visual Studio", "2022", "BuildTools")

	fnFixPath := func(path string) string {
		path = strings.TrimSpace(path)
		path = filepath.ToSlash(path)
		fixPath := strings.TrimPrefix(path, root)
		fixPath = strings.TrimPrefix(fixPath, "/")
		fixPath = strings.TrimPrefix(fixPath, "Program Files/Microsoft Visual Studio/"+vsVersion)
		fixPath = strings.ReplaceAll(fixPath, "/VC/Tools/MSVC/"+vcVersion, "")
		fixPath = strings.ReplaceAll(fixPath, "BuildTools", "sdk")
		fixPath = strings.ReplaceAll(fixPath, "DIA SDK", "dia")
		fixPath = filepath.Join(outDir, fixPath)
		fixPath = filepath.ToSlash(fixPath)
		println("fix path:", fixPath)
		return fixPath
	}

	s := setup{
		findDia: func() {
			once := sync.Once{}
			diaRoot := filepath.Join(BuildTools, "DIA SDK")
			filepath.Walk(diaRoot, func(path string, info fs.FileInfo, err error) error {
				if strings.Contains(path, "arm") {
					return nil
				}
				if info != nil {
					if info.IsDir() {
						return nil
					}
				}

				switch filepath.Ext(path) {
				case ".dll", ".lib":
					if !strings.Contains(path, "amd64") {
						return nil
					}
				}
				once.Do(func() {
					cfg.msdia140 = fnFixPath(filepath.Join(diaRoot, "bin", "amd64", "msdia140.dll"))
					cfg.userIncludes = append(cfg.userIncludes, fnFixPath(filepath.Join(diaRoot, "include")))
					cfg.user64Libs = append(cfg.user64Libs, fnFixPath(filepath.Join(diaRoot, "lib", "amd64")))
				})
				stream.CopyFile(path, fnFixPath(path))
				return err
			})
		},
		findSdk: func() {
			msvc := filepath.Join(BuildTools, "VC", "Tools", "MSVC")
			filepath.Walk(msvc, func(path string, info fs.FileInfo, err error) error {
				if filepath.Base(path) == "bin" {
					msvc = filepath.Dir(path)
					return nil
				}
				return err
			})

			include := filepath.Join(msvc, "include")
			fixInclude := fnFixPath(include)
			stream.CopyDir(include, fixInclude)

			bin64Root := filepath.Join(msvc, "bin", "Hostx64", "x64")
			lib64 := filepath.Join(msvc, "lib", "x64")
			stream.CopyDir(lib64, fnFixPath(lib64))
			stream.CopyDir(bin64Root, filepath.Join(outDir, "wdk", "bin"))

			bin32Root := filepath.Join(msvc, "bin", "Hostx64", "x86")
			lib32 := filepath.Join(msvc, "lib", "x86")
			stream.CopyDir(lib32, fnFixPath(lib32))

			stream.CopyDir(bin64Root, fnFixPath(bin64Root))
			stream.CopyDir(bin32Root, fnFixPath(bin32Root))

			Bins64 := bin{
				cc:   filepath.Join(bin64Root, "cl.exe"),
				lib:  filepath.Join(lib64, "lib.exe"),
				link: filepath.Join(bin64Root, "link.exe"),
				asm:  filepath.Join(bin64Root, "ml64.exe"),
			}
			Bins32 := bin{
				cc:   filepath.Join(bin32Root, "cl.exe"),
				lib:  filepath.Join(lib32, "lib.exe"),
				link: filepath.Join(bin32Root, "link.exe"),
				asm:  filepath.Join(bin32Root, "ml.exe"),
			}

			fixBin64 := bin{
				cc:   fnFixPath(Bins64.cc),
				lib:  fnFixPath(Bins64.lib),
				link: fnFixPath(Bins64.link),
				asm:  fnFixPath(Bins64.asm),
			}
			fixBin32 := bin{
				cc:   fnFixPath(Bins32.cc),
				lib:  fnFixPath(Bins32.lib),
				link: fnFixPath(Bins32.link),
				asm:  fnFixPath(Bins32.asm),
			}
			cfg.Bins64 = fixBin64
			cfg.Bins32 = fixBin32

			cfg.userIncludes = append(cfg.userIncludes, fnFixPath(include))
			cfg.user64Libs = append(cfg.user64Libs, fnFixPath(lib64))
			cfg.user32Libs = append(cfg.user32Libs, fnFixPath(lib32))
		},
		findWdk: func() {
			wdkRoot := filepath.Join(root, "Program Files", "Windows Kits", "10")

			kmdfLib := filepath.Join(wdkRoot, "Lib", "wdf", "kmdf", "x64", kmdfVersion)
			stream.CopyDir(kmdfLib, filepath.Join(outDir, "wdk", "Lib", "wdf", "kmdf", "x64", kmdfVersion))
			cfg.kernel64Libs = append(cfg.kernel64Libs, filepath.Join(outDir, "wdk", "Lib", "wdf", "kmdf", "x64", kmdfVersion))

			kmLib := filepath.Join(wdkRoot, "Lib", wdkVersion, "km", "x64")
			stream.CopyDir(kmLib, filepath.Join(outDir, "wdk", "Lib", wdkVersion, "km", "x64"))
			cfg.kernel64Libs = append(cfg.kernel64Libs, filepath.Join(outDir, "wdk", "Lib", wdkVersion, "km", "x64"))

			kmdfInclude := filepath.Join(wdkRoot, "Include", "wdf", "kmdf", kmdfVersion)
			stream.CopyDir(kmdfInclude, filepath.Join(outDir, "wdk", "Include", "wdf", "kmdf", kmdfVersion))
			cfg.kernelIncludes = append(cfg.kernelIncludes, filepath.Join(outDir, "wdk", "Include", "wdf", "kmdf", kmdfVersion))

			sharedInclude := filepath.Join(wdkRoot, "Include", wdkVersion, "shared")
			stream.CopyDir(sharedInclude, filepath.Join(outDir, "wdk", "Include", wdkVersion, "shared"))
			cfg.kernelIncludes = append(cfg.kernelIncludes, filepath.Join(outDir, "wdk", "Include", wdkVersion, "shared"))

			kmInclude := filepath.Join(wdkRoot, "Include", wdkVersion, "km")
			stream.CopyDir(kmInclude, filepath.Join(outDir, "wdk", "Include", wdkVersion, "km"))
			cfg.kernelIncludes = append(cfg.kernelIncludes, filepath.Join(outDir, "wdk", "Include", wdkVersion, "km"))
			cfg.kernelIncludes = append(cfg.kernelIncludes, filepath.Join(outDir, "wdk", "Include", wdkVersion, "km", "crt"))

			stream.WriteBinaryFile(filepath.Join(outDir, "wdk", "wdk.cmake"), wdkCmake)
			stream.WriteBinaryFile(filepath.Join(outDir, "sdk", "sdk.cmake"), sdkCmake)

			fnRemoveTelemetry := func(dir string) {
				filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
					if filepath.Ext(path) == ".dll" {
						if strings.Contains(path, "Microsoft.VisualStudio.") { //Microsoft.VisualStudio.RemoteControl.dll
							mylog.Check(os.Remove(path))
							mylog.Success("remove", path)
						}
					}
					return err
				})
			}

			fnRemoveTelemetry(filepath.Join(outDir, "/wdk/bin"))
			fnRemoveTelemetry(filepath.Join(outDir, "/sdk/bin"))

			mylog.Struct(cfg)
			//filepath.Walk(filepath.Join(wdkRoot, "Debuggers"), func(path string, info fs.FileInfo, err error) error {
			//	if strings.Contains(path, "arm") {
			//		return nil
			//	}
			//	if info != nil {
			//		if info.IsDir() {
			//			return nil
			//		}
			//	}
			//	fixPath := fnFixPath(path)
			//	stream.CopyFile(path, fixPath)
			//	return err
			//})
		},
	}
	s.findDia()
	s.findSdk()
	s.findWdk()
	return cfg
}

type (
	setup struct {
		findDia func()
		findSdk func() //以及32和64位编译器和lib,include
		findWdk func() //仅支持64位编译器和lib,include
	}
	bin struct { //todo.add.flags,give .h get filepath.dir into include dir
		cc   string
		lib  string
		link string
		asm  string
	}
	Config struct {
		userIncludes []string
		user64Libs   []string
		user32Libs   []string

		kernelIncludes []string
		kernel64Libs   []string

		Bins64   bin
		Bins32   bin
		msdia140 string
	}
)
