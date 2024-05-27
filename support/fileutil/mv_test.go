package fileutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMoveFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		setupSrcDestDir func() (string, string, func(), error)
		wantErr         bool
	}{
		{
			name: "source file does not exist",
			setupSrcDestDir: func() (string, string, func(), error) {
				srcFile, err := os.CreateTemp("", "source")
				if err != nil {
					return "", "", nil, err
				}
				srcPath := srcFile.Name()

				destPath := srcFile.Name() + "_dest"
				_ = os.Remove(srcPath) // remove the source file to mimic source file does not exist
				cleanup := func() {
					_ = os.Remove(destPath)
				}
				return srcPath, destPath, cleanup, nil
			},
			wantErr: true,
		},
		{
			name: "destination file already exists",
			setupSrcDestDir: func() (string, string, func(), error) {
				srcFile, err := os.CreateTemp("", "source")
				if err != nil {
					return "", "", nil, err
				}
				defer srcFile.Close()

				destFile, err := os.CreateTemp("", "dest")
				if err != nil {
					return "", "", nil, err
				}
				defer destFile.Close()

				cleanup := func() {
					// clean up temp files
					_ = os.Remove(srcFile.Name())
					_ = os.Remove(destFile.Name())
				}
				return srcFile.Name(), destFile.Name(), cleanup, nil
			},
		},
		{
			name: "move in same filesystem",
			setupSrcDestDir: func() (string, string, func(), error) {
				srcFile, err := os.CreateTemp("", "source")
				if err != nil {
					return "", "", nil, err
				}
				defer srcFile.Close()
				srcPath := srcFile.Name()

				destPath := srcPath + ".dest"
				cleanup := func() {
					_ = os.Remove(srcPath)
					_ = os.Remove(destPath)
				}
				return srcPath, destPath, cleanup, nil
			},
		},
		{
			name: "move cross filesystem",
			setupSrcDestDir: func() (string, string, func(), error) {
				if runtime.GOOS == "windows" {
					// Mount points are not common on Windows, use homedir instead.
					// This will not really move cross different filesystems.
					dir, err := os.UserHomeDir()
					if err != nil {
						return "", "", nil, err
					}
					srcFile, err := os.CreateTemp("", "source")
					if err != nil {
						return "", "", nil, err
					}
					defer srcFile.Close()

					srcPath := srcFile.Name()
					destPath := filepath.Join(dir, filepath.Base(srcPath))
					cleanup := func() {
						_ = os.Remove(srcPath)
						_ = os.Remove(destPath)
					}
					return srcPath, destPath, cleanup, nil
				} else {
					srcFile, err := os.CreateTemp("", "source")
					if err != nil {
						return "", "", nil, err
					}
					defer srcFile.Close()
					srcPath := srcFile.Name()

					// Create destination directory in current working directory
					cwd, err := os.Getwd()
					if err != nil {
						return "", "", nil, err
					}
					destDir, err := os.MkdirTemp(cwd, "destination")
					if err != nil {
						return "", "", nil, err
					}

					destPath := filepath.Join(destDir, filepath.Base(srcPath))
					cleanup := func() {
						_ = os.Remove(srcPath)
						_ = os.RemoveAll(destDir)
					}
					return srcPath, destPath, cleanup, nil
				}
			},
		},
		{
			name: "no permission to operate source or destination file",
			setupSrcDestDir: func() (string, string, func(), error) {
				srcFile, err := os.CreateTemp("", "source")
				if err != nil {
					return "", "", nil, err
				}
				// create a directory with no write permission for the current user
				var destPath string
				if runtime.GOOS == "windows" {
					destPath = `C:\Windows\System32\dest`
				} else {
					destDir, err := os.MkdirTemp("", "test")
					if err != nil {
						return "", "", nil, err
					}
					os.Chmod(destDir, 0555)
					destPath = filepath.Join(destDir, "dest")
				}
				cleanup := func() {
					_ = os.Remove(srcFile.Name())
					if runtime.GOOS != "windows" {
						_ = os.RemoveAll(filepath.Dir(destPath))
					}
				}
				return srcFile.Name(), destPath, cleanup, nil
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srcPath, destPath, cleanup, err := tc.setupSrcDestDir()
			if err != nil {
				t.Fatalf("Could not setup source and destination directories: %v", err)
			}
			defer cleanup()

			err = MoveFile(srcPath, destPath)
			if (err != nil) != tc.wantErr {
				t.Errorf("MoveFile() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil {
				// source file should not exist
				if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
					t.Errorf("Source file still exists after move")
				}
				// destination file should exist
				if _, err := os.Stat(destPath); err != nil {
					t.Errorf("Destination file does not exist after move")
				}
			}
		})
	}
}
