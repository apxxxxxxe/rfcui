package feed

import (
	"math/rand"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
)

type Feed struct {
	Title       string  `json:"FeedTitle"`
	Color       int     `json:"FeedColor"`
	Description string  `json:"FeedDescription"`
	Link        string  `json:"FeedLink"`
	FeedLink    string  `json:"FeedFeedLink"`
	Items       []*Item `json:"FeedItems"`
	Merged      bool    `json:"FeedMerged"`
}

type Item struct {
	Belong      string    `json:"ItemBelong"`
	Title       string    `json:"ItemTitle"`
	Description string    `json:"ItemDescription"`
	PubDate     time.Time `json:"ItemPubDate"`
	Link        string    `json:"ItemLink"`
}

func (a *Item) FormatTime() string {
	const timeFormat = "2006/01/02 15:04:05"
	return a.PubDate.Format(timeFormat)
}

func GetFeedFromURL(url string, forcedTitle string) (*Feed, error) {
	parser := gofeed.NewParser()

	parsedFeed, err := parser.ParseURL(url)
  if err != nil {
    return nil, err
  }
	color := rand.Intn(256)

	var title string
	if forcedTitle != "" {
		title = forcedTitle
	} else {
		title = parsedFeed.Title
	}

	feed := &Feed{title, color, parsedFeed.Description, parsedFeed.Link, url, []*Item{}, false}

	for _, item := range parsedFeed.Items {
		feed.Items = append(feed.Items, &Item{feed.Title, item.Title, item.Description, parseTime(item.Published), item.Link})
	}

	feed.Items = formatItems(feed.Items)

	return feed, nil
}

func formatItems(items []*Item) []*Item {
	result := make([]*Item, 0)
	now := time.Now()

	// 現在時刻より未来のフィードを除外
	for _, item := range items {
		if now.After(item.PubDate) {
			result = append(result, item)
		}
	}

	// 日付順にソート
	sort.Slice(result, func(i, j int) bool {
		return result[i].PubDate.After(result[j].PubDate)
	})

	return result
}

func MergeFeeds(feeds []*Feed, title string) *Feed {
	mergedItems := []*Item{}

	for _, feed := range feeds {
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
	} else {
		// 候補に該当しない形式はエラー
	}
	return tm
}
