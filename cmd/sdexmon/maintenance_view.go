package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sdexmon/sdexmon/internal/models"
)

func maintenanceView(m model) string {
	switch m.maintenanceState.Screen {
	case models.MaintenanceMenu:
		return maintenanceMenuView(m)
	case models.AssetADomainInput, models.AssetBDomainInput:
		return maintenanceDomainInputView(m)
	case models.AssetASelection, models.AssetBSelection:
		return maintenanceAssetSelectionView(m)
	case models.PairConfirmation:
		return maintenanceConfirmationView(m)
	default:
		return maintenanceMenuView(m)
	}
}

func maintenanceMenuView(m model) string {
	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Maintenance"),
		"",
		boldStyle.Render("MAINTENANCE MODE"),
		"",
		"1. Add Asset Pair",
		"",
		dimStyle.Render("Coming soon:"),
		dimStyle.Render("  2. Remove Asset Pair"),
		dimStyle.Render("  3. View Custom Pairs"),
		"",
	}

	if m.maintenanceState.ErrorMessage != "" {
		lines = append(lines, "", errorStyle.Render(m.maintenanceState.ErrorMessage))
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

func maintenanceDomainInputView(m model) string {
	isAssetA := m.maintenanceState.Screen == models.AssetADomainInput
	assetLabel := "Asset B"
	if isAssetA {
		assetLabel = "Asset A"
	}

	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Add Asset Pair"),
		"",
		boldStyle.Render("Enter " + assetLabel + " Domain"),
		"",
		dimStyle.Render("Enter the domain to search for assets (e.g., zeam.money)"),
		"",
	}

	// Show the appropriate input field
	if isAssetA {
		lines = append(lines, m.maintenanceState.DomainInputA.View())
	} else {
		lines = append(lines, m.maintenanceState.DomainInputB.View())
	}

	lines = append(lines, "")

	if m.maintenanceState.ErrorMessage != "" {
		lines = append(lines, "", errorStyle.Render(m.maintenanceState.ErrorMessage))
	}

	if m.maintenanceState.LoadingMessage != "" {
		lines = append(lines, "", dimStyle.Render(m.maintenanceState.LoadingMessage))
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

func maintenanceAssetSelectionView(m model) string {
	isAssetA := m.maintenanceState.Screen == models.AssetASelection
	assetLabel := "Asset B"
	results := m.maintenanceState.SearchResultsB
	cursor := m.maintenanceState.AssetCursorB
	if isAssetA {
		assetLabel = "Asset A"
		results = m.maintenanceState.SearchResultsA
		cursor = m.maintenanceState.AssetCursorA
	}

	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Add Asset Pair"),
		"",
		boldStyle.Render("Select " + assetLabel),
		"",
	}

	if len(results) == 0 {
		lines = append(lines, dimStyle.Render("No assets found"))
	} else {
		// Show window of results
		windowSize := 15
		start := cursor - windowSize/2
		if start < 0 {
			start = 0
		}
		end := start + windowSize
		if end > len(results) {
			end = len(results)
		}

		for i := start; i < end; i++ {
			asset := results[i]
			label := fmt.Sprintf("%s (%s...)", asset.Code, asset.Issuer[:8])
			if asset.Name != "" {
				label = fmt.Sprintf("%s - %s (%s...)", asset.Code, asset.Name, asset.Issuer[:8])
			}

			if i == cursor {
				lines = append(lines, selectedStyle.Render("> "+label))
			} else {
				lines = append(lines, pairItemStyle.Render("  "+label))
			}
		}
	}

	lines = append(lines, "")

	if m.maintenanceState.ErrorMessage != "" {
		lines = append(lines, "", errorStyle.Render(m.maintenanceState.ErrorMessage))
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

func maintenanceConfirmationView(m model) string {
	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Confirm Asset Pair"),
		"",
		boldStyle.Render("CONFIRMATION"),
		"",
	}

	if m.maintenanceState.ConfirmationData == nil {
		lines = append(lines, dimStyle.Render("Loading market data..."))
	} else {
		cd := m.maintenanceState.ConfirmationData
		lines = append(lines,
			fmt.Sprintf("Pair: %s / %s", cd.AssetA.GetCode(), cd.AssetB.GetCode()),
			"",
			fmt.Sprintf("Current best bid:       %s", formatPriceValue(cd.BestBid)),
			fmt.Sprintf("Current best ask:       %s", formatPriceValue(cd.BestAsk)),
			fmt.Sprintf("LP Locked %s:      %s", cd.AssetA.GetCode(), formatAmountValue(cd.LPLockedA)),
			fmt.Sprintf("LP Locked %s:      %s", cd.AssetB.GetCode(), formatAmountValue(cd.LPLockedB)),
			"",
			greenStyle.Render("Press ENTER to confirm and add pair"),
			dimStyle.Render("Press ESC to cancel"),
		)
	}

	lines = append(lines, "")

	if m.maintenanceState.ErrorMessage != "" {
		lines = append(lines, "", errorStyle.Render(m.maintenanceState.ErrorMessage))
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

func formatPriceValue(s string) string {
	if s == "N/A" || s == "" {
		return dimStyle.Render("N/A")
	}
	return s
}

func formatAmountValue(s string) string {
	if s == "N/A" || s == "" {
		return dimStyle.Render("N/A")
	}
	return formatAmount(s)
}
