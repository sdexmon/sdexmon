# sdexmon 🪐  
> Real-time Stellar DEX Monitor — Order Books, Trades, Liquidity Pools — in your Terminal.

[![version](https://img.shields.io/badge/version-v0.1.0-blue.svg)](#)
[![language](https://img.shields.io/badge/language-Go-00ADD8.svg)](#)
[![license](https://img.shields.io/badge/license-Custom-blue)](#)

**sdexmon** is a Go-based TUI for the Stellar Decentralized Exchange.  
View order books, monitor liquidity pools, and follow live trades — directly in your terminal.

## Features
- 📊 **Order Book Viewer** - Real-time bids/asks with depth visualization
- 💧 **Liquidity Pool Analytics** - Track locked amounts, fees, and volume
- 📈 **Live Trade Stream** - Monitor recent trades with color-coded buy/sell
- 🔍 **Asset Exposure** - View all liquidity pools containing a specific asset
- 🎨 **Beautiful TUI** - Built with Bubble Tea & Lip Gloss
- ⚡ **Fast Updates** - Powered by Horizon RPC / Stellar API

## Installation

### Quick Install (Recommended)

Install the latest release with a single command:

```bash
curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
```

This will:
- Detect your platform (macOS, Linux, Windows)
- Download the appropriate binary from GitHub releases
- Install to `/usr/local/bin`
- Make the binary executable

**Custom install directory:**
```bash
INSTALL_DIR=~/.local/bin curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash
```

### Alternative Methods

#### Manual Download
1. Go to [Releases](https://github.com/sdexmon/sdexmon/releases/latest)
2. Download the archive for your platform
3. Extract and place the binary in your PATH

#### Go Install (if you have Go)
```bash
go install github.com/sdexmon/sdexmon@latest
```

#### From Source

```bash
git clone https://github.com/sdexmon/sdexmon.git
cd sdexmon
go run .
```

### Quick Start

```bash
# Run with launcher script (recommended)
./run

# Or run directly
go run ./cmd/sdexmon
```

## Usage

The application starts with a service selection menu:

1. **View Asset Pairs** - Monitor order books, trades, and liquidity pools for trading pairs
2. **View Single Asset Exposure** - See all liquidity pools containing a specific asset

### Navigation

- `↑/↓` - Navigate lists
- `enter` - Select/confirm
- `b` - Back to previous screen
- `z` - Toggle debug view
- `,/.` - Adjust orderbook depth
- `q` - Quit application

### Environment Variables

```bash
# Horizon endpoint (default: https://horizon.stellar.org)
export HORIZON_URL="https://horizon.stellar.org"

# Optional: Start directly at a pair
export BASE_ASSET="native"  # or CODE:ISSUER
export QUOTE_ASSET="USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"

# Enable debug mode
export DEBUG="true"
```

## Development

```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Build
go build -o sdexmon ./cmd/sdexmon
```

## Project Structure

```
sdexmon/
├── cmd/sdexmon/          # Main application
├── internal/
│   ├── models/           # Data structures and constants
│   └── config/           # Configuration and environment
├── go.mod                # Module definition
└── WARP.md               # Development guide
```

## License

Custom non-commercial license. See [LICENSE](LICENSE) for details.

## Links

- **Website:** [sdexmon.host](https://sdexmon.host)
- **Documentation:** See [WARP.md](WARP.md) for detailed development guide
