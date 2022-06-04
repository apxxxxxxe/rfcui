package tui

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	mycolor "github.com/apxxxxxxe/rfcui/color"
	fd "github.com/apxxxxxxe/rfcui/feed"
	myio "github.com/apxxxxxxe/rfcui/io"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	inputField      = "InputPopup"
	descriptionPage = "descriptionPage"
	mainPage        = "MainPage"
	modalPage       = "modalPage"
)

var (
	ErrGettingFeedFailed = errors.New("failed to get feed")
	ErrRmFailed          = errors.New("faled to remove files or dirs")
	cachePath            = filepath.Join(getDataPath(), "cache")
)

type Tui struct {
	App                *tview.Application
	Pages              *tview.Pages
	GroupWidget        *GroupWidget
	FeedWidget         *FeedWidget
	SubWidget          *SubWidget
	Description        *tview.TextView
	Info               *tview.TextView
	Help               *tview.TextView
	InputWidget        *InputBox
	WaitGroup          *sync.WaitGroup
	SelectingFeeds     []*fd.Feed
	ConfirmationStatus int
	LastSelectedWidget tview.Primitive
	Modal              *tview.Modal
}

func (tui *Tui) SelectFeed() {
	const defaultColor = tcell.ColorBlack
	const selectedColor = tcell.ColorWhite

	row, column := tui.FeedWidget.Table.GetSelection()
	if tui.FeedWidget.Table.GetCell(row, column).BackgroundColor == defaultColor {
		tui.FeedWidget.Table.GetCell(row, column).SetBackgroundColor(selectedColor)
		tui.SelectingFeeds = append(tui.SelectingFeeds, tui.FeedWidget.Feeds[row])
	} else {
		tui.FeedWidget.Table.GetCell(tui.FeedWidget.Table.GetSelection()).SetBackgroundColor(defaultColor)
		targetFeed := tui.FeedWidget.Feeds[row]
		for i, f := range tui.SelectingFeeds {
			if f == targetFeed {
				tui.SelectingFeeds = append(tui.SelectingFeeds[:i], tui.SelectingFeeds[i+1:]...)
				break
			}
		}
	}
}

func (tui *Tui) updateFeed(index int) error {
	targetFeed := tui.FeedWidget.Feeds[index]

	if targetFeed.IsMerged() {
		tui.FeedWidget.Feeds[index].Items = []*fd.Item{}
		for _, url := range targetFeed.FeedLinks {
			for _, f := range tui.FeedWidget.Feeds {
				if !f.IsMerged() {
					feedLink, _ := f.GetFeedLink()
					if url == feedLink {
						tui.FeedWidget.Feeds[index].Items = append(tui.FeedWidget.Feeds[index].Items, f.Items...)
						break
					}
				}
			}
		}
		tui.FeedWidget.Feeds[index].SortItems()

	} else {
		color := targetFeed.Color
		url, err := targetFeed.GetFeedLink()
		if err != nil {
			return fmt.Errorf(targetFeed.Title, ": ", err)
		}

		feed, err := fd.GetFeedFromURL(url, "")

		if err != nil {
			feed = getInvalidFeed(url, err)
		}

		if color > 0 && color < len(mycolor.TcellColors) {
			for _, item := range feed.Items {
				item.Color = targetFeed.Color
			}
		} else {
			targetFeed.Title = feed.Title
			targetFeed.Color = feed.Color
		}
		feed.SortItems()

		targetFeed.Link = feed.Link
		targetFeed.Description = feed.Description
		targetFeed.Items = feed.Items

		if err != nil {
			return ErrGettingFeedFailed
		}
	}
	return nil
}

func (tui *Tui) AddFeedFromURL(url string) error {
	f, err := fd.GetFeedFromURL(url, "")
	if err != nil {
		return err
	}

	if f.IsMerged() {
		if err := tui.GroupWidget.SaveGroup(f); err != nil {
			return err
		}
	} else {
		if err := tui.FeedWidget.SaveFeed(f); err != nil {
			return err
		}
	}

	for i, feed := range tui.FeedWidget.Feeds {
		if f.IsMerged() {
			if f.Title == feed.Title {
				tui.FeedWidget.Feeds[i] = f
				tui.FeedWidget.setFeeds()
				return nil
			}
		} else {
			feedLink, _ := feed.GetFeedLink()
			if feedLink == url {
				tui.FeedWidget.Feeds[i] = f
				tui.FeedWidget.setFeeds()
				return nil
			}
		}
	}
	tui.FeedWidget.Feeds = append(tui.FeedWidget.Feeds, f)
	tui.FeedWidget.setFeeds()
	return nil

}

