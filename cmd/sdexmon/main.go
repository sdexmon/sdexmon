package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
)

const (
	defaultDepth      = 7 // rows per side (limit to 7 each side)
	orderbookInterval = 1200 * time.Millisecond
	tradesInterval    = 1200 * time.Millisecond
	lpInterval        = 30 * time.Second
	maxTradesKept     = 120
)

// Screen states
type screenState int

const (
	screenServiceSelection screenState = iota
	screenSelectPair
	screenPairInfo
	screenPairDebug
	screenSelectAsset
	screenViewExposure
	screenExposureDebug
	screenPairInput // custom pair input screen
)

const asciiAquila = `███████  ██████  █████  ██████       █████   ██████  ██    ██ ██ ██       █████  
██      ██      ██   ██ ██   ██     ██   ██ ██    ██ ██    ██ ██ ██      ██   ██ 
███████ ██      ███████ ██████      ███████ ██    ██ ██    ██ ██ ██      ███████ 
     ██ ██      ██   ██ ██   ██     ██   ██ ██ ▄▄ ██ ██    ██ ██ ██      ██   ██ 
███████  ██████ ██   ██ ██   ██     ██   ██  ██████   ██████  ██ ███████ ██   ██ 
                                                ▀▀                               
                                                                                 `

// Curated assets and pairs (static table)

type pairOption struct{ Base, Quote string }

var curatedAssets = map[string]txnbuild.Asset{
	"USDZ": txnbuild.CreditAsset{Code: "USDZ", Issuer: "GAKTLPC4ZV37SSCITQ5IS5AQ4WPF4CF4VZJQPPAROSGXMYOATF5U6XPR"},
	"ZARZ": txnbuild.CreditAsset{Code: "ZARZ", Issuer: "GAROH4EV3WVVTRQKEY43GZK3XSRBEYETRVZ7SVG5LHWOAANSMCTJBB3U"},
	"EURZ": txnbuild.CreditAsset{Code: "EURZ", Issuer: "GAM5BKSKTHYS6IE4OUHCISGI6YVH75XIMOCG4RB5TR74KZDJRSNKEURZ"},
	"XAUZ": txnbuild.CreditAsset{Code: "XAUZ", Issuer: "GD3MMNHD5U5H732GTLYO7DZVUNGPVP462KVNFO4HALNPP6C7ESQAGOLD"},
	"BTCZ": txnbuild.CreditAsset{Code: "BTCZ", Issuer: "GAT63G6FINKAES4473ZZZT3SYJVUIXKYBVFBQYQHEZF6EE3VY5AGBTCZ"},
	"XLM":  txnbuild.NativeAsset{},
	"USDC": txnbuild.CreditAsset{Code: "USDC", Issuer: "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"},
}

var curatedPairs = []pairOption{
	{"USDC", "USDZ"},
	{"USDZ", "ZARZ"},
	{"USDZ", "EURZ"},
	{"USDZ", "BTCZ"},
	{"USDZ", "XAUZ"},
	{"EURZ", "ZARZ"},
	{"EURZ", "XAUZ"},
	{"EURZ", "BTCZ"},
	{"ZARZ", "XAUZ"},
	{"ZARZ", "BTCZ"},
	{"XAUZ", "BTCZ"},
	{"XLM", "USDC"},
	{"XLM", "USDZ"},
	{"XLM", "EURZ"},
	{"XLM", "ZARZ"},
	{"XLM", "XAUZ"},
	{"XLM", "BTCZ"},
}

var liquidityPoolIDs = map[string]string{
	"USDC-USDZ": "314e17d86ffc767a6132fba31cc9f53f23ca359d2db788f26f0d364d75e82c57",
	"USDZ-USDC": "314e17d86ffc767a6132fba31cc9f53f23ca359d2db788f26f0d364d75e82c57",
	"USDZ-ZARZ": "d6842cf8f10ec2fc8a4599f23f7b0161bafa228b267714fc3ed6ca6d48d0b13c",
	"ZARZ-USDZ": "d6842cf8f10ec2fc8a4599f23f7b0161bafa228b267714fc3ed6ca6d48d0b13c",
	"USDZ-EURZ": "30869ce3dd1e130649c08ca0986bcb912acd4c557502378d8e32f41e1c443f55",
	"EURZ-USDZ": "30869ce3dd1e130649c08ca0986bcb912acd4c557502378d8e32f41e1c443f55",
	"USDZ-BTCZ": "645923faa8b51f09f63306db95788bf4d8aa033ff50031ac279dcdb483207f10",
	"BTCZ-USDZ": "645923faa8b51f09f63306db95788bf4d8aa033ff50031ac279dcdb483207f10",
	"USDZ-XAUZ": "f0344bb1fbde3157c745ca7c310e9516877ef30ae35cacf3f268b4b163d30788",
	"XAUZ-USDZ": "f0344bb1fbde3157c745ca7c310e9516877ef30ae35cacf3f268b4b163d30788",
	"EURZ-ZARZ": "57b50011b18e2e6a94b4cf745a569779a50d710c757caa37d38148d24d383cc9",
	"ZARZ-EURZ": "57b50011b18e2e6a94b4cf745a569779a50d710c757caa37d38148d24d383cc9",
	"EURZ-XAUZ": "1c473914c3af39f5ed04284f01f8488906ec9ddeae31e3f4dc608e9871ba4a68",
	"XAUZ-EURZ": "1c473914c3af39f5ed04284f01f8488906ec9ddeae31e3f4dc608e9871ba4a68",
	"EURZ-BTCZ": "3c3d8532451361b47986d1c864e029488453fcf923bca383af673a4fe84ef8c1",
	"BTCZ-EURZ": "3c3d8532451361b47986d1c864e029488453fcf923bca383af673a4fe84ef8c1",
	"ZARZ-XAUZ": "962528fd96913f256044daf4aa77162be04c381764fef6f92b6962b4d6c50fb1",
	"XAUZ-ZARZ": "962528fd96913f256044daf4aa77162be04c381764fef6f92b6962b4d6c50fb1",
	"ZARZ-BTCZ": "39b4a2889462d58dffb9e11a97502f2a74788d9c2b6c6b711ba2e7b0cfe2a7d8",
	"BTCZ-ZARZ": "39b4a2889462d58dffb9e11a97502f2a74788d9c2b6c6b711ba2e7b0cfe2a7d8",
	"XAUZ-BTCZ": "a4753a9faa6b256e46fb63ab900c64333d5d799ee48b70452d3fa833db350f33",
	"BTCZ-XAUZ": "a4753a9faa6b256e46fb63ab900c64333d5d799ee48b70452d3fa833db350f33",
	"XLM-USDC":  "a468d41d8e9b8f3c7209651608b74b7db7ac9952dcae0cdf24871d1d9c7b0088",
	"USDC-XLM":  "a468d41d8e9b8f3c7209651608b74b7db7ac9952dcae0cdf24871d1d9c7b0088",
	"XLM-USDZ":  "7001fca2d71456cda8a061e4733f035fce36423ccf942e92db139a116d7e557b",
	"USDZ-XLM":  "7001fca2d71456cda8a061e4733f035fce36423ccf942e92db139a116d7e557b",
	"XLM-EURZ":  "d79c741bc6371240af4a1e86c645742a561581095bd147ae86a0a3386701c545",
	"EURZ-XLM":  "d79c741bc6371240af4a1e86c645742a561581095bd147ae86a0a3386701c545",
	"XLM-ZARZ":  "fb7072d551e853826e4a5497e2da1e6765c8cc29fa938ceeeeef579adc53a9f6",
	"ZARZ-XLM":  "fb7072d551e853826e4a5497e2da1e6765c8cc29fa938ceeeeef579adc53a9f6",
	"XLM-XAUZ":  "fb0e4a67424a2851cfa02618de758f2cbaa71e737454caf25919fa51bab125e5",
	"XAUZ-XLM":  "fb0e4a67424a2851cfa02618de758f2cbaa71e737454caf25919fa51bab125e5",
	"XLM-BTCZ":  "d8905565dac7e4c5702520bdf39d1e8b385a94708628c87333862a41b62da980",
	"BTCZ-XLM":  "d8905565dac7e4c5702520bdf39d1e8b385a94708628c87333862a41b62da980",
}

