package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenMenu screen = iota
	screenBuilder
	screenOutput
)

type model struct {
	screen  screen
	cursor  int
	builder builderModel
	output  string
}

func initialModel() model {
	return model{screen: screenMenu}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case screenMenu:
			return m.updateMenu(msg)
		case screenBuilder:
			return m.updateBuilder(msg)
		case screenOutput:
			return m.updateOutput(msg)
		}
	}
	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(menuLabels)-1 {
			m.cursor++
		}
	case "enter":
		switch menuItem(m.cursor) {
		case menuDeploy:
			m.builder = newDeployBuilder()
			m.screen = screenBuilder
		case menuStatus:
			m.builder = newStatusBuilder()
			m.screen = screenBuilder
		case menuQuit:
			return m, tea.Quit
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateBuilder(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenMenu
	case "enter":
		m.output = m.builder.buildCommand()
		m.screen = screenOutput
	case "tab", "down":
		if m.builder.focus < len(m.builder.fields)-1 {
			m.builder.focus++
		}
	case "shift+tab", "up":
		if m.builder.focus > 0 {
			m.builder.focus--
		}
	case "backspace":
		f := &m.builder.fields[m.builder.focus]
		if len(f.value) > 0 {
			f.value = f.value[:len(f.value)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.builder.fields[m.builder.focus].value += msg.String()
		}
	}
	return m, nil
}

func (m model) updateOutput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "ctrl+c":
		m.screen = screenMenu
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case screenMenu:
		return renderMenu(m.cursor)
	case screenBuilder:
		return m.builder.render()
	case screenOutput:
		return renderOutput(m.output)
	}
	return ""
}

// Run starts the Bubble Tea TUI.
func Run() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
