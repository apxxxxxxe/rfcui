package feed

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/mmcdole/gofeed"
)

func GetFeedfromFile(fp string) *Feed {
	parser := gofeed.NewParser()

	bytes, err := ioutil.ReadFile(fp)
	if err != nil {
		panic(err)
	}

	parsedFeed, _ := parser.ParseString(string(bytes))
	color := rand.Intn(256)

	feed := &Feed{"", parsedFeed.Title, color, parsedFeed.Link, "", []*Article{}, false}

	for _, item := range parsedFeed.Items {
		feed.Items = append(feed.Items, &Article{feed, item.Title, parseTime(item.Published), item.Link})
	}

	feed.Items = formatArticles(feed.Items)

	return feed
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
