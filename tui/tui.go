package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"time"
)

func main() {
	index := 0

	app := tview.NewApplication()

	widget1 := tview.NewBox()
	widget1.SetTitle("widget1").
		SetBorder(true)

	widget2 := tview.NewBox()
	widget2.SetTitle("widget2").
		SetBorder(true)

	widget3 := tview.NewBox()
	widget3.SetTitle("widget3").
		SetBorder(true)

	widget4 := tview.NewTable().Select(0, 0).SetFixed(1, 1).SetSelectable(true, true)
	widget4.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			widget4.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {
		widget4.GetCell(row, column).SetTextColor(tcell.ColorRed)
		widget4.SetSelectable(false, false)
	})
	widget4.
		SetCellSimple(0, 0, "1").
		SetCellSimple(1, 0, "2").
		SetCellSimple(2, 0, "3").
		SetCellSimple(3, 0, "4").
		SetCellSimple(4, 0, "5").
		SetCellSimple(5, 0, "6").
		SetCellSimple(6, 0, "7")
	widget4.SetTitle("widget4").
		SetBorder(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyCtrlF:
				// CtrlFを押した時の処理を記述
				return event // CtrlFをInputFieldのdefaultのキー設定へ伝える

			case tcell.KeyRune:
				switch event.Rune() {
				case 'a':
					// aを押した時の処理を記述
					widget4.SetCellSimple(index, 0, fmt.Sprint(time.Now().UnixNano()))
					index++
					return nil // aを入力してもdefaultのキー設定へ伝えない
				case 'b':
					// bを押した時の処理を記述
					return nil // bを入力してもdefaultのキー設定へ伝えない
				}
			}
			return event // 上記以外のキー入力をdefaultのキーアクションへ伝える
		})

	widget5 := tview.NewBox()
	widget5.SetTitle("widget5").
		SetBorder(true)

	widget6 := tview.NewBox()
	widget6.SetTitle("widget6").
		SetBorder(true)

	flex := tview.NewFlex()
	flex.SetTitle("flex").
		SetBorder(true)

	flex.SetDirection(tview.FlexRow).
		AddItem(widget1, 3, 0, true).
		AddItem(widget2, 0, 1, false).
		AddItem(widget3, 0, 2, false)

	grid := tview.NewGrid()
	grid.SetTitle("grid").
		SetBorder(false)

	grid.SetSize(5, 5, 0, 0).
		AddItem(widget4, 0, 0, 5, 2, 0, 0, true).
		AddItem(widget5, 0, 2, 3, 4, 0, 0, true).
		AddItem(widget6, 3, 2, 2, 4, 0, 0, true)

	page := tview.NewPages()
	page.SetTitle("page1").
		SetBorder(false)

	page.AddPage("page1", flex, true, true).
		AddPage("page2", grid, true, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			curPage, _ := page.GetFrontPage()
			if curPage == "page1" {
				page.SwitchToPage("page2")
				page.SetTitle("page2")
			} else {
				page.SwitchToPage("page1")
				page.SetTitle("page1")
			}
			return nil
		}
		return event
	})

	if err := app.SetRoot(page, true).Run(); err != nil {
		panic(err)
	}
}
