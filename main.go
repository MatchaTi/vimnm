package main

// Network Manager CLI Tool with Vim Motion

import (
	"flag"
	"fmt"
	"github.com/MatchaTi/vimnm/ui"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

func main() {
	noClear := flag.Bool("no-clear", false, "do not clear terminal when exiting")
	flag.Parse()

	exitCode := 0

	p := tea.NewProgram(ui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		exitCode = 1
	}

	if !*noClear {
		clearTerminal()
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
