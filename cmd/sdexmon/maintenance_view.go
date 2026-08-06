package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/stellar/go/txnbuild"

	"github.com/sdexmon/sdexmon/internal/config"
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
	case models.PairRemoveSelection:
		return maintenancePairRemoveView(m)
	case models.PairRemoveConfirmation:
		return maintenanceRemoveConfirmationView(m)
	case models.PairList:
		return maintenancePairListView(m)
	default:
		return maintenanceMenuView(m)
	}
}

// maintenanceShortcuts returns the footer hints for the active sub-screen.
func maintenanceShortcuts(screen models.MaintenanceScreen) string {
	switch screen {
	case models.AssetADomainInput, models.AssetBDomainInput:
		return "enter: search  esc: back  ctrl+c: quit"
	case models.AssetASelection, models.AssetBSelection:
		return "up/down: navigate  enter: select  esc: back  q: quit"
	case models.PairConfirmation:
		return "enter: add pair  esc: back  q: quit"
	case models.PairRemoveSelection:
		return "up/down: navigate  enter: remove  esc: back  q: quit"
	case models.PairRemoveConfirmation:
		return "y: confirm remove  n: cancel  q: quit"
	case models.PairList:
		return "up/down: scroll  esc: back  q: quit"
	default:
		return "1: add  2: remove  3: list  esc: back  q: quit"
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
		"2. Remove Asset Pair",
		"3. View Configured Pairs",
		"",
		dimStyle.Render(fmt.Sprintf("%d pair(s) configured in %s", len(configuredPairs), config.GetConfigPath())),
		"",
	}

	if m.maintenanceState.StatusMessage != "" {
		lines = append(lines, greenStyle.Render(m.maintenanceState.StatusMessage))
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

func maintenancePairRemoveView(m model) string {
	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Remove Asset Pair"),
		"",
		boldStyle.Render("Select Pair To Remove"),
		"",
	}

	if len(configuredPairs) == 0 {
		lines = append(lines, dimStyle.Render("No pairs configured"))
	} else {
		lines = append(lines, renderMaintenancePairRows(m.maintenanceState.PairCursor, true)...)
	}

	lines = append(lines, "")

	if m.maintenanceState.ErrorMessage != "" {
		lines = append(lines, "", errorStyle.Render(m.maintenanceState.ErrorMessage))
	}

	return maintenanceFrame(m, lines)
}

func maintenanceRemoveConfirmationView(m model) string {
	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Remove Asset Pair"),
		"",
		boldStyle.Render("CONFIRM REMOVAL"),
		"",
	}

	pair, ok := maintenancePairAt(m.maintenanceState.PairCursor)
	if !ok {
		lines = append(lines, dimStyle.Render("No pair selected"))
	} else {
		baseLabel, quoteLabel := "?", "?"
		if base, quote, resolved := assetsForPair(pair); resolved {
			baseLabel = assetIdentityLabel(base)
			quoteLabel = assetIdentityLabel(quote)
		}
		lines = append(lines,
			fmt.Sprintf("Pair:  %s", pairLabel(pair)),
			fmt.Sprintf("Base:  %s", baseLabel),
			fmt.Sprintf("Quote: %s", quoteLabel),
			"",
			dimStyle.Render("The pair is removed from "+config.GetConfigPath()+"."),
			dimStyle.Render("No on-chain state is affected."),
			"",
			redStyle.Render("Press Y (or ENTER) to remove this pair"),
			dimStyle.Render("Press N (or ESC) to cancel"),
		)
	}

	if m.maintenanceState.ErrorMessage != "" {
		lines = append(lines, "", errorStyle.Render(m.maintenanceState.ErrorMessage))
	}

	return maintenanceFrame(m, lines)
}

func maintenancePairListView(m model) string {
	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Configured Pairs"),
		"",
		boldStyle.Render(fmt.Sprintf("%d PAIR(S)", len(configuredPairs))),
		dimStyle.Render(config.GetConfigPath()),
		"",
	}

	if len(configuredPairs) == 0 {
		lines = append(lines, dimStyle.Render("No pairs configured"))
	} else {
		lines = append(lines, renderMaintenancePairRows(m.maintenanceState.PairCursor, false)...)
	}

	return maintenanceFrame(m, lines)
}

// renderMaintenancePairRows renders a scrolling window of configured pairs.
// The cursor is only highlighted when the list is interactive.
func renderMaintenancePairRows(cursor int, highlight bool) []string {
	const windowSize = 15

	start := cursor - windowSize/2
	if start < 0 {
		start = 0
	}
	end := start + windowSize
	if end > len(configuredPairs) {
		end = len(configuredPairs)
	}

	rows := make([]string, 0, end-start+1)
	for i := start; i < end; i++ {
		pair := configuredPairs[i]
		label := padRight(pairLabel(pair), 14)

		detail := "unresolved assets"
		if base, quote, ok := assetsForPair(pair); ok {
			detail = fmt.Sprintf("%s -> %s", shortAssetIdentity(base), shortAssetIdentity(quote))
			if poolID := configuredPoolID(base, quote); len(poolID) >= 8 {
				detail += fmt.Sprintf("  LP %s...", poolID[:8])
			}
		}
		if pair.Favorite {
			detail += "  [fav]"
		}

		row := label + "  " + detail
		if highlight && i == cursor {
			rows = append(rows, selectedStyle.Render("> "+row))
		} else {
			rows = append(rows, pairItemStyle.Render("  "+row))
		}
	}

	if start > 0 || end < len(configuredPairs) {
		rows = append(rows, dimStyle.Render(fmt.Sprintf("  showing %d-%d of %d", start+1, end, len(configuredPairs))))
	}
	return rows
}

// shortAssetIdentity renders CODE:ISSUER with a truncated issuer so a full row
// still fits the 140 column layout.
func shortAssetIdentity(asset txnbuild.Asset) string {
	if asset.IsNative() {
		return "XLM"
	}
	value := assetString(asset)
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 || len(parts[1]) <= 12 {
		return value
	}
	return parts[0] + ":" + parts[1][:6] + ".." + parts[1][len(parts[1])-4:]
}

// assetIdentityLabel spells out an asset in full, including its issuer, for the
// confirmation screen where the exact identity being removed matters.
func assetIdentityLabel(asset txnbuild.Asset) string {
	if asset.IsNative() {
		return "XLM (native)"
	}
	return assetString(asset)
}

// configuredPoolID looks up the liquidity pool recorded for a pair, preferring
// the issuer-safe key over the legacy code-only one.
func configuredPoolID(base, quote txnbuild.Asset) string {
	if id := liquidityPoolIDs[poolMapKey(base, quote)]; id != "" {
		return id
	}
	return liquidityPoolIDs[assetShort(base)+"-"+assetShort(quote)]
}

// maintenanceFrame pads content to the terminal height and appends the footer,
// matching the layout of the other maintenance screens.
func maintenanceFrame(m model, lines []string) string {
	content := strings.Join(lines, "\n")
	targetHeight := 60
	if m.height > 0 {
		targetHeight = m.height
	}
	paddingLines := targetHeight - lipgloss.Height(content) - 2
	if paddingLines < 0 {
		paddingLines = 0
	}
	padding := strings.Repeat("\n", paddingLines)

	return lipgloss.JoinVertical(lipgloss.Left, content, padding, m.bottomLine())
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
