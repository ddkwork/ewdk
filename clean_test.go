package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ddkwork/golibrary/std/assert"
)

func TestFixPath(t *testing.T) {
	root := "/mnt/ewdk"
	path := filepath.Join(root, "Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/bin/amd64/msdia140.dll")
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, root)
	path = filepath.ToSlash(path)
	assert.Equal(t, "Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/bin/amd64/msdia140.dll", path)
}

/*
2025-06-19 13:44:58  Success ->                    root:  │ /mnt/ewdk //main.Walk+0x174 /home/runner/work/ewdk/ewdk/clean.go:72
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/bin/amd64/msdia140.dll
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/bin/amd64/msdia140.dll
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/bin/amd64/msdia140.dll
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/include
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/include
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/include
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/lib/amd64
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/lib/amd64
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/lib/amd64
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/Callback.h
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/Callback.h
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/Callback.h
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.cpp
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.cpp
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/DIA2Dump.cpp
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.h
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.h
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/DIA2Dump.h
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.sln
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.sln
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/DIA2Dump.sln
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.vcxproj
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.vcxproj
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/DIA2Dump.vcxproj
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.vcxproj.filters
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/DIA2Dump.vcxproj.filters
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/DIA2Dump.vcxproj.filters
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/PrintSymbol.cpp
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/PrintSymbol.cpp
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/PrintSymbol.cpp
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/PrintSymbol.h
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/PrintSymbol.h
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/PrintSymbol.h
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/Readme.txt
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/Readme.txt
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/Readme.txt
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/makefile
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/makefile
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/makefile
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/regs.cpp
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/regs.cpp
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/regs.cpp
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/regs.h
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/regs.h
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/regs.h
origin path:  /mnt/ewdk/Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/stdafx.cpp
fix path trimprefix:  /Program Files/Microsoft Visual Studio/2022/BuildTools/DIA SDK/Samples/DIA2Dump/stdafx.cpp
fix path:  dist/Program_Files/Microsoft_Visual_Studio/2022/sdk/DIA_SDK/Samples/DIA2Dump/stdafx.cpp


*/