func (tui *Tui) LoadCells(table *tview.Table, texts []string) {
	table.Clear()
	for i, text := range texts {
		table.SetCell(i, 0, tview.NewTableCell(text))
	}
}

func getDataPath() string {
	const dataRoot = "rfcui"
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, dataRoot)
}

func (tui *Tui) showDescription(texts [][]string) {
	var s string
	for _, line := range texts {
		for _, text := range line {
			s += text + " "
		}
		s += "\n"
	}
	tui.Description.SetText(s)
}

func (tui *Tui) Notify(text string) {
	tui.Info.SetText(text).SetTextColor(tcell.ColorReset)
}

func (tui *Tui) NotifyError(text string) {
	text = fmt.Sprint("error:\n", text)
	tui.Info.SetText(text).SetTextColor(tcell.ColorRed)
}

func (tui *Tui) UpdateHelp(text string) {
	tui.Help.SetText(text)
}

func (tui *Tui) RefreshTui() {
	switch tui.App.GetFocus() {
	case tui.GroupWidget.Table:
		row, column := tui.GroupWidget.Table.GetSelection()
		tui.selectGroupRow(row, column)
	case tui.FeedWidget.Table:
		row, column := tui.FeedWidget.Table.GetSelection()
		tui.selectFeedRow(row, column)
	case tui.SubWidget.Table:
		row, column := tui.SubWidget.Table.GetSelection()
		tui.selectSubRow(row, column)
	}
}

func (tui *Tui) setItems(paintColor, resetRow bool) {
	var (
		row   int
		items []*fd.Item
	)

	focus := tui.App.GetFocus()
	if focus == tui.GroupWidget.Table {
		row, _ = tui.GroupWidget.Table.GetSelection()
		items = tui.GroupWidget.Groups[row].Items
	} else if focus == tui.FeedWidget.Table {
		row, _ = tui.FeedWidget.Table.GetSelection()
		items = tui.FeedWidget.Feeds[row].Items
	}

	tui.SubWidget.Items = items

	table := tui.SubWidget.Table.Clear()
	for i, item := range items {
		table.SetCellSimple(i, 0, item.Title)
		if paintColor && item.Color > 0 && item.Color < len(mycolor.TcellColors) {
			table.GetCell(i, 0).SetTextColor(mycolor.TcellColors[item.Color])
		}
	}

	if tui.SubWidget.Table.GetRowCount() != 0 {
		if resetRow {
			tui.SubWidget.Table.Select(0, 0).ScrollToBeginning()
		} else {
			row, _ := tui.SubWidget.Table.GetSelection()
			max := tui.SubWidget.Table.GetRowCount() - 1
			if row > max {
				tui.SubWidget.Table.Select(max, 0).ScrollToEnd()
			} else {
				tui.SubWidget.Table.Select(row, 0)
			}
		}
	}
}

func (tui *Tui) GetTodaysFeeds() error {
	const feedname = "Today's Items"

	targetfeed, err := fd.MergeFeeds(tui.FeedWidget.Feeds, feedname)
	if err != nil {
		return err
	}

	// 現在時刻より未来のフィードを除外
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	result := make([]*fd.Item, 0)
	for _, item := range targetfeed.Items {
		if today.Before(item.PubDate) {
			result = append(result, item)
		}
	}
	targetfeed.Items = result

	isExist := false
	for i, g := range tui.GroupWidget.Groups {
		if g.Title == feedname {
			tui.GroupWidget.Groups[i] = targetfeed
			isExist = true
			break
		}
	}
	if !isExist {
		tui.GroupWidget.Groups = append(tui.GroupWidget.Groups, targetfeed)
	}
	tui.GroupWidget.setGroups()
	return nil
}

