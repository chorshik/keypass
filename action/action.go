package action

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ebladrocher/keypass/crypto/gpg"
	"github.com/ebladrocher/keypass/fsutil"
	"github.com/ebladrocher/keypass/storepass"
	"github.com/fatih/color"
	"github.com/ghodss/yaml"
)

// Action ...
type Action struct {
	Name  string
	Store *storepass.RootStore
}

// New ...
func New() *Action {
	if gdb := os.Getenv("KEYPASS_DEBUG"); gdb == "true" {
		gpg.Debug = true
	}
	if nc := os.Getenv("KEYPASS_NOCOLOR"); nc == "true" {
		color.NoColor = true
	}
	name := "keypass"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	pwDir := pwStoreDir("")

	if cfg, err := newFromFile(configFile()); err == nil && cfg != nil {
		cfg.ImportFunc = askForKeyImport
		return &Action{
			Name:  name,
			Store: cfg,
		}
	}

	cfg, err := storepass.NewRootStore(pwDir)
	if err != nil {
		log.Fatal(err)
	}

	cfg.ImportFunc = askForKeyImport
	cfg.FsckFunc = askForConfirmation

	return &Action{
		Name:  name,
		Store: cfg,
	}
}

func newFromFile(cf string) (*storepass.RootStore, error) {
	if _, err := os.Stat(cf); err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadFile(cf)
	if err != nil {
		fmt.Printf("Ошибка чтения конфига из %s: %s\n", cf, err)
		return nil, err
	}
	cfg := &storepass.RootStore{}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		fmt.Printf("Ошибка чтения конфига из %s: %s\n", cf, err)
		return nil, err
	}
	return cfg, nil
}

// String ...
func (s *Action) String() string {
	return s.Store.String()
}

func pwStoreDir(mount string) string {
	if d := os.Getenv("PASSWORD_STORE_DIR"); d != "" {
		return fsutil.CleanPath(d)
	}
	return os.Getenv("HOME") + "/.password-store"
}
