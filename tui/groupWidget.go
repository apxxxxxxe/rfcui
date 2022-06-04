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

type GroupWidget struct {
	Table  *tview.Table
	Groups []*fd.Feed
}

func (m *GroupWidget) SaveFeed(f *fd.Feed) error {
	b, err := fd.EncodeFeed(f)
	if err != nil {
		return err
	}

  hash := fmt.Sprintf("%x", md5.Sum([]byte(f.Title)))

	if err := myio.SaveBytes(b, filepath.Join(cachePath, hash)); err != nil {
		return err
	}

	return nil
}

func (m *GroupWidget) SaveFeeds() error {
	for _, f := range m.Groups {
		if err := m.SaveFeed(f); err != nil {
			return err
		}
	}
	return nil
}

func (m *GroupWidget) LoadFeeds(path string) error {
	for _, file := range myio.DirWalk(path) {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		m.Groups = append(m.Groups, fd.DecodeFeed(b))
	}
	return nil
}

func (m *GroupWidget) DeleteSelection() error {
	row, _ := m.Table.GetSelection()
	if err := m.DeleteFeedFile(row); err != nil {
		return err
	}
	m.DeleteFeed(row)
	return nil
}

func (m *GroupWidget) DeleteFeedFile(index int) error {
	var hash string

	v := m.Groups[index]

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

func (m *GroupWidget) DeleteFeed(i int) {
	m.Groups = append(m.Groups[:i], m.Groups[i+1:]...)
}

func (m *GroupWidget) sortFeeds() {
	sort.Slice(m.Groups, func(i, j int) bool {
		return strings.Compare(m.Groups[i].Title, m.Groups[j].Title) == -1
	})
	sort.Slice(m.Groups, func(i, j int) bool {
		return m.Groups[i].IsMerged() && !m.Groups[j].IsMerged()
	})
}

func (m *GroupWidget) AddMergedFeed(feeds []*fd.Feed, title string) error {
	f, err := fd.MergeFeeds(feeds, title)
	if err != nil {
		return err
	}
	m.Groups = append(m.Groups, f)
	if err := m.SaveFeed(f); err != nil {
		return err
	}
	return nil
}

func (m *GroupWidget) setFeeds() {
	m.sortFeeds()
	table := m.Table.Clear()
	for i, feed := range m.Groups {
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