func getInvalidFeed(url string, err error) *fd.Feed {
	return &fd.Feed{
		Title:       "failed to retrieve: " + url,
		Color:       1, // Red
		Description: fmt.Sprint("Failed to retrieve feed:\n", err),
		Link:        "",
		FeedLinks:   []string{url},
		Items:       []*fd.Item{},
	}
}

func (tui *Tui) updateAllFeed() error {
	length := len(tui.FeedWidget.Feeds)
	doneCount := 0

	wg := sync.WaitGroup{}

	for index, feed := range tui.FeedWidget.Feeds {
		if !feed.IsMerged() {
			wg.Add(1)
			go func(i int) {
				if err := tui.updateFeed(i); err != nil {
					if !errors.Is(err, ErrGettingFeedFailed) {
						panic(err)
					}
				} else {
					if err := tui.FeedWidget.SaveFeed(tui.FeedWidget.Feeds[i]); err != nil {
						panic(err)
					}
				}
				doneCount++
				if doneCount == length {
					tui.Notify("All feeds are up-to-date.")
				} else {
					tui.Notify(fmt.Sprint("Updating ", doneCount, "/", length, " feeds...\r"))
				}
				tui.App.ForceDraw()
				wg.Done()
			}(index)
		}
	}

	wg.Wait()

	for index, feed := range tui.FeedWidget.Feeds {
		if feed.IsMerged() {
			if err := tui.updateFeed(index); err != nil {
				return err
			}
		}
	}

	if len(tui.FeedWidget.Feeds) > 0 {
		if err := tui.GetTodaysFeeds(); err != nil {
			return err
		}
		tui.FeedWidget.Table.ScrollToBeginning()
	}
	tui.GroupWidget.setGroups()
	tui.FeedWidget.setFeeds()
	tui.RefreshTui()

	return nil
}

func (tui *Tui) selectGroupRow(row, column int) {
	var feed *fd.Feed
	tui.Notify("")
	tui.ConfirmationStatus = 0
	if len(tui.GroupWidget.Groups) > 0 {
		feed = tui.GroupWidget.Groups[row]
		tui.setItems(true, tui.LastSelectedWidget == tui.GroupWidget.Table)
	}
	if tui.App.GetFocus() == tui.GroupWidget.Table {
		if len(tui.GroupWidget.Groups) > 0 {
			feedStatus := [][]string{
				{"Title:", feed.Title},
				{"Link:", feed.Link},
				{"Description:", feed.Description},
				{"Colorcode:", strconv.Itoa(feed.Color)},
			}
			tui.showDescription(feedStatus)
		}
		tui.UpdateHelp("[l]:move to SubColumn [r]:reload selecting feed [R]:reload All feeds [q]:quit rfcui")
	}
}

func (tui *Tui) selectFeedRow(row, column int) {
	var feed *fd.Feed
	tui.Notify("")
	tui.ConfirmationStatus = 0
	if len(tui.FeedWidget.Feeds) > 0 {
		feed = tui.FeedWidget.Feeds[row]
		tui.setItems(false, tui.LastSelectedWidget == tui.FeedWidget.Table)
	}
	if tui.App.GetFocus() == tui.FeedWidget.Table {
		if len(tui.FeedWidget.Feeds) > 0 {
			feedStatus := [][]string{
				{"Title:", feed.Title},
				{"Link:", feed.Link},
				{"Description:", feed.Description},
				{"Colorcode:", strconv.Itoa(feed.Color)},
			}
			tui.showDescription(feedStatus)
		}
		tui.UpdateHelp("[l]:move to SubColumn [r]:reload selecting feed [R]:reload All feeds [q]:quit rfcui")
	}
}

func (tui *Tui) selectSubRow(row, column int) {
	var (
		item       *fd.Item
		parentFeed *fd.Feed
		feedTitle  string
	)

	tui.Notify("")
	tui.UpdateHelp("[h]:move to MainColumn [o]:open an item with $BROWSER [q]:quit rfcui")

	if len(tui.SubWidget.Items) == 0 || len(tui.FeedWidget.Feeds) == 0 {
		return
	}

	item = tui.SubWidget.Items[row]

	index, _ := tui.FeedWidget.Table.GetSelection()
	parentFeed = tui.FeedWidget.Feeds[index]

	if tui.App.GetFocus() == tui.SubWidget.Table {
		if parentFeed.IsMerged() {
			for _, feed := range tui.FeedWidget.Feeds {
				feedLink, err := feed.GetFeedLink()
				if err != fd.ErrGetFeedLinkFailed {
					if item.Belong == feedLink {
						feedTitle = fmt.Sprint(feed.Title, "\n")
					}
				}
			}
		}
		itemText := [][]string{
			{"Feed:", feedTitle},
			{"Published:", item.FormatDate()},
			{"Title:", item.Title},
			{"Link:", item.Link},
		}
		tui.showDescription(itemText)
	}
}

