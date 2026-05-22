// Package forge models the domain objects the anvil TUI reads from a Forge
// daemon: Sessions, Messages (the DAG), Refs, Resources, and a System report.
package forge

import "time"

// Session is a long-running interaction with a model + plugin set.
type Session struct {
	ID      string
	Name    string
	Title   string
	Desc    string
	Plugins string
	Model   string
	Parent  string
	Created time.Time
	Updated time.Time

	Messages  int
	TokensIn  int
	TokensOut int
	Cost      float64

	Archived bool

	Counts map[string]int
}

// Message is a single node in the session's Merkle DAG.
type Message struct {
	Hash    string
	Role    Role
	Date    time.Time
	Refs    []string
	TokIn   int
	TokOut  int
	Preview string
	Body    []string
}

// Role classifies a message node.
type Role string

const (
	RoleUser       Role = "user"
	RoleAssistant  Role = "assistant"
	RoleToolCall   Role = "tool_call"
	RoleToolResult Role = "tool_result"
	RoleSystem     Role = "system"
)

// Ref points a name at a message hash. HEAD is special — it's a pointer to
// another ref (the "target") rather than directly to a hash.
type Ref struct {
	Ref    string
	Target string
	Hash   string
	IsHead bool
}

// Resource is a versioned content blob attached to a session, archive, or
// the global namespace.
type Resource struct {
	Path     string
	Name     string
	Scope    Scope
	MIME     string
	Versions int
	Size     string
	Updated  time.Time
	Summary  string

	HEAD    string
	History []ResourceRev
}

type ResourceRev struct {
	Version int
	Hash    string
	Date    time.Time
	Delta   string
}

// Scope of a resource determines which lifecycle owns it.
type Scope string

const (
	ScopeSession Scope = "session"
	ScopeArchive Scope = "archive"
	ScopeGlobal  Scope = "global"
)

// System is the report shown on the dashboard screen.
type System struct {
	Agent struct {
		Version string
		Uptime  string
		HTTP    string
		Metrics string
	}
	Storage struct {
		Backend string
		Path    string
		Objects int
		Refs    int
		Swept   int
	}
	Plugins   []Plugin
	RecentLog []LogLine
}

type Plugin struct {
	Name, Kind, Version string
	Status              string
}

type LogLine struct {
	Level, Time, Source, Message string
}
