package ui

import (
	"fmt"
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

	projects := m.store.Projects()
	deploymentCount := len(m.currentDeployments())
	left := titleStyle.Render("lazyvercel") + " " + subtitleStyle.Render(strings.Join(filterText, " "))
	if len(projects) > 0 {
		left += dimStyle.Render(fmt.Sprintf("  %d project", len(projects)))
		if len(projects) != 1 {
			left += dimStyle.Render("s")
		}
		left += dimStyle.Render(fmt.Sprintf(" / %d deploys", deploymentCount))
	}
	right := mutedStyle.Render(status)
	padding := max(1, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return chromeStyle.Width(m.width).Padding(0, 1).Render(left + strings.Repeat(" ", padding) + right)
}

func (m Model) renderFooter() string {
	refresh := "auto off"
	if m.refreshInterval > 0 {
		effective := m.nextRefreshInterval()
		refresh = "auto " + effective.String()
		if effective != m.refreshInterval {
			refresh += " while active"
		}
	}
	help := "tab focus  j/k move  pg scroll  r refresh  l logs  d detail  o open  i inspect  q quit  " + refresh
	last := m.store.LastRefresh()
	if !last.IsZero() {
		help += "  refreshed " + last.Format("15:04:05")
	}
	return chromeStyle.Width(m.width).Padding(0, 1).Foreground(lipgloss.Color("245")).Render(help)
}

func (m Model) renderProjects(width, height int) string {
	var b strings.Builder
	projects := m.store.Projects()

	b.WriteString(titleStyle.Render("Projects"))
	b.WriteString(" ")
	b.WriteString(dimStyle.Render(fmt.Sprintf("%d", len(projects))))
	b.WriteString("\n\n")

	if len(projects) == 0 {
		b.WriteString(mutedStyle.Render("No projects found"))
		return b.String()
	}

	start, end := visibleRange(len(projects), m.projectIndex, max(1, (height-4)/4))
	for index := start; index < end; index++ {
		project := projects[index]
		name := trim(project.Name, max(8, width-5))
		context := trim(projectContext(project), max(8, width-5))
		summary := m.renderProjectSummary(project, width)
		line := fmt.Sprintf("  %s\n  %s\n  %s", titleStyle.Render(name), mutedStyle.Render(context), summary)
		if index == m.projectIndex {
			line = fmt.Sprintf("> %s\n  %s\n  %s", selectedStyle.Render(name), selectedMutedStyle.Render(context), summary)
		}
		b.WriteString(line)
		if index < end-1 {
			b.WriteString("\n\n")
		}
	}
	if end < len(projects) {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("  +%d more", len(projects)-end)))
	}
	return b.String()
}

func (m Model) renderProjectSummary(project vercel.Project, width int) string {
	deployment, ok := m.store.Summary(project)
	if !ok {
		return dimStyle.Render(trim("deploy -", max(8, width-5)))
	}
	state := deployment.StateLabel()
	target := shortTarget(fallback(deployment.Target, "preview"))
	age := relativeTime(deployment.CreatedAt)
	sha := shortSHA(metaString(deployment.Meta, "githubCommitSha"))
	text := trim(fmt.Sprintf("%s %s %s %s", state, target, age, sha), max(8, width-5))
	return styleState(state).Render(text)
}

