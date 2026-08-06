package stellar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/sdexmon/sdexmon/internal/models"
)

const (
	expertBaseURL = "https://api.stellar.expert/explorer/public"

	// expertSearchLimit caps how many fuzzy matches we ask for. The list is
	// ranked, so anything past this is noise.
	expertSearchLimit = 50
)

// expertAssetRecord is the subset of a stellar.expert asset record we use.
type expertAssetRecord struct {
	// Asset is "CODE-ISSUER-TYPE", the only field guaranteed to be present.
	Asset string `json:"asset"`
	// Domain is the issuer's verified home domain, empty when unverified.
	Domain string `json:"domain"`
	// Trustlines is [total, funded, authorized] and may be absent.
	Trustlines []int64 `json:"trustlines"`
	TomlInfo   struct {
		Code     string `json:"code"`
		Issuer   string `json:"issuer"`
		Name     string `json:"name"`
		Status   string `json:"status"`
		Decimals int    `json:"decimals"`
		OrgName  string `json:"orgName"`
	} `json:"tomlInfo"`
}

type expertSearchResponse struct {
	Embedded struct {
		Records []expertAssetRecord `json:"records"`
	} `json:"_embedded"`
}

type expertAssetListEntry struct {
	Code     string `json:"code"`
	Issuer   string `json:"issuer"`
	Name     string `json:"name"`
	Org      string `json:"org"`
	Domain   string `json:"domain"`
	Decimals int    `json:"decimals"`
}

type expertAssetList struct {
	Name   string                 `json:"name"`
	Assets []expertAssetListEntry `json:"assets"`
}

func expertClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Second}
}

func getExpertJSON(rawURL string, out interface{}) error {
	resp, err := expertClient().Get(rawURL)
	if err != nil {
		return fmt.Errorf("failed to query stellar.expert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("stellar.expert returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to parse stellar.expert response: %w", err)
	}
	return nil
}

// SearchAssetsByCode runs the fuzzy stellar.expert search.
//
// The results are deliberately not filtered by domain: this mode exists to
// discover assets by code or name. Because anyone can mint a lookalike code,
// callers must display the home domain of every result.
func SearchAssetsByCode(query string) ([]models.AssetSearchResult, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, fmt.Errorf("search term cannot be empty")
	}

	records, err := searchExpertRecords(trimmed)
	if err != nil {
		return nil, err
	}

	assets := recordsToAssets(records)
	if len(assets) == 0 {
		return nil, fmt.Errorf("no assets matched %q", trimmed)
	}

	// Rank verified, widely trusted assets first so lookalikes sink.
	sort.SliceStable(assets, func(i, j int) bool {
		if (assets[i].Domain != "") != (assets[j].Domain != "") {
			return assets[i].Domain != ""
		}
		return assets[i].Trustlines > assets[j].Trustlines
	})

	return assets, nil
}

// SearchAssetsOnExpert is the fallback for domain search when a domain has no
// reachable stellar.toml. Results are restricted to issuers whose home domain
// matches exactly, so lookalike assets from other domains are never offered.
func SearchAssetsOnExpert(domain string) ([]models.AssetSearchResult, error) {
	normalized := NormalizeDomain(domain)
	if normalized == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	records, err := searchExpertRecords(normalized)
	if err != nil {
		return nil, err
	}

	matching := make([]expertAssetRecord, 0, len(records))
	for _, record := range records {
		if strings.EqualFold(strings.TrimSpace(record.Domain), normalized) {
			matching = append(matching, record)
		}
	}

	assets := recordsToAssets(matching)
	if len(assets) == 0 {
		return nil, fmt.Errorf("stellar.expert knows no assets homed at %s", normalized)
	}

	return assets, nil
}

// TopAssets returns the stellar.expert top 50 list: the most active assets on
// the network, curated by metrics rather than by us.
func TopAssets() ([]models.AssetSearchResult, error) {
	var list expertAssetList
	if err := getExpertJSON(expertBaseURL+"/asset-list/top50", &list); err != nil {
		return nil, err
	}

	assets := make([]models.AssetSearchResult, 0, len(list.Assets))
	for _, entry := range list.Assets {
		if entry.Code == "" || !IsValidIssuer(entry.Issuer) {
			continue
		}
		name := entry.Name
		if name == "" {
			name = entry.Code
		}
		org := entry.Org
		if strings.EqualFold(org, "unknown") {
			org = ""
		}
		assets = append(assets, models.AssetSearchResult{
			Code:            entry.Code,
			Issuer:          entry.Issuer,
			Domain:          strings.ToLower(entry.Domain),
			Name:            name,
			Org:             org,
			DisplayDecimals: entry.Decimals,
		})
	}

	if len(assets) == 0 {
		return nil, fmt.Errorf("stellar.expert returned an empty top 50 list")
	}

	return assets, nil
}

func searchExpertRecords(term string) ([]expertAssetRecord, error) {
	endpoint := fmt.Sprintf("%s/asset?search=%s&limit=%d",
		expertBaseURL, url.QueryEscape(term), expertSearchLimit)

	var result expertSearchResponse
	if err := getExpertJSON(endpoint, &result); err != nil {
		return nil, err
	}
	return result.Embedded.Records, nil
}

func recordsToAssets(records []expertAssetRecord) []models.AssetSearchResult {
	assets := make([]models.AssetSearchResult, 0, len(records))
	for _, record := range records {
		code, issuer := record.TomlInfo.Code, record.TomlInfo.Issuer
		if code == "" || issuer == "" {
			// The `asset` field is "CODE-ISSUER-TYPE" and always present.
			parts := strings.Split(record.Asset, "-")
			if len(parts) < 2 {
				continue
			}
			code, issuer = parts[0], parts[1]
		}
		if code == "" || !IsValidIssuer(issuer) {
			continue
		}

		trustlines := 0
		if len(record.Trustlines) > 0 {
			trustlines = int(record.Trustlines[0])
		}

		assets = append(assets, models.AssetSearchResult{
			Code:            code,
			Issuer:          issuer,
			Domain:          strings.ToLower(strings.TrimSpace(record.Domain)),
			Name:            record.TomlInfo.Name,
			Org:             record.TomlInfo.OrgName,
			Status:          strings.ToLower(record.TomlInfo.Status),
			DisplayDecimals: record.TomlInfo.Decimals,
			Trustlines:      trustlines,
		})
	}
	return assets
}
