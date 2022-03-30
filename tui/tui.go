package tui

import (
	"fmt"
	"os"

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
	Help       *tview.TextView
	FocusIndex int
}

func (t *Tui) RefreshTui() {
	if t.MainWidget.Table.HasFocus() {
		t.SelectMainWidgetRow(0)
	}
	if t.SubWidget.Table.HasFocus() {
		t.SelectSubWidgetRow(0)
	}
}

func (t *Tui) Notify(text string) {
	t.Info.SetText(text)
}

func (t *Tui) UpdateHelp(text string) {
	t.Help.SetText(text)
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
	t.MainWidget.Table.Select(0, 0).ScrollToBeginning()
}

func (t *Tui) SetArticles(items []*feed.Article) {
	t.SubWidget.Items = items
	itemTexts := []string{}
	for _, item := range items {
		itemTexts = append(itemTexts, item.Title)
	}
	t.LoadCells(t.SubWidget.Table, itemTexts)
	t.SubWidget.Table.Select(0, 0).ScrollToBeginning()
}

func (t *Tui) UpdateSelectedFeed() {
	row, _ := t.MainWidget.Table.GetSelection()
	targetFeed := *t.MainWidget.Feeds[row]
	targetFeed = *feed.GetFeedFromUrl(targetFeed.FeedLink)
	t.SetArticles(targetFeed.Items)
	t.Notify("Updated.")
}

func (m *MainWidget) GetFeedTitles() []string {
	titles := []string{}
	for _, feed := range m.Feeds {
		titles = append(titles, feed.Title)
	}
	return titles
}

func (t *Tui) SelectMainWidgetRow(count int) {
	row, column := t.MainWidget.Table.GetSelection()
	if (count < 0 && row+count >= 0) || (count > 0 && row+count <= t.MainWidget.Table.GetRowCount()-1) {
		row += count
	}
	t.MainWidget.Table.Select(row, column)
	feed := t.MainWidget.Feeds[row]
	t.SetArticles(feed.Items)
	t.Notify(fmt.Sprint(feed.Title, "\n", feed.Link, "\n", feed.FeedLink))
}

func (t *Tui) SelectSubWidgetRow(count int) {
	row, column := t.SubWidget.Table.GetSelection()
	if (count < 0 && row+count >= 0) || (count > 0 && row+count <= t.SubWidget.Table.GetRowCount()-1) {
		row += count
	}
	t.SubWidget.Table.Select(row, column)
	item := t.SubWidget.Items[row]
	t.Notify(fmt.Sprint(item.Belong.Title, "\n", item.FormatTime(), "\n", item.Title, "\n", item.Link))
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

	infoWidget := tview.NewTextView()
	infoWidget.SetTitle("Info").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	helpWidget := tview.NewTextView().SetTextAlign(2)

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
		FocusIndex: 0,
		MainWidget: &MainWidget{mainTable, []*feed.Feed{}},
		SubWidget:  &SubWidget{subTable, []*feed.Article{}},
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
	})

	t.MainWidget.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'r':
				t.UpdateSelectedFeed()
				return nil
			case 'j':
				t.SelectMainWidgetRow(1)
				return nil
			case 'k':
				t.SelectMainWidgetRow(-1)
				return nil
			}
		}
		return event
	})

	t.SubWidget.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
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
			case 'j':
				t.SelectSubWidgetRow(1)
				return nil
			case 'k':
				t.SelectSubWidgetRow(-1)
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
				t.Notify("")
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

	t.SubWidget.Items = t.MainWidget.Feeds[0].Items
	t.LoadCells(t.MainWidget.Table, t.MainWidget.GetFeedTitles())
	t.LoadCells(t.SubWidget.Table, t.SubWidget.GetArticleTitles())

	t.App.SetRoot(t.Pages, true).SetFocus(t.MainWidget.Table)
	t.RefreshTui()

	if err := t.App.Run(); err != nil {
		t.App.Stop()
		return err
	}
	return nil
}
