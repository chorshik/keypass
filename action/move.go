package action

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// Move ...
func (s *Action) Move(c *cli.Context) error {
	force := c.Bool("force")

	if len(c.Args().Slice()) != 2 {
		return fmt.Errorf("Использование: keypass mv old-path new-path")
	}

	from := c.Args().Slice()[0]
	to := c.Args().Slice()[1]

	if !force {
		exists, err := s.Store.Exists(to)
		if err != nil {
			return err
		}
		if exists && !askForConfirmation(fmt.Sprintf("%s уже существует. Перезаписать это?", to)) {
			return fmt.Errorf("не перезаписывать ваш текущий секрет")
		}
	}

	if err := s.Store.Move(from, to); err != nil {
		return err
	}

	return nil
}
