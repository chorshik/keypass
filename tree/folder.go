package tree

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

// Folder ...
type Folder struct {
	Name    string
	Path    string
	Root    bool
	Entries map[string]Entry
}

// IsFile ...
func (f Folder) IsFile() bool { return false }

// IsDir ...
func (f Folder) IsDir() bool { return true }

// IsMount ...
func (f Folder) IsMount() bool { return f.Path != "" }

// List ...
func (f Folder) List() []string {
	return f.list("")
}

// Format ...
func (f Folder) Format() string {
	return f.format("", true)
}

// String ...
func (f Folder) String() string {
	return f.Name
}

// AddFile ...
func (f *Folder) AddFile(name string) error {
	return f.addFile(strings.Split(name, string(filepath.Separator)))
}

// AddMount ...
func (f *Folder) AddMount(name, path string) error {
	return f.addMount(strings.Split(name, string(filepath.Separator)), path)
}

func newFolder(name string) *Folder {
	return &Folder{
		Name:    name,
		Path:    "",
		Entries: make(map[string]Entry, 10),
	}
}

func newMount(name, path string) *Folder {
	f := newFolder(name)
	f.Path = path
	return f
}

func (f Folder) list(prefix string) []string {
	out := make([]string, 0, 10)
	if !f.Root {
		if prefix != "" {
			prefix += string(filepath.Separator)
		}
		prefix += f.Name
	}
	for _, key := range sortedKeys(f.Entries) {
		out = append(out, f.Entries[key].list(prefix)...)
	}
	return out
}

func (f Folder) format(prefix string, last bool) string {
	var out *bytes.Buffer
	if f.Root {
		out = bytes.NewBufferString(f.Name)
	} else {
		out = bytes.NewBufferString(prefix)

		if last {
			_, _ = out.WriteString(symLeaf)
		} else {
			_, _ = out.WriteString(symBranch)
		}

		if f.IsMount() {
			_, _ = out.WriteString(colMount(f.Name + " (" + f.Path + ")"))
		} else {
			_, _ = out.WriteString(colDir(f.Name))
		}

		if last {
			prefix += symEmpty
		} else {
			prefix += symVert
		}
	}

	_, _ = out.WriteString("\n")

	for i, key := range sortedKeys(f.Entries) {
		last := i == len(f.Entries)-1
		_, _ = out.WriteString(f.Entries[key].format(prefix, last))
	}
	return out.String()
}

func (f *Folder) getFolder(name string) *Folder {
	if next, found := f.Entries[name]; found && next.IsDir() {
		return next.(*Folder)
	}
	next := newFolder(name)
	f.Entries[name] = next
	return next
}

// FindFolder ...
func (f *Folder) FindFolder(name string) *Folder {
	return f.findFolder(strings.Split(strings.TrimSuffix(name, "/"), "/"))
}

func (f *Folder) findFolder(path []string) *Folder {
	if len(path) < 1 {
		return f
	}
	name := path[0]
	if next, found := f.Entries[name]; found && next.IsDir() {
		if f, ok := next.(*Folder); ok {
			return f.findFolder(path[1:])
		}
	}
	return nil
}

func (f *Folder) addFile(path []string) error {
	if len(path) < 1 {
		return fmt.Errorf("Path must not be empty")
	}
	name := path[0]
	if len(path) == 1 {
		if _, found := f.Entries[name]; found {
			return fmt.Errorf("File %s exists", name)
		}
		f.Entries[name] = File(name)
		return nil
	}
	next := f.getFolder(name)
	return next.addFile(path[1:])
}

func (f *Folder) addMount(path []string, dest string) error {
	if len(path) < 1 {
		return fmt.Errorf("Path must not be empty")
	}
	name := path[0]
	if len(path) == 1 {
		if e, found := f.Entries[name]; found {
			if e.IsFile() {
				return fmt.Errorf("File %s exists", name)
			}
		}
		f.Entries[name] = newMount(name, dest)
		return nil
	}
	next := f.getFolder(name)
	return next.addMount(path[1:], dest)
}
