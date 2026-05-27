package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/mwantia/anvil/internal/forge"
)

// logTreeItem identifies a single row in the expandable log tree.
// toolCallIdx == -1 means a message row; >= 0 is a tool-call sub-row.
type logTreeItem struct {
	msgIdx      int
	toolCallIdx int
}

type logState struct {
	cursor  int
	bodyExp bool         // body text expanded for current message row
	treeExp map[int]bool // msgIdx → tool-call rows shown
	walkRef string       // non-empty when walking a specific ref
}

func (a *App) logTreeItems() []logTreeItem {
	var items []logTreeItem
	for i, msg := range a.messages {
		items = append(items, logTreeItem{msgIdx: i, toolCallIdx: -1})
		if a.logState.treeExp[i] && len(msg.ToolCalls) > 0 {
			for j := range msg.ToolCalls {
				items = append(items, logTreeItem{msgIdx: i, toolCallIdx: j})
			}
		}
	}
	return items
}

func (a *App) currentLogItem() logTreeItem {
	items := a.logTreeItems()
	if len(items) == 0 {
		return logTreeItem{msgIdx: 0, toolCallIdx: -1}
	}
	c := a.logState.cursor
	if c >= len(items) {
		c = len(items) - 1
	}
	return items[c]
}

func (a *App) currentMessage() forge.Message {
	if len(a.messages) == 0 {
		return forge.Message{}
	}
	idx := a.currentLogItem().msgIdx
	if idx >= len(a.messages) {
		idx = len(a.messages) - 1
	}
	return a.messages[idx]
}

func (a *App) updateLog(m tea.KeyMsg) {
	items := a.logTreeItems()
	cur := a.logState.cursor
	if len(items) > 0 && cur >= len(items) {
		cur = len(items) - 1
		a.logState.cursor = cur
	}

	switch {
	case key.Matches(m, a.keys.Up):
		if a.logState.cursor > 0 {
			a.logState.cursor--
		}

	case key.Matches(m, a.keys.Down):
		if a.logState.cursor < len(items)-1 {
			a.logState.cursor++
		}

	case key.Matches(m, a.keys.Right):
		if len(items) == 0 {
			break
		}
		item := items[cur]
		if item.toolCallIdx == -1 {
			msg := a.messages[item.msgIdx]
			if len(msg.ToolCalls) > 0 {
				a.logState.treeExp[item.msgIdx] = true
			}
		}

	case key.Matches(m, a.keys.Left):
		if len(items) == 0 {
			break
		}
		item := items[cur]
		if item.toolCallIdx >= 0 {
			// jump to parent message row first
			for i, it := range items {
				if it.msgIdx == item.msgIdx && it.toolCallIdx == -1 {
					a.logState.cursor = i
					break
				}
			}
		}
		a.logState.treeExp[item.msgIdx] = false

	case key.Matches(m, a.keys.Diff): // d = toggle tree expand/collapse
		if len(items) == 0 {
			break
		}
		item := items[cur]
		if item.toolCallIdx >= 0 {
			// collapse to parent
			for i, it := range items {
				if it.msgIdx == item.msgIdx && it.toolCallIdx == -1 {
					a.logState.cursor = i
					break
				}
			}
			a.logState.treeExp[item.msgIdx] = false
		} else {
			msg := a.messages[item.msgIdx]
			if len(msg.ToolCalls) > 0 {
				a.logState.treeExp[item.msgIdx] = !a.logState.treeExp[item.msgIdx]
			}
		}

	case key.Matches(m, a.keys.ExpandAll): // D = smart toggle all tool-call rows
		allExpanded := true
		for i, msg := range a.messages {
			if len(msg.ToolCalls) > 0 && !a.logState.treeExp[i] {
				allExpanded = false
				break
			}
		}
		if allExpanded {
			a.logState.treeExp = map[int]bool{}
		} else {
			for i, msg := range a.messages {
				if len(msg.ToolCalls) > 0 {
					a.logState.treeExp[i] = true
				}
			}
		}

	case key.Matches(m, a.keys.Enter):
		if len(items) > 0 && items[cur].toolCallIdx == -1 {
			a.logState.bodyExp = !a.logState.bodyExp
		}

	case key.Matches(m, a.keys.Edit):
		msg := a.currentMessage()
		_, _ = a.client.EditFork(context.Background(), a.activeSession().ID, msg.Hash, "")
		a.reloadLogRefs(a.logState.walkRef)
		a.flash("forge sessions edit " + a.activeSession().Name + " " + shortHash(msg.Hash, 8))

	case key.Matches(m, a.keys.Clone):
		msg := a.currentMessage()
		_ = a.client.Checkout(context.Background(), a.activeSession().ID, msg.Hash)
		a.reloadLogRefs(a.logState.walkRef)
		a.flash("forge sessions checkout " + a.activeSession().Name + " " + shortHash(msg.Hash, 8))

	case key.Matches(m, a.keys.Yank):
		a.flash("yanked " + shortHash(a.currentMessage().Hash, 12))
	}
}

