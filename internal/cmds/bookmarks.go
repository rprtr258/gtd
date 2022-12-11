package cmds

import (
	"fmt"
	"path"
	"strings"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

var BookmarksCmd = &cli.Command{
	Name:  "bookmarks",
	Usage: "list browser bookmarks",
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()
		if len(args) >= 2 {
			return fmt.Errorf("more than one argument provided")
		}

		if len(args) == 0 {
			file, err := ini.Load(path.Join(GTD_DIR, "reference/", "browser_bookmarks.ini"))
			if err != nil {
				return err
			}

			for _, section := range file.Sections() {
				for i := range section.KeyStrings() {
					key := section.KeyStrings()[i]
					url := section.Keys()[i]
					fmt.Printf("%s/%s: %s\n", section.Name(), key, url)
				}
			}
			return nil
		}

		param := args[0]
		return Open(ctx.Context, lo.Substring(param, strings.Index(param, ": ")+2, uint(len(param))))
	},
}
