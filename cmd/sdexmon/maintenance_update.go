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
	domainA := textinput.New()
	domainA.Placeholder = "e.g., zeam.money"
	domainA.Prompt = "Domain > "
	domainA.CharLimit = 100

	domainB := textinput.New()
	domainB.Placeholder = "e.g., zeam.money"
	domainB.Prompt = "Domain > "
	domainB.CharLimit = 100

	return models.MaintenanceState{
		Screen:       models.MaintenanceMenu,
		DomainInputA: domainA,
		DomainInputB: domainB,
	}
}

func handleMaintenanceUpdate(m model, msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.maintenanceState.Screen {
		case models.MaintenanceMenu:
			return handleMaintenanceMenuKeys(m, msg)
		case models.AssetADomainInput:
			return handleAssetADomainInputKeys(m, msg)
		case models.AssetASelection:
			return handleAssetASelectionKeys(m, msg)
		case models.AssetBDomainInput:
			return handleAssetBDomainInputKeys(m, msg)
		case models.AssetBSelection:
			return handleAssetBSelectionKeys(m, msg)
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
		// Received asset search results
		if m.maintenanceState.Screen == models.AssetADomainInput {
			m.maintenanceState.SearchResultsA = msg.Assets
			m.maintenanceState.AssetCursorA = 0
			m.maintenanceState.Screen = models.AssetASelection
			m.maintenanceState.LoadingMessage = ""
		} else if m.maintenanceState.Screen == models.AssetBDomainInput {
			m.maintenanceState.SearchResultsB = msg.Assets
			m.maintenanceState.AssetCursorB = 0
			m.maintenanceState.Screen = models.AssetBSelection
			m.maintenanceState.LoadingMessage = ""
		}
		return m, nil

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

func handleMaintenanceMenuKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentScreen = screenLanding
		m.maintenanceState = initMaintenanceState()
		return m, nil
	case "1":
		// Start add asset pair flow
		m.maintenanceState.Screen = models.AssetADomainInput
		m.maintenanceState.DomainInputA.Focus()
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

func handleAssetADomainInputKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.maintenanceState = initMaintenanceState()
		return m, nil
	case "enter":
		domain := m.maintenanceState.DomainInputA.Value()
		if domain == "" {
			m.maintenanceState.ErrorMessage = "Domain cannot be empty"
			return m, nil
		}
		m.maintenanceState.LoadingMessage = "Searching stellar.expert..."
		m.maintenanceState.ErrorMessage = ""
		return m, searchAssetsCmd(domain)
	}

	// Update text input
	var cmd tea.Cmd
	m.maintenanceState.DomainInputA, cmd = m.maintenanceState.DomainInputA.Update(msg)
	return m, cmd
}

func handleAssetASelectionKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.maintenanceState.Screen = models.AssetADomainInput
		m.maintenanceState.ErrorMessage = ""
		return m, nil
	case "up", "k":
		if m.maintenanceState.AssetCursorA > 0 {
			m.maintenanceState.AssetCursorA--
		}
		return m, nil
	case "down", "j":
		if m.maintenanceState.AssetCursorA < len(m.maintenanceState.SearchResultsA)-1 {
			m.maintenanceState.AssetCursorA++
		}
		return m, nil
	case "enter":
		if len(m.maintenanceState.SearchResultsA) == 0 {
			return m, nil
		}
		selected := m.maintenanceState.SearchResultsA[m.maintenanceState.AssetCursorA]
		m.maintenanceState.SelectedAssetA = &selected
		m.maintenanceState.Screen = models.AssetBDomainInput
		m.maintenanceState.DomainInputB.Focus()
		m.maintenanceState.ErrorMessage = ""
		return m, nil
	}
	return m, nil
}

func handleAssetBDomainInputKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.maintenanceState.Screen = models.AssetASelection
		m.maintenanceState.ErrorMessage = ""
		return m, nil
	case "enter":
		domain := m.maintenanceState.DomainInputB.Value()
		if domain == "" {
			m.maintenanceState.ErrorMessage = "Domain cannot be empty"
			return m, nil
		}
		m.maintenanceState.LoadingMessage = "Searching stellar.expert..."
		m.maintenanceState.ErrorMessage = ""
		return m, searchAssetsCmd(domain)
	}

	// Update text input
	var cmd tea.Cmd
	m.maintenanceState.DomainInputB, cmd = m.maintenanceState.DomainInputB.Update(msg)
	return m, cmd
}

func handleAssetBSelectionKeys(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.maintenanceState.Screen = models.AssetBDomainInput
		m.maintenanceState.ErrorMessage = ""
		return m, nil
	case "up", "k":
		if m.maintenanceState.AssetCursorB > 0 {
			m.maintenanceState.AssetCursorB--
		}
		return m, nil
	case "down", "j":
		if m.maintenanceState.AssetCursorB < len(m.maintenanceState.SearchResultsB)-1 {
			m.maintenanceState.AssetCursorB++
		}
		return m, nil
	case "enter":
		if len(m.maintenanceState.SearchResultsB) == 0 {
			return m, nil
		}
		selected := m.maintenanceState.SearchResultsB[m.maintenanceState.AssetCursorB]
		m.maintenanceState.SelectedAssetB = &selected

		// Fetch confirmation data
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

func searchAssetsCmd(domain string) tea.Cmd {
	return func() tea.Msg {
		assets, err := stellar.SearchAssetsByDomain(domain)
		if err != nil {
			return models.MaintenanceErrMsg{Err: err}
		}
		return models.AssetSearchResultsMsg{Assets: assets}
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
