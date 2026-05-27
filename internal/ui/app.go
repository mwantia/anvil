package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mwantia/anvil/internal/forge"
)

type Screen int

const (
	ScreenSessions  Screen = iota // tab 1 — expandable session tree
	ScreenResources               // tab 2
	ScreenSystem                  // tab 3
	ScreenLog                     // sub-screen, entered via Enter on a session
)

var screenNames = []string{"sessions", "resources", "system"}
var screenLabels = []string{
	"forge sessions",
	"forge resources status",
	"forge system status",
	"forge sessions log",
}

// App is the root Bubble Tea model.
type App struct {
	client   forge.Client
	keys     KeyMap
	styles   Styles
	themeIdx int

	screen      Screen
	width       int
	height      int
	now         time.Time
	statusFlash string
	clearAt     time.Time

	sessions  []forge.Session
	messages  []forge.Message
	refs      []forge.Ref
	resources []forge.Resource
	system    forge.System

	sessionsState  sessionsState
	logState       logState
	resourcesState resourcesState
}

func NewApp(client forge.Client) *App {
	app := &App{
		client: client,
		keys:   DefaultKeys(),
		styles: NewStyles(Themes[0]),
		now:    time.Now(),
	}
	app.sessionsState.expanded = map[string]bool{}
	app.sessionsState.sessionRefs = map[string][]forge.Ref{}
	app.logState.treeExp = map[int]bool{}
	app.reloadAll()

	return app
}

func (a *App) reloadAll() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s, err := a.client.Sessions(ctx); err == nil {
		a.sessions = s
		for _, ss := range s {
			if _, exists := a.sessionsState.expanded[ss.ID]; !exists {
				a.sessionsState.expanded[ss.ID] = true
			}
		}
	}

	// Refresh refs for any currently expanded sessions.
	visible := a.visibleSessions()
	for _, ss := range visible {
		if a.sessionsState.expanded[ss.ID] {
			if rs, err := a.client.Refs(ctx, ss.ID); err == nil {
				a.sessionsState.sessionRefs[ss.ID] = rs
			}
		}
	}

	// Sync log + refs for the active session.
	if active := a.activeSession(); active.ID != "" {
		if ms, err := a.client.Log(ctx, active.ID, ""); err == nil {
			a.messages = ms
			for i := range a.sessions {
				if a.sessions[i].ID == active.ID {
					a.sessions[i].Messages, a.sessions[i].Counts = countMessages(ms)
					break
				}
			}
		}
		if rs, ok := a.sessionsState.sessionRefs[active.ID]; ok {
			a.refs = rs
		}
	}

	if rs, err := a.client.Resources(ctx, "all"); err == nil {
		a.resources = rs
	}

	if sys, err := a.client.System(ctx); err == nil {
		a.system = sys
	}
}

func countMessages(msgs []forge.Message) (int, map[string]int) {
	counts := map[string]int{}
	for _, m := range msgs {
		counts[string(m.Role)]++
	}

	return len(msgs), counts
}

func (a *App) headRefLabel() string {
	for _, r := range a.refs {
		if r.IsHead {
			return r.Target
		}
	}

	return ""
}

func (a *App) flash(msg string) {
	a.statusFlash = msg
	a.clearAt = time.Now().Add(2 * time.Second)
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (a *App) Init() tea.Cmd {
	return tick()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height
		return a, nil

	case tickMsg:
		a.now = time.Time(m)
		if !a.clearAt.IsZero() && a.now.After(a.clearAt) {
			a.statusFlash = ""
			a.clearAt = time.Time{}
		}
		return a, tick()

	case tea.KeyMsg:
		if key.Matches(m, a.keys.Quit) {
			return a, tea.Quit
		}
		if key.Matches(m, a.keys.Esc) && a.screen == ScreenLog {
			a.screen = ScreenSessions
			return a, nil
		}
		if m.String() == "t" {
			a.themeIdx = (a.themeIdx + 1) % len(Themes)
			a.styles = NewStyles(Themes[a.themeIdx])
			a.flash("theme: " + Themes[a.themeIdx].Name)
			return a, nil
		}
		switch {
		case key.Matches(m, a.keys.Tab1):
			a.screen = ScreenSessions
			return a, nil
		case key.Matches(m, a.keys.Tab2):
			a.screen = ScreenResources
			return a, nil
		case key.Matches(m, a.keys.Tab3):
			a.screen = ScreenSystem
			return a, nil
		}

		switch a.screen {
		case ScreenSessions:
			a.updateSessions(m)
		case ScreenLog:
			a.updateLog(m)
		case ScreenResources:
			a.updateResources(m)
		}
	}

	return a, nil
}

