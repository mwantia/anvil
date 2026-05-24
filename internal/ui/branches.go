package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mwantia/anvil/internal/forge"
)

type branchesState struct {
	selectedRef int
	pane        branchesPane
}

type branchesPane int

const (
	paneRefs branchesPane = iota
	paneDag
	paneDetail
)

func (a *App) updateBranches(m tea.KeyMsg) {
	switch {
	case key.Matches(m, a.keys.Up):
		if a.branchesState.pane == paneRefs && a.branchesState.selectedRef > 0 {
			a.branchesState.selectedRef--
		}

	case key.Matches(m, a.keys.Down):
		if a.branchesState.pane == paneRefs && a.branchesState.selectedRef < len(a.refs)-1 {
			a.branchesState.selectedRef++
		}

	case key.Matches(m, a.keys.Left):
		if a.branchesState.pane > 0 {
			a.branchesState.pane--
		}

	case key.Matches(m, a.keys.Right):
		if int(a.branchesState.pane) < int(paneDetail) {
			a.branchesState.pane++
		}

	case key.Matches(m, a.keys.Tab):
		a.branchesState.pane = (a.branchesState.pane + 1) % 3

	case key.Matches(m, a.keys.Enter), key.Matches(m, a.keys.Checkout):
		r := a.currentRef()
		_ = a.client.Checkout(context.Background(), a.activeSession().ID, r.Ref)
		a.reloadLogRefs()
		a.flash("forge sessions checkout " + a.activeSession().Name + " " + r.Ref)

	case key.Matches(m, a.keys.Branch):
		r := a.currentRef()
		_ = a.client.Branch(context.Background(), a.activeSession().ID, "new-branch", r.Hash)
		a.reloadLogRefs()
		a.flash("forge sessions branch " + a.activeSession().Name + " new-branch")

	case key.Matches(m, a.keys.Merge):
		r := a.currentRef()
		_ = a.client.Merge(context.Background(), a.activeSession().ID, r.Ref)
		a.flash("forge sessions merge " + r.Ref)

	case key.Matches(m, a.keys.Delete):
		r := a.currentRef()
		if !r.IsHead && r.Ref != "main" {
			_ = a.client.DeleteRef(context.Background(), a.activeSession().ID, r.Ref)
			a.reloadLogRefs()
			a.flash("forge sessions branch -d " + r.Ref)
		}

	case key.Matches(m, a.keys.Yank):
		a.flash("yanked " + shortHash(a.currentRef().Hash, 12))
	}
}

func (a *App) currentRef() forge.Ref {
	if len(a.refs) == 0 {
		return forge.Ref{}
	}

	if a.branchesState.selectedRef >= len(a.refs) {
		a.branchesState.selectedRef = len(a.refs) - 1
	}

	return a.refs[a.branchesState.selectedRef]
}

