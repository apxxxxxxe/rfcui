package tui

import (
	"crypto/md5"
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
	myio "github.com/apxxxxxxe/rfcui/io"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	datapath   = "feedcache"
	inputField = "InputPopup"
)

type Tui struct {
	App         *tview.Application
	Pages       *tview.Pages
	MainWidget  *MainWidget
	SubWidget   *SubWidget
	Description *tview.TextView
	Info        *tview.TextView
	Help        *tview.TextView
	InputWidget *InputBox
}

func (t *Tui) AddFeedFromURL(url string) error {
	f, err := feed.GetFeedFromURL(url, "")
	if err != nil {
		return err
	}

	for i, feed := range t.MainWidget.Feeds {
		if feed.FeedLink == url {
			t.MainWidget.Feeds[i] = f
			t.setFeeds(t.MainWidget.Feeds)
			return nil
		}
	}
	t.setFeeds(append(t.MainWidget.Feeds, f))
	return nil

}

func (t *Tui) LoadCells(table *tview.Table, texts []string) {
	table.Clear()
	for i, text := range texts {
		table.SetCell(i, 0, tview.NewTableCell(text))
	}
}

func getDataPath() string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, datapath)
}

func (t *Tui) showDescription(text string) {
	t.Description.SetText(text)
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

func (t *Tui) updateFeed(i int) error {
	var err error

	t.MainWidget.Feeds[i], err = feed.GetFeedFromURL(t.MainWidget.Feeds[i].FeedLink, t.MainWidget.Feeds[i].Title)
	if err != nil {
		return err
	}

	t.setItems(t.MainWidget.Feeds[i].Items)
	return nil
}

func (t *Tui) updateSelectedFeed() error {
	t.showDescription("Updating...")
	t.App.ForceDraw()

	row, _ := t.MainWidget.Table.GetSelection()
	if err := t.updateFeed(row); err != nil {
		return err
	}

	t.MainWidget.SaveFeeds()
	t.showDescription("Updated.")

	return nil
}

func (t *Tui) updateAllFeed() error {
	t.Notify("Updating...")
	t.App.ForceDraw()

	length := len(t.MainWidget.Feeds)
	doneCount := 0

	wg := sync.WaitGroup{}
	for index := range t.MainWidget.Feeds {
		wg.Add(1)
		go func(i int) {
			t.updateFeed(i)
			doneCount++
			t.Notify(fmt.Sprint("Updating ", doneCount, "/", length, " feeds..."))
			t.App.ForceDraw()
			wg.Done()
		}(index)
	}
	wg.Wait()

	t.MainWidget.SaveFeeds()
	t.Notify("All feeds have updated.")

	return nil
}

func (t *Tui) selectMainRow() {
	row, _ := t.MainWidget.Table.GetSelection()
	if len(t.MainWidget.Feeds) != 0 {
		feed := t.MainWidget.Feeds[row]
		t.setItems(feed.Items)
		t.showDescription(fmt.Sprint(feed.Title, "\n", feed.Link))
		t.UpdateHelp("[l]:move to SubColumn [r]:reload selecting feed [R]:reload All feeds [q]:quit rfcui")
	}
}

func (t *Tui) selectSubRow() {
	row, _ := t.SubWidget.Table.GetSelection()
	if len(t.SubWidget.Items) != 0 {
		item := t.SubWidget.Items[row]
		t.showDescription(fmt.Sprint(item.Belong, "\n", item.FormatTime(), "\n", item.Title, "\n", item.Link))
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

func (t *Tui) AddFeedsFromURL(path string) error {
	_, feedURLs, err := getLines(path)
	if err != nil {
		return err
	}

	fileNames := []string{}
	for _, fp := range myio.DirWalk(getDataPath()) {
		fileNames = append(fileNames, filepath.Base(fp))
	}

	newURLs := []string{}
	for _, feedLink := range feedURLs {
		isNewURL := true
		hash := fmt.Sprintf("%x", md5.Sum([]byte(feedLink)))
		for _, fileName := range fileNames {
			if filepath.Base(fileName) == hash {
				isNewURL = false
			}
		}
		if isNewURL {
			newURLs = append(newURLs, feedLink)
		}
	}

	//ch := make(chan string, count)
	//go func() {
	//	for _, url := range feedURLs {
	//		ch <- url
	//	}
	//	close(ch)
	//}()
	//for i := 0; i < count; i++ {
	//	for url := range ch {
	//		if err := t.AddFeedFromURL(url); err != nil {
	//			panic(err)
	//		}
	//	}
	//}

	wg := sync.WaitGroup{}

	for _, url := range newURLs {
		wg.Add(1)
		go func(u string) {
			_ = t.AddFeedFromURL(u)
			wg.Done()
		}(url)
	}

	wg.Wait()

	return nil
}

type MainWidget struct {
	Table *tview.Table `json:"Table"`
	Feeds []*feed.Feed `json:"Feeds"`
}

func (m *MainWidget) SaveFeeds() error {
	for _, f := range m.Feeds {
		if f.Merged {
			continue
		}

		b, err := feed.Encode(f)
		if err != nil {
			return err
		}
		hash := fmt.Sprintf("%x", md5.Sum([]byte(f.FeedLink)))
		myio.SaveBytes(b, filepath.Join(getDataPath(), hash))
	}
	return nil
}

func (m *MainWidget) LoadFeeds(path string) error {
	if !myio.IsDir(getDataPath()) {
		os.MkdirAll(getDataPath(), 0755)
	}
	for _, file := range myio.DirWalk(path) {
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

type InputBox struct {
	Input *tview.InputField
	Mode  int
}

func NewTui() *Tui {

	mainTable := tview.NewTable()
	mainTable.SetTitle("Feeds").SetBorder(true).SetTitleAlign(tview.AlignLeft)
	mainTable.Select(0, 0).SetSelectable(true, true)

	subTable := tview.NewTable()
	subTable.SetTitle("Items").SetBorder(true).SetTitleAlign(tview.AlignLeft)
	subTable.Select(0, 0).SetSelectable(true, true)

	descriptionWidget := tview.NewTextView()
	descriptionWidget.SetTitle("Description").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	infoWidget := tview.NewTextView()
	infoWidget.SetTitle("Info").SetBorder(true).SetTitleAlign(tview.AlignLeft)

	helpWidget := tview.NewTextView().SetTextAlign(1)

	inputWidget := tview.NewInputField()
	inputWidget.SetBorder(true).SetTitleAlign(tview.AlignLeft)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(mainTable, 0, 4, false).
				AddItem(infoWidget, 0, 1, false),
				0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(subTable, 0, 3, false).
				AddItem(descriptionWidget, 0, 1, false),
				0, 2, false),
			0, 1, false).AddItem(helpWidget, 1, 0, false)

	inputFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inputWidget, 3, 1, false).
			AddItem(nil, 0, 1, false), 40, 1, false).
		AddItem(nil, 0, 1, false)

	pages := tview.NewPages().
		AddPage("MainPage", mainFlex, true, true).
		AddPage(inputField, inputFlex, true, false)

	tui := &Tui{
		App:         tview.NewApplication(),
		Pages:       pages,
		MainWidget:  &MainWidget{mainTable, []*feed.Feed{}},
		SubWidget:   &SubWidget{subTable, []*feed.Item{}},
		Description: descriptionWidget,
		Info:        infoWidget,
		Help:        helpWidget,
		InputWidget: &InputBox{inputWidget, 0},
	}

	tui.setAppFunctions()

	return tui
}

func (t *Tui) setAppFunctions() {
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
		t.showDescription(fmt.Sprint(feed.Title, "\n", feed.Link))
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
		t.showDescription(fmt.Sprint(item.Belong, "\n", item.FormatTime(), "\n", item.Title, "\n", item.Link))
	}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				row, _ := t.SubWidget.Table.GetSelection()
				browser := os.Getenv("BROWSER")
				if browser == "" {
					t.showDescription("$BROWSER is empty. Set it and try again.")
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
						t.showDescription("$BROWSER is empty. Set it and try again.")
					} else {
						execCmd(true, browser, t.SubWidget.Items[row].Link)
					}
					return nil
				}
			}
			return event
		})

	t.InputWidget.Input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			//
			switch t.InputWidget.Mode {
			case 0: // new feed
				if err := t.AddFeedFromURL(t.InputWidget.Input.GetText()); err != nil {
					t.Notify(err.Error())
				}
			}
			t.InputWidget.Input.SetText("")
			t.InputWidget.Input.SetTitle("Input")
			t.Pages.HidePage(inputField)
			t.App.SetFocus(t.MainWidget.Table)
			return nil
		}
		return event
	})

	t.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if t.App.GetFocus() == t.InputWidget.Input {
			return event
		}
		switch event.Key() {
		case tcell.KeyEscape:
			t.App.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'n':
				t.InputWidget.Input.SetTitle("New Feed")
				t.InputWidget.Mode = 0
				t.Pages.ShowPage(inputField)
				t.App.SetFocus(t.InputWidget.Input)
				return nil
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

	err := t.MainWidget.LoadFeeds(getDataPath())
	if err != nil {
		return err
	}

	if err := t.AddFeedsFromURL("list.txt"); err != nil {
		return err
	}

	err = t.MainWidget.SaveFeeds()
	if err != nil {
		return err
	}

	if len(t.MainWidget.Feeds) > 0 {
		t.setFeeds(t.MainWidget.Feeds)
		t.setItems(t.MainWidget.Feeds[0].Items)
	}
	t.App.SetRoot(t.Pages, true).SetFocus(t.MainWidget.Table)
	t.RefreshTui()

	if err := t.App.Run(); err != nil {
		t.App.Stop()
		return err
	}

	return nil
}
