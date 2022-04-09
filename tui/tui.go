package tui

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/apxxxxxxe/rfcui/feed"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const datapath = "feedcache"

type Tui struct {
	App        *tview.Application
	Pages      *tview.Pages
	MainWidget *MainWidget
	SubWidget  *SubWidget
	Info       *tview.TextView
	Help       *tview.TextView
}

func (t *Tui) AddFeedFromURL(url string) error {
	for _, feed := range t.MainWidget.Feeds {
		if feed.FeedLink == url {
			return errors.New("Feed already exist.")
		}
	}
	f := feed.GetFeedFromURL(url, "")
	t.setFeeds(append(t.MainWidget.Feeds, f))
	return nil
}

func (t *Tui) LoadCells(table *tview.Table, texts []string) {
	table.Clear()
	for i, text := range texts {
		table.SetCell(i, 0, tview.NewTableCell(text))
	}
}

func (t *Tui) Notify(text string) {
	t.Info.SetText(text)
}

func (t *Tui) UpdateHelp(text string) {
	t.Help.SetText(text)
}

func (t *Tui) RefreshTui() {
	if t.MainWidget.Table.HasFocus() {
		t.selectMainRow()
	} else if t.SubWidget.Table.HasFocus() {
		t.selectSubRow()
	}
}

func (t *Tui) setItems(items []*feed.Item) {
	t.SubWidget.Items = items
	itemTexts := []string{}
	for _, item := range items {
		itemTexts = append(itemTexts, item.Title)
	}
	t.LoadCells(t.SubWidget.Table, itemTexts)
	if t.SubWidget.Table.GetRowCount() != 0 {
		t.SubWidget.Table.Select(0, 0).ScrollToBeginning()
	}
}

func (t *Tui) deleteFeed(i int) {
	a := t.MainWidget.Feeds
	a = append(a[:i], a[i+1:]...)
}

func (t *Tui) GetTodaysFeeds() {
	const feedname = "Today's Items"
	for i, f := range t.MainWidget.Feeds {
		if f.Title == feedname {
			t.deleteFeed(i)
			break
		}
	}
	targetfeed := feed.MergeFeeds(t.MainWidget.Feeds, feedname)
	t.MainWidget.Feeds = append(t.MainWidget.Feeds, targetfeed)

	// 現在時刻より未来のフィードを除外
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	result := make([]*feed.Item, 0)
	for _, item := range targetfeed.Items {
		if today.Before(item.PubDate) {
			result = append(result, item)
		}
	}
	targetfeed.Items = result
	t.setFeeds(t.MainWidget.Feeds)
}

func (t *Tui) sortFeeds() {
	sort.Slice(t.MainWidget.Feeds, func(i, j int) bool {
		return strings.Compare(t.MainWidget.Feeds[i].Title, t.MainWidget.Feeds[j].Title) == -1
	})
	sort.Slice(t.MainWidget.Feeds, func(i, j int) bool {
		// Prioritize merged feeds
		return t.MainWidget.Feeds[i].Merged && !t.MainWidget.Feeds[j].Merged
	})
}

func (t *Tui) updateFeed(i int) {
	t.MainWidget.Feeds[i] = feed.GetFeedFromURL(t.MainWidget.Feeds[i].FeedLink, t.MainWidget.Feeds[i].Title)
	t.setItems(t.MainWidget.Feeds[i].Items)
}

func (t *Tui) updateSelectedFeed() {
	t.Notify("Updating...")
	t.App.ForceDraw()

	row, _ := t.MainWidget.Table.GetSelection()
	t.updateFeed(row)

	t.MainWidget.SaveFeeds()
	t.Notify("Updated.")
}

func (t *Tui) updateAllFeed() {
	t.Notify("Updating...")
	t.App.ForceDraw()

	for i, _ := range t.MainWidget.Feeds {
		t.updateFeed(i)
	}

	t.MainWidget.SaveFeeds()
	t.Notify("Updated.")
}

func (t *Tui) selectMainRow() {
	row, _ := t.MainWidget.Table.GetSelection()
	if len(t.MainWidget.Feeds) != 0 {
		feed := t.MainWidget.Feeds[row]
		t.setItems(feed.Items)
		t.Notify(fmt.Sprint(feed.Title, "\n", feed.Link))
		t.UpdateHelp("[l]:move to SubColumn [r]:reload selecting feed [R]:reload All feeds [q]:quit rfcui")
	}
}

func (t *Tui) selectSubRow() {
	row, _ := t.SubWidget.Table.GetSelection()
	if len(t.SubWidget.Items) != 0 {
		item := t.SubWidget.Items[row]
		t.Notify(fmt.Sprint(item.Belong, "\n", item.FormatTime(), "\n", item.Title, "\n", item.Link))
	}
}

func (t *Tui) setFeeds(feeds []*feed.Feed) {
	t.MainWidget.Feeds = feeds
	t.sortFeeds()
	feedTitles := []string{}
	for _, feed := range t.MainWidget.Feeds {
		feedTitles = append(feedTitles, feed.Title)
	}
	t.LoadCells(t.MainWidget.Table, feedTitles)
	row, _ := t.MainWidget.Table.GetSelection()
	max := t.MainWidget.Table.GetRowCount() - 1
	if max < row {
		t.MainWidget.Table.Select(max, 0).ScrollToBeginning()
	}
	t.App.ForceDraw()
}

type MainWidget struct {
	Table *tview.Table `json:"Table"`
	Feeds []*feed.Feed `json:"Feeds"`
}

