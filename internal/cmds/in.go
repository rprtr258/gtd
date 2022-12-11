package cmds

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/urfave/cli/v2"
)

var InCmd = &cli.Command{
	Name:  "in",
	Usage: "add to in directory",
	Action: func(ctx *cli.Context) error {
		var IN_DIR = path.Join(GTD_DIR, "in/")
		args := ctx.Args().Slice()
		if len(args) >= 2 {
			return fmt.Errorf("more than one argument provided")
		}

		if len(args) == 0 {
			entries, err := os.ReadDir(IN_DIR)
			if err != nil {
				return err
			}

			for _, file := range entries {
				if file.IsDir() {
					fmt.Println(file.Name())
				} else if path.Ext(file.Name()) == ".md" {
					line, err := readLine(path.Join(IN_DIR, file.Name()))
					if err != nil {
						return err
					}

					fmt.Printf("%s: %s\n", file.Name(), line)
				}
			}
			return nil
		}

		param := args[0]
		// TODO: open dir
		if strings.Contains(param, ".md") { // TODO: fix no ".md" in output
			return Open(ctx.Context, path.Join(IN_DIR, param[:strings.Index(param, ": ")]))
		} else if !strings.Contains(param, ".md: ") {
			words := strings.Split(param, " ")
			// TODO: create new if already exists
			return os.WriteFile(path.Join(IN_DIR, words[0]+".md"), []byte(param+"\n"), 0o644)
		}

		return nil
	},
}
