package tui

import "github.com/charmbracelet/lipgloss"

type menuItem int

const (
	menuDeploy menuItem = iota
	menuStatus
	menuQuit
)

var menuLabels = []string{
	"Build deploy command",
	"Build status command",
	"Quit",
}

var (
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).MarginBottom(1)
)

func renderMenu(cursor int) string {
	s := titleStyle.Render("oxctl — ECS deployment tool") + "\n"
	for i, label := range menuLabels {
		if i == cursor {
			s += selectedStyle.Render("▶ " + label)
		} else {
			s += normalStyle.Render("  " + label)
		}
		s += "\n"
	}
	s += normalStyle.Render("\n↑/↓ navigate · enter select · q quit")
	return s
}
