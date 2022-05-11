package storepass

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ebladrocher/keypass/crypto/gpg"
	"github.com/ebladrocher/keypass/fsutil"
)

const (
	keyDir   = ".gpg-keys"
	fileMode = 0600
	dirMode  = 0700
)

// Load ...
func (s *Store) loadRecipients() ([]string, error) {
	f, err := os.Open(s.idFile())
	if err != nil {
		return []string{}, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Не удалось закрыть %s: %s\n", s.idFile(), err)
		}
	}()

	keys := unmarshalRecipients(f)

	for _, r := range keys {
		kl, err := gpg.ListPublicKeys(r)
		if err != nil {
			fmt.Printf("Не удалось получить открытый ключ для %s: %s\n", r, err)
			continue
		}
		if len(kl) > 0 {
			continue
		}

		if s.importFunc != nil {
			if !s.importFunc(r) {
				continue
			}
		}

		if err := s.importPublicKey(r); err != nil {
			fmt.Printf("Не удалось импортировать открытый ключ для %s: %s\n", r, err)
		}
	}

	return keys, nil
}

func unmarshalRecipients(reader io.Reader) []string {
	m := make(map[string]struct{}, 5)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			m[line] = struct{}{}
		}
	}

	lst := make([]string, 0, len(m))
	for k := range m {
		lst = append(lst, k)
	}
	sort.Strings(lst)

	return lst
}

func (s *Store) importPublicKey(r string) error {
	filename := filepath.Join(s.path, keyDir, r)
	if !fsutil.IsFile(filename) {
		return fmt.Errorf("Открытый ключ %s не найдено в %s", r, filename)
	}

	return gpg.ImportPublicKey(filename)
}

func (s *Store) exportPublicKey(r string) (string, error) {
	filename := filepath.Join(s.path, keyDir, r)
	if fsutil.IsFile(filename) {
		return filename, nil
	}

	if err := gpg.ExportPublicKey(r, filename); err != nil {
		return filename, err
	}

	return filename, nil
}

func (s *Store) saveRecipients() error {
	if err := os.MkdirAll(filepath.Dir(s.idFile()), dirMode); err != nil {
		return err
	}

	if err := ioutil.WriteFile(s.idFile(), marshalRecipients(s.recipients), fileMode); err != nil {
		return err
	}

	if !s.persistKeys {
		return nil
	}

	if err := os.MkdirAll(filepath.Join(s.path, keyDir), dirMode); err != nil {
		return err
	}

	for _, r := range s.recipients {
		path, err := s.exportPublicKey(r)
		if err != nil {
			return err
		}
		if err := s.gitAdd(path); err != nil {
			if err == ErrGitNotInit {
				continue
			}
			return err
		}
		if err := s.gitCommit(fmt.Sprintf("Exported Public Keys %s", r)); err != nil {
			return err
		}
	}

	if s.autoPush {
		if err := s.gitPush("", ""); err != nil {
			if err == ErrGitNotInit {
				return nil
			}
			return err
		}
	}

	return nil
}

func marshalRecipients(r []string) []byte {
	if len(r) == 0 {
		return []byte("\n")
	}

	m := make(map[string]struct{}, len(r))
	for _, k := range r {
		m[k] = struct{}{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := bytes.Buffer{}
	for _, k := range keys {
		_, _ = out.WriteString(k)
		_, _ = out.WriteString("\n")
	}

	return out.Bytes()
}
