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

type sessionsState struct {
	selected      int
	showArchived  bool
	focusBranches bool
}

func (a *App) visibleSessions() []forge.Session {
	if a.sessionsState.showArchived {
		return a.sessions
	}

	out := make([]forge.Session, 0, len(a.sessions))
	for _, s := range a.sessions {
		if !s.Archived {
			out = append(out, s)
		}
	}

	return out
}

func (a *App) updateSessions(m tea.KeyMsg) {
	if a.sessionsState.focusBranches {
		switch {
		case key.Matches(m, a.keys.Tab), key.Matches(m, a.keys.Esc):
			a.sessionsState.focusBranches = false

		case key.Matches(m, a.keys.Left):
			if a.branchesState.pane > 0 {
				a.branchesState.pane--
			} else {
				a.sessionsState.focusBranches = false
			}

		default:
			a.updateBranches(m)
		}

		return
	}

	switch {
	case key.Matches(m, a.keys.Up):
		if a.sessionsState.selected > 0 {
			a.sessionsState.selected--
			a.reloadRefsOnly()
		}

	case key.Matches(m, a.keys.Down):
		visible := a.visibleSessions()
		if a.sessionsState.selected < len(visible)-1 {
			a.sessionsState.selected++
			a.reloadRefsOnly()
		}

	case key.Matches(m, a.keys.Tab):
		a.sessionsState.focusBranches = true
		a.branchesState.pane = 0

	case key.Matches(m, a.keys.Enter):
		a.reloadLogRefs()
		a.screen = ScreenLog

	case key.Matches(m, a.keys.Archive):
		visible := a.visibleSessions()
		if len(visible) > 0 && a.sessionsState.selected < len(visible) {
			ss := visible[a.sessionsState.selected]
			_ = a.client.ArchiveSession(context.Background(), ss.Name)
			a.reloadAll()
			a.flash("forge sessions archive " + ss.Name)
		}

	case key.Matches(m, a.keys.New):
		_, _ = a.client.NewSession(context.Background(), "new-session", "")
		a.reloadAll()
		a.flash("forge sessions new")

	case key.Matches(m, a.keys.Clone):
		ss := a.activeSession()
		_, _ = a.client.CloneSession(context.Background(), ss.Name, ss.Name+"-clone")
		a.reloadAll()
		a.flash("forge sessions clone " + ss.Name)

	case key.Matches(m, a.keys.Delete):
		ss := a.activeSession()
		_ = a.client.DeleteSession(context.Background(), ss.Name)
		a.reloadAll()
		a.flash("forge sessions delete " + ss.Name)

	case key.Matches(m, a.keys.Filter):
		a.sessionsState.showArchived = !a.sessionsState.showArchived
		a.sessionsState.selected = 0
	}
}

func (a *App) reloadRefsOnly() {
	visible := a.visibleSessions()
	if len(visible) == 0 {
		return
	}

	idx := a.sessionsState.selected
	if idx >= len(visible) {
		idx = 0
	}

	id := visible[idx].ID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if rs, err := a.client.Refs(ctx, id); err == nil {
		a.refs = rs
		a.branchesState.selectedRef = 0
	}
}

func (a *App) reloadLogRefs() {
	visible := a.visibleSessions()
	if len(visible) == 0 {
		return
	}

	idx := a.sessionsState.selected
	if idx >= len(visible) {
		idx = 0
	}

	id := visible[idx].ID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ms, err := a.client.Log(ctx, id); err == nil {
		a.messages = ms
		a.logState.selected = 0
		for i := range a.sessions {
			if a.sessions[i].ID == id {
				a.sessions[i].Messages, a.sessions[i].Counts = countMessages(ms)
				break
			}
		}
	}

	if rs, err := a.client.Refs(ctx, id); err == nil {
		a.refs = rs
		a.branchesState.selectedRef = 0
	}
}

