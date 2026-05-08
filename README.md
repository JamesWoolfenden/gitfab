# gitfab

Opens a git repository in your default browser. Similar to `hub browse` but works with any Git hosting platform (GitHub, GitLab, Bitbucket, etc.) and can be run from any subdirectory within a repository.

## Features

- 🌐 **Universal Support** - Works with GitHub, GitLab, Bitbucket, and any Git hosting service
- 📁 **Smart Repository Detection** - Finds the Git repository from any subdirectory
- 🔄 **Multiple URL Formats** - Handles SSH, HTTPS, and ssh:// protocol URLs
- 🎯 **Remote Selection** - Choose which remote to open (defaults to `origin`)
- 🚀 **Pipeline Support** - Use `--target` flag to open Actions/Pipelines page (auto-detects platform)
- 👀 **Watch CI Builds** - Use `--watch` to follow GitHub Actions runs live in your terminal
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
- `--watch` or `-w` - Watch running CI builds in the console (GitHub only)
- `--once` or `-1` - Print CI build status once as a plain table and exit (GitHub only)
- `--wait` - Block until active CI runs finish, print result, exit non‑zero on failure (GitHub only)
- `--json` - Emit CI build status as JSON instead of a table (with `--once` or `--wait`)
- `--branch <name>` - Filter `--watch`/`--once`/`--wait` to runs on a single branch

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

**Watch CI builds in the terminal:**

```bash
gitfab --watch
# or use the shorthand
gitfab -w
```

Polls GitHub Actions every 5 seconds and renders a live table of workflow runs (status, workflow, branch, event, age). Exits automatically once all in‑progress builds finish, or press `Ctrl+C` to stop.

```text
Watching JamesWoolfenden/gitfab — Ctrl+C to stop

STATUS         WORKFLOW         BRANCH        EVENT         AGE  ID
● in_progress  CI               feat/watch    push          12s  9123456789
● queued       CodeQL           feat/watch    push          12s  9123456790
✓ success      CI               main          push          3m   9123456712
✗ failure      goreleaser       v0.1.0        push          1h   9123455100
- cancelled    CI               chore/deps    pull_request  2h   9123454001

Updated 09:41:07
```

Set `GITHUB_TOKEN` to authenticate — without it you are limited to public repos and the unauthenticated GitHub API rate limit:

```bash
export GITHUB_TOKEN=ghp_xxx
gitfab -w
```

**One‑shot status (script/pipe friendly):**

```bash
gitfab --once                 # plain table, no ANSI escapes
gitfab -1 --branch feat/foo   # only runs on feat/foo
gitfab --json | jq '.[0]'     # machine‑readable
```

**Wait for CI to finish:**

```bash
git push
gitfab --wait --branch "$(git rev-parse --abbrev-ref HEAD)"
echo "exit=$?"   # 0 if all watched runs passed, 1 otherwise
```

`--wait` polls until every run that was active during the wait has completed, prints a final table (or JSON with `--json`), and exits non‑zero if any of them failed, timed out, or was cancelled. Handy in shell scripts or as a background task in your editor/agent.

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
