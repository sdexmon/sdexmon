package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sdexmon/sdexmon/internal/config"
	"github.com/sdexmon/sdexmon/internal/models"
)

// quits reports whether a command asks the program to exit. The text inputs
// also emit a harmless blink command, so nilness alone proves nothing.
func quits(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, isQuit := cmd().(tea.QuitMsg)
	return isQuit
}

// withTempConfig points the config path at a throwaway home directory, loads
// the given pairs through the real loader, and restores the package globals
// afterwards.
func withTempConfig(t *testing.T, pairs []config.Pair) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

	previousPairs := configuredPairs
	previousPools := liquidityPoolIDs
	previousConfig := appConfig
	t.Cleanup(func() {
		configuredPairs = previousPairs
		liquidityPoolIDs = previousPools
		appConfig = previousConfig
	})

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Pairs = pairs
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if err := loadConfiguration(); err != nil {
		t.Fatal(err)
	}
}

func testPairs() []config.Pair {
	return []config.Pair{
		{Name: "XLM/USDC", Base: "XLM:native", Quote: "USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"},
		{Name: "XLM/USDZ", Base: "XLM:native", Quote: "USDZ:GAKTLPC4ZV37SSCITQ5IS5AQ4WPF4CF4VZJQPPAROSGXMYOATF5U6XPR"},
	}
}

// TestMaintenanceScreenIsReachable guards the regression where the maintenance
// views existed but no key or route led to them.
func TestMaintenanceScreenIsReachable(t *testing.T) {
	withTempConfig(t, testPairs())

	for _, screen := range []screenState{screenLanding, screenPairInfo} {
		start := model{currentScreen: screen, maintenanceState: initMaintenanceState()}
		next, _ := start.Update(keyMsg("m"))
		opened := next.(model)
		if opened.currentScreen != screenMaintenance {
			t.Fatalf("m from screen %d should open maintenance, got %d", screen, opened.currentScreen)
		}
		if opened.maintenanceState.Screen != models.MaintenanceMenu {
			t.Fatalf("maintenance should start on the menu, got %d", opened.maintenanceState.Screen)
		}
		if opened.View() == "" {
			t.Fatal("maintenance screen rendered nothing")
		}
	}
}

// TestMaintenanceMenuRoutes checks every menu entry now leads somewhere, since
// remove and list used to be labelled "coming soon".
func TestMaintenanceMenuRoutes(t *testing.T) {
	withTempConfig(t, testPairs())

	cases := []struct {
		key  string
		want models.MaintenanceScreen
	}{
		{"1", models.AssetASourceSelect},
		{"2", models.PairRemoveSelection},
		{"3", models.PairList},
	}

	for _, tc := range cases {
		start := model{currentScreen: screenMaintenance, maintenanceState: initMaintenanceState()}
		next, _ := start.Update(keyMsg(tc.key))
		got := next.(model).maintenanceState.Screen
		if got != tc.want {
			t.Errorf("menu key %q went to screen %d, want %d", tc.key, got, tc.want)
		}
	}

	start := model{currentScreen: screenMaintenance, maintenanceState: initMaintenanceState()}
	next, _ := start.Update(keyMsg("esc"))
	if got := next.(model).currentScreen; got != screenLanding {
		t.Errorf("esc should leave maintenance for the landing screen, got %d", got)
	}
}

// TestMaintenanceRemovesPairFromConfig drives the remove flow end to end and
// asserts both the file and the in-memory pair list are updated.
func TestMaintenanceRemovesPairFromConfig(t *testing.T) {
	withTempConfig(t, testPairs())

	m := model{currentScreen: screenMaintenance, maintenanceState: initMaintenanceState()}
	for _, key := range []string{"2", "enter", "y"} {
		next, _ := m.Update(keyMsg(key))
		m = next.(model)
	}

	if m.maintenanceState.ErrorMessage != "" {
		t.Fatalf("unexpected error: %s", m.maintenanceState.ErrorMessage)
	}

	saved, err := config.ListCustomPairs()
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 1 || saved[0].Name != "XLM/USDZ" {
		t.Fatalf("config pairs after removal = %+v", saved)
	}
	if len(configuredPairs) != 1 {
		t.Fatalf("in-memory pairs = %d, want 1 after reload", len(configuredPairs))
	}
	if !strings.Contains(m.maintenanceState.StatusMessage, "Removed pair") {
		t.Errorf("status message = %q, want a removal confirmation", m.maintenanceState.StatusMessage)
	}
}

