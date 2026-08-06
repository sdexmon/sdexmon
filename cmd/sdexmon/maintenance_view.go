package main

import (
	"fmt"
	"strconv"
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
	case models.AssetASourceSelect, models.AssetBSourceSelect:
		return maintenanceSourceSelectView(m)
	case models.AssetAQueryInput, models.AssetBQueryInput:
		return maintenanceQueryInputView(m)
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
	case models.AssetASourceSelect, models.AssetBSourceSelect:
		return "up/down: navigate  1-3: pick source  enter: choose  esc: back  q: quit"
	case models.AssetAQueryInput, models.AssetBQueryInput:
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

// maintenanceLegLabel names the side of the pair the current screen works on.
func maintenanceLegLabel(screen models.MaintenanceScreen) string {
	switch screen {
	case models.AssetASourceSelect, models.AssetAQueryInput, models.AssetASelection:
		return "Asset A"
	default:
		return "Asset B"
	}
}

// maintenanceSourceSelectView asks where the asset should be looked up. The
// order is deliberate: the domain's own stellar.toml is the only source that
// proves an issuer belongs to that domain, so it leads.
func maintenanceSourceSelectView(m model) string {
	assetLabel := maintenanceLegLabel(m.maintenanceState.Screen)

	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Add Asset Pair"),
		"",
		boldStyle.Render("Find " + assetLabel + " By"),
		"",
	}

	if selected := m.maintenanceState.SelectedAssetA; selected != nil && assetLabel == "Asset B" {
		lines = append(lines, dimStyle.Render("Asset A: "+assetSummaryLine(*selected)), "")
	}

	for i, mode := range models.AllAssetSearchModes {
		row := fmt.Sprintf("%d. %s", i+1, mode.Label())
		if i == m.maintenanceState.SourceCursor {
			lines = append(lines, selectedStyle.Render("> "+row))
		} else {
			lines = append(lines, pairItemStyle.Render("  "+row))
		}
		lines = append(lines, dimStyle.Render("     "+mode.Description()))
	}

	lines = append(lines, "")

	if m.maintenanceState.ErrorMessage != "" {
		lines = append(lines, errorStyle.Render(m.maintenanceState.ErrorMessage))
	}
	if m.maintenanceState.LoadingMessage != "" {
		lines = append(lines, dimStyle.Render(m.maintenanceState.LoadingMessage))
	}

	return maintenanceFrame(m, lines)
}

func maintenanceQueryInputView(m model) string {
	isAssetA := m.maintenanceState.Screen == models.AssetAQueryInput
	assetLabel := maintenanceLegLabel(m.maintenanceState.Screen)

	mode := m.maintenanceState.SearchModeB
	input := m.maintenanceState.QueryInputB
	if isAssetA {
		mode = m.maintenanceState.SearchModeA
		input = m.maintenanceState.QueryInputA
	}

	title := "Enter " + assetLabel + " Domain"
	hint := "The assets published in that domain's own stellar.toml are listed"
	if mode == models.SearchByCode {
		title = "Search For " + assetLabel
		hint = "Matches every asset on the network, so confirm the domain before you pick"
	}

	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Add Asset Pair"),
		"",
		boldStyle.Render(title),
		"",
		dimStyle.Render(hint),
		"",
		input.View(),
		"",
	}

	if m.maintenanceState.ErrorMessage != "" {
		lines = append(lines, errorStyle.Render(m.maintenanceState.ErrorMessage))
	}
	if m.maintenanceState.LoadingMessage != "" {
		lines = append(lines, dimStyle.Render(m.maintenanceState.LoadingMessage))
	}

	return maintenanceFrame(m, lines)
}

