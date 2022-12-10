package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

var (
	GTD_DIR          = "/home/rprtr258/GTD/"
	IN_DIR           = path.Join(GTD_DIR, "in/")
	NEXT_ACTIONS_DIR = path.Join(GTD_DIR, "next_actions/")
	CALENDAR_DIR     = path.Join(GTD_DIR, "calendar/")
)

var app = cli.App{
	Name:  "gtd",
	Usage: "GTD utilities",
	Commands: []*cli.Command{{
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
			return my_open(ctx.Context, lo.Substring(param, strings.Index(param, ": ")+2, uint(len(param))))
		},
	}, {
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
			out, err := checkOutput(ctx.Context, []string{"rg", param, "-i", "--iglob", "*.md"}, GTD_DIR)
			if err != nil {
				return err
			}

			fmt.Println(out)
			return nil
		},
	}, {
		Name:  "in",
		Usage: "add to in directory",
		Action: func(ctx *cli.Context) error {
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
				return my_open(ctx.Context, path.Join(IN_DIR, param[:strings.Index(param, ": ")]))
			} else if !strings.Contains(param, ".md: ") {
				words := strings.Split(param, " ")
				// TODO: create new if already exists
				return os.WriteFile(path.Join(IN_DIR, words[0]+".md"), []byte(param+"\n"), 0o644)
			}

			return nil
		},
	}, {
		Name:  "actions",
		Usage: "list next actions",
		Action: func(ctx *cli.Context) error {
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
	}, {
		Name:  "calendar",
		Usage: "list todos with known dates",
		Action: func(ctx *cli.Context) error {
			args := ctx.Args().Slice()
			if len(args) >= 2 {
				return fmt.Errorf("more than one argument provided")
			}

			entries, err := os.ReadDir(CALENDAR_DIR)
			if err != nil {
				return err
			}

			datetimenotes := make([]DatetimeNote, 0, len(entries))
			for _, x := range entries {
				if x.IsDir() || path.Ext(x.Name()) != ".md" {
					continue
				}

				xdd, err := parse_datetimenote(path.Join(CALENDAR_DIR, x.Name()))
				if err != nil {
					return err
				}

				datetimenotes = append(datetimenotes, xdd)
			}

			sort.Slice(datetimenotes, func(i, j int) bool {
				return datetimenotes[i].date.Before(datetimenotes[j].date)
			})

			tomorrow_datetime := time.Now().AddDate(0, 0, 1)

			iterr := lo.
				If(len(args) == 0, datetimenotes).
				Else(take_while(func(dtn DatetimeNote) bool { return dtn.date.Before(tomorrow_datetime) }, datetimenotes))

			for _, line := range iterr {
				fmt.Println(line)
			}
			// TODO: open

			return nil
		},
		Subcommands: []*cli.Command{{
			Name:  "today",
			Usage: "number of tasks to do today",
			Action: func(ctx *cli.Context) error {
				entries, err := os.ReadDir(CALENDAR_DIR)
				if err != nil {
					return err
				}

				datetimenotes := make([]DatetimeNote, 0, len(entries))
				for _, x := range entries {
					if x.IsDir() || path.Ext(x.Name()) != ".md" {
						continue
					}

					xdd, err := parse_datetimenote(path.Join(CALENDAR_DIR, x.Name()))
					if err != nil {
						return err
					}

					datetimenotes = append(datetimenotes, xdd)
				}

				today := time.Now().Day()

				count := lo.CountBy(datetimenotes, func(item DatetimeNote) bool {
					return item.date.Day() == today
				})
				fmt.Println(count)

				return nil
			},
		}},
	}},
}

type DatetimeNote struct {
	title    string
	date     time.Time
	nextDate *time.Time
}

func formatDateFuckingPretty(d time.Time) string {
	return fmt.Sprintf("%8s, %2d %8s %d", d.Weekday().String(), d.Day(), d.Month().String(), d.Year())
}

func (note DatetimeNote) String() string {
	return fmt.Sprintf("%s: %30s", formatDateFuckingPretty(note.date), note.title) + lo.
		If(note.nextDate == nil, "").
		ElseF(func() string { return fmt.Sprintf(" (next in %s)", note.nextDate.Format("Monday, 02 January 2006")) }) // TODO: check/rewrite?
}

func parse_datetimenote(filename string) (DatetimeNote, error) {
	var (
		title    *string
		date     *time.Time
		new_date *time.Time
	)

	file, err := os.Open(filename)
	if err != nil {
		return DatetimeNote{}, err
	}

	buf := bufio.NewScanner(file)
	for buf.Scan() {
		line := strings.Trim(buf.Text(), " \r\n\t")
		if strings.HasPrefix(line, "# ") { // TODO: migrate to markdown metadata headers
			title = lo.ToPtr(line[2:])
		} else if strings.HasPrefix(line, "Date: ") {
			date_str := line[len("Date: "):]
			date1, err := time.Parse("02.01.2006", date_str)
			if err != nil {
				return DatetimeNote{}, fmt.Errorf("invalid date found, date=%q: %w", date_str, err)
			}
			date = &date1
		} else if strings.HasPrefix(line, "Period: ") {
			if date == nil {
				return DatetimeNote{}, errors.New("period found before date header")
			}

			period_str := line[len("Period: "):]
			var (
				period_cnt  int
				period_size byte
			)
			fmt.Sscanf(period_str, "%d%c", &period_cnt, &period_size)
			period_size2 := map[byte]func(time.Time, int) time.Time{
				'y': func(d time.Time, cnt int) time.Time { return date.AddDate(cnt, 0, 0) },
				'm': func(d time.Time, cnt int) time.Time { return date.AddDate(0, cnt, 0) },
				'w': func(d time.Time, cnt int) time.Time { return date.AddDate(0, 0, 7*cnt) },
				'd': func(d time.Time, cnt int) time.Time { return date.AddDate(0, 0, cnt) },
			}[period_size](*date, period_cnt)
			new_date = &period_size2
		}
	}

	if title == nil {
		return DatetimeNote{}, errors.New("no title found")
	}
	if date == nil {
		return DatetimeNote{}, errors.New("no date found")
	}

	return DatetimeNote{
		title:    *title,
		date:     *date,
		nextDate: new_date,
	}, nil
}

func take_while[T any](p func(T) bool, xs []T) []T {
	for i, x := range xs {
		if !p(x) {
			return xs[:i]
		}
	}
	return xs
}

func checkOutput(ctx context.Context, args []string, cwd string) (string, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	bytes, err := cmd.Output()
	return string(bytes), err
}

func run(ctx context.Context, args ...string) error {
	executable, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}

	if _, err = os.StartProcess(
		executable,
		args,
		&os.ProcAttr{
			Dir:   ".",
			Env:   os.Environ(),
			Files: []*os.File{os.Stdin, nil, nil},
			Sys:   &syscall.SysProcAttr{},
		},
	); err != nil {
		return err
	}

	return nil
}

func my_open(ctx context.Context, open_what string) error {
	return run(ctx, "/usr/bin/open", open_what)
}

func readLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	buf := bufio.NewScanner(file)
	for buf.Scan() {
		return buf.Text(), nil
	}

	return "", errors.New("no lines scanned")
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
