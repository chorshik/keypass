package storepass

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ebladrocher/keypass/fsutil"
	"github.com/ebladrocher/keypass/tree"
	"github.com/fatih/color"
)

// RootStore ...
type RootStore struct {
	AutoPush    bool              `json:"autopush"`
	AutoPull    bool              `json:"autopull"`
	AutoImport  bool              `json:"autoimport"`
	AlwaysTrust bool              `json:"alwaystrust"`
	NoConfirm   bool              `json:"noconfirm"`
	PersistKeys bool              `json:"persistkeys"`
	LoadKeys    bool              `json:"loadkeys"`
	ClipTimeout int               `json:"cliptimeout"`
	Path        string            `json:"path"`
	Mount       map[string]string `json:"mounts,omitempty"`
	ImportFunc  ImportCallback    `json:"-"`
	FsckFunc    FsckCallback      `json:"-"`
	store       *Store
	mounts      map[string]*Store
}

// NewRootStore ...
func NewRootStore(path string) (*RootStore, error) {
	s := &RootStore{
		Path:   path,
		Mount:  make(map[string]string),
		mounts: make(map[string]*Store),
	}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *RootStore) init() error {
	if r.Mount == nil {
		r.Mount = make(map[string]string)
	}
	if r.mounts == nil {
		r.mounts = make(map[string]*Store, len(r.Mount))
	}
	if r.Path == "" {
		return fmt.Errorf("Путь не должен быть пустым")
	}

	s, err := NewStore("", fsutil.CleanPath(r.Path), r)
	if err != nil {
		return err
	}

	r.store = s

	for alias, path := range r.Mount {
		path = fsutil.CleanPath(path)
		if err := r.addMount(alias, path); err != nil {
			fmt.Printf("Не удалось инициализировать mount %s (%s): %s. Игнорировать\n", alias, path, err)
			continue
		}
		r.Mount[alias] = path
	}

	if err := r.checkMounts(); err != nil {
		return fmt.Errorf("проверка mounts не удалась: %s", err)
	}

	if r.ClipTimeout < 1 {
		r.ClipTimeout = 45
	}

	return nil
}

// Init tries ...
func (r *RootStore) Init(store string, ids ...string) error {
	sub := r.getStore(store)
	sub.persistKeys = r.PersistKeys
	sub.loadKeys = r.LoadKeys
	sub.alwaysTrust = r.AlwaysTrust
	return sub.Init(ids...)
}

// ListRecipients ...
func (r *RootStore) ListRecipients(store string) []string {
	return r.getStore(store).recipients
}

// Initialized ...
func (r *RootStore) Initialized() bool {
	return r.store.Initialized()
}

// Exists ...
func (r *RootStore) Exists(name string) (bool, error) {
	store := r.getStore(name)
	return store.Exists(strings.TrimPrefix(name, store.alias))
}

// SetConfirm ...
func (r *RootStore) SetConfirm(name string, content []byte, cb RecipientCallback) error {
	store := r.getStore(name)
	return store.SetConfirm(strings.TrimPrefix(name, store.alias), content, cb)
}

// Get ...
func (r *RootStore) Get(name string) ([]byte, error) {
	// forward to substore
	store := r.getStore(name)
	return store.Get(strings.TrimPrefix(name, store.alias))
}

// IsDir ...
func (r *RootStore) IsDir(name string) bool {
	store := r.getStore(name)
	return store.IsDir(strings.TrimPrefix(name, store.alias))
}

// Move ...
func (r *RootStore) Move(from, to string) error {
	subFrom := r.getStore(from)
	subTo := r.getStore(to)

	// cross-store move
	if !subFrom.equals(subTo) {
		content, err := subFrom.Get(from)
		if err != nil {
			return err
		}
		if err := subTo.Set(to, content); err != nil {
			return err
		}
		if err := subFrom.Delete(from); err != nil {
			return err
		}
		return nil
	}

	from = strings.TrimPrefix(from, subFrom.alias)
	to = strings.TrimPrefix(to, subFrom.alias)
	return subFrom.Move(from, to)
}

// AddMount ...
func (r *RootStore) AddMount(alias, path string, keys ...string) error {
	path = fsutil.CleanPath(path)
	if r.Mount == nil {
		r.Mount = make(map[string]string, 1)
	}

	if _, found := r.Mount[alias]; found {
		return fmt.Errorf("%s уже примонтировано", alias)
	}

	if err := r.addMount(alias, path, keys...); err != nil {
		return err
	}
	r.Mount[alias] = path

	if err := r.checkMounts(); err != nil {
		return err
	}
	return nil
}

