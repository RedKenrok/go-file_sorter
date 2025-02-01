package main

import (
	"flag"
	"fmt"

	app "github.com/redkenrok/go-file_sorter/internal/app"
)

var (
	version   = "Dev"
	commit    = "None"
	buildDate = "Unknown"
)

func main() {
	flag.Usage = func() {
		fmt.Println("For CLI usage run: `file_sorter --help`.\n")
		fmt.Println("For TUI usage run without options.\n")
	}
	flag.Parse()

	if flag.NFlag() == 0 {
		app.RunTUI(
			version,
			commit,
			buildDate,
		)
		return
	}

	app.RunCLI(
		version,
		commit,
		buildDate,
	)
}
