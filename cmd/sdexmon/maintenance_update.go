package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sdexmon/sdexmon/internal/config"
	"github.com/sdexmon/sdexmon/internal/models"
	"github.com/sdexmon/sdexmon/internal/stellar"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/txnbuild"
)

func initMaintenanceState() models.MaintenanceState {
	return models.MaintenanceState{
		Screen:      models.MaintenanceMenu,
		QueryInputA: newQueryInput(models.SearchByDomain),
		QueryInputB: newQueryInput(models.SearchByDomain),
	}
}

// newQueryInput builds the text field for a search mode, so the prompt always
// matches what the user is actually expected to type.
func newQueryInput(mode models.AssetSearchMode) textinput.Model {
	input := textinput.New()
	input.CharLimit = 100
	if mode == models.SearchByCode {
		input.Prompt = "Asset > "
		input.Placeholder = "e.g., USDC, ZARZ, Zeam"
		return input
	}
	input.Prompt = "Domain > "
	input.Placeholder = "e.g., zeam.money"
	return input
}

func handleMaintenanceUpdate(m model, msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.maintenanceState.Screen {
		case models.MaintenanceMenu:
			return handleMaintenanceMenuKeys(m, msg)
		case models.AssetASourceSelect:
			return handleSourceSelectKeys(m, msg, models.AssetLegA)
		case models.AssetBSourceSelect:
			return handleSourceSelectKeys(m, msg, models.AssetLegB)
		case models.AssetAQueryInput:
			return handleQueryInputKeys(m, msg, models.AssetLegA)
		case models.AssetBQueryInput:
			return handleQueryInputKeys(m, msg, models.AssetLegB)
		case models.AssetASelection:
			return handleAssetSelectionKeys(m, msg, models.AssetLegA)
		case models.AssetBSelection:
			return handleAssetSelectionKeys(m, msg, models.AssetLegB)
		case models.PairConfirmation:
			return handleConfirmationKeys(m, msg)
		case models.PairRemoveSelection:
			return handlePairRemoveSelectionKeys(m, msg)
		case models.PairRemoveConfirmation:
			return handlePairRemoveConfirmationKeys(m, msg)
		case models.PairList:
			return handlePairListKeys(m, msg)
		}

	case models.AssetSearchResultsMsg:
		return applySearchResults(m, msg), nil

	case models.ConfirmationDataMsg:
		m.maintenanceState.ConfirmationData = msg.Data
		m.maintenanceState.LoadingMessage = ""
		return m, nil

	case models.MaintenanceErrMsg:
		m.maintenanceState.ErrorMessage = msg.Err.Error()
		m.maintenanceState.LoadingMessage = ""
		return m, nil
	}

	return m, nil
}

// applySearchResults files results under the leg that requested them. The
// screen only advances when the user is still waiting on that same search, so
// late results cannot yank them out of somewhere else.
func applySearchResults(m model, msg models.AssetSearchResultsMsg) model {
	m.maintenanceState.LoadingMessage = ""

	if msg.Leg == models.AssetLegA {
		m.maintenanceState.SearchResultsA = msg.Assets
		m.maintenanceState.SearchSourceA = msg.Source
		m.maintenanceState.AssetCursorA = 0
		if m.maintenanceState.Screen == models.AssetASourceSelect || m.maintenanceState.Screen == models.AssetAQueryInput {
			m.maintenanceState.Screen = models.AssetASelection
		}
		return m
	}

	m.maintenanceState.SearchResultsB = msg.Assets
	m.maintenanceState.SearchSourceB = msg.Source
	m.maintenanceState.AssetCursorB = 0
	if m.maintenanceState.Screen == models.AssetBSourceSelect || m.maintenanceState.Screen == models.AssetBQueryInput {
		m.maintenanceState.Screen = models.AssetBSelection
	}
	return m
}

func handleMaintenanceMenuKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentScreen = screenLanding
		m.maintenanceState = initMaintenanceState()
		return m, nil
	case "1":
		// Start add asset pair flow
		m.maintenanceState.Screen = models.AssetASourceSelect
		m.maintenanceState.SourceCursor = 0
		m.maintenanceState.ErrorMessage = ""
		m.maintenanceState.StatusMessage = ""
		return m, nil
	case "2":
		// Start remove asset pair flow
		m.maintenanceState.Screen = models.PairRemoveSelection
		m.maintenanceState.PairCursor = 0
		m.maintenanceState.ErrorMessage = ""
		m.maintenanceState.StatusMessage = ""
		return m, nil
	case "3":
		// Show the configured pairs read-only
		m.maintenanceState.Screen = models.PairList
		m.maintenanceState.PairCursor = 0
		m.maintenanceState.ErrorMessage = ""
		m.maintenanceState.StatusMessage = ""
		return m, nil
	}
	return m, nil
}

