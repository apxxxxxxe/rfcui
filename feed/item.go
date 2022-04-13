package feed

import (
	"sort"
	"time"
)

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
