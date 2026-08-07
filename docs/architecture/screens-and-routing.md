# Screens and routing

Routing is a flat state machine on `model.currentScreen`, plus one nested state
machine for maintenance. The pair selector is not a screen; it is a popup
overlay controlled by `model.showPairPopup`.

## Screen states

Declared in `cmd/sdexmon/main.go`:

```go
screenLanding      // sdexmon banner, pair selector popup host
screenPairInfo     // order book, trades, liquidity pool, exposure
screenPairDebug    // raw diagnostics for the current pair
screenPairInput    // custom base/quote entry (currently unreachable)
screenUpgrade      // advisory update notice
screenMaintenance  // add, remove and list configured pairs
```

`View()` switches on `currentScreen` and falls back to the landing view for any
unrecognised state.

## Entry point

```
start
  |
  +-- display mode and a resolved pair -> screenPairInfo
  +-- otherwise                        -> screenLanding
  |
  +-- update available and not display mode -> screenUpgrade first
```

`upgradeReturn` remembers the screen the upgrade notice was opened from, so
`esc` returns there rather than always dropping to the landing screen.

## Transitions

```
screenLanding
  enter -> open pair selector popup
  m     -> screenMaintenance
  u     -> screenUpgrade (only when an update is available)

pair selector popup (landing or pair info)
  s     -> search mode
  enter -> select pair, close popup, screenPairInfo
  esc   -> close popup (or leave search mode when searching)

screenPairInfo
  p     -> open pair selector popup
  d     -> screenPairDebug
  m     -> screenMaintenance
  u     -> screenUpgrade (only when an update is available)

screenPairDebug
  d     -> screenPairInfo
  u     -> screenUpgrade (only when an update is available)

screenUpgrade
  enter -> run the installer, then quit
  esc   -> upgradeReturn

screenPairInput (unreachable: nothing sets currentScreen to it)
  tab   -> switch base/quote field
  enter -> parse both, screenPairInfo
  esc   -> screenLanding with the popup open
```

`q` and `ctrl+c` quit from every screen. `q` is suppressed only while a
maintenance text field has focus, so domains and asset codes containing `q` can
still be typed. That exception does not apply to the pair selector search, where
`q` still quits.

## Maintenance sub-state machine

`screenMaintenance` delegates to `handleMaintenanceUpdate`, which switches on
`models.MaintenanceState.Screen`:

```
MaintenanceMenu
  1 -> AssetASourceSelect
  2 -> PairRemoveSelection
  3 -> PairList
  esc -> screenLanding, state reset

AssetASourceSelect / AssetBSourceSelect
  1-3 or enter -> AssetAQueryInput / AssetBQueryInput, or straight to the
                  selection screen for query-less sources such as Top 50
  esc          -> MaintenanceMenu (leg A) or AssetASelection (leg B)

AssetAQueryInput / AssetBQueryInput
  enter -> run the search, results land on the matching selection screen
  esc   -> back to the matching source select

AssetASelection
  enter -> AssetBSourceSelect
  esc   -> AssetASourceSelect

AssetBSelection
  enter -> PairConfirmation, fetch market data
  esc   -> AssetBSourceSelect

PairConfirmation
  enter -> save, reload pairs, MaintenanceMenu with a status message
  esc   -> AssetBSelection

PairRemoveSelection
  enter -> PairRemoveConfirmation
  esc   -> MaintenanceMenu

PairRemoveConfirmation
  y or enter -> remove, reload pairs, MaintenanceMenu
  n or esc   -> PairRemoveSelection

PairList
  esc -> MaintenanceMenu
```

### Asynchronous search results

Searches run off the UI thread and return `models.AssetSearchResultsMsg`, which
carries the `AssetLeg` that requested it. `applySearchResults` files results
under that leg and only advances the screen when the user is still waiting on
that same search, so a late response cannot yank them out of somewhere else.

These messages, along with `ConfirmationDataMsg` and `MaintenanceErrMsg`, are
routed from the top-level `Update` regardless of the current screen, so results
are never lost if the user navigates away.

## Standard screen frame

Every screen except the upgrade notice renders the same vertical frame:

1. Version line, dimmed: `<version> (build <commit>)`.
2. Blank line.
3. The `sdexmon` ASCII banner.
4. A screen-specific subtitle.
5. Content.
6. Padding to push the footer to the bottom.
7. Footer: context-aware shortcuts on the left, network capacity on the right.

`bottomLine` selects the shortcut string from the current screen and popup
state, prepends `u: upgrade` when an update is available, and replaces the whole
left side with a data-freshness string in display mode.

The upgrade notice is rendered by `internal/ui.RenderUpgradeAvailable`, which
centres an amber-bordered box and does not use the standard frame.

## Adding a screen

1. Add a constant to the `screenState` block.
2. Add a `case` to the screen switch in `Update` for its key handling.
3. Add a `case` to `View()` returning the renderer.
4. Add a `case` to `bottomLine` for its footer shortcuts.
5. Add the transitions into and out of it, and update this document.
