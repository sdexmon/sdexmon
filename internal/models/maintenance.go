package models

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/stellar/go/txnbuild"
)

// MaintenanceScreen represents the maintenance mode screen states
type MaintenanceScreen int

const (
	MaintenanceMenu MaintenanceScreen = iota
	AssetADomainInput
	AssetASelection
	AssetBDomainInput
	AssetBSelection
	PairConfirmation
)

// StellarExpertAsset represents an asset from stellar.expert API
type StellarExpertAsset struct {
	Code       string  `json:"code"`
	Issuer     string  `json:"issuer"`
	Domain     string  `json:"domain"`
	Supply     string  `json:"supply"`
	Trustlines int     `json:"trustlines"`
	Name       string  `json:"name"`
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
	Screen           MaintenanceScreen
	DomainInputA     textinput.Model
	DomainInputB     textinput.Model
	SearchResultsA   []StellarExpertAsset
	SearchResultsB   []StellarExpertAsset
	SelectedAssetA   *StellarExpertAsset
	SelectedAssetB   *StellarExpertAsset
	AssetCursorA     int
	AssetCursorB     int
	ConfirmationData *PairConfirmationData
	LoadingMessage   string
	ErrorMessage     string
}

// Messages for maintenance mode
type (
	AssetSearchResultsMsg struct{ Assets []StellarExpertAsset }
	ConfirmationDataMsg   struct{ Data *PairConfirmationData }
	MaintenanceErrMsg     struct{ Err error }
)
