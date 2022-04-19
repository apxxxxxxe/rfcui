package tui

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mycolor "github.com/apxxxxxxe/rfcui/color"
	fd "github.com/apxxxxxxe/rfcui/feed"
	myio "github.com/apxxxxxxe/rfcui/io"

	"github.com/rivo/tview"
)

type MainWidget struct {
	Table *tview.Table
	Feeds []*fd.Feed
}

func (m *MainWidget) SaveFeed(f *fd.Feed) error {
	var hash string

	b, err := fd.EncodeFeed(f)
	if err != nil {
		return err
	}

	if f.IsMerged() {
		hash = fmt.Sprintf("%x", md5.Sum([]byte(f.Title)))
	} else {
		feedLink, _ := f.GetFeedLink()
		hash = fmt.Sprintf("%x", md5.Sum([]byte(feedLink)))
	}

	if err := myio.SaveBytes(b, filepath.Join(cachePath, hash)); err != nil {
		return err
	}

	return nil
}

func (m *MainWidget) SaveFeeds() error {
	for _, f := range m.Feeds {
		if err := m.SaveFeed(f); err != nil {
			return err
		}
	}
	return nil
}

func (m *MainWidget) LoadFeeds(path string) error {
	for _, file := range myio.DirWalk(path) {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		m.Feeds = append(m.Feeds, fd.DecodeFeed(b))
	}
	return nil
}

func (m *MainWidget) DeleteSelection() error {
	row, _ := m.Table.GetSelection()
	if err := m.DeleteFeedFile(row); err != nil {
		return err
	}
	m.DeleteFeed(row)
	return nil
}

func (m *MainWidget) DeleteFeedFile(index int) error {
	var hash string

	v := m.Feeds[index]

	if v.IsMerged() {
		hash = fmt.Sprintf("%x", md5.Sum([]byte(v.Title)))
	} else {

		feedLink, _ := v.GetFeedLink()
		hash = fmt.Sprintf("%x", md5.Sum([]byte(feedLink)))
	}

	if err := os.Remove(filepath.Join(cachePath, hash)); err != nil {
		return ErrRmFailed
	}
	return nil
}

func (m *MainWidget) DeleteFeed(i int) {
	m.Feeds = append(m.Feeds[:i], m.Feeds[i+1:]...)
}

func (m *MainWidget) sortFeeds() {
	sort.Slice(m.Feeds, func(i, j int) bool {
		return strings.Compare(m.Feeds[i].Title, m.Feeds[j].Title) == -1
	})
	sort.Slice(m.Feeds, func(i, j int) bool {
		return m.Feeds[i].IsMerged() && !m.Feeds[j].IsMerged()
	})
}

func (m *MainWidget) AddMergedFeed(feeds []*fd.Feed, title string) error {
	f, err := fd.MergeFeeds(feeds, title)
	if err != nil {
		return err
	}
	m.Feeds = append(m.Feeds, f)
	if err := m.SaveFeed(f); err != nil {
		return err
	}
	return nil
}

func (m *MainWidget) setFeeds() {
	m.sortFeeds()
	table := m.Table.Clear()
	for i, feed := range m.Feeds {
		table.SetCellSimple(i, 0, feed.Title)
		if !feed.IsMerged() {
			if feed.Color < 0 || feed.Color > len(mycolor.TcellColors) {
				table.GetCell(i, 0).SetTextColor(mycolor.TcellColors[15])
			} else {
				table.GetCell(i, 0).SetTextColor(mycolor.TcellColors[feed.Color])
			}
		}
	}
	row, _ := m.Table.GetSelection()
	max := m.Table.GetRowCount() - 1
	if max < row {
		m.Table.Select(max, 0).ScrollToBeginning()
	}
}
