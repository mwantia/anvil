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

type logState struct {
	selected int
	expanded bool
}

func (a *App) updateLog(m tea.KeyMsg) {
	switch {
	case key.Matches(m, a.keys.Up):
		if a.logState.selected > 0 {
			a.logState.selected--
		}

	case key.Matches(m, a.keys.Down):
		if a.logState.selected < len(a.messages)-1 {
			a.logState.selected++
		}

	case key.Matches(m, a.keys.Enter):
		a.logState.expanded = !a.logState.expanded

	case key.Matches(m, a.keys.Edit):
		msg := a.currentMessage()
		_, _ = a.client.EditFork(context.Background(), a.activeSession().ID, msg.Hash, "")
		a.reloadLogRefs()
		a.flash("forge sessions edit " + a.activeSession().Name + " " + shortHash(msg.Hash, 8))

	case key.Matches(m, a.keys.Clone):
		msg := a.currentMessage()
		_ = a.client.Checkout(context.Background(), a.activeSession().ID, msg.Hash)
		a.reloadLogRefs()
		a.flash("forge sessions checkout " + a.activeSession().Name + " " + shortHash(msg.Hash, 8))

	case key.Matches(m, a.keys.Yank):
		a.flash("yanked " + shortHash(a.currentMessage().Hash, 12))
	}
}

func (a *App) currentMessage() forge.Message {
	if len(a.messages) == 0 {
		return forge.Message{}
	}

	if a.logState.selected >= len(a.messages) {
		a.logState.selected = len(a.messages) - 1
	}

	return a.messages[a.logState.selected]
}

func (a *App) viewLog(w, h int) string {
	s := a.styles
	ss := a.activeSession()

	header := strings.Builder{}
	header.WriteString(s.Accent.Render(ss.Name) + s.Faint.Render(" · ") + s.Muted.Render(ss.Model))
	if a.headRefLabel() != "" {
		header.WriteString(s.Faint.Render(" @ ") + s.Accent.Render(a.headRefLabel()))
	}

	header.WriteString("\n")
	header.WriteString(s.Muted.Render(fmt.Sprintf("%d messages", ss.Messages)) + s.Faint.Render(fmt.Sprintf(
		"  user=%d  assistant=%d  tool_call=%d  tool_result=%d",
		ss.Counts["user"], ss.Counts["assistant"], ss.Counts["tool_call"], ss.Counts["tool_result"])) + "\n")
	header.WriteString(s.Faint.Render("tokens: ") + s.Muted.Render(fmt.Sprintf(
		"in=%d  out=%d  total=%d  cost=$%.6f",
		ss.TokensIn, ss.TokensOut, ss.TokensIn+ss.TokensOut, ss.Cost)))
	headerBox := Box(s, "log · "+ss.Name, false, w-2, header.String())

	headerH := lipgloss.Height(headerBox)
	bodyH := max(h-headerH-1, 4)

	left := (w - 4) / 2
	right := w - left - 4

	listRows := make([]string, 0, len(a.messages))
	for i, msg := range a.messages {
		listRows = append(listRows, a.renderLogRow(msg, left-4, i == a.logState.selected))
	}

	listContent := fitLines(strings.Join(listRows, "\n"), bodyH-3)
	list := Box(s, fmt.Sprintf("messages [%d]", len(a.messages)), true, left, listContent)
	list = fitLines(list, bodyH)

	msg := a.currentMessage()
	detailContent := fitLines(a.renderMessageDetail(msg), bodyH-3)
	detail := Box(s, shortHash(msg.Hash, 12), false, right, detailContent)
	detail = fitLines(detail, bodyH)

	body := lipgloss.JoinHorizontal(lipgloss.Top, list, "  ", detail)

	return headerBox + "\n" + body
}

func (a *App) renderLogRow(msg forge.Message, width int, selected bool) string {
	s := a.styles
	role := string(msg.Role)
	refsStr := ""
	for _, r := range msg.Refs {
		refsStr += s.ChipAcc.Render(r) + " "
	}

	previewWidth := max(width-28-lipgloss.Width(refsStr), 10)
	line := fmt.Sprintf("%s %s %s %s",
		s.Faint.Render(shortHash(msg.Hash, 8)),
		s.RoleStyle(role).Render(RoleGlyph(role)),
		s.RoleStyle(role).Render(fmt.Sprintf("%-11s", role)),
		Truncate(msg.Preview, previewWidth),
	)

	if refsStr != "" {
		line = fmt.Sprintf("%s (%s)", line, refsStr)
	}

	st := s.Row
	if selected {
		st = s.RowSel
	}

	return st.Width(width).Render(line)
}

func (a *App) renderMessageDetail(msg forge.Message) string {
	s := a.styles
	role := string(msg.Role)

	b := strings.Builder{}
	b.WriteString(s.Faint.Render("message ") + s.Muted.Render(msg.Hash))
	for _, r := range msg.Refs {
		b.WriteString(" " + s.ChipAcc.Render(r))
	}

	b.WriteString("\n")
	b.WriteString(KV(s, "Role", s.RoleStyle(role).Render(RoleGlyph(role)+" "+role), 8) + "\n")
	b.WriteString(KV(s, "Date", s.Muted.Render(formatDate(msg.Date)), 8) + "\n")
	if msg.TokIn > 0 || msg.TokOut > 0 {
		b.WriteString(KV(s, "Tokens", s.Muted.Render(fmt.Sprintf("in=%d out=%d", msg.TokIn, msg.TokOut)), 8) + "\n")
	}

	b.WriteString("\n")
	if a.logState.expanded && len(msg.Body) > 0 {
		b.WriteString(strings.Join(msg.Body, "\n"))
	} else {
		b.WriteString(s.Muted.Render(msg.Preview))
		if len(msg.Body) > 1 {
			b.WriteString("\n\n" + s.Faint.Render(fmt.Sprintf("(enter to expand · %d lines)", len(msg.Body))))
		}
	}

	return b.String()
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "—"
	}

	return t.Format("Mon Jan 2 15:04:05 2006")
}
