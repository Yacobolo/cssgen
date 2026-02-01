// Package main provides the cssgen CLI tool for generating type-safe CSS constants.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yacobolo/cssgen"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `cssgen - CSS constant generator and linter

Type-safe CSS class constants for Go/templ projects.
See cmd/cssgen/README.md for full documentation and examples.

Usage: cssgen [OPTIONS]

MODES:
  -lint-only              Run linter without generation (default: generate)

GENERATION OPTIONS:
  -source DIR             Source CSS directory (default: web/ui/src/styles)
  -output-dir DIR         Output directory for generated files (default: internal/web/ui)
  -package NAME           Go package name (default: ui)
  -include PATTERNS       Comma-separated glob patterns
  -format MODE            Generation format: markdown|compact (default: markdown)
  -property-limit N       Max properties per category (default: 5)
  -extract-intent         Parse @intent comments (default: true)
  -infer-layer            Infer layer from file path (default: true)
  -show-internal          Show -webkit-* properties

LINTING OPTIONS:
  -lint                   Run linter after generation
  -lint-paths PATTERNS    Files to scan (default: internal/web/**/*.{templ,go})
  -strict                 Exit 1 on any issue (CI mode)
  -threshold PERCENT      Minimum adoption percentage for -strict mode

OUTPUT CONTROL:
  -output-format MODE     Output format: issues|summary|full|json|markdown
                          (default: issues - golangci-lint style)
                            issues   = Errors/warnings only (default)
                            summary  = Statistics and Quick Wins only
                            full     = Everything (issues + stats + wins)
                            json     = Machine-readable JSON
                            markdown = Shareable Markdown report
  -quiet                  Suppress all output (exit code only)

OUTPUT OPTIONS:
  -max-issues-per-linter N  Limit issues shown (0=unlimited)
  -max-same-issues N        Deduplicate repeated issues (0=unlimited)
  -print-lines              Show source lines with issues (default: true)
  -print-linter-name        Show (csslint) suffix (default: true)
  -color                    Force color output

EXAMPLES:
  # Generate CSS constants
  cssgen

  # Lint (default: issues only - fast and clean)
  cssgen -lint-only

  # Lint with statistics and Quick Wins
  cssgen -lint-only -output-format full

  # Weekly report (stats only, no individual issues)
  cssgen -lint-only -output-format summary

  # Export to Markdown for GitHub issue
  cssgen -lint-only -output-format markdown > report.md

  # CI mode (exit 1 on any issue)
  cssgen -lint-only -strict

  # Quiet mode (exit code only, for pre-commit hooks)
  cssgen -lint-only -quiet

