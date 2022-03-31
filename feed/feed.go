package feed

import (
	"math/rand"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
)

type Feed struct {
	Group    string
	Title    string
	Color    int
	Link     string
	FeedLink string
	Items    []*Article
	Merged bool
}

type Article struct {
	Belong  *Feed
	Title   string
	PubDate time.Time
	Link    string
}

func (a *Article) FormatTime() string {
	const timeFormat = "2006/01/02 15:04:05"
	return a.PubDate.Format(timeFormat)
}

func GetFeedFromUrl(url string, forcedTitle string) *Feed {
	parser := gofeed.NewParser()

	parsedFeed, _ := parser.ParseURL(url)
	color := rand.Intn(256)

	var title string
	if forcedTitle != "" {
		title = forcedTitle
	} else {
		title = parsedFeed.Title
	}

	feed := &Feed{"", title, color, parsedFeed.Link, url, []*Article{}, false}

	for _, item := range parsedFeed.Items {
		feed.Items = append(feed.Items, &Article{feed, item.Title, parseTime(item.Published), item.Link})
	}

	feed.Items = formatArticles(feed.Items)

	return feed
}

func formatArticles(items []*Article) []*Article {
	result := make([]*Article, 0)
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

func MergeFeeds(feeds []*Feed, group string) *Feed {
	mergedItems := []*Article{}

	for _, feed := range feeds {
		mergedItems = append(mergedItems, feed.Items...)
	}
	mergedItems = formatArticles(mergedItems)

	return &Feed{
		Group:    group,
		Title:    group,
		Color:    0,
		Link:     "",
		FeedLink: "",
		Items:    mergedItems,
		Merged: true,
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
