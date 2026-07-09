package ui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/shiniken/lazyvercel/internal/vercel"
)

func (m Model) renderHeader() string {
	filters := m.store.Filters()
	filterText := []string{}
	if filters.Target != "" {
		filterText = append(filterText, "target="+filters.Target)
	}
	if filters.Branch != "" {
		filterText = append(filterText, "branch="+filters.Branch)
	}
	if len(filterText) == 0 {
		filterText = append(filterText, "all deployments")
	}

	status := m.status
	if m.loading {
		status = "loading..."
	}
	if m.err != nil {
		status = errorStyle.Render(m.err.Error())
	}

	left := titleStyle.Render("lazyvercel") + " " + mutedStyle.Render(strings.Join(filterText, " "))
	right := mutedStyle.Render(status)
	padding := max(1, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return baseStyle.Width(m.width).Render(left + strings.Repeat(" ", padding) + right)
}

func (m Model) renderFooter() string {
	help := "tab focus  j/k move  r refresh  o open deployment  i inspect  q quit"
	last := m.store.LastRefresh()
	if !last.IsZero() {
		help += "  refreshed " + last.Format("15:04:05")
	}
	return baseStyle.Width(m.width).Foreground(lipgloss.Color("245")).Render(help)
}

func (m Model) renderProjects(width int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Projects"))
	b.WriteString("\n\n")

	projects := m.store.Projects()
	if len(projects) == 0 {
		b.WriteString(mutedStyle.Render("No linked projects"))
		return b.String()
	}

	for index, project := range projects {
		line := fmt.Sprintf("%s\n%s", project.ProjectName, mutedStyle.Render(filepath.Base(project.Dir)))
		if index == m.projectIndex {
			line = selectedStyle.Width(width - 4).Render(line)
		}
		b.WriteString(line)
		if index < len(projects)-1 {
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

func (m Model) renderDeployments(width int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Deployments"))
	b.WriteString("\n\n")

	deployments := m.currentDeployments()
	if len(deployments) == 0 {
		b.WriteString(mutedStyle.Render("No deployments found"))
		return b.String()
	}

	contentWidth := max(20, width-4)
	for index, deployment := range deployments {
		state := styleState(deployment.StateLabel()).Render(padRight(deployment.StateLabel(), 12))
		target := mutedStyle.Render(padRight(fallback(deployment.Target, "preview"), 10))
		age := mutedStyle.Render(relativeTime(deployment.CreatedAt))
		title := trim(oneLine(fallback(metaString(deployment.Meta, "githubCommitMessage"), deployment.URL, deployment.UID)), max(12, contentWidth-26))

		line := fmt.Sprintf("%s %s %s\n%s", state, target, age, mutedStyle.Render(title))
		if index == m.deploymentIndex {
			line = selectedStyle.Width(contentWidth).Render(line)
		}
		b.WriteString(line)
		if index < len(deployments)-1 {
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

func (m Model) renderDetails(width, height int) string {
	m.viewport.Width = max(20, width-4)
	m.viewport.Height = max(5, height-2)
	return m.viewport.View()
}

func (m *Model) syncViewport() {
	var b strings.Builder
	if m.detail == nil {
		if m.err != nil {
			b.WriteString(errorStyle.Render(m.err.Error()))
		} else {
			b.WriteString(mutedStyle.Render("Select a deployment to inspect it."))
		}
		m.viewport.SetContent(b.String())
		return
	}

	detail := *m.detail
	b.WriteString(titleStyle.Render("Detail"))
	b.WriteString("\n\n")
	writeRow(&b, "Project", detail.Project.ProjectName)
	writeRow(&b, "State", detail.StateLabel())
	writeRow(&b, "Target", fallback(detail.Target, "preview"))
	writeRow(&b, "URL", detail.URL)
	writeRow(&b, "Inspector", detail.InspectorURL)
	writeRow(&b, "Created", absoluteTime(detail.CreatedAt))
	writeRow(&b, "Building", absoluteTime(detail.BuildingAt))
	writeRow(&b, "Ready", absoluteTime(detail.Ready))
	writeRow(&b, "Creator", fallback(detail.Creator.Username, detail.Creator.GitHubLogin, detail.Creator.Email))

	if detail.ErrorCode != "" || detail.ErrorMessage != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error"))
		b.WriteString("\n")
		writeRow(&b, "Code", detail.ErrorCode)
		writeRow(&b, "Message", detail.ErrorMessage)
	}

	if len(detail.Alias) > 0 {
		b.WriteString("\n")
		b.WriteString(titleStyle.Render("Aliases"))
		b.WriteString("\n")
		for _, alias := range detail.Alias {
			b.WriteString("  https://")
			b.WriteString(alias)
			b.WriteString("\n")
		}
	}

	if len(detail.Meta) > 0 {
		b.WriteString("\n")
		b.WriteString(titleStyle.Render("Git Metadata"))
		b.WriteString("\n")
		writeMeta(&b, detail.Meta)
	}

	if len(detail.Builds) > 0 {
		b.WriteString("\n")
		b.WriteString(titleStyle.Render("Builds"))
		b.WriteString("\n")
		for _, build := range detail.Builds {
			writeRow(&b, fallback(build.Src, build.Entrypoint, "build"), fallback(build.ReadyState, build.Runtime, build.Use))
			if build.ErrorCode != "" || build.ErrorMessage != "" {
				writeRow(&b, "  error", strings.TrimSpace(build.ErrorCode+" "+build.ErrorMessage))
			}
		}
	}

	if (detail.ProjectSettings != vercel.ProjectSettings{}) {
		b.WriteString("\n")
		b.WriteString(titleStyle.Render("Project Settings"))
		b.WriteString("\n")
		writeRow(&b, "Framework", detail.ProjectSettings.Framework)
		writeRow(&b, "Build", detail.ProjectSettings.BuildCommand)
		writeRow(&b, "Install", detail.ProjectSettings.InstallCommand)
		writeRow(&b, "Output", detail.ProjectSettings.OutputDirectory)
		writeRow(&b, "Root", detail.ProjectSettings.RootDirectory)
		writeRow(&b, "Node", detail.ProjectSettings.NodeVersion)
	}

	m.viewport.SetContent(strings.TrimRight(b.String(), "\n"))
	m.viewport.GotoTop()
}

func writeRow(b *strings.Builder, label, value string) {
	if strings.TrimSpace(value) == "" {
		value = "-"
	}
	b.WriteString(mutedStyle.Render(fmt.Sprintf("%-10s", label)))
	b.WriteString(" ")
	b.WriteString(value)
	b.WriteString("\n")
}

func writeMeta(b *strings.Builder, meta map[string]any) {
	keys := make([]string, 0, len(meta))
	for key := range meta {
		if strings.HasPrefix(key, "github") || strings.HasPrefix(key, "gitlab") || strings.HasPrefix(key, "git") {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		for key := range meta {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		writeRow(b, key, fmt.Sprint(meta[key]))
	}
}

func metaString(meta map[string]any, key string) string {
	if meta == nil {
		return ""
	}
	value, ok := meta[key]
	if !ok || value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func oneLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func styleState(state string) lipgloss.Style {
	switch state {
	case "READY":
		return successStyle
	case "ERROR", "CANCELED":
		return errorStyle
	case "BUILDING", "INITIALIZING", "QUEUED":
		return warnStyle
	default:
		return mutedStyle
	}
}

func relativeTime(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	t := time.UnixMilli(ms)
	d := time.Since(t)
	if d < time.Minute {
		return "now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

func absoluteTime(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	return time.UnixMilli(ms).Format("2006-01-02 15:04:05")
}

func padRight(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return value + strings.Repeat(" ", width-len(value))
}
