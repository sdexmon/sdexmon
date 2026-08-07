# WARP.md

Guidance for WARP (warp.dev) when working in this repository.

Detailed documentation lives in `docs/`. This file is the orientation layer plus
the things that are easy to get wrong. When they disagree, the code wins, and
whichever document was wrong should be fixed.

## Overview

SDEXMON is a terminal UI (Go, Bubble Tea, Lip Gloss) for monitoring Stellar spot
markets. It is read-only: it submits no transactions and holds no keys.

Module path: `github.com/sdexmon/sdexmon`

Capabilities:

- Live order books, trades, and liquidity pool metrics for a selected pair.
- Liquidity pool exposure panels for both sides of the pair.
- A pair selector popup with search, over pairs from the user's config file.
- In-app pair maintenance: add, remove and list, with trust-aware asset lookup
  via SEP-1 `stellar.toml`, fuzzy stellar.expert search, or the Top 50 list.
- A startup release check with an advisory upgrade notice and in-app upgrade.
- An unattended `display` mode with pair rotation, for kiosk boards.

## Commands

```bash
./run                    # dev launcher: .env, DEBUG=true, 140x60, version flags
go run ./cmd/sdexmon     # without the launcher

make build               # go build -o sdexmon ./cmd/sdexmon
make fmt                 # go fmt ./...
make vet                 # go vet ./...
make test                # go test ./...
make readme-gen          # regenerate README.ansi from README.md
make readme-check        # fail if README.ansi has drifted

go test -race ./...              # what CI runs
go test -run '^TestName$' ./...  # single test
go mod tidy                      # after module changes
```

Release builds inject version information; plain builds report
`dev (build unknown)`:

```bash
go build -ldflags="\
  -X main.appVersion=$(git describe --tags --always) \
  -X main.gitCommit=$(git rev-parse --short HEAD)" \
  -o sdexmon ./cmd/sdexmon
```

`sdexmon --version` prints `<version> (build <commit>)` and exits.

IMPORTANT: `README.ansi` is generated. After editing `README.md`, run
`make readme-gen` and commit both files, or `make readme-check` will fail.

## Repository layout

```
cmd/sdexmon/
  main.go                 entry point, model, update, views, Horizon calls
  maintenance_update.go   maintenance key handling and commands
  maintenance_view.go     maintenance screen renderers
  *_test.go               layout, order book request, maintenance, upgrade
internal/
  config/                 YAML config, asset parsing, pair CRUD, debug logger
  models/maintenance.go   maintenance state machine types
  models/types.go         DEAD CODE, see below
  models/constants.go     DEAD CODE, see below
  stellar/                stellar.toml resolver, stellar.expert client
  ui/upgrade.go           upgrade notice renderer
  version/checker.go      GitHub release check and semver comparison
docs/                     documentation, see docs/README.md
packaging/systemd/        sdexmon.service unit
scripts/gen-readme-ansi.py
```

## Environment variables

- `HORIZON_URL` -- Horizon endpoint. Default `https://horizon.stellar.org`.
- `DEBUG` -- `true` or `1` enables debug mode.
- `BASE_ASSET`, `QUOTE_ASSET` -- `native`, `XLM`, `XLM:native`, or
  `CODE:ISSUER`. These preselect the highlighted pair. They only cause a jump
  straight to Pair Info in `display` mode.
- `LP_POOL_ID` -- force a specific pool, applied to every pair. Debugging aid.

The `run` script loads `.env` if present.

## Screens and keys

Screens: Landing, Pair Info, Pair Debug, Pair Input, Upgrade, Maintenance. The
pair selector is a popup overlay, not a screen.

```
Landing      enter: pair selector   m: maintenance   u: upgrade*   q: quit
Selector     up/down (k/j)  enter: select  s: search  esc: close   q: quit
Pair Info    p: selector  d: debug detail  m: maintenance  u: upgrade*  q: quit
Pair Debug   d: back   u: upgrade*   q: quit
Upgrade      enter: run installer   esc: back   q: quit
Maintenance  1: add   2: remove   3: list   esc: back   q: quit
```

`*` The `u` shortcut exists and is advertised only while the startup check found
a newer release.

`q` and `ctrl+c` quit everywhere, except that `q` types normally in a
maintenance text field (`AcceptsTextInput`), where only `ctrl+c` quits.

Full reference: `docs/guide/usage.md`. State machine:
`docs/architecture/screens-and-routing.md`.

## Trading pairs

IMPORTANT: `~/.config/sdexmon/config.yaml` is the source of truth at runtime.
`loadConfiguration()` in `cmd/sdexmon/main.go` builds `configuredPairs` and
`liquidityPoolIDs` from its `pairs:` entries.

The `curatedPairs`, `curatedAssets` and `fallbackLiquidityPoolIDs` tables in
`cmd/sdexmon/main.go` are fallbacks only, used when the config fails to load or
yields no usable pairs.

IMPORTANT: `internal/models/constants.go` is NOT in the load path. Adding pairs
there has no effect.

Prefer the in-app maintenance screen (`m`) over hand editing: it validates
issuers, rejects duplicates, and reloads immediately without a restart.

Details: `docs/guide/pair-management.md` and
`docs/architecture/configuration-system.md`.

### Asset trust

IMPORTANT: an asset code proves nothing about who issued it. Every result row
shows its home domain, and the selection and confirmation screens spell out the
issuer.

