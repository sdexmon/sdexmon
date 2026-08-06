package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(key string) tea.KeyMsg {
	if key == "esc" {
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

// TestUpgradeScreenIsEscapable guards the regression where the upgrade screen
// swallowed every key, including quit, leaving users on an outdated build with
// no way out.
func TestUpgradeScreenIsEscapable(t *testing.T) {
	m := model{
		currentScreen:   screenUpgrade,
		updateAvailable: true,
		latestVersion:   "v9.9.9",
		upgradeReturn:   screenPairInfo,
	}

	if _, cmd := m.Update(keyMsg("q")); cmd == nil {
		t.Error("q on the upgrade screen must quit")
	}
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); cmd == nil {
		t.Error("ctrl+c on the upgrade screen must quit")
	}

	next, _ := m.Update(keyMsg("esc"))
	if got := next.(model).currentScreen; got != screenPairInfo {
		t.Errorf("esc should return to the previous screen, got screen %d", got)
	}
}

// TestUpgradeShortcutRequiresAvailableUpdate makes sure the u shortcut only
// works while the startup check actually found a newer release.
func TestUpgradeShortcutRequiresAvailableUpdate(t *testing.T) {
	upToDate := model{currentScreen: screenLanding}
	next, _ := upToDate.Update(keyMsg("u"))
	if got := next.(model).currentScreen; got != screenLanding {
		t.Errorf("u without an available update should be a no-op, got screen %d", got)
	}

	outdated := model{currentScreen: screenLanding, updateAvailable: true, latestVersion: "v9.9.9"}
	next, _ = outdated.Update(keyMsg("u"))
	updated := next.(model)
	if updated.currentScreen != screenUpgrade {
		t.Errorf("u should open the upgrade notice, got screen %d", updated.currentScreen)
	}
	if updated.upgradeReturn != screenLanding {
		t.Errorf("upgrade notice should remember the landing screen, got %d", updated.upgradeReturn)
	}
}

// TestUpgradeFooterHint checks the shortcut is advertised only when relevant.
func TestUpgradeFooterHint(t *testing.T) {
	withUpdate := model{currentScreen: screenLanding, updateAvailable: true, width: 140}
	if got := withUpdate.bottomLine(); !strings.Contains(got, "u: upgrade") {
		t.Error("footer should advertise u: upgrade when an update is available")
	}

	withoutUpdate := model{currentScreen: screenLanding, width: 140}
	if got := withoutUpdate.bottomLine(); strings.Contains(got, "u: upgrade") {
		t.Error("footer should not advertise u: upgrade when up to date")
	}
}
