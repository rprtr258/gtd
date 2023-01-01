package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/rprtr258/gtd/internal/cmds"
)

var app = cli.App{
	Name:  "gtd",
	Usage: "GTD utilities",
	Commands: []*cli.Command{
		cmds.BookmarksCmd,
		cmds.FincCmd,
		cmds.InCmd,
		cmds.CalendarCmd,
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
