package action

import (
	"context"

	ctxutil "github.com/ebladrocher/keypass/ctxutils"
	"github.com/urfave/cli/v2"
)

// GetCommands ...
func (s *Action) GetCommands() []*cli.Command {
	ctx := context.Background()

	return []*cli.Command{
		{
			Name:  "generate",
			Usage: "Сгенерируйте новый пароль",
			Description: "" +
				"Сгенерировать новый пароль указанной длины" +
				"При желании поместите его в буфер обмена и очистите доску через 45 секунд",
			Before: s.Initialized,
			Action: s.Generate,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "clip,",
					Aliases: []string{"c"},
					Usage:   "Скопировать пароль в буфер обмена ",
				},
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Принудительно перезаписать существующий пароль",
				},
				&cli.BoolFlag{
					Name:    "no-symbols",
					Aliases: []string{"n"},
					Usage:   "Не использовать символы в пароле ",
				},
			},
		},
		{
			Name:  "show",
			Usage: "Показать существующий секрет и при желании поместить его в буфер обмена.",
			Description: "" +
				"Показать существующий секрет и при желании поместить его в буфер обмена. " +
				"Если поместить в буфер обмена, он будет очищен за 45 секунд.",
			Before: s.Initialized,
			Action: s.Show,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "clip",
					Aliases: []string{"c"},
					Usage:   "Скопируйте пароль в буфер обмена",
				},
			},
		},
		{
			Name:  "init",
			Usage: "Инициализируйте новое хранилище паролей",
			Description: "" +
				"Инициализируйте новое хранилище паролей и использует gpg-id для шифрования.",
			Action: s.Init,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "store",
					Aliases: []string{"s"},
					Usage:   "Set the sub store to operate on",
				},
				&cli.BoolFlag{
					Name:  "nogit",
					Usage: "Не инициализировать git репозиторий",
				},
			},
		},
		{
			Name:  "insert",
			Usage: "Вставить содержимое в секретный файл",
			Description: "" +
				"Вставить новый секрет, запись может быть многострочной",
			Action: s.Insert,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "multiline",
					Aliases: []string{"m"},
					Usage:   "Вставить с помощью $EDITOR",
				},
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Перезаписать существующий",
				},
			},
		},
		{
			Name:        "edit",
			Usage:       "Редактировать существующий секрет",
			Description: "Создать новый или редактировать существующий секрет с помощью $EDITOR.",
			Before:      s.Initialized,
			Action:      s.Edit,
		},
		{
			Name:  "delete",
			Usage: "Удалите существующий секрет",
			Description: "" +
				"Эта команда удаляет секреты. Она может рекурсивно работать с папками. ",
			Aliases: []string{"rm"},
			Before:  s.Initialized,
			Action:  s.Delete,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"rf"},
					Usage:   "Рекурсивно удалить папку",
				},
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Принудительно удалить секрет",
				},
			},
		},
		{
			Name:    "find",
			Usage:   "Перечислите секреты, соответствующие поисковому запросу.",
			Before:  s.Initialized,
			Action:  s.Find,
			Aliases: []string{"s"},
		},
		{
			Name:  "list",
			Usage: "Перечислить существующие секреты.",
			Description: "" +
				"Эта команда выведет список всех существующих секретов",
			Aliases: []string{"ls"},
			Before:  s.Initialized,
			Action:  s.List,
		},
		{
			Name:        "git",
			Usage:       "Использовать git",
			Description: "Если хранилище паролей является репозиторием git, выполните команду git, указанную git-command-args.",
			Before:      s.Initialized,
			Action:      s.Git,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "store",
					Usage: "Store to operate on",
				},
			},
			Subcommands: []*cli.Command{
				{
					Name:        "init",
					Usage:       "Инициализировать git репозиторий",
					Description: "Создайте и инициализируйте новое репозиторий git в хранилище",
					Before:      s.Initialized,
					Action:      s.GitInit,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "store",
							Usage: "Store to operate on",
						},
						&cli.StringFlag{
							Name:  "sign-key",
							Usage: "Ключ GPG для подписания коммитов",
						},
					},
				},
			},
		},
		{
			Name:        "config",
			Usage:       "Изменить конфигурацию",
			Description: "Чтобы управлять конфигурацией keypass",
			Action:      s.Config,
		},
		{
			Name:   "grep",
			Usage:  "Поиск секретов, содержащих строку поиска при расшифровке.",
			Before: s.Initialized,
			Action: s.Grep,
		},
		{
			Name:    "move",
			Aliases: []string{"mv"},
			Usage:   "Переместить секреты из одного места в другое.",
			Description: "" +
				"Эта команда перемещает секрет с одного пути на другой",
			Before: s.Initialized,
			Action: s.Move,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "force, f",
					Usage: "Принудительно переместить секрет и перезаписать существующий",
				},
			},
		},
		{
			Name:  "clone",
			Usage: "Клонировать магазин из git",
			Description: "" +
				"Эта команда клонирует существующее хранилище паролей с git remote в " +
				"локальное хранилище паролей. Может использоваться для инициализации нового корневого хранилища." +
				"" +
				"Пример:" +
				"keypass clone git@example.com/store.git example/dir",
			Action: s.Clone,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Usage: "Путь клонирования репозитория",
				},
			},
		},
		{
			Name:        "jsonapi",
			Usage:       "Запустите keypass как jsonapi, например. для плагинов браузера",
			Description: "Настройте и запустите keypass как собственные узлы обмена сообщениями, например для плагинов браузера.",
			Hidden:      true,
			Subcommands: []*cli.Command{
				{
					Name:  "listen",
					Usage: "Слушайте и отвечайте на сообщения через stdin/stdout",
					Description: "" +
						"Keypass запускается в режиме прослушивания из подключаемых модулей браузера с использованием оболочки, " +
						"указанной в манифестах узла обмена сообщениями.",
					Action: func(c *cli.Context) error {
						return s.JSONAPI(withGlobalFlags(ctx, c), c)
					},
				},
				{
					Name:  "configure",
					Usage: "Настройка манифеста встроенного обмена сообщениями keypass для выбранного браузера",
					Description: "" +
						"Чтобы получить доступ к keypass из подключаемых модулей браузера, " +
						"в правильном месте должен быть установлен собственный манифест приложения.",
					Action: func(c *cli.Context) error {
						return s.SetupNativeMessaging(withGlobalFlags(ctx, c), c)
					},
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "browser",
							Usage: "Один из 'chrome' и 'firefox'",
						},
						&cli.StringFlag{
							Name:  "path",
							Usage: "Путь для установки 'keypass_wrapper.sh'",
						},
						&cli.BoolFlag{
							Name:  "global",
							Usage: "Установить для всех пользователей, требуются права суперпользователя",
						},
						&cli.StringFlag{
							Name:  "libpath",
							Usage: "Путь к библиотеке для глобальной установки в Linux. По умолчанию /usr/lib",
						},
						&cli.BoolFlag{
							Name:  "print-only",
							Usage: "печатать только сводку по установке, но не создавать никаких файлов",
						},
					},
				},
			},
		},
	}
}

func withGlobalFlags(ctx context.Context, c *cli.Context) context.Context {
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	return ctx
}
