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
	inputField      = "InputPopup"
	descriptionPage = "descriptionPage"
	mainPage        = "MainPage"
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
	WaitGroup   *sync.WaitGroup
}

func (tui *Tui) AddFeedFromURL(url string) error {
	f, err := feed.GetFeedFromURL(url, "")
	if err != nil {
		return err
	}

	for i, feed := range tui.MainWidget.Feeds {
		if feed.FeedLink == url {
			tui.MainWidget.Feeds[i] = f
			tui.setFeeds(tui.MainWidget.Feeds)
			return nil
		}
	}
	tui.setFeeds(append(tui.MainWidget.Feeds, f))
	return nil

}

func (tui *Tui) LoadCells(table *tview.Table, texts []string) {
	table.Clear()
	for i, text := range texts {
		table.SetCell(i, 0, tview.NewTableCell(text))
	}
}

func getDataPath() string {
	const datapath = "feedcache"
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, datapath)
}

func (tui *Tui) showDescription(text string) {
	tui.Description.SetText(text)
}

func (tui *Tui) Notify(text string) {
	tui.Info.SetText(text)
}

func (tui *Tui) UpdateHelp(text string) {
	tui.Help.SetText(text)
}

func (tui *Tui) RefreshTui() {
	focus := tui.App.GetFocus()
	if focus == tui.MainWidget.Table {
		row, column := tui.MainWidget.Table.GetSelection()
		tui.selectMainRow(row, column)
	} else if focus == tui.SubWidget.Table {
		row, column := tui.SubWidget.Table.GetSelection()
		tui.selectSubRow(row, column)
	}
}

func (tui *Tui) setItems(paintColor bool) {
	row, _ := tui.MainWidget.Table.GetSelection()
	items := tui.MainWidget.Feeds[row].Items

	tui.SubWidget.Items = items

	table := tui.SubWidget.Table.Clear()
	for i, item := range items {
		table.SetCellSimple(i, 0, item.Title)
		if paintColor {
			table.GetCell(i, 0).SetTextColor(tcellColors[item.Color])
		}
	}

	if tui.SubWidget.Table.GetRowCount() != 0 {
		tui.SubWidget.Table.Select(0, 0).ScrollToBeginning()
	}
}

func (tui *Tui) deleteFeed(i int) {
	a := tui.MainWidget.Feeds
	a = append(a[:i], a[i+1:]...)
}

func (tui *Tui) GetTodaysFeeds() {
	const feedname = "Today's Items"

	targetfeed := feed.MergeFeeds(tui.MainWidget.Feeds, feedname)

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

	isExist := false
	for i, f := range tui.MainWidget.Feeds {
		if f.Title == feedname {
			tui.MainWidget.Feeds[i] = targetfeed
			isExist = true
			break
		}
	}
	if !isExist {
		tui.MainWidget.Feeds = append(tui.MainWidget.Feeds, targetfeed)
	}
	tui.setFeeds(tui.MainWidget.Feeds)
}

func (tui *Tui) GetAllItems() {
	const feedname = "All Items"

	targetfeed := feed.MergeFeeds(tui.MainWidget.Feeds, feedname)

	isExist := false
	for i, f := range tui.MainWidget.Feeds {
		if f.Title == feedname {
			tui.MainWidget.Feeds[i] = targetfeed
			isExist = true
			break
		}
	}
	if !isExist {
		tui.MainWidget.Feeds = append(tui.MainWidget.Feeds, targetfeed)
	}
	tui.setFeeds(tui.MainWidget.Feeds)
}

func (tui *Tui) sortFeeds() {
	sort.Slice(tui.MainWidget.Feeds, func(i, j int) bool {
		return strings.Compare(tui.MainWidget.Feeds[i].Title, tui.MainWidget.Feeds[j].Title) == -1
	})
	sort.Slice(tui.MainWidget.Feeds, func(i, j int) bool {
		// Prioritize merged feeds
		return tui.MainWidget.Feeds[i].Merged && !tui.MainWidget.Feeds[j].Merged
	})
}