// Messages

type (
	orderbookTickMsg struct{}
	tradesTickMsg    struct{}
	lpTickMsg        struct{}
	orderbookDataMsg struct{ ob hProtocol.OrderBookSummary }
	tradesDataMsg    struct{ list []hProtocol.Trade }
	lpDataMsg        struct{ data Liquidity }
	lpNoteMsg        string
	exposureDataMsg  struct{ pools []Liquidity }
	errMsg           error
)

// Model

type model struct {
	client *horizonclient.Client

	// Screen state
	currentScreen screenState

	// Asset selection
	base          txnbuild.Asset
	quote         txnbuild.Asset
	selectedAsset txnbuild.Asset // for single asset exposure view

	orderbook hProtocol.OrderBookSummary
	trades    []hProtocol.Trade

	tradeCursor string // paging token of last trade we processed

	// liquidity data
	lp            Liquidity
	lpPoolID      string
	lpMessage     string
	exposurePools []Liquidity // for view exposure screen

	// debug log buffer
	debugLogs []string

	width  int
	height int
	depth  int

	// input and selection state
	pairIndex  int
	assetIndex int
	baseInput  textinput.Model
	quoteInput textinput.Model

	// liveness
	lastOrderbookAt time.Time
	lastTradesAt    time.Time
	lastLPAt        time.Time

	// debug modes
	debugMode bool

	status string
	err    error
}

func initialModel(client *horizonclient.Client, base, quote txnbuild.Asset) model {
	b := textinput.New()
	b.Placeholder = "native or CODE:ISSUER (base)"
	b.Prompt = "BASE > "
	b.CharLimit = 80

	q := textinput.New()
	q.Placeholder = "native or CODE:ISSUER (quote)"
	q.Prompt = "QUOTE > "
	q.CharLimit = 80

	// Check for debug mode
	debugMode := os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1"

	// Setup custom log writer if debug mode
	if debugMode {
		setupDebugLogger()
	}

	// Always start at Service Selection
	// Environment variables BASE_ASSET/QUOTE_ASSET are stored as defaults
	// but don't skip the landing page
	initialScreen := screenServiceSelection

	return model{
		client:        client,
		currentScreen: initialScreen,
		base:          base,
		quote:         quote,
		trades:        make([]hProtocol.Trade, 0, 64),
		depth:         defaultDepth,
		baseInput:     b,
		quoteInput:    q,
		debugMode:     debugMode,
		debugLogs:     make([]string, 0, 100),
		exposurePools: make([]Liquidity, 0),
		status:        "Select service to begin",
	}
}

func (m model) Init() tea.Cmd {
	// Don't start polling at initialization
	// Polling starts when user selects a pair and navigates to Pair Info screen
	return nil
}

