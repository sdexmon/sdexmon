# Pair Selector Search Feature

## Overview

Added a search function to the Pair Selector popup that allows users to quickly filter trading pairs by searching for either asset in the pair.

## Usage

### Activating Search

1. Open the Pair Selector popup by pressing `enter` on the landing page or `p` on the Pair Info screen
2. Press `s` to activate search mode
3. A search input field (🔍) will appear at the top of the popup
4. Start typing to filter pairs in real-time

### Search Behavior

- **Case-insensitive**: Search queries are automatically converted to uppercase
- **Matches both sides**: The search looks for matches in either the base or quote asset
- **Real-time filtering**: Results update as you type
- **Scrollable results**: Use ↑/↓ to navigate through filtered pairs

### Example Searches

- Type `USD` → Shows pairs containing USDC, USDZ (e.g., USDC/USDZ, USDZ/ZARZ, XLM/USDC)
- Type `XLM` → Shows all pairs with XLM (e.g., XLM/USDC, XLM/USDZ, XLM/EURZ)
- Type `Z` → Shows all "Z" assets (USDZ, ZARZ, EURZ, XAUZ, BTCZ)

### Keyboard Shortcuts

**When search is active:**
- `↑/↓` or `k/j`: Navigate through filtered results
- `enter`: Select the highlighted pair
- `esc`: Exit search mode (returns to full pair list)
- `q`: Quit application

**When search is inactive:**
- `↑/↓` or `k/j`: Navigate through all pairs
- `s`: Activate search mode
- `enter`: Select the highlighted pair
- `esc`: Close popup
- `q`: Quit application

## Implementation Details

### New Model Fields

```go
searchInput    textinput.Model  // search input for pair selector
searchMode     bool             // whether search is active in pair selector
filteredPairs  []pairOption     // filtered list based on search
```

### Key Functions

- **`filterPairs()`**: Filters the configured pairs based on the search query
  - Trims and uppercases the search query
  - Matches against both base and quote assets
  - Resets pair index if it goes out of bounds

- **`pairSelectorPopup()`**: Updated to:
  - Show search input when in search mode
  - Use `filteredPairs` when searching, `configuredPairs` otherwise
  - Display "No pairs found" when filter returns no results
  - Show context-appropriate keyboard shortcuts

### State Management

The search state is properly reset when:
- Exiting search mode (press `esc`)
- Selecting a pair (press `enter`)
- Closing the popup

This ensures the user always starts with a clean state when reopening the selector.

## User Experience

### Visual Feedback

- Search input appears with 🔍 icon when active
- Selected pair is highlighted in cyan/bold
- Footer updates to show relevant shortcuts
- "No pairs found" message when search yields no results

### Smooth Workflow

1. User opens pair selector
2. Presses `s` to search
3. Types partial asset name (e.g., "XLM")
4. Sees filtered results immediately
5. Uses arrows to select
6. Presses `enter` to load pair data
7. App switches to Pair Info screen with live data

## Testing

Build and run to test:

```bash
./run
```

Or:

```bash
go run ./cmd/sdexmon
```

Then:
1. Press `enter` to open pair selector
2. Press `s` to activate search
3. Type various asset codes to test filtering
4. Verify navigation and selection work correctly
5. Test edge cases (no results, empty search, etc.)
