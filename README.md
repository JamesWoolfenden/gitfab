# gitfab

Opens a git repository in your default browser. Similar to `hub browse` but works with any Git hosting platform (GitHub, GitLab, Bitbucket, etc.) and can be run from any subdirectory within a repository.

## Features

- 🌐 **Universal Support** - Works with GitHub, GitLab, Bitbucket, and any Git hosting service
- 📁 **Smart Repository Detection** - Finds the Git repository from any subdirectory
- 🔄 **Multiple URL Formats** - Handles SSH, HTTPS, and ssh:// protocol URLs
- 🎯 **Remote Selection** - Choose which remote to open (defaults to `origin`)
- 🚀 **Pipeline Support** - Use `--target` flag to open Actions/Pipelines page (auto-detects platform)
- ⚡ **Fast & Lightweight** - Single binary with no dependencies

## Installation

### Homebrew (macOS/Linux)

```bash
brew install JamesWoolfenden/tap/gitfab
```

### From Source

```bash
go install github.com/jameswoolfenden/gitfab@latest
```

### Pre-built Binaries

Download the latest binary from the [releases page](https://github.com/JamesWoolfenden/gitfab/releases).

## Usage

### Basic Usage

Simply run `gitfab` from anywhere within a Git repository:

```bash
gitfab
```

This will open the repository's `origin` remote URL in your default browser.

### Command-Line Flags

```bash
gitfab [flags]
```

#### Flags

- `--help` - Show help information
- `--version` - Show version information
- `--remote <name>` - Specify which remote to open (default: "origin")
- `--target` or `-t` - Open pipeline/actions page (auto-detects based on platform)

### Examples

**Open the default remote (origin):**

```bash
gitfab
```

**Open a specific remote:**

```bash
gitfab --remote upstream
```

**Check version:**

```bash
gitfab --version
```

**Show help:**

```bash
gitfab --help
```

**Open the pipelines/actions page:**

```bash
gitfab --target
# or use the shorthand
gitfab -t
```

This will open:
- GitHub: `/actions` page
- GitLab: `/-/pipelines` page
- Bitbucket: `/pipelines` page

## How It Works

1. Detects the Git repository by walking up the directory tree from your current location
2. Reads the Git configuration to find the specified remote
3. Converts the Git URL to an HTTP(S) URL (handles SSH and HTTPS formats)
4. Opens the URL in your default browser

## Supported Platforms

- macOS
- Linux
- Windows
- FreeBSD
- OpenBSD
- Solaris

## Development

### Build

```bash
make build
```

### Run Tests

```bash
go test -v ./...
```

### Create Release

```bash
make bump
```

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