// Update

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		// Screen-specific navigation
		switch m.currentScreen {
		case screenServiceSelection:
			switch msg.String() {
			case "1":
				m.currentScreen = screenSelectPair
				m.pairIndex = currentPairIndex(m.base, m.quote)
				return m, nil
			case "2":
				m.currentScreen = screenSelectAsset
				m.assetIndex = 0
				return m, nil
			}

		case screenSelectPair:
			switch msg.String() {
			case "esc":
				m.currentScreen = screenServiceSelection
				return m, nil
			case "a":
				m.currentScreen = screenPairInput
				m.baseInput.SetValue(assetString(m.base))
				m.quoteInput.SetValue(assetString(m.quote))
				m.baseInput.Focus()
				return m, nil
			case "up", "k":
				if m.pairIndex > 0 {
					m.pairIndex--
				}
				return m, nil
			case "down", "j":
				if m.pairIndex < len(curatedPairs)-1 {
					m.pairIndex++
				}
				return m, nil
			case "enter":
				if len(curatedPairs) > 0 {
					opt := curatedPairs[m.pairIndex]
					base, ok1 := curatedAssets[opt.Base]
					quote, ok2 := curatedAssets[opt.Quote]
					if ok1 && ok2 {
						m.base, m.quote = base, quote
						m.tradeCursor = ""
						m.currentScreen = screenPairInfo
						m.status = "pair selected"
						return m, tea.Batch(
							fetchOrderbookCmd(m.client, m.base, m.quote),
							fetchTradesCmd(m.client, m.base, m.quote, m.tradeCursor, true),
							resolveAndFetchLPCmd(m.client, m.base, m.quote),
							tea.Tick(orderbookInterval, func(time.Time) tea.Msg { return orderbookTickMsg{} }),
							tea.Tick(tradesInterval, func(time.Time) tea.Msg { return tradesTickMsg{} }),
						)
					}
				}
				return m, nil
			}

		case screenPairInput:
			switch msg.String() {
			case "esc":
				m.currentScreen = screenSelectPair
				return m, nil
			case "enter":
				base, err1 := parseAsset(strings.TrimSpace(m.baseInput.Value()))
				quote, err2 := parseAsset(strings.TrimSpace(m.quoteInput.Value()))
				if err1 != nil {
					m.err = fmt.Errorf("base asset: %w", err1)
					return m, nil
				}
				if err2 != nil {
					m.err = fmt.Errorf("quote asset: %w", err2)
					return m, nil
				}
				m.base, m.quote = base, quote
				m.tradeCursor = ""
				m.currentScreen = screenPairInfo
				m.status = "pair updated"
				return m, tea.Batch(
					fetchOrderbookCmd(m.client, m.base, m.quote),
					fetchTradesCmd(m.client, m.base, m.quote, m.tradeCursor, true),
					resolveAndFetchLPCmd(m.client, m.base, m.quote),
					tea.Tick(orderbookInterval, func(time.Time) tea.Msg { return orderbookTickMsg{} }),
					tea.Tick(tradesInterval, func(time.Time) tea.Msg { return tradesTickMsg{} }),
				)
			case "tab":
				if m.baseInput.Focused() {
					m.baseInput.Blur()
					m.quoteInput.Focus()
				} else {
					m.quoteInput.Blur()
					m.baseInput.Focus()
				}
				return m, nil
			}
			// Pass other keys to text inputs
			var cmd1, cmd2 tea.Cmd
			m.baseInput, cmd1 = m.baseInput.Update(msg)
			m.quoteInput, cmd2 = m.quoteInput.Update(msg)
			return m, tea.Batch(cmd1, cmd2)

		case screenPairInfo:
			switch msg.String() {
			case "b":
				m.currentScreen = screenSelectPair
				return m, nil
			case "z":
				m.currentScreen = screenPairDebug
				return m, nil
			case ",":
				if m.depth > 1 {
					m.depth -= 1
				}
				return m, nil
			case ".":
				if m.depth < 7 {
					m.depth += 1
				}
				return m, nil
			}

		case screenPairDebug:
			switch msg.String() {
			case "z":
				m.currentScreen = screenPairInfo
				return m, nil
			case "b":
				m.currentScreen = screenSelectPair
				return m, nil
			}

		case screenSelectAsset:
			assetList := []string{"XLM", "USDZ", "ZARZ", "EURZ", "XAUZ", "BTCZ", "USDC"}
			switch msg.String() {
			case "esc":
				m.currentScreen = screenServiceSelection
				return m, nil
			case "up", "k":
				if m.assetIndex > 0 {
					m.assetIndex--
				}
				return m, nil
			case "down", "j":
				if m.assetIndex < len(assetList)-1 {
					m.assetIndex++
				}
				return m, nil
			case "enter":
				if m.assetIndex < len(assetList) {
					assetCode := assetList[m.assetIndex]
					if asset, ok := curatedAssets[assetCode]; ok {
						m.selectedAsset = asset
						m.currentScreen = screenViewExposure
						m.status = "loading exposure"
						return m, fetchExposureCmd(m.client, asset)
					}
				}
				return m, nil
			}

		case screenViewExposure:
			switch msg.String() {
			case "b":
				m.currentScreen = screenSelectAsset
				return m, nil
			case "z":
				m.currentScreen = screenExposureDebug
				return m, nil
			}

		case screenExposureDebug:
			switch msg.String() {
			case "z":
				m.currentScreen = screenViewExposure
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case orderbookTickMsg:
		return m, tea.Batch(
			fetchOrderbookCmd(m.client, m.base, m.quote),
			tea.Tick(orderbookInterval, func(time.Time) tea.Msg { return orderbookTickMsg{} }),
		)
	case tradesTickMsg:
		return m, tea.Batch(
			fetchTradesCmd(m.client, m.base, m.quote, m.tradeCursor, false),
			tea.Tick(tradesInterval, func(time.Time) tea.Msg { return tradesTickMsg{} }),
		)
	case lpTickMsg:
		return m, tea.Batch(
			resolveAndFetchLPCmd(m.client, m.base, m.quote),
			tea.Tick(lpInterval, func(time.Time) tea.Msg { return lpTickMsg{} }),
		)

	case orderbookDataMsg:
		m.orderbook = msg.ob
		m.lastOrderbookAt = time.Now()
		m.err = nil
		return m, nil
	case tradesDataMsg:
		if len(msg.list) > 0 {
			// append and cap
			m.trades = append(m.trades, msg.list...)
			if len(m.trades) > maxTradesKept {
				m.trades = m.trades[len(m.trades)-maxTradesKept:]
			}
			// advance cursor
			m.tradeCursor = msg.list[len(msg.list)-1].PagingToken()
		}
		m.lastTradesAt = time.Now()
		m.err = nil
		return m, nil
	case lpDataMsg:
		m.lp = msg.data
		m.lpMessage = ""
		m.lastLPAt = time.Now()
		return m, nil
	case lpNoteMsg:
		m.lpMessage = string(msg)
		return m, nil
	case exposureDataMsg:
		m.exposurePools = msg.pools
		m.err = nil
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// View

func (m model) View() string {
	switch m.currentScreen {
	case screenServiceSelection:
		return serviceSelectionView(m)
	case screenSelectPair:
		return pairSelectorView(m)
	case screenPairInput:
		return pairInputView(m)
	case screenPairInfo:
		return pairInfoView(m)
	case screenPairDebug:
		return pairDebugView(m)
	case screenSelectAsset:
		return selectAssetView(m)
	case screenViewExposure:
		return viewExposureView(m)
	case screenExposureDebug:
		return exposureDebugView(m)
	default:
		return serviceSelectionView(m)
	}
}

func pairInfoView(m model) string {
	if m.base == nil || m.quote == nil {
		// Fallback if somehow we're here without a pair
		return serviceSelectionView(m)
	}

	subtitle := fmt.Sprintf("Pair Info - %s/%s", assetShort(m.base), assetShort(m.quote))

	ob := m.renderOrderbook()
	tr := m.renderTrades()
	lp := m.renderLiquidity()

	// layout: two rows
	// Top row: ORDER BOOK (left) and TRADES (right)
	// Bottom row: LIQUIDITY POOL (full width)
	leftW := 66
	rightW := 44

	leftTop := panelStyle.Width(leftW).Render(ob)
	rightTop := panelStyle.Width(rightW).Render(tr)
	row1 := lipgloss.JoinHorizontal(lipgloss.Left, leftTop, " ", rightTop)

	// Full width panel
	lpW := leftW + rightW + 1 // Combined width + spacer
	row2 := panelStyle.Width(lpW).Render(lp)

	bottom := m.bottomLine()

	// Build content
	content := lipgloss.JoinVertical(lipgloss.Left,
		renderHeader(),
		renderSubtitle(subtitle),
		row1,
		"", // 1 row spacer
		row2,
	)
	contentHeight := lipgloss.Height(content)
	targetHeight := 60
	if m.height > 0 {
		targetHeight = m.height
	}

	paddingLines := targetHeight - contentHeight - 2 // -2 for bottom line itself
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		padding,
		bottom,
	)
}

func (m model) renderOrderbook() string {
	allBids := m.orderbook.Bids
	allAsks := m.orderbook.Asks

	// Limit to 7 levels per side; we will pad to always show 7
	maxRows := 7
	bids := allBids
	asks := allAsks
	if len(bids) > maxRows {
		bids = bids[:maxRows]
	}
	if len(asks) > maxRows {
		asks = asks[:maxRows]
	}

	// Panel title and column headers
	priceUnit := assetShort(m.quote)
	amountUnit := assetShort(m.base)
	// Compute integer width for decimal alignment of prices across visible rows
	priceIntW := maxPriceIntWidth(bids, asks, maxRows)
	fracW := 7
	priceW := priceIntW + 1 + fracW
	amountW, totalW := 16, 16
	barW := 12

	title := boldStyle.Render("ORDER BOOK")
	head := lipgloss.JoinHorizontal(lipgloss.Top,
		dimStyle.Render(padRightVis("PRICE ("+priceUnit+")", priceW)),
		padRight("", 2),
		dimStyle.Render(padRightVis("AMOUNT ("+amountUnit+")", amountW)),
		padRight("", 2),
		dimStyle.Render(padRightVis("TOTAL (cum)", totalW)),
	)

	rows := []string{title, head}

	// ----- ASKS (upwards): render 7 rows: pad missing at the top, then worst->best -----
	nA := minInt(len(asks), maxRows)
	padA := maxRows - nA
	// build best-first slice and cumulative from best outward
	asksBest := make([]hProtocol.PriceLevel, nA)
	copy(asksBest, allAsks[:nA])
	askCumBest := make([]float64, nA)
	var sum float64
	for i := 0; i < nA; i++ {
		p, _ := strconv.ParseFloat(asksBest[i].Price, 64)
		a, _ := strconv.ParseFloat(asksBest[i].Amount, 64)
		sum += a * p
		askCumBest[i] = sum
	}
	askMax := 0.0
	if nA > 0 {
		askMax = askCumBest[nA-1]
	}
	// padding blanks (top empty asks)
	for k := 0; k < padA; k++ {
		rows = append(rows, orderbookBlankRow(priceW, amountW, totalW, barW))
	}
	// display worst->best: index from nA-1 down to 0
	for di := 0; di < nA; di++ {
		idx := nA - 1 - di // worst -> best
		a := asksBest[idx]
		pStr := alignDecimalFixed(a.Price, priceIntW, fracW)
		amtStr := formatWithCommas(formatAmount(a.Amount))
		cumStr := formatFloatWithCommas(askCumBest[idx])
		ratio := 0.0
		if askMax > 0 {
			ratio = askCumBest[idx] / askMax
		}
		bar := depthBar(barW, ratio, lipgloss.Color("52")) // red-ish
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			padLeftVis(redStyle.Render(pStr), priceW), padRight("", 2),
			padLeftVis(redStyle.Render(amtStr), amountW), padRight("", 2),
			padLeftVis(redStyle.Render(cumStr), totalW), padRight("", 2), bar,
		)
		rows = append(rows, row)
	}

	// ----- Spread line -----
	bestBid, bestAsk := "", ""
	if len(allBids) > 0 {
		bestBid = allBids[0].Price
	}
	if len(allAsks) > 0 {
		bestAsk = allAsks[0].Price
	}
	spreadPct := spreadPercent(bestBid, bestAsk)
	rows = append(rows, dimStyle.Render(fmt.Sprintf("Spread  %s", spreadPct)))

	// ----- BIDS (downwards): render best->worse, then pad missing below -----
	nB := minInt(len(bids), maxRows)
	bidCum := make([]float64, nB)
	sum = 0
	for i := 0; i < nB; i++ {
		p, _ := strconv.ParseFloat(bids[i].Price, 64)
		a, _ := strconv.ParseFloat(bids[i].Amount, 64)
		sum += a * p
		bidCum[i] = sum
	}
	bidMax := 0.0
	if nB > 0 {
		bidMax = bidCum[nB-1]
	}
	for i := 0; i < nB; i++ {
		b := bids[i]
		pStr := alignDecimalFixed(b.Price, priceIntW, fracW)
		amtStr := formatWithCommas(formatAmount(b.Amount))
		cumStr := formatFloatWithCommas(bidCum[i])
		ratio := 0.0
		if bidMax > 0 {
			ratio = bidCum[i] / bidMax
		}
		bar := depthBar(barW, ratio, lipgloss.Color("24")) // teal-ish
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			padLeftVis(greenStyle.Render(pStr), priceW), padRight("", 2),
			padLeftVis(greenStyle.Render(amtStr), amountW), padRight("", 2),
			padLeftVis(greenStyle.Render(cumStr), totalW), padRight("", 2), bar,
		)
		rows = append(rows, row)
	}
	// pad remaining empty bid rows
	for k := 0; k < maxRows-nB; k++ {
		rows = append(rows, orderbookBlankRow(priceW, amountW, totalW, barW))
	}

	return lipgloss.NewStyle().Render(strings.Join(rows, "\n"))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func truncateMiddle(s string, max int) string {
	if len(s) <= max {
		return s
	}
	half := (max - 3) / 2
	if half < 0 {
		return s
	}
	return s[:half] + "..." + s[len(s)-half:]
}

