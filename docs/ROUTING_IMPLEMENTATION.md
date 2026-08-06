# Routing System Implementation Summary

## Overview
Successfully implemented a comprehensive routing system for the SCAR AQUILA TUI application with proper screen navigation, standardized UI components, and a state machine architecture.

## Implementation Details

### 1. Screen State Architecture
- **Screen States**: Created `screenState` enum with 8 distinct screens:
  - `screenServiceSelection` - Landing page
  - `screenSelectPair` - Asset pair selection
  - `screenPairInput` - Custom pair input
  - `screenPairInfo` - Pair monitoring (orderbook, trades, LP)
  - `screenPairDebug` - Debug info for pairs
  - `screenSelectAsset` - Single asset selection
  - `screenViewExposure` - LP exposure view for selected asset
  - `screenExposureDebug` - Debug info for exposure

### 2. Model Structure Updates
- Added `currentScreen` field to track active screen
- Added `selectedAsset` for single asset exposure functionality
- Added `exposurePools []Liquidity` for exposure data
- Added `assetIndex` for asset selection navigation
- Removed legacy flags: `pairSelectorActive`, `pairInputActive`, `debugPageActive`

### 3. Reusable UI Components
Created three standard components for consistent screen layout:

```go
renderHeader()           // SCAR AQUILA ASCII art
renderSubtitle(title)    // Screen title
renderFooter(shortcuts, isLive)  // Context-aware shortcuts + status
```

Every screen now has:
1. SCAR AQUILA header
2. Screen-specific subtitle
3. Content area
4. Context-aware footer with shortcuts

### 4. Navigation State Machine
Implemented in `Update()` function with screen-specific key handlers:

#### Service Selection
- `1` → Select Pair
- `2` → Select Asset
- `q` → Quit

#### Select Pair
- `↑/↓` → Navigate pairs
- `enter` → Go to Pair Info (starts data polling)
- `a` → Custom pair input
- `esc` → Back to Service Selection
- `q` → Quit

#### Pair Input
- `enter` → Apply and go to Pair Info
- `tab` → Switch between base/quote fields
- `esc` → Back to Select Pair
- `q` → Quit

#### Pair Info
- `b` → Back to Select Pair
- `z` → Toggle to Pair Debug
- `,/.` → Adjust depth
- `q` → Quit

#### Pair Debug
- `z` → Back to Pair Info
- `b` → Back to Select Pair
- `q` → Quit

#### Select Asset
- `↑/↓` → Navigate assets
- `enter` → Go to View Exposure
- `esc` → Back to Service Selection
- `q` → Quit

#### View Exposure
- `b` → Back to Select Asset
- `z` → Toggle to Exposure Debug
- `q` → Quit

#### Exposure Debug
- `z` → Back to View Exposure
- `q` → Quit

### 5. New Features

#### Single Asset Exposure
- New functionality to view all liquidity pools containing a specific asset
- `fetchExposureCmd()` searches liquidityPoolIDs map for matching pools
- Fetches pool data from stellar.expert API
- Displays locked amounts for each asset in the pools
- Shows up to 10 pools per asset

#### Messages
Added new message types:
- `exposureDataMsg` - Contains fetched exposure data

### 6. Screen Views
All screen views refactored with consistent structure:

- `serviceSelectionView()` - New landing page
- `pairSelectorView()` - Refactored with new components
- `pairInputView()` - Refactored with new components
- `pairInfoView()` - Main pair monitoring view
- `pairDebugView()` - Pair debug information
- `selectAssetView()` - New asset selection screen
- `viewExposureView()` - New exposure display
- `exposureDebugView()` - New exposure debug screen

### 7. View Router
Simplified `View()` function with switch statement on `currentScreen`

## Navigation Flow

```
./run
  ↓
Service Selection
  ├─ [1] → Select Pair → Pair Info ⇄ Pair Debug
  │                         ↑          ↓
  │                         └──────────┘
  └─ [2] → Select Asset → View Exposure ⇄ Exposure Debug
                            ↑              ↓
                            └──────────────┘
```

## Testing
- ✅ Code compiles successfully
- ✅ All screens implemented
- ✅ Navigation paths defined
- ✅ go fmt applied
- ✅ go vet passes with no warnings

## Usage
```bash
./run         # Start with Service Selection screen
./tui         # Direct run without launcher
```

## Environment Variables (unchanged)
- `HORIZON_URL` - Horizon endpoint (defaults to https://horizon.stellar.org)
- `BASE_ASSET`, `QUOTE_ASSET` - If set, starts at Pair Info
- `LP_POOL_ID` - Override pool ID
- `DEBUG` - Enable debug mode

## Key Benefits
1. **Consistent UI**: Every screen has header, subtitle, and footer
2. **Clear Navigation**: State machine enforces valid transitions
3. **Maintainable**: Centralized routing logic
4. **Extensible**: Easy to add new screens
5. **User-Friendly**: Context-aware shortcuts in footer
6. **New Functionality**: Single asset exposure tracking

## Files Modified
- `main.go` - Complete refactor of routing system

## Notes
- Existing pair monitoring functionality preserved
- All original features (orderbook, trades, LP) still work
- Data polling continues to work for pair screens
- Debug mode compatible with new routing
