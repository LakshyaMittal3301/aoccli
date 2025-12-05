package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/LakshyaMittal3301/aoccli/internal/aoc"
	"github.com/LakshyaMittal3301/aoccli/internal/config"
)

type appState int

const (
	stateConfig appState = iota
	stateLoading
	stateLeaderboard
)

type Model struct {
	state  appState
	cfg    config.Config
	cfgErr error

	textInput textinput.Model

	leaderboard *aoc.Leaderboard
	entries     []aoc.DayEntry
	currentDay  int
	maxDay      int

	dayPicker bool
	pickerDay int

	width, height int

	err error
}

// Messages for async work.
type leaderboardLoadedMsg struct {
	lb *aoc.Leaderboard
}

type errMsg struct {
	err error
}

// New builds the Bubbletea model, using cfg and cfgErr from config.Load().
func New(cfg config.Config, cfgErr error) Model {
	ti := textinput.New()
	ti.Placeholder = "Paste AoC private leaderboard JSON URL"
	ti.CharLimit = 512
	ti.Width = 80
	if cfg.LeaderboardURL == "" {
		ti.Focus()
	}

	m := Model{
		cfg:        cfg,
		cfgErr:     cfgErr,
		textInput:  ti,
		currentDay: 1,
	}

	if cfgErr == nil && cfg.LeaderboardURL != "" {
		m.state = stateLoading
	} else {
		m.state = stateConfig
		if cfgErr != nil && !errors.Is(cfgErr, config.ErrNotFound) {
			m.err = cfgErr
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	switch m.state {
	case stateConfig:
		return textinput.Blink
	case stateLoading:
		return fetchLeaderboardCmd(m.cfg.LeaderboardURL)
	default:
		return nil
	}
}

func fetchLeaderboardCmd(url string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		lb, err := aoc.FetchLeaderboard(ctx, url)
		if err != nil {
			return errMsg{err: err}
		}
		return leaderboardLoadedMsg{lb: lb}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case errMsg:
		m.err = msg.err
		// keep current state (if loading, we fall back to leaderboard view with error)
		if m.state == stateLoading {
			m.state = stateLeaderboard
		}
		return m, nil

	case leaderboardLoadedMsg:
		m.leaderboard = msg.lb
		m.maxDay = aoc.MaxAvailableDay(msg.lb)
		if m.currentDay < 1 {
			m.currentDay = 1
		}
		if m.currentDay > m.maxDay {
			m.currentDay = m.maxDay
		}
		m.entries = aoc.BuildDayEntries(msg.lb, m.currentDay)
		m.state = stateLeaderboard
		m.err = nil
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateConfig:
			return m.updateConfigKey(msg)
		case stateLoading:
			if key := msg.String(); key == "ctrl+c" || key == "q" {
				return m, tea.Quit
			}
			return m, nil
		case stateLeaderboard:
			return m.updateLeaderboardKey(msg)
		}
	}

	// Let the text input handle messages in config mode (e.g. cursor, typing).
	if m.state == stateConfig {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) updateConfigKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		url := strings.TrimSpace(m.textInput.Value())
		if url == "" {
			m.err = errors.New("URL cannot be empty")
			return m, nil
		}
		m.cfg.LeaderboardURL = url
		if err := config.Save(m.cfg); err != nil {
			m.err = err
			return m, nil
		}
		m.state = stateLoading
		m.err = nil
		return m, fetchLeaderboardCmd(url)

	case "ctrl+c", "esc":
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) updateLeaderboardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" || key == "q" {
		return m, tea.Quit
	}

	if m.leaderboard == nil {
		return m, nil
	}

	// Day picker mode
	if m.dayPicker {
		switch key {
		case "up", "k":
			if m.pickerDay > 1 {
				m.pickerDay--
			}
		case "down", "j":
			if m.pickerDay < m.maxDay {
				m.pickerDay++
			}
		case "enter":
			m.currentDay = m.pickerDay
			m.entries = aoc.BuildDayEntries(m.leaderboard, m.currentDay)
			m.dayPicker = false
		case "esc", "d":
			m.dayPicker = false
		}
		return m, nil
	}

	// Normal leaderboard navigation.
	switch key {
	case "left", "h":
		if m.currentDay > 1 {
			m.currentDay--
			m.entries = aoc.BuildDayEntries(m.leaderboard, m.currentDay)
		}
	case "right", "l":
		if m.currentDay < m.maxDay {
			m.currentDay++
			m.entries = aoc.BuildDayEntries(m.leaderboard, m.currentDay)
		}
	case "d":
		m.dayPicker = true
		m.pickerDay = m.currentDay
	case "r":
		m.state = stateLoading
		return m, fetchLeaderboardCmd(m.cfg.LeaderboardURL)
	}

	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateConfig:
		return m.viewConfig()
	case stateLoading:
		return m.viewLoading()
	case stateLeaderboard:
		return m.viewLeaderboard()
	default:
		return ""
	}
}

