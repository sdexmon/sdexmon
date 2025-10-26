package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/txnbuild"
	"gopkg.in/yaml.v3"
)

// HorizonURL returns the Horizon endpoint from environment or default
func HorizonURL() string {
	if v := os.Getenv("HORIZON_URL"); v != "" {
		return v
	}
	// Default to public Stellar Horizon mainnet
	return "https://horizon.stellar.org"
}

// NewClient creates a new Horizon client
func NewClient() *horizonclient.Client {
	return &horizonclient.Client{HorizonURL: HorizonURL()}
}

// IsDebugMode returns true if debug mode is enabled via environment
func IsDebugMode() bool {
	return os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1"
}

// GetBaseAsset returns the base asset from environment if set
func GetBaseAsset() (txnbuild.Asset, error) {
	if b := os.Getenv("BASE_ASSET"); b != "" {
		return ParseAsset(b)
	}
	return nil, nil
}

// GetQuoteAsset returns the quote asset from environment if set
func GetQuoteAsset() (txnbuild.Asset, error) {
	if q := os.Getenv("QUOTE_ASSET"); q != "" {
		return ParseAsset(q)
	}
	return nil, nil
}

// GetLPPoolID returns the LP pool ID override from environment if set
func GetLPPoolID() string {
	return os.Getenv("LP_POOL_ID")
}

// Custom log writer to capture logs in memory
var debugLogBuffer []string
var debugLogMutex sync.Mutex

type debugLogWriter struct{}

func (w debugLogWriter) Write(p []byte) (n int, err error) {
	debugLogMutex.Lock()
	defer debugLogMutex.Unlock()
	line := string(p)
	if line != "" {
		debugLogBuffer = append(debugLogBuffer, line)
		if len(debugLogBuffer) > 100 {
			debugLogBuffer = debugLogBuffer[len(debugLogBuffer)-100:]
		}
	}
	return len(p), nil
}

// SetupDebugLogger configures logging for debug mode
func SetupDebugLogger() {
	// Only write to debug buffer, not to stderr to keep TUI clean
	log.SetOutput(debugLogWriter{})
}

// GetDebugLogs returns the current debug log buffer
func GetDebugLogs() []string {
	debugLogMutex.Lock()
	defer debugLogMutex.Unlock()
	result := make([]string, len(debugLogBuffer))
	copy(result, debugLogBuffer)
	return result
}

// YAML Configuration Support

// Config represents the YAML configuration structure
type Config struct {
	App struct {
		Version     string `yaml:"version"`
		DefaultPair string `yaml:"default_pair"`
	} `yaml:"app"`
	
	Pairs []Pair `yaml:"pairs"`
	
	Assets []Asset `yaml:"assets"`
	
	Preferences struct {
		DefaultOrderBookDepth int  `yaml:"default_order_book_depth"`
		DefaultLiquidityPools int  `yaml:"default_liquidity_pools"`
		AutoRefresh           bool `yaml:"auto_refresh"`
		RefreshIntervalMs     int  `yaml:"refresh_interval_ms"`
		ShowDebug             bool `yaml:"show_debug"`
	} `yaml:"preferences"`
	
	SystemSettings struct {
		TerminalSize struct {
			Width  int `yaml:"width"`
			Height int `yaml:"height"`
		} `yaml:"terminal_size"`
	} `yaml:"system_settings"`
}

// Pair represents a trading pair in the configuration
type Pair struct {
	Name         string `yaml:"name"`
	Base         string `yaml:"base"`
	Quote        string `yaml:"quote"`
	LP           string `yaml:"lp"`
	Favorite     bool   `yaml:"favorite"`
	ShowDecimals int    `yaml:"show_decimals"`
}

