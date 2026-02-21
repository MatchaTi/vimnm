package main

// Network Manager CLI Tool with Vim Motion

import (
	"fmt"
	"github.com/MatchaTi/vimnm/ui"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	p := tea.NewProgram(ui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
