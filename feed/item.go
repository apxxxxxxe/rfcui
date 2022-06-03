package feed

import (
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

const timeFormat = "2006/01/02 15:04:05"

func (a *Item) FormatDate() string {
	return a.PubDate.Format(timeFormat)
}

func (a *Item) FormatTime() string {
	const format = "15:04"
	return a.PubDate.Format(format)
}
