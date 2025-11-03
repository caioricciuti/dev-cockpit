package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	githubRepo    = "caioricciuti/dev-cockpit"
	githubAPIURL  = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	binaryName    = "devcockpit-darwin-arm64"
	checksumName  = "devcockpit-darwin-arm64.sha256"
)

// FetchLatestRelease queries GitHub API for the latest release
func FetchLatestRelease() (*Release, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(githubAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: HTTP %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release data: %w", err)
	}

	// Validate release has required assets
	if release.TagName == "" {
		return nil, fmt.Errorf("invalid release: missing tag name")
	}

	if release.FindAsset(binaryName) == nil {
		return nil, fmt.Errorf("release missing binary asset: %s", binaryName)
	}

	if release.FindAsset(checksumName) == nil {
		return nil, fmt.Errorf("release missing checksum asset: %s", checksumName)
	}

	return &release, nil
}

// HasUpdate compares current version with latest and returns true if update is available
func HasUpdate(current, latest string) (bool, error) {
	// Normalize versions (ensure "v" prefix for semver)
	if !strings.HasPrefix(current, "v") {
		current = "v" + current
	}
	if !strings.HasPrefix(latest, "v") {
		latest = "v" + latest
	}

	// Handle dev builds - always allow update
	if current == "vdev" {
		return true, nil
	}

	// Validate versions
	if !semver.IsValid(current) {
		return false, fmt.Errorf("invalid current version: %s", current)
	}
	if !semver.IsValid(latest) {
		return false, fmt.Errorf("invalid latest version: %s", latest)
	}

	// Compare: -1 means current < latest (update available)
	return semver.Compare(current, latest) < 0, nil
}
