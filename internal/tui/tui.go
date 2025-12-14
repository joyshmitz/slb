// Package tui implements the Bubble Tea terminal UI for SLB.
// Uses the Charmbracelet ecosystem: Bubble Tea, Bubbles, Lip Gloss, Glamour.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the main TUI model.
type Model struct {
	ready  bool
	width  int
	height int
}

// New creates a new TUI model.
func New() Model {
	return Model{}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	return "SLB TUI - Press q to quit"
}

// Run starts the TUI.
func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
