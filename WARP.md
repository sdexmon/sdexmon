# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Overview

Terminal UI (Go, Bubble Tea/Lip Gloss) for visualizing Stellar spot markets. Features:
- Asset pair monitoring: order books, trades, and liquidity pools
- Navigation-based routing with pair selection landing page
- Polls Horizon for order books/trades; fetches LP metrics from stellar.expert
- Defaults to curated asset pairs, 140×60 layout, and 2–7 decimal rendering
- **Automatic version checking**: Checks for updates on startup and shows an advisory upgrade notice with a `u: upgrade` shortcut
- **Pair maintenance in-app**: `m` opens a maintenance screen to add, remove and list pairs, persisted to `~/.config/sdexmon/config.yaml`

Key files:
- `main.go`: entire TUI (~2085 lines) containing routing, model/update/view, Horizon calls, LP fetch, key bindings
- `run`: convenience launcher script that sets safe defaults (Horizon URL, debug mode, terminal size) and runs `go run .`
- `install.sh`: installer script that creates wrapper for proper environment setup
- `go.mod`: dependency manifest (Bubble Tea, Lip Gloss, Stellar Go SDK)
- `docs/ROUTING_IMPLEMENTATION.md`: detailed routing system documentation
- `docs/MIGRATION.md`: guide for users upgrading from pre-wrapper installations
- `cmd/sdexmon/maintenance_update.go`, `cmd/sdexmon/maintenance_view.go`: pair maintenance screen (add/remove/list)
- `internal/config/user_config.go`: reads and writes pairs in the YAML config
- `docs/MAINTENANCE_MODE.md`: legacy maintenance mode notes
- `.env`: local environment variables (not tracked in git)
- `tui`: compiled binary

## Commands

- Quick start (recommended for development):
  ```bash
  ./run
  ```
  Sets `HORIZON_URL` to public Stellar Horizon, enables debug, adjusts terminal size, and executes `go run .`.

- Install for production use:
  ```bash
  curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
  ```
  Installs binary as `.sdexmon-bin` and creates wrapper script `sdexmon` that:
  - Sets `DEBUG=true` by default
  - Configures optimal terminal size (140×60)
  - Sets default Horizon URL
  - Runs the actual binary

- Run without the helper script:
  ```bash
  go run .
  ```

- Build binary:
  ```bash
  go build -o sdexmon ./cmd/sdexmon
  ```
  Then run with `./sdexmon`.

- Build with version info:
  ```bash
  go build -o sdexmon -ldflags="-X main.appVersion=$(git describe --tags --always) -X main.gitCommit=$(git rev-parse --short HEAD)" ./cmd/sdexmon
  ```
  `appVersion` and `gitCommit` default to `dev`/`unknown` for plain `go build`.
  Release builds get real values injected by GoReleaser (see `.goreleaser.yml`).

- Check version:
  ```bash
  ./sdexmon --version
  ```

- Format and basic lint:
  ```bash
  go fmt ./...
  go vet ./...
  ```

- Tests:
  - All tests (none exist yet; for when tests are added):
    ```bash
    go test ./...
    ```
  - Single test by name:
    ```bash
    go test -run '^TestName$' ./...
    ```

- Dependency tidy (useful after module changes):
  ```bash
  go mod tidy
  ```

## Environment

These environment variables are read at runtime:
- **Horizon REST**
  - `HORIZON_URL`: Horizon endpoint for REST reads (order books, trades). Defaults to `https://horizon.stellar.org` (public mainnet).
- **Default pair** (optional, allows skipping service selection)
  - `BASE_ASSET`, `QUOTE_ASSET`: `native` or `CODE:ISSUER` (e.g., `USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN`). If set, app starts directly at Pair Info screen.
- **Liquidity pool** (optional)
  - `LP_POOL_ID`: Force specific pool ID (otherwise auto-resolved from liquidityPoolIDs map)
- **Debug**
  - `DEBUG`: Set to `true` or `1` to enable debug mode with extra logging and `z` key to toggle debug screens

**Note**: The `run` script automatically loads `.env` if present.

## Architecture and data flow

