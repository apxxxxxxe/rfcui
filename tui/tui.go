package tui

import (
	"fmt"

	"github.com/apxxxxxxe/rfcui/feed"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Tui struct {
	App        *tview.Application
	Pages      *tview.Pages
	MainWidget *MainWidget
	SubWidget  *SubWidget
	Info       *tview.TextView
	FocusIndex int
}

func (t *Tui) Notify(text string) {
	t.Info.SetText(text)
}

func (t *Tui) LoadCells(table *tview.Table, texts []string) {
	table.Clear()
	for i, text := range texts {
		table.SetCell(i, 0, tview.NewTableCell(text))
	}
}

type MainWidget struct {
	Table *tview.Table
	Feeds []*feed.Feed
}

type SubWidget struct {
	Table *tview.Table
	Items []*feed.Article
}

func (t *Tui) SetFeeds(feeds []*feed.Feed) {
	t.MainWidget.Feeds = feeds
	feedTitles := []string{}
	for _, feed := range feeds {
		feedTitles = append(feedTitles, feed.Title)
	}
	t.LoadCells(t.MainWidget.Table, feedTitles)
}

func (t *Tui) SetArticles(items []*feed.Article) {
	t.SubWidget.Items = items
	itemTexts := []string{}
	for _, item := range items {
		itemTexts = append(itemTexts, item.Title)
	}
	t.LoadCells(t.SubWidget.Table, itemTexts)
}

func (m *MainWidget) GetFeedTitles() []string {
	titles := []string{}
	for _, feed := range m.Feeds {
		titles = append(titles, feed.Title)
	}
	return titles
}

func (s *SubWidget) GetArticleTitles() []string {
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
	subTable.SetTitle("Articles").SetBorder(true).SetTitleAlign(tview.AlignLeft)
	subTable.Select(0, 0).SetSelectable(true, true)

	infoWidget := tview.NewTextView().SetTextAlign(1)
	infoWidget.SetTitle("Details").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	grid := tview.NewGrid()
	grid.SetSize(6, 5, 0, 0).
		AddItem(mainTable, 0, 0, 6, 2, 0, 0, true).
		AddItem(subTable, 0, 2, 4, 4, 0, 0, true).
		AddItem(infoWidget, 4, 2, 2, 4, 0, 0, true)

	tui := &Tui{
		App:        tview.NewApplication(),
		Pages:      tview.NewPages().AddPage("MainPage", grid, true, true),
		FocusIndex: 0,
		MainWidget: &MainWidget{mainTable, []*feed.Feed{}},
		SubWidget:  &SubWidget{subTable, []*feed.Article{}},
		Info:       infoWidget,
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
	})

	t.MainWidget.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				row, column := t.MainWidget.Table.GetSelection()
				if row < t.MainWidget.Table.GetRowCount() {
					t.MainWidget.Table.Select(row+1, column)
				}
				row, _ = t.MainWidget.Table.GetSelection()
				t.SetArticles(t.MainWidget.Feeds[row].Items)
				t.Notify(string(row))
				return nil
			case 'k':
				row, column := t.MainWidget.Table.GetSelection()
				if row > 0 {
					t.MainWidget.Table.Select(row-1, column)
				}
				row, _ = t.MainWidget.Table.GetSelection()
				t.SetArticles(t.MainWidget.Feeds[row].Items)
				t.Notify(string(row))
				return nil
			}
		}
		return event
	})

	t.SubWidget.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				row, column := t.SubWidget.Table.GetSelection()
				if row < t.SubWidget.Table.GetRowCount() {
					t.SubWidget.Table.Select(row+1, column)
				}
				row, _ = t.SubWidget.Table.GetSelection()
				item := t.SubWidget.Items[row]
				t.Notify(fmt.Sprint(item.Belong.Title, "\n", item.PubDate, "\n", item.Title, "\n", item.Link))
				return nil
			case 'k':
				row, column := t.SubWidget.Table.GetSelection()
				if row > 0 {
					t.SubWidget.Table.Select(row-1, column)
				}
				row, _ = t.SubWidget.Table.GetSelection()
				item := t.SubWidget.Items[row]
				t.Notify(fmt.Sprint(item.Belong.Title, "\n", item.PubDate, "\n", item.Title, "\n", item.Link))
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
				return nil
			case 'l':
				t.App.SetFocus(t.SubWidget.Table)
				return nil
			case 'q':
				t.App.Stop()
				return nil
			}
		}
		return event
	})

	t.SubWidget.Items = t.MainWidget.Feeds[0].Items
	t.LoadCells(t.MainWidget.Table, t.MainWidget.GetFeedTitles())
	t.LoadCells(t.SubWidget.Table, t.SubWidget.GetArticleTitles())

	if err := t.App.SetRoot(t.Pages, true).SetFocus(t.MainWidget.Table).Run(); err != nil {
		t.App.Stop()
		return err
	}
	return nil
}