`)
	}
}

func main() {
	var (
		// Generation options
		source        = flag.String("source", "web/ui/src/styles", "Source CSS directory")
		outputDir     = flag.String("output-dir", "internal/web/ui", "Output directory for generated files")
		pkg           = flag.String("package", "ui", "Go package name")
		include       = flag.String("include", "", "Comma-separated glob patterns")
		format        = flag.String("format", "markdown", "Output format: markdown, compact")
		propertyLimit = flag.Int("property-limit", 5, "Max properties per category")
		showInternal  = flag.Bool("show-internal", false, "Show -webkit-* properties")
		extractIntent = flag.Bool("extract-intent", true, "Parse @intent comments")
		inferLayer    = flag.Bool("infer-layer", true, "Infer layer from file path")

		// Linter modes
		lint     = flag.Bool("lint", false, "Run linter after generation")
		lintOnly = flag.Bool("lint-only", false, "Run linter without generation")

		// Linter options
		lintPaths = flag.String("lint-paths", "internal/web/features/**/*.templ,internal/web/features/**/*.go", "Comma-separated glob patterns for files to scan")
		strict    = flag.Bool("strict", false, "Exit with code 1 if linting issues found (for CI)")
		threshold = flag.Float64("threshold", 0.0, "Minimum adoption percentage for -strict mode")

		// Output control
		outputFormat = flag.String("output-format", "", "Output format: issues|summary|full|json|markdown (default: auto)")
		quiet        = flag.Bool("quiet", false, "Suppress all output (exit code only)")
		verbose      = flag.Bool("verbose", false, "Enable verbose generation logging")

		// Output options
		maxIssuesPerLinter = flag.Int("max-issues-per-linter", 0, "Maximum issues to show (0=unlimited)")
		maxSameIssues      = flag.Int("max-same-issues", 0, "Maximum same issues to show (0=unlimited)")
		printLines         = flag.Bool("print-lines", true, "Show source lines with issues")
		printLinter        = flag.Bool("print-linter-name", true, "Show (csslint) suffix")
		colorFlag          = flag.Bool("color", false, "Force color output")
	)
	flag.Parse()

	// Build config
	config := cssgen.Config{
		SourceDir:          *source,
		OutputDir:          *outputDir,
		PackageName:        *pkg,
		Verbose:            *verbose, // Still used for generation logging
		Format:             *format,
		PropertyLimit:      *propertyLimit,
		ShowInternal:       *showInternal,
		ExtractIntent:      *extractIntent,
		LayerInferFromPath: *inferLayer,
	}

	// Parse includes
	if *include != "" {
		config.Includes = strings.Split(*include, ",")
	} else {
		// Default includes
		config.Includes = []string{
			"layers/components/**/*.css",
			"layers/utilities.css",
			"layers/base.css",
		}
	}

	// Generate (unless lint-only)
	if !*lintOnly {
		result, err := cssgen.Generate(config)
		if err != nil {
			log.Fatalf("Generation failed: %v", err)
		}

		// Report
		fmt.Printf("✓ Generated files in %s\n", config.OutputDir)
		fmt.Printf("  Files scanned: %d\n", result.FilesScanned)
		fmt.Printf("  Classes generated: %d\n", result.ClassesGenerated)

		if len(result.Warnings) > 0 {
			fmt.Printf("\n⚠ Warnings:\n")
			for _, w := range result.Warnings {
				fmt.Printf("  - %s\n", w)
			}
		}
	}

	// Run linter if requested
	if *lint || *lintOnly {
		parsedPaths := parsePaths(*lintPaths)

		// Determine the generated file path for linter
		// AllCSSClasses map is always in styles.gen.go
		generatedFile := filepath.Join(config.OutputDir, "styles.gen.go")

		lintConfig := cssgen.LintConfig{
			GeneratedFile: generatedFile,
			PackageName:   *pkg,
			ScanPaths:     parsedPaths,
			Verbose:       *verbose, // Still used for generation logging
			Strict:        *strict,
			Threshold:     *threshold,

			// Output configuration
			MaxIssuesPerLinter: *maxIssuesPerLinter,
			MaxSameIssues:      *maxSameIssues,
			ShowStats:          true, // Controlled by output format now
			PrintIssuedLines:   *printLines,
			PrintLinterName:    *printLinter,
			UseColors:          *colorFlag,
		}

		lintResult, err := cssgen.Lint(lintConfig)
		if err != nil {
			log.Fatalf("Lint failed: %v", err)
		}

		// Determine output format
		format := cssgen.DetermineOutputFormat(*outputFormat, *quiet)

		// Write output (unless -quiet)
		if !*quiet {
			cssgen.WriteOutput(os.Stdout, lintResult, format, lintConfig)
		}

		// Exit code logic - "Soft Gate" approach
		var exitCode int
		if *strict {
			// Strict mode: any issue (error or warning) fails the build
			if len(lintResult.Issues) > 0 {
				exitCode = 1
			}

			// Also check threshold if specified
			belowThreshold := *threshold > 0 && lintResult.UsagePercentage < *threshold
			if belowThreshold && !*quiet {
				fmt.Fprintf(os.Stderr, "\nStrict mode: Usage percentage %.1f%% is below threshold %.1f%%\n",
					lintResult.UsagePercentage, *threshold)
				exitCode = 1
			}
		} else {
			// Default "Soft Gate" mode: only errors fail the build
			if lintResult.ErrorCount > 0 {
				exitCode = 1
			}
		}

		os.Exit(exitCode)
	}

	os.Exit(0)
}

// parsePaths splits comma-separated paths into a slice
func parsePaths(paths string) []string {
	if paths == "" {
		return []string{}
	}

	parts := strings.Split(paths, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