- Bubble Tea program in `main.go`
  - **Routing**: State machine with 6 screens (Landing, Pair Info, Pair Debug, Pair Input, Upgrade, Maintenance)
  - **Model** holds: current screen, selected assets, Horizon order book/trades, trade cursor, LP metrics, UI state, version info
  - **Startup**:
    - `main()` checks the GitHub API for the latest release and passes the result into the model
    - If a newer release exists (and not in `--display` mode), the app opens on the Upgrade screen; `esc` continues into the app
    - When base/quote are set, schedules three tickers (order book, trades, LP)
  - **Update**: Screen-based navigation state machine
    - Upgrade: `enter` runs the installer via `tea.ExecProcess` (the TUI quits afterwards since the binary was replaced), `esc` returns to the previous screen, `q`/`ctrl+c` always quits
    - Landing: Displays sdexmon ASCII art with version and commit info + pair selector popup
    - Pair screens: Horizon polling via `fetchOrderbookCmd`, `fetchTradesCmd`, `resolveAndFetchLPCmd`
    - Maintenance: sub-state machine in `maintenance_update.go`; async asset search and market lookups arrive as `models.AssetSearchResultsMsg` / `models.ConfirmationDataMsg` / `models.MaintenanceErrMsg` and are routed from the top-level `Update`
  - **View**: Router switches on currentScreen to render appropriate view
    - Upgrade: Centered amber notice box with version info and upgrade instructions
    - Landing: sdexmon ASCII branding with version display (top-left)
    - All other screens: SCAR AQUILA header, subtitle, content, context-aware footer
    - Pair Info: Three panels (Order Book, Trades, Liquidity Pool) + Exposure panels

## Navigation Flow

```
./run -> Landing (with Pair Selector Popup)
         |- Select Pair -> Pair Info <-> Pair Debug
         |- Custom Input -> Pair Info <-> Pair Debug
         |- m -> Maintenance -> Add / Remove / List
```

## UI Controls

The `u: upgrade` shortcut is only shown and only active while the startup check
found a newer release. It is available on the Landing, Pair Info, and Pair Debug
screens.

### Landing Screen
- `enter` (⏎): Open pair selector popup
- `m`: Open pair maintenance
- `u`: Open upgrade notice (only when an update is available)
- `q`: Quit

### Pair Selector Popup (from Landing)
- `↑/↓`: Navigate pairs
- `enter`: Select pair (start monitoring)
- `esc`: Close popup
- `q`: Quit

### Pair Input (Custom Entry)
- `tab`: Switch base/quote fields
- `enter`: Apply and start monitoring
- `esc`: Back to landing
- `q`: Quit

### Pair Info
- `p`: Open pair selector popup
- `d`: Toggle debug detail view
- `m`: Open pair maintenance
- `u`: Open upgrade notice (only when an update is available)
- `q`: Quit

### Pair Debug Detail
- `d`: Back to pair info
- `u`: Open upgrade notice (only when an update is available)
- `q`: Quit

### Upgrade Notice
- `enter`: Run the installer now (app exits afterwards; restart to use the new build)
- `esc`: Continue on the current version
- `q`: Quit

### Maintenance (from Landing or Pair Info with `m`)
- `1`: Add asset pair (domain search -> pick asset A -> pick asset B -> confirm)
- `2`: Remove asset pair (pick pair -> confirm with `y`/`n`)
- `3`: View configured pairs (read-only, `↑/↓` to scroll)
- `esc`: Back one step, and back to Landing from the menu
- `q`: Quit, except while a domain field has focus (use `ctrl+c` there)

## Trading Pairs Management

**IMPORTANT:** `~/.config/sdexmon/config.yaml` is the source of truth at runtime.
`loadConfiguration()` builds `configuredPairs` and `liquidityPoolIDs` from its
`pairs:` entries. The `curatedPairs` / `curatedAssets` / `fallbackLiquidityPoolIDs`
tables in `cmd/sdexmon/main.go` are only fallbacks, used when the config fails to
load or contains no usable pairs. `internal/models/constants.go` is **not** in the
load path.

### In-app (preferred)

