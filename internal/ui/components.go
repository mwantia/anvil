package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	runewidth "github.com/mattn/go-runewidth"
)

// Box renders a titled bordered panel of the given width.
func (s Styles) RenderBox(title string, focused bool, width int, body string) string {
	st := s.Box
	titleStyle := s.BoxTitle
	if focused {
		st = s.BoxFocused
		titleStyle = s.BoxTitle.Foreground(s.ColAccent)
	}

	if width > 0 {
		st = st.Width(width)
	}

	if title == "" {
		return st.Render(body)
	}

	return st.Render(titleStyle.Render(s.RenderFocusedGlyph(focused)+" "+title) + "\n" + body)
}

func (Styles) RenderFocusedGlyph(focused bool) string {
	if focused {
		return "▸"
	}

	return "·"
}

// Spark renders a unicode sparkline from the given series.
func (Styles) RenderSpark(values []int) string {
	const blocks = "▁▂▃▄▅▆▇█"
	if len(values) == 0 {
		return ""
	}

	maxVal := 1
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	runes := []rune(blocks)
	var b strings.Builder
	for _, v := range values {
		idx := max((v*(len(runes)-1))/maxVal, 0)
		if idx >= len(runes) {
			idx = len(runes) - 1
		}

		b.WriteRune(runes[idx])
	}

	return b.String()
}

func (s Styles) RenderHorizontalDashedRule(width int) string {
	if width <= 0 {
		return ""
	}

	return lipgloss.NewStyle().Foreground(s.ColRule).Render(strings.Repeat("─", width))
}

func (s Styles) RenderKeyValue(label, value string, width int) string {
	return s.Faint.Render(s.PadToRight(label, width)) + " " + value
}

func (Styles) PadToRight(s string, n int) string {
	if len(s) >= n {
		return s
	}

	return s + strings.Repeat(" ", n-len(s))
}

// TruncateRunes cuts s to fit n terminal columns, appending ellipsis if it overflows.
// Wide characters (emoji, CJK) count as 2 columns each.
func (Styles) TruncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	cols := 0
	var out []rune
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if cols+w > n-1 {
			break
		}

		if r == '🩶' {
			out = append(out, ' ')
		} else {
			out = append(out, r)
		}
		cols += w
	}
	return string(out) + "…"
}