func (tui *Tui) updateFeed(i int) error {
	if tui.MainWidget.Feeds[i].Merged {
		//return errors.New("merged feed can't update")
		return nil
	}

	var err error
	tui.MainWidget.Feeds[i], err = feed.GetFeedFromURL(tui.MainWidget.Feeds[i].FeedLink, tui.MainWidget.Feeds[i].Title)
	if err != nil {
		return err
	}

	return nil
}

func (tui *Tui) updateSelectedFeed() error {
	tui.Notify("Updating...")
	tui.App.ForceDraw()

	row, _ := tui.MainWidget.Table.GetSelection()
	if err := tui.updateFeed(row); err != nil {
		return err
	}

	tui.MainWidget.SaveFeeds()
	tui.setItems(tui.MainWidget.Feeds[row].Merged)
	tui.GetTodaysFeeds()
	tui.GetAllItems()
	tui.Notify("Updated.")
	tui.App.SetFocus(tui.MainWidget.Table)

	return nil
}

func (tui *Tui) updateAllFeed() error {
	tui.Notify("Updating...")
	tui.App.ForceDraw()

	length := len(tui.MainWidget.Feeds)
	doneCount := 0

	wg := sync.WaitGroup{}

	for index := range tui.MainWidget.Feeds {
		wg.Add(1)
		go func(i int) {
			tui.updateFeed(i)
			doneCount++
			tui.Notify(fmt.Sprint("Updating ", doneCount, "/", length, " feeds..."))
			tui.App.ForceDraw()
			wg.Done()
		}(index)
	}
	wg.Wait()

	tui.GetTodaysFeeds()
	tui.GetAllItems()
	tui.MainWidget.SaveFeeds()
	tui.Notify("All feeds have updated.")

	return nil
}

func (tui *Tui) selectMainRow(row, column int) {
	feed := tui.MainWidget.Feeds[row]
	tui.setItems(tui.MainWidget.Feeds[row].Merged)
	if tui.App.GetFocus() == tui.MainWidget.Table {
		tui.showDescription(fmt.Sprint(feed.Title, "\n", feed.Link))
		tui.UpdateHelp("[l]:move to SubColumn [r]:reload selecting feed [R]:reload All feeds [q]:quit rfcui")
	}
}

func (tui *Tui) selectSubRow(row, column int) {
	item := tui.SubWidget.Items[row]
	if tui.App.GetFocus() == tui.SubWidget.Table {
		tui.showDescription(fmt.Sprint(item.Belong, "\n", item.FormatTime(), "\n", item.Title, "\n", item.Link))
		tui.UpdateHelp("[h]:move to MainColumn [o]:open an item with $BROWSER [q]:quit rfcui")
	}
}

func (tui *Tui) setFeeds(feeds []*feed.Feed) {
	tui.MainWidget.Feeds = feeds
	tui.sortFeeds()
	table := tui.MainWidget.Table.Clear()
	for i, feed := range tui.MainWidget.Feeds {
		table.SetCellSimple(i, 0, feed.Title)
		if !feed.Merged {
			table.GetCell(i, 0).SetTextColor(tcellColors[feed.Color])
		}
	}
	row, _ := tui.MainWidget.Table.GetSelection()
	max := tui.MainWidget.Table.GetRowCount() - 1
	if max < row {
		tui.MainWidget.Table.Select(max, 0).ScrollToBeginning()
	}
	tui.App.ForceDraw()
}

