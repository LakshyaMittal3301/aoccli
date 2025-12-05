package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/LakshyaMittal3301/aoccli/internal/config"
	"github.com/LakshyaMittal3301/aoccli/internal/tui"
)

func main() {
	reset := false
	help := false

	fs := flag.NewFlagSet("aoccli", flag.ExitOnError)
	fs.BoolVar(&reset, "reset-config", false, "delete the saved config file and exit")
	fs.BoolVar(&help, "help", false, "show help")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		fs.PrintDefaults()
	}

	// Let -h/--help work; ExitOnError already handles -h, but we also honor --help via the bool.
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if help {
		fs.Usage()
		return
	}

	if reset {
		if err := deleteConfig(); err != nil {
			fmt.Fprintln(os.Stderr, "aoccli: failed to reset config:", err)
			os.Exit(1)
		}
		fmt.Println("aoccli: config deleted")
		return
	}

	cfg, err := config.Load()
	m := tui.New(cfg, err)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "aoccli:", err)
		os.Exit(1)
	}
}

func deleteConfig() error {
	p, err := config.Path()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
