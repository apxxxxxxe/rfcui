package main

import (
	//"github.com/apxxxxxxe/rfcui/db"
	//"github.com/apxxxxxxe/rfcui/tui"

	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"io"
	"io/ioutil"
	"net/http"
	"os"

	"path/filepath"

	// フィード取得&フォーマット
	"github.com/mmcdole/gofeed"

	// terminalのwidth,height取得用
	"golang.org/x/term"
)

type Feed struct {
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

func getFeedfromFile(filepath string) *Feed {
	parser := gofeed.NewParser()

	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	parsedFeed, _ := parser.ParseString(string(bytes))
	color := rand.Intn(256)

	feed := &Feed{Title: parsedFeed.Title, Color: color}

	for _, item := range parsedFeed.Items {
		feed.Items = append(feed.Items, &Article{feed, item.Title, parseTime(item.Published), item.Link})
	}

	return feed
}

func color256Sprint(num int, text string) string {
	const (
		setColor   = "\x1b[38;5;%dm"
		resetColor = "\x1b[0m"
	)
	n := int(math.Abs(float64(num))) % 256
	return fmt.Sprintf(setColor+text+resetColor, n)
}

func bar() error {
	width, _, err := term.GetSize(0)
	if err != nil {
		return err
	}
	println(strings.Repeat("─", width))
	return nil
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

func downloadFeed(url string) (string, error) {
	workDir, _ := os.Getwd()
	basedir := workDir + "/feed"
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

func deleteAfterNow(items []*Article) []*Article {
	now := time.Now()
	ret := make([]*Article, 0)

	for _, item := range items {
		if now.After(item.PubDate) {
			ret = append(ret, item)
		}
	}
	return ret
}

func main() {

	const timeFormat = "2006/01/02 15:04:05"

	feedURLs := []string{
		"https://nitter.net/NJSLYR/rss",
		"https://nitter.net/NaoS__/rss",
		"https://nitter.net/apxxxxxxe/rss",
		"https://nitter.net/tyatya_1026/rss",
		"https://nitter.net/_nunog_/rss",
		"https://www.corocoro.jp/rss/series/3269754496804959379",
		"https://shonenjumpplus.com/rss/series/3269754496501949051",
		"https://yuchrszk.blogspot.com/rss.xml",
		"https://ch.nicovideo.jp/paleo/blomaga/nico/feed",
		"https://tonarinoyj.jp/rss/series/13932016480028984490",
		"https://shonenjumpplus.com/rss/series/10833519556325021827",
		"https://tonarinoyj.jp/rss/series/3269754496306260262",
		"https://readingmonkey.blog.fc2.com/?xml",
	}

	const hasFeed = false

	// フィードをファイルにダウンロードする
	feedFiles := []string{}
	if hasFeed {
		workDir, _ := os.Getwd()
		basedir := workDir + "/feed"
		files, _ := ioutil.ReadDir(basedir + "/")
		for _, file := range files {
			feedFiles = append(feedFiles, basedir+"/"+file.Name())
		}
	} else {
		for i, feedURL := range feedURLs {
			fmt.Printf("\x1b[2Kdownloading %s (%d/%d)\r", feedURL, i+1, len(feedURLs))
			filename, err := downloadFeed(feedURL)
			if err != nil {
				panic(err)
			}
			feedFiles = append(feedFiles, filename)
		}
		fmt.Print("\x1b[2K\r")
	}

	// ファイルからFeedクラスを作る
	var feeds []*Feed
	for _, path := range feedFiles {
		feeds = append(feeds, getFeedfromFile(path))
	}

	// すべてのフィードから記事を集めて配列を作る
	allFeedItems := make([]*Article, 0)
	for _, feed := range feeds {
		allFeedItems = append(allFeedItems, feed.Items...)
	}

	// 日付順にソート
	sort.Slice(allFeedItems, func(i, j int) bool {
		return allFeedItems[i].PubDate.Before(allFeedItems[j].PubDate)
	})

	// 現在時刻より未来のフィードを除外
	allFeedItems = deleteAfterNow(allFeedItems)

	// フィードを表示
	bar()
	for _, item := range allFeedItems {
		fmt.Printf("%s [%s] \n%s\n\n", color256Sprint(item.Belong.Color, item.Belong.Title), item.PubDate.Format(timeFormat), item.Link)
		fmt.Println(item.Title)
		bar()
	}
	return

}