// handleSourceSelectKeys drives the "how do you want to find this asset?" menu
// that precedes every asset lookup.
func handleSourceSelectKeys(m model, msg tea.KeyMsg, leg models.AssetLeg) (model, tea.Cmd) {
	switch key := msg.String(); key {
	case "esc":
		m.maintenanceState.ErrorMessage = ""
		m.maintenanceState.LoadingMessage = ""
		if leg == models.AssetLegA {
			m.maintenanceState.Screen = models.MaintenanceMenu
		} else {
			m.maintenanceState.Screen = models.AssetASelection
		}
		return m, nil
	case "up", "k":
		if m.maintenanceState.SourceCursor > 0 {
			m.maintenanceState.SourceCursor--
		}
		return m, nil
	case "down", "j":
		if m.maintenanceState.SourceCursor < len(models.AllAssetSearchModes)-1 {
			m.maintenanceState.SourceCursor++
		}
		return m, nil
	case "1", "2", "3":
		index := int(key[0] - '1')
		if index >= len(models.AllAssetSearchModes) {
			return m, nil
		}
		m.maintenanceState.SourceCursor = index
		return startSearchMode(m, leg, models.AllAssetSearchModes[index])
	case "enter":
		index := m.maintenanceState.SourceCursor
		if index < 0 || index >= len(models.AllAssetSearchModes) {
			return m, nil
		}
		return startSearchMode(m, leg, models.AllAssetSearchModes[index])
	}
	return m, nil
}

// startSearchMode either opens the query field or, for query-less sources such
// as the top 50 list, fires the lookup straight away.
func startSearchMode(m model, leg models.AssetLeg, mode models.AssetSearchMode) (model, tea.Cmd) {
	m.maintenanceState.ErrorMessage = ""

	if leg == models.AssetLegA {
		m.maintenanceState.SearchModeA = mode
	} else {
		m.maintenanceState.SearchModeB = mode
	}

	if !mode.NeedsQuery() {
		m.maintenanceState.LoadingMessage = loadingMessageFor(mode, "")
		return m, searchAssetsCmd(leg, mode, "")
	}

	input := newQueryInput(mode)
	input.Focus()
	if leg == models.AssetLegA {
		m.maintenanceState.QueryInputA = input
		m.maintenanceState.Screen = models.AssetAQueryInput
	} else {
		m.maintenanceState.QueryInputB = input
		m.maintenanceState.Screen = models.AssetBQueryInput
	}
	return m, nil
}

func handleQueryInputKeys(m model, msg tea.KeyMsg, leg models.AssetLeg) (model, tea.Cmd) {
	mode := m.maintenanceState.SearchModeA
	if leg == models.AssetLegB {
		mode = m.maintenanceState.SearchModeB
	}

	switch msg.String() {
	case "esc":
		m.maintenanceState.ErrorMessage = ""
		m.maintenanceState.LoadingMessage = ""
		if leg == models.AssetLegA {
			m.maintenanceState.Screen = models.AssetASourceSelect
		} else {
			m.maintenanceState.Screen = models.AssetBSourceSelect
		}
		return m, nil
	case "enter":
		query := m.maintenanceState.QueryInputA.Value()
		if leg == models.AssetLegB {
			query = m.maintenanceState.QueryInputB.Value()
		}
		if query == "" {
			m.maintenanceState.ErrorMessage = "Enter a value to search for"
			return m, nil
		}
		m.maintenanceState.LoadingMessage = loadingMessageFor(mode, query)
		m.maintenanceState.ErrorMessage = ""
		return m, searchAssetsCmd(leg, mode, query)
	}

	var cmd tea.Cmd
	if leg == models.AssetLegA {
		m.maintenanceState.QueryInputA, cmd = m.maintenanceState.QueryInputA.Update(msg)
	} else {
		m.maintenanceState.QueryInputB, cmd = m.maintenanceState.QueryInputB.Update(msg)
	}
	return m, cmd
}

func loadingMessageFor(mode models.AssetSearchMode, query string) string {
	switch mode {
	case models.SearchByCode:
		return fmt.Sprintf("Searching stellar.expert for %q...", query)
	case models.SearchTop50:
		return "Loading stellar.expert Top 50..."
	default:
		return "Reading " + stellar.TomlURL(query) + "..."
	}
}

