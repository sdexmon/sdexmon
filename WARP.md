# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

**sdexmon** is a Go-based Terminal User Interface (TUI) for monitoring the Stellar Decentralized Exchange in real-time. The application displays order books, liquidity pools, and live trades directly in the terminal.

**Technology Stack:**
- Language: Go
- TUI Framework: Bubble Tea & Lip Gloss
- Data Source: Horizon RPC / Stellar API
- Build Tool: GoReleaser

**Project Status:** Early stage - repository structure established, implementation pending

## Development Commands

### Building
```bash
# Once implemented, typical Go build
go build -o sdexmon ./cmd/sdexmon
```

### Running
```bash
# Local execution
go run ./cmd/sdexmon

# After installation
sdexmon
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Code Quality
```bash
# Format code
go fmt ./...
gofumpt -w .

# Lint code
golangci-lint run

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

### Release
```bash
# Create release with GoReleaser (requires tag)
goreleaser release --clean

# Test release process without publishing
goreleaser release --snapshot --clean
```

## Architecture

### Expected Project Structure
Based on Go standards and the GoReleaser configuration:

```
sdexmon/
├── cmd/sdexmon/         # Main application entry point
│   └── main.go
├── internal/            # Private application code
│   ├── ui/              # Bubble Tea TUI components
│   ├── stellar/         # Stellar API integration
│   └── models/          # Data models
├── pkg/                 # Public library code (if needed)
├── go.mod               # Go module definition
└── go.sum               # Dependency checksums
```

### Key Components (To Be Implemented)

1. **TUI Layer (Bubble Tea)**
   - Order book viewer
   - Liquidity pool analytics display
   - Live trade stream
   - Navigation and keyboard controls

2. **Stellar Integration**
   - Horizon API client
   - WebSocket streaming for real-time data
   - Order book data fetching
   - Liquidity pool monitoring

3. **Data Models**
   - Order book structures
   - Trade events
   - Liquidity pool metrics

## Stellar-Specific Guidelines

### Decimal Precision
- **CRITICAL:** All Stellar transactions and amounts MUST use **7 or fewer decimal places**
- Display amounts with at least 2 decimal places, up to 7 when needed
- Never truncate or round beyond 7 decimals

### API Endpoint
- Use ValidationCloud endpoint: `https://mainnet.stellar.validationcloud.io/v1/jcRGf8fyg_vHRumAMzbD0uENOzQ20kXYtV65DX_ly3w`
- Prefer MAINNET for development with small real amounts
- Implement proper error handling for API failures and rate limits

### Real-time Data
- Use Horizon streaming endpoints for live trade data
- Implement reconnection logic for WebSocket failures
- Buffer and manage high-frequency trade updates

## Go Coding Standards

### Naming Conventions
- **Packages:** lowercase, single word (e.g., `stellar`, `ui`)
- **Files:** snake_case (e.g., `order_book.go`, `liquidity_pool.go`)
- **Exported:** PascalCase (e.g., `OrderBook`, `FetchTrades()`)
- **Unexported:** camelCase (e.g., `parseResponse`, `apiClient`)
- **Interfaces:** PascalCase with "er" suffix when possible (e.g., `Trader`, `Fetcher`)

### Code Organization
- Keep `main.go` minimal - only application initialization
- Group related functionality in packages
- Use composition over inheritance
- Handle all errors explicitly

### Required Practices
- Always use `context.Context` for cancellation and timeouts
- Include health check endpoint/functionality
- Implement graceful shutdown for cleanup
- Use structured logging

## Dependencies

When adding dependencies:
```bash
# Add new dependency
go get <package>

# Update dependencies
go get -u ./...

# Clean up unused dependencies
go mod tidy
```

## Testing Strategy

### Test Files
- Place test files alongside source: `order_book_test.go`
- Use table-driven tests for multiple scenarios
- Mock external dependencies (Stellar API)

### Coverage Requirements
- Aim for 80%+ code coverage
- Focus on business logic and critical paths
- UI components may have lower coverage due to TUI complexity

## License

This project uses a **custom non-commercial license**. Users can use, copy, and modify for personal use only. Commercial use requires prior written consent. Attribution to Daniel van Tonder is required.

## CI/CD Notes

GoReleaser configuration exists (`.goreleaser.yml`) for multi-platform builds:
- Platforms: Linux, macOS, Windows
- Architectures: amd64, arm64
- Binary location: `./cmd/sdexmon`
