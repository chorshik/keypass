package action

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"

	"github.com/ebladrocher/keypass/crypto/gpg"
	"golang.org/x/crypto/ssh/terminal"
)

func (s *Action) confirmRecipients(name string, recipients []string) ([]string, error) {
	if s.Store.NoConfirm {
		return recipients, nil
	}
	for {
		fmt.Printf("keypass: Шифрование %s для этих получателей:\n", name)
		sort.Strings(recipients)
		for _, r := range recipients {
			kl, err := gpg.ListPublicKeys(r)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if len(kl) < 1 {
				fmt.Println("ключ не найден", r)
				continue
			}
			fmt.Printf(" - %s\n", kl[0].OneLine())
		}
		fmt.Println("")

		yes, err := askForBool("Вы хотите продолжить?", true)
		if err != nil {
			return recipients, err
		}

		if yes {
			return recipients, nil
		}

		return recipients, fmt.Errorf("пользователь прерван")
	}
}

func askForConfirmation(text string) bool {
	for {
		if choice, err := askForBool(text, false); err == nil {
			return choice
		}
	}
}

func askForBool(text string, def bool) (bool, error) {
	choices := "y/N"
	if def {
		choices = "Y/n"
	}

	str, err := askForString(text, choices)
	if err != nil {
		return false, err
	}
	switch str {
	case "Y/n":
		return true, nil
	case "y/N":
		return false, nil
	}

	str = strings.ToLower(string(str[0]))
	switch str {
	case "y":
		return true, nil
	case "n":
		return false, nil
	default:
		return false, fmt.Errorf("Неизвестный ответ: %s", str)
	}
}

func askForKeyImport(key string) bool {
	ok, err := askForBool("Вы хотите импортировать открытый ключ '%s' в свою связку ключе?", false)
	if err != nil {
		return false
	}
	return ok
}

func askForInt(text string, def int) (int, error) {
	str, err := askForString(text, strconv.Itoa(def))
	if err != nil {
		return 0, err
	}
	intVal, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return intVal, nil
}

func askForString(text, def string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [%s]: ", text, def)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		input = def
	}
	return input, nil
}

func askForPassword(name string, askFn func(string) (string, error)) (string, error) {
	if askFn == nil {
		askFn = promptPass
	}
	for {
		pass, err := askFn(fmt.Sprintf("Введите пароль для %s", name))
		if err != nil {
			return "", err
		}

		passAgain, err := askFn(fmt.Sprintf("Повторите пароль для %s", name))
		if err != nil {
			return "", err
		}

		if pass == passAgain {
			return strings.TrimSpace(pass), nil
		}

		fmt.Println("Ошибка: введенный пароль не совпадает")
	}
}

func askForPrivateKey(prompt string) (string, error) {
	kl, err := gpg.ListPrivateKeys()
	if err != nil {
		return "", err
	}
	kl = kl.UseableKeys()
	if len(kl) < 1 {
		return "", fmt.Errorf("Не найдено ни одного пригодного для использования закрытого ключа ")
	}
	for {
		fmt.Println(prompt)
		for i, k := range kl {
			fmt.Printf("[%d] %s\n", i, k.OneLine())
		}
		iv, err := askForInt(fmt.Sprintf("Пожалуйста, введите номер ключа (0-%d)", len(kl)-1), 0)
		if err != nil {
			continue
		}
		if iv >= 0 && iv < len(kl) {
			return kl[iv].Fingerprint, nil
		}
	}
}

func promptPass(prompt string) (pass string, err error) {
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.GetState(fd)
	if err != nil {
		return "", fmt.Errorf("Не удалось получить состояние терминала: %s", err)
	}
	defer func() {
		if err := terminal.Restore(fd, oldState); err != nil {
			fmt.Printf("Не удалось восстановить терминал: %s\n", err)
		}
	}()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	go func() {
		for range sigch {
			if err := terminal.Restore(fd, oldState); err != nil {
				fmt.Printf("Не удалось восстановить терминал: %s\n", err)
			}
			os.Exit(1)
		}
	}()

	fmt.Printf("%s: ", prompt)
	passBytes, err := terminal.ReadPassword(fd)
	fmt.Println("")
	return string(passBytes), err
}
