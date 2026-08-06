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
		if pairMatches(existing, base, quote) {
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

// RemoveCustomPair deletes the first configured pair whose assets match the
// given ones, in either orientation. Issuers are compared so a pair is never
// removed just because another issuer uses the same asset code.
func RemoveCustomPair(assetA, assetB txnbuild.Asset) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	base := AssetToString(assetA)
	quote := AssetToString(assetB)
	for i, existing := range cfg.Pairs {
		if !pairMatches(existing, base, quote) {
			continue
		}
		cfg.Pairs = append(cfg.Pairs[:i], cfg.Pairs[i+1:]...)
		return SaveConfig(cfg)
	}

	return fmt.Errorf("pair not found in %s", GetConfigPath())
}

// ListCustomPairs returns the configured pairs in file order.
func ListCustomPairs() ([]Pair, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return cfg.Pairs, nil
}

// pairMatches reports whether a configured pair refers to the same two assets
// as the given canonical asset strings, in either orientation. Both sides are
// normalised first so the equivalent spellings of XLM ("native" and
// "XLM:native") compare equal.
func pairMatches(pair Pair, base, quote string) bool {
	existingBase, err := ParseAsset(pair.Base)
	if err != nil {
		return false
	}
	existingQuote, err := ParseAsset(pair.Quote)
	if err != nil {
		return false
	}

	left := AssetString(existingBase)
	right := AssetString(existingQuote)
	return (left == base && right == quote) || (left == quote && right == base)
}

// AssetToString converts an asset to the canonical config and Horizon format.
func AssetToString(asset txnbuild.Asset) string {
	return AssetString(asset)
}

// StringToAsset converts a canonical config string into a Stellar asset.
func StringToAsset(value string) (txnbuild.Asset, error) {
	return ParseAsset(value)
}
