package main

import (
	"path/filepath"

	"github.com/xhd2015/xgo/support/filecopy"
	"github.com/xhd2015/xgo/support/fileutil"
)

func genXgoRuntime(rootDir string) error {
	runtimeDir := filepath.Join(rootDir, "runtime")
	genRuntimeDir := filepath.Join(rootDir, "cmd", "xgo", "runtime_gen")

	err := filecopy.NewOptions().Ignore("test").IncludeSuffix(".go", "go.mod").IgnoreSuffix("_test.go").CopyReplaceDir(runtimeDir, genRuntimeDir)
	if err != nil {
		return err
	}
	// err = os.Rename(filepath.Join(genRuntimeDir, "go.mod"), filepath.Join(genRuntimeDir, "go.mod.txt"))
	err = fileutil.MoveFile(filepath.Join(genRuntimeDir, "go.mod"), filepath.Join(genRuntimeDir, "go.mod.txt"))
	if err != nil {
		return err
	}
	return nil
}