// Asset represents an asset with its decimal configuration
type Asset struct {
	Name         string `yaml:"name"`
	ShowDecimals int    `yaml:"show_decimals"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "sdexmon", "config.yaml")
}

// LoadConfig loads the configuration from YAML file
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()
	
	// If config doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}
	
	return &config, nil
}

// SaveConfig saves the configuration to YAML file
func SaveConfig(config *Config) error {
	configPath := GetConfigPath()
	
	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// AddPair adds a new pair to the config and saves it
func AddPair(config *Config, name, base, quote, lpID string) error {
	// Determine show_decimals based on base and quote assets
	showDecimals := 2 // default
	
	// Check if either base or quote contains BTCZ (should be 0 decimals)
	if strings.Contains(base, "BTCZ") || strings.Contains(quote, "BTCZ") {
		showDecimals = 0
	}
	// Check if either base or quote contains XAUZ (should be 7 decimals)
	if strings.Contains(base, "XAUZ") || strings.Contains(quote, "XAUZ") {
		showDecimals = 7
	}
	
	newPair := Pair{
		Name:         name,
		Base:         base,
		Quote:        quote,
		LP:           lpID,
		Favorite:     false,
		ShowDecimals: showDecimals,
	}
	
	config.Pairs = append(config.Pairs, newPair)
	return SaveConfig(config)
}

// GetPairDecimals returns the decimal configuration for a trading pair
// Returns [baseDecimals, quoteDecimals] by finding the matching pair or asset in config
func (c *Config) GetPairDecimals(baseName, quoteName string) (int, int) {
	// Look for exact pair match first (if pairs have show_decimals)
	for _, pair := range c.Pairs {
		pairBaseName := parseAssetCode(pair.Base)
		pairQuoteName := parseAssetCode(pair.Quote)
		if (pairBaseName == baseName && pairQuoteName == quoteName) ||
		   (pairBaseName == quoteName && pairQuoteName == baseName) {
			if pair.ShowDecimals > 0 {
				return pair.ShowDecimals, pair.ShowDecimals
			}
		}
	}
	
	// Look up individual assets in the Assets array
	baseDecimals := c.GetAssetDecimals(baseName)
	quoteDecimals := c.GetAssetDecimals(quoteName)
	return baseDecimals, quoteDecimals
}

// GetAssetDecimals returns the decimal configuration for an asset
func (c *Config) GetAssetDecimals(assetName string) int {
	// First try to find by asset name (e.g., "USDC:GA5ZS...")
	for _, asset := range c.Assets {
		if asset.Name == assetName {
			return asset.ShowDecimals
		}
	}
	
	// Try to match by asset code (e.g., "USDC")
	assetCode := parseAssetCode(assetName)
	for _, asset := range c.Assets {
		if parseAssetCode(asset.Name) == assetCode {
			return asset.ShowDecimals
		}
	}
	
	// Fall back to default based on asset characteristics
	return getAssetDefaultDecimals(assetCode)
}

// parseAssetCode extracts the asset code from YAML asset string (e.g., "USDC:GA5ZS..." -> "USDC")
func parseAssetCode(assetString string) string {
	if assetString == "XLM:native" {
		return "XLM"
	}
	parts := strings.SplitN(assetString, ":", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return assetString
}

// getAssetDefaultDecimals returns default decimals for an asset based on its characteristics
func getAssetDefaultDecimals(assetCode string) int {
	switch assetCode {
	case "BTCZ":
		return 0 // Bitcoin - whole numbers
	case "XAUZ":
		return 7 // Gold - maximum precision
	default:
		return 2 // Standard currency display
	}
}

// getDefaultConfig returns a default configuration when no config file exists
func getDefaultConfig() *Config {
	return &Config{
		App: struct {
			Version     string `yaml:"version"`
			DefaultPair string `yaml:"default_pair"`
		}{
			Version:     "0.1.0",
			DefaultPair: "USDC/USDZ",
		},
		Pairs: []Pair{}, // Empty, will be populated from curated pairs if needed
		Preferences: struct {
			DefaultOrderBookDepth int  `yaml:"default_order_book_depth"`
			DefaultLiquidityPools int  `yaml:"default_liquidity_pools"`
			AutoRefresh           bool `yaml:"auto_refresh"`
			RefreshIntervalMs     int  `yaml:"refresh_interval_ms"`
			ShowDebug             bool `yaml:"show_debug"`
		}{
			DefaultOrderBookDepth: 7,
			DefaultLiquidityPools: 10,
			AutoRefresh:           true,
			RefreshIntervalMs:     1500,
			ShowDebug:             false,
		},
		SystemSettings: struct {
			TerminalSize struct {
				Width  int `yaml:"width"`
				Height int `yaml:"height"`
			} `yaml:"terminal_size"`
		}{
			TerminalSize: struct {
				Width  int `yaml:"width"`
				Height int `yaml:"height"`
			}{
				Width:  140,
				Height: 60,
			},
		},
	}
}
