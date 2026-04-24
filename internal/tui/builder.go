package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type field struct {
	label       string
	placeholder string
	value       string
}

type builderModel struct {
	fields    []field
	focus     int
	forDeploy bool
}

func newDeployBuilder() builderModel {
	return builderModel{
		forDeploy: true,
		fields: []field{
			{label: "cluster", placeholder: "my-cluster"},
			{label: "service", placeholder: "my-service"},
			{label: "image", placeholder: "123.dkr.ecr.us-east-1.amazonaws.com/app:sha"},
			{label: "container-name", placeholder: "app"},
			{label: "task-def", placeholder: "./task-def.json"},
			{label: "wait (true/false)", placeholder: "false"},
			{label: "timeout (seconds)", placeholder: "300"},
		},
	}
}

func newStatusBuilder() builderModel {
	return builderModel{
		forDeploy: false,
		fields: []field{
			{label: "cluster", placeholder: "my-cluster"},
			{label: "service", placeholder: "my-service"},
		},
	}
}

var (
	focusedField = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("205")).Padding(0, 1)
	blurredField = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	labelStyle   = lipgloss.NewStyle().Width(22).Foreground(lipgloss.Color("245"))
)

func (b builderModel) render() string {
	var sb strings.Builder
	if b.forDeploy {
		sb.WriteString(titleStyle.Render("Deploy command builder") + "\n\n")
	} else {
		sb.WriteString(titleStyle.Render("Status command builder") + "\n\n")
	}
	for i, f := range b.fields {
		disp := f.value
		if disp == "" {
			disp = f.placeholder
		}
		row := labelStyle.Render(f.label+":") + " "
		if i == b.focus {
			row += focusedField.Render(disp)
		} else {
			row += blurredField.Render(disp)
		}
		sb.WriteString(row + "\n")
	}
	sb.WriteString(normalStyle.Render("\ntype to edit · tab/↓ next · shift+tab/↑ prev · enter generate · esc back"))
	return sb.String()
}

func (b builderModel) buildCommand() string {
	labels := []string{"cluster", "service", "image", "container-name", "task-def", "wait (true/false)", "timeout (seconds)"}
	flags := []string{"cluster", "service", "image", "container-name", "task-def", "wait", "timeout"}
	vals := make(map[string]string, len(labels))
	for i, f := range b.fields {
		v := f.value
		if v == "" {
			v = f.placeholder
		}
		vals[labels[i]] = v
	}
	if b.forDeploy {
		parts := []string{"oxctl deploy"}
		for i, lbl := range labels {
			v := vals[lbl]
			if flags[i] == "wait" {
				if v == "true" {
					parts = append(parts, "--wait")
				}
				continue
			}
			if flags[i] == "timeout" && vals["wait (true/false)"] != "true" {
				continue
			}
			parts = append(parts, fmt.Sprintf("--%s %s", flags[i], v))
		}
		return strings.Join(parts, " \\\n  ")
	}
	return fmt.Sprintf("oxctl status \\\n  --cluster %s \\\n  --service %s",
		vals["cluster"], vals["service"])
}
