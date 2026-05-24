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

type resourcesState struct {
	selected int
	scope    string
}

var scopeOrder = []string{"all", "session", "archive", "global"}

func (a *App) updateResources(m tea.KeyMsg) {
	switch {
	case key.Matches(m, a.keys.Up):
		if a.resourcesState.selected > 0 {
			a.resourcesState.selected--
			a.loadCurrentResourceDetail()
		}

	case key.Matches(m, a.keys.Down):
		if a.resourcesState.selected < len(a.filteredResources())-1 {
			a.resourcesState.selected++
			a.loadCurrentResourceDetail()
		}

	case key.Matches(m, a.keys.Left):
		a.resourcesState.scope = cycleScope(a.resourcesState.scope, -1)
		a.resourcesState.selected = 0
		_ = a.reloadResources()
		a.loadCurrentResourceDetail()

	case key.Matches(m, a.keys.Right):
		a.resourcesState.scope = cycleScope(a.resourcesState.scope, +1)
		a.resourcesState.selected = 0
		_ = a.reloadResources()
		a.loadCurrentResourceDetail()

	case key.Matches(m, a.keys.Yank):
		r := a.currentResource()
		a.flash("yanked " + r.HEAD)
	}
}

func (a *App) loadCurrentResourceDetail() {
	rs := a.filteredResources()
	if len(rs) == 0 {
		return
	}

	idx := a.resourcesState.selected
	if idx >= len(rs) {
		return
	}

	r := rs[idx]
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	detail, err := a.client.ResourceDetail(ctx, r.Path, r.Name)
	if err != nil {
		return
	}

	for i := range a.resources {
		if a.resources[i].Path == r.Path && a.resources[i].Name == r.Name {
			a.resources[i] = detail
			return
		}
	}
}

func cycleScope(s string, dir int) string {
	if s == "" {
		s = "all"
	}

	for i, x := range scopeOrder {
		if x == s {
			next := (i + dir + len(scopeOrder)) % len(scopeOrder)
			return scopeOrder[next]
		}
	}

	return "all"
}

func (a *App) reloadResources() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scope := a.resourcesState.scope
	if scope == "" {
		scope = "all"
	}

	rs, err := a.client.Resources(ctx, scope)
	if err != nil {
		return err
	}

	a.resources = rs
	return nil
}

func (a *App) filteredResources() []forge.Resource {
	if a.resourcesState.scope == "" || a.resourcesState.scope == "all" {
		return a.resources
	}

	var out []forge.Resource
	for _, r := range a.resources {
		if string(r.Scope) == a.resourcesState.scope {
			out = append(out, r)
		}
	}

	return out
}

func (a *App) currentResource() forge.Resource {
	rs := a.filteredResources()
	if len(rs) == 0 {
		return forge.Resource{}
	}

	if a.resourcesState.selected >= len(rs) {
		a.resourcesState.selected = len(rs) - 1
	}

	return rs[a.resourcesState.selected]
}

func (a *App) viewResources(w int) string {
	s := a.styles

	scopeW := 22
	listW := (w - scopeW - 6) * 5 / 11
	detailW := w - scopeW - listW - 6
	if listW < 28 {
		listW = 28
	}

	if detailW < 32 {
		detailW = 32
	}

	scopeBody := strings.Builder{}
	for _, sc := range scopeOrder {
		count := 0
		for _, r := range a.resources {
			if sc == "all" || string(r.Scope) == sc {
				count++
			}
		}

		active := sc == a.resourcesState.scope || (a.resourcesState.scope == "" && sc == "all")
		line := fmt.Sprintf("%-12s %d", sc, count)
		st := s.Row
		if active {
			st = s.RowSel
		}

		scopeBody.WriteString(st.Width(scopeW-4).Render(line) + "\n")
	}

	scopeBody.WriteString("\n")
	scopeBody.WriteString(s.Faint.Render("STORE") + "\n")
	scopeBody.WriteString(KV(s, "backend", s.Muted.Render(a.system.Storage.Backend), 9) + "\n")
	scopeBody.WriteString(KV(s, "objects", s.Muted.Render(fmt.Sprintf("%d", a.system.Storage.Objects)), 9) + "\n")
	scopeBody.WriteString(KV(s, "refs", s.Muted.Render(fmt.Sprintf("%d", a.system.Storage.Refs)), 9) + "\n")
	scopeBody.WriteString(KV(s, "swept", s.Muted.Render(fmt.Sprintf("%d", a.system.Storage.Swept)), 9))
	scopeBox := Box(s, "scope", false, scopeW, scopeBody.String())

	filtered := a.filteredResources()
	listBody := strings.Builder{}
	listBody.WriteString(s.Header.Render(fmt.Sprintf("%-26s %4s %8s %s", "PATH", "VER", "SIZE", "SCOPE")) + "\n")
	listBody.WriteString(Hr(s, listW-4) + "\n")
	for i, r := range filtered {
		line := fmt.Sprintf("%-26s %4s %8s %s",
			Truncate(r.Path, 26),
			fmt.Sprintf("v%d", r.Versions),
			r.Size,
			s.Chip.Render(string(r.Scope)),
		)

		st := s.Row
		if i == a.resourcesState.selected {
			st = s.RowSel
		}

		listBody.WriteString(st.Width(listW-4).Render(line) + "\n")
	}

	if len(filtered) == 0 {
		listBody.WriteString("\n" + s.Faint.Render("  no resources in scope "+a.resourcesState.scope))
	}

	listBox := Box(s, fmt.Sprintf("resources [%d]", len(filtered)), true, listW, listBody.String())

	r := a.currentResource()
	detail := strings.Builder{}
	detail.WriteString(KV(s, "path", s.Muted.Render(r.Path), 9) + "\n")
	detail.WriteString(KV(s, "scope", s.Chip.Render(string(r.Scope)), 9) + "\n")
	detail.WriteString(KV(s, "mime", s.Muted.Render(r.MIME), 9) + "\n")
	detail.WriteString(KV(s, "versions", s.Muted.Render(fmt.Sprintf("%d", r.Versions)), 9) + "\n")
	detail.WriteString(KV(s, "size", s.Muted.Render(r.Size), 9) + "\n")
	detail.WriteString(KV(s, "updated", s.Muted.Render(r.Updated.Format("2006-01-02 15:04:05")), 9) + "\n")
	detail.WriteString(KV(s, "HEAD", s.Accent.Render(r.HEAD), 9) + "\n")
	detail.WriteString("\n")

	detail.WriteString(s.Faint.Render(fmt.Sprintf("HISTORY · %d of %d", len(r.History), r.Versions)) + "\n")

	for i, h := range r.History {
		marker := "  "
		if i == 0 {
			marker = s.Accent.Render("▍ ")
		}

		line := fmt.Sprintf("%sv%-3d %s  %s  %s",
			marker, h.Version,
			s.Faint.Render(h.Hash),
			s.Faint.Render(h.Date.Format("2006-01-02 15:04")),
			s.Muted.Render(Truncate(h.Delta, detailW-32)),
		)

		detail.WriteString(line + "\n")
	}
	detail.WriteString("\n")

	if r.Summary != "" {
		detail.WriteString(s.Faint.Render("SUMMARY") + "\n")
		detail.WriteString(s.Muted.Render(r.Summary))
	}

	detailBox := Box(s, r.Name, false, detailW, detail.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, scopeBox, "  ", listBox, "  ", detailBox)
}
