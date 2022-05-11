package action

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ebladrocher/keypass/fsutil"
	"github.com/urfave/cli/v2"
)

// Clone ...
func (s *Action) Clone(c *cli.Context) error {
	if len(c.Args().Slice()) < 1 {
		return fmt.Errorf("Использование: keypass clone repo [mount]")
	}

	repo := c.Args().Slice()[0]
	mount := ""
	if len(c.Args().Slice()) > 1 {
		mount = c.Args().Slice()[1]
	}

	path := c.String("path")
	if path == "" {
		path = pwStoreDir(mount)
	}

	if mount == "" && s.Store.Initialized() {
		return fmt.Errorf("Невозможно клонировать %s в корневое хранилище, так как это хранилище уже инициализировано.  Попробуйте клонировать submount: `keypass clone %s sub`", repo, repo)
	}

	if err := gitClone(repo, path); err != nil {
		return err
	}

	if mount != "" {
		if !s.Store.Initialized() {
			return fmt.Errorf("Root-Store не инициализирован. Сначала клонируйте или инициализируйте корневое хранилище")
		}
		fmt.Printf("Монтирование хранилища паролей %s в точке монтирования `%s` ...\n", path, mount)
		if err := s.Store.AddMount(mount, path); err != nil {
			return err
		}
	}

	if err := writeConfig(s.Store); err != nil {
		return err
	}

	fmt.Printf("Ваше хранилище паролей готово к использованию! Посмотрите: `keypass %s`\n", mount)

	return nil
}

func gitClone(repo, path string) error {
	if fsutil.IsDir(path) {
		return fmt.Errorf("%s это дирректория", path)
	}

	fmt.Printf("клонирование репозитория %s в %s ...\n", repo, path)

	cmd := exec.Command("git", "clone", repo, path)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
