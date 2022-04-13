package tui

import (
  "github.com/apxxxxxxe/rfcui/feed"

  "github.com/rivo/tview"
)

type SubWidget struct {
  Table *tview.Table
  Items []*feed.Item
}

