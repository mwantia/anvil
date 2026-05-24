package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TermBar renders the top window chrome.
func TermBar(s Styles, w int, screenLabel string) string {
	dot := lipgloss.NewStyle().Foreground(s.ColAccent).Render("●")
	name := lipgloss.NewStyle().Bold(true).Foreground(ColFg).Render("anvil")
	path := s.Faint.Render(screenLabel + " · ~/forge")

	left := dot + " " + name + "  " + path
	right := s.Faint.Render("go 1.22 · bubbletea")

	return s.TermBar.Width(w).Render(pad(left, right, w-2))
}

// TabBar renders the 1-3 hotkey tab strip. The right side shows the forge
// daemon address and a health indicator instead of session/HEAD state
// (that information is already present in the StatusBar).
func TabBar(s Styles, w int, active int, screenNames []string, address, health string) string {
	var parts []string
	for i, n := range screenNames {
		st := s.Tab
		if i == active {
			st = s.TabActive
		}

		parts = append(parts, st.Render(fmt.Sprintf("%d %s", i+1, n)))
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Bottom, parts...)
	right := s.Faint.Render("forge") + " " + s.Muted.Render(address) + " " + health

	return pad(tabs, right, w-2)
}

// StatusBar renders the bottom accent-colored bar.
func StatusBar(s Styles, w int, left, right []string) string {
	l := " " + strings.Join(left, "   ") + " "
	r := " " + strings.Join(right, "   ") + " "

	return s.StatusBar.Width(w).Render(pad(l, r, w))
}

// KeyHints renders the chip footer above the status bar.
func KeyHints(s Styles, items [][2]string) string {
	cells := make([]string, 0, len(items))
	for _, it := range items {
		cells = append(cells, s.KeyCap.Render(it[0])+" "+s.KeyHint.Render(it[1]))
	}

	return "  " + strings.Join(cells, "   ")
}

func pad(left, right string, w int) string {
	return left + strings.Repeat(" ", max(w-lipgloss.Width(left)-lipgloss.Width(right), 1)) + right
}
