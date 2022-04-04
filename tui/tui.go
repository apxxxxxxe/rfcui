package tui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/apxxxxxxe/rfcui/feed"
	"github.com/apxxxxxxe/rfcui/io"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
	f := feed.GetFeedFromUrl(url, "")
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

func (t *Tui) updateSelectedFeed() {
	t.Notify("Updating...")
	t.App.ForceDraw()
	row, _ := t.MainWidget.Table.GetSelection()
	targetFeed := *t.MainWidget.Feeds[row]
	targetFeed = *feed.GetFeedFromUrl(targetFeed.FeedLink, targetFeed.Title)
	t.setItems(targetFeed.Items)
	t.Notify("Updated.")
}

func (t *Tui) updateAllFeed() {
	t.Notify("Updating...")
	t.App.ForceDraw()
	for _, f := range t.MainWidget.Feeds {
		f = feed.GetFeedFromUrl(f.FeedLink, f.Title)
		t.setItems(f.Items)
	}
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
		t.Notify(fmt.Sprint(item.Belong.Title, "\n", item.FormatTime(), "\n", item.Title, "\n", item.Link))
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
  Table *tview.Table
  Feeds []*feed.Feed
}

func (m *MainWidget) GetFeedTitles() []string {
  titles := []string{}
  for _, feed := range m.Feeds {
    titles = append(titles, feed.Title)
  }
  return titles
}

type SubWidget struct {
  Table *tview.Table
  Items []*feed.Item
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
		t.Notify(fmt.Sprint(item.Belong.Title, "\n", item.FormatTime(), "\n", item.Title, "\n", item.Link))
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

	feedURLs, err := io.GetLines("list.txt")
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, url := range feedURLs {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if err := t.AddFeedFromURL(u); err != nil {
				log.Fatal(err)
			}
		}(url)
	}

	if len(t.MainWidget.Feeds) > 0 {
		t.SubWidget.Items = t.MainWidget.Feeds[0].Items
	}
	t.LoadCells(t.MainWidget.Table, t.MainWidget.GetFeedTitles())
	t.LoadCells(t.SubWidget.Table, t.SubWidget.GetItemTitles())

	t.App.SetRoot(t.Pages, true).SetFocus(t.MainWidget.Table)
	t.RefreshTui()

	if err := t.App.Run(); err != nil {
		t.App.Stop()
		return err
	}

	wg.Wait()

	return nil
}


