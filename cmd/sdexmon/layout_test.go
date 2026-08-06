package main

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stellar/go/txnbuild"
)

// TestPairInfoViewFitsTerminal guards against layout regressions where the
// rendered Pair Info screen is wider or taller than the terminal. Overflow
// makes the terminal wrap lines, which corrupts the whole screen.
func TestPairInfoViewFitsTerminal(t *testing.T) {
	sizes := []struct {
		width  int
		height int
	}{
		{140, 60}, // size used by the run script and installed wrapper
		{120, 50},
		{100, 45},
		{180, 60},
		{200, 70},
	}

	for _, size := range sizes {
		m := model{
			currentScreen: screenPairInfo,
			base:          txnbuild.NativeAsset{},
			quote:         curatedAssets["USDC"],
			width:         size.width,
			height:        size.height,
		}

		view := pairInfoView(m)

		if got := lipgloss.Width(view); got > size.width {
			t.Errorf("width %d/height %d: rendered width %d exceeds terminal width %d",
				size.width, size.height, got, size.width)
		}
		if got := lipgloss.Height(view); got > size.height {
			t.Errorf("width %d/height %d: rendered height %d exceeds terminal height %d",
				size.width, size.height, got, size.height)
		}
	}
}