Press `m` on the Landing or Pair Info screen:
- **Add**: enter a domain, pick asset A and asset B from the stellar.expert
  results, review the market summary, then confirm. Requires network access, and
  only finds assets published under a home domain, so native XLM cannot be added
  this way.
- **Remove**: pick a pair and confirm. Matching is issuer-aware and orientation
  agnostic, so BASE/QUOTE and QUOTE/BASE both resolve to the same entry.
- **List**: read-only view of every configured pair with issuers and pool IDs.

Both mutations write the file and reload it immediately, so the pair selector
updates without a restart.

### By hand

Edit `~/.config/sdexmon/config.yaml`:
```yaml
pairs:
  - name: XLM/USDC
    base: XLM:native
    quote: USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN
    lp: a468d41d8e9b8f3c7209651608b74b7db7ac9952dcae0cdf24871d1d9c7b0088
    favorite: true
    show_decimals: 7
```
`base`/`quote` accept `native`, `XLM:native` or `CODE:ISSUER`. `lp` is optional;
when omitted the pool is resolved from Horizon at runtime.

### Finding Required Data

- **Asset Issuers**: Use stellar.expert asset search
- **Liquidity Pool IDs**: Use stellar.expert liquidity pools section
- **Validation**: Asset codes must be 1-12 chars A-Z/0-9, issuer addresses 56 chars starting with 'G'

## Data Sources

- **User config** (`~/.config/sdexmon/config.yaml`): the pairs actually shown in
  the selector, plus per-pair LP IDs, favourites and decimal preferences
- **Fallback tables** (in `cmd/sdexmon/main.go`, used only when the config yields
  no pairs):
  - `curatedAssets`: XLM, USDZ, ZARZ, EURZ, XAUZ, BTCZ, USDC with issuer addresses
  - `curatedPairs`: Predefined trading pairs
  - `fallbackLiquidityPoolIDs`: Static map of pool IDs for known pairs (bidirectional)

- **Rendering/layout**:
  - Fixed‑width layout designed for ~140×60
  - All screens: Header + Subtitle + Content + Footer
  - Pair Info: Order Book (left) + Trades (right) / Liquidity Pool (full width)
  - Decimal alignment: 2–7 places with space separators

## Stellar-Specific Guidelines

### Decimal Precision
- **CRITICAL:** All Stellar transactions and amounts MUST use **7 or fewer decimal places**
- Display amounts with at least 2 decimal places, up to 7 when needed
- Never truncate or round beyond 7 decimals

### API Endpoint
- Default: `https://horizon.stellar.org` (public mainnet)
- Preferred for production: ValidationCloud endpoint at `https://mainnet.stellar.validationcloud.io/v1/jcRGf8fyg_vHRumAMzbD0uENOzQ20kXYtV65DX_ly3w`
- Set via `HORIZON_URL` environment variable
- Prefer MAINNET for development with small real amounts

## Project Structure

Follows standard Go project layout:

```
sdexmon/
├── cmd/sdexmon/              # Main application
│   ├── main.go               # Entry point (~2700 lines)
│   ├── maintenance_update.go # Maintenance key handling and commands
│   ├── maintenance_view.go   # Maintenance screen renderers
│   ├── maintenance_test.go   # Maintenance routing and removal tests
│   └── upgrade_test.go       # Upgrade notice tests
├── internal/                 # Private packages
│   ├── models/               # Data structures
│   │   ├── types.go          # Model, ScreenState, Messages
│   │   ├── constants.go      # Curated assets, pairs, pool IDs
│   │   └── maintenance.go    # Maintenance mode types
│   ├── config/               # Configuration
│   │   ├── config.go         # Environment & logging
│   │   ├── assets.go         # Asset parsing utilities
│   │   └── user_config.go    # User configuration handling
│   ├── ui/                   # UI components
│   │   └── upgrade.go        # Upgrade notice screen renderer
│   ├── version/              # Version management
│   │   ├── checker.go        # GitHub release checker
│   │   └── checker_test.go   # Version comparison tests
│   └── stellar/              # Stellar API helpers
│       ├── confirmation.go   # Asset confirmation
│       └── expert.go         # stellar.expert API client
├── docs/                     # Project documentation
│   ├── MAINTENANCE_MODE.md   # Legacy maintenance mode notes
│   ├── MIGRATION.md          # Pre-wrapper upgrade guide
│   ├── ROUTING_IMPLEMENTATION.md # Routing system documentation
│   └── raspberry-pi.md       # Raspberry Pi deployment notes
├── go.mod                    # Module: github.com/sdexmon/sdexmon
├── go.sum                    # Dependencies
├── run                       # Launcher script
├── install.sh                # Installation script
├── tui                       # Pre-compiled binary
└── WARP.md                   # This file
```