func handleAssetSelectionKeys(m model, msg tea.KeyMsg, leg models.AssetLeg) (model, tea.Cmd) {
	results := m.maintenanceState.SearchResultsA
	cursor := m.maintenanceState.AssetCursorA
	if leg == models.AssetLegB {
		results = m.maintenanceState.SearchResultsB
		cursor = m.maintenanceState.AssetCursorB
	}

	setCursor := func(m model, value int) model {
		if leg == models.AssetLegA {
			m.maintenanceState.AssetCursorA = value
		} else {
			m.maintenanceState.AssetCursorB = value
		}
		return m
	}

	switch msg.String() {
	case "esc":
		m.maintenanceState.ErrorMessage = ""
		if leg == models.AssetLegA {
			m.maintenanceState.Screen = models.AssetASourceSelect
		} else {
			m.maintenanceState.Screen = models.AssetBSourceSelect
		}
		return m, nil
	case "up", "k":
		if cursor > 0 {
			m = setCursor(m, cursor-1)
		}
		return m, nil
	case "down", "j":
		if cursor < len(results)-1 {
			m = setCursor(m, cursor+1)
		}
		return m, nil
	case "enter":
		if cursor < 0 || cursor >= len(results) {
			return m, nil
		}
		selected := results[cursor]

		if leg == models.AssetLegA {
			m.maintenanceState.SelectedAssetA = &selected
			m.maintenanceState.Screen = models.AssetBSourceSelect
			m.maintenanceState.SourceCursor = 0
			m.maintenanceState.ErrorMessage = ""
			return m, nil
		}

		if m.maintenanceState.SelectedAssetA == nil {
			m.maintenanceState.ErrorMessage = "Asset A is no longer selected, start again"
			return m, nil
		}
		if m.maintenanceState.SelectedAssetA.Code == selected.Code &&
			m.maintenanceState.SelectedAssetA.Issuer == selected.Issuer {
			m.maintenanceState.ErrorMessage = "A pair needs two different assets"
			return m, nil
		}
		m.maintenanceState.SelectedAssetB = &selected

		assetA := txnbuild.CreditAsset{
			Code:   m.maintenanceState.SelectedAssetA.Code,
			Issuer: m.maintenanceState.SelectedAssetA.Issuer,
		}
		assetB := txnbuild.CreditAsset{
			Code:   selected.Code,
			Issuer: selected.Issuer,
		}

		m.maintenanceState.Screen = models.PairConfirmation
		m.maintenanceState.LoadingMessage = "Fetching market data..."
		m.maintenanceState.ErrorMessage = ""
		return m, fetchConfirmationDataCmd(m.client, assetA, assetB)
	}
	return m, nil
}

func handlePairRemoveSelectionKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.maintenanceState.Screen = models.MaintenanceMenu
		m.maintenanceState.ErrorMessage = ""
		return m, nil
	case "up", "k":
		if m.maintenanceState.PairCursor > 0 {
			m.maintenanceState.PairCursor--
		}
		return m, nil
	case "down", "j":
		if m.maintenanceState.PairCursor < len(configuredPairs)-1 {
			m.maintenanceState.PairCursor++
		}
		return m, nil
	case "enter":
		if _, ok := maintenancePairAt(m.maintenanceState.PairCursor); !ok {
			return m, nil
		}
		m.maintenanceState.Screen = models.PairRemoveConfirmation
		m.maintenanceState.ErrorMessage = ""
		return m, nil
	}
	return m, nil
}

func handlePairRemoveConfirmationKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.maintenanceState.Screen = models.PairRemoveSelection
		m.maintenanceState.ErrorMessage = ""
		return m, nil
	case "enter", "y":
		pair, ok := maintenancePairAt(m.maintenanceState.PairCursor)
		if !ok {
			m.maintenanceState.Screen = models.PairRemoveSelection
			return m, nil
		}
		base, quote, ok := assetsForPair(pair)
		if !ok {
			m.maintenanceState.ErrorMessage = "Pair has unresolved assets and cannot be removed"
			return m, nil
		}
		if err := config.RemoveCustomPair(base, quote); err != nil {
			m.maintenanceState.ErrorMessage = fmt.Sprintf("Failed to remove: %v", err)
			return m, nil
		}

		label := pairLabel(pair)
		m = reloadPairsAfterMaintenance(m)
		m.maintenanceState.Screen = models.MaintenanceMenu
		m.maintenanceState.PairCursor = 0
		m.maintenanceState.StatusMessage = fmt.Sprintf("Removed pair %s", label)
		m.status = fmt.Sprintf("Removed pair %s", label)
		return m, nil
	}
	return m, nil
}

func handlePairListKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.maintenanceState.Screen = models.MaintenanceMenu
		return m, nil
	case "up", "k":
		if m.maintenanceState.PairCursor > 0 {
			m.maintenanceState.PairCursor--
		}
		return m, nil
	case "down", "j":
		if m.maintenanceState.PairCursor < len(configuredPairs)-1 {
			m.maintenanceState.PairCursor++
		}
		return m, nil
	}
	return m, nil
}

// maintenancePairAt returns the configured pair at index, guarding against a
// cursor left over from a longer list.
func maintenancePairAt(index int) (pairOption, bool) {
	if index < 0 || index >= len(configuredPairs) {
		return pairOption{}, false
	}
	return configuredPairs[index], true
}

// reloadPairsAfterMaintenance re-reads the config file so the pair selector
// reflects the change immediately, and keeps dependent indexes in range.
func reloadPairsAfterMaintenance(m model) model {
	if err := loadConfiguration(); err != nil {
		m.maintenanceState.ErrorMessage = fmt.Sprintf("Saved, but reload failed: %v", err)
	}
	m.filteredPairs = configuredPairs
	if m.pairIndex >= len(configuredPairs) {
		m.pairIndex = 0
	}
	if m.maintenanceState.PairCursor >= len(configuredPairs) {
		m.maintenanceState.PairCursor = 0
	}
	return m
}

func handleConfirmationKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.maintenanceState.Screen = models.AssetBSelection
		m.maintenanceState.ConfirmationData = nil
		m.maintenanceState.ErrorMessage = ""
		return m, nil
	case "enter":
		if m.maintenanceState.ConfirmationData == nil {
			return m, nil
		}

		// Save the pair to config
		cd := m.maintenanceState.ConfirmationData
		err := config.AddCustomPair(cd.AssetA, cd.AssetB)
		if err != nil {
			m.maintenanceState.ErrorMessage = fmt.Sprintf("Failed to save: %v", err)
			return m, nil
		}

		// Success! Reload config and return to the menu so several pairs can be
		// maintained in one visit.
		label := fmt.Sprintf("%s/%s", cd.AssetA.GetCode(), cd.AssetB.GetCode())
		m.maintenanceState = initMaintenanceState()
		m = reloadPairsAfterMaintenance(m)
		m.maintenanceState.StatusMessage = fmt.Sprintf("Added pair %s", label)
		m.status = fmt.Sprintf("Added pair %s", label)
		return m, nil
	}
	return m, nil
}

// Commands

// searchAssetsCmd runs the lookup for the chosen source off the UI thread and
// reports where the results came from, so the selection screen can say how far
// they can be trusted.
func searchAssetsCmd(leg models.AssetLeg, mode models.AssetSearchMode, query string) tea.Cmd {
	return func() tea.Msg {
		var (
			assets []models.AssetSearchResult
			source string
			err    error
		)

		switch mode {
		case models.SearchByCode:
			assets, err = stellar.SearchAssetsByCode(query)
			source = fmt.Sprintf("stellar.expert search for %q - unverified, check each domain", query)
		case models.SearchTop50:
			assets, err = stellar.TopAssets()
			source = "stellar.expert Top 50 - ranked by network activity, not endorsed"
		default:
			assets, err = stellar.ResolveDomainAssets(query)
			source = stellar.TomlURL(query) + " (SEP-1 verified)"
			if len(assets) > 0 && !assets[0].Verified {
				source = fmt.Sprintf("stellar.expert home domain %s - no stellar.toml found",
					stellar.NormalizeDomain(query))
			}
		}

		if err != nil {
			return models.MaintenanceErrMsg{Err: err}
		}
		return models.AssetSearchResultsMsg{Leg: leg, Assets: assets, Source: source}
	}
}

func fetchConfirmationDataCmd(client *horizonclient.Client, assetA, assetB txnbuild.Asset) tea.Cmd {
	return func() tea.Msg {
		data, err := stellar.FetchPairConfirmationData(client, assetA, assetB, liquidityPoolIDs)
		if err != nil {
			return models.MaintenanceErrMsg{Err: err}
		}
		return models.ConfirmationDataMsg{Data: data}
	}
}
