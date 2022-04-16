package tui

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	mycolor "github.com/apxxxxxxe/rfcui/color"
	fd "github.com/apxxxxxxe/rfcui/feed"
	myio "github.com/apxxxxxxe/rfcui/io"

	"github.com/rivo/tview"
)

type MainWidget struct {
	Table  *tview.Table
	Groups []*fd.Group
	Feeds  []*fd.Feed
}

func (m *MainWidget) SaveFeed(f *fd.Feed) error {
	if f.IsMerged() {
		return nil
	}

	b, err := fd.EncodeFeed(f)
	if err != nil {
		return err
	}
	feedLink, err := f.GetFeedLink()
	if err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(feedLink)))

	if err := myio.SaveBytes(b, filepath.Join(getDataPath()[0], hash)); err != nil {
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

func (m *MainWidget) SaveGroup(g *fd.Group) error {
	b, err := fd.EncodeGroup(g)
	if err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(g.Title)))

	if err := myio.SaveBytes(b, filepath.Join(getDataPath()[1], hash)); err != nil {
		return err
	}

	return nil
}

func (m *MainWidget) SaveGroups() error {
	for _, g := range m.Groups {
		if err := m.SaveGroup(g); err != nil {
			return err
		}
	}
	return nil
}

func (m *MainWidget) LoadGroups(path string) error {
	for _, file := range myio.DirWalk(path) {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		m.Groups = append(m.Groups, fd.DecodeGroup(b))
	}
	return nil
}

func (m *MainWidget) DeleteSelection() error {
	row, _ := m.Table.GetSelection()
	if err := m.DeleteItem(row); err != nil {
		return err
	}
	return nil
}

func (m *MainWidget) DeleteItem(index int) error {
	v := m.Feeds[index]

	m.deleteFeed(index)

	var dataPath, hash string
	if v.IsMerged() {
		dataPath = getDataPath()[1]
		hash = fmt.Sprintf("%x", md5.Sum([]byte(v.Title)))
		if err := m.deleteGroup(v.Title); err != nil {
			return err
		}
	} else {
		dataPath = getDataPath()[0]
		feedLink, err := v.GetFeedLink()
		if err != nil {
			return err
		}
		hash = fmt.Sprintf("%x", md5.Sum([]byte(feedLink)))
	}

	if err := os.Remove(filepath.Join(dataPath, hash)); err != nil {
		return ErrRmFailed
	}
	return nil
}

func (m *MainWidget) deleteFeed(i int) {
	m.Feeds = append(m.Feeds[:i], m.Feeds[i+1:]...)
}

func (m *MainWidget) deleteGroup(title string) error {
	if err := m.deleteGroupData(title); err != nil {
		return err
	}
	for i, g := range m.Groups {
		if title == g.Title {
			m.Groups = append(m.Groups[:i], m.Groups[i+1:]...)
		}
	}
	return nil
}

func (m *MainWidget) deleteGroupData(title string) error {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(title)))
	if err := os.Remove(filepath.Join(getDataPath()[1], hash)); err != nil {
		return ErrRmFailed
	}
	return nil
}

func (m *MainWidget) sortFeeds() {
	sort.Slice(m.Feeds, func(i, j int) bool {
		a := []byte(m.Feeds[i].Title)
		b := []byte(m.Feeds[j].Title)
		return bytes.Compare(a, b) == -1
	})
	sort.Slice(m.Feeds, func(i, j int) bool {
		return m.Feeds[i].IsMerged() && !m.Feeds[j].IsMerged()
	})
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

func (m *MainWidget) setGroups() error {

	for i, f := range m.Feeds {
		if f.IsMerged() {
			m.deleteFeed(i)
		}
	}

	results := []*fd.Feed{}
	for _, g := range m.Groups {
		feeds := []*fd.Feed{}
		for _, link := range g.FeedLinks {
			for _, f := range m.Feeds {
				feedLink, err := f.GetFeedLink()
				if err != nil {
					return err
				}
				if link == feedLink {
					feeds = append(feeds, f)
				}
			}
		}
		mergedFeed, err := fd.MergeFeeds(feeds, g.Title)
		if err != nil {
			return err
		}
		results = append(results, mergedFeed)
	}
	m.Feeds = append(m.Feeds, results...)
	return nil
}
