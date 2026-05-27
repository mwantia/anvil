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

// treeItem identifies a single row in the expandable session tree.
// refIdx == -1 means this is a session header row; >= 0 is a ref row.
type treeItem struct {
	sessionIdx int
	refIdx     int
}

type sessionsState struct {
	cursor       int
	offset       int // viewport scroll offset
	showArchived bool
	expanded     map[string]bool        // session ID → expanded
	sessionRefs  map[string][]forge.Ref // session ID → refs (lazy-loaded on expand)
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

// treeItems builds the flat, ordered list of rows from visible sessions
// and any that are currently expanded.
func (a *App) treeItems() []treeItem {
	visible := a.visibleSessions()
	var items []treeItem
	for i, ss := range visible {
		items = append(items, treeItem{sessionIdx: i, refIdx: -1})
		if a.sessionsState.expanded[ss.ID] {
			for j := range a.sessionsState.sessionRefs[ss.ID] {
				items = append(items, treeItem{sessionIdx: i, refIdx: j})
			}
		}
	}

	return items
}

func (a *App) currentItem() treeItem {
	items := a.treeItems()
	if len(items) == 0 {
		return treeItem{sessionIdx: 0, refIdx: -1}
	}

	c := a.sessionsState.cursor
	if c >= len(items) {
		c = len(items) - 1
	}

	return items[c]
}

func (a *App) activeSession() forge.Session {
	visible := a.visibleSessions()
	if len(visible) == 0 {
		return forge.Session{}
	}

	idx := a.currentItem().sessionIdx
	if idx >= len(visible) {
		idx = len(visible) - 1
	}

	return visible[idx]
}

// expandSession loads refs for the given session ID and marks it expanded.
func (a *App) expandSession(id string) {
	a.sessionsState.expanded[id] = true
	if _, ok := a.sessionsState.sessionRefs[id]; ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if rs, err := a.client.Refs(ctx, id); err == nil {
		a.sessionsState.sessionRefs[id] = rs
	}
}

// syncActiveRefs keeps a.refs in sync with the session under the cursor.
func (a *App) syncActiveRefs() {
	ss := a.activeSession()
	if rs, ok := a.sessionsState.sessionRefs[ss.ID]; ok {
		a.refs = rs
	}
}

func (a *App) updateSessions(m tea.KeyMsg) {
	items := a.treeItems()
	cur := a.sessionsState.cursor
	if len(items) > 0 && cur >= len(items) {
		cur = len(items) - 1
		a.sessionsState.cursor = cur
	}

	switch {
	case key.Matches(m, a.keys.Up):
		if cur > 0 {
			a.sessionsState.cursor--
			a.syncActiveRefs()
		}

	case key.Matches(m, a.keys.Down):
		if cur < len(items)-1 {
			a.sessionsState.cursor++
			a.syncActiveRefs()
		}

	case key.Matches(m, a.keys.Right):
		// expand the selected session header
		if len(items) == 0 {
			break
		}
		item := items[cur]
		if item.refIdx == -1 {
			a.expandSession(a.visibleSessions()[item.sessionIdx].ID)
		}

	case key.Matches(m, a.keys.Left):
		// if on a ref row: jump to its session header, then collapse
		// if already on a session header: collapse
		if len(items) == 0 {
			break
		}
		item := items[cur]
		if item.refIdx >= 0 {
			for i, it := range items {
				if it.sessionIdx == item.sessionIdx && it.refIdx == -1 {
					a.sessionsState.cursor = i
					break
				}
			}
		}
		ss := a.visibleSessions()[item.sessionIdx]
		a.sessionsState.expanded[ss.ID] = false
		a.syncActiveRefs()

	case key.Matches(m, a.keys.Diff): // d = toggle expand/collapse
		if len(items) == 0 {
			break
		}
		item := items[cur]
		ss := a.visibleSessions()[item.sessionIdx]
		if item.refIdx >= 0 {
			// on a ref row: collapse to parent header
			for i, it := range items {
				if it.sessionIdx == item.sessionIdx && it.refIdx == -1 {
					a.sessionsState.cursor = i
					break
				}
			}
			a.sessionsState.expanded[ss.ID] = false
		} else if a.sessionsState.expanded[ss.ID] {
			a.sessionsState.expanded[ss.ID] = false
		} else {
			a.expandSession(ss.ID)
		}
		a.syncActiveRefs()

	case key.Matches(m, a.keys.ExpandAll): // D = smart toggle all
		allExpanded := true
		for _, ss := range a.visibleSessions() {
			if !a.sessionsState.expanded[ss.ID] {
				allExpanded = false
				break
			}
		}
		if allExpanded {
			curSession := 0
			if len(items) > 0 {
				curSession = items[cur].sessionIdx
			}
			a.sessionsState.expanded = map[string]bool{}
			a.sessionsState.cursor = curSession
			a.sessionsState.offset = 0
		} else {
			for _, ss := range a.visibleSessions() {
				a.expandSession(ss.ID)
			}
		}
		a.syncActiveRefs()

	case key.Matches(m, a.keys.Enter):
		ref := ""
		if len(items) > 0 && cur < len(items) && items[cur].refIdx >= 0 {
			refs := a.sessionsState.sessionRefs[a.activeSession().ID]
			if items[cur].refIdx < len(refs) {
				r := refs[items[cur].refIdx]
				if !r.IsHead {
					ref = r.Ref
				}
			}
		}
		a.reloadLogRefs(ref)
		a.screen = ScreenLog

	case key.Matches(m, a.keys.Archive):
		ss := a.activeSession()
		if ss.ID != "" {
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
		a.sessionsState.cursor = 0
		a.sessionsState.offset = 0
	}
}

func (a *App) reloadLogRefs(ref string) {
	ss := a.activeSession()
	if ss.ID == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if ms, err := a.client.Log(ctx, ss.ID, ref); err == nil {
		a.messages = ms
		a.logState.cursor = 0
		a.logState.bodyExp = false
		a.logState.walkRef = ref
		treeExp := map[int]bool{}
		for i, msg := range ms {
			if len(msg.ToolCalls) > 0 {
				treeExp[i] = true
			}
		}
		a.logState.treeExp = treeExp
		for i := range a.sessions {
			if a.sessions[i].ID == ss.ID {
				a.sessions[i].Messages, a.sessions[i].Counts = countMessages(ms)
				break
			}
		}
	}

	if rs, err := a.client.Refs(ctx, ss.ID); err == nil {
		a.sessionsState.sessionRefs[ss.ID] = rs
		a.refs = rs
	}
}

const treeHeaderRows = 2 // column header + Hr

func (a *App) viewSessions(w, h int) string {
	s := a.styles
	visible := a.visibleSessions()
	items := a.treeItems()

	leftW := w * 70 / 100
	rightW := w - leftW - 2

	// TITLE is the only variable-width column; all others are fixed.
	// Fixed chars per row = prefix(2) + ID(10) + sep(1) + NAME(14) + sep(1)
	//   + sep(1) + PLUGINS(10) + sep(1) + MSGS(5) + sep(2) + UPDATED(16) = 63.
	// innerW = leftW minus Box border(1 each side) and padding(1 each side).
	// The Row style's own left-border + left/right padding consume 3 more chars
	// from the Width budget, so the usable content area is innerW-3.
	innerW := leftW - 4
	titleW := max(innerW-66, 10)

	// Clamp cursor and update viewport scroll.
	if len(items) > 0 && a.sessionsState.cursor >= len(items) {
		a.sessionsState.cursor = len(items) - 1
	}
	// Available lines inside the Box for data rows (subtract border + title + header rows).
	viewH := max(h-treeHeaderRows-4, 4)
	cur := a.sessionsState.cursor
	if cur < a.sessionsState.offset {
		a.sessionsState.offset = cur
	}
	if cur >= a.sessionsState.offset+viewH {
		a.sessionsState.offset = cur - viewH + 1
	}

	// Column header.
	archTag := ""
	if a.sessionsState.showArchived {
		archTag = " " + s.ChipAcc.Render("all")
	}
	headLine := s.Header.Render(fmt.Sprintf("  %-10s %-14s %-*s %-10s %5s  %s",
		"ID", "NAME", titleW, "TITLE", "PLUGINS", "MSGS", "UPDATED"))
	treeRows := []string{headLine, s.RenderHorizontalDashedRule(leftW - 4)}

	end := min(a.sessionsState.offset+viewH, len(items))
	for i := a.sessionsState.offset; i < end; i++ {
		item := items[i]
		selected := i == cur
		st := s.Row
		if selected {
			st = s.RowSel
		}

		var line string
		if item.refIdx == -1 {
			// Session header row.
			ss := visible[item.sessionIdx]
			archMark := ""
			if ss.Archived {
				archMark = " " + s.Chip.Render("arch")
			}
			glyph := s.Faint.Render("▸")
			if a.sessionsState.expanded[ss.ID] {
				glyph = s.Accent.Render("▾")
			}
			line = glyph + " " + fmt.Sprintf("%-10s %-14s %-*s %-10s %5d  %s%s",
				shortHash(ss.ID, 10),
				s.TruncateRunes(ss.Name, 14),
				titleW, s.TruncateRunes(ss.Title, titleW),
				s.TruncateRunes(ss.Plugins, 10),
				ss.Messages,
				ss.Updated.Format("2006-01-02 15:04"),
				archMark,
			)
		} else {
			// Ref row: indented under the session header.
			// Unselected rows are fully faint; selected rows light up with
			// colored ref labels so the selection is immediately obvious.
			refs := a.sessionsState.sessionRefs[visible[item.sessionIdx].ID]
			if item.refIdx < len(refs) {
				r := refs[item.refIdx]
				hash := fmt.Sprintf("%-10s", shortHash(r.Hash, 10))
				if selected {
					line = s.Faint.Render("    · ") + s.Faint.Render(hash) + "  " + refLabel(s, r)
				} else {
					line = s.Faint.Render("    · " + hash + "  " + r.Ref)
				}
			}
		}

		treeRows = append(treeRows, st.Width(leftW-4).Render(line))
	}

	table := s.RenderBox(fmt.Sprintf("sessions [%d]%s", len(visible), archTag), true, leftW, strings.Join(treeRows, "\n"))
	table = fitLines(table, h)

	// Detail panel — session info + optional ref detail when on a ref row.
	item := a.currentItem()
	ss := a.activeSession()

	maxCount := 1
	for _, n := range ss.Counts {
		if n > maxCount {
			maxCount = n
		}
	}

	detail := strings.Builder{}
	kvw := 11
	detail.WriteString(s.RenderKeyValue("ID", s.Muted.Render(shortHash(ss.ID, 16)), kvw) + "\n")
	detail.WriteString(s.RenderKeyValue("Name", s.Accent.Render(ss.Name), kvw) + "\n")
	detail.WriteString(s.RenderKeyValue("Title", s.TruncateRunes(ss.Title, rightW-14), kvw) + "\n")
	detail.WriteString(s.RenderKeyValue("Model", s.Muted.Render(ss.Model), kvw) + "\n")
	detail.WriteString(s.RenderKeyValue("Parent", s.Info.Render(ss.Parent), kvw) + "\n")
	detail.WriteString(s.RenderKeyValue("Created", s.Muted.Render(ss.Created.Format("01-02 15:04")), kvw) + "\n")
	detail.WriteString(s.RenderKeyValue("Updated", s.Muted.Render(ss.Updated.Format("01-02 15:04")), kvw) + "\n")

	detail.WriteString("\n")
	detail.WriteString(s.Faint.Render(fmt.Sprintf("MESSAGES · %d", ss.Messages)) + "\n")
	for _, role := range []string{"user", "assistant", "tool", "tool_call", "tool_result", "system"} {
		n := ss.Counts[role]
		if n == 0 {
			continue
		}
		bar := miniBar(n, maxCount, 40, s.ColAccent, s.ColRule2)
		line := fmt.Sprintf("%s %-11s  %s %3d",
			s.RoleStyle(role).Render(RoleGlyph(role)),
			s.RoleStyle(role).Render(fmt.Sprintf("%-11s", role)),
			bar, n,
		)
		detail.WriteString(line + "\n")
	}

	detail.WriteString("\n")
	detail.WriteString(s.Faint.Render("COST") + "\n")
	detail.WriteString(s.RenderKeyValue("Estimated", s.Accent.Render(fmt.Sprintf("$%.4f", ss.Cost)), kvw) + "\n")
	detail.WriteString(s.RenderKeyValue("Tokens", s.Muted.Render(fmt.Sprintf("in=%d out=%d", ss.TokensIn, ss.TokensOut)), kvw) + "\n")

	// Append ref detail when cursor is on a ref row.
	if item.refIdx >= 0 {
		refs := a.sessionsState.sessionRefs[ss.ID]
		if item.refIdx < len(refs) {
			r := refs[item.refIdx]
			detail.WriteString("\n")
			detail.WriteString(s.Faint.Render("REF") + "\n")
			detail.WriteString(s.RenderKeyValue("name", refLabel(s, r), kvw) + "\n")
			detail.WriteString(s.RenderKeyValue("hash", s.Muted.Render(r.Hash), kvw) + "\n")
			detail.WriteString(s.RenderKeyValue("type", s.Muted.Render(refType(r)), kvw) + "\n")
		}
	}

	rightBox := s.RenderBox(ss.Name, false, rightW, detail.String())
	rightBox = fitLines(rightBox, h)

	return lipgloss.JoinHorizontal(lipgloss.Top, table, "  ", rightBox)
}

func miniBar(n, maxVal, width int, accent, empty lipgloss.Color) string {
	if maxVal <= 0 {
		maxVal = 1
	}

	cells := min((n*width)/maxVal, width)
	on := lipgloss.NewStyle().Foreground(accent).Render(strings.Repeat("█", cells))
	off := lipgloss.NewStyle().Foreground(empty).Render(strings.Repeat("░", width-cells))

	return on + off
}

func shortHash(h string, n int) string {
	if len(h) <= n {
		return h
	}

	return h[:n]
}
