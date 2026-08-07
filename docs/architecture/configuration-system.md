# Configuration system

Configuration lives in `internal/config`. It owns the YAML schema, asset string
parsing, the pair CRUD helpers used by the maintenance screen, and the Horizon
client and debug logger constructors.

## File location

`config.GetConfigPath()` returns `$HOME/.config/sdexmon/config.yaml`, or an
empty string when the home directory cannot be resolved.

The directory is created with mode 0755 and the file written with mode 0644 on
save.

## Schema

```go
type Config struct {
    App struct {
        Version     string
        DefaultPair string
    }
    Pairs  []Pair
    Assets []Asset
    Preferences struct {
        DefaultOrderBookDepth int
        DefaultLiquidityPools int
        AutoRefresh           bool
        RefreshIntervalMs     int
        ShowDebug             bool
    }
    SystemSettings struct {
        TerminalSize struct {
            Width  int
            Height int
        }
    }
}

type Pair struct {
    Name         string
    Base         string
    Quote        string
    LP           string
    Favorite     bool
    ShowDecimals int
}

type Asset struct {
    Name         string
    ShowDecimals int
}
```

Only `Pairs` and `Assets` affect the running application today. `App`,
`Preferences` and `SystemSettings` are parsed and round-tripped on save, but
nothing reads them: the order book depth is derived from the terminal height,
the refresh intervals are compile-time constants, debug mode comes from the
`DEBUG` environment variable, and terminal size comes from the live terminal.

## Load path

`LoadConfig` returns a default `Config` with an empty `Pairs` slice when the
file does not exist. A read or parse failure returns an error.

`loadConfiguration` in `cmd/sdexmon/main.go` converts that into runtime state:

1. On error, set `configuredPairs = curatedPairs` and
   `liquidityPoolIDs = fallbackLiquidityPoolIDs`, then return the error. The
   caller logs a warning and continues.
2. Otherwise, walk `cfg.Pairs`. Parse `base` and `quote` with
   `config.ParseAsset`. A pair whose assets do not parse is logged and skipped,
   so one bad entry does not take the rest of the file down.
3. Append a `pairOption` carrying the label, the short codes, the resolved
   assets, and the favorite flag.
4. When `lp` is set, register it in `liquidityPoolIDs` under four keys: the
   code pair in both orientations (`BASE-QUOTE`, `QUOTE-BASE`) and the
   issuer-qualified pair in both orientations. The issuer-qualified key is
   preferred at lookup time; the code-only keys exist for the legacy fallback
   table.
5. If no pairs survived, fall back to `curatedPairs`.
6. Fill in any missing `liquidityPoolIDs` entries from
   `fallbackLiquidityPoolIDs`.

### Fallback tables

`cmd/sdexmon/main.go` holds `curatedAssets`, `curatedPairs` and
`fallbackLiquidityPoolIDs`. These are safety nets, not the configuration.
`assetsForPair` also uses `curatedAssets` to resolve a pair that carries only
codes.

The similar-looking tables in `internal/models/constants.go` are dead code and
are not in the load path. See
[the architecture overview](overview.md#known-technical-debt).

## Asset string parsing

`config.ParseAsset` accepts:

- `""`, `native`, `XLM`, and `XLM:native` -> `txnbuild.NativeAsset{}`
- `CODE:ISSUER` -> `txnbuild.CreditAsset{Code, Issuer}`, with the code
  upper-cased and both sides trimmed

Anything else is an error. Note that `cmd/sdexmon/main.go` also has a local
`parseAsset` used for the environment variables and the custom pair input; it
does not accept the `XLM:native` spelling.

Two representations exist on the way out:

- `config.AssetString` produces `native` or `CODE:ISSUER`. This is the canonical
  form used for pair comparison and for values written by the maintenance
  screen.
- `getAssetName` in `cmd/sdexmon/main.go` produces `XLM:native` or
  `CODE:ISSUER`. This is the form used for decimal lookups.

The issuer is part of the asset identity throughout. Two assets sharing a code
from different issuers are never treated as equal.

## Pair CRUD

`internal/config/user_config.go` backs the maintenance screen:

- `AddCustomPair` loads the config, rejects a duplicate, appends a `Pair` named
  `CODEA/CODEB` with `ShowDecimals: 7`, and saves.
- `RemoveCustomPair` loads the config, finds the first pair matching both assets
  in either orientation, removes it, and saves. Returns an error naming the
  config path when nothing matches.
- `ListCustomPairs` returns the configured pairs in file order.

`pairMatches` normalises both sides through `ParseAsset` before comparing, so
the equivalent spellings of XLM compare equal, and comparison is
orientation agnostic.

`config.AddPair` is a separate, lower-level helper that takes raw strings and
derives `show_decimals` from the asset codes. It is not used by the maintenance
flow.

## Decimal precision

Stellar amounts never exceed 7 decimal places. Within that ceiling, the number
of displayed decimals is resolved per pair.

`Config.GetPairDecimals(baseName, quoteName)` returns a base and a quote
precision:

1. Scan `Pairs` for an entry matching both asset codes in either orientation. If
   found and its `show_decimals` is greater than zero, use that value for both
   sides.
2. Otherwise resolve each side independently with `GetAssetDecimals`.

`GetAssetDecimals(assetName)`:

1. Exact match on `Assets[].name`.
2. Match on asset code, so an `assets` entry named `USDC:GA5ZS...` also matches
   a lookup for `USDC`.
3. Fall back to a built-in default by code: `BTCZ` gets 0, `XAUZ` gets 7,
   everything else gets 2.

Because step 1 only applies when `show_decimals` is greater than zero, a pair
configured with `show_decimals: 0` falls through to the asset lookup. For BTCZ
pairs that still resolves to 0, so the intended result holds, but a deliberate
zero on a non-BTCZ pair will not be honoured.

Where this is applied:

- Order book prices always use 7 decimals, for granularity.
- Order book amounts use the base precision; cumulative totals use the quote
  precision.
- Trade prices use the quote precision; trade amounts use the base precision.
- Liquidity pool and exposure figures are trimmed to 2 decimals, since they are
  aggregate values rather than tradeable amounts.

When `appConfig` is nil or the pair is unset, both precisions default to 2.
