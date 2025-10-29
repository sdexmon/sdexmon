# Maintenance Mode - Add Asset Pair Feature

## Overview

Maintenance mode allows users to add custom trading pairs to sdexmon through an interactive TUI workflow. Custom pairs are saved to `~/.config/sdexmon/config.yaml` and persist between sessions.

## Accessing Maintenance Mode

Press `m` from:
- Landing screen (when pair selector popup is closed)
- Pair Info screen (when monitoring a trading pair)

## Workflow

### 1. Maintenance Menu
- Press `1` to start "Add Asset Pair" flow
- Future options (2 & 3) are coming soon

### 2. Asset A - Domain Input
- Enter a domain name to search for assets (e.g., `zeam.money`)
- Press `enter` to search stellar.expert API
- Press `esc` to go back

### 3. Asset A - Selection
- Browse search results with `↑`/`↓` or `k`/`j`
- Assets display as: `CODE - Name (issuer...)`
- Press `enter` to select
- Press `esc` to go back and try a different domain

### 4. Asset B - Domain Input
- Same as Asset A domain input
- Enter domain for the second asset in the pair

### 5. Asset B - Selection
- Same as Asset A selection
- Select the second asset

### 6. Confirmation Screen
- Displays:
  - Pair name (e.g., USDC / USDZ)
  - Current best bid from order book
  - Current best ask from order book
  - LP Locked amounts (if liquidity pool exists)
- Press `enter` to confirm and save
- Press `esc` to go back

### 7. Success
- Pair is saved to `~/.config/sdexmon/config.yaml`
- Returns to landing screen
- **Note:** Currently requires app restart to see new pair in selector

## File Structure

### New Files Created

```
internal/
├── models/
│   └── maintenance.go          # Data structures for maintenance mode
├── stellar/
│   ├── expert.go               # stellar.expert API integration
│   └── confirmation.go         # Market data fetching for confirmation
└── config/
    └── user_config.go          # User config persistence

cmd/sdexmon/
├── maintenance_view.go         # TUI view rendering functions
└── maintenance_update.go       # Event handlers and commands
```

### Modified Files

- `cmd/sdexmon/main.go`:
  - Added `models.MaintenanceState` to model
  - Added `screenMaintenance` constant
  - Added 'm' key handler in landing and pair info screens
  - Updated `View()` to route to maintenance views
  - Updated bottom line shortcuts to show 'm: maintenance'

## Configuration Format

Custom pairs are saved in `~/.config/sdexmon/config.yaml`:

```yaml
custom_pairs:
  - asset_a: "USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"
    asset_b: "USDZ:GAKTLPC4ZV37SSCITQ5IS5AQ4WPF4CF4VZJQPPAROSGXMYOATF5U6XPR"
    label: "USDC/USDZ"
  - asset_a: "native"
    asset_b: "ZARZ:GAROH4EV3WVVTRQKEY43GZK3XSRBEYETRVZ7SVG5LHWOAANSMCTJBB3U"
    label: "XLM/ZARZ"
```

## API Integration

### stellar.expert API

**Endpoint:** `https://api.stellar.expert/explorer/public/asset?search=<domain>`

Used to search for assets by domain name. Returns asset details including:
- Code
- Issuer
- Domain
- Supply
- Trustlines
- Name

### Horizon API

Used to fetch current order book data (best bid/ask) for the confirmation screen.

## Keyboard Shortcuts

### Maintenance Menu
- `1`: Add asset pair
- `esc` or `q`: Back to landing screen

### Domain Input Screens
- `enter`: Search for assets
- `esc`: Go back
- (text input): Type domain name

### Asset Selection Screens
- `↑`/`↓` or `k`/`j`: Navigate asset list
- `enter`: Select asset
- `esc`: Go back to domain input

### Confirmation Screen
- `enter`: Confirm and save pair
- `esc`: Go back to previous screen
- `q`: Quit

## Known Limitations

1. **Requires restart**: New custom pairs only appear in the pair selector after restarting sdexmon
2. **Remove pairs**: Not yet implemented (option 2 in menu is greyed out)
3. **View custom pairs**: Not yet implemented (option 3 in menu is greyed out)
4. **LP data**: Confirmation screen shows "--" for LP locked amounts (could be enhanced to fetch from stellar.expert)
5. **Native assets**: Currently all assets from stellar.expert are treated as credit assets. Native XLM support could be enhanced.

## Future Enhancements

- [ ] Dynamic reload of custom pairs without restart
- [ ] Remove custom pair functionality
- [ ] View/manage all custom pairs
- [ ] Fetch full LP data from stellar.expert in confirmation
- [ ] Better handling of native XLM asset selection
- [ ] Search history / recent domains
- [ ] Favorite/bookmark frequently used assets
- [ ] Validate assets exist on-chain before saving
- [ ] Bulk import/export of custom pairs

## Testing

To test the feature:

```bash
./run
# Press 'm' to enter maintenance mode
# Follow the workflow to add a pair
# Check ~/.config/sdexmon/config.yaml to verify it was saved
# Restart and verify pair appears in selector
```

## Error Handling

The maintenance mode includes error handling for:
- Empty domain input
- stellar.expert API failures
- Network errors
- Duplicate pairs (prevents adding same pair twice)
- Invalid asset formats
- YAML save failures

Errors are displayed in red at the bottom of each screen.
