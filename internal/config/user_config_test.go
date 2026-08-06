package config

import (
	"testing"

	"github.com/stellar/go/txnbuild"
)

func TestAddCustomPairUsesUnifiedConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := getDefaultConfig()
	cfg.Pairs = []Pair{{
		Name:  "XLM/USDC",
		Base:  "native",
		Quote: "USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
	}}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	base := txnbuild.CreditAsset{
		Code:   "USD",
		Issuer: "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWHF",
	}
	if err := AddCustomPair(base, txnbuild.NativeAsset{}); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Pairs) != 2 {
		t.Fatalf("pairs = %d, want 2", len(loaded.Pairs))
	}
	if loaded.Pairs[1].Base != AssetString(base) {
		t.Fatalf("base = %q, want full issuer identity %q", loaded.Pairs[1].Base, AssetString(base))
	}
	if loaded.App.Version == "" {
		t.Fatal("saving a custom pair discarded the main config schema")
	}
}

func TestRemoveCustomPairMatchesEitherOrientation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	usdc := txnbuild.CreditAsset{
		Code:   "USDC",
		Issuer: "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
	}
	other := txnbuild.CreditAsset{
		Code:   "USDC",
		Issuer: "GAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAWHF",
	}

	cfg := getDefaultConfig()
	cfg.Pairs = []Pair{
		{Name: "XLM/USDC", Base: "XLM:native", Quote: AssetString(usdc)},
		{Name: "XLM/USDC (other issuer)", Base: "native", Quote: AssetString(other)},
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	// Reversed orientation, and the "native" spelling of the "XLM:native" entry.
	if err := RemoveCustomPair(usdc, txnbuild.NativeAsset{}); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Pairs) != 1 {
		t.Fatalf("pairs = %d, want 1", len(loaded.Pairs))
	}
	if loaded.Pairs[0].Quote != AssetString(other) {
		t.Fatalf("removed the wrong issuer: remaining quote = %q", loaded.Pairs[0].Quote)
	}
}

func TestRemoveCustomPairReportsMissingPair(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := getDefaultConfig()
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	err := RemoveCustomPair(txnbuild.NativeAsset{}, txnbuild.CreditAsset{
		Code:   "USDC",
		Issuer: "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
	})
	if err == nil {
		t.Fatal("removing an unknown pair should report an error")
	}
}

func TestListCustomPairsPreservesFileOrder(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := getDefaultConfig()
	cfg.Pairs = []Pair{
		{Name: "A/B", Base: "native", Quote: "USDC:GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"},
		{Name: "C/D", Base: "native", Quote: "USDZ:GAKTLPC4ZV37SSCITQ5IS5AQ4WPF4CF4VZJQPPAROSGXMYOATF5U6XPR"},
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	pairs, err := ListCustomPairs()
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 2 || pairs[0].Name != "A/B" || pairs[1].Name != "C/D" {
		t.Fatalf("unexpected pair list: %+v", pairs)
	}
}
