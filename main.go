package main

import (
	//"github.com/apxxxxxxe/rfcui/db"
	//"log"

	//"github.com/apxxxxxxe/rfcui/db"
	//"github.com/apxxxxxxe/rfcui/feed"
	"log"

	"github.com/apxxxxxxe/rfcui/tui"

	"fmt"
	"math"

	"sync"
)

func color256Sprint(num int, text string) string {
	const (
		setColor   = "\x1b[38;5;%dm"
		resetColor = "\x1b[0m"
	)
	n := int(math.Abs(float64(num))) % 256
	return fmt.Sprintf(setColor+text+resetColor, n)
}

func main() {

	feedURLs := []string{
		"https://www.corocoro.jp/rss/series/3269754496804959379",
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

	var wg sync.WaitGroup

	t := tui.NewTui()

	for _, url := range feedURLs {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if err := t.AddFeedFromURL(u); err != nil {
				log.Fatal(err)
			}
		}(url)
	}

	t.UpdateHelp("q: exit rfcui")
	if err := t.Run(); err != nil {
		panic(err)
	}

	wg.Wait()

	//t.MainWidget.Feeds = append([]*feed.Feed{feed.MergeFeeds(t.MainWidget.Feeds, "◆AllArticles")}, t.MainWidget.Feeds...)

	return

}