func (m Model) renderDeployments(width, height int) string {
	var b strings.Builder
	deployments := m.currentDeployments()
	ready, active, failed := deploymentCounts(deployments)
	b.WriteString(titleStyle.Render("Deployments"))
	b.WriteString(" ")
	b.WriteString(successStyle.Render(fmt.Sprintf("%d ready", ready)))
	if active > 0 {
		b.WriteString(" ")
		b.WriteString(warnStyle.Render(fmt.Sprintf("%d active", active)))
	}
	if failed > 0 {
		b.WriteString(" ")
		b.WriteString(errorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	b.WriteString("\n\n")

	if len(deployments) == 0 {
		b.WriteString(mutedStyle.Render("No deployments found"))
		return b.String()
	}

	contentWidth := max(20, width-4)
	start, end := visibleRange(len(deployments), m.deploymentIndex, max(1, (height-4)/3))
	for index := start; index < end; index++ {
		deployment := deployments[index]
		selected := index == m.deploymentIndex
		stateLabel := deployment.StateLabel()
		state := badgeStyle(stateLabel, selected).Width(12).Render(trim(stateLabel, 10))
		target := fallback(deployment.Target, "preview")
		age := relativeTime(deployment.CreatedAt)
		sha := shortSHA(metaString(deployment.Meta, "githubCommitSha"))
		title := trim(oneLine(fallback(metaString(deployment.Meta, "githubCommitMessage"), deployment.URL, deployment.UID)), max(12, contentWidth-4))
		line := renderDeploymentRow(contentWidth, state, target, age, sha, title, selected)
		b.WriteString(line)
		if index < end-1 {
			b.WriteString("\n")
		}
	}
	if end < len(deployments) {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("  +%d more", len(deployments)-end)))
	}

	return b.String()
}

func (m Model) renderDetails(width, height int) string {
	m.viewport.Width = max(20, width-4)
	m.viewport.Height = max(5, height-2)
	return m.viewport.View()
}

func (m *Model) syncViewport() {
	if m.rightMode == rightModeLogs {
		m.syncLogsViewport()
		return
	}

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
	width := max(40, m.viewport.Width)
	b.WriteString(titleStyle.Render("Detail"))
	b.WriteString(" ")
	b.WriteString(badgeStyle(detail.StateLabel(), false).Render(detail.StateLabel()))
	b.WriteString("\n\n")

	writeSection(&b, "Deployment")
	writeRow(&b, "Project", detail.Project.Name, width)
	writeRow(&b, "Account", fallback(detail.Project.AccountSlug, detail.Project.AccountName, detail.Project.AccountID), width)
	writeRow(&b, "Target", fallback(detail.Target, "preview"), width)
	writeRow(&b, "Created", absoluteTime(detail.CreatedAt)+" ("+relativeTime(detail.CreatedAt)+")", width)
	writeRow(&b, "Building", absoluteTime(detail.BuildingAt), width)
	writeRow(&b, "Ready", absoluteTime(detail.Ready), width)
	writeRow(&b, "Creator", fallback(detail.Creator.Username, detail.Creator.GitHubLogin, detail.Creator.Email), width)
	writeRow(&b, "URL", detail.URL, width)
	writeRow(&b, "Inspect", detail.InspectorURL, width)

	if detail.ErrorCode != "" || detail.ErrorMessage != "" {
		b.WriteString("\n")
		writeSection(&b, "Error")
		writeRow(&b, "Code", detail.ErrorCode, width)
		writeRow(&b, "Message", detail.ErrorMessage, width)
	}

	if len(detail.Alias) > 0 {
		b.WriteString("\n")
		writeSection(&b, "Aliases")
		for _, alias := range detail.Alias {
			b.WriteString("  ")
			b.WriteString(trim("https://"+alias, width-2))
			b.WriteString("\n")
		}
	}

	if len(detail.Meta) > 0 {
		b.WriteString("\n")
		writeSection(&b, "Git")
		writeRow(&b, "Repo", repoName(detail.Meta), width)
		writeRow(&b, "Branch", metaString(detail.Meta, "githubCommitRef"), width)
		writeRow(&b, "Commit", shortSHA(metaString(detail.Meta, "githubCommitSha")), width)
		writeRow(&b, "Author", fallback(metaString(detail.Meta, "githubCommitAuthorName"), metaString(detail.Meta, "githubCommitAuthorLogin"), metaString(detail.Meta, "githubCommitAuthorEmail")), width)
		writeRow(&b, "Message", oneLine(metaString(detail.Meta, "githubCommitMessage")), width)
	}

	if len(detail.Builds) > 0 {
		b.WriteString("\n")
		writeSection(&b, "Builds")
		for _, build := range detail.Builds {
			writeRow(&b, fallback(build.Src, build.Entrypoint, "build"), fallback(build.ReadyState, build.Runtime, build.Use), width)
			if build.ErrorCode != "" || build.ErrorMessage != "" {
				writeRow(&b, "error", strings.TrimSpace(build.ErrorCode+" "+build.ErrorMessage), width)
			}
		}
	}

	if (detail.ProjectSettings != vercel.ProjectSettings{}) {
		b.WriteString("\n")
		writeSection(&b, "Project Settings")
		writeRow(&b, "Framework", detail.ProjectSettings.Framework, width)
		writeRow(&b, "Build", detail.ProjectSettings.BuildCommand, width)
		writeRow(&b, "Install", detail.ProjectSettings.InstallCommand, width)
		writeRow(&b, "Output", detail.ProjectSettings.OutputDirectory, width)
		writeRow(&b, "Root", detail.ProjectSettings.RootDirectory, width)
		writeRow(&b, "Node", detail.ProjectSettings.NodeVersion, width)
	}

	m.viewport.SetContent(strings.TrimRight(b.String(), "\n"))
	m.viewport.GotoTop()
}

