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
