package action

import (
	"fmt"

	"github.com/ebladrocher/keypass/storepass"
	"github.com/urfave/cli/v2"
)

// Delete ...
func (s *Action) Delete(c *cli.Context) error {
	force := c.Bool("force")
	recursive := c.Bool("recursive")

	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("укажите секретное имя")
	}

	found, err := s.Store.Exists(name)
	if err != nil && err != storepass.ErrNotFound {
		return fmt.Errorf("не удалось увидеть, если  %s существуют", name)
	}

	if !force { // don't check if it's force anyway
		recStr := ""
		if recursive {
			recStr = "recursively "
		}
		if found && !askForConfirmation(fmt.Sprintf("Вы уверены, что хотите %sудалить %s?", recStr, name)) {
			return nil
		}
	}

	if recursive {
		return s.Store.Prune(name)
	}

	if s.Store.IsDir(name) {
		return fmt.Errorf("Невозможно удалить  '%s': Это дирректория. Используйте 'keypass rm -r %s' для удаления", name, name)
	}

	return s.Store.Delete(name)
}
