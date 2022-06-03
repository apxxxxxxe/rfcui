package feed

import (
	"math/rand"
	"net/url"
	"os/exec"
	"sort"
	"strings"
	"time"

	mycolor "github.com/apxxxxxxe/rfcui/color"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
)

var ErrGetFeedLinkFailed = errors.New("tried to get feed link from a merged feed")

type Feed struct {
	Title       string
	Color       int
	Description string
	Link        string
	FeedLinks   []string
	Items       []*Item
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func GetFeedFromURL(url string, forcedTitle string) (*Feed, error) {
	var (
		parsedFeed *gofeed.Feed
		feed       *Feed
		err        error
	)
	parser := gofeed.NewParser()

	if IsUrl(url) {
		parsedFeed, err = parser.ParseURL(url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		cmd := strings.Split(strings.TrimSpace(url), " ")
		output, err := exec.Command(cmd[0], cmd[1:]...).Output()
		if err != nil {
			return nil, err
		}
		parsedFeed, err = parser.ParseString(string(output))
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	color := getComfortableColorIndex()

	var title string
	if forcedTitle != "" {
		title = forcedTitle
	} else {
		title = parsedFeed.Title
	}

	feed = &Feed{
		Title:       title,
		Color:       color,
		Description: parsedFeed.Description,
		Link:        parsedFeed.Link,
		FeedLinks:   []string{url},
		Items:       []*Item{},
	}

	for _, item := range parsedFeed.Items {
		feedLink, err := feed.GetFeedLink()
		if err != nil {
			return nil, err
		}
		if time.Now().After(parseTime(item.Published)) {
			feed.Items = append(feed.Items, &Item{
				Belong:      feedLink,
				Color:       feed.Color,
				Title:       item.Title,
				Description: item.Description,
				PubDate:     parseTime(item.Published),
				Link:        item.Link,
			})
		}
	}

	feed.SortItems()

	return feed, nil
}

func (feed *Feed) GetFeedLink() (string, error) {
	if feed.IsMerged() {
		return "", ErrGetFeedLinkFailed
	}
	return feed.FeedLinks[0], nil
}

func (feed *Feed) IsMerged() bool {
	return len(feed.FeedLinks) > 1
}

func MergeFeeds(feeds []*Feed, title string) (*Feed, error) {
	mergedItems := []*Item{}
	mergedFeedlinks := []string{}

	for _, feed := range feeds {
		if !feed.IsMerged() {
			mergedItems = append(mergedItems, feed.Items...)
			feedLink, _ := feed.GetFeedLink()
			mergedFeedlinks = append(mergedFeedlinks, feedLink)
		}
	}

	resultFeed := &Feed{
		Title:       title,
		Color:       15,
		Description: "",
		Link:        "",
		FeedLinks:   mergedFeedlinks,
		Items:       mergedItems,
	}

	resultFeed.SortItems()

	return resultFeed, nil
}

func parseTime(clock string) time.Time {
	const (
		ISO8601  = "2006-01-02T15:04:05+09:00"
		ISO8601Z = "2006-01-02T15:04:05Z"
	)
	var (
		tm          time.Time
		finalFormat = ISO8601
		formats     = []string{
			ISO8601,
      ISO8601Z,
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
			break
		}
	}

	tm, _ = time.Parse(finalFormat, clock)

	return tm
}

func (feed *Feed) SortItems() {
	sort.Slice(feed.Items, func(i, j int) bool {
		a := feed.Items[i].PubDate
		b := feed.Items[j].PubDate
		return a.After(b)
	})
}

func getComfortableColorIndex() int {
	return int(mycolor.ComfortableColorCode[rand.Intn(len(mycolor.ComfortableColorCode))])
}
