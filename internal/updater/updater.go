package updater

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/go-selfupdate"
)

const (
	// GitHub repository for updates
	githubRepo = "phravins/devcli"
	// Current version of DevCLI
	currentVersion = "1.0.0"
)

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	CurrentVersion    string
	LatestVersion     string
	IsUpdateAvailable bool
	ReleaseURL        string
	ReleaseNotes      string
}

// CheckForUpdates checks if a new version is available on GitHub
func CheckForUpdates() (*UpdateInfo, error) {
	latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug(githubRepo))
	if err != nil {
		return nil, fmt.Errorf("error checking for updates: %w", err)
	}

	if !found {
		return nil, fmt.Errorf("no releases found")
	}

	currentVer, err := semver.NewVersion(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid current version: %w", err)
	}

	latestVer, err := semver.NewVersion(latest.Version())
	if err != nil {
		return nil, fmt.Errorf("invalid latest version: %w", err)
	}

	info := &UpdateInfo{
		CurrentVersion:    currentVersion,
		LatestVersion:     latest.Version(),
		IsUpdateAvailable: latestVer.GreaterThan(currentVer),
		ReleaseURL:        latest.URL,
		ReleaseNotes:      latest.ReleaseNotes,
	}

	return info, nil
}

// PerformUpdate downloads and installs the latest version
func PerformUpdate() error {
	latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug(githubRepo))
	if err != nil {
		return fmt.Errorf("error detecting latest version: %w", err)
	}

	if !found {
		return fmt.Errorf("no releases found")
	}

	currentVer, err := semver.NewVersion(currentVersion)
	if err != nil {
		return fmt.Errorf("invalid current version: %w", err)
	}

	latestVer, err := semver.NewVersion(latest.Version())
	if err != nil {
		return fmt.Errorf("invalid latest version: %w", err)
	}

	if !latestVer.GreaterThan(currentVer) {
		return fmt.Errorf("already running the latest version (%s)", currentVersion)
	}

	// Configure the updater
	config := selfupdate.Config{
		Validator: &selfupdate.ChecksumValidator{
			UniqueFilename: getAssetName(),
		},
	}

	updater, err := selfupdate.NewUpdater(config)
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	// Perform the update with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := updater.UpdateTo(ctx, latest, latest.AssetURL); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

// getAssetName returns the expected asset name for the current platform
func getAssetName() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	switch osName {
	case "windows":
		return fmt.Sprintf("devcli-windows-%s.exe", arch)
	case "darwin":
		return fmt.Sprintf("devcli-darwin-%s", arch)
	case "linux":
		return fmt.Sprintf("devcli-linux-%s", arch)
	default:
		return fmt.Sprintf("devcli-%s-%s", osName, arch)
	}
}

// GetCurrentVersion returns the current version of DevCLI
func GetCurrentVersion() string {
	return currentVersion
}
