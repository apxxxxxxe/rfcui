package main

import (
	//"github.com/apxxxxxxe/rfcui/db"
	"sort"

	"github.com/apxxxxxxe/rfcui/feed"
	"github.com/apxxxxxxe/rfcui/tui"

	"fmt"
	"math"
	"strings"

	// terminalのwidth,height取得用
	"golang.org/x/term"
)

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

func main() {

	feedURLs := []string{
		"https://nazology.net/feed",
		"https://tonarinoyj.jp/rss/series/3269754496421404509",
		"https://nitter.domain.glass/search/rss?f=tweets&q=from%3Aapxxxxxxe",
		"https://nitter.domain.glass/search/rss?f=tweets&q=from%3ANaoS__",
		"https://nitter.domain.glass/search/rss?f=tweets&q=from%3A_nunog_",
		"https://viewer.heros-web.com/rss/series/13933686331695925339",
		"https://shonenjumpplus.com/rss/series/3269754496501949051",
		"https://yuchrszk.blogspot.com/rss.xml",
		"https://ch.nicovideo.jp/paleo/blomaga/nico/feed",
		"https://tonarinoyj.jp/rss/series/13932016480028984490",
		"https://shonenjumpplus.com/rss/series/10833519556325021827",
		"https://tonarinoyj.jp/rss/series/3269754496306260262",
		"https://readingmonkey.blog.fc2.com/?xml",
	}

	const hasFeed = true

	// urlからFeedクラスを作る
	var feeds []*feed.Feed
	for i, url := range feedURLs {
		fmt.Printf("\x1b[2Kdownloading %s (%d/%d)\r", url, i+1, len(feedURLs))
		feeds = append(feeds, feed.GetFeedFromUrl(url, ""))
	}
	fmt.Print("\x1b[2K\r")

	t := tui.NewTui()
	t.MainWidget.Feeds = feeds
	t.MainWidget.Feeds = append([]*feed.Feed{feed.MergeFeeds(feeds, "AllArticles")}, t.MainWidget.Feeds...)
	sort.Slice(t.MainWidget.Feeds, func(i, j int) bool {
		return bool(strings.Compare(t.MainWidget.Feeds[i].Title, t.MainWidget.Feeds[j].Title) == -1)
	})
	sort.Slice(t.MainWidget.Feeds, func(i, j int) bool {
		// Prioritize merged feeds
		return t.MainWidget.Feeds[i].Merged && !t.MainWidget.Feeds[j].Merged
	})
	t.UpdateHelp("q: exit rfcui")
	t.Run()

	return

}
