// Package cssgen provides CSS constant generation and linting for Go/templ projects.
//
// cssgen generates type-safe Go constants from CSS files and provides a linter
// to eliminate hardcoded class strings and catch typos at build time.
//
// # Generation
//
// Generate Go constants from CSS files:
//
//	config := cssgen.Config{
//		SourceDir:   "web/styles",
//		OutputDir:   "internal/ui",
//		PackageName: "ui",
//		Includes:    []string{"**/*.css"},
//	}
//	result, err := cssgen.Generate(config)
//
// # Linting
//
// Lint CSS class usage in Go/templ files:
//
//	lintConfig := cssgen.LintConfig{
//		SourceDir:      "web/styles",
//		OutputDir:      "internal/ui",
//		LintPaths:      []string{"internal/**/*.{templ,go}"},
//		StrictMode:     false,
//	}
//	result, err := cssgen.Lint(lintConfig)
//
// # CLI Tool
//
// cssgen also provides a CLI tool. Install with:
//
//	go install github.com/yacobolo/cssgen/cmd/cssgen@latest
//
// See cmd/cssgen/README.md for CLI documentation.
package cssgen

// Public API is exported via linter.go:
// - Generate(config Config) (GenerateResult, error)
// - Lint(config LintConfig) (LintResult, error)
// - DetermineOutputFormat(requested string, quiet bool) OutputFormat
// - WriteOutput(w io.Writer, result LintResult, format OutputFormat, config LintConfig) error
