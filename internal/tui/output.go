package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	cmdBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(1, 2).
			MarginTop(1)
	hintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).MarginTop(1)
)

func renderOutput(cmd string) string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Generated command") + "\n")
	sb.WriteString(cmdBoxStyle.Render(cmd) + "\n")
	sb.WriteString(hintStyle.Render("Select all and copy, then press esc or q to quit."))
	return sb.String()
}
