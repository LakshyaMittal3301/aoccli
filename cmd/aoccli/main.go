package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LakshyaMittal3301/aoccli/internal/config"
	"github.com/LakshyaMittal3301/aoccli/internal/tui"
)

func main() {
	cfg, err := config.Load()
	m := tui.New(cfg, err)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "aoccli:", err)
		os.Exit(1)
	}
}
