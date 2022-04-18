package tui

import (
	"crypto/md5"
	"errors"
	"fmt"
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
)

var (
	ErrGettingFeedFailed = errors.New("failed to get feed")
	ErrRmFailed          = errors.New("faled to remove files or dirs")
)

type Tui struct {
	App                *tview.Application
	Pages              *tview.Pages
	MainWidget         *MainWidget
	SubWidget          *SubWidget
	Description        *tview.TextView
	Info               *tview.TextView
	Help               *tview.TextView
	InputWidget        *InputBox
	WaitGroup          *sync.WaitGroup
	SelectingFeeds     []*fd.Feed
	ConfirmationStatus int
}

func (tui *Tui) SelectFeed() {
	const defaultColor = tcell.ColorBlack
	const selectedColor = tcell.ColorWhite

	row, column := tui.MainWidget.Table.GetSelection()
	if tui.MainWidget.Table.GetCell(row, column).BackgroundColor == defaultColor {
		tui.MainWidget.Table.GetCell(row, column).SetBackgroundColor(selectedColor)
		tui.SelectingFeeds = append(tui.SelectingFeeds, tui.MainWidget.Feeds[row])
	} else {
		tui.MainWidget.Table.GetCell(tui.MainWidget.Table.GetSelection()).SetBackgroundColor(defaultColor)
		targetFeed := tui.MainWidget.Feeds[row]
		for i, f := range tui.SelectingFeeds {
			if f == targetFeed {
				tui.SelectingFeeds = append(tui.SelectingFeeds[:i], tui.SelectingFeeds[i+1:]...)
				break
			}
		}
	}
}

func (tui *Tui) updateFeed(index int) error {
	targetFeed := tui.MainWidget.Feeds[index]

	if targetFeed.IsMerged() {
		tui.MainWidget.Feeds[index].Items = []*fd.Item{}
		for _, url := range targetFeed.FeedLinks {
			for _, f := range tui.MainWidget.Feeds {
				if !f.IsMerged() {
					feedLink, _ := f.GetFeedLink()
					if url == feedLink {
						tui.MainWidget.Feeds[index].Items = append(tui.MainWidget.Feeds[index].Items, f.Items...)
						break
					}
				}
			}
		}
		tui.MainWidget.Feeds[index].SortItems()

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

	if err := tui.MainWidget.SaveFeed(f); err != nil {
		return err
	}

	for i, feed := range tui.MainWidget.Feeds {
		if f.IsMerged() {
			if f.Title == feed.Title {
				tui.MainWidget.Feeds[i] = f
				tui.MainWidget.setFeeds()
				return nil
			}
		} else {
			feedLink, _ := feed.GetFeedLink()
			if feedLink == url {
				tui.MainWidget.Feeds[i] = f
				tui.MainWidget.setFeeds()
				return nil
			}
		}
	}
	tui.MainWidget.Feeds = append(tui.MainWidget.Feeds, f)
	tui.MainWidget.setFeeds()
	return nil

}

func (tui *Tui) LoadCells(table *tview.Table, texts []string) {
	table.Clear()
	for i, text := range texts {
		table.SetCell(i, 0, tview.NewTableCell(text))
	}
}

func getDataPath() string {
	const dataRoot = "cache"
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, dataRoot)
}

func (tui *Tui) showDescription(text string) {
	tui.Description.SetText(text)
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
		if paintColor && item.Color > 0 && item.Color < len(mycolor.TcellColors) {
			table.GetCell(i, 0).SetTextColor(mycolor.TcellColors[item.Color])
		}
	}

	if tui.SubWidget.Table.GetRowCount() != 0 {
		tui.SubWidget.Table.Select(0, 0).ScrollToBeginning()
	}
}

