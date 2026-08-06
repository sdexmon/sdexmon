package models

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/stellar/go/txnbuild"
)

// MaintenanceScreen represents the maintenance mode screen states
type MaintenanceScreen int

const (
	MaintenanceMenu MaintenanceScreen = iota
	AssetASourceSelect
	AssetAQueryInput
	AssetASelection
	AssetBSourceSelect
	AssetBQueryInput
	AssetBSelection
	PairConfirmation
	PairRemoveSelection
	PairRemoveConfirmation
	PairList
)

// AssetLeg identifies which side of the pair a search belongs to. Async results
// carry it so they land on the right list no matter which screen is showing.
type AssetLeg int

const (
	AssetLegA AssetLeg = iota
	AssetLegB
)

// AssetSearchMode is how the user wants to look an asset up.
type AssetSearchMode int

const (
	// SearchByDomain reads the domain's own SEP-1 stellar.toml. Authoritative:
	// only the domain owner can publish an issuer there.
	SearchByDomain AssetSearchMode = iota
	// SearchByCode is a fuzzy stellar.expert search. Convenient but unverified,
	// so results always show their home domain.
	SearchByCode
	// SearchTop50 lists the stellar.expert top 50 assets by activity.
	SearchTop50
)

// AllAssetSearchModes is the menu order of the search sources.
var AllAssetSearchModes = []AssetSearchMode{SearchByDomain, SearchByCode, SearchTop50}

// Label is the short menu entry for a search mode.
func (m AssetSearchMode) Label() string {
	switch m {
	case SearchByDomain:
		return "Domain search (stellar.toml)"
	case SearchByCode:
		return "Asset search (code or name)"
	case SearchTop50:
		return "stellar.expert Top 50"
	default:
		return "Unknown"
	}
}

// Description explains the trust level of a search mode.
func (m AssetSearchMode) Description() string {
	switch m {
	case SearchByDomain:
		return "Only assets the domain publishes in its own SEP-1 stellar.toml"
	case SearchByCode:
		return "Fuzzy search over all assets - always check the domain column"
	case SearchTop50:
		return "Most active assets on the network, ranked by stellar.expert"
	default:
		return ""
	}
}

// NeedsQuery reports whether the mode has to prompt for text before searching.
func (m AssetSearchMode) NeedsQuery() bool {
	return m == SearchByDomain || m == SearchByCode
}

// AssetSearchResult is a selectable asset returned by any of the search modes.
type AssetSearchResult struct {
	Code            string
	Issuer          string
	Domain          string
	Name            string
	Org             string
	Status          string
	DisplayDecimals int
	Trustlines      int
	// Verified is true when the asset was read from its home domain's own
	// stellar.toml, the only source that proves the domain claims it.
	Verified bool
}

// PairConfirmationData holds data for the confirmation screen
type PairConfirmationData struct {
	AssetA    txnbuild.Asset
	AssetB    txnbuild.Asset
	BestBid   string
	BestAsk   string
	LPLockedA string
	LPLockedB string
	LPPoolID  string
}

// MaintenanceState holds all maintenance mode UI state
type MaintenanceState struct {
	Screen MaintenanceScreen

	// Search source selection
	SourceCursor int
	SearchModeA  AssetSearchMode
	SearchModeB  AssetSearchMode

	// Query entry, holding a domain or an asset code depending on the mode.
	QueryInputA textinput.Model
	QueryInputB textinput.Model

	SearchResultsA []AssetSearchResult
	SearchResultsB []AssetSearchResult
	SearchSourceA  string
	SearchSourceB  string
	SelectedAssetA *AssetSearchResult
	SelectedAssetB *AssetSearchResult
	AssetCursorA   int
	AssetCursorB   int

	ConfirmationData *PairConfirmationData
	// PairCursor is the highlighted row on the remove and list screens.
	PairCursor     int
	LoadingMessage string
	StatusMessage  string
	ErrorMessage   string
}

// AcceptsTextInput reports whether the current maintenance screen is typing
// into a text field. Single-letter global shortcuts must stand down while it
// is true, otherwise queries containing those letters cannot be entered.
func (s MaintenanceState) AcceptsTextInput() bool {
	return s.Screen == AssetAQueryInput || s.Screen == AssetBQueryInput
}

// Messages for maintenance mode
type (
	// AssetSearchResultsMsg carries search results back to the leg that asked
	// for them, with a human readable description of where they came from.
	AssetSearchResultsMsg struct {
		Leg    AssetLeg
		Assets []AssetSearchResult
		Source string
	}
	ConfirmationDataMsg struct{ Data *PairConfirmationData }
	MaintenanceErrMsg   struct{ Err error }
)
