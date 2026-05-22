package forge

import (
	"context"
	"errors"
	"time"
)

// Fixture is an in-memory Client for running anvil without a live daemon.
type Fixture struct {
	sessions  []Session
	messages  map[string][]Message
	refs      map[string][]Ref
	resources []Resource
	system    System
}

// NewFixture returns a Client preloaded with demo data.
func NewFixture() *Fixture {
	f := &Fixture{
		messages: map[string][]Message{},
		refs:     map[string][]Ref{},
	}
	f.seed()
	return f
}

func mustParse(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		panic(err)
	}
	return t
}

func (f *Fixture) seed() {
	f.sessions = []Session{
		{
			ID: "73d5176296fa51228a2f28881f194c37", Name: "test",
			Title: "Infrastructure State", Plugins: "all",
			Model:   "forge/prometheus",
			Created: mustParse("2026-05-16 22:38:53"), Updated: mustParse("2026-05-21 15:59:09"),
			Messages: 35, TokensIn: 1190516, TokensOut: 16335,
			Counts: map[string]int{
				"user": 3, "assistant": 6, "tool_call": 3, "tool_result": 22, "system": 1,
			},
		},
		{
			ID: "3b3eed21de61dc76471051681fff294a", Name: "discord",
			Plugins: "all", Model: "forge/prometheus",
			Created: mustParse("2026-05-15 10:36:39"), Updated: mustParse("2026-05-21 11:23:04"),
			Messages: 12, TokensIn: 84320, TokensOut: 1844,
			Counts: map[string]int{"user": 4, "assistant": 4, "tool_call": 1, "tool_result": 2, "system": 1},
		},
		{
			ID: "556a1991dccbc3ff2fe103ca4579612b", Name: "monitoring",
			Plugins: "all", Model: "forge/prometheus",
			Created: mustParse("2026-05-11 10:03:16"), Updated: mustParse("2026-05-20 18:47:30"),
			Messages: 89, TokensIn: 2541002, TokensOut: 30188,
			Counts: map[string]int{"user": 9, "assistant": 14, "tool_call": 11, "tool_result": 54, "system": 1},
		},
		{
			ID: "ae12bd44f0091ab9f7ea223adf118c0e", Name: "demo-resumed",
			Title: "PR #247 · agent-registry refactor",
			Desc:  "cloned from demo @ a29cae3e",
			Plugins: "skills,consul", Model: "forge/assistant",
			Parent:  "demo (archived)",
			Created: mustParse("2026-05-20 09:11:02"), Updated: mustParse("2026-05-20 14:32:18"),
			Messages: 18, TokensIn: 220411, TokensOut: 5022, Cost: 0.0184,
			Counts: map[string]int{"user": 5, "assistant": 5, "tool_call": 4, "tool_result": 3, "system": 1},
		},
		{
			ID: "f47ac10b58cc4372a5670e02b2c3d479", Name: "rag-eval-7B",
			Plugins: "all", Model: "forge/eval",
			Created: mustParse("2026-05-19 12:01:08"), Updated: mustParse("2026-05-19 14:22:55"),
			Messages: 42, TokensIn: 988421, TokensOut: 9018, Cost: 0.0042, Archived: true,
			Counts: map[string]int{"user": 7, "assistant": 8, "tool_call": 6, "tool_result": 20, "system": 1},
		},
		{
			ID: "8c2e1b04a91706f9e23dca7708211e3f", Name: "skills-debug",
			Title: "tool-call debugging",
			Desc:  "isolated repro for skills__execute_script truncation",
			Plugins: "skills", Model: "forge/prometheus",
			Created: mustParse("2026-05-18 08:35:14"), Updated: mustParse("2026-05-18 17:50:22"),
			Messages: 27, TokensIn: 412309, TokensOut: 4811,
			Counts: map[string]int{"user": 6, "assistant": 7, "tool_call": 4, "tool_result": 9, "system": 1},
		},
	}

	test := f.sessions[0].ID
	f.messages[test] = []Message{
		{
			Hash: "a29cae3eb772", Refs: []string{"HEAD", "edit-79b09eb1"},
			Role: RoleAssistant, Date: mustParse("2026-05-21 15:59:09"),
			TokIn: 42066, TokOut: 642,
			Preview: "The data has arrived. Let me distill the atmospheric truth for Berlin.",
			Body: []string{
				"The data has arrived. Let me distill the atmospheric truth for Berlin.",
				"", "---", "## Berlin — Current Weather", "",
				"* Temperature:    18.4°C (feels 17.9°C)",
				"* Conditions:     Partly cloudy, light breeze SW 12 km/h",
				"* Humidity:       64%",
				"* Pressure:       1014 hPa, falling",
				"", "Outlook: drizzle by 19:00, clearing overnight. Bring a layer.",
			},
		},
		{
			Hash: "7aa6bec2295f", Role: RoleToolResult,
			Date:    mustParse("2026-05-21 15:59:03"),
			Preview: "stdout 41 lines · skills__execute_script · exit=0",
		},
		{
			Hash: "92cab7f10934", Role: RoleToolCall,
			Date: mustParse("2026-05-21 15:59:01"),
			TokIn: 31089, TokOut: 31,
			Preview: "→ skills__execute_script(args=[Berlin], script=scripts/weather.sh, skill=get-weather)",
		},
		{
			Hash:    "1dc7bfcb9fe5", Role: RoleUser,
			Date:    mustParse("2026-05-21 15:58:56"),
			Preview: "Use the 'get-weather' skill to get data from Berlin",
		},
		{
			Hash: "3f9bd71a04c8", Role: RoleAssistant,
			Date: mustParse("2026-05-21 14:12:30"),
			TokIn: 29812, TokOut: 511,
			Preview: "Looking at the current Prometheus scrape targets, three jobs are unhealthy.",
		},
		{
			Hash: "65c45b08254d", Refs: []string{"fork-986bbfdb"},
			Role: RoleAssistant, Date: mustParse("2026-05-19 11:42:08"),
			TokIn: 28104, TokOut: 488,
			Preview: "Alternative interpretation: the alert is firing because the recording rule sees stale data.",
		},
		{
			Hash:    "4f2a01d6b88e", Role: RoleUser,
			Date:    mustParse("2026-05-19 11:41:55"),
			Preview: "try a different angle on the staging spike",
		},
		{
			Hash:    "8e3c1b09f1a2", Role: RoleToolResult,
			Date:    mustParse("2026-05-19 10:30:14"),
			Preview: "consul__kv_get config/forge/staging.hcl · 4.2 kB",
		},
		{
			Hash: "6e9af221c053", Role: RoleAssistant,
			Date: mustParse("2026-05-18 22:14:08"),
			TokIn: 22107, TokOut: 402,
			Preview: "I need the staging config first. Let me read it from consul.",
		},
		{
			Hash:    "6435124c238f", Refs: []string{"fork-d1870730"},
			Role:    RoleUser, Date: mustParse("2026-05-18 17:55:22"),
			Preview: "check what the staging cluster looked like last Friday",
		},
		{
			Hash:    "1d23519edc47", Refs: []string{"main"},
			Role:    RoleSystem, Date: mustParse("2026-05-16 22:38:53"),
			Preview: "You are Prometheus. Plugins: all.",
		},
	}

	f.refs[test] = []Ref{
		{Ref: "HEAD", Target: "edit-79b09eb1", Hash: "a29cae3eb772", IsHead: true},
		{Ref: "edit-79b09eb1", Hash: "a29cae3eb772"},
		{Ref: "fork-986bbfdb", Hash: "65c45b08254d"},
		{Ref: "fork-d1870730", Hash: "6435124c238f"},
		{Ref: "main", Hash: "1d23519edc47"},
	}

	f.resources = []Resource{
		{
			Path: "sessions/test/preferences", Name: "preferences", Scope: ScopeSession,
			MIME: "application/yaml", Versions: 4, Size: "312 B",
			Updated: mustParse("2026-05-21 09:14:22"),
			Summary: "editor + model defaults for this session",
			HEAD:    "7f2a1c08",
			History: []ResourceRev{
				{4, "7f2a1c08", mustParse("2026-05-21 09:14:00"), "+model · forge/prometheus"},
				{3, "b1c9ee44", mustParse("2026-05-18 14:02:00"), "+tool_call.allow=[skills,consul]"},
				{2, "2c08aa11", mustParse("2026-05-17 21:48:00"), "init from defaults"},
				{1, "00aa0000", mustParse("2026-05-16 22:38:00"), "create"},
			},
		},
		{
			Path: "sessions/test/notes", Name: "notes", Scope: ScopeSession,
			MIME: "text/markdown", Versions: 12, Size: "4.1 kB",
			Updated: mustParse("2026-05-20 17:55:01"),
			Summary: "researcher scratchpad — Prometheus targets, staging spike notes",
			HEAD:    "0d4e2299",
			History: []ResourceRev{
				{12, "0d4e2299", mustParse("2026-05-20 17:55:00"), "+section \"fork-986bbfdb summary\""},
				{11, "4a17f9c1", mustParse("2026-05-19 11:50:00"), "+staging-spike"},
				{10, "b22d810f", mustParse("2026-05-18 22:30:00"), "reorg"},
				{9, "7e019aac", mustParse("2026-05-18 18:01:00"), "+config kv-paths"},
			},
		},
		{
			Path: "archives/demo", Name: "envelope", Scope: ScopeArchive,
			MIME: "application/x-forge-envelope", Versions: 1, Size: "184 kB",
			Updated: mustParse("2026-05-20 14:32:18"),
			Summary: "sealed archive of `demo` session · cloned to `demo-resumed`",
			HEAD:    "99c411bd",
			History: []ResourceRev{
				{1, "99c411bd", mustParse("2026-05-20 14:32:00"), "archive demo @ a29cae3e"},
			},
		},
		{
			Path: "global/style-guide", Name: "style-guide", Scope: ScopeGlobal,
			MIME: "text/markdown", Versions: 7, Size: "8.4 kB",
			Updated: mustParse("2026-05-14 10:09:44"),
			Summary: "shared writing/tool-use guide injected into every session",
			HEAD:    "3a55f201",
			History: []ResourceRev{
				{7, "3a55f201", mustParse("2026-05-14 10:09:00"), "+tone: terse"},
				{6, "4f0b91e7", mustParse("2026-05-11 16:30:00"), "rewrite \"tool errors\""},
				{5, "11ed20aa", mustParse("2026-04-29 14:22:00"), "+pgvector usage"},
			},
		},
	}

	f.system = System{
		Plugins: []Plugin{
			{"ollama", "provider", "0.5.2", "up"},
			{"skills", "tools", "0.4.0", "up"},
			{"consul", "tools", "0.3.1", "up"},
			{"openviking", "resource", "0.2.0-rc", "degraded"},
		},
		RecentLog: []LogLine{
			{"INFO", "15:59:09", "session", "test/edit-79b09eb1 → a29cae3eb772 · in=42066 out=642"},
			{"INFO", "15:59:01", "pipeline", "tool_call skills__execute_script(args=[Berlin])"},
			{"DEBUG", "15:58:56", "pipeline", "dispatch session=test ref=edit-79b09eb1"},
			{"WARN", "15:42:11", "plugin", "openviking degraded · rt=1.82s exceeds slo=800ms"},
			{"INFO", "14:32:18", "session", "demo-resumed cloned from demo @ a29cae3e"},
		},
	}
	f.system.Agent.Version = "v0.7.2"
	f.system.Agent.Uptime = "2h 14m"
	f.system.Agent.HTTP = "127.0.0.1:9280"
	f.system.Agent.Metrics = "127.0.0.1:9500"
	f.system.Storage.Backend = "file"
	f.system.Storage.Path = "./data"
	f.system.Storage.Objects = 14823
	f.system.Storage.Refs = 47
	f.system.Storage.Swept = 312
}

