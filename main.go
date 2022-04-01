package main

import (

	//"github.com/apxxxxxxe/rfcui/db"
	"github.com/apxxxxxxe/rfcui/io"
	//"github.com/apxxxxxxe/rfcui/feed"

	"github.com/apxxxxxxe/rfcui/tui"

	"log"

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

type Person struct {
	Name string
	Age  int
}

func main() {

	feedURLs := io.GetLines("list.txt")

	var wg sync.WaitGroup

	t := tui.NewTui()

	//t.LoadFeeds()

	for _, url := range feedURLs {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if err := t.AddFeedFromURL(u); err != nil {
				log.Fatal(err)
			}
		}(url)
	}

	if err := t.Run(); err != nil {
		panic(err)
	}

	wg.Wait()

	//t.MainWidget.Feeds = append([]*feed.Feed{feed.MergeFeeds(t.MainWidget.Feeds, "◆AllArticles")}, t.MainWidget.Feeds...)

	return

}