func prettyJSON(s string) string {
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(s), "", "  "); err != nil {
		return s
	}
	return out.String()
}

func (m model) renderLiquidity() string {
	title := boldStyle.Render("LIQUIDITY POOL")
	lines := []string{title}
	if m.lpMessage != "" {
		lines = append(lines, dimStyle.Render(m.lpMessage))
		return strings.Join(lines, "\n")
	}
	if len(m.lp.Codes) == 2 && m.lp.Codes[0] != "" && m.lp.Codes[1] != "" {
		// First section: LOCKED, FEES (1D), FEES (7D)
		// Column widths: code=10, locked=20, fees1d=18, fees7d=18
		header1 := padRightVis("", 22) + dimStyle.Render("LOCKED") +
			padRightVis("", 9) + dimStyle.Render("FEES (1D)") +
			padRightVis("", 9) + dimStyle.Render("FEES (7D)")
		lines = append(lines, header1)

		for i := 0; i < 2; i++ {
			code := padRightVis(m.lp.Codes[i], 10)
			locked := padLeftVis(m.lp.Locked[i], 20)
			fees1d := padLeftVis(m.lp.Fees1d[i], 18)
			fees7d := padLeftVis(m.lp.Fees7d[i], 18)
			row := code + locked + padRightVis("", 2) + fees1d + padRightVis("", 2) + fees7d
			lines = append(lines, row)
		}

		lines = append(lines, "")

		// Second section: VOLUME (1D), VOLUME (7D)
		// Column widths: code=10, vol1d=28, vol7d=28
		header2 := padRightVis("", 26) + dimStyle.Render("VOLUME (1D)") +
			padRightVis("", 16) + dimStyle.Render("VOLUME (7D)")
		lines = append(lines, header2)

		for i := 0; i < 2; i++ {
			code := padRightVis(m.lp.Codes[i], 10)
			vol1d := padLeftVis(m.lp.Vol1d[i], 28)
			vol7d := padLeftVis(m.lp.Vol7d[i], 28)
			row := code + vol1d + padRightVis("", 2) + vol7d
			lines = append(lines, row)
		}
	} else {
		lines = append(lines, dimStyle.Render("Loading pool metrics..."))
	}
	return strings.Join(lines, "\n")
}

func (m model) renderTrades() string {
	rows := []string{boldStyle.Render("TRADES (latest)")}
	rows = append(rows, dimStyle.Render("ELAPSED   PRICE         AMOUNT"))
	limit := 15 // 7 + 7 + 1
	count := 0
	now := time.Now().UTC()
	for i := len(m.trades) - 1; i >= 0 && count < limit; i-- {
		t := m.trades[i]
		isSell := t.BaseIsSeller
		price := alignNum(tradePriceString(t.Price))
		amount := alignNum(formatAmount(t.BaseAmount))
		elapsed := humanElapsedShort(now.Sub(time.Time(t.LedgerCloseTime)))
		line := fmt.Sprintf("%s  %s  %s", padLeftVis(elapsed, 8), price, amount)
		if isSell {
			rows = append(rows, redStyle.Render(line))
		} else {
			rows = append(rows, greenStyle.Render(line))
		}
		count++
	}
	return lipgloss.NewStyle().Render(strings.Join(rows, "\n"))
}

func (m model) midPrice() string {
	if len(m.orderbook.Bids) == 0 || len(m.orderbook.Asks) == 0 {
		return ""
	}
	bestBid, _ := strconv.ParseFloat(m.orderbook.Bids[0].Price, 64)
	bestAsk, _ := strconv.ParseFloat(m.orderbook.Asks[0].Price, 64)
	if bestBid <= 0 || bestAsk <= 0 {
		return ""
	}
	mid := (bestBid + bestAsk) / 2
	return formatPrice(mid)
}

// Commands

func fetchOrderbookCmd(client *horizonclient.Client, base, quote txnbuild.Asset) tea.Cmd {
	return func() tea.Msg {
		if client == nil || base == nil || quote == nil {
			return errMsg(fmt.Errorf("not configured"))
		}
		req := horizonclient.OrderBookRequest{}
		applySellingAsset(&req, base)
		applyBuyingAsset(&req, quote)
		ob, err := client.OrderBook(req)
		if err != nil {
			return errMsg(err)
		}
		return orderbookDataMsg{ob: ob}
	}
}

func fetchTradesCmd(client *horizonclient.Client, base, quote txnbuild.Asset, cursor string, bootstrap bool) tea.Cmd {
	return func() tea.Msg {
		if client == nil || base == nil || quote == nil {
			return errMsg(fmt.Errorf("not configured"))
		}
		req := horizonclient.TradeRequest{}
		applyBaseAsset(&req, base)
		applyCounterAsset(&req, quote)
		if cursor == "" {
			// Bootstrap: get the most recent 50 in descending order
			req.Limit = 50
			req.Order = horizonclient.OrderDesc
		} else {
			req.Cursor = cursor
			req.Order = horizonclient.OrderAsc
			req.Limit = 200
		}
		page, err := client.Trades(req)
		if err != nil {
			return errMsg(err)
		}
		recs := page.Embedded.Records
		if cursor == "" {
			// reverse so newest last
			for i, j := 0, len(recs)-1; i < j; i, j = i+1, j-1 {
				recs[i], recs[j] = recs[j], recs[i]
			}
		}
		return tradesDataMsg{list: recs}
	}
}

// Helpers

func assetTypeEnum(a txnbuild.Asset) horizonclient.AssetType {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		return horizonclient.AssetType("native")
	case txnbuild.CreditAsset:
		if len(v.Code) > 4 {
			return horizonclient.AssetType("credit_alphanum12")
		}
		return horizonclient.AssetType("credit_alphanum4")
	default:
		return ""
	}
}

func applySellingAsset(req *horizonclient.OrderBookRequest, a txnbuild.Asset) {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		req.SellingAssetType = "native"
	case txnbuild.CreditAsset:
		req.SellingAssetType = assetTypeEnum(v)
		req.SellingAssetCode = v.Code
		req.SellingAssetIssuer = v.Issuer
	}
}

func applyBuyingAsset(req *horizonclient.OrderBookRequest, a txnbuild.Asset) {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		req.BuyingAssetType = "native"
	case txnbuild.CreditAsset:
		req.BuyingAssetType = assetTypeEnum(v)
		req.BuyingAssetCode = v.Code
		req.BuyingAssetIssuer = v.Issuer
	}
}

func applyBaseAsset(req *horizonclient.TradeRequest, a txnbuild.Asset) {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		req.BaseAssetType = "native"
	case txnbuild.CreditAsset:
		req.BaseAssetType = assetTypeEnum(v)
		req.BaseAssetCode = v.Code
		req.BaseAssetIssuer = v.Issuer
	}
}

func applyCounterAsset(req *horizonclient.TradeRequest, a txnbuild.Asset) {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		req.CounterAssetType = "native"
	case txnbuild.CreditAsset:
		req.CounterAssetType = assetTypeEnum(v)
		req.CounterAssetCode = v.Code
		req.CounterAssetIssuer = v.Issuer
	}
}