func (m *MainWidget) GetFeedTitles() []string {
	titles := []string{}
	for _, feed := range m.Feeds {
		titles = append(titles, feed.Title)
	}
	return titles
}

func (m *MainWidget) SaveFeeds() error {
	for _, f := range m.Feeds {
		b, err := feed.Encode(f)
		if err != nil {
			return err
		}
		hash := fmt.Sprintf("%x", md5.Sum([]byte(f.FeedLink)))
		feed.SaveBytes(b, filepath.Join(datapath, hash))
	}
	return nil
}

func (m *MainWidget) LoadFeeds(path string) error {
	for _, file := range feed.DirWalk(path) {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		m.Feeds = append(m.Feeds, feed.Decode(b))
	}
	return nil
}

type SubWidget struct {
	Table *tview.Table `json:"SubTable"`
	Items []*feed.Item `json:"Items"`
}

func (s *SubWidget) GetItemTitles() []string {
	titles := []string{}
	for _, item := range s.Items {
		titles = append(titles, item.Title)
	}
	return titles
}

func NewTui() *Tui {

	mainTable := tview.NewTable()
	mainTable.SetTitle("Feeds").SetBorder(true).SetTitleAlign(tview.AlignLeft)
	mainTable.Select(0, 0).SetSelectable(true, true)

	subTable := tview.NewTable()
	subTable.SetTitle("Items").SetBorder(true).SetTitleAlign(tview.AlignLeft)
	subTable.Select(0, 0).SetSelectable(true, true)

	infoWidget := tview.NewTextView()
	infoWidget.SetTitle("Info").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	helpWidget := tview.NewTextView().SetTextAlign(1)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(mainTable, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(subTable, 0, 3, false).
				AddItem(infoWidget, 0, 1, false),
				0, 2, false),
			0, 1, false).AddItem(helpWidget, 1, 0, false)

	tui := &Tui{
		App:        tview.NewApplication(),
		Pages:      tview.NewPages().AddPage("MainPage", flex, true, true),
		MainWidget: &MainWidget{mainTable, []*feed.Feed{}},
		SubWidget:  &SubWidget{subTable, []*feed.Item{}},
		Info:       infoWidget,
		Help:       helpWidget,
	}

	return tui
}

func execCmd(attachStd bool, cmd string, args ...string) error {
	command := exec.Command(cmd, args...)

	if attachStd {
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}
	defer func() {
		command.Stdin = nil
		command.Stdout = nil
		command.Stderr = nil
	}()

	return command.Run()
}

func (t *Tui) Run() error {

	t.MainWidget.Table.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			t.MainWidget.Table.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {
		t.MainWidget.Table.GetCell(row, column).SetTextColor(tcell.ColorRed)
		t.MainWidget.Table.SetSelectable(false, false)
	}).SetSelectionChangedFunc(func(row, column int) {
		feed := t.MainWidget.Feeds[row]
		t.setItems(feed.Items)
		t.Notify(fmt.Sprint(feed.Title, "\n", feed.Link))
	})

	t.MainWidget.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 't':
				t.GetTodaysFeeds()
				t.UpdateHelp("gettting Today's Feed...")
				return nil
			case 'R':
				t.updateAllFeed()
				return nil
			case 'r':
				t.updateSelectedFeed()
				return nil
			}
		}
		return event
	})

	t.SubWidget.Table.SetSelectionChangedFunc(func(row, column int) {
		item := t.SubWidget.Items[row]
		t.Notify(fmt.Sprint(item.Belong, "\n", item.FormatTime(), "\n", item.Title, "\n", item.Link))
	}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				row, _ := t.SubWidget.Table.GetSelection()
				browser := os.Getenv("BROWSER")
				if browser == "" {
					t.Notify("$BROWSER is empty. Set it and try again.")
				} else {
					execCmd(true, browser, t.SubWidget.Items[row].Link)
				}
				return nil
			case tcell.KeyRune:
				switch event.Rune() {
				case 'o':
					row, _ := t.SubWidget.Table.GetSelection()
					browser := os.Getenv("BROWSER")
					if browser == "" {
						t.Notify("$BROWSER is empty. Set it and try again.")
					} else {
						execCmd(true, browser, t.SubWidget.Items[row].Link)
					}
					return nil
				}
			}
			return event
		})

	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			t.App.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'h':
				t.App.SetFocus(t.MainWidget.Table)
				t.RefreshTui()
				return nil
			case 'l':
				t.App.SetFocus(t.SubWidget.Table)
				t.RefreshTui()
				return nil
			case 'q':
				t.App.Stop()
				return nil
			}
		}
		return event
	})

	if !feed.IsDir(datapath) {
		os.MkdirAll(datapath, 0755)
	}

	err := t.MainWidget.LoadFeeds(datapath)
	if err != nil {
		return err
	}

	feedURLs, err := getLines("list.txt")
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, url := range feedURLs {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			_ = t.AddFeedFromURL(u)
		}(url)
	}

	if len(t.MainWidget.Feeds) > 0 {
		t.SubWidget.Items = t.MainWidget.Feeds[0].Items
	}
	t.LoadCells(t.MainWidget.Table, t.MainWidget.GetFeedTitles())
	t.LoadCells(t.SubWidget.Table, t.SubWidget.GetItemTitles())

	t.App.SetRoot(t.Pages, true).SetFocus(t.MainWidget.Table)
	t.RefreshTui()

	wg.Wait()

	err = t.MainWidget.SaveFeeds()
	if err != nil {
		return err
	}

	if err := t.App.Run(); err != nil {
		t.App.Stop()
		return err
	}

	return nil
}
