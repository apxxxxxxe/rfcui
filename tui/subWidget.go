package tui

import (
  fd "github.com/apxxxxxxe/rfcui/feed"

  "github.com/rivo/tview"
)

type SubWidget struct {
  Table *tview.Table
  Items []*fd.Item
}