func currentPairIndex(base, quote txnbuild.Asset) int {
	bc := assetShort(base)
	qc := assetShort(quote)
	for i, p := range curatedPairs {
		if p.Base == bc && p.Quote == qc {
			return i
		}
	}
	return 0
}

func parseAsset(s string) (txnbuild.Asset, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "native") || strings.EqualFold(s, "XLM") && !strings.Contains(s, ":") {
		return txnbuild.NativeAsset{}, nil
	}
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("expected CODE:ISSUER or 'native'")
	}
	code := strings.ToUpper(strings.TrimSpace(parts[0]))
	issuer := strings.TrimSpace(parts[1])
	if code == "XLM" && issuer == "" {
		return txnbuild.NativeAsset{}, nil
	}
	if code == "" || issuer == "" {
		return nil, fmt.Errorf("invalid asset spec")
	}
	return txnbuild.CreditAsset{Code: code, Issuer: issuer}, nil
}

func assetShort(a txnbuild.Asset) string {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		return "XLM"
	case txnbuild.CreditAsset:
		return v.Code
	default:
		return "?"
	}
}

func assetString(a txnbuild.Asset) string {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		return "native"
	case txnbuild.CreditAsset:
		return fmt.Sprintf("%s:%s", v.Code, v.Issuer)
	default:
		return ""
	}
}

func formatAmount(s string) string {
	// Ensure at least 2 and up to 7 decimals.
	if s == "" {
		return "0.00"
	}
	if !strings.Contains(s, ".") {
		return s + ".00"
	}
	parts := strings.SplitN(s, ".", 2)
	dec := parts[1]
	if len(dec) < 2 {
		dec = dec + strings.Repeat("0", 2-len(dec))
	}
	if len(dec) > 7 {
		dec = dec[:7]
	}
	return parts[0] + "." + dec
}

func formatPrice(f float64) string {
	// 7 decimal max, min 2
	s := strconv.FormatFloat(f, 'f', 7, 64)
	return trimDecimalsKeepMin2(s)
}

func trimDecimalsKeepMin2(s string) string {
	if !strings.Contains(s, ".") {
		return s + ".00"
	}
	parts := strings.SplitN(s, ".", 2)
	dec := parts[1]
	// trim trailing zeros but keep at least 2
	dec = strings.TrimRight(dec, "0")
	if len(dec) < 2 {
		dec = dec + strings.Repeat("0", 2-len(dec))
	}
	if len(dec) > 7 {
		dec = dec[:7]
	}
	return parts[0] + "." + dec
}

func alignNum(s string) string {
	// fixed width numeric alignment to 12 chars
	s = trimDecimalsKeepMin2(formatAmount(s))
	if len(s) >= 12 {
		return s
	}
	return strings.Repeat(" ", 12-len(s)) + s
}

func tradePriceString(tp hProtocol.TradePrice) string {
	// Convert rational price to decimal string with up to 7 decimals
	if tp.D == 0 {
		return "0.00"
	}
	f := float64(tp.N) / float64(tp.D)
	return formatPrice(f)
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func padRightVis(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func orderbookBlankRow(priceW, amountW, totalW, barW int) string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		padLeftVis("", priceW), padRight("", 2),
		padLeftVis("", amountW), padRight("", 2),
		padLeftVis("", totalW), padRight("", 2), strings.Repeat(" ", barW),
	)
}

func padLeftVis(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return strings.Repeat(" ", n-w) + s
}

func depthBar(width int, ratio float64, color lipgloss.Color) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	n := int(ratio*float64(width) + 0.5)
	if n < 0 {
		n = 0
	}
	if n > width {
		n = width
	}
	sty := lipgloss.NewStyle().Background(color)
	return sty.Render(strings.Repeat(" ", n)) + strings.Repeat(" ", width-n)
}

func formatWithCommas(s string) string {
	// assumes s already normalized for decimals (2-7). Adds space separators to integer part.
	if s == "" {
		return "0.00"
	}
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	parts := strings.SplitN(s, ".", 2)
	intp := parts[0]
	decp := ""
	if len(parts) == 2 {
		decp = parts[1]
	}
	// add spaces to intp
	var out []byte
	for i, c := range []byte(intp) {
		if i != 0 && (len(intp)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, c)
	}
	res := string(out)
	if neg {
		res = "-" + res
	}
	if decp != "" {
		res += "." + decp
	} else {
		res += ".00"
	}
	return res
}

func formatFloatWithCommas(f float64) string {
	s := strconv.FormatFloat(f, 'f', 7, 64)
	return formatWithCommas(trimDecimalsKeepMin2(s))
}

func toFixedDecimals(s string, n int) string {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return trimDecimalsKeepMin2(s)
	}
	return strconv.FormatFloat(f, 'f', n, 64)
}

func alignDecimalFixed(s string, intW, fracW int) string {
	s = toFixedDecimals(s, fracW)
	if !strings.Contains(s, ".") {
		s = s + "." + strings.Repeat("0", fracW)
	}
	parts := strings.SplitN(s, ".", 2)
	intp, frac := parts[0], parts[1]
	if len(frac) < fracW {
		frac = frac + strings.Repeat("0", fracW-len(frac))
	}
	if len(frac) > fracW {
		frac = frac[:fracW]
	}
	if len(intp) < intW {
		intp = strings.Repeat(" ", intW-len(intp)) + intp
	}
	return intp + "." + frac
}

func maxPriceIntWidth(bids []hProtocol.PriceLevel, asks []hProtocol.PriceLevel, maxRows int) int {
	maxW := 1
	// bids best->worse
	for i := 0; i < len(bids) && i < maxRows; i++ {
		s := toFixedDecimals(bids[i].Price, 7)
		parts := strings.SplitN(s, ".", 2)
		if len(parts[0]) > maxW {
			maxW = len(parts[0])
		}
	}
	// asks displayed from worst->best (iterate reverse)
	for c, i := 0, len(asks)-1; i >= 0 && c < maxRows; i, c = i-1, c+1 {
		s := toFixedDecimals(asks[i].Price, 7)
		parts := strings.SplitN(s, ".", 2)
		if len(parts[0]) > maxW {
			maxW = len(parts[0])
		}
	}
	return maxW
}

func spreadPercent(bestBid, bestAsk string) string {
	if bestBid == "" || bestAsk == "" {
		return "-"
	}
	bb, err1 := strconv.ParseFloat(bestBid, 64)
	ba, err2 := strconv.ParseFloat(bestAsk, 64)
	if err1 != nil || err2 != nil || bb <= 0 || ba <= 0 {
		return "-"
	}
	sp := (ba - bb) / ((ba + bb) / 2) * 100
	if sp < 0 {
		sp = 0
	}
	return trimDecimalsKeepMin2(strconv.FormatFloat(sp, 'f', 4, 64)) + "%"
}

// Reusable UI components

func renderHeader() string {
	return asciiAquila
}

func renderSubtitle(title string) string {
	return boldStyle.Render(title)
}

func renderFooter(shortcuts string, isLive bool) string {
	statusText := "LIVE"
	if !isLive {
		statusText = "OFFLINE"
	}

	w := 140 // fixed width
	leftText := shortcuts
	rightText := statusText
	gap := w - lipgloss.Width(leftText) - lipgloss.Width(rightText) - 2
	if gap < 1 {
		gap = 1
	}

	line := leftText + strings.Repeat(" ", gap) + rightText
	return inverseStyle.Render(line)
}

func (m model) bottomLine() string {
	var shortcuts string
	switch m.currentScreen {
	case screenServiceSelection:
		shortcuts = "1: pairs  2: assets  q: quit"
	case screenSelectPair:
		shortcuts = "↑/↓: navigate  enter: select  a: custom pair  esc: back  q: quit"
	case screenPairInfo:
		shortcuts = "b: back  ,/. depth  z: debug  q: quit"
	case screenPairDebug:
		shortcuts = "z: pair info  b: select pair  q: quit"
	case screenSelectAsset:
		shortcuts = "↑/↓: navigate  enter: select  esc: back  q: quit"
	case screenViewExposure:
		shortcuts = "b: back  z: debug  q: quit"
	case screenExposureDebug:
		shortcuts = "z: exposure  q: quit"
	case screenPairInput:
		shortcuts = "enter: apply  tab: switch field  esc: back  q: quit"
	default:
		shortcuts = "q: quit"
	}
	return renderFooter(shortcuts, m.isLive())
}

