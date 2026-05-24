package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up, Down, Left, Right key.Binding
	Tab, ShiftTab         key.Binding
	Enter, Esc            key.Binding

	Tab1, Tab2, Tab3 key.Binding

	New, Clone, Archive, Delete, Edit key.Binding
	Branch, Merge, Checkout, Diff     key.Binding
	Yank, Filter, Follow              key.Binding

	ExpandAll, CollapseAll key.Binding

	Help, Quit key.Binding
}

func DefaultKeys() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "back")),
		Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "fwd")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next pane")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift-tab", "prev pane")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
		Esc:      key.NewBinding(key.WithKeys("backspace"), key.WithHelp("bs", "back")),

		Tab1: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "sessions")),
		Tab2: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "resources")),
		Tab3: key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "system")),

		New:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		Clone:    key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clone/checkout")),
		Archive:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive")),
		Delete:   key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
		Edit:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit/fork")),
		Branch:   key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "branch")),
		Merge:    key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "merge")),
		Checkout: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "checkout")),
		Diff:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "expand/collapse")),
		Yank:     key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yank hash")),
		Filter:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Follow:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "follow")),

		ExpandAll:   key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "expand all")),
		CollapseAll: key.NewBinding(key.WithKeys("J"), key.WithHelp("J", "collapse all")),

		Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}