func (f *Fixture) Sessions(_ context.Context) ([]Session, error) {
	out := make([]Session, len(f.sessions))
	copy(out, f.sessions)
	return out, nil
}

func (f *Fixture) Session(_ context.Context, idOrName string) (Session, error) {
	for _, s := range f.sessions {
		if s.ID == idOrName || s.Name == idOrName {
			return s, nil
		}
	}
	return Session{}, errors.New("session not found: " + idOrName)
}

func (f *Fixture) Log(_ context.Context, sessionID string) ([]Message, error) {
	if ms, ok := f.messages[sessionID]; ok {
		return ms, nil
	}
	for _, ms := range f.messages {
		return ms, nil
	}
	return nil, nil
}

func (f *Fixture) Refs(_ context.Context, sessionID string) ([]Ref, error) {
	if rs, ok := f.refs[sessionID]; ok {
		return rs, nil
	}
	for _, rs := range f.refs {
		return rs, nil
	}
	return nil, nil
}

func (f *Fixture) ResourceDetail(_ context.Context, path, name string) (Resource, error) {
	for _, r := range f.resources {
		if r.Path == path && r.Name == name {
			return r, nil
		}
	}
	return Resource{}, errors.New("resource not found: " + path + "/" + name)
}

func (f *Fixture) Resources(_ context.Context, scope string) ([]Resource, error) {
	if scope == "" || scope == "all" {
		return f.resources, nil
	}
	var out []Resource
	for _, r := range f.resources {
		if string(r.Scope) == scope {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *Fixture) System(_ context.Context) (System, error) { return f.system, nil }

func (f *Fixture) NewSession(_ context.Context, _, _ string) (Session, error)   { return Session{}, nil }
func (f *Fixture) CloneSession(_ context.Context, _, _ string) (Session, error) { return Session{}, nil }
func (f *Fixture) ArchiveSession(_ context.Context, _ string) error             { return nil }
func (f *Fixture) DeleteSession(_ context.Context, _ string) error              { return nil }
func (f *Fixture) Checkout(_ context.Context, _, _ string) error                { return nil }
func (f *Fixture) Branch(_ context.Context, _, _, _ string) error               { return nil }
func (f *Fixture) DeleteRef(_ context.Context, _, _ string) error               { return nil }
func (f *Fixture) Merge(_ context.Context, _, _ string) error                   { return nil }
func (f *Fixture) EditFork(_ context.Context, _, _, _ string) (Message, error)  { return Message{}, nil }