func bottomBarFixed(shortcuts string, isLive bool) string {
	return renderFooter(shortcuts, isLive)
}

func (m model) isLive() bool {
	ok1 := time.Since(m.lastOrderbookAt) < 3*orderbookInterval
	ok2 := time.Since(m.lastTradesAt) < 3*tradesInterval
	return ok1 && ok2
}

func humanElapsedShort(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	days := int(d.Hours()) / 24
	h := int(d.Hours()) % 24
	return fmt.Sprintf("%dd%02dh", days, h)
}

// Styles

var (
	boldStyle     = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	greenStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	redStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	headerStyle   = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252")).Padding(0, 1)
	statusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).PaddingTop(1)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)
	pairItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	panelStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	inverseStyle  = lipgloss.NewStyle().Reverse(true)
)

func serviceSelectionView(m model) string {
	lines := []string{
		renderHeader(),
		renderSubtitle("Service Selection"),
		"",
		dimStyle.Render("Choose a service:"),
		"",
		selectedStyle.Render("  1. View Asset Pairs"),
		dimStyle.Render("     Monitor order books, trades, and liquidity pools for asset pairs"),
		"",
		selectedStyle.Render("  2. View Single Asset Exposure"),
		dimStyle.Render("     View all liquidity pools containing a specific asset"),
		"",
	}

	if m.err != nil {
		lines = append(lines, "")
		lines = append(lines, errorStyle.Render(m.err.Error()))
	}

	content := strings.Join(lines, "\n")
	contentHeight := lipgloss.Height(content)
	targetHeight := 60
	if m.height > 0 {
		targetHeight = m.height
	}
	paddingLines := targetHeight - contentHeight - 2
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)

	bottom := m.bottomLine()
	return lipgloss.JoinVertical(lipgloss.Left, content, padding, bottom)
}

func pairInputView(m model) string {
	lines := []string{
		renderHeader(),
		renderSubtitle("Type Asset Pair"),
		"",
		dimStyle.Render("Enter assets as 'native' or 'CODE:ISSUER'. Press Tab to switch fields."),
		"",
	}
	if !m.baseInput.Focused() && !m.quoteInput.Focused() {
		m.baseInput.Focus()
	}
	lines = append(lines, m.baseInput.View())
	lines = append(lines, m.quoteInput.View())
	lines = append(lines, "")
	if m.err != nil {
		lines = append(lines, "")
		lines = append(lines, errorStyle.Render(m.err.Error()))
	}

	content := strings.Join(lines, "\n")
	contentHeight := lipgloss.Height(content)
	targetHeight := 60
	if m.height > 0 {
		targetHeight = m.height
	}
	paddingLines := targetHeight - contentHeight - 2
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)

	bottom := m.bottomLine()
	return lipgloss.JoinVertical(lipgloss.Left, content, padding, bottom)
}

func pairSelectorView(m model) string {
	lines := []string{
		renderHeader(),
		renderSubtitle("Select Pair"),
		"",
	}
	for i, p := range curatedPairs {
		label := fmt.Sprintf("%s/%s", p.Base, p.Quote)
		if i == m.pairIndex {
			lines = append(lines, selectedStyle.Render("> "+label))
		} else {
			lines = append(lines, pairItemStyle.Render("  "+label))
		}
	}
	if m.err != nil {
		lines = append(lines, "")
		lines = append(lines, errorStyle.Render(m.err.Error()))
	}

	content := strings.Join(lines, "\n")
	contentHeight := lipgloss.Height(content)
	targetHeight := 60
	if m.height > 0 {
		targetHeight = m.height
	}
	paddingLines := targetHeight - contentHeight - 2
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)

	bottom := m.bottomLine()
	return lipgloss.JoinVertical(lipgloss.Left, content, padding, bottom)
}

func selectAssetView(m model) string {
	lines := []string{
		renderHeader(),
		renderSubtitle("Select Asset"),
		"",
	}

	// Build list of unique assets from curatedAssets
	assetList := []string{"XLM", "USDZ", "ZARZ", "EURZ", "XAUZ", "BTCZ", "USDC"}

	for i, assetCode := range assetList {
		label := assetCode
		if i == m.assetIndex {
			lines = append(lines, selectedStyle.Render("> "+label))
		} else {
			lines = append(lines, pairItemStyle.Render("  "+label))
		}
	}

	if m.err != nil {
		lines = append(lines, "")
		lines = append(lines, errorStyle.Render(m.err.Error()))
	}

	content := strings.Join(lines, "\n")
	contentHeight := lipgloss.Height(content)
	targetHeight := 60
	if m.height > 0 {
		targetHeight = m.height
	}
	paddingLines := targetHeight - contentHeight - 2
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)

	bottom := m.bottomLine()
	return lipgloss.JoinVertical(lipgloss.Left, content, padding, bottom)
}

func viewExposureView(m model) string {
	lines := []string{
		renderHeader(),
		renderSubtitle("View Exposure - " + assetShort(m.selectedAsset)),
		"",
	}

	if len(m.exposurePools) == 0 {
		lines = append(lines, dimStyle.Render("Loading liquidity pools..."))
	} else {
		// Build exposure entries with locked amounts for selected asset
		type exposureEntry struct {
			otherAsset string
			amount     string
			numericAmt float64
		}

		selectedCode := assetShort(m.selectedAsset)
		entries := []exposureEntry{}

		for _, pool := range m.exposurePools {
			// Find which index has the selected asset
			var selectedIdx, otherIdx int
			if strings.EqualFold(pool.Codes[0], selectedCode) {
				selectedIdx = 0
				otherIdx = 1
			} else if strings.EqualFold(pool.Codes[1], selectedCode) {
				selectedIdx = 1
				otherIdx = 0
			} else {
				continue // pool doesn't contain selected asset
			}

			// Parse the locked amount to float for sorting
			amtStr := pool.Locked[selectedIdx]
			// Remove spaces and commas for parsing
			amtClean := strings.ReplaceAll(strings.ReplaceAll(amtStr, " ", ""), ",", "")
			numeric, err := strconv.ParseFloat(amtClean, 64)
			if err != nil {
				numeric = 0
			}

			entries = append(entries, exposureEntry{
				otherAsset: pool.Codes[otherIdx],
				amount:     amtStr,
				numericAmt: numeric,
			})
		}

		// Sort by numeric amount descending
		for i := 0; i < len(entries); i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[j].numericAmt > entries[i].numericAmt {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		// Find max amount for bar scaling
		maxAmt := 0.0
		for _, e := range entries {
			if e.numericAmt > maxAmt {
				maxAmt = e.numericAmt
			}
		}

		// Render compact view
		barWidth := 20 // width for the background bar
		for _, e := range entries {
			// Format pair as "CODE/ZARZ" (4 chars for code + 1 for / + selected asset code)
			pairStr := fmt.Sprintf("%4s/%s", e.otherAsset, selectedCode)

			// Format amount with thousand separators and 2 decimal places
			amtFormatted := formatFloatWithCommas(e.numericAmt)
			// Trim to 2 decimals for compact display
			if idx := strings.Index(amtFormatted, "."); idx >= 0 {
				intPart := amtFormatted[:idx]
				decPart := amtFormatted[idx+1:]
				if len(decPart) > 2 {
					decPart = decPart[:2]
				}
				amtFormatted = intPart + "." + decPart
			}
			// Right-align the amount (16 chars to accommodate large numbers with spaces)
			amtFormatted = padLeftVis(amtFormatted, 16)

			// Calculate bar ratio for background color
			ratio := 0.0
			if maxAmt > 0 {
				ratio = e.numericAmt / maxAmt
			}
			bar := depthBar(barWidth, ratio, lipgloss.Color("24")) // same teal as bid bars

			line := lipgloss.JoinHorizontal(lipgloss.Top, pairStr, "   ", amtFormatted, " ", bar)
			lines = append(lines, line)
		}
	}

	if m.err != nil {
		lines = append(lines, "")
		lines = append(lines, errorStyle.Render(m.err.Error()))
	}

	content := strings.Join(lines, "\n")
	contentHeight := lipgloss.Height(content)
	targetHeight := 60
	if m.height > 0 {
		targetHeight = m.height
	}
	paddingLines := targetHeight - contentHeight - 2
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)

	bottom := m.bottomLine()
	return lipgloss.JoinVertical(lipgloss.Left, content, padding, bottom)
}

