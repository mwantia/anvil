# TUI Design Reference

This document captures the current layout, palette, and component conventions for the Anvil TUI, plus a **Design Direction** section where planned changes and open questions live. Update the direction section freely — it is the primary input for design work.

---

## Layout skeleton

```
┌─ TermBar ────────────────────────────────────────────────────────────────────┐
│ ● anvil  forge <screen-label> · ~/forge                go 1.22 · bubbletea   │
├─ TabBar ─────────────────────────────────────────────────────────────────────┤
│   1 sessions   2 resources   3 system          forge <address> <health>      │
│                                                                              │
│  <body — varies per screen>                                                  │
│                                                                              │
├─ KeyHints ───────────────────────────────────────────────────────────────────┤
│ ↑↓ select   enter log   →/d expand   ←  collapse   K/J all   …   1-3 tab  q │
├─ StatusBar (amber) ──────────────────────────────────────────────────────────┤
│   anvil v…   session <name>   HEAD main                            15:04:05  │
└──────────────────────────────────────────────────────────────────────────────┘
```

- **TabBar right side**: `forge <address> <health>` — `●` (green) when agent HTTP is reachable, `○` (faint) otherwise.
- Chrome pieces live in `chrome.go`. Heights are computed in `app.go:View()` and the body is clipped/padded to fill exactly `height − chromeH` lines via `fitLines`.

---

## Screens

### 1 · Sessions (`ScreenSessions`)

Side-by-side layout: 70 % left (session tree) / 30 % right (detail panel).

```
┌─ sessions [N] ────────────────────────────────────────────────────────────────┐  ┌─ <session-name> ──┐
│   ID         NAME           TITLE              PLUGINS MSGS  UPDATED          │  │ ID       <hash>   │
│ ───────────────────────────────────────────────────────────────────────────── │  │ Name     <name>   │
│ ▸ <id>       <name>         <title>            <plug>   12  2006-01-02 15:04  │  │ Title    …        │
│ ▾ <id>       <name>         <title>            <plug>    4  2006-01-02 15:04  │  │ Model    …        │
│     ·  <hash>       HEAD → main                                               │  │ Parent   …        │
│     ·  <hash>       main                                                      │  │ Created  …        │
│     ·  <hash>       fork-abc123                                               │  │ Updated  …        │
│ ▸ <id>       <name>         <title>            <plug>    1  2006-01-02 15:04  │  │                   │
└───────────────────────────────────────────────────────────────────────────────┘  │ MESSAGES · N      │
                                                                                   │   user     ██░ 3  │
                                                                                   │   assistant ███ 4 │
                                                                                   │   tool_call █░░ 1 │
                                                                                   │   tool_result …   │
                                                                                   │                   │
                                                                                   │ COST              │
                                                                                   │ Estimated $0.0014 │
                                                                                   │ Tokens in=… out=… │
                                                                                   │                   │
                                                                                   │ REF (when on ref) │
                                                                                   │ name  HEAD→main   │
                                                                                   │ hash  <hash>      │
                                                                                   │ type  symref      │
                                                                                   └───────────────────┘
```

**Tree behaviour:**
- `▸` = collapsed, `▾` = expanded (Amber). Ref rows are indented with `·`.
- `→` or `d` — expand selected session header; `←` — collapse (jumps to header if on a ref row first).
- `d` on an expanded header — collapse; `d` on a ref row — collapse to parent.
- `K` — expand all sessions; `J` — collapse all, cursor jumps to owning session header.
- Unselected ref rows are fully faint (`ColFgFaint #4a5364`). Selected ref row: faint bullet+hash, colored ref label via `refLabel()`.
- `Enter` on a session header — open log for HEAD. `Enter` on a ref row — open log walking that specific ref (non-HEAD refs only; HEAD rows pass empty ref string).
- Right panel shows a **REF** section appended below COST when cursor is on a ref row.