Domain search reads `https://<domain>/.well-known/stellar.toml` and lists only
what that domain publishes about itself. It is the only authoritative source.
Entries with `status = "dead"`, a `code_template`, or an invalid issuer are
dropped; per-currency `toml` links are followed.

IMPORTANT: do NOT reintroduce the old behaviour of passing a domain to the fuzzy
stellar.expert `?search=` endpoint. It matches substrings and returned lookalike
assets from unrelated issuers. The stellar.expert fallback for an unreachable
stellar.toml filters to an exact home domain match.

## Data sources and polling

Horizon: order books (limit 200), trades (bootstrap 50 descending, then paged
from a cursor at limit 200, capped at 120 kept), liquidity pool discovery and
detail, and `/fee_stats` for the footer capacity indicator.

stellar.expert (`https://api.stellar.expert/explorer/public`): liquidity pool
fee and volume enrichment (`/liquidity-pool/<id>`, amounts in stroops), fuzzy
asset search (`/asset?search=`), and `/asset-list/top50`. Horizon is the
fallback for pool reserves when stellar.expert fails.

Intervals: order book 1200 ms, trades 1200 ms, liquidity pool 30 s, network
stats 10 s, display rotation `--rotate` (default 30 s, 0 disables).

## Rendering constraints

These are load-bearing. Breaking them produces a garbled alternate screen.

- Never write into the last terminal column. Terminals auto-wrap, which shifts
  every following line.
- Never emit more lines than the terminal has. Content is hard-truncated in
  `pairInfoView` for this reason.
- Never write to stderr while the TUI is running. `setupDebugLogger` redirects
  `log` into an in-memory ring buffer of the last 100 lines, restored on exit.

The layout targets 140x60. `fitPanelRows` shrinks the exposure list before the
order book; depth 7 down to 3, exposure 10 down to 3. Below 102 columns the
paired panels stack vertically. Content width is capped at 180.

## Stellar rules

- Amounts MUST use 7 or fewer decimal places. Display at least 2, up to 7.
- Thousand separators are spaces, not commas.
- Do not assume a currency symbol.
- The issuer is part of the asset identity. Never compare or deduplicate assets
  on code alone.
- Development runs against mainnet with the public Horizon endpoint.

## Go coding standards

- Packages: lowercase, single word. Files: snake_case.
- Exported: PascalCase. Unexported: camelCase. Interfaces: "er" suffix where it
  reads naturally.
- Handle all errors explicitly. Use `context.Context` for outbound timeouts.
- Prefer composition over inheritance.
- Run `go fmt` and `go vet` before committing.
- Comments should explain intent and constraints, not restate the code.

## Documentation rules

- Long-form docs go in `docs/<category>/`, not the repository root. Only
  `README.md`, `README.ansi`, `WARP.md` and `LICENSE` belong at the root.
- Markdown uses plain ASCII. No emoji, no Unicode arrows or dashes. Use `->`,
  `--`, `+/-`, `>=`.
- Do not create point-in-time status or summary documents. Describe the current
  state instead.
- On macOS, never use `sed` for in-place edits; BSD `sed` corrupts multi-byte
  UTF-8. Use Python with `encoding='utf-8'`.

## Known technical debt

- `cmd/sdexmon/main.go` is roughly 3 100 lines and still holds the model, the
  update loop, every view, and the Horizon integration. The remaining split
  would be `internal/stellar` for the Horizon wrapper and `internal/ui` for the
  views, styles and formatting.
- `internal/models/types.go` and `internal/models/constants.go` are dead code.
  Nothing imports `models.Model`, `models.ScreenState`, `models.CuratedAssets`,
  `models.CuratedPairs` or `models.LiquidityPoolIDs`; `cmd/sdexmon/main.go`
  declares its own equivalents. Only `internal/models/maintenance.go` is live.
- `config.GetBaseAsset`, `config.GetQuoteAsset` and `config.GetLPPoolID` are
  unused; `main.go` reads those environment variables directly.
- `config.ParseAsset` accepts `XLM:native`; the local `parseAsset` in `main.go`
  does not. Two parsers with different behaviour still exist.
- The Pair Input screen has a handler and a view but no key binding routes to
  it, so it is unreachable.
- `q` still quits from the pair selector search field, so search terms
  containing `q` cannot be typed there.
- Test coverage is targeted, not broad. Missing: view snapshots and mocked
  stellar.expert searches.

## Version history

Current tags run v0.1.0 through v0.2.2. Releases v1.0.0 through v1.0.3 were
published with a broken build configuration and have been deprecated; the valid
sequence continues from v0.1.0 upward.

## Documentation map

```
docs/README.md                              index
docs/guide/installation.md                  install, wrapper, uninstall
docs/guide/configuration.md                 env vars, flags, config schema
docs/guide/usage.md                         every screen and key
docs/guide/pair-management.md               add, remove, list pairs
docs/guide/upgrading.md                     update check and upgrade paths
docs/guide/migration.md                     pre-v0.1.1 installations
docs/deployment/raspberry-pi.md             kiosk board with systemd
docs/architecture/overview.md               layout, data flow, tech debt
docs/architecture/screens-and-routing.md    state machines
docs/architecture/configuration-system.md   YAML schema, decimals
docs/development/building-and-testing.md    build, test, standards
docs/development/releasing.md               CI, GoReleaser, installer contract
```

## License

Custom non-commercial license, see `LICENSE`. Personal non-commercial use is
allowed with attribution to the original author. Commercial use, distribution
and sublicensing require prior written consent.
