package main

import (
	_ "embed"
	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type (
	setup struct {
		findDiaSdk         func()
		findMsvcBuildTools func()
		findWdk            func()
	}
)
type bin struct {
	cc   string
	lib  string
	link string
	asm  string
}
type Config struct {
	userIncludes []string
	user64Libs   []string
	user32Libs   []string

	kernelIncludes []string
	kernel64Libs   []string
	kernel32Libs   []string

	Bins64   bin
	Bins32   bin
	msdia140 string
}

func main() {
	Walk()
}
func Walk() Config {
	cfg := Config{
		userIncludes:   make([]string, 0),
		user64Libs:     make([]string, 0),
		user32Libs:     make([]string, 0),
		kernelIncludes: make([]string, 0),
		kernel64Libs:   make([]string, 0),
		kernel32Libs:   make([]string, 0),
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
		root = "/mnt/ewdk"
	}
	mylog.Success("root: ", root)
	outDir := "D:/ewdk/dist"
	if stream.IsRunningOnGitHubActions() {
		outDir = "dist"
	}
	os.RemoveAll(outDir)
	BuildTools := filepath.Join(root, "Program Files", "Microsoft Visual Studio", "2022", "BuildTools")

	fnFixPath := func(path string) string {
		path = filepath.ToSlash(path)
		println("origin path: ", path)
		fixPath := strings.TrimPrefix(path, root)
		println("fix path trimprefix: ", fixPath)
		//fixPath = strings.TrimPrefix(fixPath, "Program Files\\Microsoft Visual Studio\\2022")
		fixPath = strings.TrimPrefix(fixPath, filepath.ToSlash("Program Files/Microsoft Visual Studio/2022"))
		//fixPath = strings.ReplaceAll(fixPath, "VC\\Tools\\MSVC\\14.41.34120", "")
		fixPath = strings.ReplaceAll(fixPath, filepath.ToSlash("VC\\Tools\\MSVC\\14.41.34120"), "")
		fixPath = strings.ReplaceAll(fixPath, "BuildTools", "sdk")

		fixPath = strings.ReplaceAll(fixPath, " ", "_")
		fixPath = filepath.Join(outDir, fixPath)
		fixPath = filepath.ToSlash(fixPath)
		println("fix path: ", fixPath)
		return fixPath
	}

	s := setup{
		findDiaSdk: func() {
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
		findMsvcBuildTools: func() {
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

			//dist/Program_Files/Windows_Kits/10/Lib/10.0.26100.0/km/x64
			//dist/Program_Files/Windows_Kits/10/Lib/wdf/kmdf/x64/1.35

			//dist/Program_Files/Windows_Kits/10/Include/wdf/kmdf/1.35
			//dist/Program_Files/Windows_Kits/10/Include/10.0.26100.0/shared
			//dist/Program_Files/Windows_Kits/10/Include/10.0.26100.0/km
			//dist/Program_Files/Windows_Kits/10/Include/10.0.26100.0/km/crt

			kmdfLib := filepath.Join(wdkRoot, "Lib", "wdf", "kmdf", "x64", "1.35")
			stream.CopyDir(kmdfLib, filepath.Join(outDir, "wdk", "Lib", "wdf", "kmdf", "x64", "1.35"))
			cfg.kernel64Libs = append(cfg.kernel64Libs, filepath.Join(outDir, "wdk", "Lib", "wdf", "kmdf", "x64", "1.35"))

			kmLib := filepath.Join(wdkRoot, "Lib", "10.0.26100.0", "km", "x64")
			stream.CopyDir(kmLib, filepath.Join(outDir, "wdk", "Lib", "10.0.26100.0", "km", "x64"))
			cfg.kernel64Libs = append(cfg.kernel64Libs, filepath.Join(outDir, "wdk", "Lib", "10.0.26100.0", "km", "x64"))

			kmdfInclude := filepath.Join(wdkRoot, "Include", "wdf", "kmdf", "1.35")
			stream.CopyDir(kmdfInclude, filepath.Join(outDir, "wdk", "Include", "wdf", "kmdf", "1.35"))
			cfg.kernelIncludes = append(cfg.kernelIncludes, filepath.Join(outDir, "wdk", "Include", "wdf", "kmdf", "1.35"))

			sharedInclude := filepath.Join(wdkRoot, "Include", "10.0.26100.0", "shared")
			stream.CopyDir(sharedInclude, filepath.Join(outDir, "wdk", "Include", "10.0.26100.0", "shared"))
			cfg.kernelIncludes = append(cfg.kernelIncludes, fnFixPath(sharedInclude))

			kmInclude := filepath.Join(wdkRoot, "Include", "10.0.26100.0", "km")
			stream.CopyDir(kmInclude, filepath.Join(outDir, "wdk", "Include", "10.0.26100.0", "km"))
			cfg.kernelIncludes = append(cfg.kernelIncludes, filepath.Join(outDir, "wdk", "Include", "10.0.26100.0", "km"))

			//crtInclude := filepath.Join(wdkRoot, "Include", "10.0.26100.0", "km", "crt")
			cfg.kernelIncludes = append(cfg.kernelIncludes, filepath.Join(outDir, "wdk", "Include", "10.0.26100.0", "km", "crt"))

			stream.WriteBinaryFile(filepath.Join(outDir, "wdk", "wdk.cmake"), wdkCmake)

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

			mylog.Struct(cfg)
		},
	}
	s.findDiaSdk()
	s.findMsvcBuildTools()
	s.findWdk()
	return cfg
}

//go:embed wdk.cmake
var wdkCmake string
