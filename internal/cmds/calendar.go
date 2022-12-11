package cmds

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"time"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
	gm "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v2"
)

var CALENDAR_DIR = path.Join(GTD_DIR, "calendar/")

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
	// file, err := os.Open(filename)
	file, err := os.ReadFile(filename)
	if err != nil {
		return DatetimeNote{}, err
	}

	node := gm.New().Parser().Parse(text.NewReader(file))

	var metadata string
	if true { // TODO: many safety checks
		lines := node.
			FirstChild().
			NextSibling().
			Lines()
		metadata = string(file[lines.At(0).Start:lines.At(lines.Len()-1).Stop])
	} else {
		return DatetimeNote{}, errors.New("incorrect file")
	}

	type metadataT struct {
		Title  *string   `yaml:"title"`
		Date   *myDate   `yaml:"date"`
		Period *myPeriod `yaml:"period"`
	}
	var meta metadataT
	if err := yaml.NewDecoder(bytes.NewReader([]byte(metadata + "\n"))).Decode(&meta); err != nil {
		return DatetimeNote{}, err
	}

	if meta.Date == nil {
		return DatetimeNote{}, errors.New("date is missing")
	}

	if meta.Title == nil {
		return DatetimeNote{}, errors.New("title is missing")
	}

	var new_date *time.Time
	if meta.Period != nil {
		xdd := meta.Period.Apply(time.Time(*meta.Date))
		new_date = &xdd
	}

	return DatetimeNote{
		title:    *meta.Title,
		date:     time.Time(*meta.Date),
		nextDate: new_date,
	}, nil
}

type myDate time.Time

var _ = (encoding.TextUnmarshaler)(&myDate{})

func (date *myDate) UnmarshalText(data []byte) error {
	date1, err := time.Parse("02.01.2006", string(data))
	if err != nil {
		return fmt.Errorf("invalid date, date=%q: %w", string(data), err)
	}

	*date = myDate(date1)
	return nil
}

type myPeriod struct {
	cnt  int
	size byte
}

var (
	_myPeriod myPeriod
	_         = (encoding.TextUnmarshaler)(&_myPeriod)
)

func (period *myPeriod) Apply(d time.Time) time.Time {
	switch period.size {
	case 'y':
		return d.AddDate(period.cnt, 0, 0)
	case 'm':
		return d.AddDate(0, period.cnt, 0)
	case 'w':
		return d.AddDate(0, 0, 7*period.cnt)
	case 'd':
		return d.AddDate(0, 0, period.cnt)
	default:
		panic(fmt.Errorf("invalid period %+v", period))
	}
}

func (period *myPeriod) UnmarshalText(data []byte) error {
	if _, err := fmt.Sscanf(string(data), "%d%c", &period.cnt, &period.size); err != nil {
		return fmt.Errorf("failed scanning: %w", err)
	}
	switch period.size {
	case 'y', 'm', 'w', 'd':
	default:
		return fmt.Errorf("invalid period: %s", string(data))
	}
	return nil
}

func take_while[T any](p func(T) bool, xs []T) []T {
	for i, x := range xs {
		if !p(x) {
			return xs[:i]
		}
	}
	return xs
}

var CalendarCmd = &cli.Command{
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
				return fmt.Errorf("failed parsing %s: %w", x.Name(), err)
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
}