func exposureDebugView(m model) string {
	lines := []string{
		renderHeader(),
		renderSubtitle("Exposure Debug"),
		"",
	}

	// Selected asset info
	assetStr := assetString(m.selectedAsset)
	if assetStr == "native" {
		assetStr = "XLM:native"
	}
	lines = append(lines, dimStyle.Render("Selected asset:"), assetStr)
	lines = append(lines, "")

	// Pool count
	lines = append(lines, dimStyle.Render(fmt.Sprintf("Pools found: %d", len(m.exposurePools))))
	lines = append(lines, "")

	// Debug logs
	lines = append(lines, boldStyle.Render("Logs (latest)"))
	logStart := len(m.debugLogs) - 30
	if logStart < 0 {
		logStart = 0
	}
	for i := logStart; i < len(m.debugLogs); i++ {
		lines = append(lines, dimStyle.Render(m.debugLogs[i]))
	}

	content := strings.Join(lines, "\n")
	contentHeight := lipgloss.Height(content)
	targetHeight := 60
	if m.height > 0 {
		targetHeight = m.height
	}
	paddingLines := targetHeight - contentHeight - 2
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)
	bottom := m.bottomLine()
	return lipgloss.JoinVertical(lipgloss.Left, content, padding, bottom)
}

func pairDebugView(m model) string {
	lines := []string{
		renderHeader(),
		renderSubtitle("Pair Debug"),
		"",
	}
	// Pair
	pair := fmt.Sprintf("%s/%s", assetShort(m.base), assetShort(m.quote))
	lines = append(lines, dimStyle.Render("Pair selected:"), pair)
	lines = append(lines, "")
	// Assets full
	baseStr := assetString(m.base)
	if baseStr == "native" {
		baseStr = "XLM:native"
	}
	quoteStr := assetString(m.quote)
	if quoteStr == "native" {
		quoteStr = "XLM:native"
	}
	lines = append(lines, dimStyle.Render("Base asset (code:issuer):"), baseStr)
	lines = append(lines, dimStyle.Render("Counter asset (code:issuer):"), quoteStr)
	lines = append(lines, "")
	// LP ID
	pairKey := fmt.Sprintf("%s-%s", assetShort(m.base), assetShort(m.quote))
	lpID, found := liquidityPoolIDs[pairKey]
	if !found {
		lpID = "(not found)"
	}
	lines = append(lines, dimStyle.Render("LP Pool ID:"), lpID)
	lines = append(lines, "")
	// Debug logs
	lines = append(lines, boldStyle.Render("Logs (latest)"))
	logStart := len(m.debugLogs) - 30
	if logStart < 0 {
		logStart = 0
	}
	for i := logStart; i < len(m.debugLogs); i++ {
		lines = append(lines, dimStyle.Render(m.debugLogs[i]))
	}

	content := strings.Join(lines, "\n")
	contentHeight := lipgloss.Height(content)
	targetHeight := 60
	if m.height > 0 {
		targetHeight = m.height
	}
	paddingLines := targetHeight - contentHeight - 2
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)
	bottom := m.bottomLine()
	return lipgloss.JoinVertical(lipgloss.Left, content, padding, bottom)
}

func horizonURL() string {
	if v := os.Getenv("HORIZON_URL"); v != "" {
		return v
	}
	// Default to public Stellar Horizon mainnet
	return "https://horizon.stellar.org"
}

func newClient() *horizonclient.Client {
	return &horizonclient.Client{HorizonURL: horizonURL()}
}

// ----- Liquidity fetch -----

const defaultPoolID = "7001fca2d71456cda8a061e4733f035fce36423ccf942e92db139a116d7e557b"

func reserveParam(a txnbuild.Asset) string {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		return "native"
	case txnbuild.CreditAsset:
		return fmt.Sprintf("%s:%s", v.Code, v.Issuer)
	default:
		return ""
	}
}

// Liquidity holds display-ready strings

type Liquidity struct {
	Codes    [2]string
	Decimals [2]int
	Locked   [2]string
	Fees1d   [2]string
	Fees7d   [2]string
	Vol1d    [2]string
	Vol7d    [2]string
}

type lpAPIResponse struct {
	Assets []struct {
		Amount string `json:"amount"`
		Asset  string `json:"asset"`
		Name   string `json:"name"`
		Toml   struct {
			Code     string `json:"code"`
			Issuer   string `json:"issuer"`
			Decimals int    `json:"decimals"`
		} `json:"toml_info"`
	} `json:"assets"`
	EarnedFees []struct {
		Asset string          `json:"asset"`
		D1    json.RawMessage `json:"1d"`
		D7    json.RawMessage `json:"7d"`
	} `json:"earned_fees"`
	Volume []struct {
		Asset string          `json:"asset"`
		D1    json.RawMessage `json:"1d"`
		D7    json.RawMessage `json:"7d"`
	} `json:"volume"`
	Updated int64 `json:"updated"`
}

func resolveAndFetchLPCmd(client *horizonclient.Client, base, quote txnbuild.Asset) tea.Cmd {
	return func() tea.Msg {
		// Allow override
		if override := os.Getenv("LP_POOL_ID"); override != "" {
			if data, err := fetchLPByID(override); err == nil {
				return lpDataMsg{data: data}
			} else {
				return lpNoteMsg(fmt.Sprintf("Error loading pool: %v", err))
			}
		}
		if base == nil || quote == nil {
			return lpNoteMsg("No pool: not configured")
		}

		// Look up pool ID from our predefined map
		pairKey := fmt.Sprintf("%s-%s", assetShort(base), assetShort(quote))
		poolID, found := liquidityPoolIDs[pairKey]
		if !found {
			return lpNoteMsg("No pool for " + pairKey)
		}

		data, err := fetchLPByID(poolID)
		if err != nil {
			return lpNoteMsg(fmt.Sprintf("Pool fetch error: %v", err))
		}
		return lpDataMsg{data: data}
	}
}

func fetchLPByID(poolID string) (Liquidity, error) {
	url := "https://api.stellar.expert/explorer/public/liquidity-pool/" + poolID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Liquidity{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<10))
		return Liquidity{}, fmt.Errorf("lp http %d: %s", resp.StatusCode, string(b))
	}
	var api lpAPIResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&api); err != nil {
		return Liquidity{}, err
	}
	data := Liquidity{}
	for i := 0; i < len(api.Assets) && i < 2; i++ {
		code := api.Assets[i].Toml.Code
		if code == "" {
			code = strings.Split(api.Assets[i].Asset, "-")[0]
		}
		data.Codes[i] = code
		data.Decimals[i] = api.Assets[i].Toml.Decimals
		if data.Decimals[i] == 0 {
			data.Decimals[i] = 7 // default to 7 if not specified
		}
		// stellar.expert returns amounts in stroops (always 7 decimals)
		data.Locked[i] = formatLPAmount(api.Assets[i].Amount)
	}
	for _, ef := range api.EarnedFees {
		code := strings.Split(ef.Asset, "-")[0]
		idx := indexOfCode(data.Codes, code)
		if idx >= 0 {
			data.Fees1d[idx] = parseFlexNumberWithDecimals(ef.D1, data.Decimals[idx])
			data.Fees7d[idx] = parseFlexNumberWithDecimals(ef.D7, data.Decimals[idx])
		}
	}
	for _, v := range api.Volume {
		code := strings.Split(v.Asset, "-")[0]
		idx := indexOfCode(data.Codes, code)
		if idx >= 0 {
			data.Vol1d[idx] = parseFlexNumberWithDecimals(v.D1, data.Decimals[idx])
			data.Vol7d[idx] = parseFlexNumberWithDecimals(v.D7, data.Decimals[idx])
		}
	}
	return data, nil
}