func (a *App) viewLog(w, h int) string {
	s := a.styles
	ss := a.activeSession()

	header := strings.Builder{}
	header.WriteString(s.Accent.Render(ss.Name) + s.Faint.Render(" · ") + s.Muted.Render(ss.Model))
	displayRef := a.logState.walkRef
	if displayRef == "" {
		displayRef = a.headRefLabel()
	}
	if displayRef != "" {
		header.WriteString(s.Faint.Render(" @ ") + s.Accent.Render(displayRef))
	}

	header.WriteString("\n")
	countParts := []string{}
	for _, role := range []string{"user", "assistant", "tool", "system"} {
		if n := ss.Counts[role]; n > 0 {
			countParts = append(countParts, fmt.Sprintf("%s=%d", role, n))
		}
	}
	header.WriteString(s.Muted.Render(fmt.Sprintf("%d messages", ss.Messages)) + s.Faint.Render("  "+strings.Join(countParts, "  ")) + "\n")
	header.WriteString(s.Faint.Render("tokens: ") + s.Muted.Render(fmt.Sprintf(
		"in=%d  out=%d  total=%d  cost=$%.6f",
		ss.TokensIn, ss.TokensOut, ss.TokensIn+ss.TokensOut, ss.Cost)))
	headerBox := s.RenderBox("log · "+ss.Name, false, w-2, header.String())

	headerH := lipgloss.Height(headerBox)
	bodyH := max(h-headerH-1, 4)

	left := (w - 4) / 2
	right := w - left - 4

	items := a.logTreeItems()

	// Clamp cursor.
	if len(items) > 0 && a.logState.cursor >= len(items) {
		a.logState.cursor = len(items) - 1
	}

	listRows := make([]string, 0, len(items))
	for i, item := range items {
		selected := i == a.logState.cursor
		if item.toolCallIdx == -1 {
			msg := a.messages[item.msgIdx]
			listRows = append(listRows, a.renderLogRow(msg, left-4, selected))
		} else {
			tc := a.messages[item.msgIdx].ToolCalls[item.toolCallIdx]
			listRows = append(listRows, a.renderToolCallRow(tc, left-4, selected))
		}
	}

	listContent := fitLines(strings.Join(listRows, "\n"), bodyH-3)
	list := s.RenderBox(fmt.Sprintf("messages [%d]", len(a.messages)), true, left, listContent)
	list = fitLines(list, bodyH)

	curItem := a.currentLogItem()
	var title strings.Builder
	var content strings.Builder

	if curItem.toolCallIdx >= 0 {
		tc := a.messages[curItem.msgIdx].ToolCalls[curItem.toolCallIdx]
		title.WriteString(tc.Name)
		if tc.ID != "" {
			title.WriteString(s.Faint.Render(" · ") + s.Muted.Render(tc.ID))
		}
		content.WriteString(fitLines(a.renderToolCallDetail(tc), bodyH-3))
	} else {
		msg := a.currentMessage()
		title.WriteString(shortHash(msg.Hash, 12))
		for _, r := range msg.Refs {
			title.WriteString(" " + s.ChipAcc.Render(r))
		}
		content.WriteString(fitLines(a.renderMessageDetail(msg), bodyH-3))
	}

	detail := s.RenderBox(title.String(), false, right, content.String())
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

	// Expand/collapse glyph for messages with tool calls.
	glyph := "  "
	if len(msg.ToolCalls) > 0 {
		msgIdx := -1
		for i := range a.messages {
			if a.messages[i].Hash == msg.Hash {
				msgIdx = i
				break
			}
		}
		if msgIdx >= 0 && a.logState.treeExp[msgIdx] {
			glyph = s.Accent.Render("▾ ")
		} else {
			glyph = s.Faint.Render("▸ ")
		}
	}

	previewWidth := max(width-30-lipgloss.Width(refsStr), 10)
	line := glyph + fmt.Sprintf("%s %s %s %s",
		s.Faint.Render(shortHash(msg.Hash, 8)),
		s.RoleStyle(role).Render(RoleGlyph(role)),
		s.RoleStyle(role).Render(fmt.Sprintf("%-11s", role)),
		s.TruncateRunes(msg.Preview, previewWidth),
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

func (a *App) renderToolCallRow(tc forge.ToolCall, width int, selected bool) string {
	s := a.styles

	// Compact single-line arg preview: first key=value, value collapsed to one line.
	argPreview := ""
	if len(tc.Arguments) > 0 {

		keys := make([]string, 0, len(tc.Arguments))
		for k := range tc.Arguments {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		parts := make([]string, 0, 2)
		for _, k := range keys {
			val := tc.Arguments[k]

			var vs string
			switch vt := val.(type) {
			case string:
				vs = strings.ReplaceAll(strings.ReplaceAll(vt, "\n", " "), "\r", "")
			default:
				b, _ := json.Marshal(val)
				vs = string(b)
			}
			parts = append(parts, fmt.Sprintf("%s=%s", k, vs))
			if len(parts) >= 2 {
				break
			}
		}
		argPreview = strings.Join(parts, " ")
	}

	previewWidth := max(width-34, 10)
	truncated := s.TruncateRunes(argPreview, previewWidth)
	var line string
	if selected {
		line = s.Faint.Render("    · ") + s.Info.Render(fmt.Sprintf("%-20s", tc.Name)) + " " + s.Muted.Render(truncated)
	} else {
		line = s.Faint.Render("    · " + fmt.Sprintf("%-20s", tc.Name) + " " + truncated)
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

	body := strings.Builder{}

	body.WriteString(s.RenderKeyValue("Role", s.RoleStyle(role).Render(RoleGlyph(role)+" "+role), 8) + "\n")
	body.WriteString(s.RenderKeyValue("Date", s.Muted.Render(formatDate(msg.Date)), 8) + "\n")
	body.WriteString(s.RenderKeyValue("Tokens", s.Muted.Render(fmt.Sprintf("in=%d out=%d", msg.TokIn, msg.TokOut)), 8) + "\n")
	if len(msg.ToolCalls) > 0 {
		body.WriteString(s.RenderKeyValue("Calls", s.Muted.Render(fmt.Sprintf("%d tool call(s) · →/d expand", len(msg.ToolCalls))), 8) + "\n")
	}

	body.WriteString("\n")
	if a.logState.bodyExp && len(msg.Body) > 0 {
		body.WriteString(renderBody(string(msg.Role), msg.Body, a.detailWidth()))
	} else {
		body.WriteString(s.Muted.Render(msg.Preview))
		if len(msg.Body) > 1 {
			body.WriteString("\n\n" + s.Faint.Render(fmt.Sprintf("(enter to expand · %d lines)", len(msg.Body))))
		}
	}

	return body.String()
}

func (a *App) renderToolCallDetail(tc forge.ToolCall) string {
	s := a.styles

	body := strings.Builder{}
	body.WriteString(s.RenderKeyValue("Tool", s.Info.Render(tc.Name), 8) + "\n")
	if tc.ID != "" {
		body.WriteString(s.RenderKeyValue("ID", s.Faint.Render(tc.ID), 8) + "\n")
	}
	body.WriteString("\n")

	if len(tc.Arguments) > 0 {
		body.WriteString(s.Faint.Render("ARGUMENTS") + "\n")
		raw, err := json.MarshalIndent(tc.Arguments, "", "  ")
		if err == nil {
			fenced := "```json\n" + string(raw) + "\n```"
			r, rerr := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(a.detailWidth()),
			)
			if rerr == nil {
				if rendered, rerr := r.Render(fenced); rerr == nil {
					body.WriteString(strings.TrimRight(rendered, "\n"))
					return body.String()
				}
			}
			body.WriteString(s.Muted.Render(string(raw)))
		}
	}

	return body.String()
}

// detailWidth returns the approximate character width available for the detail panel.
func (a *App) detailWidth() int {
	right := a.width - (a.width-4)/2 - 4
	if right < 20 {
		return 20
	}
	return right
}

// renderBody renders full message body with markdown or JSON syntax highlighting.
func renderBody(role string, lines []string, width int) string {
	raw := strings.Join(lines, "\n")

	switch role {
	case "tool_call", "tool_result":
		// Pretty-print JSON if possible, then wrap in a fenced code block for glamour.
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err == nil {
			if pretty, err := json.MarshalIndent(v, "", "  "); err == nil {
				raw = string(pretty)
			}
		}
		raw = "```json\n" + raw + "\n```"
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return raw
	}

	rendered, err := r.Render(raw)
	if err != nil {
		return raw
	}

	return strings.TrimRight(rendered, "\n")
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "—"
	}

	return t.Format("Mon Jan 2 15:04:05 2006")
}
