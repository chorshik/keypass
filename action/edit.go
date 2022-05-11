package action

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/ebladrocher/keypass/fsutil"
	"github.com/ebladrocher/keypass/storepass"
	"github.com/urfave/cli/v2"
)

// Edit ...
func (s *Action) Edit(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		return fmt.Errorf("укажите секретное имя ")
	}

	exists, err := s.Store.Exists(name)
	if err != nil && err != storepass.ErrNotFound {
		return fmt.Errorf("не удалось увидеть, если  %s существует", name)
	}

	var content []byte
	if exists {
		content, err = s.Store.Get(name)
		if err != nil {
			return fmt.Errorf("не удалось расшифровать %s: %v", name, err)
		}
	}

	nContent, err := s.editor(content)
	if err != nil {
		return err
	}

	if bytes.Equal(content, nContent) {
		return nil
	}

	return s.Store.SetConfirm(name, nContent, s.confirmRecipients)
}

func (s *Action) editor(content []byte) ([]byte, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return []byte{}, fmt.Errorf("не удалось отредактировать, установите  $EDITOR")
	}

	tmpfile, err := ioutil.TempFile(fsutil.Tempdir(), "keypass-edit")
	if err != nil {
		return []byte{}, fmt.Errorf("не удалось создать tmpfile для начала  %s: %v", editor, tmpfile.Name())
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			log.Fatal(err)
		}
	}()

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return []byte{}, fmt.Errorf("не удалось создать tmpfile для начала  %s: %v", editor, tmpfile.Name())
	}
	if err := tmpfile.Close(); err != nil {
		return []byte{}, fmt.Errorf("не удалось создать tmpfile для начала  %s: %v", editor, tmpfile.Name())
	}

	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return []byte{}, fmt.Errorf("не удалось запустить %s с %s файл", editor, tmpfile.Name())
	}

	nContent, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return []byte{}, fmt.Errorf("не удалось прочитать из tmpfile: %v", err)
	}

	return nContent, nil
}
