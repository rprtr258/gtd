package cmds

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var FincCmd = &cli.Command{
	Name:  "find",
	Usage: "search through gtd files",
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()
		if len(args) >= 2 {
			return fmt.Errorf("more than one argument provided")
		}

		if len(args) == 0 {
			// TODO: open file
			return nil
		}

		param := args[0]
		// TODO: open file
		// TODO: don't search binary files
		// TODO: print results nicely
		// TODO: search in file names also
		out, err := CheckOutput(ctx.Context, []string{"rg", param, "-i", "--iglob", "*.md"}, GTD_DIR)
		if err != nil {
			return err
		}

		fmt.Println(out)
		return nil
	},
}
