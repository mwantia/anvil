package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (a *App) viewSystem(w int) string {
	s := a.styles

	tileW := (w - 12) / 4

	agentTile := strings.Builder{}
	agentTile.WriteString(lipgloss.NewStyle().Foreground(s.ColAccent).Bold(true).Render(a.system.Agent.Version))
	agentTile.WriteString("  " + s.Faint.Render(a.system.Agent.Uptime) + "\n")
	agentTile.WriteString(s.Faint.Render("http ") + s.Muted.Render(a.system.Agent.HTTP) + "\n")
	agentTile.WriteString(s.Faint.Render("metr ") + s.Muted.Render(a.system.Agent.Metrics) + "\n")
	agentTile.WriteString(s.OK.Render("● healthy"))

	live := 0
	archived := 0
	for _, ss := range a.sessions {
		if ss.Archived {
			archived++
		} else {
			live++
		}
	}

	sessTile := strings.Builder{}
	sessTile.WriteString(lipgloss.NewStyle().Foreground(s.ColAccent).Bold(true).Render(fmt.Sprintf("%d", live)))
	sessTile.WriteString("  " + s.Faint.Render("live\n"))
	sessTile.WriteString(s.Accent.Render(s.RenderSpark([]int{1, 2, 2, 3, 3, 4, 3, 4, 5, 4, 5, 5})) + "\n")
	sessTile.WriteString(s.Faint.Render("arch ") + s.Muted.Render(fmt.Sprintf("%d", archived)) + "  ")
	sessTile.WriteString(s.Faint.Render("total ") + s.Muted.Render(fmt.Sprintf("%d", len(a.sessions))))

	tokTile := strings.Builder{}
	tokIn, tokOut := 0, 0
	for _, ss := range a.sessions {
		tokIn += ss.TokensIn
		tokOut += ss.TokensOut
	}

	total := tokIn + tokOut
	tokLabel := fmt.Sprintf("%d", total)
	if total > 1_000_000 {
		tokLabel = fmt.Sprintf("%.1fM", float64(total)/1_000_000)
	} else if total > 1_000 {
		tokLabel = fmt.Sprintf("%.1fK", float64(total)/1_000)
	}

	tokTile.WriteString(lipgloss.NewStyle().Foreground(s.ColAccent).Bold(true).Render(tokLabel))
	tokTile.WriteString("  " + s.Faint.Render("total tokens\n"))
	tokTile.WriteString(s.Faint.Render("in  ") + s.Muted.Render(fmt.Sprintf("%d", tokIn)) + "\n")
	tokTile.WriteString(s.Faint.Render("out ") + s.Muted.Render(fmt.Sprintf("%d", tokOut)))

	storTile := strings.Builder{}
	storTile.WriteString(lipgloss.NewStyle().Foreground(s.ColAccent).Bold(true).Render(fmt.Sprintf("%d", a.system.Storage.Objects)))
	storTile.WriteString("  " + s.Faint.Render("objects\n"))
	storTile.WriteString(s.Faint.Render("backend ") + s.Muted.Render(a.system.Storage.Backend) + "\n")
	storTile.WriteString(s.Faint.Render("refs    ") + s.Muted.Render(fmt.Sprintf("%d", a.system.Storage.Refs)) + "\n")
	storTile.WriteString(s.Faint.Render("swept   ") + s.Muted.Render(fmt.Sprintf("%d", a.system.Storage.Swept)))

	tiles := lipgloss.JoinHorizontal(lipgloss.Top,
		s.RenderBox("agent", false, tileW, agentTile.String()), "  ",
		s.RenderBox("sessions", false, tileW, sessTile.String()), "  ",
		s.RenderBox("tokens · total", false, tileW, tokTile.String()), "  ",
		s.RenderBox("storage", false, tileW, storTile.String()),
	)

	colW := (w - 8) / 3

	plugins := strings.Builder{}
	for _, p := range a.system.Plugins {
		statusStyle := s.OK
		switch p.Status {
		case "degraded":
			statusStyle = s.Warn
		case "unhealthy", "down":
			statusStyle = s.Danger
		}

		kind := ""
		if len(p.Types) > 0 {
			kind = p.Types[0]
		}

		latency := FormatPrettyDuration(time.Duration(p.Latency))
		fmt.Fprintf(&plugins, "%s %-12s %s %s %s\n",
			s.Chip.Render(fmt.Sprintf("%-8s", kind)),
			p.Name,
			statusStyle.Render("●"),
			statusStyle.Render(fmt.Sprintf("%-10s", p.Status)),
			s.Faint.Render(latency),
		)

		if p.Status != "healthy" && p.Message != "" {
			msg := s.TruncateRunes(p.Message, colW-6)
			plugins.WriteString("\t" + s.Faint.Render(msg) + "\n")
		}
		if p.Action != "" {
			plugins.WriteString("\t  " + s.Faint.Render("→ ") + s.Warn.Render(s.TruncateRunes(p.Action, colW-8)) + "\n")
		}
	}

	plugBox := s.RenderBox(fmt.Sprintf("plugins [%d]", len(a.system.Plugins)), false, colW, "\n"+plugins.String())

	act := strings.Builder{}
	for _, e := range a.system.RecentLog {
		lvlStyle := s.Faint
		switch e.Level {
		case "WARN":
			lvlStyle = s.Warn

		case "ERROR":
			lvlStyle = s.Danger

		case "INFO":
			lvlStyle = s.Info
		}

		fmt.Fprintf(&act, "%s %s %s %s\n",
			lvlStyle.Render(fmt.Sprintf("%-5s", e.Level)),
			s.Faint.Render(e.Time),
			s.Accent.Render(fmt.Sprintf("%-8s", e.Source)),
			s.Muted.Render(s.TruncateRunes(e.Message, colW-26)),
		)
	}
	act.WriteString("\n" + s.Accent.Render("›") + " " + s.Faint.Render("tail —follow"))
	actBox := s.RenderBox("recent activity", false, colW, act.String())

	dagBox := s.RenderBox("dag · "+a.activeSession().Name, false, colW, a.renderMiniDag())
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, plugBox, "  ", actBox, "  ", dagBox)

	return tiles + "\n\n" + row2
}

