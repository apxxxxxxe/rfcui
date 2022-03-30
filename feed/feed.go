package feed

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"io"
	"io/ioutil"
	"net/http"

	// フィード取得&フォーマット
	"github.com/mmcdole/gofeed"
)

type Feed struct {
	Group string
	Title string
	Color int
	Items []*Article
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

func GetFeedfromFile(fp string) *Feed {
	parser := gofeed.NewParser()

	bytes, err := ioutil.ReadFile(fp)
	if err != nil {
		panic(err)
	}

	parsedFeed, _ := parser.ParseString(string(bytes))
	color := rand.Intn(256)

	feed := &Feed{"", parsedFeed.Title, color, []*Article{}}

	for _, item := range parsedFeed.Items {
		feed.Items = append(feed.Items, &Article{feed, item.Title, parseTime(item.Published), item.Link})
	}

	feed.Items = formatArticles(feed.Items)

	return feed
}

func downloadFile(fp string, url string) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(fp), 0777); err != nil {
		return err
	}

	out, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func DownloadFeed(url string) (string, error) {
	workDir, _ := os.Getwd()
	basedir := workDir + "/feedcache"
	tmpfile := basedir + "/" + fmt.Sprint(time.Now().UnixNano())
	fp := gofeed.NewParser()

	if err := downloadFile(tmpfile, url); err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(tmpfile)
	if err != nil {
		return "", err
	}

	feed, err := fp.ParseString(string(data))
	if err != nil {
		return "", err
	}

	filename := strings.ReplaceAll(feed.Title, " ", "")
	filename = strings.ReplaceAll(filename, "/", "／")
	filename = basedir + "/" + filename

	if err = os.Rename(tmpfile, filename); err != nil {
		return "", err
	}
	return filename, nil
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

func CombineFeeds(feeds []*Feed, group string) *Feed {
	combinedItems := []*Article{}

	for _, feed := range feeds {
		combinedItems = append(combinedItems, feed.Items...)
	}
	combinedItems = formatArticles(combinedItems)

	return &Feed{
		Group: group,
		Title: "",
		Color: 0,
		Items: combinedItems,
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
