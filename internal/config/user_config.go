package config

import (
	"fmt"

	"github.com/stellar/go/txnbuild"
)

// AddCustomPair adds a pair to the application's single configuration schema.
// Asset issuers are retained so equal asset codes from different issuers remain
// distinct.
func AddCustomPair(assetA, assetB txnbuild.Asset) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	base := AssetToString(assetA)
	quote := AssetToString(assetB)
	for _, existing := range cfg.Pairs {
		if existing.Base == base && existing.Quote == quote {
			return fmt.Errorf("pair already exists")
		}
	}

	cfg.Pairs = append(cfg.Pairs, Pair{
		Name:         fmt.Sprintf("%s/%s", assetA.GetCode(), assetB.GetCode()),
		Base:         base,
		Quote:        quote,
		Favorite:     false,
		ShowDecimals: 7,
	})
	return SaveConfig(cfg)
}

// AssetToString converts an asset to the canonical config and Horizon format.
func AssetToString(asset txnbuild.Asset) string {
	return AssetString(asset)
}

// StringToAsset converts a canonical config string into a Stellar asset.
func StringToAsset(value string) (txnbuild.Asset, error) {
	return ParseAsset(value)
}
