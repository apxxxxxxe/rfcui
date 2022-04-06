package main

import (
	"github.com/apxxxxxxe/rfcui/tui"
)

func main() {
	if err := tui.NewTui().Run(); err != nil {
		panic(err)
	}
}
