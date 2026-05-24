// Package ui implements the Bubble Tea + Lip Gloss layer for anvil.
//
// Files:
//
//	theme.go       palette · borders · Styles bundle
//	keys.go        keymap
//	components.go  reusable widgets (box, chip, spark)
//	chrome.go      term-bar · tab-bar · key-hints · status-bar
//	app.go         root Model — owns Screen + active forge.Client
//	sessions.go    screen 1 · forge sessions status
//	log.go         screen 2 · forge sessions log
//	branches.go    screen 3 · forge sessions branch
//	resources.go   screen 4 · forge resources status
//	system.go      screen 5 · forge system status
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// ─── color tokens ──────────────────────────────────────────────────────

var (
	ColBg      = lipgloss.Color("#161d27")
	ColFg      = lipgloss.Color("#e6e8eb")
	ColFgDim   = lipgloss.Color("#8a93a3")
	ColFgFaint = lipgloss.Color("#4a5364")
	ColRule    = lipgloss.Color("#1f2731")
	ColRule2   = lipgloss.Color("#2a3340")

	ColOk     = lipgloss.Color("#4ade80")
	ColInfo   = lipgloss.Color("#60a5fa")
	ColWarn   = lipgloss.Color("#fbbf24")
	ColDanger = lipgloss.Color("#f87171")

	AccentAmber = lipgloss.Color("#f59e0b")
)

// ─── styles bundle ─────────────────────────────────────────────────────

type Styles struct {
	App        lipgloss.Style
	TermBar    lipgloss.Style
	Box        lipgloss.Style
	BoxFocused lipgloss.Style
	BoxTitle   lipgloss.Style

	Row    lipgloss.Style
	RowSel lipgloss.Style

	Header lipgloss.Style

	Tab       lipgloss.Style
	TabActive lipgloss.Style

	StatusBar lipgloss.Style
	KeyHint   lipgloss.Style
	KeyCap    lipgloss.Style

	Chip    lipgloss.Style
	ChipAcc lipgloss.Style

	Muted, Faint, Accent   lipgloss.Style
	OK, Info, Warn, Danger lipgloss.Style

	User, Assistant, ToolCall, ToolResult, System lipgloss.Style
}

func NewStyles() Styles {
	acc := AccentAmber
	bx := lipgloss.RoundedBorder()

	box := lipgloss.NewStyle().
		Border(bx).
		BorderForeground(ColRule2).
		Foreground(ColFg).
		Padding(0, 1)

	boxFocused := box.BorderForeground(acc)

	row := lipgloss.NewStyle().
		Foreground(ColFgDim).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: " "}).
		Padding(0, 1)
	rowSel := row.
		Foreground(ColFg).
		Background(lipgloss.Color("#121920")).
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "▍"}).
		BorderForeground(acc)

	chip := lipgloss.NewStyle().
		Foreground(ColFgDim).
		Padding(0, 1)
	chipAcc := chip.
		Foreground(acc).
		Background(lipgloss.Color("#0e1a24"))

	return Styles{
		App: lipgloss.NewStyle().Foreground(ColFg),
		TermBar: lipgloss.NewStyle().
			Foreground(ColFgDim).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColRule),
		Box:        box,
		BoxFocused: boxFocused,
		BoxTitle: lipgloss.NewStyle().
			Foreground(ColFgDim).
			Bold(true).
			Padding(0, 1),

		Row:    row,
		RowSel: rowSel,
		Header: lipgloss.NewStyle().
			Foreground(ColFgFaint).
			Padding(0, 1),

		Tab: lipgloss.NewStyle().
			Foreground(ColFgDim).
			Padding(0, 2),
		TabActive: lipgloss.NewStyle().
			Foreground(acc).
			Padding(0, 2),

		StatusBar: lipgloss.NewStyle().
			Background(acc).
			Foreground(ColBg).
			Bold(true),

		KeyHint: lipgloss.NewStyle().
			Foreground(ColFgFaint),
		KeyCap: lipgloss.NewStyle().
			Foreground(ColFgDim).
			Padding(0, 1),

		Chip: chip, ChipAcc: chipAcc,

		Muted: lipgloss.NewStyle().
			Foreground(ColFgDim),
		Faint: lipgloss.NewStyle().
			Foreground(ColFgFaint),
		Accent: lipgloss.NewStyle().
			Foreground(acc),
		OK: lipgloss.NewStyle().
			Foreground(ColOk),
		Info: lipgloss.NewStyle().
			Foreground(ColInfo),
		Warn: lipgloss.NewStyle().
			Foreground(ColWarn),
		Danger: lipgloss.NewStyle().
			Foreground(ColDanger),

		User: lipgloss.NewStyle().
			Foreground(acc),
		Assistant: lipgloss.NewStyle().
			Foreground(ColFgDim),
		ToolCall: lipgloss.NewStyle().
			Foreground(ColInfo),
		ToolResult: lipgloss.NewStyle().
			Foreground(ColFgFaint),
		System: lipgloss.NewStyle().
			Foreground(ColWarn),
	}
}

func (s Styles) RoleStyle(role string) lipgloss.Style {
	switch role {
	case "user":
		return s.User
	case "assistant":
		return s.Assistant
	case "tool_call":
		return s.ToolCall
	case "tool_result":
		return s.ToolResult
	case "system":
		return s.System
	}

	return s.Muted
}

func RoleGlyph(role string) string {
	switch role {
	case "user":
		return "◆"
	case "assistant":
		return "○"
	case "tool_call":
		return "→"
	case "tool_result":
		return "←"
	case "system":
		return "§"
	}

	return "·"
}
