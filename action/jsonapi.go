package action

import (
	"context"
	"os"
	"runtime"
	"strings"

	"github.com/ebladrocher/keypass/jsonapi"
	"github.com/ebladrocher/keypass/jsonapi/manifest"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// JSONAPI reads a json message on stdin and responds on stdout
func (s *Action) JSONAPI(ctx context.Context, c *cli.Context) error {
	api := jsonapi.API{Store: s.Store, Reader: os.Stdin, Writer: os.Stdout}
	if err := api.ReadAndRespond(ctx); err != nil {
		return api.RespondError(err)
	}
	return nil
}

// SetupNativeMessaging sets up manifest for keypass as native messaging host
func (s *Action) SetupNativeMessaging(ctx context.Context, c *cli.Context) error {
	browser, err := s.getBrowser(ctx, c)
	if err != nil {
		return err
	}

	globalInstall, err := s.getGlobalInstall(ctx, c)
	if err != nil {
		return err
	}

	libpath, err := s.getLibPath(ctx, c, browser, globalInstall)
	if err != nil {
		return err
	}

	wrapperPath, err := s.getWrapperPath(ctx, c)
	if err != nil {
		return err
	}

	if err := manifest.PrintSummary(browser, wrapperPath, libpath, globalInstall); err != nil {
		return err
	}

	if c.Bool("print-only") {
		return nil
	}

	install, err := askForBool(color.BlueString("Install manifest and wrapper?"), true)
	if install && err == nil {
		return manifest.SetUp(browser, wrapperPath, libpath, globalInstall)
	}
	return err
}

func (s *Action) getBrowser(ctx context.Context, c *cli.Context) (string, error) {
	browser := c.String("browser")
	if browser != "" {
		return browser, nil
	}

	browser, err := askForString(color.BlueString("Для какого браузера вы хотите установить собственный обмен сообщениями keepass? [%s]", strings.Join(manifest.ValidBrowsers[:], ",")), manifest.DefaultBrowser)
	if err != nil {
		return "", errors.Wrapf(err, "не удалось запросить ввод пользователя")
	}
	if !stringInSlice(browser, manifest.ValidBrowsers) {
		return "", errors.Errorf("%s not one of %s", browser, strings.Join(manifest.ValidBrowsers[:], ","))
	}
	return browser, nil
}

func (s *Action) getGlobalInstall(ctx context.Context, c *cli.Context) (bool, error) {
	if !c.IsSet("global") {
		return askForBool(color.BlueString("Установить для всех пользователей? (может потребоваться sudo keypass)"), false)
	}
	return c.Bool("global"), nil
}

func (s *Action) getLibPath(ctx context.Context, c *cli.Context, browser string, global bool) (string, error) {
	if !c.IsSet("libpath") && runtime.GOOS == "linux" && browser == "firefox" && global {
		return askForString(color.BlueString("Какой у вас путь к lib"), "/usr/lib")
	}
	return c.String("libpath"), nil
}

func (s *Action) getWrapperPath(ctx context.Context, c *cli.Context) (string, error) {
	path := c.String("path")
	if path != "" {
		return path, nil
	}
	path, err := askForString(color.BlueString("По какому пути должен быть установлен keypass_wrapper.sh?"), os.Getenv("HOME")+"/.config")
	if err != nil {
		return "", errors.Wrapf(err, "не удалось запросить ввод пользователя")
	}
	return path, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