func maintenanceAssetSelectionView(m model) string {
	isAssetA := m.maintenanceState.Screen == models.AssetASelection
	assetLabel := maintenanceLegLabel(m.maintenanceState.Screen)

	results := m.maintenanceState.SearchResultsB
	cursor := m.maintenanceState.AssetCursorB
	source := m.maintenanceState.SearchSourceB
	if isAssetA {
		results = m.maintenanceState.SearchResultsA
		cursor = m.maintenanceState.AssetCursorA
		source = m.maintenanceState.SearchSourceA
	}

	lines := []string{
		renderVersionInfo(),
		"",
		renderHeader(),
		renderSubtitle("Add Asset Pair"),
		"",
		boldStyle.Render("Select " + assetLabel),
	}

	if source != "" {
		lines = append(lines, dimStyle.Render("Source: "+source))
	}
	lines = append(lines, "", dimStyle.Render(assetRowHeader()))

	if len(results) == 0 {
		lines = append(lines, dimStyle.Render("No assets found"))
	} else {
		const windowSize = 15
		start := cursor - windowSize/2
		if start < 0 {
			start = 0
		}
		end := start + windowSize
		if end > len(results) {
			end = len(results)
		}

		for i := start; i < end; i++ {
			row := assetRow(results[i])
			if i == cursor {
				lines = append(lines, selectedStyle.Render("> "+row))
			} else {
				lines = append(lines, pairItemStyle.Render("  "+row))
			}
		}

		if start > 0 || end < len(results) {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("  showing %d-%d of %d", start+1, end, len(results))))
		}
	}

	lines = append(lines, "")

	if m.maintenanceState.ErrorMessage != "" {
		lines = append(lines, errorStyle.Render(m.maintenanceState.ErrorMessage))
	}

	return maintenanceFrame(m, lines)
}

const (
	assetColCode   = 10
	assetColName   = 24
	assetColDomain = 26
	assetColIssuer = 14
)

func assetRowHeader() string {
	return "  " + padRight("CODE", assetColCode) +
		padRight("NAME", assetColName) +
		padRight("HOME DOMAIN", assetColDomain) +
		padRight("ISSUER", assetColIssuer) + "NOTES"
}

// assetRow always shows the home domain, because an asset code on its own says
// nothing about who issued it.
func assetRow(asset models.AssetSearchResult) string {
	name := firstNonEmpty(asset.Name, asset.Org, "-")
	domain := asset.Domain
	if domain == "" {
		domain = "no home domain"
	}

	notes := []string{}
	if asset.Verified {
		notes = append(notes, "stellar.toml")
	}
	if asset.Status != "" && asset.Status != "live" {
		notes = append(notes, asset.Status)
	}
	if asset.Trustlines > 0 {
		notes = append(notes, formatCount(asset.Trustlines)+" trustlines")
	}

	return padRight(truncateEnd(asset.Code, assetColCode-1), assetColCode) +
		padRight(truncateEnd(name, assetColName-1), assetColName) +
		padRight(truncateEnd(domain, assetColDomain-1), assetColDomain) +
		padRight(shortIssuer(asset.Issuer), assetColIssuer) +
		strings.Join(notes, "  ")
}

// assetSummaryLine describes a chosen asset in one line, issuer included, so a
// lookalike cannot slip through the later screens unnoticed.
func assetSummaryLine(asset models.AssetSearchResult) string {
	domain := asset.Domain
	if domain == "" {
		domain = "no home domain"
	}
	suffix := ""
	if asset.Verified {
		suffix = " [stellar.toml]"
	}
	return fmt.Sprintf("%s:%s  %s%s", asset.Code, shortIssuer(asset.Issuer), domain, suffix)
}

func shortIssuer(issuer string) string {
	if len(issuer) <= 12 {
		return issuer
	}
	return issuer[:6] + ".." + issuer[len(issuer)-4:]
}

func truncateEnd(s string, max int) string {
	if max <= 3 || len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// formatCount groups digits with spaces, matching how amounts are rendered
// elsewhere in the app.
func formatCount(n int) string {
	digits := strconv.Itoa(n)
	var out []byte
	for i := 0; i < len(digits); i++ {
		if i != 0 && (len(digits)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, digits[i])
	}
	return string(out)
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

	// Spell out both issuers and domains: this is the last chance to notice a
	// lookalike asset before it is written to the config.
	if selected := m.maintenanceState.SelectedAssetA; selected != nil {
		lines = append(lines, "Asset A: "+assetSummaryLine(*selected))
	}
	if selected := m.maintenanceState.SelectedAssetB; selected != nil {
		lines = append(lines, "Asset B: "+assetSummaryLine(*selected))
	}
	lines = append(lines, "")

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
