package tree

import (
	"fmt"
	"path/filepath"
)

// File ...
type File string

// IsFile ...
func (f File) IsFile() bool { return true }

// IsDir ...
func (f File) IsDir() bool { return false }

// IsMount ...
func (f File) IsMount() bool { return false }

// Add ...
func (f File) Add(Entry) error {
	return fmt.Errorf("%s is a file", f)
}

// String ...
func (f File) String() string {
	return string(f)
}

func (f File) list(prefix string) []string {
	return []string{filepath.Join(prefix, string(f))}
}

func (f File) format(prefix string, last bool) string {
	sym := symBranch
	if last {
		sym = symLeaf
	}
	return prefix + sym + string(f) + "\n"
}
