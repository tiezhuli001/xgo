package fileutil

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

func MoveFile(sourcePath, destPath string) error {
	// os.Rename first
	err := os.Rename(sourcePath, destPath)
	if err == nil {
		return nil
	}

	if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err == syscall.EXDEV {
		fmt.Println("rename failed cross-device link, exec CopyAndRemoveFile()")
		return CopyAndRemoveFile(sourcePath, destPath)
	}
	return err
}

func CopyAndRemoveFile(sourcePath, destPath string) error {
	src, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %v", err)
	}
	defer src.Close()

	// preserves permissions
	fi, err := src.Stat()
	if err != nil {
		return fmt.Errorf("couldn't get source fileinfo: %v", err)
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := fi.Mode() & os.ModePerm
	dst, err := os.OpenFile(destPath, flag, perm)
	if err != nil {
		return fmt.Errorf("couldn't open dest file: %v", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("couldn't copy to dest from source: %v", err)
	}
	// for Windows, close before trying to remove: https://stackoverflow.com/a/64943554/246801
	src.Close()

	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't remove source file: %v", err)
	}
	return nil
}
