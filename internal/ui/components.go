package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Box renders a titled bordered panel of the given width.
func Box(s Styles, title string, focused bool, width int, body string) string {
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

	return st.Render(titleStyle.Render(focusedGlyph(focused)+" "+title) + "\n" + body)
}

func focusedGlyph(focused bool) string {
	if focused {
		return "▸"
	}

	return "·"
}

// Spark renders a unicode sparkline from the given series.
func Spark(values []int) string {
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

// Truncate cuts s to fit n runes, appending ellipsis if it overflows.
func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}

	r := []rune(s)
	if len(r) <= n {
		return s
	}

	if n <= 1 {
		return "…"
	}

	return string(r[:n-1]) + "…"
}

// Hr renders a horizontal dashed rule.
func Hr(s Styles, width int) string {
	if width <= 0 {
		return ""
	}

	return lipgloss.NewStyle().Foreground(s.ColRule).Render(strings.Repeat("─", width))
}

// KV renders one "label  value" row with a dim right-padded label.
func KV(s Styles, label, value string, labelWidth int) string {
	return s.Faint.Render(padRight(label, labelWidth)) + " " + value
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}

	return s + strings.Repeat(" ", n-len(s))
}
