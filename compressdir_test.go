package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompressDir(t *testing.T) {
	source := filepath.Clean(os.Args[1])
	source = filepath.Clean("../VSLayout")
	compressType := "zip" // os.Args[2]
	dirToCompress := filepath.Base(source)
	target := source + "." + compressType
	CompressDir(source, target, dirToCompress)
}
