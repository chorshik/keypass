package action

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Git ...
func (s *Action) Git(c *cli.Context) error {
	store := c.String("store")
	return s.Store.Git(store, c.Args().Slice()...)
}

// GitInit ...
func (s *Action) GitInit(c *cli.Context) error {
	store := c.String("store")
	sk := c.String("sign-key")
	if sk == "" {
		s, err := askForPrivateKey("Пожалуйста, выберите ключ для подписи Git Commits")
		if err == nil {
			sk = s
		}
	}

	if err := s.Store.GitInit(store, sk); err != nil {
		return err
	}
	fmt.Println(color.GreenString("Git инициализировано"))
	return nil
}
