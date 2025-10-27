# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Overview

Terminal UI (Go, Bubble Tea/Lip Gloss) for visualizing Stellar spot markets. Features:
- Asset pair monitoring: order books, trades, and liquidity pools
- Navigation-based routing with pair selection landing page
- Polls Horizon for order books/trades; fetches LP metrics from stellar.expert
- Defaults to curated asset pairs, 140×60 layout, and 2–7 decimal rendering
- **Note:** Maintenance UI has been removed - pairs are now managed via code

Key files:
- `main.go`: entire TUI (~2085 lines) containing routing, model/update/view, Horizon calls, LP fetch, key bindings
- `run`: convenience launcher script that sets safe defaults (Horizon URL, debug mode, terminal size) and runs `go run .`
- `go.mod`: dependency manifest (Bubble Tea, Lip Gloss, Stellar Go SDK)
- `ROUTING_IMPLEMENTATION.md`: detailed routing system documentation
- `.env`: local environment variables (not tracked in git)
- `tui`: compiled binary

## Commands

- Quick start (recommended):
  ```bash
  ./run
  ```
  Sets `HORIZON_URL` to public Stellar Horizon, enables debug, adjusts terminal size, and executes `go run .`.

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
  go build -o sdexmon -ldflags="-X main.gitCommit=$(git rev-parse --short HEAD)" ./cmd/sdexmon
  ```

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
  - **Routing**: State machine with 4 screens (Landing, Pair Info, Pair Debug, Pair Input)
  - **Model** holds: current screen, selected assets, Horizon order book/trades, trade cursor, LP metrics, UI state
  - **Init**: when base/quote are set, schedules three tickers (order book, trades, LP)
  - **Update**: Screen-based navigation state machine
    - Landing: Displays sdexmon ASCII art with version and commit info + pair selector popup
    - Pair screens: Horizon polling via `fetchOrderbookCmd`, `fetchTradesCmd`, `resolveAndFetchLPCmd`
  - **View**: Router switches on currentScreen to render appropriate view
    - Landing: sdexmon ASCII branding with version display (top-left)
    - All other screens: SCAR AQUILA header, subtitle, content, context-aware footer
    - Pair Info: Three panels (Order Book, Trades, Liquidity Pool) + Exposure panels

## Navigation Flow

```
./run → Landing (with Pair Selector Popup)
         └─ Select Pair → Pair Info ⇄ Pair Debug
         └─ Custom Input → Pair Info ⇄ Pair Debug
```

## UI Controls

### Landing Screen
- `enter` (⏎): Open pair selector popup
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
- `q`: Quit

### Pair Debug Detail
- `d`: Back to pair info
- `q`: Quit

## Trading Pairs Management

**IMPORTANT:** The maintenance UI has been removed for deployment. Trading pairs are now managed by editing the `internal/models/constants.go` file directly.

### Adding a New Asset

1. Edit `internal/models/constants.go`
2. Add to the `CuratedAssets` map:
   ```go
   "CODE": txnbuild.CreditAsset{Code: "CODE", Issuer: "G..."},
   ```
3. For native XLM, use: `txnbuild.NativeAsset{}`
4. Find issuer addresses on stellar.expert

### Adding a New Trading Pair

1. Ensure both assets exist in `CuratedAssets`
2. Add to the `CuratedPairs` slice:
   ```go
   {"BASE", "QUOTE"}, // BASE/QUOTE - Description
   ```
3. Optionally add liquidity pool ID (both directions):
   ```go
   "BASE-QUOTE": "pool_id_64_hex_chars",
   "QUOTE-BASE": "pool_id_64_hex_chars", // Same ID
   ```

### Removing a Trading Pair

1. Remove from `CuratedPairs` slice
2. Remove both directions from `LiquidityPoolIDs` map
3. Optionally remove unused assets from `CuratedAssets`
4. Test changes by building and running

### Finding Required Data

- **Asset Issuers**: Use stellar.expert asset search
- **Liquidity Pool IDs**: Use stellar.expert liquidity pools section
- **Validation**: Asset codes must be 1-12 chars A-Z/0-9, issuer addresses 56 chars starting with 'G'

## Data Sources

- **Curated data** (in `internal/models/constants.go`):
  - `CuratedAssets`: XLM, USDZ, ZARZ, EURZ, XAUZ, BTCZ, USDC with issuer addresses
  - `CuratedPairs`: Predefined trading pairs available in pair selector
  - `LiquidityPoolIDs`: Static map of pool IDs for known pairs (bidirectional)

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
│   └── main.go               # Entry point (~2085 lines)
├── internal/                 # Private packages
│   ├── models/               # Data structures
│   │   ├── types.go          # Model, ScreenState, Messages
│   │   └── constants.go      # Curated assets, pairs, pool IDs
│   └── config/               # Configuration
│       ├── config.go         # Environment & logging
│       └── assets.go         # Asset parsing utilities
├── go.mod                    # Module: github.com/sdexmon/sdexmon
├── go.sum                    # Dependencies
├── run                       # Launcher script
├── tui                       # Pre-compiled binary
├── main_monolithic.go        # Backup of original single-file version
└── WARP.md                   # This file
```

## Known Issues & Technical Debt

1. **Code organization:** Main business logic still in single file
   - All TUI code (~2085 lines) in `cmd/sdexmon/main.go`
   - Should be split into: `internal/ui/`, `internal/stellar/`, `internal/format/`
   - Created packages (`models`, `config`) are first step
   - Further refactoring recommended but not blocking

2. **No tests:** Zero test coverage
   - No `*_test.go` files exist
   - Testing framework not set up
   - Should add: unit tests, mocked API tests, format tests

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

### Phase 1: Module & Build Fixes
1. Update `go.mod` module name to proper path
2. Fix `.goreleaser.yml` to point to actual main location
3. Verify builds work cross-platform

### Phase 2: Code Organization
1. **Move main.go → cmd/sdexmon/main.go**
2. **Extract packages to internal/:**
   - `internal/ui/` - Bubble Tea components, views, routing
   - `internal/stellar/` - Horizon client wrapper, API calls
   - `internal/models/` - Data structures (OrderBook, Trade, Liquidity)
   - `internal/config/` - Environment variable handling
3. **Maintain single entry point** in cmd/sdexmon/main.go that orchestrates packages

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