**Column widths:**
- ID, NAME, PLUGINS, MSGS, UPDATED are fixed-width. TITLE expands to fill remaining space (`titleW = innerW − 66`, min 10).
- Row style `Width(leftW−4)` sets **outer** rendered width; content area = `leftW−7` (border + padding consume 3 chars).

### 2 · Log (`ScreenLog`) — sub-screen of Sessions

```
┌─ log · <session> ────────────────────────────────────────────────┐
│ <name> · <model> @ <ref-or-HEAD>                                 │
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

- **`Enter`** — toggle expanded body (full `msg.Body` vs preview).
- **Expanded rendering** (`renderBody` in `log.go`):
  - `user` / `assistant`: rendered as **Markdown** via [glamour](https://github.com/charmbracelet/glamour), word-wrapped to panel width.
  - `tool_call` / `tool_result`: content is JSON-pretty-printed (if valid), wrapped in a ` ```json ``` ` fence, then rendered by glamour (chroma syntax highlighting).
  - Falls back to raw text on any render error.
- **`⌫` (Backspace)** — return to Sessions screen.
- `e` → EditFork on selected message hash.
- `c` → Checkout to selected message hash.
- `y` → Yank hash.
- Header shows `@ <ref>` when walking a non-HEAD branch; the ref persists across edit/checkout actions via `logState.walkRef`.

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
| `ColFgFaint`  | `#4a5364` | Labels, borders, HR lines, unselected refs |
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

| Component    | File           | Notes                                                                        |
|--------------|----------------|------------------------------------------------------------------------------|
| `Box`        | components.go  | Rounded border, optional title. `BoxFocused` swaps border to Amber.          |
| `Row`/`RowSel` | theme.go     | Left-border glyph: `" "` (dim) vs `"▍"` (amber). Bg `#121920` on sel.      |
| `Chip`/`ChipAcc` | theme.go   | Inline badges. ChipAcc = Amber fg + `#0e1a24` bg.                           |
| `Hr`         | components.go  | `─` repeated, `ColRule` fg.                                                  |
| `KV`         | components.go  | `Faint(padRight(label, w)) + " " + value`.                                   |
| `Spark`      | components.go  | `▁▂▃▄▅▆▇█` sparkline from int slice.                                         |
| `miniBar`    | sessions.go    | Inline horizontal bar: `█` (Amber) + `░` (ColRule2).                         |
| `renderBody` | log.go         | Glamour markdown renderer; JSON roles get pretty-printed + fenced code block. |
| `TermBar`    | chrome.go      | Top bar: dot + name + path left, runtime right.                               |
| `TabBar`     | chrome.go      | Numbered tabs + `forge <address> <health>` right.                             |
| `StatusBar`  | chrome.go      | Full-width Amber bg bar, bold. Flash messages prepend left side.              |
| `KeyHints`   | chrome.go      | `KeyCap key  KeyHint label` pairs, spaced.                                    |

---

## Design Direction

> **How to use this section:** Write what you want changed, why, and any constraints. Be as rough or as specific as you like. Claude reads this file before any TUI work and uses it to stay aligned with your intent.

### Open questions

- [ ] Should `ScreenLog` remain a sub-screen (no tab number), or get its own tab? The current "tab 1 stays lit while in log" is a subtle convention.
- [ ] Is the amber `StatusBar` the right primary accent, or should it be a subtler dark bar with amber text only?
- [ ] Should `reloadAll()` poll on a timer, or remain manual (only reloads on navigation events)?

### Planned changes

<!-- Add items here as you identify them. Format: what, where, why. -->

### Design constraints

- Terminal must be usable at 80 columns minimum (currently assumes ~140).
- No mouse dependency — all navigation must be keyboard-only.
- Colour palette should remain dark-only; no light-mode variant planned.
- Avoid adding new top-level tabs without a strong reason — the 1-3 number row is the primary navigation.
