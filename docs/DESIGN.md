# TUI Design Reference

This document captures the current layout, palette, and component conventions for the Anvil TUI, plus a **Design Direction** section where planned changes and open questions live. Update the direction section freely — it is the primary input for design work.

---

## Layout skeleton

> User: Let's remove `session <name>  [HEAD→main]` from the **TabBar** - We already have the same state in the **StatusBar**, creating two areas with the same content.
> We can replace this with: `forge <address> <health-state>` to indicate the current connection.

```
┌─ TermBar ────────────────────────────────────────────────────────────────────┐
│ ● anvil  forge <screen-label> · ~/forge                go 1.22 · bubbletea   │
├─ TabBar ─────────────────────────────────────────────────────────────────────┤
│   1 sessions   2 resources   3 system          session <name>  [HEAD→main]   │
│                                                                              │
│  <body — varies per screen>                                                  │
│                                                                              │
├─ KeyHints ───────────────────────────────────────────────────────────────────┤
│ ↑↓ select   enter log   tab branches   n new   c clone   …   1-3 tab  q quit │
├─ StatusBar (amber) ──────────────────────────────────────────────────────────┤
│   anvil v…   session <name>   HEAD main                            15:04:05  │
└──────────────────────────────────────────────────────────────────────────────┘
```

Chrome pieces live in `chrome.go`. Heights are computed in `app.go:View()` and the body is clipped/padded to fill exactly `height − chromeH` lines via `fitLines`.

---

## Screens

### 1 · Sessions (`ScreenSessions`)

Split layout, 55 % top / 45 % bottom:

> User: Currently `sessions [N]` and `<session-name>` aren't sized correctly (both having separate heights).
> I want to make both layouts uniform to each other and turn ``sessions [N]` into a list/table that supports groups.
> Each session can be expanded (by using `→` or `d`) to list down all existing refs for this session.
> These can be collapsed by using `←` or `d` (toggles). Using the hotkey `k` expands all sessions listed (Using `j` collapses all again)
> With this change, we should be able to remove `refs [N]`, `dag · log` and `<ref>` completely (most is now compacted into `sessions [N]`).
> The additional information about the selected ref (e.g. `<ref>`) can be added to `<session-name>` at the bottom as dedicated group or combined with the already existing info.

```
┌─ sessions [N] ─────────────────────────────────┐  ┌─ <session-name> ──────┐
│ ID           NAME    TITLE   PLUGINS MSGS UPD  │  │ ID       <hash>       │
│ ────────────────────────────────────────────── │  │ Name     <name>       │
│ <selected row> …                               │  │ Title    …            │
│  <dim row> …                                   │  │ Model    …            │
└────────────────────────────────────────────────┘  │ Parent   …            │
┌─ refs [N] ──────────┐  ┌─ dag · log ───────────┐  │ Created  …            │
│ HEAD → main         │  │ *  <hash>  HEAD,main  │  │ Updated  …            │
│ main                │  │ *  <hash>  …          │  │                       │
│ fork-abc            │  │ …                     │  │ MESSAGES · N          │
│ ──────────────────  │  └───────────────────────┘  │   user        ██░░ 3  │
│ c checkout  b branch│  ┌─ <ref> ───────────────┐  │   assistant   ███░ 4  │
│ m merge     x delete│  │ ref    …              │  │   tool_call   █░░░ 1  │
└─────────────────────┘  │ hash   …              │  │   tool_result █░░░ 1  │
                         │ type   …              │  │                       │
                         │ ACTIONS               │  │ COST                  │
                         │ c checkout  b branch  │  │ Estimated  $0.0014    │
                         └───────────────────────┘  │ Tokens  in=… out=…    │
                                                    └───────────────────────┘
```

- Left panel (70 % width): session table. Focused border when `!focusBranches`.
- Right panel (30 % width): session detail with mini-bar charts for message counts.
- Bottom: branches sub-panel, active when `focusBranches = true`. Three panes — refs list, DAG graph, ref detail — navigated with `←→` or `Tab`.

### 2 · Log (`ScreenLog`) — sub-screen of Sessions

```
┌─ log · <session> ────────────────────────────────────────────────┐
│ <name> · <model> @ <HEAD>                                        │
│ N messages   user=1 assistant=1 tool_call=1 tool_result=1        │
│ tokens: in=…  out=…  total=…  cost=$…                            │
├─ messages [N] ──────────────┐  ┌─ <hash[0:12]> ──────────────────┤
│ <hash[0:8]>  user      …    │  │ message <full-hash>  [refs]     │
│ <hash[0:8]>  assistant …    │  │ Role    user                    │
│ <selected>   tool_call …    │  │ Date    Mon Jan 2 15:04         │
│                             │  │ Tokens  in=… out=…              │
│                             │  │                                 │
│                             │  │ <preview>                       │
│                             │  │ (enter to expand · N lines)     │
└─────────────────────────────┘  └─────────────────────────────────┘
```

- Enter on a row: toggle expanded body (full `msg.Body` lines vs preview).
- `e` → EditFork on selected message hash.
- `c` → Checkout to selected message hash.

### 3 · Resources (`ScreenResources`)

Three-column layout:

```
┌─ scope ───────┐  ┌─ resources [N] ─────────────────────┐  ┌─ <resource-name> ─────┐
│  all       N  │  │ PATH              VER    SIZE SCOPE │  │ path     …            │
│  session   N  │  │ ─────────────────────────────────── │  │ scope    …            │
│  archive   N  │  │  <path>           v1   1.2kb  all   │  │ mime     …            │
│  global    N  │  │  <selected>       v2    840b sess   │  │ versions N            │
│               │  │                                     │  │ size     …            │
│ STORE         │  │                                     │  │ updated  …            │
│ backend  file │  │                                     │  │ HEAD     <hash>       │
│ objects  123  │  │                                     │  │                       │
│ refs      45  │  │                                     │  │ HISTORY · 2 of N      │
│ swept      0  │  │                                     │  │  v2  <hash>  <date> … │
└───────────────┘  └─────────────────────────────────────┘  │  v1  <hash>  <date> … │
                                                            │ SUMMARY               │
                                                            │ …                     │
                                                            └───────────────────────┘