// Delete ...
func (r *RootStore) Delete(name string) error {
	store := r.getStore(name)
	sn := strings.TrimPrefix(name, store.alias)
	if sn == "" {
		return fmt.Errorf("не возможно удалить точку монтирования. Использовать  `keypass mount remove %s`", store.alias)
	}
	return store.Delete(sn)
}

// Prune ...
func (r *RootStore) Prune(tree string) error {
	for mp := range r.mounts {
		if strings.HasPrefix(mp, tree) {
			return fmt.Errorf("нельзя обрезать поддерево с помощью mounts. Отмонтировать сначала : `keypass mount remove %s`", mp)
		}
	}

	store := r.getStore(tree)
	return store.Prune(strings.TrimPrefix(tree, store.alias))
}

// String ...
func (r *RootStore) String() string {
	ms := make([]string, 0, len(r.mounts))
	for alias, sub := range r.mounts {
		ms = append(ms, alias+"="+sub.String())
	}
	return fmt.Sprintf("RootStore(Path: %s, Mounts: %+v)", r.store.path, strings.Join(ms, ","))
}

// GitInit ...
func (r *RootStore) GitInit(store, sk string) error {
	return r.getStore(store).GitInit(sk)
}

// Git ...
func (r *RootStore) Git(store string, args ...string) error {
	return r.getStore(store).Git(args...)
}

// List ...
func (r *RootStore) List() ([]string, error) {
	t, err := r.Tree()
	if err != nil {
		return []string{}, err
	}
	return t.List(), nil
}

// Tree ...
func (r *RootStore) Tree() (*tree.Folder, error) {
	root := tree.New("keypass")
	addFunc := func(in ...string) {
		for _, f := range in {
			if err := root.AddFile(f); err != nil {
				fmt.Printf("Не удалось добавить файл %s в дерево: %s\n", f, err)
			}
		}
	}
	mps := r.mountPoints()
	sort.Sort(sort.Reverse(byLen(mps)))
	for _, alias := range mps {
		substore := r.mounts[alias]
		if substore == nil {
			continue
		}
		if err := root.AddMount(alias, substore.path); err != nil {
			return nil, fmt.Errorf("не удалось добавить mount: %s", err)
		}
		sf, err := substore.List(alias)
		if err != nil {
			return nil, fmt.Errorf("Не удалось добавить файл: %s", err)
		}
		addFunc(sf...)
	}

	sf, err := r.store.List("")
	if err != nil {
		return nil, err
	}
	addFunc(sf...)

	return root, nil
}

func (r *RootStore) addMount(alias, path string, keys ...string) error {
	if r.mounts == nil {
		r.mounts = make(map[string]*Store, 1)
	}
	if _, found := r.mounts[alias]; found {
		return fmt.Errorf("%s уже примонтировано", alias)
	}

	s, err := NewStore(alias, fsutil.CleanPath(path), r)
	if err != nil {
		return err
	}

	if !s.Initialized() {
		if len(keys) < 1 {
			return fmt.Errorf("password store %s не инициализировано. Попробуйте keypass init", path)
		}
		if err := s.Init(keys...); err != nil {
			return err
		}
		fmt.Printf("Password store %s инициализировано для:", path)
		for _, r := range s.recipients {
			color.Yellow(r)
		}
	}

	r.mounts[alias] = s
	return nil
}

func (r *RootStore) checkMounts() error {
	paths := make(map[string]string, len(r.mounts))
	for k, v := range r.mounts {
		if _, found := paths[v.path]; found {
			return fmt.Errorf("Doubly mounted path at %s: %s", v.path, k)
		}
		paths[v.path] = k
	}
	return nil
}

func (r *RootStore) getStore(name string) *Store {
	name = strings.TrimSuffix(name, "/")
	mp := r.mountPoint(name)
	if sub, found := r.mounts[mp]; found {
		return sub
	}
	return r.store
}

func (r *RootStore) mountPoints() []string {
	mps := make([]string, 0, len(r.mounts))
	for k := range r.mounts {
		mps = append(mps, k)
	}
	sort.Sort(byLen(mps))
	return mps
}

func (r *RootStore) mountPoint(name string) string {
	for _, mp := range r.mountPoints() {
		if strings.HasPrefix(name, mp) {
			return mp
		}
	}
	return ""
}

type rootStore RootStore

// UnmarshalJSON implements a custom JSON unmarshaler
// that will also make sure the store is properly initialized
// after loading
func (r *RootStore) UnmarshalJSON(b []byte) error {
	s := rootStore{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*r = RootStore(s)
	if err := r.init(); err != nil {
		return err
	}
	return nil
}

