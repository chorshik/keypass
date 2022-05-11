package action

import (
	"fmt"

	"github.com/ebladrocher/keypass/crypto/gpg"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Initialized ...
func (s *Action) Initialized(*cli.Context) error {
	if !s.Store.Initialized() {
		return fmt.Errorf("хранилище паролей не инициализировано. Попробуйте '%s init'", s.Name)
	}
	return nil
}

// Init ...
func (s *Action) Init(c *cli.Context) error {
	store := c.String("store")
	nogit := c.Bool("nogit")

	if !hasConfig() {
		s.Store.AutoPush = true
		s.Store.AutoPull = true
		s.Store.AutoImport = false
		s.Store.NoConfirm = true
		s.Store.PersistKeys = true
		s.Store.LoadKeys = false
		s.Store.ClipTimeout = 45
	}

	keys := c.Args().Slice()
	if len(keys) < 1 {
		nk, err := askForPrivateKey("Пожалуйста, выберите закрытый ключ для шифрования:")
		if err != nil {
			return err
		}
		keys = []string{nk}
	}

	if err := s.Store.Init(store, keys...); err != nil {
		return err
	}

	fmt.Printf(color.GreenString("Хранилище паролей инициализировано для: "))
	for _, recipient := range s.Store.ListRecipients(store) {
		r := "0x" + recipient
		if kl, err := gpg.ListPublicKeys(recipient); err == nil && len(kl) > 0 {
			r = kl[0].OneLine()
		}
		color.Yellow(r)
	}
	fmt.Println("")

	if err := writeConfig(s.Store); err != nil {
		color.Red(fmt.Sprintf("Не удалось записать конфигурацию: %s", err))
	}

	if nogit {
		return nil
	}

	return s.GitInit(c)
}
