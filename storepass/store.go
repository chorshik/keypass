package storepass

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ebladrocher/keypass/crypto/gpg"
	"github.com/ebladrocher/keypass/fsutil"
	"github.com/fatih/color"
)

const (
	gpgID = ".gpg-id"
)

var (
	// ErrEncrypt ...
	ErrEncrypt = fmt.Errorf("Не удалось зашифровать")
	// ErrNotFound ...
	ErrNotFound = fmt.Errorf("Запись отсутствует в хранилище паролей")
	// ErrDecrypt ...
	ErrDecrypt = fmt.Errorf("Не удалось расшифровать")
	// ErrSneaky ...
	ErrSneaky = fmt.Errorf("you've attempted to pass a sneaky path to keypass. go home")
)

// RecipientCallback ...
type RecipientCallback func(string, []string) ([]string, error)

// ImportCallback ...
type ImportCallback func(string) bool

// FsckCallback ...
type FsckCallback func(string) bool

// Store ...
type Store struct {
	autoPush    bool
	autoPull    bool
	autoImport  bool
	persistKeys bool
	loadKeys    bool
	recipients  []string
	alias       string
	path        string
	alwaysTrust bool
	importFunc  ImportCallback
	fsckFunc    FsckCallback
}

// NewStore ...
func NewStore(alias, path string, r *RootStore) (*Store, error) {
	if r == nil {
		r = &RootStore{}
	}
	if path == "" {
		return nil, fmt.Errorf("Нужен путь ")
	}
	s := &Store{
		autoPush:    r.AutoPush,
		autoPull:    r.AutoPull,
		autoImport:  r.AutoImport,
		loadKeys:    r.LoadKeys,
		alias:       alias,
		path:        path,
		alwaysTrust: r.AlwaysTrust,
		importFunc:  r.ImportFunc,
		fsckFunc:    r.FsckFunc,
		recipients:  make([]string, 0, 5),
	}

	if fsutil.IsFile(s.idFile()) {
		keys, err := s.loadRecipients()
		if err != nil {
			return nil, err
		}
		s.recipients = keys
	}
	return s, nil
}

// Initialized ...
func (s *Store) Initialized() bool {
	return fsutil.IsFile(s.idFile())
}

// Init ...
func (s *Store) Init(ids ...string) error {
	if s.Initialized() {
		return fmt.Errorf("Хранилище уже инициализирован")
	}

	s.recipients = make([]string, 0, len(ids))

	for _, id := range ids {
		if id == "" {
			continue
		}
		kl, err := gpg.ListPublicKeys(id)
		if err != nil || len(kl) < 1 {
			fmt.Println("Не удалось получить открытый ключ:", id)
			continue
		}
		s.recipients = append(s.recipients, kl[0].Fingerprint)
	}

	if len(s.recipients) < 1 {
		return fmt.Errorf("не удалось инициализировать хранилище: не указаны действительные получатели")
	}

	kl, err := gpg.ListPrivateKeys(s.recipients...)
	if err != nil {
		return fmt.Errorf("Не удалось получить доступные закрытые ключи: %s", err)
	}

	if len(kl) < 1 {
		return fmt.Errorf("Ни у одного из получателей нет секретного ключа. Вы не сможете расшифровать добавленные вами секреты")
	}

	if err := s.saveRecipients(); err != nil {
		return fmt.Errorf("не удалось инициализировать хранилище: %v", err)
	}

	return nil
}

// Exists ...
func (s *Store) Exists(name string) (bool, error) {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return false, ErrSneaky
	}

	return fsutil.IsFile(p), nil
}

// Get ...
func (s *Store) Get(name string) ([]byte, error) {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return []byte{}, ErrSneaky
	}

	if !fsutil.IsFile(p) {
		return []byte{}, ErrNotFound
	}

	content, err := gpg.Decrypt(p)
	if err != nil {
		return []byte{}, ErrDecrypt
	}

	return content, nil
}

// IsDir ...
func (s *Store) IsDir(name string) bool {
	return fsutil.IsDir(filepath.Join(s.path, name))
}

