package action

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Grep ...
func (s *Action) Grep(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("Использование: keypass grep arg")
	}

	search := c.Args().First()

	l, err := s.Store.List()
	if err != nil {
		return err
	}

	for _, v := range l {
		content, err := s.Store.Get(v)
		if err != nil {
			fmt.Printf("не удалось расшифровать %s: %v", v, err)
			continue
		}

		if strings.Contains(string(content), search) {
			fmt.Printf("%s:\n%s", color.BlueString(v), string(content))
		}
	}

	return nil
}
