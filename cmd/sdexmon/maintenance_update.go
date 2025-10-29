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
	case "esc", "q":
		m.currentScreen = screenLanding
		m.maintenanceState = initMaintenanceState()
		return m, nil
	case "1":
		// Start add asset pair flow
		m.maintenanceState.Screen = models.AssetADomainInput
		m.maintenanceState.DomainInputA.Focus()
		m.maintenanceState.ErrorMessage = ""
		return m, nil
	}
	return m, nil
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

		// Success! Reload config and return to landing
		m.maintenanceState = initMaintenanceState()
		m.currentScreen = screenLanding
		m.status = fmt.Sprintf("Added pair %s/%s", cd.AssetA.GetCode(), cd.AssetB.GetCode())

		// Reload configured pairs to include the new pair
		return m, reloadConfigCmd()
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

func reloadConfigCmd() tea.Cmd {
	return func() tea.Msg {
		// Reload config to pick up new pairs
		// For now, this is a no-op since we need to restart
		// In the future, we could dynamically reload
		return nil
	}
}
