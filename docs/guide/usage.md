# Usage

## Starting

```bash
sdexmon
```

From a source checkout:

```bash
./run
```

SDEXMON opens on the landing screen. If the startup release check found a newer
version, the upgrade notice is shown first; press `esc` to continue. See
[Upgrading](upgrading.md).

The layout targets a 140x60 terminal. Narrower terminals still work: below 102
columns the order book and trades panels stack vertically instead of sitting
side by side, and the order book depth and exposure rows shrink to fit the
available height.

## Global keys

`q` and `ctrl+c` quit from anywhere. The single exception is a maintenance text
field, where `q` types normally and only `ctrl+c` quits.

`u` opens the upgrade notice, but only while the startup check found a newer
release. It is available on the landing, Pair Info and Pair Debug screens, and
is only advertised in the footer when an update exists.

## Landing

The landing screen shows the SDEXMON banner with the version and build.

```
enter   open the pair selector
m       pair maintenance
u       upgrade notice (only when an update is available)
q       quit
```

## Pair selector

A popup, opened with `enter` from the landing screen or `p` from Pair Info. It
shows a window of ten pairs around the highlighted entry.

```
up / down   move (k / j also work)
enter       select the pair and start monitoring
s           search
esc         close the popup
q           quit
```

### Search

Press `s` and type to filter. Matching is case insensitive and checks the pair
label as well as the base and quote asset codes, so `USD` matches `USDC/USDZ`
and `XLM/USDC` alike. Results update as you type, and `No pairs found` is shown
when nothing matches.

```
up / down   move through the filtered results
enter       select
esc         leave search and restore the full list
```

Note that `q` still quits while searching, so it cannot be typed into the pair
search field. Search terms containing `q` are not currently supported here.

## Pair Info

The main monitoring screen. It renders five panels:

- ORDER BOOK -- asks above, the spread, then bids, each row showing price,
  amount, cumulative total and a depth bar. Prices always use 7 decimals;
  amounts and totals use the configured precision for the pair. Outliers are
  filtered out so one absurd offer cannot flatten the depth bars.
- TRADES (latest) -- newest last, green for buys and red for sells, with the
  elapsed time since the ledger closed.
- LIQUIDITY POOL -- locked amounts, 1-day and 7-day fees, and 1-day and 7-day
  volume for both reserves.
- Two exposure panels -- the largest liquidity pools holding the base asset and
  the quote asset respectively, sorted by locked amount.

```
p       open the pair selector
d       toggle the detail (debug) view
m       pair maintenance
u       upgrade notice (only when an update is available)
q       quit
```

The footer shows the current shortcuts on the left and Stellar ledger capacity
usage on the right, polled every 10 seconds.

## Pair Debug

Raw diagnostics for the current pair.

```
d       back to Pair Info
u       upgrade notice (only when an update is available)
q       quit
```

Set `DEBUG=true` before starting for the captured log buffer to be populated.

## Upgrade notice

```
enter   run the installer now
esc     continue on the current version
q       quit
```

The notice is advisory. Quitting and dismissing always work, so an outdated
build can never lock you out. Running the installer replaces the running binary,
so SDEXMON exits afterwards and asks you to start it again.

## Maintenance

Press `m` from the landing or Pair Info screen. The menu header shows how many
pairs are configured and the config file path.

```
1       add asset pair
2       remove asset pair
3       view configured pairs
esc     back to the landing screen
q       quit
```

See [Pair management](pair-management.md) for the full add, remove and list
workflows including the asset search sources.

## Display mode

`sdexmon display` is unattended: there is no key handling beyond quitting, the
upgrade notice is suppressed, and the footer shows data freshness instead of
shortcuts. See [Configuration](configuration.md) for the flags and
[the Raspberry Pi guide](../deployment/raspberry-pi.md) for a systemd setup.

## Not currently reachable

The custom pair input screen (free-text base and quote entry) is implemented and
rendered, but no key binding currently routes to it. Use the maintenance screen
to add a pair, or edit the config file directly.