func FormatPrettyDuration(d time.Duration) string {
	switch {
	case d >= time.Second:
		return fmt.Sprintf("%.2fs", d.Seconds())

	case d >= time.Millisecond:
		return fmt.Sprintf("%.2fms", float64(d)/float64(time.Millisecond))

	case d >= time.Microsecond:
		return fmt.Sprintf("%.2fµs", float64(d)/float64(time.Microsecond))

	default:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
}

func (a *App) renderMiniDag() string {
	s := a.styles
	if len(a.messages) == 0 {
		return s.Faint.Render("no messages loaded")
	}

	refsByHash := map[string][]string{}
	for _, r := range a.refs {
		if r.IsHead {
			refsByHash[r.Hash] = append(refsByHash[r.Hash], "HEAD")
			continue
		}

		refsByHash[r.Hash] = append(refsByHash[r.Hash], r.Ref)
	}

	var lines []string
	for _, msg := range a.messages {
		role := string(msg.Role)
		var line strings.Builder
		line.WriteString(s.Accent.Render("*") + "  " + s.Faint.Render(shortHash(msg.Hash, 8)) + "  ")

		for _, rn := range refsByHash[msg.Hash] {
			if rn == "main" {
				line.WriteString(s.Info.Render(rn) + " ")
			} else if strings.HasPrefix(rn, "fork-") {
				line.WriteString(s.Warn.Render(shortHash(rn, 12)) + " ")
			} else {
				line.WriteString(s.ChipAcc.Render(shortHash(rn, 14)) + " ")
			}
		}

		line.WriteString(s.RoleStyle(role).Render(role))
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}
