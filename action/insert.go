package action

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/ebladrocher/keypass/storepass"
	"github.com/urfave/cli/v2"
)

// Insert ...
func (s *Action) Insert(c *cli.Context) error {
	echo := c.Bool("echo")
	multiline := c.Bool("multiline")
	force := c.Bool("force")

	name := c.Args().Get(0)
	if name == "" {
		return fmt.Errorf("укажите секретное имя")
	}

	replacing, err := s.Store.Exists(name)
	if err != nil && err != storepass.ErrNotFound {
		return fmt.Errorf("не удалось увидеть, если %s существует", name)
	}

	if !force {
		if replacing && !askForConfirmation(fmt.Sprintf("Запись для %s уже существует. Перезаписать ее?", name)) {
			return fmt.Errorf("не перезаписывать ваш текущий секрет")
		}
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("Failed to stat stdin: %s", err)
	}

	if info.Mode()&os.ModeCharDevice == 0 {
		content := &bytes.Buffer{}

		if written, err := io.Copy(content, os.Stdin); err != nil {
			return fmt.Errorf("Не удалось скопировать после  %d байт: %s", written, err)
		}

		return s.Store.SetConfirm(name, content.Bytes(), s.confirmRecipients)
	}

	if multiline {
		content, err := s.editor([]byte{})
		if err != nil {
			return err
		}
		return s.Store.SetConfirm(name, []byte(content), s.confirmRecipients)
	}

	var promptFn func(string) (string, error)
	if echo {
		promptFn = func(prompt string) (string, error) {
			return askForString(prompt, "")
		}
	}

	content, err := askForPassword(name, promptFn)
	if err != nil {
		return fmt.Errorf("не удалось спросить пароль: %v", err)
	}

	return s.Store.SetConfirm(name, []byte(content), s.confirmRecipients)
}
