package config

import (
	"log"
	"os"
	"sync"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/txnbuild"
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
