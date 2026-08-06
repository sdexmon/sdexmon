package stellar

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const sampleToml = `
NETWORK_PASSPHRASE = "Public Global Stellar Network ; September 2015"

[[CURRENCIES]]
code = "ZARZ"
issuer = "GAROH4EV3WVVTRQKEY43GZK3XSRBEYETRVZ7SVG5LHWOAANSMCTJBB3U"
name = "zeam ZAR"
display_decimals = 2
status = "live"
# a boolean where SEP-1 documents a string: must not fail the whole document
regulated = false

[[CURRENCIES]]
code = "yZARZ"
issuer = "GDZBAWPUGAJI4CQTO53O6Y33WEZ4IRVDBLDYUY6EKGICP7OK53OYZARZ"
name = "Yield ZARZ"
status = "test"

[[CURRENCIES]]
code = "OLD"
issuer = "GAKTLPC4ZV37SSCITQ5IS5AQ4WPF4CF4VZJQPPAROSGXMYOATF5U6XPR"
status = "dead"

[[CURRENCIES]]
code_template = "BTC?"
issuer = "GAKTLPC4ZV37SSCITQ5IS5AQ4WPF4CF4VZJQPPAROSGXMYOATF5U6XPR"

[[CURRENCIES]]
code = "NOISSUER"
`

// TestParseTomlCurrencies checks the SEP-1 subset we depend on, including the
// entries that must be dropped rather than offered as selectable assets.
func TestParseTomlCurrencies(t *testing.T) {
	doc, err := parseTomlDocument([]byte(sampleToml))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	assets := currenciesToAssets(doc.Currencies, "zeam.money")
	if len(assets) != 2 {
		t.Fatalf("got %d assets, want 2 (dead, template and issuer-less entries dropped): %+v", len(assets), assets)
	}

	first := assets[0]
	if first.Code != "ZARZ" || first.Name != "zeam ZAR" {
		t.Errorf("first asset = %+v, want the ZARZ entry", first)
	}
	if first.Domain != "zeam.money" {
		t.Errorf("domain = %q, want zeam.money", first.Domain)
	}
	if !first.Verified {
		t.Error("assets read from a domain's own stellar.toml must be marked verified")
	}
	if first.DisplayDecimals != 2 {
		t.Errorf("display decimals = %d, want 2", first.DisplayDecimals)
	}
	if assets[1].Status != "test" {
		t.Errorf("status = %q, want test to be carried through", assets[1].Status)
	}
}

// TestFetchAssetsFromDomainReadsWellKnownPath is the regression for domain
// search hitting a fuzzy index instead of the domain's own SEP-1 file.
func TestFetchAssetsFromDomainReadsWellKnownPath(t *testing.T) {
	var requested string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = r.URL.Path
		w.Write([]byte(sampleToml))
	}))
	defer server.Close()

	doc, err := fetchTomlDocument(server.Client(), server.URL+wellKnownPath)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if requested != wellKnownPath {
		t.Errorf("requested %q, want %q", requested, wellKnownPath)
	}
	if len(doc.Currencies) != 5 {
		t.Errorf("parsed %d currencies, want all 5 raw entries", len(doc.Currencies))
	}
}

func TestTomlURL(t *testing.T) {
	cases := map[string]string{
		"zeam.money":          "https://zeam.money/.well-known/stellar.toml",
		"https://zeam.money/": "https://zeam.money/.well-known/stellar.toml",
		"  ZEAM.money  ":      "https://zeam.money/.well-known/stellar.toml",
		"http://zeam.money/x": "https://zeam.money/.well-known/stellar.toml",
	}
	for input, want := range cases {
		if got := TomlURL(input); got != want {
			t.Errorf("TomlURL(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestIsValidIssuer(t *testing.T) {
	valid := "GAROH4EV3WVVTRQKEY43GZK3XSRBEYETRVZ7SVG5LHWOAANSMCTJBB3U"
	if !IsValidIssuer(valid) {
		t.Errorf("%s should be a valid issuer", valid)
	}
	for _, invalid := range []string{"", "GABC", valid[:55], "M" + valid[1:], valid[:55] + "1"} {
		if IsValidIssuer(invalid) {
			t.Errorf("%q should not be a valid issuer", invalid)
		}
	}
}