## Known Issues & Technical Debt

1. **Code organization:** Main business logic still in single file
   - All TUI code (~2085 lines) in `cmd/sdexmon/main.go`
   - Should be split into: `internal/ui/`, `internal/stellar/`, `internal/format/`
   - Created packages (`models`, `config`) are first step
   - Further refactoring recommended but not blocking

2. **Thin test coverage:** Only targeted regression tests exist
   - Covered: upgrade notice, maintenance routing/removal, config pair CRUD,
     version comparison, order book request shape, layout helpers
   - Missing: broad view snapshots and mocked stellar.expert asset search
   - Should add: unit tests, mocked API tests, format tests

## Troubleshooting

### Wrong landing page after install
If you see "SCAR AQUILA" and "Service Selection" instead of the "sdexmon_" landing page:
- This was a build issue in early releases (v1.0.0-v1.0.3 had incorrect goreleaser config)
- Solution: Reinstall with the latest version (v0.1.1+)
  ```bash
  curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
  ```

### Version history note
Versions v1.0.0 through v1.0.3 were released with incorrect build configuration and have been deprecated.
The correct versioning continues from v0.1.0 → v0.1.1+

## Go Coding Standards

### Naming Conventions
- **Packages:** lowercase, single word (e.g., `stellar`, `ui`, `orderbook`)
- **Files:** snake_case (e.g., `order_book.go`, `liquidity_pool.go`)
- **Exported:** PascalCase (e.g., `OrderBook`, `FetchTrades()`)
- **Unexported:** camelCase (e.g., `parseResponse`, `apiClient`)
- **Interfaces:** PascalCase with "er" suffix when possible (e.g., `Trader`, `Fetcher`)

### Code Organization Principles
- Keep `main.go` minimal - only application initialization
- Group related functionality in packages
- Use composition over inheritance
- Handle all errors explicitly
- Always use `context.Context` for cancellation and timeouts

### Required Practices
- Use `go fmt` and `go vet` before committing
- Implement structured logging (currently using `log` package)
- Add graceful shutdown handlers for cleanup
- Mock external dependencies (Horizon API, stellar.expert) in tests

## Future Refactoring Plan

To align with Go best practices and team standards:

### Phase 1: Module & Build Fixes ✅
1. ✅ Update `go.mod` module name to proper path
2. ✅ Fix `.goreleaser.yml` to point to actual main location (`./cmd/sdexmon`)
3. ✅ Verify builds work cross-platform

### Phase 2: Code Organization (Partially Complete)
1. ✅ **Move main.go → cmd/sdexmon/main.go**
2. 🟡 **Extract packages to internal/:** (Started)
   - ❌ `internal/ui/` - Bubble Tea components, views, routing (TODO)
   - ❌ `internal/stellar/` - Horizon client wrapper, API calls (TODO)
   - ✅ `internal/models/` - Data structures (OrderBook, Trade, Liquidity)
   - ✅ `internal/config/` - Environment variable handling
3. ✅ **Maintain single entry point** in cmd/sdexmon/main.go that orchestrates packages

### Phase 3: Testing
1. Add unit tests for data transformations
2. Mock Horizon API responses for integration tests
3. Table-driven tests for price formatting and decimal handling
4. Target 80%+ code coverage

## License

Custom non-commercial license (see LICENSE file):
- Personal, non-commercial use allowed
- Attribution required to Daniel van Tonder
- Commercial use prohibited without written consent