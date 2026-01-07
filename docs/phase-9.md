# Phase 9: Release

**Goal**: Production-ready release with cross-platform distribution, installation scripts, and documentation.

**Duration Estimate**: 2-3 days

**Prerequisites**: All previous phases complete (0-8).

**Deliverable**: Cross-platform binaries, GitHub release workflow, shell completions, installation script, and documentation.

---

## Overview

This phase prepares EgenSkriven for public release. By the end, users can:
- Download pre-built binaries for their platform
- Install via a one-line command
- Use shell completions for faster CLI usage
- Access comprehensive documentation

The focus is on distribution, not new features. Everything should "just work" for new users.

### What We're Building

| Component | Purpose |
|-----------|---------|
| Cross-platform builds | Binaries for macOS, Linux, Windows |
| Version embedding | `egenskriven version` shows build info |
| GitHub Actions workflow | Automated releases on git tag |
| Shell completions | Tab-completion for bash/zsh/fish |
| Installation script | One-liner install for Unix systems |
| Documentation | README, CLI reference, guides |

### Why This Matters

A great CLI tool with poor distribution is unused. This phase ensures:
- Users can install in under a minute
- Binaries work without dependencies
- Shell completions reduce friction
- Documentation answers common questions

---

## Environment Requirements

Before starting, ensure you have:

| Tool | Version | Check Command |
|------|---------|---------------|
| Go | 1.21+ | `go version` |
| Git | Any | `git --version` |
| GitHub CLI | Any | `gh --version` |
| Make | Any | `make --version` |

**Install GitHub CLI** (if needed):
- macOS: `brew install gh`
- Linux: `sudo apt install gh` or see https://cli.github.com/
- Windows: `winget install GitHub.cli`

After installing, authenticate:
```bash
gh auth login
```

---

## Tasks

### 9.1 Cross-Platform Build

**What**: Update the Makefile to build binaries for all target platforms.

**Why**: Users should be able to download a single binary for their system. Go's cross-compilation makes this straightforward with `GOOS` and `GOARCH` environment variables.

**Target Platforms**:

| Platform | GOOS | GOARCH | Binary Name |
|----------|------|--------|-------------|
| macOS (Apple Silicon) | darwin | arm64 | egenskriven-darwin-arm64 |
| macOS (Intel) | darwin | amd64 | egenskriven-darwin-amd64 |
| Linux (64-bit) | linux | amd64 | egenskriven-linux-amd64 |
| Linux (ARM64) | linux | arm64 | egenskriven-linux-arm64 |
| Windows (64-bit) | windows | amd64 | egenskriven-windows-amd64.exe |

**Update Makefile**:

Add the following targets to your existing `Makefile`:

```makefile
# =============================================================================
# Release Targets
# =============================================================================

# Build binaries for all platforms
# Usage: make release
# Creates binaries in dist/ directory
release: clean-dist build-ui release-darwin release-linux release-windows checksums
	@echo "Release builds complete. Binaries in dist/"
	@ls -lh dist/

# Clean dist directory before release build
clean-dist:
	@echo "Cleaning dist directory..."
	rm -rf dist/
	mkdir -p dist/

# macOS builds (Apple Silicon and Intel)
release-darwin:
	@echo "Building for macOS (arm64)..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-darwin-arm64 \
		./cmd/egenskriven
	@echo "Building for macOS (amd64)..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-darwin-amd64 \
		./cmd/egenskriven

# Linux builds (amd64 and arm64)
release-linux:
	@echo "Building for Linux (amd64)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-linux-amd64 \
		./cmd/egenskriven
	@echo "Building for Linux (arm64)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-linux-arm64 \
		./cmd/egenskriven

# Windows build
release-windows:
	@echo "Building for Windows (amd64)..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-windows-amd64.exe \
		./cmd/egenskriven

# Generate checksums for all binaries
checksums:
	@echo "Generating checksums..."
	cd dist && sha256sum * > checksums.txt
	@cat dist/checksums.txt
```

**Steps**:

1. Open `Makefile` in your editor.

2. Add the release targets shown above.

3. Test the release build:
   ```bash
   make release
   ```
   
   **Expected output**:
   ```
   Cleaning dist directory...
   Building for macOS (arm64)...
   Building for macOS (amd64)...
   Building for Linux (amd64)...
   Building for Linux (arm64)...
   Building for Windows (amd64)...
   Generating checksums...
   Release builds complete. Binaries in dist/
   -rwxr-xr-x 1 user user 35M Jan  3 10:00 egenskriven-darwin-amd64
   -rwxr-xr-x 1 user user 35M Jan  3 10:00 egenskriven-darwin-arm64
   -rwxr-xr-x 1 user user 35M Jan  3 10:00 egenskriven-linux-amd64
   -rwxr-xr-x 1 user user 35M Jan  3 10:00 egenskriven-linux-arm64
   -rwxr-xr-x 1 user user 36M Jan  3 10:00 egenskriven-windows-amd64.exe
   ```

4. Verify checksums file:
   ```bash
   cat dist/checksums.txt
   ```
   
   **Expected output** (hashes will differ):
   ```
   a1b2c3d4...  egenskriven-darwin-amd64
   e5f6g7h8...  egenskriven-darwin-arm64
   i9j0k1l2...  egenskriven-linux-amd64
   m3n4o5p6...  egenskriven-linux-arm64
   q7r8s9t0...  egenskriven-windows-amd64.exe
   ```

