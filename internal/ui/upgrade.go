package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	upgradeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")). // Red border
			Padding(2, 4).
			MarginTop(2).
			MarginBottom(2)

	upgradeHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196")).
				MarginBottom(1)

	upgradeTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	upgradeCommandStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("40")). // Green
				Bold(true).
				MarginTop(1).
				MarginBottom(1)
)

// RenderUpgradeRequired renders the upgrade required screen
func RenderUpgradeRequired(currentVersion, latestVersion string, width, height int) string {
	content := upgradeHeaderStyle.Render("⚠ UPDATE REQUIRED ⚠") + "\n\n"

	content += upgradeTextStyle.Render(
		fmt.Sprintf("Your version: %s\n", currentVersion),
	)
	content += upgradeTextStyle.Render(
		fmt.Sprintf("Latest version: %s\n\n", latestVersion),
	)

	content += upgradeTextStyle.Render(
		"A new version of sdexmon is available and must be installed.\n\n",
	)

	content += upgradeTextStyle.Render("To upgrade, run:\n")
	content += upgradeCommandStyle.Render(
		"  curl -sSL https://raw.githubusercontent.com/sdexmon/sdexmon/main/install.sh | bash",
	)

	content += "\n\n"
	content += upgradeTextStyle.Render("Or download from: https://github.com/sdexmon/sdexmon/releases/latest")

	box := upgradeBoxStyle.Render(content)

	// Center the box
	boxHeight := lipgloss.Height(box)
	boxWidth := lipgloss.Width(box)

	verticalPadding := (height - boxHeight) / 2
	horizontalPadding := (width - boxWidth) / 2

	if verticalPadding < 0 {
		verticalPadding = 0
	}
	if horizontalPadding < 0 {
		horizontalPadding = 0
	}

	topPadding := ""
	for i := 0; i < verticalPadding; i++ {
		topPadding += "\n"
	}

	leftPadding := ""
	for i := 0; i < horizontalPadding; i++ {
		leftPadding += " "
	}

	return topPadding + leftPadding + box
}
