package config

import (
	"fmt"
	"strings"

	"github.com/stellar/go/txnbuild"
)

// ParseAsset parses an asset string (native, XLM, or CODE:ISSUER)
func ParseAsset(s string) (txnbuild.Asset, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "native") || strings.EqualFold(s, "XLM") && !strings.Contains(s, ":") {
		return txnbuild.NativeAsset{}, nil
	}
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("expected CODE:ISSUER or 'native'")
	}
	code := strings.ToUpper(strings.TrimSpace(parts[0]))
	issuer := strings.TrimSpace(parts[1])
	if code == "XLM" && issuer == "" {
		return txnbuild.NativeAsset{}, nil
	}
	if code == "" || issuer == "" {
		return nil, fmt.Errorf("invalid asset spec")
	}
	return txnbuild.CreditAsset{Code: code, Issuer: issuer}, nil
}

// AssetShort returns the short name for an asset
func AssetShort(a txnbuild.Asset) string {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		return "XLM"
	case txnbuild.CreditAsset:
		return v.Code
	default:
		return "?"
	}
}

// AssetString returns the full string representation of an asset
func AssetString(a txnbuild.Asset) string {
	switch v := a.(type) {
	case txnbuild.NativeAsset:
		return "native"
	case txnbuild.CreditAsset:
		return fmt.Sprintf("%s:%s", v.Code, v.Issuer)
	default:
		return ""
	}
}
