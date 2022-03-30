package main

import (
	//"github.com/apxxxxxxe/rfcui/db"
	"github.com/apxxxxxxe/rfcui/feed"
	"github.com/apxxxxxxe/rfcui/tui"

	"fmt"
	"io/ioutil"
	"math"
	"os"
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
		"https://shonenjumpplus.com/rss/series/3269754496501949051",
		"https://yuchrszk.blogspot.com/rss.xml",
		"https://ch.nicovideo.jp/paleo/blomaga/nico/feed",
		"https://tonarinoyj.jp/rss/series/13932016480028984490",
		"https://shonenjumpplus.com/rss/series/10833519556325021827",
		"https://tonarinoyj.jp/rss/series/3269754496306260262",
		"https://readingmonkey.blog.fc2.com/?xml",
	}

	const hasFeed = true

	// フィードをファイルにダウンロードする
	feedFiles := []string{}
	if hasFeed {
		workDir, _ := os.Getwd()
		basedir := workDir + "/feedcache"
		files, _ := ioutil.ReadDir(basedir + "/")
		for _, file := range files {
			feedFiles = append(feedFiles, basedir+"/"+file.Name())
		}
	} else {
		for i, feedURL := range feedURLs {
			fmt.Printf("\x1b[2Kdownloading %s (%d/%d)\r", feedURL, i+1, len(feedURLs))
			filename, err := feed.DownloadFeed(feedURL)
			if err != nil {
				panic(err)
			}
			feedFiles = append(feedFiles, filename)
		}
		fmt.Print("\x1b[2K\r")
	}

	// ファイルからFeedクラスを作る
	var feeds []*feed.Feed
	for _, path := range feedFiles {
		feeds = append(feeds, feed.GetFeedfromFile(path))
	}

	// すべてのフィードから記事を集めて配列を作る
	conbinedFeeds := feed.CombineFeeds(feeds, "AllArticles")

	itemNames := []string{}
	for _, item := range conbinedFeeds.Items {
		itemNames = append(itemNames, item.Title)
	}

	t := tui.NewTui()
	t.MainWidget.Feeds = feeds
	t.Notify("vim!")
	t.Run()

	// フィードを表示
	//bar()
	//for _, item := range conbinedFeeds.Items {
	//	fmt.Printf("%s [%s] \n%s\n\n", color256Sprint(item.Belong.Color, item.Belong.Title), item.PubDate.Format(timeFormat), item.Link)
	//	fmt.Println(item.Title)
	//	bar()
	//}
	return

}