func (a *App) View() string {
	if a.width == 0 {
		a.width = 140
	}
	// Log is a sub-screen of sessions; keep sessions tab highlighted.
	tabIdx := int(a.screen)
	if a.screen == ScreenLog {
		tabIdx = int(ScreenSessions)
	}

	header := TermBar(a.styles, a.width, screenLabels[a.screen])

	health := a.styles.Faint.Render("○")
	if a.system.Agent.HTTP != "" {
		health = a.styles.OK.Render("●")
	}
	tabs := TabBar(a.styles, a.width, tabIdx, screenNames, a.system.Agent.HTTP, health)

	hints := KeyHints(a.styles, a.screenHints())

	left := []string{
		fmt.Sprintf("anvil %s", a.system.Agent.Version),
		fmt.Sprintf("session %s", a.activeSession().Name),
		fmt.Sprintf("HEAD %s", a.headRefLabel()),
	}
	right := []string{a.now.Format("15:04:05")}
	if a.statusFlash != "" {
		left = append([]string{a.statusFlash}, left...)
	}
	status := StatusBar(a.styles, a.width, left, right)

	chromeH := lipgloss.Height(header) + lipgloss.Height(tabs) +
		lipgloss.Height(hints) + lipgloss.Height(status) + 2
	bodyH := max(a.height-chromeH, 1)

	var body string
	switch a.screen {
	case ScreenSessions:
		body = a.viewSessions(a.width, bodyH)
	case ScreenLog:
		body = a.viewLog(a.width, bodyH)
	case ScreenResources:
		body = a.viewResources(a.width)
	case ScreenSystem:
		body = a.viewSystem(a.width)
	}

	body = fitLines(body, bodyH)

	return a.styles.App.Render(
		header + "\n" + tabs + "\n\n" + body + "\n\n" + hints + "\n" + status,
	)
}

// fitLines clips s to at most n lines, and pads with blank lines if shorter.
func fitLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		return strings.Join(lines[:n], "\n")
	}

	if len(lines) < n {
		return s + strings.Repeat("\n", n-len(lines))
	}

	return s
}

func (a *App) screenHints() [][2]string {
	common := [][2]string{{"1-3", "tab"}, {"t", "theme"}, {"q", "quit"}}
	var hints [][2]string

	switch a.screen {
	case ScreenSessions:
		hints = [][2]string{
			{"↑↓", "select"}, {"→/d", "expand"}, {"←", "collapse"},
			{"K/J", "all"}, {"enter", "log"},
			{"n", "new"}, {"c", "clone"}, {"a", "archive"}, {"x", "delete"},
		}

	case ScreenLog:
		hints = [][2]string{
			{"↑↓", "walk"}, {"→/d", "calls"}, {"←", "collapse"}, {"enter", "body"},
			{"e", "edit·fork"}, {"c", "checkout"}, {"y", "yank"}, {"⌫", "back"},
		}

	case ScreenResources:
		hints = [][2]string{{"↑↓", "select"}, {"←→", "scope"}, {"y", "yank"}}

	case ScreenSystem:
		hints = [][2]string{{"r", "reload"}}
	}

	return append(hints, common...)
}
