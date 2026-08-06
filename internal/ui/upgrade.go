package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	upgradeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("214")). // Amber border
			Padding(2, 4).
			MarginTop(2).
			MarginBottom(2)

	upgradeHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214")). // Amber: advisory, not fatal
				MarginBottom(1)

	upgradeHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	upgradeTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	upgradeCommandStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("40")). // Green
				Bold(true).
				MarginTop(1).
				MarginBottom(1)
)

// RenderUpgradeAvailable renders the update notice screen. The notice is
// advisory: the caller must keep esc (dismiss) and q/ctrl+c (quit) working so an
// outdated build can never lock the user out of the app.
func RenderUpgradeAvailable(currentVersion, latestVersion, upgradeCommand string, width, height int) string {
	content := upgradeHeaderStyle.Render("UPDATE AVAILABLE") + "\n\n"

	content += upgradeTextStyle.Render(
		fmt.Sprintf("Your version: %s\n", currentVersion),
	)
	content += upgradeTextStyle.Render(
		fmt.Sprintf("Latest version: %s\n\n", latestVersion),
	)

	content += upgradeTextStyle.Render(
		"Press enter to run the installer now, or upgrade manually with:\n",
	)
	content += upgradeCommandStyle.Render("  " + upgradeCommand)

	content += "\n\n"
	content += upgradeTextStyle.Render("Or download from: https://github.com/sdexmon/sdexmon/releases/latest")
	content += "\n\n"
	content += upgradeHintStyle.Render("enter: upgrade now   esc: continue without upgrading   q: quit")

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
