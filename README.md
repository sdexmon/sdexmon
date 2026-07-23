NOTE:
-----
This README includes an optional ANSI-colored version for terminal viewing.
If you are reading this on GitHub, scroll down for the plain-text version.

```ansi
\x1b[36m███████╗██████╗ ███████╗██╗  ██╗███╗   ███╗ ██████╗ ███╗   ██╗
██╔════╝██╔══██╗██╔════╝╚██╗██╔╝████╗ ████║██╔═══██╗████╗  ██║
███████╗██║  ██║█████╗   ╚███╔╝ ██╔████╔██║██║   ██║██╔██╗ ██║
╚════██║██║  ██║██╔══╝   ██╔██╗ ██║╚██╔╝██║██║   ██║██║╚██╗██║
███████║██████╔╝███████╗██╔╝ ██╗██║ ╚═╝ ██║╚██████╔╝██║ ╚████║
╚══════╝╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝         ╚═════╝ ╚═╝  ╚═══╝\x1b[0m

\x1b[36mSTELLAR DEX MONITORING — TERMINAL NATIVE\x1b[0m
\x1b[90m------------------------------------------------------------\x1b[0m
\x1b[32m[ 1 ]\x1b[0m \x1b[36mOVERVIEW\x1b[0m   \x1b[32m[ 2 ]\x1b[0m \x1b[36mFEATURES\x1b[0m   \x1b[32m[ 3 ]\x1b[0m \x1b[36mQUICKSTART\x1b[0m
\x1b[32m[ 4 ]\x1b[0m \x1b[36mINSTALL\x1b[0m    \x1b[32m[ 5 ]\x1b[0m \x1b[36mUSAGE\x1b[0m      \x1b[32m[ 6 ]\x1b[0m \x1b[36mCONFIG\x1b[0m
\x1b[32m[ 7 ]\x1b[0m \x1b[36mDEV\x1b[0m        \x1b[32m[ 8 ]\x1b[0m \x1b[36mLINKS\x1b[0m
\x1b[90m------------------------------------------------------------\x1b[0m
\x1b[36mTerminal view:\x1b[0m \x1b[32mless -R README.ansi\x1b[0m  \x1b[90m|\x1b[0m  \x1b[32mmake readme | less -R\x1b[0m
\x1b[36mSDEXMON — monitor the DEX. trust the terminal.\x1b[0m
```

# SDEXMON
STELLAR DEX MONITORING — TERMINAL NATIVE
======================================

SDEXMON is a real-time monitor for the Stellar Decentralized Exchange.
It runs entirely in your terminal and focuses on signal, not dashboards.

No web UI.
No accounts.
No noise.

------------------------------------------------------------

[ 1 ] OVERVIEW
[ 2 ] FEATURES
[ 3 ] QUICKSTART
[ 4 ] INSTALL
[ 5 ] USAGE
[ 6 ] CONFIGURATION
[ 7 ] DEVELOPMENT
[ 8 ] LINKS

------------------------------------------------------------


[ 1 ] OVERVIEW
--------------
sdexmon is a Go-based TUI (terminal user interface) for observing live
activity on the Stellar DEX.

It allows operators, developers, and traders to:
- Inspect live order books
- Monitor liquidity pools
- Follow real-time trades
- Analyze asset exposure across pools

All output is rendered directly in the terminal.


[ 2 ] FEATURES
--------------
- Order book viewer (live bids / asks with depth)
- Liquidity pool analytics (locked amounts, fees, volume)
- Live trade stream (color-coded buy / sell)
- Asset exposure across pools
- Fast, event-driven updates via Horizon RPC
- Terminal-first UX built with Bubble Tea


[ 3 ] QUICKSTART
----------------
Install and run with a single command:

    curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash

The installer will:
- Detect your platform (macOS / Linux / Windows)
- Download the correct binary from GitHub Releases
- Install into /usr/local/bin by default
- Create a wrapper with sensible defaults


[ 4 ] INSTALL
-------------
One-line install (recommended):

    curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash

Custom install directory:

    INSTALL_DIR=~/.local/bin \
    curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash

Alternative methods:

Manual download:
- https://github.com/sdexmon/sdexmon/releases/latest

Go install (requires Go toolchain):

    go install github.com/sdexmon/sdexmon/cmd/sdexmon@latest

From source:

    git clone https://github.com/sdexmon/sdexmon.git
    cd sdexmon
    go run ./cmd/sdexmon


[ 5 ] USAGE
-----------
Start the application:

    sdexmon

Or, when running from source:

    go run ./cmd/sdexmon

On startup you can:
1. View asset pairs (order books, trades, liquidity pools)
2. View single-asset exposure across pools

Navigation keys:
- Up / Down : move
- Enter     : select
- b         : back
- z         : toggle debug view
- , / .     : adjust order book depth
- q         : quit


[ 6 ] CONFIGURATION
-------------------
The installer creates a wrapper that sets sensible defaults.

Optional environment variables:

    # Horizon endpoint
    export HORIZON_URL="https://horizon.stellar.org"

    # Supply the pair used by display mode
    export BASE_ASSET="native"
    export QUOTE_ASSET="USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"

    # Disable debug mode
    export DEBUG="false"

Passive display mode:

    # Start immediately and rotate favorite pairs every 30 seconds
    sdexmon display --favorites-only --rotate=30s --no-update-check

    # Pin the display to one configured pair
    sdexmon display --pair XLM/USDC --rotate=0

Raspberry Pi 5 and systemd setup:

    docs/raspberry-pi.md


[ 7 ] DEVELOPMENT
-----------------
Format code:

    go fmt ./...

Run checks:

    go vet ./...

Build binary:

    go build -o sdexmon ./cmd/sdexmon


Project structure:

    sdexmon/
    ├── cmd/sdexmon/      main application
    ├── internal/
    │   ├── models/      data structures
    │   └── config/      configuration
    └── go.mod


[ 8 ] LINKS
-----------
Repository : https://github.com/sdexmon/sdexmon
Website    : https://sdexmon.host
Discord    : d4n13vt

------------------------------------------------------------
SDEXMON — monitor the DEX. trust the terminal.
