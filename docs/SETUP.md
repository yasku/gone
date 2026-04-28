# gone — Setup and Installation

## Prerequisites

- **Go 1.26+** — The project uses Go 1.26.1 features
- **macOS** — gone is macOS-exclusive (uses `osascript` for Trash integration)
- **Terminal** — Any modern terminal (iTerm2, Terminal.app, Alacritty, etc.)

## Installation Methods

### Homebrew (Recommended)

```bash
brew install yasku/tap/gone
```

### Pre-built Binary

Download the latest release from [GitHub Releases](https://github.com/yasku/gone/releases/latest):

```bash
# Apple Silicon (M1/M2/M3/M4)
curl -L -o gone https://github.com/yasku/gone/releases/latest/download/gone-darwin-arm64
chmod +x gone

# Intel
curl -L -o gone https://github.com/yasku/gone/releases/latest/download/gone-darwin-amd64
chmod +x gone

# First run — clear the quarantine bit (binaries are unsigned)
xattr -d com.apple.quarantine gone

./gone
```

**Verify integrity**: Check against `checksums.txt` from the release:
```bash
shasum -a 256 gone
```

### Go Install

```bash
go install github.com/yasku/gone/cmd/gone@latest
```

### From Source

```bash
git clone https://github.com/yasku/gone.git
cd gone
go build -o gone ./cmd/gone
./gone
```

## Optional External Tools

gone works without these tools but with enhanced functionality:

| Tool | Purpose | Install Command | Required |
|------|---------|-----------------|----------|
| `fd` | Fast file finder (accelerates scanning) | `brew install fd` | No |
| `osquery` | SQL OS queries for security audit | `brew install osquery` | No |
| `glances` | Alternative system monitor | `brew install glances` | No |
| `bmon` | Network monitor | `brew install bmon` | No |

To check which tools are installed:
```bash
./scripts/check-tools.sh
```

## Running gone

### Basic Usage

```bash
gone                    # Start with Uninstall tab
gone claude             # Start and immediately scan for "claude"
gone nvm rustup node    # Scan for multiple terms
```

### Command Line Options

gone accepts an initial search term as CLI arguments:

```bash
gone [search-term]
```

## Environment Variables

gone does not require environment variables. Optional configuration may be added in future versions.

## Build from Source

### Prerequisites

1. **Go 1.26+**
   ```bash
   go version  # Should output go1.26 or higher
   ```

2. **Git**
   ```bash
   git --version
   ```

### Build Steps

```bash
# Clone repository
git clone https://github.com/yasku/gone.git
cd gone

# Download dependencies
go mod download

# Build binary
go build -o gone ./cmd/gone

# Run
./gone
```

### Build for Different Architectures

```bash
# Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o gone-darwin-arm64 ./cmd/gone

# Intel
GOOS=darwin GOARCH=amd64 go build -o gone-darwin-amd64 ./cmd/gone
```

## Testing Locally

### Run All Tests

```bash
./scripts/test-all.sh
```

This runs:
- `go test ./cmd/... ./internal/... -v` — All tests
- `go test -race ./internal/...` — Race detector
- `go vet ./cmd/... ./internal/...` — Vet checks
- `go fmt ./cmd/... ./internal/...` — Format check

### Run Specific Tests

```bash
# Test a specific package
go test ./internal/tui/... -v

# Run with race detector
go test -race ./internal/...

# Run benchmarks
go test -bench=. ./internal/scanner/
```

### Manual Testing

```bash
# Build
go build -o gone ./cmd/gone

# Run with sample search
./gone claude

# Check available tools
./scripts/check-tools.sh
```

## Troubleshooting

### "loading..." hangs on startup

- Ensure you're on macOS
- Try running with `TERM=xterm-256color gone`

### Binary is unsigned (macOS警告)

```bash
xattr -d com.apple.quarantine gone
```

### osquery audit shows "osquery not found"

Install osquery:
```bash
brew install osquery
```

gone works without osquery but the Audit tab will show install instructions.

### Scanning is slow

Install `fd` for faster file finding:
```bash
brew install fd
```

### "Permission denied" errors

Some directories require permissions (e.g., `/usr/local`). Run:
```bash
# Check directory permissions
ls -ld /usr/local
```

## Uninstall gone

If installed via Homebrew:
```bash
brew uninstall gone
```

If installed via binary, simply delete the `gone` file:
```bash
rm /path/to/gone
```

Configuration and operation logs are in `~/.config/gone/` — delete manually if desired.
