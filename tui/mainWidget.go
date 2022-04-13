package tui

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apxxxxxxe/rfcui/feed"
	myio "github.com/apxxxxxxe/rfcui/io"

	"github.com/rivo/tview"
)

type MainWidget struct {
	Table  *tview.Table
	Groups []*feed.Group
	Feeds  []*feed.Feed
}

func (m *MainWidget) SaveFeed(f *feed.Feed) error {
	if f.Merged {
		return nil
	}

	b, err := feed.EncodeFeed(f)
	if err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(f.FeedLink)))
	myio.SaveBytes(b, filepath.Join(getDataPath()[0], hash))

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
		m.Feeds = append(m.Feeds, feed.DecodeFeed(b))
	}
	return nil
}

func (m *MainWidget) SaveGroup(g *feed.Group) error {
	b, err := feed.EncodeGroup(g)
	if err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(g.Title)))
	myio.SaveBytes(b, filepath.Join(getDataPath()[1], hash))
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
		m.Groups = append(m.Groups, feed.DecodeGroup(b))
	}
	return nil
}

func (m *MainWidget) DeleteSelection() error {
	row, _ := m.Table.GetSelection()
	v := m.Feeds[row]

	m.deleteFeed(row)

	var dataPath, hash string
	if v.Merged {
		dataPath = getDataPath()[1]
		hash = fmt.Sprintf("%x", md5.Sum([]byte(v.Title)))
	} else {
		dataPath = getDataPath()[0]
		hash = fmt.Sprintf("%x", md5.Sum([]byte(v.FeedLink)))
	}

	if err := os.Remove(filepath.Join(dataPath, hash)); err != nil {
		return err
	}
	return nil
}

func (m *MainWidget) deleteFeed(i int) {
	m.Feeds = append(m.Feeds[:i], m.Feeds[i+1:]...)
}