// ---- View helpers ----

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	headerStyle = lipgloss.NewStyle().Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	helpStyle   = lipgloss.NewStyle().Faint(true)
)

func (m Model) viewConfig() string {
	var b strings.Builder

	fmt.Fprintln(&b, titleStyle.Render("Advent of Code – aoccli"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Enter your private leaderboard JSON URL:")
	fmt.Fprintln(&b, m.textInput.View())
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, helpStyle.Render("Press Enter to save, Esc or Ctrl+C to quit."))

	if m.err != nil {
		fmt.Fprintln(&b, errorStyle.Render("Error: "+m.err.Error()))
	}

	return b.String()
}

func (m Model) viewLoading() string {
	var b strings.Builder

	fmt.Fprintln(&b, titleStyle.Render("Advent of Code – aoccli"))
	fmt.Fprintln(&b)
	msg := "Loading leaderboard..."
	if m.cfg.LeaderboardURL == "" {
		msg = "No leaderboard URL configured."
	}
	fmt.Fprintln(&b, msg)

	if m.err != nil {
		fmt.Fprintln(&b, errorStyle.Render("Error: "+m.err.Error()))
	}

	fmt.Fprintln(&b)
	fmt.Fprintln(&b, helpStyle.Render("Press q to quit."))

	return b.String()
}

func (m Model) viewLeaderboard() string {
	var b strings.Builder

	if m.leaderboard == nil {
		fmt.Fprintln(&b, "No leaderboard loaded.")
		fmt.Fprintln(&b, helpStyle.Render("Press r to reload or q to quit."))
		return b.String()
	}

	header := fmt.Sprintf("Advent of Code %s – Day %d / %d", m.leaderboard.Event, m.currentDay, m.maxDay)
	fmt.Fprintln(&b, titleStyle.Render(header))

	if m.err != nil {
		fmt.Fprintln(&b, errorStyle.Render("Error: "+m.err.Error()))
	}
	fmt.Fprintln(&b)

	// Day picker overlay.
	if m.dayPicker {
		fmt.Fprintln(&b, headerStyle.Render("Select day"))
		for d := 1; d <= m.maxDay; d++ {
			cursor := "  "
			if d == m.pickerDay {
				cursor = "➤ "
			}
			fmt.Fprintf(&b, "%sDay %02d\n", cursor, d)
		}
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, helpStyle.Render("↑/↓ or j/k to move · Enter to select · d/Esc to cancel · q to quit"))
		return b.String()
	}

	fmt.Fprintln(&b, headerStyle.Render(
		fmt.Sprintf("%4s  %4s  %-8s  %-8s  %-30s",
			"Pos", "Pts", "P1", "P2", "Name"),
	))

	// Rows.
	// Rows.
	lastPos := -1
	for _, e := range m.entries {
		p1 := "-"
		p2 := "-"

		if e.HasPart1 {
			p1 = formatDuration(e.Part1Since)
		}
		if e.HasPart2 {
			p2 = formatDuration(e.Part2Since)
		}

		// Show position only for the first person with that position (ties get blank).
		posStr := ""
		if e.Pos != lastPos {
			posStr = formatPosition(e.Pos)
			lastPos = e.Pos
		}

		name := truncate(e.Name+" "+strings.Repeat("★", e.StarsToday), 30)

		fmt.Fprintf(
			&b,
			"%4s  %4d  %-8s  %-8s  %-30s\n",
			posStr,
			e.DayScore,
			p1,
			p2,
			name,
		)
	}

	fmt.Fprintln(&b)
	fmt.Fprintln(&b, helpStyle.Render("←/h prev day · →/l next day · d day list · r refresh · q quit"))
	fmt.Fprintln(&b, helpStyle.Render("Times are HH:MM:SS since midnight (UTC-5) release."))

	return b.String()
}

// formatPosition renders AoC-style positions (" 1)", "13)", etc.).
func formatPosition(pos int) string {
	return fmt.Sprintf("%2d)", pos)
}

// formatDuration renders a duration as HH:MM:SS.
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSeconds := int64(d.Seconds())
	h := totalSeconds / 3600
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// truncate shortens a string to max runes and appends … if needed.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}
