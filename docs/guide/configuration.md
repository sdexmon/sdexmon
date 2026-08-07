# Configuration

SDEXMON reads settings from three places, in this order of specificity:

1. Command-line flags -- per-invocation behaviour, mainly display mode.
2. Environment variables -- endpoint, debug, and pair overrides.
3. `~/.config/sdexmon/config.yaml` -- the trading pairs shown in the selector.

## Environment variables

### HORIZON_URL

The Horizon REST endpoint used for order books, trades, liquidity pool details
and network fee stats.

Default: `https://horizon.stellar.org`

```bash
export HORIZON_URL="https://horizon.stellar.org"
```

The installer wrapper and the `run` script both set this default, so you only
need to export it when pointing at a different Horizon instance.

### DEBUG

Set to `true` or `1` to enable debug mode. Debug mode makes the Pair Debug
screen useful; the `d` key toggles it either way.

Default: `false` from the installer wrapper, `true` from the `run` script.

```bash
DEBUG=true sdexmon
```

Log output is captured in an in-memory ring buffer of the last 100 lines while
the TUI is running, so it cannot corrupt the alternate screen. Startup
diagnostics before the TUI starts still go to stderr.

### BASE_ASSET and QUOTE_ASSET

Preselect a pair. Accepted formats are `native`, `XLM`, `XLM:native`, and
`CODE:ISSUER`.

```bash
export BASE_ASSET="native"
export QUOTE_ASSET="USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"
```

In normal interactive use these set the initially highlighted pair, but the app
still opens on the landing screen. Only `display` mode starts straight on the
Pair Info screen when both are set. An unparseable value is logged and ignored.

### LP_POOL_ID

Force a specific liquidity pool instead of resolving one for the current pair.
When set, it applies to every pair, so it is mainly a debugging aid.

```bash
LP_POOL_ID=a468d41d8e9b8f3c7209651608b74b7db7ac9952dcae0cdf24871d1d9c7b0088 sdexmon
```

## Command-line flags

### --version

```bash
sdexmon --version
```

Prints `<version> (build <commit>)` and exits. Must be the first argument.

### display

`sdexmon display` starts unattended display mode: the app opens directly on Pair
Info and never shows the upgrade notice, so an un-attended board keeps rendering
market data. `--display` works as an equivalent flag.

```
--display            start directly in unattended display mode
--favorites-only     rotate only configured favorite pairs
--no-update-check    disable the startup release check
--pair <label>       initial configured pair, for example XLM/USDC
--rotate <duration>  pair rotation interval, default 30s, use 0 to disable
```

Examples:

```bash
# Rotate through favorite pairs every 30 seconds, no update check
sdexmon display --favorites-only --rotate=30s --no-update-check

# Pin the display to one configured pair
sdexmon display --pair XLM/USDC --rotate=0
```

`--pair` matches the configured pair `name`, or the `BASE/QUOTE` code pair, case
insensitively. An unknown label is a fatal error, so a typo in a systemd unit
fails loudly instead of silently showing the wrong market.

`--favorites-only` falls back to all configured pairs when no pair is marked
`favorite: true`.

In display mode the footer replaces the key hints with data freshness, for
example `order book 12s ago`, and appends `| reconnecting` while a data source
is failing. The last good snapshot stays on screen.

## Configuration file

Location: `~/.config/sdexmon/config.yaml`

The directory is created with mode 0755 and the file with mode 0644 when
SDEXMON writes it. If the file does not exist, SDEXMON starts with a built-in
default configuration that has no pairs, and falls back to the curated pair
table compiled into the binary.

```yaml
app:
  version: "0.1.0"
  default_pair: "USDC/USDZ"

pairs:
  - name: XLM/USDC
    base: XLM:native
    quote: USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN
    lp: a468d41d8e9b8f3c7209651608b74b7db7ac9952dcae0cdf24871d1d9c7b0088
    favorite: true
    show_decimals: 7

assets:
  - name: XAUZ
    show_decimals: 7

preferences:
  default_order_book_depth: 7
  auto_refresh: true
  refresh_interval_ms: 1500
  show_debug: false

system_settings:
  terminal_size:
    width: 140
    height: 60
```

### Field reference

`pairs[].name` is the label shown in the selector. When omitted, the label
falls back to `BASE/QUOTE` derived from the asset codes.

`pairs[].base` and `pairs[].quote` accept `native`, `XLM`, `XLM:native`, or
`CODE:ISSUER`. The issuer is part of the asset identity: two assets with the
same code from different issuers are different assets.

`pairs[].lp` is an optional 64-character hex liquidity pool ID. When omitted,
SDEXMON asks Horizon for a pool matching the two reserves at runtime.

`pairs[].favorite` marks the pair for `display --favorites-only`.

`pairs[].show_decimals` sets the displayed decimal places for that pair, 0 to 7.
See [the configuration system](../architecture/configuration-system.md) for how
this interacts with the `assets` list and the built-in defaults.

`app`, `preferences` and `system_settings` are parsed and preserved on save, but
only `pairs` and `assets` currently influence the running application. Terminal
size is taken from the live terminal, not from `system_settings`.

### Editing

Prefer the in-app maintenance screen; press `m` from the landing or Pair Info
screen. It validates issuers, prevents duplicates, and reloads immediately. See
[Pair management](pair-management.md).

Hand edits take effect on the next start. An invalid pair entry is logged and
skipped rather than aborting the load, so one bad line will not lock you out.
