# AGENTS.md - Development Guidelines

This file provides instructions for agentic coding tools operating in this repository.

## Project Overview

Fantasy Grounds Character Extractor - A Go tool that parses Fantasy Grounds `db.xml` campaign files and extracts individual character sheets into separate XML files.

## Language

- **Go 1.25.7** - Modern Go with minimal dependencies (standard library only)
- Platform: macOS (arm64), macOS (x86_64), windows (x86_64)
- Environment: Managed via [devbox](https://www.jetpack.io/devbox/)

## Build Commands

```bash
# Build the application for current platform
task build

# Build for all platforms (macOS x86_64, macOS arm64, Windows x86_64)
task build-all

# Run with devbox (recommended)
cd /Users/sdelrio/github/sdelrio/fg-char-extract && devbox run -- task build-all
```

## Test Commands

```bash
# Run all tests (Taskfile)
task test

# Run all tests (Go)
go test -v ./...

# Run single test file
go test -v -run TestRun

# Run single test
go test -v -run "TestRun" -count=1

# Run test against specific XML file
# (No built-in test for this, edit main_test.go)
```

## Devbox Setup

```bash
# Setup devbox environment
cd /Users/sdelrio/github/sdelrio/fg-char-extract && devbox shell

# Install devbox if not present
curl -fsSL https://get.jetpack.io/devbox/install.sh | bash
```

## Code Style Guidelines

## Imports

- Import standard library packages first, then external packages (none required)
- Alphabetize imports within each category
- Use backticks for package references in documentation
- No blank imports

```go
import (
    "encoding/xml"
    "fmt"
    "io"
    "os"
    "regexp"
    "strings"
)
```

## Formatting

- **no gofmt required** - Code is manually formatted (uses tabs for indentation)
- 2-space indent for function bodies within braces
- No trailing whitespace
- Single blank line between logical blocks
- Maximum line length ~80-100 characters

## Types

- Use named struct types that describe the domain entity
- Export (capitalized) fields for struct tags and public access
- Export methods that have side effects or are part of public API
- Keep structs focused - one responsibility per struct
- Use `xml.` prefix for XML-specific types when needed

```go
type Character struct {
    ID     string
    Level  string
    Tokens []xml.Token
}
```

## Naming Conventions

### Functions

- `snake_case` for function names (consistent with Go conventions)
- Start with verb for methods
- Use concise but descriptive names
- Use `run` for main entry logic
- Use `write*` for file I/O operations

### Variables

- `snake_case` for variables
- Use short names when meaning is clear from context (`f`, `c`)
- Be explicit in error handling paths
- Use descriptive names for parameters

```go
filename := "db.xml"
level := strings.TrimSpace(c.Level)
```

### Constants

- `PascalCase` or `UPPER_SNAKE_CASE` depending on scope
- Local constants: `pascalCase`
- Global constants: `CONST_NAME`

## Error Handling

- Use `error` return values for all functions that can fail
- Use `fmt.Errorf` with `%w` for wrapping errors
- Contextualize errors with specific messages
- Use `fmt.Fprintf(os.Stderr, ...)` for error output
- Call `os.Exit(1)` after printing stderr error message
- Don't ignore errors (except in cleanup paths)

```go
func run(filename string) error {
    f, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("opening %s: %w", filename, err)
    }
    defer f.Close()
    // ...
}
```

## XML Handling

- Use `xml.NewDecoder` and manual token traversal for complex XML
- Escape special characters: `&`, `<`, `>` (not tabs/newlines)
- Flush encoder after writing tokens when necessary
- Preserve original whitespace and formatting for character data
- Use `strings.ReplaceAll` for efficient bulk escaping
- Maintain XML depth tracking for nested element parsing

## File I/O

- Use `os.Open` for reading, `os.Create` for writing
- Always use `defer f.Close()`
- Check write errors after every file write operation
- Construct output filenames in format: `character_<ID>_<Level>.xml`
- Use `filepath.Join` for platform-independent paths

## XML Structure

- Root element: `<root version="3.1" release="7|CoreRPG:3">`
- Character wrapper: `<character>` (always indented)
- Filter out ignored tags: `<public>`, `<holder*`
- Preserve all other character data and structure
- Include XML declaration at start: `<?xml version="1.0" encoding="UTF-8"?>`

## Testing

### Golden File Pattern

- Store expected output in `tests/expected_<output>.xml`
- Compare byte-for-byte with generated output
- Use `defer os.Remove()` for test cleanup
- Place test XML inputs in `tests/` directory

```go
func TestRun(t *testing.T) {
    defer os.Remove("character_id-00001_4.xml")
    if err := run("tests/db.xml"); err != nil {
        t.Fatalf("run() failed: %v", err)
    }
    // Compare golden file
}
```

## Taskfile

This project uses [Task](https://taskfile.dev/) for task management:

```yaml
version: '3'

vars:
  PLATFORMS: darwin/amd64 darwin/arm64 windows/amd64

tasks:
  build:
    desc: Build the application for current platform
    cmds:
      - mkdir -p dist
      - go build -o dist/fg-char-extract main.go
    sources:
      - "*.go"
    generates:
      - dist/fg-char-extract

  build-all:
    desc: Build for all supported platforms
    deps: [clean]
    cmds:
      - for: { var: PLATFORMS }
        task: build-for
        vars: { TARGET: '{{.ITEM}}' }

  build-for:
    internal: true
    vars:
      GOOS: '{{index (splitList "/" .TARGET) 0}}'
      GOARCH: '{{index (splitList "/" .TARGET) 1}}'
      BIN: 'dist/fg-char-extract-{{.GOOS}}-{{.GOARCH}}{{if eq .GOOS "windows"}}.exe{{end}}'
    cmds:
      - echo "Building for {{.GOOS}}/{{.GOARCH}}..."
      - mkdir -p dist
      - GOOS={{.GOOS}} GOARCH={{.GOARCH}} go build -o {{.BIN}} main.go

  run:
    desc: Run against local db.xml
    cmds:
      - go run main.go ./tests/db.xml

  test:
    desc: Run unit tests
    cmds:
      - go test -v ./...

  clean:
    desc: Remove binary and generated XML files
    cmds:
      - rm -rf dist/
      - rm -f character_*.xml
```

## Clean Commands

```bash
# Taskfile clean
task clean

# Manual clean
rm -f fg-char-extract character_*.xml
```

## Directory Structure

```
fg-char-extract/
├── main.go               # Main application logic
├── main_test.go          # Unit tests
├── go.mod                # Go module definition
├── Taskfile.yml          # Task definitions
├── tests/
│   ├── db.xml            # Test input
│   └── expected_*.xml    # Golden files
├── character_*.xml       # Extracted output (not tracked)
├── fg-char-extract       # Binary (not tracked)
└── .envrc                # Devbox environment
```

## Git Ignore Patterns

Binary and generated files should not be tracked:
- `fg-char-extract`
- `*.xml` (generated character files)

## Git Guidelines

Use **Conventional Commits** format: `<type>(<scope>): <description>`

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

**Examples**:
```
feat(main): add proficiency bonus to character skillsset xml
fix: resolve test failures on checking skills
docs: update deployment instructions
```

## License

MIT