func (tui *Tui) AddFeedsFromURL(path string) error {
	_, feedURLs, err := myio.GetLines(path)
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

	wg := sync.WaitGroup{}

	for _, url := range newURLs {
		wg.Add(1)
		go func(u string) {
			_ = tui.AddFeedFromURL(u)
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

	descriptionFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(mainTable, 0, 4, false).
				AddItem(infoWidget, 0, 1, false),
				0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(subTable, 0, 2, false).
				AddItem(descriptionWidget, 0, 3, false),
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
		AddPage(mainPage, mainFlex, true, true).
		AddPage(descriptionPage, descriptionFlex, true, false).
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
		WaitGroup:   &sync.WaitGroup{},
	}

	tui.setAppFunctions()

	return tui
}

func (tui *Tui) setAppFunctions() {
	tui.MainWidget.Table.SetSelectionChangedFunc(func(row, column int) {
		tui.selectMainRow(row, column)
	})

	tui.MainWidget.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'R':
				tui.updateAllFeed()
				return nil
			case 'r':
				tui.updateSelectedFeed()
				return nil
			case 'l':
				tui.App.SetFocus(tui.SubWidget.Table)
				tui.RefreshTui()
				return nil
			}
		}
		return event
	})

	tui.SubWidget.Table.SetSelectionChangedFunc(func(row, column int) {
		tui.selectSubRow(row, column)
	}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				row, _ := tui.SubWidget.Table.GetSelection()
				browser := os.Getenv("BROWSER")
				if browser == "" {
					tui.showDescription("$BROWSER is empty. Set it and try again.")
				} else {
					execCmd(true, browser, tui.SubWidget.Items[row].Link)
				}
				return nil
			case tcell.KeyRune:
				switch event.Rune() {
				case 'h':
					tui.App.SetFocus(tui.MainWidget.Table)
					tui.RefreshTui()
					return nil
				case 'l':
					tui.Pages.SwitchToPage(descriptionPage)
					tui.App.SetFocus(tui.Description)
					return nil
				case 'o':
					row, _ := tui.SubWidget.Table.GetSelection()
					browser := os.Getenv("BROWSER")
					if browser == "" {
						tui.showDescription("$BROWSER is empty. Set it and try again.")
					} else {
						execCmd(true, browser, tui.SubWidget.Items[row].Link)
					}
					return nil
				}
			}
			return event
		})

	tui.Description.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'h':
				tui.Pages.ShowPage(mainPage)
				tui.Pages.HidePage(descriptionPage)
				tui.App.SetFocus(tui.SubWidget.Table)
				return nil
			}
		}
		return event
	})

	tui.InputWidget.Input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			//
			switch tui.InputWidget.Mode {
			case 0: // new feed
				if err := tui.AddFeedFromURL(tui.InputWidget.Input.GetText()); err != nil {
					tui.Notify(err.Error())
				}
			}
			tui.InputWidget.Input.SetText("")
			tui.InputWidget.Input.SetTitle("Input")
			tui.Pages.HidePage(inputField)
			tui.App.SetFocus(tui.MainWidget.Table)
			return nil
		}
		return event
	})

	tui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if tui.App.GetFocus() == tui.InputWidget.Input {
			return event
		}
		switch event.Key() {
		case tcell.KeyEscape:
			tui.App.Stop()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'n':
				tui.InputWidget.Input.SetTitle("New Feed")
				tui.InputWidget.Mode = 0
				tui.Pages.ShowPage(inputField)
				tui.App.SetFocus(tui.InputWidget.Input)
				return nil
			case 'q':
				tui.App.Stop()
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

func (tui *Tui) Run() error {

	err := tui.MainWidget.LoadFeeds(getDataPath())
	if err != nil {
		return err
	}

	if err := tui.AddFeedsFromURL("list.txt"); err != nil {
		return err
	}

	//tui.GetTodaysFeeds()
	//tui.GetAllItems()

	err = tui.MainWidget.SaveFeeds()
	if err != nil {
		return err
	}

	if len(tui.MainWidget.Feeds) > 0 {
		tui.setFeeds(tui.MainWidget.Feeds)
		tui.setItems(tui.MainWidget.Feeds[0].Merged)
	}
	tui.App.SetRoot(tui.Pages, true).SetFocus(tui.MainWidget.Table)
	tui.RefreshTui()

	tui.WaitGroup.Add(1)
	go func() {
		if err := tui.updateAllFeed(); err != nil {
			panic(err)
		}
		tui.WaitGroup.Done()
	}()

	if err := tui.App.Run(); err != nil {
		tui.WaitGroup.Wait()
		tui.App.Stop()
		return err
	}

	return nil
}