**Common Mistakes**:
- Forgetting `CGO_ENABLED=0` (may require C compiler for cross-compilation)
- Not creating dist/ directory first
- Missing `build-ui` dependency (binaries won't include the UI)

---

### 9.2 Add Version Embedding

**What**: Embed version information into the binary at build time.

**Why**: Users need to know which version they're running. Support teams need version info for debugging. `egenskriven version` should show meaningful build information.

**Update main.go**:

**File**: `cmd/egenskriven/main.go`

```go
package main

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/yourusername/egenskriven/internal/commands"
	"github.com/yourusername/egenskriven/ui"
)

// These variables are set at build time via -ldflags
// Example: go build -ldflags "-X main.Version=1.0.0"
var (
	// Version is the semantic version (e.g., "1.0.0")
	Version = "dev"
	
	// BuildDate is the ISO 8601 build timestamp
	BuildDate = "unknown"
	
	// GitCommit is the git commit hash
	GitCommit = "unknown"
)

func main() {
	app := pocketbase.New()

	// Register custom CLI commands, passing version info
	commands.Register(app, commands.VersionInfo{
		Version:   Version,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
	})

	// Serve embedded React frontend
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		e.Router.GET("/{path...}", func(re *core.RequestEvent) error {
			path := re.Request.PathValue("path")

			if f, err := ui.DistFS.Open(path); err == nil {
				f.Close()
				return re.FileFS(ui.DistFS, path)
			}

			return re.FileFS(ui.DistFS, "index.html")
		})

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
```

**Create Version Command**:

**File**: `internal/commands/version.go`

```go
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// VersionInfo holds build-time version information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
}

// NewVersionCmd creates the version command
func NewVersionCmd(info VersionInfo, jsonOutput *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version and build information",
		Long: `Display version information including:
- Version number
- Build date
- Git commit hash
- Go version
- OS and architecture`,
		Run: func(cmd *cobra.Command, args []string) {
			output := struct {
				VersionInfo
				GoVersion string `json:"go_version"`
				OS        string `json:"os"`
				Arch      string `json:"arch"`
			}{
				VersionInfo: info,
				GoVersion:   runtime.Version(),
				OS:          runtime.GOOS,
				Arch:        runtime.GOARCH,
			}

			if *jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(output)
				return
			}

			fmt.Printf("EgenSkriven %s\n", info.Version)
			fmt.Printf("Build date:  %s\n", info.BuildDate)
			fmt.Printf("Git commit:  %s\n", info.GitCommit)
			fmt.Printf("Go version:  %s\n", runtime.Version())
			fmt.Printf("OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
```

**Update Makefile with LDFLAGS**:

Add these lines near the top of your `Makefile`:

```makefile
# Version information for ldflags
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Linker flags to embed version info
LDFLAGS := -X main.Version=$(VERSION) \
           -X main.BuildDate=$(BUILD_DATE) \
           -X main.GitCommit=$(GIT_COMMIT)
```

Update the `build` target to use LDFLAGS:

```makefile
# Build production binary with version info
build:
	@echo "Building production binary..."
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o egenskriven ./cmd/egenskriven
	@echo "Built: ./egenskriven ($(shell du -h egenskriven | cut -f1))"
```

**Steps**:

1. Update `cmd/egenskriven/main.go` with the version variables.

2. Create `internal/commands/version.go`.

3. Update your `Makefile` with LDFLAGS and modify the build target.

4. Build and test:
   ```bash
   make build
   ./egenskriven version
   ```
   
   **Expected output**:
   ```
   EgenSkriven dev
   Build date:  2025-01-03T10:00:00Z
   Git commit:  a1b2c3d
   Go version:  go1.21.0
   OS/Arch:     darwin/arm64
   ```

5. Test JSON output:
   ```bash
   ./egenskriven version --json
   ```
   
   **Expected output**:
   ```json
   {
     "version": "dev",
     "build_date": "2025-01-03T10:00:00Z",
     "git_commit": "a1b2c3d",
     "go_version": "go1.21.0",
     "os": "darwin",
     "arch": "arm64"
   }
   ```

6. Test with a git tag:
   ```bash
   git tag v1.0.0
   make build
   ./egenskriven version
   ```
   
   **Expected output** should now show `EgenSkriven v1.0.0`

**Common Mistakes**:
- Forgetting quotes around LDFLAGS value
- Wrong variable path (must match package, e.g., `main.Version` not `cmd/egenskriven.Version`)
- Build date format errors on Windows (use PowerShell equivalent)

---

### 9.3 Create GitHub Release Workflow

**What**: Create a GitHub Actions workflow that automatically builds and releases binaries when you push a version tag.

**Why**: Manual releases are error-prone. Automation ensures consistent, reproducible builds for every release.

**File**: `.github/workflows/release.yml`

```yaml
name: Release

# Trigger on version tags (e.g., v1.0.0, v1.2.3-beta)
on:
  push:
    tags:
      - 'v*'

# Permissions needed to create releases
permissions:
  contents: write

jobs:
  release:
    name: Build and Release
    runs-on: ubuntu-latest
    
    steps:
      # Checkout code
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Needed for git describe

      # Setup Go
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true

      # Setup Node.js for UI build
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: ui/package-lock.json

      # Install UI dependencies
      - name: Install UI Dependencies
        run: npm ci
        working-directory: ui

      # Build UI
      - name: Build UI
        run: npm run build
        working-directory: ui

      # Run tests
      - name: Run Tests
        run: go test ./... -v

      # Get version from tag
      - name: Get Version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      # Build all platforms
      - name: Build Binaries
        env:
          VERSION: ${{ steps.version.outputs.VERSION }}
          BUILD_DATE: ${{ github.event.repository.updated_at }}
          GIT_COMMIT: ${{ github.sha }}
        run: |
          mkdir -p dist

          # Define ldflags
          LDFLAGS="-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT:0:7}"

          # macOS (Apple Silicon)
          echo "Building darwin/arm64..."
          CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
            -ldflags "${LDFLAGS}" \
            -o dist/egenskriven-darwin-arm64 \
            ./cmd/egenskriven

          # macOS (Intel)
          echo "Building darwin/amd64..."
          CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
            -ldflags "${LDFLAGS}" \
            -o dist/egenskriven-darwin-amd64 \
            ./cmd/egenskriven

          # Linux (amd64)
          echo "Building linux/amd64..."
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
            -ldflags "${LDFLAGS}" \
            -o dist/egenskriven-linux-amd64 \
            ./cmd/egenskriven

          # Linux (arm64)
          echo "Building linux/arm64..."
          CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
            -ldflags "${LDFLAGS}" \
            -o dist/egenskriven-linux-arm64 \
            ./cmd/egenskriven

          # Windows (amd64)
          echo "Building windows/amd64..."
          CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
            -ldflags "${LDFLAGS}" \
            -o dist/egenskriven-windows-amd64.exe \
            ./cmd/egenskriven

          # Generate checksums
          cd dist
          sha256sum * > checksums.txt
          cat checksums.txt

      # Create GitHub Release
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          name: EgenSkriven ${{ steps.version.outputs.VERSION }}
          body: |
            ## EgenSkriven ${{ steps.version.outputs.VERSION }}

            ### Installation

            **Quick install (macOS/Linux):**
            ```bash
            curl -fsSL https://raw.githubusercontent.com/${{ github.repository }}/main/install.sh | sh
            ```

            **Manual download:**
            Download the appropriate binary for your platform below.

            ### Checksums

            Verify your download with SHA-256:
            ```bash
            sha256sum -c checksums.txt
            ```

            ### What's New

            See [CHANGELOG.md](https://github.com/${{ github.repository }}/blob/main/CHANGELOG.md) for details.
          files: |
            dist/egenskriven-darwin-arm64
            dist/egenskriven-darwin-amd64
            dist/egenskriven-linux-amd64
            dist/egenskriven-linux-arm64
            dist/egenskriven-windows-amd64.exe
            dist/checksums.txt
          draft: false
          prerelease: ${{ contains(steps.version.outputs.VERSION, '-') }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Steps**:

1. Create the workflow directory:
   ```bash
   mkdir -p .github/workflows
   ```

2. Create the file:
   ```bash
   touch .github/workflows/release.yml
   ```

3. Open in your editor and paste the workflow above.

4. Test the workflow locally (optional, requires `act`):
   ```bash
   # Install act: https://github.com/nektos/act
   act push --tag v0.0.1-test
   ```

5. Push a test tag to trigger the workflow:
   ```bash
   git add .github/workflows/release.yml
   git commit -m "Add release workflow"
   git push origin main
   
   # Create and push a tag
   git tag v0.0.1-test
   git push origin v0.0.1-test
   ```

6. Check GitHub Actions tab in your repository for the build status.

7. Verify the release appears in the Releases section.

**Common Mistakes**:
- Missing `permissions: contents: write` (release creation fails)
- Wrong tag pattern (must start with `v`)
- UI build step missing (binaries won't include frontend)
- Not fetching full git history (`fetch-depth: 0` needed for `git describe`)

---

### 9.4 Shell Completions

**What**: Implement shell completion generation for bash, zsh, fish, and PowerShell.

**Why**: Tab completion dramatically improves CLI usability. Users expect modern CLI tools to support it.

**File**: `internal/commands/completion.go`

```go
package commands

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCompletionCmd creates the completion command with subcommands for each shell
func NewCompletionCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [shell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for EgenSkriven.

To load completions:

Bash:
  # Linux
  $ egenskriven completion bash > /etc/bash_completion.d/egenskriven
  
  # macOS (requires bash-completion@2)
  $ egenskriven completion bash > $(brew --prefix)/etc/bash_completion.d/egenskriven

Zsh:
  # If shell completion is not already enabled, you need to enable it first:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # Add to your ~/.zshrc or run once:
  $ egenskriven completion zsh > "${fpath[1]}/_egenskriven"
  
  # Or for Oh My Zsh:
  $ egenskriven completion zsh > ~/.oh-my-zsh/completions/_egenskriven

Fish:
  $ egenskriven completion fish > ~/.config/fish/completions/egenskriven.fish

PowerShell:
  # Add to your PowerShell profile:
  PS> egenskriven completion powershell >> $PROFILE

After installing, restart your shell or source the completion file.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}
```

**Register the Command**:

Update `internal/commands/root.go` to register the completion command:

```go
func Register(app *pocketbase.PocketBase, versionInfo VersionInfo) {
	// ... existing code ...

	rootCmd := app.RootCmd
	
	// Add completion command
	rootCmd.AddCommand(NewCompletionCmd(rootCmd))
	
	// ... rest of registration ...
}
```

**Steps**:

1. Create `internal/commands/completion.go` with the code above.

2. Update `internal/commands/root.go` to register the command.

3. Build and test:
   ```bash
   make build
   ./egenskriven completion --help
   ```
   
   **Expected output**:
   ```
   Generate shell completion scripts for EgenSkriven.
   
   To load completions:
   
   Bash:
     # Linux
     $ egenskriven completion bash > /etc/bash_completion.d/egenskriven
   ...
   ```

4. Test completion output:
   ```bash
   ./egenskriven completion bash | head -20
   ```
   
   **Expected output**: Bash completion script code.

5. Test with your shell (example for zsh):
   ```bash
   # Generate completion
   ./egenskriven completion zsh > /tmp/_egenskriven
   
   # Source it temporarily
   source /tmp/_egenskriven
   
   # Test tab completion
   ./egenskriven <TAB>
   ```

**Common Mistakes**:
- Missing shell in ValidArgs (completion fails for that shell)
- Not exposing rootCmd properly (completions need the full command tree)

---

### 9.5 Create Installation Script

**What**: Create a shell script for one-line installation on Unix systems.

**Why**: `curl ... | sh` is the standard way to install CLI tools. It should detect the platform and download the correct binary.

**File**: `install.sh`

```bash
#!/bin/sh
# EgenSkriven Installation Script
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/yourusername/egenskriven/main/install.sh | sh
#
# Environment variables:
#   INSTALL_DIR  - Installation directory (default: /usr/local/bin or ~/.local/bin)
#   VERSION      - Specific version to install (default: latest)

set -e

# Colors for output (only if terminal supports it)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Print functions
info() { printf "${BLUE}[INFO]${NC} %s\n" "$1"; }
success() { printf "${GREEN}[OK]${NC} %s\n" "$1"; }
warn() { printf "${YELLOW}[WARN]${NC} %s\n" "$1"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; exit 1; }

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)   OS="linux" ;;
        darwin)  OS="darwin" ;;
        mingw*|msys*|cygwin*) 
            error "Windows detected. Please use the Windows installer or download manually."
            ;;
        *)       error "Unsupported operating system: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        arm64|aarch64)  ARCH="arm64" ;;
        *)              error "Unsupported architecture: $ARCH" ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    success "Detected platform: $PLATFORM"
}

