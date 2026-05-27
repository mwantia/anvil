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

// ─── fixed color tokens (same across all themes) ──────────────────────────

var (
	ColFg      = lipgloss.Color("#e6e8eb")
	ColFgDim   = lipgloss.Color("#8a93a3")
	ColFgFaint = lipgloss.Color("#4a5364")

	ColOk     = lipgloss.Color("#4ade80")
	ColInfo   = lipgloss.Color("#60a5fa")
	ColWarn   = lipgloss.Color("#fbbf24")
	ColDanger = lipgloss.Color("#f87171")
)

// ─── theme ────────────────────────────────────────────────────────────────

// Theme holds the per-theme color values. Background tones are tinted subtly
// toward the accent hue to give each theme a distinct atmosphere.
type Theme struct {
	Name      string
	Accent    lipgloss.Color
	Bg        lipgloss.Color // app background
	Rule      lipgloss.Color // HR dividers
	Rule2     lipgloss.Color // box borders, empty bar fill
	RowSelBg  lipgloss.Color // selected row background
	ChipAccBg lipgloss.Color // accent chip background
}

var Themes = []Theme{
	{
		Name:      "amber",
		Accent:    lipgloss.Color("#f59e0b"),
		Bg:        lipgloss.Color("#161d27"),
		Rule:      lipgloss.Color("#1f2731"),
		Rule2:     lipgloss.Color("#2a3340"),
		RowSelBg:  lipgloss.Color("#121920"),
		ChipAccBg: lipgloss.Color("#0e1a24"),
	},
	{
		Name:      "cyan",
		Accent:    lipgloss.Color("#06b6d4"),
		Bg:        lipgloss.Color("#141e21"),
		Rule:      lipgloss.Color("#1d2c30"),
		Rule2:     lipgloss.Color("#263a3f"),
		RowSelBg:  lipgloss.Color("#0f1c1f"),
		ChipAccBg: lipgloss.Color("#0b181c"),
	},
	{
		Name:      "slate",
		Accent:    lipgloss.Color("#94a3b8"),
		Bg:        lipgloss.Color("#161820"),
		Rule:      lipgloss.Color("#1f222e"),
		Rule2:     lipgloss.Color("#282b3a"),
		RowSelBg:  lipgloss.Color("#11131e"),
		ChipAccBg: lipgloss.Color("#0e101c"),
	},
	{
		Name:      "lime",
		Accent:    lipgloss.Color("#84cc16"),
		Bg:        lipgloss.Color("#151e14"),
		Rule:      lipgloss.Color("#1e2b1c"),
		Rule2:     lipgloss.Color("#263826"),
		RowSelBg:  lipgloss.Color("#101a10"),
		ChipAccBg: lipgloss.Color("#0c160c"),
	},
	{
		Name:      "magenta",
		Accent:    lipgloss.Color("#d946ef"),
		Bg:        lipgloss.Color("#1a1520"),
		Rule:      lipgloss.Color("#261e2e"),
		Rule2:     lipgloss.Color("#32263a"),
		RowSelBg:  lipgloss.Color("#14101a"),
		ChipAccBg: lipgloss.Color("#100d16"),
	},
}

// ─── styles bundle ─────────────────────────────────────────────────────────

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

	// Raw colors for ad-hoc lipgloss.NewStyle() calls in other files.
	ColAccent lipgloss.Color
	ColRule   lipgloss.Color
	ColRule2  lipgloss.Color
	ColBg     lipgloss.Color
}

func NewStyles(t Theme) Styles {
	acc := t.Accent
	bx := lipgloss.RoundedBorder()

	box := lipgloss.NewStyle().
		Border(bx).
		BorderForeground(t.Rule2).
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
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "▍"}).
		BorderForeground(acc)

	chip := lipgloss.NewStyle().
		Foreground(ColFgDim).
		Padding(0, 1)
	chipAcc := chip.
		Foreground(acc).
		Background(t.ChipAccBg)

	return Styles{
		App: lipgloss.NewStyle().Foreground(ColFg),
		TermBar: lipgloss.NewStyle().
			Foreground(ColFgDim).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(t.Rule),
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
			Foreground(t.Bg).
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

		ColAccent: acc,
		ColRule:   t.Rule,
		ColRule2:  t.Rule2,
		ColBg:     t.Bg,
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
