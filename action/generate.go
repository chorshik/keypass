package action

import (
	"fmt"
	"strconv"

	"github.com/ebladrocher/keypass/pass"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	defaultLength = 16
)

// Generate ...
func (s *Action) Generate(c *cli.Context) error {
	noSymbols := c.Bool("no-symbols")
	force := c.Bool("force")

	name := c.Args().Get(0)
	length := c.Args().Get(1)

	if name == "" {
		var err error
		name, err = askForString("Какое имя вы хотите использовать?", "")
		if err != nil || name == "" {
			return fmt.Errorf(color.RedString("укажите имя пароля"))
		}
	}

	replacing, err := s.Store.Exists(name)
	if err != nil {
		return fmt.Errorf("не удалось увидеть, если  %s существует: %s", name, err)
	}

	if length == "" {
		length = strconv.Itoa(defaultLength)
		if l, err := askForInt("Какой длины должен быть пароль ?", defaultLength); err == nil {
			length = strconv.Itoa(l)
		}
	}

	if !force {
		if replacing && !askForConfirmation(fmt.Sprintf("Запись для  %s уже существует. Перезаписать это?", name)) {
			return fmt.Errorf("не перезаписывать ваш текущий пароль ")
		}
	}

	pwlen, err := strconv.Atoi(length)
	if err != nil {
		return fmt.Errorf("длина пароля должна быть числом")
	}
	if pwlen < 1 {
		return fmt.Errorf("длина пароля должна быть больше чем  0")
	}

	password := pass.GeneratePassword(pwlen, !noSymbols)

	if err := s.Store.SetConfirm(name, password, s.confirmRecipients); err != nil {
		return err
	}

	if c.Bool("clip") {
		return s.copyToClipboard(name, password)
	}

	fmt.Printf(
		"Сгенерированный пароль для %s:\n%s\n", name,
		color.YellowString(string(password)),
	)

	return nil
}
