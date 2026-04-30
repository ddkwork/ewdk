package cmake

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
)

type Info struct {
	Root   string
	Bin    string
	Module string
}

func InstallInfo() Info {
	// 1. 检查cmake是否安装
	cmd := exec.Command("cmake", "--version")
	mylog.Check(cmd.Run())

	// 2. 查询cmake实际所在路径
	cmd = exec.Command("where", "cmake")
	out := mylog.Check2(cmd.CombinedOutput())
	cmakeExePath := strings.TrimSpace(string(out))

	for s := range strings.Lines(cmakeExePath) {
		println(s)

		dir := filepath.Dir(cmakeExePath) // xxx/bin
		cmakeRoot := filepath.Dir(dir)    // xxx/CMake

		mylog.Success(cmakeRoot)
		filepath.WalkDir(cmakeRoot, func(path string, d fs.DirEntry, err error) error {
			mylog.Info(path)
			return nil
		})
	}

	panic(111)
	mylog.Struct(cmakeExePath)

	// 3. 截取CMake安装根目录
	dir := filepath.Dir(cmakeExePath) // xxx/bin
	cmakeRoot := filepath.Dir(dir)    // xxx/CMake

	mylog.Success(cmakeRoot)

	filepath.WalkDir(cmakeRoot, func(path string, d fs.DirEntry, err error) error {
		mylog.Info(path)
		return nil
	})

	// 4. Windows原生全局cmake搜索路径
	// C:\Program Files\CMake\share\cmake-*\Modules
	shareDir := filepath.Join(cmakeRoot, "share")
	matches, e := filepath.Glob(filepath.Join(shareDir, "cmake-*"))
	if e != nil || len(matches) == 0 {
		mylog.Warning("未找到 cmake 模块目录")
		mylog.Check(e)
	}
	globalModuleDir := filepath.Join(matches[0], "Modules")
	if !stream.IsDir(globalModuleDir) {
		panic("模块路径不存在: " + globalModuleDir)
	}

	return Info{
		Root:   cmakeRoot,
		Bin:    dir,
		Module: globalModuleDir,
	}
}
