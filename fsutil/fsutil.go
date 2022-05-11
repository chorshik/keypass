package fsutil

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// CleanPath ...
// http://stackoverflow.com/questions/17609732/expand-tilde-to-home-directory
func CleanPath(path string) string {
	if path[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		path = strings.Replace(path, "~/", dir+"/", 1)
	}
	if p, err := filepath.Abs(path); err == nil {
		return p
	}
	return filepath.Clean(path)
}

// IsDir ...
// https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
func IsDir(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// not found
			return false
		}
		fmt.Printf("не удалось проверить директорию%s: %s\n", path, err)
		return false
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		fmt.Printf("директория %s это символическая ссылка. игнорировать", path)
		return false
	}

	return fi.IsDir()
}

// IsFile ...
func IsFile(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// not found
			return false
		}
		fmt.Printf("не удалось проверить директорию %s: %s\n", path, err)
		return false
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		fmt.Printf("директория %s это символическая ссылка. игнорирование", path)
		return false
	}

	return fi.Mode().IsRegular()
}

// Tempdir ...
func Tempdir() string {
	shmDir := "/dev/shm"
	if fi, err := os.Stat(shmDir); err == nil {
		if fi.IsDir() {
			if unix.Access(shmDir, unix.W_OK) == nil {
				return shmDir
			}
		}
	}
	return ""
}
