package feed

import (
	"math/rand"
	"time"

	"github.com/mmcdole/gofeed"
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
		return nil, err
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
		Color:       0,
		Description: "",
		Link:        "",
		FeedLink:    "",
		Items:       mergedItems,
		Merged:      true,
	}
}

func parseTime(clock string) time.Time {
	// 時刻の表示形式を一定のものに整形して返す
	const (
		ISO8601 = "2006-01-02T15:04:05+09:00"
	)
	var tm time.Time
	delimita := [3]string{clock[3:4], clock[4:5], clock[10:11]}
	if delimita[2] == "T" {
		tm, _ = time.Parse(ISO8601, clock)
	} else if delimita[0] == "," && delimita[1] == " " {
		tm, _ = time.Parse(time.RFC1123, clock)
	} // else {
	// 候補に該当しない形式はエラー
	//}
	return tm
}

func getComfortableColorIndex() int {
	return validColorCode[rand.Intn(len(validColorCode))]
}
