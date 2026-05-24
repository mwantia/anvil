# Anvil

> [!WARNING]
> This project is in **early development** and is **not production-ready**.
> Expect broken plugins, missing features, and frequent breaking changes to config schema and APIs.
> Do not deploy in production or critical environments.
>
> This project is AI assisted - As seen by the `Co-Authored-By` tags in the commit history, as well as the GitHub's contributors tab.
> I review changes where possible, but cannot guarantee full correctness or quality done by AI in all instances.

Terminal UI for [Forge](https://github.com/mwantia/forge) — inspect and manage AI sessions, conversation logs, branches, resources, and system status from the keyboard.

## Requirements

- A running [Forge](../README.md) daemon
- Go 1.25+
- [Task](https://taskfile.dev/) (optional, wraps the build commands)

## Install

```bash
task build          # → ./build/anvil
# or
go build -ldflags '-s -w' -o ./build/anvil ./cmd/main.go
```

## Usage

```bash
anvil -address http://127.0.0.1:9280 -token <bearer-token>
```

Or via environment variables with `task run`:

```bash
export FORGE_ADDRESS=http://127.0.0.1:9280
export FORGE_TOKEN=<bearer-token>
task run
```

## Screens

| Key | Screen |
|-----|--------|
| `1` | **Sessions** — list, detail, and branch management |
| `2` | **Resources** — stored memory and archive files |
| `3` | **System** — agent health, plugins, token usage, storage |

Press `Enter` on a session to open its **Log** sub-screen.

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `1` / `2` / `3` | Switch tab |
| `q` / `Ctrl+C` | Quit |

### Sessions

| Key | Action |
|-----|--------|
| `↑` / `k`, `↓` / `j` | Select session |
| `Enter` | Open conversation log |
| `Tab` | Focus branches panel |
| `n` | New session |
| `c` | Clone session |
| `a` | Archive session |
| `x` | Delete session |
| `/` | Toggle archived filter |

### Branches panel (inside Sessions)

| Key | Action |
|-----|--------|
| `↑` / `↓` | Select ref |
| `←` / `→` | Navigate panes |
| `c` | Checkout ref |
| `b` | Create branch from ref |
| `m` | Merge into HEAD |
| `x` | Delete ref |
| `Tab` / `Esc` | Return to sessions |

### Log

| Key | Action |
|-----|--------|
| `↑` / `↓` | Walk messages |
| `Enter` | Expand / collapse message body |
| `e` | Edit-fork from this message |
| `c` | Checkout to this message hash |
| `y` | Yank message hash |
| `Esc` | Back to sessions |

### Resources

| Key | Action |
|-----|--------|
| `↑` / `↓` | Select resource |
| `←` / `→` | Cycle scope (all / session / archive / global) |
| `y` | Yank resource HEAD hash |

## Development

The `go.mod` replace directive points the `forge-sdk` dependency to `../sdk` — that sibling module must be present when building.

```bash
task setup   # go mod download && go mod tidy
task build   # compile
task run     # build + run (requires FORGE_ADDRESS and FORGE_TOKEN)
```

See [docs/DESIGN.md](docs/DESIGN.md) for screen layouts, palette reference, and planned design changes.
