package cmds

import (
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli/v2"
)

var ActionsCmd = &cli.Command{
	Name:  "actions",
	Usage: "list next actions",
	Action: func(ctx *cli.Context) error {
		var NEXT_ACTIONS_DIR = path.Join(GTD_DIR, "next_actions/")
		args := ctx.Args().Slice()
		if len(args) >= 2 {
			return fmt.Errorf("more than one argument provided")
		}

		if len(args) == 0 {
			entries, err := os.ReadDir(NEXT_ACTIONS_DIR)
			if err != nil {
				return err
			}

			for _, file := range entries {
				if file.IsDir() || path.Ext(file.Name()) != ".md" {
					continue
				}

				line, err := readLine(path.Join(NEXT_ACTIONS_DIR, file.Name()))
				if err != nil {
					return err
				}

				fmt.Printf("%s: %s\n", file.Name(), line)
			}
			return nil
		}

		// param := args[0]
		// TODO: open file/clip content
		return nil
	},
}