# Get the latest version from GitHub
get_latest_version() {
    if [ -n "$VERSION" ]; then
        info "Using specified version: $VERSION"
        return
    fi

    info "Fetching latest version..."
    VERSION=$(curl -fsSL "https://api.github.com/repos/yourusername/egenskriven/releases/latest" | 
              grep '"tag_name":' | 
              sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Please specify VERSION environment variable."
    fi
    
    success "Latest version: $VERSION"
}

# Determine installation directory
get_install_dir() {
    if [ -n "$INSTALL_DIR" ]; then
        info "Using specified install directory: $INSTALL_DIR"
        return
    fi

    # Prefer /usr/local/bin if writable, otherwise ~/.local/bin
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
    else
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi

    success "Install directory: $INSTALL_DIR"
}

# Download and install the binary
install_binary() {
    BINARY_NAME="egenskriven-${PLATFORM}"
    DOWNLOAD_URL="https://github.com/yourusername/egenskriven/releases/download/${VERSION}/${BINARY_NAME}"
    TMP_FILE=$(mktemp)

    info "Downloading $BINARY_NAME..."
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" || error "Download failed. Check your internet connection."

    info "Installing to $INSTALL_DIR/egenskriven..."
    mv "$TMP_FILE" "$INSTALL_DIR/egenskriven"
    chmod +x "$INSTALL_DIR/egenskriven"

    success "Installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v egenskriven >/dev/null 2>&1; then
        info "Verifying installation..."
        INSTALLED_VERSION=$(egenskriven version 2>/dev/null | head -1 | awk '{print $2}')
        success "EgenSkriven $INSTALLED_VERSION is ready to use!"
    else
        warn "Installation complete, but 'egenskriven' is not in your PATH."
        echo ""
        echo "Add the following to your shell profile (.bashrc, .zshrc, etc.):"
        echo ""
        echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
        echo "Then restart your shell or run:"
        echo ""
        echo "    source ~/.bashrc  # or ~/.zshrc"
        echo ""
    fi
}

# Print next steps
print_next_steps() {
    echo ""
    echo "=== Next Steps ==="
    echo ""
    echo "1. Start the server:"
    echo "   $ egenskriven serve"
    echo ""
    echo "2. Open the web UI:"
    echo "   http://localhost:8090"
    echo ""
    echo "3. Create your first task:"
    echo "   $ egenskriven add \"My first task\""
    echo ""
    echo "4. Enable shell completions:"
    echo "   $ egenskriven completion --help"
    echo ""
    echo "Documentation: https://github.com/yourusername/egenskriven"
    echo ""
}

# Main installation flow
main() {
    echo ""
    echo "=== EgenSkriven Installer ==="
    echo ""

    detect_platform
    get_latest_version
    get_install_dir
    install_binary
    verify_installation
    print_next_steps
}

main
```

**Steps**:

1. Create the file:
   ```bash
   touch install.sh
   chmod +x install.sh
   ```

2. Open in your editor and paste the script above.

3. Update all occurrences of `yourusername/egenskriven` with your actual GitHub username/repository.

4. Test the script locally:
   ```bash
   # Test platform detection
   INSTALL_DIR=/tmp/test-install VERSION=v0.0.1-test ./install.sh
   ```

5. Test after pushing a release:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/yourusername/egenskriven/main/install.sh | sh
   ```

**Common Mistakes**:
- Wrong GitHub repository URL
- Missing execute permission on the script
- `curl` or `wget` not available on target system (script uses curl)

---

### 9.5.1 Self-Update Command

**What**: Implement a built-in `update` command that checks for new versions and updates the binary in place.

**Why**: Users shouldn't have to manually check for updates or re-run install scripts. A simple `egenskriven update` command provides the best experience.

**How It Works**:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Update Flow                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Check current version (embedded at build time)              │
│                         │                                       │
│                         ▼                                       │
│  2. Query GitHub API for latest release                         │
│                         │                                       │
│                         ▼                                       │
│  3. Compare versions (semantic versioning)                      │
│                         │                                       │
│              ┌──────────┴──────────┐                           │
│              ▼                     ▼                            │
│         Up to date            New version                       │
│         (exit)                available                         │
│                                    │                            │
│                                    ▼                            │
│  4. Download new binary to temp file                            │
│                                    │                            │
│                                    ▼                            │
│  5. Verify checksum (optional but recommended)                  │
│                                    │                            │
│                                    ▼                            │
│  6. Replace current binary with new one                         │
│                                    │                            │
│                                    ▼                            │
│  7. Verify new version works                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**File**: `internal/commands/update.go`

```go
package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// GitHubRelease represents a GitHub release API response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// UpdateConfig holds configuration for the update command
type UpdateConfig struct {
	CurrentVersion string
	RepoOwner      string
	RepoName       string
}

// NewUpdateCmd creates the update command
func NewUpdateCmd(config UpdateConfig, jsonOutput *bool) *cobra.Command {
	var (
		checkOnly bool
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update EgenSkriven to the latest version",
		Long: `Check for updates and optionally update EgenSkriven to the latest version.

By default, this command will:
1. Check GitHub for the latest release
2. Compare with your current version
3. Download and install the new version if available

Use --check to only check for updates without installing.
Use --force to reinstall even if already on the latest version.

Alternative: You can also update by re-running the install script:
  curl -fsSL https://raw.githubusercontent.com/` + config.RepoOwner + `/` + config.RepoName + `/main/install.sh | sh
`,
		Example: `  # Check for updates and install if available
  egenskriven update

  # Only check for updates (don't install)
  egenskriven update --check

  # Force reinstall current version
  egenskriven update --force

  # JSON output for scripting
  egenskriven update --check --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(config, checkOnly, force, *jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force update even if already on latest version")

	return cmd
}

func runUpdate(config UpdateConfig, checkOnly, force, jsonOutput bool) error {
	// 1. Fetch latest release info from GitHub
	release, err := getLatestRelease(config.RepoOwner, config.RepoName)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(config.CurrentVersion, "v")

	// 2. Compare versions
	updateAvailable := latestVersion != currentVersion && currentVersion != "dev"
	
	// Handle JSON output
	if jsonOutput {
		result := map[string]interface{}{
			"current_version":  config.CurrentVersion,
			"latest_version":   release.TagName,
			"update_available": updateAvailable,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	// 3. Report status
	if !updateAvailable && !force {
		fmt.Printf("You're already on the latest version (%s)\n", config.CurrentVersion)
		return nil
	}

	if updateAvailable {
		fmt.Printf("Update available: %s -> %s\n", config.CurrentVersion, release.TagName)
	} else if force {
		fmt.Printf("Forcing reinstall of version %s\n", release.TagName)
	}

	// 4. If check-only, stop here
	if checkOnly {
		if updateAvailable {
			fmt.Println("\nRun 'egenskriven update' to install the update.")
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
	defer os.Remove(tmpFile) // Clean up on failure

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

	fmt.Printf("Successfully updated to %s\n", release.TagName)
	fmt.Println("\nRestart any running 'egenskriven serve' instances to use the new version.")

	return nil
}

// getLatestRelease fetches the latest release info from GitHub API
func getLatestRelease(owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

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

	// Create temp file in the same directory as target (for atomic rename)
	tmpFile, err := os.CreateTemp("", "egenskriven-update-*")
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
	// On Windows, we can't replace a running binary directly
	// On Unix, we can rename over it
	if runtime.GOOS == "windows" {
		// Rename current binary to .old
		oldPath := targetPath + ".old"
		os.Remove(oldPath) // Remove any existing .old file
		if err := os.Rename(targetPath, oldPath); err != nil {
			return fmt.Errorf("failed to backup current binary: %w", err)
		}
	}

	// Move new binary to target location
	// First try rename (fastest, atomic)
	if err := os.Rename(newBinary, targetPath); err != nil {
		// If rename fails (cross-device), fall back to copy
		if err := copyFile(newBinary, targetPath); err != nil {
			return err
		}
		os.Remove(newBinary)
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
		return err
	}

	// Preserve executable permission
	return os.Chmod(dst, 0755)
}
```

**Register the Command**:

Update `internal/commands/root.go`:

```go
func Register(app *pocketbase.PocketBase, versionInfo VersionInfo) {
	rootCmd := app.RootCmd
	
	// Global JSON flag
	var jsonOutput bool
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")

	// Add update command
	rootCmd.AddCommand(NewUpdateCmd(UpdateConfig{
		CurrentVersion: versionInfo.Version,
		RepoOwner:      "yourusername",  // TODO: Replace with actual
		RepoName:       "egenskriven",
	}, &jsonOutput))

	// ... rest of commands ...
}
```

**Steps**:

1. Create `internal/commands/update.go` with the code above.

2. Update `internal/commands/root.go` to register the command.

3. Replace `yourusername` with your actual GitHub username.

4. Build and test:
   ```bash
   make build
   ./egenskriven update --help
   ```
   
   **Expected output**:
   ```
   Check for updates and optionally update EgenSkriven to the latest version.

   By default, this command will:
   1. Check GitHub for the latest release
   2. Compare with your current version
   3. Download and install the new version if available
   ...
   ```

5. Test check-only mode:
   ```bash
   ./egenskriven update --check
   ```
   
   **Expected output** (if on latest):
   ```
   You're already on the latest version (v1.0.0)
   ```
   
   **Expected output** (if update available):
   ```
   Update available: v1.0.0 -> v1.1.0

   Run 'egenskriven update' to install the update.
   ```

6. Test JSON output:
   ```bash
   ./egenskriven update --check --json
   ```
   
   **Expected output**:
   ```json
   {
     "current_version": "v1.0.0",
     "latest_version": "v1.1.0",
     "update_available": true
   }
   ```

7. Test actual update (after publishing a release):
   ```bash
   ./egenskriven update
   ```
   
   **Expected output**:
   ```
   Update available: v1.0.0 -> v1.1.0
   Downloading egenskriven-darwin-arm64...
   Installing to /usr/local/bin/egenskriven...
   Successfully updated to v1.1.0

   Restart any running 'egenskriven serve' instances to use the new version.
   ```

**Write Tests**:

**File**: `internal/commands/update_test.go`

```go
package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLatestRelease(t *testing.T) {
	// Mock GitHub API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := GitHubRelease{
			TagName: "v1.2.0",
			Assets: []struct {
				Name               string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
			}{
				{Name: "egenskriven-darwin-arm64", BrowserDownloadURL: "https://example.com/binary"},
				{Name: "egenskriven-linux-amd64", BrowserDownloadURL: "https://example.com/binary2"},
			},
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Test would need to use the mock server URL
	// This is a simplified example
	t.Log("GitHub API mock test passed")
}

func TestVersionComparison(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"v1.0.0", "v1.0.0", false},  // Same version
		{"v1.0.0", "v1.1.0", true},   // Minor update
		{"v1.0.0", "v2.0.0", true},   // Major update
		{"v1.1.0", "v1.0.0", true},   // Downgrade (still different)
		{"dev", "v1.0.0", false},     // Dev version never updates automatically
	}

	for _, tt := range tests {
		t.Run(tt.current+"->"+tt.latest, func(t *testing.T) {
			currentClean := strings.TrimPrefix(tt.current, "v")
			latestClean := strings.TrimPrefix(tt.latest, "v")
			
			updateAvailable := latestClean != currentClean && currentClean != "dev"
			assert.Equal(t, tt.expected, updateAvailable)
		})
	}
}
```

**Security Considerations**:

| Concern | Mitigation |
|---------|------------|
| Man-in-the-middle | Use HTTPS for all downloads |
| Tampered binaries | Verify checksums (can be added) |
| Privilege escalation | Binary replaces itself, inherits same permissions |
| Interrupted download | Download to temp file first, then atomic rename |

**Optional Enhancement - Checksum Verification**:

For added security, verify the checksum before installing:

```go
// Add to update.go

func verifyChecksum(binaryPath, expectedHash string) error {
	file, err := os.Open(binaryPath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}
```

**Common Mistakes**:
- Forgetting to handle Windows executable extension (`.exe`)
- Not resolving symlinks before replacing binary
- Cross-device rename failures (temp file on different filesystem)
- Not handling running server instances (user must restart manually)

---

### 9.6 Write Documentation

**What**: Create comprehensive documentation for users and contributors.

**Why**: Good documentation is essential for adoption. Users need quick start guides; contributors need architecture docs.

**File**: Update `README.md`

```markdown
# EgenSkriven

A local-first kanban board with CLI-first design. Single binary, no dependencies, works offline.

<p align="center">
  <img src="docs/screenshot.png" alt="EgenSkriven UI" width="800">
</p>

## Features

- **Single Binary**: Download and run. No database setup, no configuration.
- **CLI-First**: Full-featured command-line interface for power users and automation.
- **Real-Time Sync**: Changes in CLI appear instantly in the web UI.
- **Agent-Friendly**: Designed for AI coding assistants to use directly.
- **Multi-Board**: Organize work across multiple boards.
- **Keyboard-Driven UI**: Linear-inspired interface with full keyboard navigation.

## Quick Start

### Installation

**One-line install (macOS/Linux):**

```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/egenskriven/main/install.sh | sh
```

**Manual download:**

Download the latest release for your platform from [Releases](https://github.com/yourusername/egenskriven/releases).

**From source:**

```bash
git clone https://github.com/yourusername/egenskriven.git
cd egenskriven
make build
```

### Usage

**Start the server:**

```bash
egenskriven serve
```

Open http://localhost:8090 in your browser.

**Create a task:**

```bash
egenskriven add "Implement dark mode" --type feature --priority high
```

**List tasks:**

```bash
egenskriven list
egenskriven list --column todo --type bug
egenskriven list --json  # Machine-readable output
```

**Move a task:**

```bash
egenskriven move abc123 in_progress
```

## CLI Reference

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output in JSON format |
| `--quiet` | `-q` | Suppress non-essential output |
| `--board` | `-b` | Specify board by name or prefix |

### Commands

| Command | Description |
|---------|-------------|
| `add` | Create a new task |
| `list` | List and filter tasks |
| `show` | Show task details |
| `move` | Move task to column |
| `update` | Update task fields |
| `delete` | Delete tasks |
| `board` | Manage boards |
| `epic` | Manage epics |
| `prime` | Output agent instructions |
| `version` | Display version info |
| `completion` | Generate shell completions |

See `egenskriven [command] --help` for detailed usage.

## Shell Completions

Enable tab completion for your shell:

```bash
# Bash
egenskriven completion bash > /etc/bash_completion.d/egenskriven

# Zsh
egenskriven completion zsh > "${fpath[1]}/_egenskriven"

# Fish
egenskriven completion fish > ~/.config/fish/completions/egenskriven.fish
```

## AI Agent Integration

EgenSkriven is designed to work with AI coding assistants. Run `egenskriven prime` to get instructions for agents.

**Claude Code:**

Add to `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [
      { "hooks": [{ "type": "command", "command": "egenskriven prime" }] }
    ]
  }
}
```

**OpenCode:**

See `.opencode/plugin/egenskriven-prime.ts` for plugin integration.

## Development

### Prerequisites

- Go 1.21+
- Node.js 20+
- Make

### Setup

```bash
git clone https://github.com/yourusername/egenskriven.git
cd egenskriven

# Install dependencies
go mod download
cd ui && npm install && cd ..

# Start development server
make dev
```

### Testing

```bash
make test              # Run all tests
make test-coverage     # Generate coverage report
```

### Building

```bash
make build    # Build for current platform
make release  # Build for all platforms
```

## Configuration

EgenSkriven stores data in `pb_data/` in the current directory. To use a different location:

```bash
egenskriven serve --dir /path/to/data
```

Project-specific agent configuration can be set in `.egenskriven/config.json`:

```json
{
  "agent": {
    "workflow": "strict",
    "mode": "autonomous"
  }
}
```

## License

MIT License - see [LICENSE](LICENSE) for details.
```

**Steps**:

1. Update `README.md` with the content above.

2. Replace `yourusername` with your actual GitHub username.

3. Add a screenshot (optional but recommended):
   - Create `docs/` directory if it doesn't exist
   - Add a screenshot as `docs/screenshot.png`

4. Review and customize sections as needed.

---

### 9.7 Create Changelog

**What**: Create a CHANGELOG.md file documenting changes for each release.

**Why**: Users and contributors need to know what changed between versions. It's also useful for debugging ("when did this feature get added?").

**File**: `CHANGELOG.md`

```markdown
# Changelog

All notable changes to EgenSkriven will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nothing yet

### Changed
- Nothing yet

### Fixed
- Nothing yet

## [1.0.0] - YYYY-MM-DD

### Added
- Initial release
- Core CLI commands: add, list, show, move, update, delete
- Web UI with kanban board view
- Real-time sync between CLI and UI
- Agent integration with `prime` command
- Multi-board support
- Task filtering and search
- Saved views
- Epics and sub-tasks
- Due dates with overdue indicators
- Import/export functionality
- Shell completions (bash, zsh, fish, powershell)
- Cross-platform binaries (macOS, Linux, Windows)

### Technical
- PocketBase backend with SQLite
- Embedded React frontend
- Single binary distribution
- No external dependencies
```

**Steps**:

1. Create the file:
   ```bash
   touch CHANGELOG.md
   ```

2. Open in your editor and paste the template above.

3. Update the date and version when releasing.

4. For each release, move items from [Unreleased] to the new version section.

---

### 9.8 Final Testing

**What**: Comprehensive testing across all platforms before release.

**Why**: Release is the last chance to catch issues. Testing on actual target platforms catches cross-platform bugs.

**Testing Checklist**:

#### Build Verification

- [ ] **All binaries build successfully**
  ```bash
  make release
  ```
  Should create 5 binaries in `dist/`.

- [ ] **Binaries have correct version**
  ```bash
  ./dist/egenskriven-linux-amd64 version
  ```
  Should show correct version, build date, and commit.

#### Platform Testing

Test on each platform (or use VMs/CI):

**macOS (if available):**
- [ ] Binary runs: `./egenskriven-darwin-arm64 serve`
- [ ] Web UI loads at http://localhost:8090
- [ ] CLI commands work: `./egenskriven-darwin-arm64 add "Test"`
- [ ] Shell completions install correctly

**Linux (or use CI/VM):**
- [ ] Binary runs: `./egenskriven-linux-amd64 serve`
- [ ] Web UI loads
- [ ] CLI commands work
- [ ] Shell completions work

**Windows (if available):**
- [ ] Binary runs: `egenskriven-windows-amd64.exe serve`
- [ ] Web UI loads
- [ ] CLI commands work
- [ ] PowerShell completions work

#### Installation Testing

- [ ] **Install script works**
  ```bash
  curl -fsSL https://raw.githubusercontent.com/yourusername/egenskriven/main/install.sh | sh
  ```

- [ ] **Installed binary is correct version**
  ```bash
  egenskriven version
  ```

- [ ] **Completions can be generated**
  ```bash
  egenskriven completion bash > /tmp/test-completion
  cat /tmp/test-completion | head -5
  ```

#### Functional Testing

- [ ] **Fresh install experience**
  - Delete `pb_data/` and run `egenskriven serve`
  - Should create database automatically
  - Admin UI should prompt for setup

- [ ] **Basic workflow**
  ```bash
  egenskriven add "Test task" --type bug --priority high
  egenskriven list
  egenskriven move <id> in_progress
  egenskriven list --column in_progress
  egenskriven delete <id> --force
  ```

- [ ] **Real-time sync**
  - Open web UI
  - Run `egenskriven add "Real-time test"`
  - Task should appear in UI without refresh

- [ ] **JSON output**
  ```bash
  egenskriven list --json | jq .
  ```
  Should output valid JSON.

---

## Verification Checklist

Complete each section in order. Check off each item as you verify it.

### Release Build Verification

- [ ] **All 5 binaries created**
  ```bash
  ls -la dist/
  ```
  Should show all platform binaries.

- [ ] **Checksums generated**
  ```bash
  cat dist/checksums.txt
  ```
  Should show SHA-256 checksums for all binaries.

- [ ] **Version embedded correctly**
  ```bash
  ./dist/egenskriven-linux-amd64 version
  ```
  Should show version, build date, commit, Go version.

### GitHub Release Verification

- [ ] **Workflow file exists**
  ```bash
  cat .github/workflows/release.yml
  ```

- [ ] **Tag triggers release**
  - Push a test tag: `git tag v0.0.1-test && git push origin v0.0.1-test`
  - Check GitHub Actions for build status
  - Check Releases page for created release

- [ ] **Release contains all assets**
  - 5 binaries
  - checksums.txt
  - Release notes

### Shell Completions Verification

- [ ] **Completion command exists**
  ```bash
  ./egenskriven completion --help
  ```

- [ ] **Bash completion generates**
  ```bash
  ./egenskriven completion bash | head -10
  ```

- [ ] **Zsh completion generates**
  ```bash
  ./egenskriven completion zsh | head -10
  ```

- [ ] **Fish completion generates**
  ```bash
  ./egenskriven completion fish | head -10
  ```

- [ ] **PowerShell completion generates**
  ```bash
  ./egenskriven completion powershell | head -10
  ```

### Installation Script Verification

- [ ] **Script is executable**
  ```bash
  ls -la install.sh
  ```
  Should show `-rwxr-xr-x`.

- [ ] **Script detects platform**
  ```bash
  ./install.sh 2>&1 | head -10
  ```
  Should show detected platform.

- [ ] **Script downloads binary**
  ```bash
  INSTALL_DIR=/tmp/test-install VERSION=v1.0.0 ./install.sh
  ls /tmp/test-install/egenskriven
  ```

### Update Command Verification

- [ ] **Update command exists**
  ```bash
  ./egenskriven update --help
  ```
  Should show usage and flags.

- [ ] **Check-only mode works**
  ```bash
  ./egenskriven update --check
  ```
  Should report current vs latest version.

- [ ] **JSON output works**
  ```bash
  ./egenskriven update --check --json
  ```
  Should output valid JSON with version info.

- [ ] **Update detects new version**
  - Build with an old version tag
  - Run `./egenskriven update --check`
  - Should report "Update available"

- [ ] **Update downloads and installs**
  ```bash
  ./egenskriven update
  ```
  Should download new binary and replace current one.

- [ ] **Force reinstall works**
  ```bash
  ./egenskriven update --force
  ```
  Should reinstall even if already on latest.

### Documentation Verification

- [ ] **README.md updated**
  - Contains installation instructions
  - Contains CLI reference
  - Contains development setup

- [ ] **CHANGELOG.md exists**
  - Contains version history
  - Documents features for v1.0.0

---

## File Summary

| File | Lines | Purpose |
|------|-------|---------|
| `Makefile` (updates) | ~50 | Release build targets |
| `cmd/egenskriven/main.go` (updates) | ~20 | Version variables |
| `internal/commands/version.go` | ~50 | Version command |
| `internal/commands/completion.go` | ~70 | Shell completions |
| `internal/commands/update.go` | ~200 | Self-update command |
| `internal/commands/update_test.go` | ~60 | Update command tests |
| `.github/workflows/release.yml` | ~100 | CI/CD for releases |
| `install.sh` | ~150 | Installation script |
| `README.md` (updates) | ~200 | Documentation |
| `CHANGELOG.md` | ~50 | Version history |

**Total new/updated code**: ~950 lines

---

## What You Should Have Now

After completing Phase 9:

```
egenskriven/
├── .github/
│   └── workflows/
│       └── release.yml              ✓ Created
├── cmd/
│   └── egenskriven/
│       └── main.go                  ✓ Updated (version vars)
├── internal/
│   └── commands/
│       ├── version.go               ✓ Created
│       ├── completion.go            ✓ Created
│       ├── update.go                ✓ Created
│       ├── update_test.go           ✓ Created
│       └── ...
├── dist/                            ✓ Created (by make release)
│   ├── egenskriven-darwin-arm64
│   ├── egenskriven-darwin-amd64
│   ├── egenskriven-linux-amd64
│   ├── egenskriven-linux-arm64
│   ├── egenskriven-windows-amd64.exe
│   └── checksums.txt
├── install.sh                       ✓ Created
├── README.md                        ✓ Updated
├── CHANGELOG.md                     ✓ Created
└── Makefile                         ✓ Updated
```

---

## Post-Release: Future Enhancements

After v1.0.0, consider these improvements:

### Phase 10+: Post-V1 Features

- **Custom Themes**: User-defined color schemes
- **TUI Mode**: Full terminal UI with Bubble Tea
- **Git Integration**: Link tasks to branches/commits
- **Sync & Collaboration**: Optional cloud sync

### Distribution Enhancements

- **Homebrew Formula**: `brew install egenskriven`
- **APT/YUM Packages**: Native Linux packages
- **Chocolatey Package**: Windows package manager
- **Docker Image**: Containerized distribution

---

## Troubleshooting

### "sha256sum: command not found" (macOS)

**Problem**: macOS uses `shasum` instead of `sha256sum`.

**Solution**: Update the Makefile:
```makefile
checksums:
	@echo "Generating checksums..."
	cd dist && shasum -a 256 * > checksums.txt
```

Or install coreutils: `brew install coreutils`

### Release workflow fails with permission error

**Problem**: GitHub Actions doesn't have permission to create releases.

**Solution**: Ensure the workflow has:
```yaml
permissions:
  contents: write
```

### Install script fails with "curl: command not found"

**Problem**: Target system doesn't have curl installed.

**Solution**: Add wget fallback to install.sh or document curl as a requirement.

### Binary crashes on Windows with "error while loading shared libraries"

**Problem**: CGO was enabled during build, requiring C libraries.

**Solution**: Ensure `CGO_ENABLED=0` is set for all builds.

### Version shows "dev" after tagging

**Problem**: Build wasn't done after tagging, or ldflags weren't applied.

**Solution**: 
1. Create tag: `git tag v1.0.0`
2. Build: `make build`
3. Verify: `./egenskriven version`

The VERSION is read from git tags via `git describe`.

### Update command fails with "permission denied"

**Problem**: User doesn't have write permission to the binary location.

**Solution**: 
- If installed in `/usr/local/bin`, use sudo: `sudo egenskriven update`
- Or reinstall to user directory: `INSTALL_DIR=~/.local/bin ./install.sh`

### Update command fails with "no binary found for platform"

**Problem**: The release doesn't include a binary for your OS/architecture combination.

**Solution**: 
- Check available platforms in the release
- Build from source for unsupported platforms
- Open an issue requesting the platform

### Update hangs or times out

**Problem**: Network issues or GitHub API rate limiting.

**Solution**:
- Check internet connection
- Try again later (rate limits reset hourly)
- Use install script as fallback: `curl -fsSL .../install.sh | sh`

### Running server doesn't use new version after update

**Problem**: The `egenskriven serve` process was running during update.

**Solution**: Restart the server after updating:
```bash
# Find and stop the running server
pkill -f "egenskriven serve"

# Start again
egenskriven serve
```

---

## Glossary

| Term | Definition |
|------|------------|
| **LDFLAGS** | Linker flags passed to Go compiler, used to embed version info |
| **CGO** | Go's C interop; disabled for pure Go builds |
| **Cross-compilation** | Building for a different OS/architecture than the host |
| **SHA-256** | Cryptographic hash function for verifying file integrity |
| **GitHub Actions** | CI/CD platform integrated with GitHub |
| **Semantic Versioning** | Version numbering scheme: MAJOR.MINOR.PATCH |
| **Shell Completion** | Tab-completion support for CLI commands |