// TestMaintenanceRemoveCanBeCancelled makes sure the confirmation step is a
// real gate and not just decoration.
func TestMaintenanceRemoveCanBeCancelled(t *testing.T) {
	withTempConfig(t, testPairs())

	m := model{currentScreen: screenMaintenance, maintenanceState: initMaintenanceState()}
	for _, key := range []string{"2", "enter", "n"} {
		next, _ := m.Update(keyMsg(key))
		m = next.(model)
	}

	if m.maintenanceState.Screen != models.PairRemoveSelection {
		t.Errorf("cancelling should return to the pair list, got screen %d", m.maintenanceState.Screen)
	}
	saved, err := config.ListCustomPairs()
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 2 {
		t.Fatalf("config pairs = %d, want 2 after cancelling", len(saved))
	}
}

// TestDomainInputAcceptsQ covers the global quit shortcut swallowing letters
// that are legitimate domain characters.
func TestDomainInputAcceptsQ(t *testing.T) {
	withTempConfig(t, testPairs())

	m := model{currentScreen: screenMaintenance, maintenanceState: initMaintenanceState()}
	for _, key := range []string{"1", "1"} { // add pair -> domain search
		next, _ := m.Update(keyMsg(key))
		m = next.(model)
	}
	if m.maintenanceState.Screen != models.AssetAQueryInput {
		t.Fatalf("domain search should open the query field, got screen %d", m.maintenanceState.Screen)
	}

	next, cmd := m.Update(keyMsg("q"))
	if quits(cmd) {
		t.Fatal("q while typing a domain must not quit the app")
	}
	if got := next.(model).maintenanceState.QueryInputA.Value(); got != "q" {
		t.Fatalf("domain input = %q, want %q", got, "q")
	}

	// ctrl+c must still be an unconditional escape hatch.
	if _, cmd := next.(model).Update(tea.KeyMsg{Type: tea.KeyCtrlC}); !quits(cmd) {
		t.Error("ctrl+c must always quit, even while typing")
	}
}

// TestSearchSourceSelection covers the three ways an asset can be looked up.
// The domain and code sources prompt for a query; the top 50 list must not.
func TestSearchSourceSelection(t *testing.T) {
	withTempConfig(t, testPairs())

	cases := []struct {
		key        string
		wantMode   models.AssetSearchMode
		wantScreen models.MaintenanceScreen
		wantPrompt string
	}{
		{"1", models.SearchByDomain, models.AssetAQueryInput, "Domain > "},
		{"2", models.SearchByCode, models.AssetAQueryInput, "Asset > "},
		{"3", models.SearchTop50, models.AssetASourceSelect, ""},
	}

	for _, tc := range cases {
		m := model{currentScreen: screenMaintenance, maintenanceState: initMaintenanceState()}
		next, _ := m.Update(keyMsg("1"))
		m = next.(model)

		next, cmd := m.Update(keyMsg(tc.key))
		m = next.(model)

		if m.maintenanceState.SearchModeA != tc.wantMode {
			t.Errorf("source key %q selected mode %d, want %d", tc.key, m.maintenanceState.SearchModeA, tc.wantMode)
		}
		if m.maintenanceState.Screen != tc.wantScreen {
			t.Errorf("source key %q went to screen %d, want %d", tc.key, m.maintenanceState.Screen, tc.wantScreen)
		}
		if tc.wantPrompt != "" {
			if got := m.maintenanceState.QueryInputA.Prompt; got != tc.wantPrompt {
				t.Errorf("source key %q prompt = %q, want %q", tc.key, got, tc.wantPrompt)
			}
			continue
		}
		if cmd == nil {
			t.Errorf("source key %q should start the lookup immediately", tc.key)
		}
		if m.maintenanceState.LoadingMessage == "" {
			t.Errorf("source key %q should report that it is loading", tc.key)
		}
	}
}

