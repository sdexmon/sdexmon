package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	localParts, localPre, localOK := parseVersion(local)
	remoteParts, remotePre, remoteOK := parseVersion(remote)
	if !localOK || !remoteOK {
		return false
	}
	for i := range localParts {
		if remoteParts[i] != localParts[i] {
			return remoteParts[i] > localParts[i]
		}
	}
	// A stable release is newer than a prerelease of the same version.
	if localPre != remotePre {
		if localPre == "" {
			return false
		}
		if remotePre == "" {
			return true
		}
		return comparePrerelease(remotePre, localPre) > 0
	}
	return false
}

func parseVersion(value string) ([3]int, string, bool) {
	var result [3]int
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	core, prerelease, _ := strings.Cut(value, "-")
	core, _, _ = strings.Cut(core, "+")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return result, "", false
	}
	for i, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return result, "", false
		}
		result[i] = number
	}
	return result, prerelease, true
}

func comparePrerelease(left, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	for i := 0; i < len(leftParts) && i < len(rightParts); i++ {
		if leftParts[i] == rightParts[i] {
			continue
		}
		leftNumber, leftErr := strconv.Atoi(leftParts[i])
		rightNumber, rightErr := strconv.Atoi(rightParts[i])
		switch {
		case leftErr == nil && rightErr == nil:
			if leftNumber < rightNumber {
				return -1
			}
			return 1
		case leftErr == nil:
			return -1
		case rightErr == nil:
			return 1
		case leftParts[i] < rightParts[i]:
			return -1
		default:
			return 1
		}
	}
	switch {
	case len(leftParts) < len(rightParts):
		return -1
	case len(leftParts) > len(rightParts):
		return 1
	default:
		return 0
	}
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