func (tui *Tui) GetTodaysFeeds() error {
	const feedname = "Today's Items"

	targetfeed, err := fd.MergeFeeds(tui.MainWidget.Feeds, feedname)
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
	tui.MainWidget.setFeeds()
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
	length := len(tui.MainWidget.Feeds)
	doneCount := 0

	wg := sync.WaitGroup{}

	for index, feed := range tui.MainWidget.Feeds {
		if !feed.IsMerged() {
			wg.Add(1)
			go func(i int) {
				if err := tui.updateFeed(i); err != nil {
					if !errors.Is(err, ErrGettingFeedFailed) {
						panic(err)
					}
				} else {
					if err := tui.MainWidget.SaveFeed(tui.MainWidget.Feeds[i]); err != nil {
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

	for index, feed := range tui.MainWidget.Feeds {
		if feed.IsMerged() {
			if err := tui.updateFeed(index); err != nil {
				return err
			}
		}
	}

	if len(tui.MainWidget.Feeds) > 0 {
		if err := tui.GetTodaysFeeds(); err != nil {
			return err
		}
		tui.MainWidget.Table.ScrollToBeginning()
	}
	tui.MainWidget.setFeeds()
	tui.RefreshTui()

	return nil
}

func (tui *Tui) selectMainRow(row, column int) {
	var feed *fd.Feed
	tui.Notify("")
	tui.ConfirmationStatus = 0
	if len(tui.MainWidget.Feeds) > 0 {
		feed = tui.MainWidget.Feeds[row]
		tui.setItems(tui.MainWidget.Feeds[row].IsMerged())
	}
	if tui.App.GetFocus() == tui.MainWidget.Table {
		if len(tui.MainWidget.Feeds) > 0 {
			tui.showDescription(fmt.Sprint("Title:       ", feed.Title, "\n", "Link:        ", feed.Link, "\n", "Description: ", feed.Description, "\n", "Colorcode:   ", feed.Color))
		}
		tui.UpdateHelp("[l]:move to SubColumn [r]:reload selecting feed [R]:reload All feeds [q]:quit rfcui")
	}
}

func (tui *Tui) selectSubRow(row, column int) {
	var (
		item      *fd.Item
		feedTitle = ""
	)
	tui.Notify("")
	if len(tui.SubWidget.Items) > 0 {
		item = tui.SubWidget.Items[row]
	}

	index, _ := tui.MainWidget.Table.GetSelection()
	parentFeed := tui.MainWidget.Feeds[index]

	if tui.App.GetFocus() == tui.SubWidget.Table {
		if len(tui.SubWidget.Items) > 0 {
			if parentFeed.IsMerged() {
				for _, feed := range tui.MainWidget.Feeds {
					feedLink, err := feed.GetFeedLink()
					if err != fd.ErrGetFeedLinkFailed {
						if item.Belong == feedLink {
							feedTitle = fmt.Sprint(feed.Title, "\n")
						}
					}
				}
			}
			tui.showDescription(fmt.Sprint(feedTitle, item.FormatDate(), "\n", item.Title, "\n", item.Link))
		}
		tui.UpdateHelp("[h]:move to MainColumn [o]:open an item with $BROWSER [q]:quit rfcui")
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
		tui.MainWidget.Feeds = append(tui.MainWidget.Feeds, f)
	}
	tui.MainWidget.setFeeds()

	return nil
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
		App:                tview.NewApplication(),
		Pages:              pages,
		MainWidget:         &MainWidget{mainTable, []*fd.Feed{}},
		SubWidget:          &SubWidget{subTable, []*fd.Item{}},
		Description:        descriptionWidget,
		Info:               infoWidget,
		Help:               helpWidget,
		InputWidget:        &InputBox{inputWidget, 0},
		WaitGroup:          &sync.WaitGroup{},
		SelectingFeeds:     []*fd.Feed{},
		ConfirmationStatus: 0,
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
				tui.App.SetFocus(tui.SubWidget.Table)
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
					if err := tui.MainWidget.DeleteSelection(); err != nil {
						if !errors.Is(err, ErrRmFailed) {
							panic(err)
						}
					}
					tui.MainWidget.setFeeds()
					tui.Notify("Deleted.")
					tui.ConfirmationStatus = 0
				} else {
					tui.Notify("Press d again to delete the feed.")
					tui.ConfirmationStatus = 1
				}
			case 'u':
				row, _ := tui.MainWidget.Table.GetSelection()
				selectedFeed := tui.MainWidget.Feeds[row]
				if !selectedFeed.IsMerged() {
					if tui.ConfirmationStatus == 2 {
						feedLink, _ := selectedFeed.GetFeedLink()
						feed, err := fd.GetFeedFromURL(feedLink, "")
						if err != nil {
							panic(err)
						}
						selectedFeed.Title = feed.Title

						if err := tui.MainWidget.SaveFeed(selectedFeed); err != nil {
							panic(err)
						}

						tui.MainWidget.setFeeds()
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
						tui.Notify("$BROWSER is empty. Set it and try again.")
					} else {
						if err := execCmd(true, browser, tui.SubWidget.Items[row].Link); err != nil {
							panic(err)
						}
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
		case tcell.KeyESC:
			tui.InputWidget.Input.SetText("")
			tui.InputWidget.Input.SetTitle("Input")
			tui.Pages.HidePage(inputField)
			tui.App.SetFocus(tui.MainWidget.Table)
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
				for i, feed := range tui.MainWidget.Feeds {
					if feed.IsMerged() && title == feed.Title {
						existIndex = i
						break
					}
				}

				if existIndex != -1 {
					for _, f := range tui.SelectingFeeds {
						if !f.IsMerged() {
							isNewFeedLink := true
							feedLink, _ := f.GetFeedLink()
							for _, url := range tui.MainWidget.Feeds[existIndex].FeedLinks {
								if feedLink == url {
									isNewFeedLink = false
									break
								}
							}
							if isNewFeedLink {
								tui.MainWidget.Feeds[existIndex].FeedLinks = append(tui.MainWidget.Feeds[existIndex].FeedLinks, feedLink)
							}
						}
					}
				} else {
					mergedFeed, err := fd.MergeFeeds(tui.SelectingFeeds, title)
					if err != nil {
						panic(err)
					}
					tui.MainWidget.Feeds = append(tui.MainWidget.Feeds, mergedFeed)
					if err := tui.MainWidget.SaveFeed(mergedFeed); err != nil {
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
				row, _ := tui.MainWidget.Table.GetSelection()
				tui.MainWidget.Feeds[row].Color = number
				for _, item := range tui.MainWidget.Feeds[row].Items {
					item.Color = number
				}
				if err := tui.MainWidget.SaveFeed(tui.MainWidget.Feeds[row]); err != nil {
					panic(err)
				}
				tui.MainWidget.setFeeds()
				tui.Notify(fmt.Sprint("set color-number as ", number))
			case 3:
				title := tui.InputWidget.Input.GetText()
				row, _ := tui.MainWidget.Table.GetSelection()
				selectedFeed := tui.MainWidget.Feeds[row]
				selectedFeed.Title = title
				if selectedFeed.IsMerged() {
					tui.MainWidget.DeleteFeedFile(row)
				}
				if err := tui.MainWidget.SaveFeed(selectedFeed); err != nil {
					panic(err)
				}
				tui.MainWidget.setFeeds()
			}
			tui.SelectingFeeds = []*fd.Feed{}
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

	fmt.Print("loading...\r")

	if !myio.IsDir(getDataPath()) {
		if err := os.MkdirAll(getDataPath(), 0755); err != nil {
			return err
		}
	}

	if err := tui.MainWidget.LoadFeeds(getDataPath()); err != nil {
		return err
	}

	if err := tui.AddFeedsFromURL("list.txt"); err != nil {
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

	if err := tui.App.SetRoot(tui.Pages, true).SetFocus(tui.MainWidget.Table).Run(); err != nil {
		tui.WaitGroup.Wait()
		tui.App.Stop()
		return err
	}

	return nil
}
