package storepass

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ebladrocher/keypass/fsutil"
	"github.com/fatih/color"
)

var (
	// ErrGitInit ...
	ErrGitInit = fmt.Errorf("git уже инициализирован")
	// ErrGitNotInit ...
	ErrGitNotInit = fmt.Errorf("git не инициализирован")
	// ErrGitNoRemote ...
	ErrGitNoRemote = fmt.Errorf("git has no remote origin")
)

// Git ...
func (s *Store) Git(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// GitInit ...
func (s *Store) GitInit(signKey string) error {
	if s.isGit() {
		return ErrGitInit
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Не удалось инициализировать git: %s", err)
	}

	if err := s.gitAdd(s.path); err != nil {
		return err
	}
	if err := s.gitCommit("Add current contents of password store."); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(s.path, ".gitattributes"), []byte("*.gpg diff=gpg\n"), fileMode); err != nil {
		return fmt.Errorf("Не удалось инициализировать git: %s", err)
	}
	if err := s.gitAdd(s.path + "/.gitattributes"); err != nil {
		fmt.Println(color.YellowString("Предупреждение: не удалось добавить .gitattributes в git "))
	}
	if err := s.gitCommit("Configure git repository for gpg file diff."); err != nil {
		fmt.Println(color.YellowString("Предупреждение: не удалось зафиксировать .gitattributes в git"))
	}

	cmd = exec.Command("git", "config", "--local", "diff.gpg.binary", "true")
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Не удалось инициализировать git: %s\n", err)
	}

	if err := s.gitSetSignKey(signKey); err != nil {
		fmt.Printf("Не удалось настроить подписание Git GPG Commit: %s\n", err)
	}

	return nil
}

func (s *Store) isGit() bool {
	return fsutil.IsDir(filepath.Join(s.path, ".git"))
}

func (s *Store) gitAdd(files ...string) error {
	if !s.isGit() {
		return ErrGitNotInit
	}

	args := []string{"add", "--all"}
	args = append(args, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("не удалось добавить файлы в git: %v", err)
	}

	return nil
}

func (s *Store) gitCommit(msg string) error {
	if !s.isGit() {
		return ErrGitNotInit
	}

	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("не удалось добавить файлы в git: %v", err)
	}

	return nil
}

func (s *Store) gitSetSignKey(sk string) error {
	if sk == "" {
		return fmt.Errorf("SignKey не установлен")
	}

	cmd := exec.Command("git", "config", "--local", "user.signingkey", sk)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "config", "--local", "commit.gpgsign", "true")
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (s *Store) gitPush(remote, branch string) error {
	if !s.isGit() {
		return ErrGitNotInit
	}

	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		branch = "master"
	}

	if v, err := s.gitConfigValue("remote." + remote + ".url"); err != nil || v == "" {
		return ErrGitNoRemote
	}

	if s.autoPull {
		if err := s.Git("pull", remote, branch); err != nil {
			return err
		}
	}

	return s.Git("push", remote, branch)
}

func (s *Store) gitConfigValue(key string) (string, error) {
	if !s.isGit() {
		return "", ErrGitNotInit
	}

	buf := &bytes.Buffer{}

	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = s.path
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
