# Agent Development Guide for cssgen

This document provides essential information for AI coding agents working on the cssgen project.

## Project Overview

**cssgen** is a type-safe CSS class constant generator and linter for Go/templ projects. It parses CSS files and generates Go constants with 1:1 mapping, providing build-time validation and IDE support.

**Tech Stack:** Go 1.21+, tdewolff/parse (CSS parsing), testify (testing)

**Philosophy:** 1:1 CSS-to-Go mapping. Each CSS class becomes exactly one Go constant. No "joined" constants like `"btn btn--primary"`. Composition happens in templates: `{ ui.Btn, ui.BtnPrimary }`.

## Build, Lint, Test Commands

### Task Commands (Preferred)

```bash
task build              # Build cssgen binary to bin/
task test               # Run tests with race detection
task test:coverage      # Run tests with HTML coverage report
task lint               # Run golangci-lint
task lint:fix           # Auto-fix linting issues
task check              # Run all checks (test + lint)
task fmt                # Format all Go code
task clean              # Remove build artifacts
task install            # Install to $GOPATH/bin
task example            # Run cssgen on testdata
```

### Running a Single Test

```bash
go test -v -race -run TestFunctionName ./...                      # Specific test
go test -v -race -run TestBEM ./...                               # BEM tests
go test -v -race -coverprofile=coverage.txt -run TestName ./...   # With coverage
```

### Direct Go Commands

```bash
go test -v -race ./...                    # All tests
go build -o bin/cssgen ./cmd/cssgen      # Build
golangci-lint run                         # Lint
```

## Code Style Guidelines

### Formatting & Imports

- Use `gofmt`, tabs for indentation, run `task fmt` before committing
- Import order: (1) stdlib, (2) external deps, (3) internal (blank lines between)

### Naming Conventions

- **Exported types/functions:** PascalCase (`CSSClass`, `Generate`, `LintConfig`)
- **Unexported types/functions:** camelCase (`parseFile`, `scanFile`, `parserState`)
- **Constants:** PascalCase for exported (`CategoryVisual`), not SCREAMING_SNAKE_CASE
- **Test functions:** `Test<FunctionName>` or `Test<Feature><Scenario>` (e.g., `TestBEMDetection`)
- **Files:** `lowercase.go`, `feature_test.go`, `output_format.go` (snake_case)
- **Generated files:** `*.gen.go` suffix

### Types, Errors, Comments

- Strong typing with custom types, type aliases for enums: `type PropertyCategory string`
- Pointer receivers for state modification, value receivers otherwise
- Error wrapping: `fmt.Errorf("context: %w", err)`, no `panic()` in library code
- All exported items need godoc starting with item name
- Package docs in `internal/cssgen/cssgen.go`, justify `#nosec` exclusions

### Linting Configuration

**Enabled linters (.golangci.yml):**

- Core: `govet`, `staticcheck`, `unused`
- Error handling: `errname`, `errorlint`, `nilerr`
- Logging: `sloglint` (structured logging with snake_case keys, static messages)
- Quality: `revive`, `misspell`, `nolintlint`, `testifylint`

**Disabled:** `errcheck`, `gosec`, `gocritic`

**Special rules:**

- Tests: Skip `gosec`, `errcheck` in `*_test.go`
- Output/reporting files: Ignored fmt errors (`linter.go`, `reporter*.go`, `output*.go`, `scanner.go`)
- Use `task lint:fix` for auto-fixable issues

## Testing

**Framework:** `github.com/stretchr/testify` (assert, require)

- Table-driven tests, `t.TempDir()` for auto-cleanup, >80% coverage on new code

## File Structure

```
cssgen/
├── cmd/cssgen/              # CLI entry point (main.go)
├── internal/cssgen/         # Core library code (private)
│   ├── testdata/            # Test fixtures (CSS files)
│   ├── *.go                 # Library implementation
│   └── *_test.go            # Test files
├── examples/                # Example CSS inputs/outputs
├── Taskfile.yml             # Task runner config
├── .golangci.yml            # Linter config
└── README.md                # User documentation
```

**Core files (in `internal/cssgen/`):**

- `generator.go` - Main entry point, file scanning, orchestration
- `parser.go` - CSS parsing, selector extraction
- `analyzer.go` - BEM detection, GoName generation
- `writer.go` - Go code generation
- `linter.go` - Linting logic
- `scanner.go` - File scanning, class reference extraction
- `types.go` - Core data types

## Common Patterns

### CSS to Go Naming & Performance

- `.btn` → `Btn = "btn"`, `.btn--primary` → `BtnPrimary`, `.card__header` → `CardHeader`, `._internal` → `_Internal`
- Cross-platform: use `filepath` package, normalize with `strings.ReplaceAll(path, "\\", "/")`
- Map lookups for constant resolution, deduplicate during parsing (maps → slices)
- No goroutines in core logic (sequential is fast enough)

## Git Workflow

- Create feature branches from `main`
- Run `task check` before committing
- Write descriptive commit messages
- Reference issues in commits: `fix: resolve BEM parsing issue (#123)`
- CI runs on all PRs (tests on Go 1.21, 1.22, 1.23 + linting)

## Common Tasks

**Add new output format:**

1. Add enum to `OutputFormat` in `types.go`
2. Create `output_<format>.go` with writer function
3. Update `WriteOutput()` in `output.go` to handle new format
4. Add tests in `output_test.go`
5. Update README.md with format documentation

**Add new CSS property category:**

1. Add constant to `PropertyCategory` in `types.go`
2. Update `categorizeProperty()` in `categorizer.go`
3. Add test cases in `generator_test.go`

**Fix linting issue:**

1. Reproduce: `go test -v -run TestLint<Scenario> ./...`
2. Debug in `linter.go` or `scanner.go`
3. Add regression test
4. Verify: `task check`

## References

- **User Docs:** README.md, CONTRIBUTING.md
- **Code Style:** https://go.dev/doc/effective_go
- **Testing:** https://go.dev/doc/tutorial/add-a-test
- **CI Config:** .github/workflows/ci.yml