func (a *App) viewSessions(w, h int) string {
	s := a.styles
	visible := a.visibleSessions()

	if len(visible) > 0 && a.sessionsState.selected >= len(visible) {
		a.sessionsState.selected = len(visible) - 1
	}

	topH := max(h*55/100, 8)
	botH := max(h-topH-1, 6)

	leftW := w * 70 / 100
	rightW := w - leftW - 2

	// sessions list
	archTag := ""
	if a.sessionsState.showArchived {
		archTag = " " + s.ChipAcc.Render("all")
	}

	head := s.Header.Render(fmt.Sprintf("%-12s %-14s %-22s %-10s %5s  %s", "ID", "NAME", "TITLE", "PLUGINS", "MSGS", "UPDATED"))
	rows := []string{head, Hr(leftW - 4)}
	for i, ss := range visible {
		archMark := ""
		if ss.Archived {
			archMark = " " + s.Chip.Render("arch")
		}

		line := fmt.Sprintf("%-12s %-14s %-22s %-10s %5d  %s%s",
			shortHash(ss.ID, 12),
			Truncate(ss.Name, 14),
			Truncate(ss.Title, 22),
			Truncate(ss.Plugins, 10),
			ss.Messages,
			ss.Updated.Format("2006-01-02 15:04"),
			archMark,
		)

		st := s.Row
		if i == a.sessionsState.selected {
			st = s.RowSel
		}

		rows = append(rows, st.Width(leftW-4).Render(line))
	}

	tableTitle := fmt.Sprintf("sessions [%d]%s", len(visible), archTag)
	table := Box(s, tableTitle, !a.sessionsState.focusBranches, leftW, strings.Join(rows, "\n"))
	table = fitLines(table, topH)

	// session detail
	var ss forge.Session
	if len(visible) > 0 && a.sessionsState.selected < len(visible) {
		ss = visible[a.sessionsState.selected]
	}

	maxCount := 1
	for _, n := range ss.Counts {
		if n > maxCount {
			maxCount = n
		}
	}

	detail := strings.Builder{}
	kvw := 11

	detail.WriteString(KV(s, "ID", s.Muted.Render(shortHash(ss.ID, 16)), kvw) + "\n")
	detail.WriteString(KV(s, "Name", s.Accent.Render(ss.Name), kvw) + "\n")
	detail.WriteString(KV(s, "Title", Truncate(ss.Title, rightW-14), kvw) + "\n")
	detail.WriteString(KV(s, "Model", s.Muted.Render(ss.Model), kvw) + "\n")
	detail.WriteString(KV(s, "Parent", s.Info.Render(ss.Parent), kvw) + "\n")
	detail.WriteString(KV(s, "Created", s.Muted.Render(ss.Created.Format("01-02 15:04")), kvw) + "\n")
	detail.WriteString(KV(s, "Updated", s.Muted.Render(ss.Updated.Format("01-02 15:04")), kvw) + "\n")

	detail.WriteString("\n")
	detail.WriteString(s.Faint.Render(fmt.Sprintf("MESSAGES · %d", ss.Messages)) + "\n")
	for _, role := range []string{"user", "assistant", "tool_call", "tool_result"} {
		n := ss.Counts[role]
		bar := miniBar(n, maxCount, 40, AccentAmber)
		line := fmt.Sprintf("%s %-11s  %s %3d",
			s.RoleStyle(role).Render(RoleGlyph(role)),
			s.RoleStyle(role).Render(fmt.Sprintf("%-11s", role)),
			bar, n,
		)

		detail.WriteString(line + "\n")
	}

	detail.WriteString("\n")
	detail.WriteString(s.Faint.Render("COST") + "\n")
	detail.WriteString(KV(s, "Estimated", s.Accent.Render(fmt.Sprintf("$%.4f", ss.Cost)), kvw) + "\n")
	detail.WriteString(KV(s, "Tokens", s.Muted.Render(fmt.Sprintf("in=%d out=%d", ss.TokensIn, ss.TokensOut)), kvw) + "\n")

	rightBox := Box(s, ss.Name, false, rightW, detail.String())
	rightBox = fitLines(rightBox, topH)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, table, "  ", rightBox)

	// bottom: branches panel
	botRow := a.viewBranches(w, a.sessionsState.focusBranches)
	botRow = fitLines(botRow, botH)

	return topRow + "\n" + botRow
}

func miniBar(n, maxVal, width int, accent lipgloss.Color) string {
	if maxVal <= 0 {
		maxVal = 1
	}

	cells := min((n*width)/maxVal, width)
	on := lipgloss.NewStyle().Foreground(accent).Render(strings.Repeat("█", cells))
	off := lipgloss.NewStyle().Foreground(ColRule2).Render(strings.Repeat("░", width-cells))

	return on + off
}

func shortHash(h string, n int) string {
	if len(h) <= n {
		return h
	}

	return h[:n]
}
