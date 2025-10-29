package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	githubAPIURL = "https://api.github.com/repos/sdexmon/sdexmon/releases/latest"
	checkTimeout = 5 * time.Second
)

// GitHubRelease represents the GitHub API response for latest release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Name    string `json:"name"`
}

// CompareVersions returns true if remote is newer than local
// Assumes semantic versioning (e.g., v0.1.2)
func CompareVersions(local, remote string) bool {
	// Remove 'v' prefix if present
	local = strings.TrimPrefix(local, "v")
	remote = strings.TrimPrefix(remote, "v")

	// Simple string comparison works for semantic versioning
	// e.g., "0.1.0" < "0.1.1" < "0.1.2" < "0.2.0"
	return remote > local
}

// FetchLatestVersion fetches the latest release version from GitHub
func FetchLatestVersion() (string, string, error) {
	client := &http.Client{
		Timeout: checkTimeout,
	}

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to avoid GitHub API rate limiting
	req.Header.Set("User-Agent", "sdexmon-version-checker")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	return release.TagName, release.HTMLURL, nil
}

// CheckForUpdate checks if an update is available
func CheckForUpdate(currentVersion string) (updateAvailable bool, latestVersion, downloadURL string, err error) {
	latestVersion, downloadURL, err = FetchLatestVersion()
	if err != nil {
		return false, "", "", err
	}

	updateAvailable = CompareVersions(currentVersion, latestVersion)
	return updateAvailable, latestVersion, downloadURL, nil
}
