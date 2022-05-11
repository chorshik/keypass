package main

import (
	"log"
	"os"

	"github.com/ebladrocher/keypass/action"
	"github.com/urfave/cli/v2"
)

func main() {
	action := action.New()
	app := cli.NewApp()

	app.Name = action.Name

	app.Usage = "unix менеджер паролей написанный на golang"

	app.Commands = action.GetCommands()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