// SetConfirm ...
func (s *Store) SetConfirm(name string, content []byte, cb RecipientCallback) error {
	p := s.passfile(name)

	if !strings.HasPrefix(p, s.path) {
		return ErrSneaky
	}

	if s.IsDir(name) {
		return fmt.Errorf("папка с таким именем %s уже существует", name)
	}

	recipients := make([]string, len(s.recipients))
	copy(recipients, s.recipients)

	if cb != nil {
		newRecipients, err := cb(name, recipients)
		if err != nil {
			return err
		}
		recipients = newRecipients
	}

	if err := gpg.Encrypt(p, content, recipients, s.alwaysTrust); err != nil {
		return ErrEncrypt
	}

	if err := s.gitAdd(p); err != nil {
		if err == ErrGitNotInit {
			return nil
		}
		return err
	}

	if err := s.gitCommit(fmt.Sprintf("Сохранить секрет в %s.", name)); err != nil {
		if err == ErrGitNotInit {
			return nil
		}
		return err
	}

	if s.autoPush {
		if err := s.gitPush("", ""); err != nil {
			if err == ErrGitNotInit {
				msg := "Warning: git is not initialized for this store. Ignoring auto-push option\n" +
					"Run: keypass git init"
				fmt.Println(color.RedString(msg))
				return nil
			}
			if err == ErrGitNoRemote {
				msg := "Warning: git has not remote. Ignoring auto-push option\n" +
					"Run: keypass git remote add origin ..."
				fmt.Println(color.RedString(msg))
				return nil
			}
			return err
		}
	}

	return nil
}

// Delete ...
func (s *Store) Delete(name string) error {
	return s.delete(name, false)
}

// Prune ...
func (s *Store) Prune(tree string) error {
	return s.delete(tree, true)
}

func (s *Store) delete(name string, recurse bool) error {
	path := s.passfile(name)
	rf := os.Remove
	if recurse {
		path = filepath.Join(s.path, name)
		rf = os.RemoveAll
	}

	if !recurse && !fsutil.IsFile(path) {
		return ErrNotFound
	}
	if recurse && !fsutil.IsDir(path) {
		return ErrNotFound
	}

	if err := rf(path); err != nil {
		return fmt.Errorf("Не удалось удалить секрет: %v", err)
	}

	if err := s.gitAdd(path); err != nil {
		if err == ErrGitNotInit {
			return nil
		}
		return err
	}
	if err := s.gitCommit(fmt.Sprintf("Remove %s from store.", name)); err != nil {
		if err == ErrGitNotInit {
			return nil
		}
		return err
	}

	if s.autoPush {
		if err := s.gitPush("", ""); err != nil {
			if err == ErrGitNotInit || err == ErrGitNoRemote {
				return nil
			}
			return err
		}
	}

	return nil
}

// Move ...
func (s *Store) Move(from, to string) error {
	// recursive move
	if s.IsDir(from) {
		if found, err := s.Exists(to); err != nil || found {
			return fmt.Errorf("Не удается переместить каталог в файл")
		}
		sf, err := s.List("")
		if err != nil {
			return err
		}
		destPrefix := to
		if s.IsDir(to) {
			destPrefix = filepath.Join(to, filepath.Base(from))
		}
		for _, e := range sf {
			if !strings.HasPrefix(e, strings.TrimSuffix(from, "/")+"/") {
				continue
			}
			et := filepath.Join(destPrefix, strings.TrimPrefix(e, from))
			if err := s.Move(e, et); err != nil {
				fmt.Println(err)
			}
		}
		return nil
	}

	content, err := s.Get(from)
	if err != nil {
		return err
	}
	if err := s.Set(to, content); err != nil {
		return err
	}
	if err := s.Delete(from); err != nil {
		return err
	}
	return nil
}

func (s *Store) equals(other *Store) bool {
	if other == nil {
		return false
	}
	return s.path == other.path
}

// Set ...
func (s *Store) Set(name string, content []byte) error {
	return s.SetConfirm(name, content, nil)
}

// List ...
func (s *Store) List(prefix string) ([]string, error) {
	lst := make([]string, 0, 10)
	addFunc := func(in ...string) {
		for _, s := range in {
			lst = append(lst, s)
		}
	}

	if err := filepath.Walk(s.path, mkStoreWalkerFunc(prefix, s.path, addFunc)); err != nil {
		return lst, err
	}

	return lst, nil
}

// String ...
func (s *Store) String() string {
	return fmt.Sprintf("Store(Alias: %s, Path: %s)", s.alias, s.path)
}

func (s *Store) idFile() string {
	return fsutil.CleanPath(filepath.Join(s.path, gpgID))
}

func (s *Store) passfile(name string) string {
	return fsutil.CleanPath(filepath.Join(s.path, name) + ".gpg")
}

func mkStoreWalkerFunc(alias, folder string, fn func(...string)) func(string, os.FileInfo, error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != folder {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		if path == folder {
			return nil
		}
		if path == filepath.Join(folder, gpgID) {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		s := strings.TrimPrefix(path, folder+"/")
		s = strings.TrimSuffix(s, ".gpg")
		if alias != "" {
			s = alias + "/" + s
		}
		fn(s)
		return nil
	}
}
