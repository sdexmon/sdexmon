# Trading Pairs Management Guide

Since the maintenance UI has been removed for deployment, trading pairs must be managed by editing the code directly. This guide explains how to add, edit, and remove trading pairs in sdexmon.

## Quick Start

All pair configuration is in: `internal/models/constants.go`

## Adding a New Asset

1. Open `internal/models/constants.go`
2. Find the `CuratedAssets` map
3. Add your new asset:

```go
// For issued assets
"USDT": txnbuild.CreditAsset{Code: "USDT", Issuer: "GCQTGZQQ5G4PTM2GL7CDIFKUBIPEC52BROAQIAPW53XBRJVN6ZJVTG6V"},

// For native XLM (already included)
"XLM": txnbuild.NativeAsset{},
```

**Finding Asset Issuers:**
- Go to [stellar.expert](https://stellar.expert)
- Search for your asset
- Copy the issuer address (56 characters starting with 'G')

## Adding a New Trading Pair

1. Make sure both assets exist in `CuratedAssets`
2. Find the `CuratedPairs` slice
3. Add your pair:

```go
{"USDT", "USDZ"}, // USDT/USDZ pair
```

## Adding Liquidity Pool Data (Optional)

If the pair has a liquidity pool on Stellar:

1. Find the `LiquidityPoolIDs` map
2. Add both directions:

```go
"USDT-USDZ": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
"USDZ-USDT": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
```

**Finding Pool IDs:**
- Go to [stellar.expert](https://stellar.expert)
- Navigate to "Liquidity Pools"
- Search for your asset pair
- Copy the pool ID (64 hex characters)

## Removing a Trading Pair

1. Remove from `CuratedPairs` slice
2. Remove both directions from `LiquidityPoolIDs` map
3. Optionally remove unused assets from `CuratedAssets`

## Testing Your Changes

```bash
# Build and test
go build -o sdexmon ./cmd/sdexmon
./sdexmon --version

# Run the app
./run
```

## Example: Adding USDT/USDZ Pair

```go
// 1. Add USDT asset (if not already present)
var CuratedAssets = map[string]txnbuild.Asset{
    // ... existing assets ...
    "USDT": txnbuild.CreditAsset{Code: "USDT", Issuer: "GCQTGZQQ5G4PTM2GL7CDIFKUBIPEC52BROAQIAPW53XBRJVN6ZJVTG6V"},
}

// 2. Add trading pair
var CuratedPairs = []PairOption{
    // ... existing pairs ...
    {"USDT", "USDZ"}, // USDT/USDZ
}

// 3. Add liquidity pool (if exists)
var LiquidityPoolIDs = map[string]string{
    // ... existing pools ...
    "USDT-USDZ": "abc123...", // 64 hex chars
    "USDZ-USDT": "abc123...", // same ID
}
```

## Important Rules

- ✅ Asset codes: 1-12 characters, A-Z and 0-9 only
- ✅ Issuer addresses: exactly 56 characters starting with 'G'
- ✅ Pool IDs: exactly 64 hex characters (0-9, a-f)
- ✅ Always add both directions for liquidity pools
- ✅ Test your changes by building and running

## Need Help?

- Check the detailed documentation in `WARP.md`
- Look at existing examples in `constants.go`
- Use stellar.expert to find asset and pool information