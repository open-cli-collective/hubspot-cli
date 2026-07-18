# CLAUDE.md

This file provides guidance for AI agents working with the hubspot-cli codebase.

## Project Overview

hubspot-cli is a command-line interface for HubSpot CRM written in Go. It uses the Cobra framework for commands and provides table/JSON/plain output formats.

## Quick Commands

```bash
# Build
make build

# Run tests
make test

# Run tests with coverage
make test-cover

# Lint
make lint

# Format and verify
make tidy

# Install locally
make install

# Clean build artifacts
make clean
```

## Architecture

```
hubspot-cli/
├── cmd/hspt/main.go              # Entry point - registers commands, calls Execute()
├── api/                          # Public Go library (importable) - TODO
├── internal/
│   ├── cmd/                      # Cobra commands (one package per resource)
│   │   ├── root/                 # Root command, Options struct, global flags
│   │   ├── initcmd/              # hspt init
│   │   ├── configcmd/            # hspt config {show,test,clear,set}
│   │   └── completion/           # Shell completion
│   ├── config/                   # JSON config loading
│   ├── version/                  # Build-time version injection via ldflags
│   ├── view/                     # Output formatting (table, JSON, plain)
│   └── exitcode/                 # Exit code constants
├── Makefile                      # Build, test, lint targets
└── go.mod                        # Module: github.com/open-cli-collective/hubspot-cli
```

## Key Patterns

### Options Struct Pattern

Commands use an Options struct for dependency injection:

```go
// Root options (global flags)
type Options struct {
    Output  string
    NoColor bool
}
```

### Register Pattern

Each command package exports a Register function:

```go
func Register(rootCmd *cobra.Command, opts *root.Options) {
    cmd := &cobra.Command{
        Use:   "contacts",
        Short: "Manage HubSpot contacts",
    }
    cmd.AddCommand(newListCmd(opts))
    cmd.AddCommand(newGetCmd(opts))
    rootCmd.AddCommand(cmd)
}
```

### View Pattern

Use the View struct for formatted output:

```go
v := view.New(opts.Output, opts.NoColor)

// Table output
headers := []string{"ID", "EMAIL", "NAME"}
rows := [][]string{{"123", "john@example.com", "John Doe"}}
v.Table(headers, rows)

// JSON output
v.JSON(data)
```

## Testing

- Unit tests in `*_test.go` files alongside source
- Use `testify/assert` for assertions
- Table-driven tests for multiple scenarios
- Use `httptest.NewServer()` to mock API responses

Run tests: `make test`

Coverage report: `make test-cover && open coverage.html`

## Commit Conventions

Use conventional commits:

```
type(scope): description

feat(contacts): add list command
fix(config): handle empty token
docs(readme): add examples
```

| Prefix | Purpose | Triggers Release? |
|--------|---------|-------------------|
| `feat:` | New features | Yes |
| `fix:` | Bug fixes | Yes |
| `docs:` | Documentation only | No |
| `test:` | Adding/updating tests | No |
| `refactor:` | Code changes that don't fix bugs or add features | No |
| `chore:` | Maintenance tasks | No |
| `ci:` | CI/CD changes | No |

## CI & Release Workflow

Releases are automated with a dual-gate system to avoid unnecessary releases:

**Gate 1 - Path filter:** Triggers when Go code (`**.go`, `go.mod`, `go.sum`) **or** release-affecting config (`.goreleaser.yml`, `version.txt`, the `auto-release`/`release` workflow files) changes. Packaging config affects the shipped artifact, so a fix there must be releasable on its own (see #62).
**Gate 2 - Commit prefix:** Only `feat:` and `fix:` commits create releases.

**Manual lever:** the `Auto Release` workflow also accepts a `workflow_dispatch` (scoped to `main`) to cut a release on demand for a change the path filter can't catch. The commit-prefix gate does not apply to a manual dispatch.

This means:
- `feat: add command` / `fix: handle edge case` + a watched path changed → release
- A packaging/release-config fix (e.g. a Homebrew cask fix in `.goreleaser.yml`) → release, so it reaches the tap
- `docs:`, `ci:`, `test:`, `refactor:`, or changes only to docs / non-release CI → no release
- Actions → Auto Release → Run workflow (on `main`) → release

**After merging a release-triggering PR:** The workflow creates a tag, which triggers GoReleaser to build binaries and publish to Homebrew. Chocolatey and Winget are published automatically.

## Environment Variables

| Setting | Environment Variable |
|---------|---------------------|
| Access Token | `HUBSPOT_ACCESS_TOKEN` |

## Dependencies

Key dependencies:
- `github.com/spf13/cobra` - CLI framework
- `github.com/fatih/color` - Colored terminal output
- `github.com/stretchr/testify` - Testing assertions
