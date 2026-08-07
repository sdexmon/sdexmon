# Architecture overview

SDEXMON is a single-binary Bubble Tea TUI. It has no server component, no
database, and no persistent state beyond one YAML config file.

## Module and layout

Module path: `github.com/sdexmon/sdexmon`

```
cmd/sdexmon/
  main.go                 entry point, model, update, views, Horizon calls
  maintenance_update.go   maintenance key handling and commands
  maintenance_view.go     maintenance screen renderers
  layout_test.go          panel-fitting tests
  main_test.go            order book request shape tests
  maintenance_test.go     maintenance routing and removal tests
  upgrade_test.go         upgrade notice tests
internal/
  config/
    config.go             Horizon client, debug logging, YAML config
    assets.go             asset parsing and formatting
    user_config.go        add, remove and list configured pairs
    user_config_test.go
  models/
    maintenance.go        maintenance state machine types
    types.go              unused, see technical debt
    constants.go          unused, see technical debt
  stellar/
    toml.go               SEP-1 stellar.toml resolver (domain search)
    expert.go             stellar.expert fuzzy search and Top 50
    confirmation.go       market data for the pair confirmation screen
    toml_test.go
  ui/
    upgrade.go            upgrade notice renderer
  version/
    checker.go            GitHub release check and semver comparison
    checker_test.go
packaging/systemd/         sdexmon.service unit
scripts/gen-readme-ansi.py README.ansi generator
```

## Startup sequence

1. `main` handles `--version` first and exits without starting the TUI.
2. A leading `display` subcommand sets display mode, then the remaining
   arguments are parsed as flags.
3. Unless `--no-update-check` is set, `version.CheckForUpdate` queries the
   GitHub releases API with a 5 second timeout. Failures are logged, not fatal.
4. `loadConfiguration` reads `~/.config/sdexmon/config.yaml` and populates the
   `configuredPairs` slice and the `liquidityPoolIDs` map. See
   [the configuration system](configuration-system.md).
5. `BASE_ASSET` and `QUOTE_ASSET` are parsed if set. Invalid values are logged
   and ignored.
6. In display mode, the pair is resolved from `--pair`, or the first available
   pair. Failure here is fatal, so a misconfigured unattended board fails
   loudly instead of showing the wrong market.
7. `initialModel` builds the model and picks the initial screen.
8. `setupDebugLogger` redirects `log` output into an in-memory buffer for the
   lifetime of the TUI, so a failing poll cannot corrupt the alternate screen.
   The logger is restored on exit.

## Data sources

### Horizon (REST)

Configured by `HORIZON_URL`, default `https://horizon.stellar.org`.

- Order book: `OrderBook` with the base as selling and the quote as buying,
  limit 200. A single canonical request returns both sides; a reverse-direction
  request describes the same offers and must not be merged in.
- Trades: `Trades` with base and counter assets. The first call bootstraps with
  the 50 most recent in descending order and reverses them so the newest is
  last. Subsequent calls page forward from the stored cursor with limit 200.
  The buffer is capped at 120 trades.
- Liquidity pools: `LiquidityPools` by reserves to discover a pool ID, and
  `LiquidityPoolDetail` as the reliable source for current reserves.
- Network stats: `GET {HORIZON_URL}/fee_stats`, using `ledger_capacity_usage`
  for the footer indicator.

### stellar.expert

Base URL: `https://api.stellar.expert/explorer/public`

- `/liquidity-pool/<id>` enriches pool data with 1-day and 7-day fees and
  volume. Amounts come back in stroops, so they are always scaled by 7 decimals.
  Horizon is the fallback when this call fails, in which case the fee and volume
  columns render as `--`.
- `/asset?search=<term>&limit=50` backs the fuzzy asset search, and the fallback
  for a domain with no reachable stellar.toml (then filtered to an exact home
  domain match).
- `/asset-list/top50` backs the Top 50 source.

### SEP-1 stellar.toml

`https://<domain>/.well-known/stellar.toml` is the authoritative source for
domain search, since only the domain owner can publish an issuer there. Reads
are capped at 100 KiB with a 10 second timeout, and up to 12 per-currency `toml`
links are followed.

## Polling

Bubble Tea ticks drive all refreshes. Each tick handler re-arms its own timer.

```
order book     1200 ms
trades         1200 ms
liquidity pool 30 s
network stats  10 s
display rotate --rotate, default 30 s, 0 disables
```

Order book and trades tickers start when a pair is selected. The network stats
ticker starts in `Init` and runs on every screen. The liquidity pool ticker
re-arms itself once triggered.

## Rendering

Lip Gloss handles all styling. The Pair Info screen is laid out as three rows:

- Row 1: ORDER BOOK and TRADES side by side.
- Row 2: LIQUIDITY POOL full width.
- Row 3: base asset exposure and quote asset exposure side by side.

`fitPanelRows` sizes the variable-height panels to the terminal, shrinking the
exposure list before the order book. Depth ranges from 7 down to 3 rows per
side; exposure ranges from 10 down to 3 rows. Content is hard-truncated to the
terminal height, because overflow scrolls the alternate screen and leaves a
garbled frame behind. The last terminal column is deliberately left unwritten to
avoid auto-wrap.

Below 102 columns the paired panels stack vertically instead.

## Number formatting

Stellar amounts never exceed 7 decimal places. Displays show at least 2.

Order book prices always render with 7 decimals for granularity. Amounts and
totals use the resolved per-pair precision. Thousand separators are spaces, not
commas.

## Known technical debt

`cmd/sdexmon/main.go` is roughly 3 100 lines and still holds the model, the
update loop, every view, and the Horizon integration. The extraction that has
happened so far moved configuration, maintenance types, the Stellar helpers, the
upgrade renderer, and the version checker into `internal/`. The remaining split
would be `internal/stellar` for the Horizon client wrapper and `internal/ui` for
the views, styles and number formatting.

`internal/models/types.go` and `internal/models/constants.go` are dead code.
Nothing imports `models.Model`, `models.ScreenState`, `models.CuratedAssets`,
`models.CuratedPairs` or `models.LiquidityPoolIDs`. `cmd/sdexmon/main.go`
declares its own `model`, `screenState`, `Liquidity`, `curatedAssets`,
`curatedPairs` and `fallbackLiquidityPoolIDs`. Only `internal/models/maintenance.go`
is live. Do not add pairs to `internal/models/constants.go` expecting them to
appear; they will not.

`config.GetBaseAsset`, `config.GetQuoteAsset` and `config.GetLPPoolID` are also
unused. `cmd/sdexmon/main.go` reads those environment variables directly.

The custom pair input screen has a handler and a view but no key binding routes
to it, so it is currently unreachable.

Test coverage is targeted rather than broad. Covered: the upgrade notice,
maintenance routing and removal, search source selection, stellar.toml parsing
and validation, config pair CRUD, version comparison, order book request shape,
and panel layout. Not covered: view snapshots and mocked stellar.expert searches.