func (a *App) viewBranches(w int, focused bool) string {
	s := a.styles
	leftW := 28
	rightW := 36
	midW := max(w-leftW-rightW-6, 30)

	refsBody := strings.Builder{}
	refsBody.WriteString(s.Header.Render(fmt.Sprintf("%-22s %s", "REF", "HASH")) + "\n")
	refsBody.WriteString(Hr(leftW-4) + "\n")

	for i, r := range a.refs {
		label := refLabel(s, r)
		line := fmt.Sprintf("%-22s %s", label, s.Faint.Render(shortHash(r.Hash, 12)))
		st := s.Row
		if i == a.branchesState.selectedRef {
			st = s.RowSel
		}
		refsBody.WriteString(st.Width(leftW-4).Render(line) + "\n")
	}

	refsBody.WriteString("\n")
	refsBody.WriteString(s.ChipAcc.Render("c") + " " + s.Faint.Render("checkout") + "  ")
	refsBody.WriteString(s.Chip.Render("b") + " " + s.Faint.Render("branch") + "\n")
	refsBody.WriteString(s.Chip.Render("m") + " " + s.Faint.Render("merge") + "  ")
	refsBody.WriteString(s.Chip.Render("x") + " " + s.Faint.Render("delete"))

	refsBox := Box(s, fmt.Sprintf("refs [%d]", len(a.refs)), focused && a.branchesState.pane == paneRefs, leftW, refsBody.String())
	dagBox := Box(s, "dag · log --graph", focused && a.branchesState.pane == paneDag, midW, a.renderDag())

	r := a.currentRef()
	detail := strings.Builder{}
	detail.WriteString(KV(s, "ref", refLabel(s, r), 8) + "\n")
	if r.Target != "" {
		detail.WriteString(KV(s, "→", s.Accent.Render(r.Target), 8) + "\n")
	}

	detail.WriteString(KV(s, "hash", s.Muted.Render(r.Hash), 8) + "\n")
	detail.WriteString(KV(s, "type", s.Muted.Render(refType(r)), 8) + "\n")
	detail.WriteString("\n")
	detail.WriteString(s.Faint.Render("ACTIONS") + "\n")
	detail.WriteString(s.ChipAcc.Render("c") + " " + s.Muted.Render("forge sessions checkout "+a.activeSession().Name+" "+r.Ref) + "\n")
	detail.WriteString(s.Chip.Render("b") + " " + s.Muted.Render("branch from "+shortHash(r.Hash, 8)) + "\n")
	detail.WriteString(s.Chip.Render("m") + " " + s.Muted.Render("merge into HEAD") + "\n")

	if !r.IsHead && r.Ref != "main" {
		detail.WriteString(s.Chip.Render("x") + " " + s.Muted.Render("delete ref") + "\n")
	}

	detailBox := Box(s, r.Ref, focused && a.branchesState.pane == paneDetail, rightW, detail.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, refsBox, "  ", dagBox, "  ", detailBox)
}

func refLabel(s Styles, r forge.Ref) string {
	switch {
	case r.IsHead:
		return s.Accent.Render("HEAD") + s.Faint.Render(" → ") + s.Accent.Render(r.Target)

	case r.Ref == "main":
		return s.Info.Render(r.Ref)

	case strings.HasPrefix(r.Ref, "edit-"):
		return s.Accent.Render(r.Ref)

	case strings.HasPrefix(r.Ref, "fork-"):
		return s.Warn.Render(r.Ref)
	}

	return r.Ref
}

func refType(r forge.Ref) string {
	switch {
	case r.IsHead:
		return "HEAD pointer"

	case r.Ref == "main":
		return "protected"

	case strings.HasPrefix(r.Ref, "edit-"):
		return "edit branch"

	case strings.HasPrefix(r.Ref, "fork-"):
		return "fork branch"
	}

	return "ref"
}

func (a *App) renderDag() string {
	s := a.styles

	refsByHash := map[string][]string{}
	for _, r := range a.refs {
		if r.IsHead {
			refsByHash[r.Hash] = append(refsByHash[r.Hash], "HEAD,"+r.Target)
			continue
		}
		refsByHash[r.Hash] = append(refsByHash[r.Hash], r.Ref)
	}

	type row struct {
		lane string
		hash string
	}

	// Build DAG rows from messages if available.
	var rows []row
	if len(a.messages) > 0 {
		for _, msg := range a.messages {
			rows = append(rows, row{"*  ", msg.Hash})
		}
	}

	b := strings.Builder{}
	for _, r := range rows {
		laneStyled := s.Accent.Render(r.lane)
		if r.hash == "" {
			laneStyled = s.Faint.Render(r.lane)
		}
		b.WriteString(laneStyled)
		if r.hash == "" {
			b.WriteString("\n")
			continue
		}

		var meta forge.Message
		for _, mm := range a.messages {
			if mm.Hash == r.hash {
				meta = mm
				break
			}
		}

		b.WriteString(" " + s.Faint.Render(shortHash(r.hash, 8)) + "  ")
		for _, rn := range refsByHash[r.hash] {
			for _, name := range strings.Split(rn, ",") {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				if strings.HasPrefix(name, "HEAD") {
					b.WriteString(s.ChipAcc.Render(name) + " ")
				} else if name == "main" {
					b.WriteString(s.Info.Render(name) + " ")
				} else if strings.HasPrefix(name, "fork-") {
					b.WriteString(s.Warn.Render(name) + " ")
				} else {
					b.WriteString(s.Chip.Render(name) + " ")
				}
			}
		}

		role := string(meta.Role)
		if role != "" {
			b.WriteString(s.RoleStyle(role).Render(role))
		}

		b.WriteString("\n")
	}

	return b.String()
}
