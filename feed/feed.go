package feed

import (
	"math/rand"
	"sort"
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

type Item struct {
	Belong      string
	Color       int
	Title       string
	Description string
	PubDate     time.Time
	Link        string
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

func getComfortableColorIndex() int {
	var validCode = []int{
		61,
		62,
		63,
		65,
		66,
		67,
		68,
		69,
		71,
		72,
		73,
		74,
		75,
		77,
		78,
		79,
		80,
		81,
		83,
		84,
		85,
		86,
		87,
		95,
		96,
		97,
		98,
		99,
		101,
		103,
		104,
		105,
		107,
		108,
		109,
		110,
		111,
		113,
		114,
		115,
		116,
		117,
		119,
		120,
		121,
		122,
		123,
		131,
		132,
		133,
		134,
		135,
		137,
		138,
		139,
		140,
		141,
		143,
		144,
		147,
		149,
		150,
		153,
		155,
		156,
		157,
		158,
		159,
		167,
		168,
		169,
		170,
		171,
		173,
		174,
		175,
		176,
		177,
		179,
		180,
		183,
		185,
		186,
		191,
		192,
		193,
		203,
		204,
		205,
		206,
		207,
		209,
		210,
		211,
		212,
		213,
		215,
		216,
		217,
		218,
		219,
		221,
		222,
		223,
		227,
		228,
		229,
	}

	return validCode[rand.Intn(len(validCode))]
}
