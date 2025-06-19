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


 */