func (tui *Tui) AddFeedsFromURL(path string) error {
	if !myio.IsFile(path) {
		return nil
	}

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

	for _, url := range newURLs {
		f := &fd.Feed{
			Title:       "getting " + url + "...",
			Color:       -1,
			Description: "update to get details",
			Link:        "",
			FeedLinks:   []string{url},
			Items:       []*fd.Item{},
		}
		tui.FeedWidget.Feeds = append(tui.FeedWidget.Feeds, f)
	}
	tui.FeedWidget.setFeeds()

	return nil
}

func (tui *Tui) LoadFeeds(path string) error {
	for _, file := range myio.DirWalk(path) {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		feed := fd.DecodeFeed(b)
		if feed.IsMerged() {
			tui.GroupWidget.Groups = append(tui.GroupWidget.Groups, feed)
		} else {
			tui.FeedWidget.Feeds = append(tui.FeedWidget.Feeds, feed)
		}
	}
	return nil
}

func NewTui() *Tui {

	groupTable := tview.NewTable()
	groupTable.SetTitle("Groups").SetBorder(true).SetTitleAlign(tview.AlignLeft)
	groupTable.Select(0, 0).SetSelectable(true, true)

	feedTable := tview.NewTable()
	feedTable.SetTitle("Feeds").SetBorder(true).SetTitleAlign(tview.AlignLeft)
	feedTable.Select(0, 0).SetSelectable(true, true)

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
				AddItem(groupTable, 0, 2, false).
				AddItem(feedTable, 0, 2, false).
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
				AddItem(groupTable, 0, 2, false).
				AddItem(feedTable, 0, 2, false).
				AddItem(infoWidget, 0, 1, false),
				0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(subTable, 0, 1, false).
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

	modal := tview.NewModal()
	modal.SetBorder(true).SetTitleAlign(0)
	modal.SetBackgroundColor(tcell.ColorBlack)

	pages := tview.NewPages().
		AddPage(mainPage, mainFlex, true, true).
		AddPage(descriptionPage, descriptionFlex, true, false).
		AddPage(inputField, inputFlex, true, false).
		AddPage(modalPage, modal, true, false)

	tui := &Tui{
		App:                tview.NewApplication(),
		Pages:              pages,
		GroupWidget:        &GroupWidget{groupTable, []*fd.Feed{}},
		FeedWidget:         &FeedWidget{feedTable, []*fd.Feed{}},
		SubWidget:          &SubWidget{subTable, []*fd.Item{}},
		Description:        descriptionWidget,
		Info:               infoWidget,
		Help:               helpWidget,
		InputWidget:        &InputBox{inputWidget, 0},
		WaitGroup:          &sync.WaitGroup{},
		SelectingFeeds:     []*fd.Feed{},
		ConfirmationStatus: 0,
		LastSelectedWidget: feedTable,
		Modal:              modal,
	}

	tui.setAppFunctions()

	return tui
}

