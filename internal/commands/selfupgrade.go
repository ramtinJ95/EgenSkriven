package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	repoOwner = "ramtinJ95"
	repoName  = "EgenSkriven"
)

// GitHubRelease represents a GitHub release API response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// newSelfUpgradeCmd creates the self-upgrade command
func newSelfUpgradeCmd() *cobra.Command {
	var (
		checkOnly bool
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "self-upgrade",
		Short: "Upgrade EgenSkriven to the latest version",
		Long: `Check for updates and optionally upgrade EgenSkriven to the latest version.

By default, this command will:
1. Check GitHub for the latest release
2. Compare with your current version
3. Download and install the new version if available

Use --check to only check for updates without installing.
Use --force to reinstall even if already on the latest version.

Alternative: You can also upgrade by re-running the install script:
  curl -fsSL https://raw.githubusercontent.com/ramtinJ95/EgenSkriven/main/install.sh | sh
`,
		Example: `  # Check for updates and install if available
  egenskriven self-upgrade

  # Only check for updates (don't install)
  egenskriven self-upgrade --check

  # Force reinstall current version
  egenskriven self-upgrade --force

  # JSON output for scripting
  egenskriven self-upgrade --check --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSelfUpgrade(checkOnly, force)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force upgrade even if already on latest version")

	return cmd
}

func runSelfUpgrade(checkOnly, force bool) error {
	formatter := getFormatter()

	// 1. Fetch latest release info from GitHub
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(Version, "v")

	// 2. Compare versions using semver comparison
	// Only consider update available if latest is actually newer (not equal or older)
	updateAvailable := currentVersion != "dev" && isNewerVersion(latestVersion, currentVersion)

	// Handle JSON output
	if formatter.JSON {
		result := map[string]interface{}{
			"current_version":  Version,
			"latest_version":   release.TagName,
			"update_available": updateAvailable,
		}
		formatter.WriteJSON(result)
		return nil
	}

	// 3. Report status
	if !updateAvailable && !force {
		formatter.Success(fmt.Sprintf("You're already on the latest version (%s)", Version))
		return nil
	}

	if updateAvailable {
		fmt.Printf("Update available: %s -> %s\n", Version, release.TagName)
	} else if force {
		fmt.Printf("Forcing reinstall of version %s\n", release.TagName)
	}

	// 4. If check-only, stop here
	if checkOnly {
		if updateAvailable {
			fmt.Println("\nRun 'egenskriven self-upgrade' to install the update.")
		}
		return nil
	}

	// 5. Find the right asset for this platform
	assetName := fmt.Sprintf("egenskriven-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, release.TagName)
	}

	// 6. Download new binary
	fmt.Printf("Downloading %s...\n", assetName)
	tmpFile, err := downloadBinary(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(tmpFile) // Clean up temp file (no-op if already moved)

	// 7. Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// 8. Replace the binary
	fmt.Printf("Installing to %s...\n", execPath)
	if err := replaceBinary(tmpFile, execPath); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	formatter.Success(fmt.Sprintf("Successfully upgraded to %s", release.TagName))
	fmt.Println("\nRestart any running 'egenskriven serve' instances to use the new version.")

	return nil
}

// getLatestRelease fetches the latest release info from GitHub API
func getLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadBinary downloads a file to a temporary location
func downloadBinary(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "egenskriven-upgrade-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Copy downloaded content
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// replaceBinary replaces the current binary with the new one
func replaceBinary(newBinary, targetPath string) error {
	// On both Windows and Linux, we can't overwrite a running binary directly.
	// However, we CAN rename a running binary, then place the new one at the original path.
	// This works because the OS keeps the inode open for the running process.
	oldPath := targetPath + ".old"
	os.Remove(oldPath) // Remove any existing .old file

	// Rename current binary to .old (this works even while running)
	if err := os.Rename(targetPath, oldPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary to target location
	// First try rename (fastest, atomic)
	if err := os.Rename(newBinary, targetPath); err != nil {
		// If rename fails (cross-device), fall back to copy
		if err := copyFile(newBinary, targetPath); err != nil {
			// Try to restore the old binary on failure
			os.Rename(oldPath, targetPath)
			return err
		}
		os.Remove(newBinary)
	}

	// Clean up .old file (optional - leave it on Windows for safety)
	if runtime.GOOS != "windows" {
		os.Remove(oldPath)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		// Clean up partial file on copy failure
		os.Remove(dst)
		return err
	}

	// Preserve executable permission
	if err := os.Chmod(dst, 0755); err != nil {
		// Clean up on chmod failure
		os.Remove(dst)
		return err
	}
	return nil
}

// isNewerVersion returns true if version a is newer than version b.
// Versions should be in semver format (e.g., "1.2.3" or "1.2.3-beta").
// Pre-release versions (containing "-") are considered older than release versions.
func isNewerVersion(a, b string) bool {
	// Parse version parts, stripping any pre-release suffix for comparison
	aParts, aPrerelease := parseVersion(a)
	bParts, bPrerelease := parseVersion(b)

	// Compare major.minor.patch
	for i := 0; i < 3; i++ {
		aVal, bVal := 0, 0
		if i < len(aParts) {
			aVal = aParts[i]
		}
		if i < len(bParts) {
			bVal = bParts[i]
		}

		if aVal > bVal {
			return true
		}
		if aVal < bVal {
			return false
		}
	}

	// Same version numbers - check pre-release
	// A release version is newer than a pre-release of the same version
	if bPrerelease != "" && aPrerelease == "" {
		return true
	}

	return false
}

// parseVersion parses a semver string into numeric parts and pre-release suffix.
// Returns ([]int{major, minor, patch}, prerelease)
func parseVersion(v string) ([]int, string) {
	prerelease := ""
	if idx := strings.Index(v, "-"); idx != -1 {
		prerelease = v[idx+1:]
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	result := make([]int, len(parts))
	for i, p := range parts {
		val, err := strconv.Atoi(p)
		if err != nil {
			val = 0
		}
		result[i] = val
	}
	return result, prerelease
}
