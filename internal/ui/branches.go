package ui

import (
	"strings"

	"github.com/mwantia/anvil/internal/forge"
)

// refLabel renders a ref name with role-appropriate styling.
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