```

Scope cycles with `←→`. Resource detail is fetched on selection change (live call, 3 s timeout).

### 4 · System (`ScreenSystem`)

Two rows:

**Row 1 — four equal tiles:**

```
┌─ agent ───────┐  ┌─ sessions ─────┐  ┌── tokens · total ─┐  ┌─ storage ─────┐
│ v0.x.x  12h   │  │ 3  live        │  │  14.2K total tok  │  │ 247  objects  │
│ http  :9280   │  │                │  │ in   8200         │  │ backend  file │
│ metr  :9500   │  │ arch 1 total 4 │  │ out  6000         │  │ refs     45   │
│ ● healthy     │  │                │  │                   │  │ swept    0    │
└───────────────┘  └────────────────┘  └───────────────────┘  └───────────────┘
```

**Row 2 — three equal columns:**

```
┌─ plugins [N] ──────┐  ┌─ recent activity ──┐  ┌─ dag · <session> ──────┐
│ [ollama] ollama v… │  │ INFO  … session  … │  │ *  <hash>  HEAD,main   │
│   ● healthy        │  │ WARN  …            │  │ *  <hash>  …           │
│ [skills] skills v… │  │ …                  │  │ …                      │
│   ● healthy        │  │ › tail —follow     │  │                        │
└────────────────────┘  └────────────────────┘  └────────────────────────┘
```

---

## Palette

| Token         | Hex       | Usage                                      |
|---------------|-----------|--------------------------------------------|
| `ColBg`       | `#161d27` | App background                             |
| `ColFg`       | `#e6e8eb` | Primary text                               |
| `ColFgDim`    | `#8a93a3` | Secondary text, inactive rows              |
| `ColFgFaint`  | `#4a5364` | Labels, borders, HR lines                  |
| `ColRule`     | `#1f2731` | HR dividers                                |
| `ColRule2`    | `#2a3340` | Empty bar fill, box borders                |
| `AccentAmber` | `#f59e0b` | Active tab, selected row marker, highlights|
| `ColOk`       | `#4ade80` | Healthy status                             |
| `ColInfo`     | `#60a5fa` | Info level, `main` branch                  |
| `ColWarn`     | `#fbbf24` | Warn level, `fork-*` branches              |
| `ColDanger`   | `#f87171` | Error level, down status                   |

Role → color mapping: `user` → Amber, `assistant` → Dim, `tool_call` → Info, `tool_result` → Faint, `system` → Warn.

---

## Components

| Component  | File           | Notes                                                                 |
|------------|----------------|-----------------------------------------------------------------------|
| `Box`      | components.go  | Rounded border, optional title. `BoxFocused` swaps border to Amber.  |
| `Row`/`RowSel` | theme.go   | Left-border glyph: `" "` (dim) vs `"▍"` (amber). Bg `#121920` on sel.|
| `Chip`/`ChipAcc` | theme.go | Inline badges. ChipAcc = Amber fg + `#0e1a24` bg.                   |
| `Hr`       | components.go  | `─` repeated, `ColRule` fg.                                           |
| `KV`       | components.go  | `Faint(padRight(label, w)) + " " + value`.                           |
| `Spark`    | components.go  | `▁▂▃▄▅▆▇█` sparkline from int slice.                                 |
| `miniBar`  | sessions.go    | Inline horizontal bar: `█` (Amber) + `░` (ColRule2).                 |
| `TermBar`  | chrome.go      | Top bar: dot + name + path left, runtime right.                       |
| `TabBar`   | chrome.go      | Numbered tabs + session/HEAD context right.                           |
| `StatusBar`| chrome.go      | Full-width Amber bg bar, bold. Flash messages prepend left side.      |
| `KeyHints` | chrome.go      | `KeyCap key  KeyHint label` pairs, spaced.                            |

---

## Design Direction

> **How to use this section:** Write what you want changed, why, and any constraints. Be as rough or as specific as you like. Claude reads this file before any TUI work and uses it to stay aligned with your intent.

### Open questions

- [ ] Should the Sessions screen split branches into a dedicated top-level tab (tab 4), or keep the current bottom sub-panel activated by `Tab`?
- [ ] Should `ScreenLog` remain a sub-screen (no tab number), or get its own tab? The current "tab 1 stays lit while in log" is a subtle convention.
- [ ] Is the amber `StatusBar` the right primary accent, or should it be a subtler dark bar with amber text only?
- [ ] Should `reloadAll()` poll on a timer, or remain manual (only reloads on navigation events)?

### Planned changes

<!-- Add items here as you identify them. Format: what, where, why. -->

- **Example:** The branches DAG pane (`renderDag`) always renders a flat `* hash role` list — no actual branching lines. Replace with a real git-log-style graph when there are divergent refs.

### Design constraints

- Terminal must be usable at 80 columns minimum (currently assumes ~140).
- No mouse dependency — all navigation must be keyboard-only.
- Colour palette should remain dark-only; no light-mode variant planned.
- Avoid adding new top-level tabs without a strong reason — the 1-3 number row is the primary navigation.
