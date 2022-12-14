package cmds

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
	gm "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v2"

	"github.com/rprtr258/gtd/internal/rofi"
)

var CALENDAR_DIR = path.Join(GTD_DIR, "calendar/")

type DatetimeNote struct {
	filename string
	title    string
	date     time.Time
	nextDate *time.Time
}

func formatDateFuckingPretty(d time.Time) string {
	weekday := d.Weekday().String()[:2]
	day := d.Day()
	month := d.Month().String()[:3]
	year := d.Year()
	return fmt.Sprintf("%s, %2d %s %d", weekday, day, month, year)
}

func padRight(line string, cnt int) string {
	realCnt := lo.If(lo.RuneLength(line) < cnt, cnt-lo.RuneLength(line)).Else(0)
	return line + strings.Repeat(" ", realCnt)
}

func (note DatetimeNote) String() string {
	date := formatDateFuckingPretty(note.date)
	title := padRight(note.title, 30)
	nextIn := lo.
		If(note.nextDate == nil, "").
		ElseF(func() string {
			return fmt.Sprintf(" (next in %s)", formatDateFuckingPretty(*note.nextDate))
		})
	return fmt.Sprintf("%s: %s%s", date, title, nextIn)
}

func parseDatetimeNote(r io.Reader) (DatetimeNote, error) {
	file, err := io.ReadAll(r)
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

var _ = encoding.TextUnmarshaler(&myDate{})

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

var _ = encoding.TextUnmarshaler(&myPeriod{})

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

var CalendarCmd = &cli.Command{
	Name:  "calendar",
	Usage: "list todos with known dates",
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()
		if len(args) >= 2 {
			return fmt.Errorf("more than one argument provided")
		}

		if rofi.IsFirstOpen() {
			datetimenotes, err := getCalendar(CALENDAR_DIR)
			if err != nil {
				return err
			}

			sort.Slice(datetimenotes, func(i, j int) bool {
				return datetimenotes[i].date.Before(datetimenotes[j].date)
			})

			for _, line := range datetimenotes {
				rofi.YieldItemWithInfo(line.String(), line.filename)
			}

			return nil
		}

		filename := path.Join(CALENDAR_DIR, rofi.GetInfo())

		file, err := os.Open(filename)
		if err != nil {
			return err
		}

		note, err := parseDatetimeNote(file)
		if err != nil {
			return err
		}

		if note.nextDate == nil {
			return nil
		}

		input, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		lines := strings.Split(string(input), "\n")

		for i, line := range lines {
			if strings.HasPrefix(line, "date:") {
				lines[i] = fmt.Sprintf("date: %s", note.nextDate.Format("02.01.2006"))
			}
		}
		output := strings.Join(lines, "\n")
		err = os.WriteFile(filename, []byte(output), 0644)
		if err != nil {
			return err
		}

		return nil
	},
	Subcommands: []*cli.Command{{
		Name:  "today",
		Usage: "number of tasks to do today",
		Action: func(ctx *cli.Context) error {
			datetimenotes, err := getCalendar(CALENDAR_DIR)
			if err != nil {
				return err
			}

			now := time.Now()

			count := lo.CountBy(datetimenotes, func(item DatetimeNote) bool {
				switch {
				case item.date.Year() != now.Year():
					return item.date.Year() < now.Year()
				case item.date.Month() != now.Month():
					return item.date.Month() < now.Month()
				default:
					return item.date.Day() <= now.Day()
				}
			})
			fmt.Println(count)

			return nil
		},
	}},
}

func getCalendar(dir string) ([]DatetimeNote, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s failed: %w", dir, err)
	}

	datetimenotes := make([]DatetimeNote, 0, len(entries))
	for _, x := range entries {
		if x.IsDir() || path.Ext(x.Name()) != ".md" {
			continue
		}

		file, err := os.Open(path.Join(dir, x.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed opening file %s: %w", x.Name(), err)
		}

		datetimeNote, err := parseDatetimeNote(file)
		if err != nil {
			return nil, fmt.Errorf("failed parsing %s: %w", x.Name(), err)
		}

		datetimeNote.filename = x.Name()
		datetimenotes = append(datetimenotes, datetimeNote)
	}

	return datetimenotes, nil
}
