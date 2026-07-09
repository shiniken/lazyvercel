package ui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shiniken/lazyvercel/internal/vercel"
)

type pane int

const (
	paneProjects pane = iota
	paneDeployments
	paneDetails
)

type Model struct {
	store *vercel.Store

	width  int
	height int
	focus  pane

	projectIndex    int
	deploymentIndex int
	detail          *vercel.DeploymentDetail
	viewport        viewport.Model

	loading bool
	status  string
	err     error
}

type refreshMsg struct {
	err error
}

type detailMsg struct {
	detail vercel.DeploymentDetail
	err    error
}

type openMsg struct {
	err error
}

func NewModel(store *vercel.Store) Model {
	vp := viewport.New(0, 0)
	return Model{
		store:    store,
		focus:    paneDeployments,
		viewport: vp,
		status:   "ready",
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadSelectedDetail()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = max(20, msg.Width/2)
		m.viewport.Height = max(5, msg.Height-6)
		m.syncViewport()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case refreshMsg:
		m.loading = false
		m.err = msg.err
		if msg.err != nil {
			m.status = "refresh failed"
			return m, nil
		}
		m.status = "refreshed " + time.Now().Format("15:04:05")
		m.normalizeIndexes()
		return m, m.loadSelectedDetail()

	case detailMsg:
		m.loading = false
		m.err = msg.err
		if msg.err != nil {
			m.detail = nil
			m.status = "detail failed"
			m.syncViewport()
			return m, nil
		}
		m.detail = &msg.detail
		m.status = "loaded detail"
		m.syncViewport()
		return m, nil

	case openMsg:
		m.loading = false
		m.err = msg.err
		if msg.err != nil {
			m.status = "open failed"
		} else {
			m.status = "opened"
		}
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading lazyvercel..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	bodyHeight := max(5, m.height-lipgloss.Height(header)-lipgloss.Height(footer))

	leftWidth := clamp(24, 34, m.width/4)
	midWidth := clamp(36, 58, m.width/3)
	rightWidth := max(28, m.width-leftWidth-midWidth-4)

	projects := panelStyle(leftWidth, bodyHeight, m.focus == paneProjects).Render(m.renderProjects(leftWidth))
	deployments := panelStyle(midWidth, bodyHeight, m.focus == paneDeployments).Render(m.renderDeployments(midWidth))
	details := panelStyle(rightWidth, bodyHeight, m.focus == paneDetails).Render(m.renderDetails(rightWidth, bodyHeight))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, projects, deployments, details),
		footer,
	)
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Tab):
		m.focus = (m.focus + 1) % 3
		return m, nil
	case key.Matches(msg, keys.BackTab):
		m.focus = (m.focus + 2) % 3
		return m, nil
	case key.Matches(msg, keys.Refresh):
		m.loading = true
		m.status = "refreshing..."
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			return refreshMsg{err: m.store.Refresh(ctx)}
		}
	case key.Matches(msg, keys.Open):
		return m.openCurrent(false)
	case key.Matches(msg, keys.OpenInspector):
		return m.openCurrent(true)
	case key.Matches(msg, keys.Up):
		m.moveSelection(-1)
		return m, m.loadSelectedDetail()
	case key.Matches(msg, keys.Down):
		m.moveSelection(1)
		return m, m.loadSelectedDetail()
	case key.Matches(msg, keys.PageUp):
		if m.focus == paneDetails {
			m.viewport.HalfViewUp()
		} else {
			m.moveSelection(-10)
			return m, m.loadSelectedDetail()
		}
	case key.Matches(msg, keys.PageDown):
		if m.focus == paneDetails {
			m.viewport.HalfViewDown()
		} else {
			m.moveSelection(10)
			return m, m.loadSelectedDetail()
		}
	}
	return m, nil
}

func (m *Model) moveSelection(delta int) {
	switch m.focus {
	case paneProjects:
		projects := m.store.Projects()
		if len(projects) == 0 {
			return
		}
		m.projectIndex = clampIndex(m.projectIndex+delta, len(projects))
		m.deploymentIndex = 0
		m.detail = nil
	case paneDeployments:
		deployments := m.currentDeployments()
		if len(deployments) == 0 {
			return
		}
		m.deploymentIndex = clampIndex(m.deploymentIndex+delta, len(deployments))
		m.detail = nil
	case paneDetails:
		if delta < 0 {
			m.viewport.LineUp(-delta)
		} else {
			m.viewport.LineDown(delta)
		}
	}
}

func (m Model) loadSelectedDetail() tea.Cmd {
	deployment, ok := m.currentDeployment()
	if !ok {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		detail, err := m.store.Detail(ctx, deployment)
		return detailMsg{detail: detail, err: err}
	}
}

func (m Model) openCurrent(inspector bool) (Model, tea.Cmd) {
	deployment, ok := m.currentDeployment()
	if !ok {
		m.status = "nothing to open"
		return m, nil
	}

	target := ""
	if inspector {
		if m.detail != nil {
			target = m.detail.InspectorURL
		}
		if target == "" {
			target = deployment.InspectorURL
		}
	} else if deployment.URL != "" {
		target = "https://" + deployment.URL
	}

	if target == "" {
		m.status = "no url available"
		return m, nil
	}

	m.loading = true
	m.status = "opening..."
	return m, func() tea.Msg {
		return openMsg{err: openURL(target)}
	}
}

func (m *Model) normalizeIndexes() {
	projects := m.store.Projects()
	if len(projects) == 0 {
		m.projectIndex = 0
		m.deploymentIndex = 0
		return
	}
	m.projectIndex = clampIndex(m.projectIndex, len(projects))
	deployments := m.store.Deployments(projects[m.projectIndex])
	if len(deployments) == 0 {
		m.deploymentIndex = 0
		return
	}
	m.deploymentIndex = clampIndex(m.deploymentIndex, len(deployments))
}

func (m Model) currentDeployments() []vercel.Deployment {
	projects := m.store.Projects()
	if len(projects) == 0 || m.projectIndex >= len(projects) {
		return nil
	}
	return m.store.Deployments(projects[m.projectIndex])
}

func (m Model) currentDeployment() (vercel.Deployment, bool) {
	deployments := m.currentDeployments()
	if len(deployments) == 0 || m.deploymentIndex >= len(deployments) {
		return vercel.Deployment{}, false
	}
	return deployments[m.deploymentIndex], true
}

func openURL(target string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", target).Start()
	case "linux":
		return exec.Command("xdg-open", target).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Start()
	default:
		return fmt.Errorf("opening URLs is not supported on %s", runtime.GOOS)
	}
}

func clampIndex(index, length int) int {
	if length <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= length {
		return length - 1
	}
	return index
}

func clamp(minimum, maximum, value int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func trim(value string, width int) string {
	if width <= 1 || len(value) <= width {
		return value
	}
	if width <= 3 {
		return value[:width]
	}
	return value[:width-3] + "..."
}

func fallback(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "-"
}
