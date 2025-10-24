package models

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
)

// Screen states
type ScreenState int

const (
	ScreenServiceSelection ScreenState = iota
	ScreenSelectPair
	ScreenPairInfo
	ScreenPairDebug
	ScreenSelectAsset
	ScreenViewExposure
	ScreenExposureDebug
	ScreenPairInput
)

// PairOption represents a trading pair
type PairOption struct {
	Base  string
	Quote string
}

// Liquidity holds display-ready strings for liquidity pool data
type Liquidity struct {
	Codes    [2]string
	Decimals [2]int
	Locked   [2]string
	Fees1d   [2]string
	Fees7d   [2]string
	Vol1d    [2]string
	Vol7d    [2]string
}

// Model represents the TUI application state
type Model struct {
	Client *horizonclient.Client

	// Screen state
	CurrentScreen ScreenState

	// Asset selection
	Base          txnbuild.Asset
	Quote         txnbuild.Asset
	SelectedAsset txnbuild.Asset // for single asset exposure view

	Orderbook hProtocol.OrderBookSummary
	Trades    []hProtocol.Trade

	TradeCursor string // paging token of last trade we processed

	// liquidity data
	LP            Liquidity
	LPPoolID      string
	LPMessage     string
	ExposurePools []Liquidity // for view exposure screen

	// debug log buffer
	DebugLogs []string

	Width  int
	Height int
	Depth  int

	// input and selection state
	PairIndex  int
	AssetIndex int
	BaseInput  textinput.Model
	QuoteInput textinput.Model

	// liveness
	LastOrderbookAt time.Time
	LastTradesAt    time.Time
	LastLPAt        time.Time

	// debug modes
	DebugMode bool

	Status string
	Err    error
}

// Messages for Bubble Tea
type (
	OrderbookTickMsg struct{}
	TradesTickMsg    struct{}
	LPTickMsg        struct{}
	OrderbookDataMsg struct{ OB hProtocol.OrderBookSummary }
	TradesDataMsg    struct{ List []hProtocol.Trade }
	LPDataMsg        struct{ Data Liquidity }
	LPNoteMsg        string
	ExposureDataMsg  struct{ Pools []Liquidity }
	ErrMsg           error
)
