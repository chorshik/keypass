package action

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// List ...
func (s *Action) List(c *cli.Context) error {
	filter := c.Args().First()

	l, err := s.Store.Tree()
	if err != nil {
		return err
	}

	if filter == "" {
		fmt.Println(l.Format())
		return nil
	}

	if subtree := l.FindFolder(filter); subtree != nil {
		subtree.Root = true
		subtree.Name = filter
		fmt.Println(subtree.Format())
		return nil
	}

	return nil
}
