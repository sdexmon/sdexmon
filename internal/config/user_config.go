package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stellar/go/txnbuild"
	"gopkg.in/yaml.v3"
)

// UserConfig holds user-specific configuration
type UserConfig struct {
	CustomPairs []CustomPair `yaml:"custom_pairs"`
}

// CustomPair represents a user-added trading pair
type CustomPair struct {
	AssetA string `yaml:"asset_a"` // "CODE:ISSUER" or "native"
	AssetB string `yaml:"asset_b"` // "CODE:ISSUER" or "native"
	Label  string `yaml:"label,omitempty"`
}

// GetUserConfigPath returns the path to the user config file
func GetUserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "sdexmon")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// LoadUserConfig loads the user configuration from disk
func LoadUserConfig() (*UserConfig, error) {
	path, err := GetUserConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &UserConfig{CustomPairs: []CustomPair{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg UserConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveUserConfig saves the user configuration to disk
func SaveUserConfig(cfg *UserConfig) error {
	path, err := GetUserConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddCustomPair adds a new custom pair to the user config
func AddCustomPair(assetA, assetB txnbuild.Asset) error {
	cfg, err := LoadUserConfig()
	if err != nil {
		return err
	}

	pair := CustomPair{
		AssetA: AssetToString(assetA),
		AssetB: AssetToString(assetB),
		Label:  fmt.Sprintf("%s/%s", assetA.GetCode(), assetB.GetCode()),
	}

	// Check for duplicates
	for _, existing := range cfg.CustomPairs {
		if existing.AssetA == pair.AssetA && existing.AssetB == pair.AssetB {
			return fmt.Errorf("pair already exists")
		}
	}

	cfg.CustomPairs = append(cfg.CustomPairs, pair)
	return SaveUserConfig(cfg)
}

// AssetToString converts a txnbuild.Asset to string format
func AssetToString(asset txnbuild.Asset) string {
	if asset.IsNative() {
		return "native"
	}
	return fmt.Sprintf("%s:%s", asset.GetCode(), asset.GetIssuer())
}

// StringToAsset converts a string to txnbuild.Asset
func StringToAsset(s string) (txnbuild.Asset, error) {
	if s == "native" {
		return txnbuild.NativeAsset{}, nil
	}

	// Parse "CODE:ISSUER" format
	parts := splitAssetString(s)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid asset format: %s (expected CODE:ISSUER or native)", s)
	}

	return txnbuild.CreditAsset{
		Code:   parts[0],
		Issuer: parts[1],
	}, nil
}

// splitAssetString splits "CODE:ISSUER" format
func splitAssetString(s string) []string {
	for i, c := range s {
		if c == ':' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
