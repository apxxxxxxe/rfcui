package tui

import(
  "github.com/apxxxxxxe/rfcui/feed"

  "github.com/rivo/tview"
)

type MainWidget struct {
  Table *tview.Table
  Feeds []*feed.Feed
}

func (m *MainWidget) GetFeedTitles() []string {
  titles := []string{}
  for _, feed := range m.Feeds {
    titles = append(titles, feed.Title)
  }
  return titles
}

