package stellar

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sdexmon/sdexmon/internal/models"
)

const (
	// wellKnownPath is where SEP-1 requires the stellar.toml to be published.
	wellKnownPath = "/.well-known/stellar.toml"

	// tomlMaxSize is the SEP-1 maximum stellar.toml size (100 KiB). Reading is
	// capped so a hostile or misconfigured host cannot stream us forever.
	tomlMaxSize = 100 * 1024

	// maxLinkedCurrencyFetches bounds how many per-currency `toml` links we
	// follow, since each one is an extra network round trip.
	maxLinkedCurrencyFetches = 12
)

// tomlCurrency is the subset of a SEP-1 [[CURRENCIES]] entry we need.
//
// It deliberately does not mirror the full SEP-1 schema: several issuers
// publish fields with types that differ from the spec (for example a boolean
// `regulated` where a string is expected), and a strict struct would fail the
// whole document over an unrelated field.
type tomlCurrency struct {
	Code            string `toml:"code"`
	CodeTemplate    string `toml:"code_template"`
	Issuer          string `toml:"issuer"`
	Name            string `toml:"name"`
	Desc            string `toml:"desc"`
	Status          string `toml:"status"`
	DisplayDecimals int    `toml:"display_decimals"`
	Toml            string `toml:"toml"`
}

type tomlDocument struct {
	Currencies []tomlCurrency `toml:"CURRENCIES"`
}

// NormalizeDomain strips anything a user might paste around a bare host name,
// e.g. "https://zeam.money/", so the well-known URL is always well formed.
func NormalizeDomain(domain string) string {
	d := strings.TrimSpace(strings.ToLower(domain))
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	if idx := strings.IndexAny(d, "/?#"); idx >= 0 {
		d = d[:idx]
	}
	return strings.Trim(d, ". ")
}

// TomlURL returns the SEP-1 well-known location for a domain.
func TomlURL(domain string) string {
	return "https://" + NormalizeDomain(domain) + wellKnownPath
}

// FetchAssetsFromDomain resolves the assets a domain actually vouches for by
// reading its SEP-1 stellar.toml. This is the authoritative source: anyone can
// name an asset "ZARZ", but only the domain owner can list an issuer in their
// own stellar.toml.
func FetchAssetsFromDomain(domain string) ([]models.AssetSearchResult, error) {
	normalized := NormalizeDomain(domain)
	if normalized == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	doc, err := fetchTomlDocument(client, TomlURL(normalized))
	if err != nil {
		return nil, err
	}

	currencies := make([]tomlCurrency, 0, len(doc.Currencies))
	linkFetches := 0

	for _, currency := range doc.Currencies {
		// SEP-1 allows a currency to be published in its own linked file.
		if currency.Toml != "" && (currency.Code == "" || currency.Issuer == "") {
			if linkFetches >= maxLinkedCurrencyFetches {
				continue
			}
			linkFetches++
			linked, err := fetchTomlDocument(client, currency.Toml)
			if err != nil {
				continue
			}
			currencies = append(currencies, linked.Currencies...)
			continue
		}
		currencies = append(currencies, currency)
	}

	assets := currenciesToAssets(currencies, normalized)
	if len(assets) == 0 {
		return nil, fmt.Errorf("%s lists no usable assets", TomlURL(normalized))
	}

	return assets, nil
}

func fetchTomlDocument(client *http.Client, rawURL string) (*tomlDocument, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return nil, fmt.Errorf("invalid stellar.toml URL %q", rawURL)
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned status %d", rawURL, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, tomlMaxSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", rawURL, err)
	}

	doc, err := parseTomlDocument(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", rawURL, err)
	}
	return doc, nil
}

func parseTomlDocument(body []byte) (*tomlDocument, error) {
	var doc tomlDocument
	if _, err := toml.Decode(string(body), &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func currenciesToAssets(currencies []tomlCurrency, domain string) []models.AssetSearchResult {
	assets := make([]models.AssetSearchResult, 0, len(currencies))
	for _, currency := range currencies {
		if asset, ok := currencyToAsset(currency, domain); ok {
			assets = append(assets, asset)
		}
	}
	return assets
}

// currencyToAsset validates a [[CURRENCIES]] entry and converts it. Entries
// that are templates, retired, or missing an issuer are dropped rather than
// shown as selectable assets.
func currencyToAsset(currency tomlCurrency, domain string) (models.AssetSearchResult, bool) {
	code := strings.TrimSpace(currency.Code)
	issuer := strings.TrimSpace(currency.Issuer)

	if code == "" || currency.CodeTemplate != "" {
		return models.AssetSearchResult{}, false
	}
	if !IsValidIssuer(issuer) {
		return models.AssetSearchResult{}, false
	}
	if strings.EqualFold(strings.TrimSpace(currency.Status), "dead") {
		return models.AssetSearchResult{}, false
	}

	return models.AssetSearchResult{
		Code:            code,
		Issuer:          issuer,
		Domain:          domain,
		Name:            strings.TrimSpace(currency.Name),
		Status:          strings.ToLower(strings.TrimSpace(currency.Status)),
		DisplayDecimals: currency.DisplayDecimals,
		Verified:        true,
	}, true
}

// IsValidIssuer reports whether an account looks like a Stellar public key.
func IsValidIssuer(issuer string) bool {
	if len(issuer) != 56 || !strings.HasPrefix(issuer, "G") {
		return false
	}
	for _, r := range issuer {
		isUpper := r >= 'A' && r <= 'Z'
		isDigit := r >= '2' && r <= '7'
		if !isUpper && !isDigit {
			return false
		}
	}
	return true
}

// ResolveDomainAssets returns the assets published by a domain.
//
// The domain's own stellar.toml is authoritative and always tried first. Only
// when it cannot be reached do we fall back to stellar.expert, and even then
// results are restricted to issuers whose home domain matches exactly, so
// lookalike assets from unrelated domains are never offered.
func ResolveDomainAssets(domain string) ([]models.AssetSearchResult, error) {
	normalized := NormalizeDomain(domain)
	if normalized == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	assets, tomlErr := FetchAssetsFromDomain(normalized)
	if tomlErr == nil {
		return assets, nil
	}

	fallback, expertErr := SearchAssetsOnExpert(normalized)
	if expertErr != nil || len(fallback) == 0 {
		return nil, fmt.Errorf("no assets found for %s (%v)", normalized, tomlErr)
	}

	return fallback, nil
}