func (m *Model) syncLogsViewport() {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Build Logs"))
	b.WriteString("\n\n")

	deployment, ok := m.currentDeployment()
	if !ok {
		b.WriteString(mutedStyle.Render("Select a deployment to inspect logs."))
		m.viewport.SetContent(b.String())
		return
	}

	writeSection(&b, "Deployment")
	writeRow(&b, "State", deployment.StateLabel(), max(40, m.viewport.Width))
	writeRow(&b, "Target", fallback(deployment.Target, "preview"), max(40, m.viewport.Width))
	writeRow(&b, "Commit", shortSHA(metaString(deployment.Meta, "githubCommitSha")), max(40, m.viewport.Width))
	writeRow(&b, "Message", oneLine(metaString(deployment.Meta, "githubCommitMessage")), max(40, m.viewport.Width))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()))
		m.viewport.SetContent(b.String())
		m.viewport.GotoTop()
		return
	}

	if len(m.logLines) == 0 {
		if m.loading {
			b.WriteString(mutedStyle.Render("Loading build logs..."))
		} else {
			b.WriteString(mutedStyle.Render("No build log events returned for this deployment."))
		}
		m.viewport.SetContent(b.String())
		m.viewport.GotoTop()
		return
	}

	writeSection(&b, fmt.Sprintf("Events (%d)", len(m.logLines)))
	width := max(40, m.viewport.Width)
	for _, line := range m.logLines {
		prefix := logPrefix(line)
		text := oneLine(fallback(line.Text, line.Step, line.Entrypoint, line.StatusCode))
		if text == "-" {
			continue
		}
		b.WriteString(dimStyle.Render(prefix))
		b.WriteString(" ")
		if looksLikeFailure(line) {
			b.WriteString(errorStyle.Render(trim(text, width-lipgloss.Width(prefix)-1)))
		} else {
			b.WriteString(trim(text, width-lipgloss.Width(prefix)-1))
		}
		b.WriteString("\n")
	}

	m.viewport.SetContent(strings.TrimRight(b.String(), "\n"))
	m.viewport.GotoTop()
}

func writeSection(b *strings.Builder, title string) {
	b.WriteString(sectionStyle.Render(title))
	b.WriteString("\n")
}