// TestSearchResultsShowHomeDomain is the regression for domain search offering
// lookalike assets: whichever source is used, every row names the home domain
// so an unrelated issuer is obvious.
func TestSearchResultsShowHomeDomain(t *testing.T) {
	withTempConfig(t, testPairs())

	m := model{currentScreen: screenMaintenance, maintenanceState: initMaintenanceState()}
	for _, key := range []string{"1", "1"} {
		next, _ := m.Update(keyMsg(key))
		m = next.(model)
	}

	next, _ := m.Update(models.AssetSearchResultsMsg{
		Leg:    models.AssetLegA,
		Source: "https://zeam.money/.well-known/stellar.toml (SEP-1 verified)",
		Assets: []models.AssetSearchResult{{
			Code:     "ZARZ",
			Issuer:   "GAROH4EV3WVVTRQKEY43GZK3XSRBEYETRVZ7SVG5LHWOAANSMCTJBB3U",
			Domain:   "zeam.money",
			Name:     "zeam ZAR",
			Verified: true,
		}},
	})
	m = next.(model)

	if m.maintenanceState.Screen != models.AssetASelection {
		t.Fatalf("results should open the selection screen, got %d", m.maintenanceState.Screen)
	}

	view := m.View()
	for _, want := range []string{"zeam.money", "HOME DOMAIN", "stellar.toml"} {
		if !strings.Contains(view, want) {
			t.Errorf("selection view is missing %q", want)
		}
	}
}

// TestLateResultsDoNotHijackScreen guards against a slow search dragging the
// user back into a list they already navigated away from.
func TestLateResultsDoNotHijackScreen(t *testing.T) {
	withTempConfig(t, testPairs())

	m := model{currentScreen: screenMaintenance, maintenanceState: initMaintenanceState()}
	for _, key := range []string{"1", "1", "esc", "esc"} {
		next, _ := m.Update(keyMsg(key))
		m = next.(model)
	}
	if m.maintenanceState.Screen != models.MaintenanceMenu {
		t.Fatalf("expected to be back on the menu, got screen %d", m.maintenanceState.Screen)
	}

	next, _ := m.Update(models.AssetSearchResultsMsg{
		Leg: models.AssetLegA,
		Assets: []models.AssetSearchResult{{
			Code:   "ZARZ",
			Issuer: "GAROH4EV3WVVTRQKEY43GZK3XSRBEYETRVZ7SVG5LHWOAANSMCTJBB3U",
			Domain: "zeam.money",
		}},
	})

	if got := next.(model).maintenanceState.Screen; got != models.MaintenanceMenu {
		t.Errorf("late results moved the user to screen %d, want the menu", got)
	}
}

// TestEmptyPairListIsSafeToBrowse guards the cursor bounds when the config has
// no pairs left to remove.
func TestEmptyPairListIsSafeToBrowse(t *testing.T) {
	withTempConfig(t, []config.Pair{})

	previous := configuredPairs
	configuredPairs = nil
	t.Cleanup(func() { configuredPairs = previous })

	m := model{currentScreen: screenMaintenance, maintenanceState: initMaintenanceState()}
	for _, key := range []string{"2", "down", "enter"} {
		next, _ := m.Update(keyMsg(key))
		m = next.(model)
	}

	if m.maintenanceState.Screen != models.PairRemoveSelection {
		t.Errorf("empty list should stay on the selection screen, got %d", m.maintenanceState.Screen)
	}
	if m.View() == "" {
		t.Error("empty pair list rendered nothing")
	}
}
