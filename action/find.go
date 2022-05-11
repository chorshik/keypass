package action

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

// Find ...
func (s *Action) Find(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("Использование: keypass find arg")
	}

	l, err := s.Store.List()
	if err != nil {
		return err
	}
	for _, value := range l {
		if strings.Contains(value, c.Args().First()) {
			fmt.Println(value)
		}
	}

	return nil
}