func (tui *Tui) setAppFunctions() {

	tui.GroupWidget.Table.SetSelectionChangedFunc(func(row, column int) {
		tui.selectGroupRow(row, column)
	})
	tui.GroupWidget.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			tui.LastSelectedWidget = tui.GroupWidget.Table
			tui.App.SetFocus(tui.SubWidget.Table)
			tui.RefreshTui()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'R':
				if err := tui.updateAllFeed(); err != nil {
					panic(err)
				}
				return nil
			case 'r':
				tui.InputWidget.Input.SetTitle("rename the feed")
				tui.InputWidget.Mode = 3
				tui.Pages.ShowPage(inputField)
				tui.App.SetFocus(tui.InputWidget.Input)
				return nil
			case 'l':
				tui.App.SetFocus(tui.FeedWidget.Table)
				tui.RefreshTui()
				return nil
			case 'v':
				tui.SelectFeed()
			case 'c':
				tui.InputWidget.Input.SetTitle("Change the feed's color")
				tui.InputWidget.Mode = 2
				tui.Pages.ShowPage(inputField)
				tui.App.SetFocus(tui.InputWidget.Input)
				tui.Notify("input 256 color code, or input \"random\" to randomize color.")
			case 'm':
				if len(tui.SelectingFeeds) > 0 {
					tui.InputWidget.Input.SetTitle("Make a Group")
					tui.InputWidget.Mode = 1
					tui.Pages.ShowPage(inputField)
					tui.App.SetFocus(tui.InputWidget.Input)
				} else {
					tui.Notify("Select feeds with the s key to make a group.")
				}
				return nil
			case 'd':
				if tui.ConfirmationStatus == 1 {
					if err := tui.GroupWidget.DeleteSelection(); err != nil {
						if !errors.Is(err, ErrRmFailed) {
							panic(err)
						}
					}
					tui.GroupWidget.setGroups()
          tui.RefreshTui()
					tui.Notify("Deleted.")
					tui.ConfirmationStatus = 0
				} else {
					tui.Notify("Press d again to delete the feed.")
					tui.ConfirmationStatus = 1
				}
			case 'x':
				texts := []string{
					"c: recolor selecting feed",
					"d: delete selecting feed",
					"l: move to SubColumn",
					"r: rename selecting feed",
					"R: reload feeds",
					"q: Exit rfcui",
				}
				text := ""
				for _, line := range texts {
					text += line + "\n"
				}
				tui.Modal.SetTitle("keymaps")
				tui.Modal.SetText(text)
				tui.Pages.ShowPage(modalPage)
				tui.App.SetFocus(tui.Modal)
			}
		}
		return event
	})

	tui.FeedWidget.Table.SetSelectionChangedFunc(func(row, column int) {
		tui.selectFeedRow(row, column)
	})
	tui.FeedWidget.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			tui.LastSelectedWidget = tui.FeedWidget.Table
			tui.App.SetFocus(tui.SubWidget.Table)
			tui.RefreshTui()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'R':
				if err := tui.updateAllFeed(); err != nil {
					panic(err)
				}
				return nil
			case 'r':
				tui.InputWidget.Input.SetTitle("rename the feed")
				tui.InputWidget.Mode = 3
				tui.Pages.ShowPage(inputField)
				tui.App.SetFocus(tui.InputWidget.Input)
				return nil
			case 'h':
				tui.App.SetFocus(tui.GroupWidget.Table)
				tui.RefreshTui()
				return nil
			case 'v':
				tui.SelectFeed()
			case 'c':
				tui.InputWidget.Input.SetTitle("Change the feed's color")
				tui.InputWidget.Mode = 2
				tui.Pages.ShowPage(inputField)
				tui.App.SetFocus(tui.InputWidget.Input)
				tui.Notify("input 256 color code, or input \"random\" to randomize color.")
			case 'm':
				if len(tui.SelectingFeeds) > 0 {
					tui.InputWidget.Input.SetTitle("Make a Group")
					tui.InputWidget.Mode = 1
					tui.Pages.ShowPage(inputField)
					tui.App.SetFocus(tui.InputWidget.Input)
				} else {
					tui.Notify("Select feeds with the s key to make a group.")
				}
				return nil
			case 'd':
				if tui.ConfirmationStatus == 1 {
					if err := tui.FeedWidget.DeleteSelection(); err != nil {
						if !errors.Is(err, ErrRmFailed) {
							panic(err)
						}
					}
					tui.FeedWidget.setFeeds()
					tui.Notify("Deleted.")
					tui.ConfirmationStatus = 0
				} else {
					tui.Notify("Press d again to delete the feed.")
					tui.ConfirmationStatus = 1
				}
			case 'u':
				row, _ := tui.FeedWidget.Table.GetSelection()
				selectedFeed := tui.FeedWidget.Feeds[row]
				if !selectedFeed.IsMerged() {
					if tui.ConfirmationStatus == 2 {
						feedLink, _ := selectedFeed.GetFeedLink()
						feed, err := fd.GetFeedFromURL(feedLink, "")
						if err != nil {
							panic(err)
						}
						selectedFeed.Title = feed.Title

						if err := tui.FeedWidget.SaveFeed(selectedFeed); err != nil {
							panic(err)
						}

						tui.FeedWidget.setFeeds()
						tui.Notify("Reset.")
						tui.ConfirmationStatus = 0
					} else {
						tui.Notify("Press u again to reset the feed's title.")
						tui.ConfirmationStatus = 2
					}
				} else {
					tui.Notify("A mergedFeed cannot be reset.")
					tui.ConfirmationStatus = 0
				}
			case 'x':
				texts := []string{
					"c: recolor selecting feed",
					"d: delete selecting feed",
					"l: move to SubColumn",
					"r: rename selecting feed",
					"R: reload feeds",
					"q: Exit rfcui",
				}
				text := ""
				for _, line := range texts {
					text += line + "\n"
				}
				tui.Modal.SetTitle("keymaps")
				tui.Modal.SetText(text)
				tui.Pages.ShowPage(modalPage)
				tui.App.SetFocus(tui.Modal)
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
					tui.Notify("$BROWSER is empty. Set it and try again.")
				} else {
					if err := execCmd(true, browser, tui.SubWidget.Items[row].Link); err != nil {
						panic(err)
					}
				}
				return nil
			case tcell.KeyEscape:
				tui.App.SetFocus(tui.LastSelectedWidget)
				tui.LastSelectedWidget = tui.SubWidget.Table
				tui.RefreshTui()
				return nil
			case tcell.KeyRune:
				switch event.Rune() {
				case 'l':
					tui.Pages.SwitchToPage(descriptionPage)
					tui.App.SetFocus(tui.Description)
					return nil
				case 'o':
					row, _ := tui.SubWidget.Table.GetSelection()
					browser := os.Getenv("BROWSER")
					if browser == "" {
						tui.Notify("$BROWSER is empty. Set it and try again.")
					} else {
						if err := execCmd(true, browser, tui.SubWidget.Items[row].Link); err != nil {
							panic(err)
						}
					}
					return nil
				case 'x':
					texts := []string{
						"h: move to DescriptionColumn",
						"l: move to MainColumn",
						"q: Exit rfcui",
					}
					text := ""
					for _, line := range texts {
						text += line + "\n"
					}
					tui.Modal.SetTitle("keymaps")
					tui.Modal.SetText(text)
					tui.Pages.ShowPage(modalPage)
					tui.App.SetFocus(tui.Modal)
				}
			}
			return event
		})

	tui.Description.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'h':
				tui.Pages.SwitchToPage(mainPage)
				tui.App.SetFocus(tui.SubWidget.Table)
				return nil
			}
		}
		return event
	})

	tui.InputWidget.Input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			tui.InputWidget.Input.SetText("")
			tui.InputWidget.Input.SetTitle("Input")
			tui.Pages.HidePage(inputField)
			tui.App.SetFocus(tui.FeedWidget.Table)
			tui.Notify("")
			return nil
		case tcell.KeyEnter:
			switch tui.InputWidget.Mode {
			case 0: // new feed
				if err := tui.AddFeedFromURL(tui.InputWidget.Input.GetText()); err != nil {
					tui.NotifyError(err.Error())
				}
				tui.WaitGroup.Add(1)
				go func() {
					if err := tui.updateAllFeed(); err != nil {
						panic(err)
					}
					tui.App.QueueUpdateDraw(func() {})
					tui.WaitGroup.Done()
				}()
			case 1: // merge feeds
				title := tui.InputWidget.Input.GetText()
				existIndex := -1
				for i, feed := range tui.GroupWidget.Groups {
					if feed.IsMerged() && title == feed.Title {
						existIndex = i
						break
					}
				}

				if existIndex != -1 {
					for _, f := range tui.SelectingFeeds {
						isNewFeedLink := true
						feedLink, _ := f.GetFeedLink()
						for _, url := range tui.GroupWidget.Groups[existIndex].FeedLinks {
							if feedLink == url {
								isNewFeedLink = false
								break
							}
						}
						if isNewFeedLink {
							tui.GroupWidget.Groups[existIndex].FeedLinks = append(tui.GroupWidget.Groups[existIndex].FeedLinks, feedLink)
						}
					}
				} else {
					mergedFeed, err := fd.MergeFeeds(tui.SelectingFeeds, title)
					if err != nil {
						panic(err)
					}
					tui.GroupWidget.Groups = append(tui.GroupWidget.Groups, mergedFeed)
					if err := tui.GroupWidget.SaveGroup(mergedFeed); err != nil {
						panic(err)
					}
				}

				tui.WaitGroup.Add(1)
				go func() {
					if err := tui.updateAllFeed(); err != nil {
						panic(err)
					}
					tui.App.QueueUpdateDraw(func() {})
					tui.WaitGroup.Done()
				}()
			case 2: // change feed color
				var (
					number int
					err    error
				)
				if tui.InputWidget.Input.GetText() == "random" {
					number = int(mycolor.ComfortableColorCode[rand.Intn(len(mycolor.ComfortableColorCode))])
				} else {
					number, err = strconv.Atoi(tui.InputWidget.Input.GetText())
					if err != nil {
						tui.NotifyError(err.Error())
						break
					}
				}
				if number < 0 {
					number *= -1
				}
				number %= len(mycolor.ValidColorCode)
				row, _ := tui.FeedWidget.Table.GetSelection()
				tui.FeedWidget.Feeds[row].Color = number
				for _, item := range tui.FeedWidget.Feeds[row].Items {
					item.Color = number
				}
				if err := tui.FeedWidget.SaveFeed(tui.FeedWidget.Feeds[row]); err != nil {
					panic(err)
				}
				tui.FeedWidget.setFeeds()
				tui.setItems(tui.FeedWidget.Feeds[row].IsMerged(), false)
				tui.Notify(fmt.Sprint("set color-number as ", number))
			case 3:
				title := tui.InputWidget.Input.GetText()
				row, _ := tui.FeedWidget.Table.GetSelection()
				selectedFeed := tui.FeedWidget.Feeds[row]
				if selectedFeed.IsMerged() {
					// rename the cache file
					oldFileName := filepath.Join(getDataPath(), fmt.Sprintf("%x", md5.Sum([]byte(selectedFeed.Title))))
					newFileName := filepath.Join(getDataPath(), fmt.Sprintf("%x", md5.Sum([]byte(title))))
					if err := os.Rename(oldFileName, newFileName); err != nil {
						panic(err)
					}
				}
				selectedFeed.Title = title
				if err := tui.FeedWidget.SaveFeed(selectedFeed); err != nil {
					panic(err)
				}
				tui.FeedWidget.setFeeds()
			}
			tui.SelectingFeeds = []*fd.Feed{}
			tui.InputWidget.Input.SetText("")
			tui.InputWidget.Input.SetTitle("Input")
			tui.Pages.HidePage(inputField)
			tui.App.SetFocus(tui.FeedWidget.Table)
			return nil
		}
		return event
	})

	tui.Modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		tui.Pages.HidePage(modalPage)
		tui.Modal.SetText("")
		tui.App.SetFocus(tui.FeedWidget.Table)
		return event
	})

	tui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if tui.App.GetFocus() == tui.InputWidget.Input {
			return event
		}
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'n':
				tui.InputWidget.Input.SetTitle("New Feed")
				tui.InputWidget.Mode = 0
				tui.Pages.ShowPage(inputField)
				tui.App.SetFocus(tui.InputWidget.Input)
				tui.Notify("Enter a feed URL or a command to output feed as xml.")
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
	fmt.Print("loading...\r")

	if !myio.IsDir(cachePath) {
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			return err
		}
	}

	if err := tui.LoadFeeds(cachePath); err != nil {
		return err
	}

	listPath := filepath.Join(getDataPath(), "list.txt")
	if err := tui.AddFeedsFromURL(listPath); err != nil {
		return err
	}

	tui.WaitGroup.Add(1)
	go func() {
		if err := tui.updateAllFeed(); err != nil {
			panic(err)
		}
		tui.App.QueueUpdateDraw(func() {})
		tui.WaitGroup.Done()
	}()

	if err := tui.App.SetRoot(tui.Pages, true).SetFocus(tui.FeedWidget.Table).Run(); err != nil {
		tui.WaitGroup.Wait()
		tui.App.Stop()
		return err
	}

	return nil
}