func indexOfCode(arr [2]string, code string) int {
	for i := 0; i < len(arr); i++ {
		if strings.EqualFold(arr[i], code) {
			return i
		}
	}
	return -1
}

func parseFlexNumberWithDecimals(raw json.RawMessage, decimals int) string {
	// Try parsing as integer first
	var intVal int64
	if err := json.Unmarshal(raw, &intVal); err == nil {
		return formatLPAmount(strconv.FormatInt(intVal, 10))
	}
	// Try as string
	var strVal string
	if err := json.Unmarshal(raw, &strVal); err == nil {
		return formatLPAmount(strVal)
	}
	// Fallback
	return "0.00"
}

func formatLPAmount(s string) string {
	// stellar.expert API returns values as stroops (integers)
	// Stellar stroops are always 7 decimal places: 1 stroop = 0.0000001 units
	const stellarDecimals = 7
	s = strings.TrimSpace(s)

	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}

	// Pad with zeros if needed to have enough digits
	for len(s) <= stellarDecimals {
		s = "0" + s
	}

	// Split into whole and fractional parts (always 7 decimals for stroops)
	whole := s[:len(s)-stellarDecimals]
	frac := s[len(s)-stellarDecimals:]

	// Add space separators to whole part (every 3 digits)
	var out []byte
	for i, c := range []byte(whole) {
		if i != 0 && (len(whole)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, c)
	}
	whole = string(out)

	// Trim trailing zeros from fractional part but keep at least 2 decimals
	frac = strings.TrimRight(frac, "0")
	if len(frac) < 2 {
		frac = frac + strings.Repeat("0", 2-len(frac))
	}
	// Limit max decimals shown (e.g., 10 max)
	if len(frac) > 10 {
		frac = frac[:10]
		frac = strings.TrimRight(frac, "0")
		if len(frac) < 2 {
			frac = frac + strings.Repeat("0", 2-len(frac))
		}
	}

	res := whole + "." + frac
	if neg {
		res = "-" + res
	}
	return res
}

func fixed7FromIntString(s string) string {
	s = strings.TrimSpace(s)
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	for len(s) < 8 { // ensure at least 8 to safely split; pad left
		s = "0" + s
	}
	whole := s[:len(s)-7]
	frac := s[len(s)-7:]
	res := formatWithCommas(whole + "." + frac)
	if neg {
		res = "-" + res
	}
	return res
}

// Ensure a string has a decimal point and exactly 7 decimals (keeps commas)
func ensure7Decimals(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "."); idx >= 0 {
		intp := s[:idx]
		frac := s[idx+1:]
		if len(frac) < 7 {
			frac = frac + strings.Repeat("0", 7-len(frac))
		}
		if len(frac) > 7 {
			frac = frac[:7]
		}
		return intp + "." + frac
	}
	return s + ".0000000"
}

// Align decimal point across two rows with code prefix (fixed width)
func alignDecimalsWithCode(codes [2]string, vals [2]string, fieldWidth int) [2]string {
	codeW := max(lipgloss.Width(codes[0]), lipgloss.Width(codes[1]))
	valW := fieldWidth - codeW - 2
	aligned := [2]string{}
	nums := []string{ensure7Decimals(vals[0]), ensure7Decimals(vals[1])}
	// compute max integer width (including commas and sign)
	maxInt := 0
	for _, s := range nums {
		idx := strings.Index(s, ".")
		if idx < 0 {
			idx = len(s)
		}
		if idx > maxInt {
			maxInt = idx
		}
	}
	for i := 0; i < 2; i++ {
		idx := strings.Index(nums[i], ".")
		if idx < 0 {
			idx = len(nums[i])
		}
		leftPad := maxInt - idx
		num := strings.Repeat(" ", leftPad) + nums[i]
		num = padRightVis(num, valW)
		aligned[i] = padRightVis(codes[i], codeW) + "  " + num
	}
	return aligned
}

// Align decimal point across two rows without code prefix
func alignDecimalsNoCode(vals [2]string, fieldWidth int) [2]string {
	res := [2]string{}
	nums := []string{ensure7Decimals(vals[0]), ensure7Decimals(vals[1])}
	maxInt := 0
	for _, s := range nums {
		idx := strings.Index(s, ".")
		if idx < 0 {
			idx = len(s)
		}
		if idx > maxInt {
			maxInt = idx
		}
	}
	for i := 0; i < 2; i++ {
		idx := strings.Index(nums[i], ".")
		if idx < 0 {
			idx = len(nums[i])
		}
		leftPad := maxInt - idx
		s := strings.Repeat(" ", leftPad) + nums[i]
		res[i] = padRightVis(s, fieldWidth)
	}
	return res
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maskSecret(s string) string {
	if s == "" {
		return "(empty)"
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

// Custom log writer to capture logs in memory
var debugLogBuffer []string
var debugLogMutex sync.Mutex

type debugLogWriter struct{}

func (w debugLogWriter) Write(p []byte) (n int, err error) {
	debugLogMutex.Lock()
	defer debugLogMutex.Unlock()
	line := strings.TrimSpace(string(p))
	if line != "" {
		debugLogBuffer = append(debugLogBuffer, line)
		if len(debugLogBuffer) > 100 {
			debugLogBuffer = debugLogBuffer[len(debugLogBuffer)-100:]
		}
	}
	return len(p), nil
}

func setupDebugLogger() {
	// Only write to debug buffer, not to stderr to keep TUI clean
	log.SetOutput(debugLogWriter{})
}

func getDebugLogs() []string {
	debugLogMutex.Lock()
	defer debugLogMutex.Unlock()
	result := make([]string, len(debugLogBuffer))
	copy(result, debugLogBuffer)
	return result
}

// fetchExposureCmd fetches all liquidity pools containing the specified asset
func fetchExposureCmd(client *horizonclient.Client, asset txnbuild.Asset) tea.Cmd {
	return func() tea.Msg {
		if asset == nil {
			return errMsg(fmt.Errorf("no asset selected"))
		}

		// Find all relevant pool IDs for this asset
		assetCode := assetShort(asset)
		var poolIDs []string
		var pairKeys []string

		// Search through our liquidityPoolIDs map for pairs containing this asset
		for pairKey, poolID := range liquidityPoolIDs {
			if strings.Contains(pairKey, assetCode) {
				// Check if we already have this pool ID
				found := false
				for _, existingID := range poolIDs {
					if existingID == poolID {
						found = true
						break
					}
				}
				if !found {
					poolIDs = append(poolIDs, poolID)
					pairKeys = append(pairKeys, pairKey)
				}
			}
		}

		if len(poolIDs) == 0 {
			return exposureDataMsg{pools: []Liquidity{}}
		}

		// Fetch all pools
		var pools []Liquidity
		for _, poolID := range poolIDs {
			data, err := fetchLPByID(poolID)
			if err != nil {
				// Log error but continue with other pools
				log.Printf("Failed to fetch pool %s: %v", poolID, err)
				continue
			}
			pools = append(pools, data)
		}

		return exposureDataMsg{pools: pools}
	}
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	client := newClient()

	// optional defaults via env
	var base, quote txnbuild.Asset
	if b := os.Getenv("BASE_ASSET"); b != "" {
		var err error
		base, err = parseAsset(b)
		if err != nil {
			log.Printf("BASE_ASSET invalid: %v", err)
		}
	}
	if q := os.Getenv("QUOTE_ASSET"); q != "" {
		var err error
		quote, err = parseAsset(q)
		if err != nil {
			log.Printf("QUOTE_ASSET invalid: %v", err)
		}
	}

	p := tea.NewProgram(initialModel(client, base, quote), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// Clear terminal on exit
	fmt.Print("\033[2J\033[H")
}
