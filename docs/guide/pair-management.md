# Pair management

`~/.config/sdexmon/config.yaml` is the source of truth for the pairs shown in
the selector. The curated tables compiled into the binary are only a fallback,
used when the config cannot be read or yields no usable pairs.

There are two ways to manage pairs: the in-app maintenance screen, and editing
the YAML by hand.

## In-app maintenance

Press `m` from the landing or Pair Info screen.

```
1       add asset pair
2       remove asset pair
3       view configured pairs
esc     back to the landing screen
q       quit
```

Both add and remove write the file and reload it immediately, so the pair
selector reflects the change without a restart.

### Adding a pair

The flow is: pick a search source for asset A, find and select asset A, pick a
search source for asset B, find and select asset B, review, confirm.

#### 1. Choose a search source

```
up / down   move (k / j also work)
1 / 2 / 3   pick a source directly
enter       choose the highlighted source
esc         back
q           quit
```

The three sources differ in how much they prove:

**1. Domain search (stellar.toml)** is authoritative and comes first. It reads
`https://<domain>/.well-known/stellar.toml` and lists only the `[[CURRENCIES]]`
that domain publishes about itself. Only the domain owner can list an issuer
there, so a lookalike asset from an unrelated issuer cannot appear. Entries
marked `status = "dead"`, entries using a `code_template`, and entries with an
invalid issuer are dropped. Per-currency `toml` links are followed, up to 12 of
them. The document is read with a 100 KiB cap and a 10 second timeout.

If the stellar.toml cannot be fetched at all, SDEXMON falls back to
stellar.expert but restricts the results to issuers whose home domain matches
exactly.

**2. Asset search (code or name)** is the fuzzy stellar.expert lookup, for when
you do not know the domain. It matches every asset on the network, so it is
convenient but unverified. Results are ranked with domain-bearing assets first,
then by trustline count, and every row shows its home domain.

**3. stellar.expert Top 50** is the most active assets on the network by
stellar.expert's own metrics. That ranking is not an endorsement.

#### 2. Enter the query

Domain search and asset search prompt for text. Top 50 loads straight away.

```
enter     run the lookup
esc       back to the source menu
ctrl+c    quit
```

`q` types normally here, so `ctrl+c` is the only way to quit from a text field.

#### 3. Select the asset

```
up / down   move (k / j also work)
enter       select
esc         back to the source menu
q           quit
```

Each row shows the code, name, home domain, issuer and notes. The notes flag
stellar.toml verification, a non-live status, and the trustline count.

An asset code proves nothing about who issued it. Always check the domain and
issuer columns before selecting.

#### 4. Confirm

```
enter   add the pair
esc     back to the asset B list
q       quit
```

The confirmation screen shows the pair name, best bid, best ask, and liquidity
pool locked amounts when a pool is known. Confirming appends the pair, saves the
file, reloads, and returns to the maintenance menu so you can add another.

Adding the same two assets twice is rejected. So is picking the same asset for
both sides.

### Removing a pair

```
up / down   move through the configured pairs
enter       select the pair to remove
esc         back to the menu
q           quit
```

Then confirm:

```
y or enter   remove
n or esc     cancel
q            quit
```

Matching is issuer-aware and orientation agnostic, so `BASE/QUOTE` and
`QUOTE/BASE` resolve to the same entry, and a pair is never removed just because
another issuer uses the same asset code.

### Viewing configured pairs

Option `3` is a read-only list of every configured pair with its issuers and
pool IDs.

```
up / down   scroll
esc         back to the menu
q           quit
```

### Limitations

None of the three search sources list native XLM, because all of them enumerate
issued assets. XLM pairs must be added by hand.

## Editing by hand

Edit `~/.config/sdexmon/config.yaml` and restart SDEXMON.

```yaml
pairs:
  - name: XLM/USDC
    base: XLM:native
    quote: USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN
    lp: a468d41d8e9b8f3c7209651608b74b7db7ac9952dcae0cdf24871d1d9c7b0088
    favorite: true
    show_decimals: 7
```

`base` and `quote` accept `native`, `XLM`, `XLM:native`, or `CODE:ISSUER`.

`lp` is optional. When omitted, SDEXMON asks Horizon for a pool matching the two
reserves at runtime.

A pair whose assets cannot be parsed is logged and skipped, so a single bad
entry will not prevent the rest of the file from loading.

### Validation rules

- Asset codes are 1 to 12 characters, A-Z and 0-9.
- Issuer addresses are exactly 56 characters, starting with `G`, using the
  base32 alphabet (A-Z and 2-7).
- Liquidity pool IDs are exactly 64 hexadecimal characters.

### Finding the values

- Asset issuers: search the asset on <https://stellar.expert>, or read the
  issuer's own `https://<domain>/.well-known/stellar.toml`, which is the
  authoritative source.
- Liquidity pool IDs: the liquidity pools section on stellar.expert, or leave
  `lp` empty and let Horizon resolve it.
