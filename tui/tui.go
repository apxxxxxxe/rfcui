package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"os"
	"os/exec"
)

type Tui struct {
	App        *tview.Application
	Pages      *tview.Pages
	Widgets    []*tview.Table
	FocusIndex int
}

func newTui() *Tui {
	tui := &Tui{
		App:        tview.NewApplication(),
		Pages:      tview.NewPages(),
		FocusIndex: 0,
		Widgets:    make([]*tview.Table, 0),
	}

	mainWidget := tview.NewTable().Select(0, 0).SetFixed(1, 1).SetSelectable(true, true)
	mainWidget.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			tui.App.Stop()
		}
		if key == tcell.KeyEnter {
			mainWidget.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {
		mainWidget.GetCell(row, column).SetTextColor(tcell.ColorRed)
		mainWidget.SetSelectable(false, false)
	})

	mainWidget.
		SetCellSimple(0, 0, "1").
		SetCellSimple(1, 0, "2").
		SetCellSimple(2, 0, "3").
		SetCellSimple(3, 0, "4").
		SetCellSimple(4, 0, "5").
		SetCellSimple(5, 0, "6").
		SetCellSimple(6, 0, "7")
	mainWidget.SetTitle("Feeds").SetBorder(true)

	subWidget := tview.NewTable()
	subWidget.SetTitle("Articles").
		SetBorder(true)

	descWidget := tview.NewBox()
	descWidget.SetTitle("Details").
		SetBorder(true)

	tui.Widgets = append(tui.Widgets, mainWidget, subWidget)

	tui.App.SetFocus(tui.Widgets[0])

	mainWidget.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlF:
			// CtrlFを押した時の処理を記述
			return event // CtrlFをInputFieldのdefaultのキー設定へ伝える
		case tcell.KeyRune:
			switch event.Rune() {
			case 'a':
				// aを押した時の処理を記述
				return nil // aを入力してもdefaultのキー設定へ伝えない
			case 'b':
				// bを押した時の処理を記述
				return nil // bを入力してもdefaultのキー設定へ伝えない
			}
		}
		return event // 上記以外のキー入力をdefaultのキーアクションへ伝える
	})

	grid := tview.NewGrid()
	grid.SetTitle("grid").SetBorder(false)
	grid.SetSize(6, 5, 0, 0).
		AddItem(mainWidget, 0, 0, 6, 2, 0, 0, true).
		AddItem(subWidget, 0, 2, 4, 4, 0, 0, true).
		AddItem(descWidget, 4, 2, 2, 4, 0, 0, true)

	tui.Pages.SetTitle("page1").SetBorder(false)
	tui.Pages.AddPage("MainPage", grid, true, true)

	tui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'h':
				tui.App.SetFocus(tui.Widgets[0])
				return nil
			case 'l':
				tui.App.SetFocus(tui.Widgets[1])
				return nil
			}
		}
		return event
	})

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

func (t Tui) loadCells(index int, texts []string) {
	for i, text := range texts {
		t.Widgets[index].SetCell(i, 0, tview.NewTableCell(text))
	}
}

func main() {
	tui := newTui()

	texts := []string{"a", "b", "c"}
	tui.loadCells(0, texts)

	if err := tui.App.SetRoot(tui.Pages, true).Run(); err != nil {
		panic(err)
	}
}
