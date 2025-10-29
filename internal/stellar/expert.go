package stellar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sdexmon/sdexmon/internal/models"
)

// flexNumber can unmarshal both JSON numbers and strings
type flexNumber string

func (f *flexNumber) UnmarshalJSON(data []byte) error {
	// Try as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = flexNumber(s)
		return nil
	}
	
	// Try as number
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = flexNumber(strconv.FormatFloat(n, 'f', 0, 64))
		return nil
	}
	
	return fmt.Errorf("supply must be a number or string")
}

func (f flexNumber) String() string {
	return string(f)
}

// expertAssetRecord represents a single asset from stellar.expert API
type expertAssetRecord struct {
	Asset    string     `json:"asset"`
	Supply   flexNumber `json:"supply"`
	Domain   string     `json:"domain"`
	TomlInfo struct {
		Code   string `json:"code"`
		Issuer string `json:"issuer"`
		Name   string `json:"name"`
	} `json:"tomlInfo"`
}

// ExpertResponse represents the API response from stellar.expert
type ExpertResponse struct {
	Embedded struct {
		Records []expertAssetRecord `json:"records"`
	} `json:"_embedded"`
}

// SearchAssetsByDomain queries stellar.expert for assets by domain
func SearchAssetsByDomain(domain string) ([]models.StellarExpertAsset, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	url := fmt.Sprintf("https://api.stellar.expert/explorer/public/asset?search=%s", domain)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query stellar.expert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stellar.expert returned status %d", resp.StatusCode)
	}

	var result ExpertResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to our model format
	assets := make([]models.StellarExpertAsset, 0, len(result.Embedded.Records))
	for _, record := range result.Embedded.Records {
		// Convert supply to string
		supplyStr := record.Supply.String()
		
		asset := models.StellarExpertAsset{
			Code:       record.TomlInfo.Code,
			Issuer:     record.TomlInfo.Issuer,
			Domain:     record.Domain,
			Supply:     supplyStr,
			Trustlines: 0, // Not easily available in this format
			Name:       record.TomlInfo.Name,
		}
		
		// Skip if no code or issuer (invalid asset)
		if asset.Code == "" || asset.Issuer == "" {
			// Try parsing from asset string as fallback
			parts := strings.Split(record.Asset, "-")
			if len(parts) >= 2 {
				asset.Code = parts[0]
				asset.Issuer = parts[1]
			} else {
				continue // Skip invalid assets
			}
		}
		
		assets = append(assets, asset)
	}

	return assets, nil
}