func writeRow(b *strings.Builder, label, value string, width int) {
	if strings.TrimSpace(value) == "" {
		value = "-"
	}
	labelWidth := 10
	valueWidth := max(8, width-labelWidth-2)
	b.WriteString(labelStyle.Render(fmt.Sprintf("%-*s", labelWidth, trim(label, labelWidth))))
	b.WriteString(" ")
	b.WriteString(trim(value, valueWidth))
	b.WriteString("\n")
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

func renderDeploymentRow(width int, state, target, age, sha, title string, selected bool) string {
	marker := "  "
	if selected {
		marker = "> "
	}
	rest := fmt.Sprintf(" %-10s %-9s %s", target, age, sha)
	restWidth := max(0, width-lipgloss.Width(marker)-lipgloss.Width(state))
	firstLine := marker + state + trim(rest, restWidth)
	secondLine := "  " + title
	if selected {
		firstLine = marker + state + selectedStyle.Render(trim(rest, restWidth))
		secondLine = selectedMutedStyle.Render(trim(secondLine, width))
	} else {
		firstLine = marker + state + subtitleStyle.Render(trim(rest, restWidth))
		secondLine = mutedStyle.Render(trim(secondLine, width))
	}
	return firstLine + "\n" + secondLine
}

func deploymentCounts(deployments []vercel.Deployment) (ready int, active int, failed int) {
	for _, deployment := range deployments {
		switch deployment.StateLabel() {
		case "READY":
			ready++
		case "ERROR", "CANCELED":
			failed++
		case "BUILDING", "INITIALIZING", "QUEUED":
			active++
		}
	}
	return ready, active, failed
}

func visibleRange(length, selected, capacity int) (int, int) {
	if length <= 0 {
		return 0, 0
	}
	if capacity <= 0 || capacity >= length {
		return 0, length
	}
	selected = clampIndex(selected, length)
	start := selected - capacity/2
	if start < 0 {
		start = 0
	}
	if start+capacity > length {
		start = length - capacity
	}
	return start, start + capacity
}

func shortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if sha == "" || sha == "<nil>" {
		return "-"
	}
	if len(sha) > 10 {
		return sha[:10]
	}
	return sha
}

func shortTarget(target string) string {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "production":
		return "prod"
	case "preview":
		return "prev"
	case "":
		return "-"
	default:
		return target
	}
}

func repoName(meta map[string]any) string {
	org := metaString(meta, "githubOrg")
	repo := metaString(meta, "githubRepo")
	switch {
	case org != "" && repo != "":
		return org + "/" + repo
	case repo != "":
		return repo
	case org != "":
		return org
	default:
		return ""
	}
}

func projectContext(project vercel.Project) string {
	prefix := ""
	switch {
	case project.LinkedCWD:
		prefix = "cwd "
	case project.Pinned:
		prefix = "pin "
	}
	account := fallback(project.AccountSlug, project.AccountName, project.AccountID)
	if project.Link.Repo != "" {
		repo := project.Link.Repo
		if project.Link.Org != "" {
			repo = project.Link.Org + "/" + repo
		}
		return prefix + repo
	}
	if account != "-" {
		return prefix + account
	}
	if project.LinkedDir != "" {
		return prefix + project.LinkedDir
	}
	return prefix + "-"
}

func logPrefix(line vercel.BuildLogLine) string {
	parts := []string{}
	if formatted := logTime(line.CreatedAt); formatted != "" {
		parts = append(parts, formatted)
	}
	if line.Type != "" {
		parts = append(parts, trim(line.Type, 8))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " ")
}

func logTime(value int64) string {
	if value <= 0 {
		return ""
	}
	if value > 1_000_000_000_000 {
		return time.UnixMilli(value).Format("15:04:05")
	}
	return time.Unix(value, 0).Format("15:04:05")
}

func looksLikeFailure(line vercel.BuildLogLine) bool {
	text := strings.ToLower(line.Text + " " + line.Step + " " + line.StatusCode)
	return strings.EqualFold(line.Type, "stderr") ||
		strings.Contains(text, "error") ||
		strings.Contains(text, "failed") ||
		strings.Contains(text, "exception") ||
		strings.HasPrefix(line.StatusCode, "4") ||
		strings.HasPrefix(line.StatusCode, "5")
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
