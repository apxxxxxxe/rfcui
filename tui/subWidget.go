package tui

import(
  "github.com/apxxxxxxe/rfcui/feed"

  "github.com/rivo/tview"
)

type SubWidget struct {
  Table *tview.Table
  Items []*feed.Article
}

func (s *SubWidget) GetArticleTitles() []string {
  titles := []string{}
  for _, item := range s.Items {
    titles = append(titles, item.Title)
  }
  return titles
}

