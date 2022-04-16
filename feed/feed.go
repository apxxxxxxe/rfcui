package feed

import (
	"math/rand"
	"time"

	mycolor "github.com/apxxxxxxe/rfcui/color"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
)

type Feed struct {
	Title       string
	Color       int
	Description string
	Link        string
	FeedLink    string
	Items       []*Item
	Merged      bool
}

func GetFeedFromURL(url string, forcedTitle string) (*Feed, error) {
	parser := gofeed.NewParser()

	parsedFeed, err := parser.ParseURL(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	color := getComfortableColorIndex()

	var title string
	if forcedTitle != "" {
		title = forcedTitle
	} else {
		title = parsedFeed.Title
	}

	feed := &Feed{title, color, parsedFeed.Description, parsedFeed.Link, url, []*Item{}, false}

	for _, item := range parsedFeed.Items {
		feed.Items = append(feed.Items, &Item{feed.FeedLink, feed.Color, item.Title, item.Description, parseTime(item.Published), item.Link})
	}

	feed.Items = formatItems(feed.Items)

	return feed, nil
}

func MergeFeeds(feeds []*Feed, title string) *Feed {
	mergedItems := []*Item{}

	for _, feed := range feeds {
		if feed == nil {
			continue
		}
		if !feed.Merged {
			mergedItems = append(mergedItems, feed.Items...)
		}
	}
	mergedItems = formatItems(mergedItems)

	return &Feed{
		Title:       title,
		Color:       15,
		Description: "",
		Link:        "",
		FeedLink:    "",
		Items:       mergedItems,
		Merged:      true,
	}
}

func parseTime(clock string) time.Time {
	const ISO8601 = "2006-01-02T15:04:05+09:00"
	var (
		tm          time.Time
		finalFormat string
		formats     = []string{
			ISO8601,
			time.ANSIC,
			time.UnixDate,
			time.RubyDate,
			time.RFC822,
			time.RFC822Z,
			time.RFC850,
			time.RFC1123,
			time.RFC1123Z,
			time.RFC3339,
			time.RFC3339Nano,
		}
	)

	for _, format := range formats {
		if len(clock) == len(format) {
			switch len(clock) {
			case len(ISO8601):
				if clock[19:20] == "Z" {
					finalFormat = time.RFC3339
				} else {
					finalFormat = ISO8601
				}
			case len(time.RubyDate):
				if clock[3:4] == " " {
					finalFormat = time.RubyDate
				} else {
					finalFormat = time.RFC850
				}
			default:
				finalFormat = format
			}
		}
	}

	tm, _ = time.Parse(finalFormat, clock)

	return tm
}

func getComfortableColorIndex() int {
	return int(mycolor.ValidColorCode[rand.Intn(len(mycolor.ValidColorCode))])
}
