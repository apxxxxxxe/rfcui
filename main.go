package main

import (
	"github.com/apxxxxxxe/rfcui/tui"

	"log"
)

func main() {
	if err := tui.NewTui().Run(); err != nil {
		log.Fatal(err)
	}
}
