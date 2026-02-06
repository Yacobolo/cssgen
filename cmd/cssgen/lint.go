package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yacobolo/cssgen/internal/cssgen"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint CSS class usage in Go/templ files",
	Long: `Check that CSS class references in Go and templ files use generated constants.
Detects hardcoded class strings, invalid classes, and unused constants.`,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return loadConfig(cmd)
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		outputDir := getStringWithFallback("output-dir", "generate.output-dir", "internal/web/ui")
		pkg := getStringWithFallback("package", "package", "ui")
		return runLint(outputDir, pkg)
	},
}

func init() {
	f := lintCmd.Flags()
	f.StringSlice("paths", []string{
		"internal/web/features/**/*.templ",
		"internal/web/features/**/*.go",
	}, "File patterns to scan for class references")
	f.String("output-dir", "internal/web/ui", "Output directory containing generated files")
	f.Bool("strict", false, "Exit 1 on any issue (CI mode)")
	f.Float64("threshold", 0.0, "Minimum adoption percentage for strict mode")
	f.String("output-format", "", "Output format: issues|summary|full|json|markdown")
	f.Int("max-issues-per-linter", 0, "Max issues to show per linter (0=unlimited)")
	f.Int("max-same-issues", 0, "Max repeated issues to show (0=unlimited)")
	f.Bool("print-lines", true, "Show source lines with issues")
	f.Bool("print-linter-name", true, "Show (csslint) suffix on issues")
}

// runLint is shared between `cssgen lint` and `cssgen generate --lint`.
func runLint(outputDir, pkg string) error {
	generatedFile := filepath.Join(outputDir, "styles.gen.go")
	lintConfig := buildLintConfig(generatedFile)
	// Override package name from the parameter (may come from generate config)
	lintConfig.PackageName = pkg

	lintResult, err := cssgen.Lint(lintConfig)
	if err != nil {
		return fmt.Errorf("lint failed: %w", err)
	}

	quiet := getBoolWithFallback("quiet", "quiet", false)
	outputFormat := getStringWithFallback("output-format", "lint.output-format", "")
	format := cssgen.DetermineOutputFormat(outputFormat, quiet)

	if !quiet {
		cssgen.WriteOutput(os.Stdout, lintResult, format, lintConfig)
	}

	// Exit code logic - "Soft Gate" approach
	strict := getBoolWithFallback("strict", "lint.strict", false)
	if strict {
		// Strict mode: any issue (error or warning) fails the build
		if len(lintResult.Issues) > 0 {
			os.Exit(1)
		}

		// Also check threshold if specified
		threshold := getFloat64WithFallback("threshold", "lint.threshold", 0.0)
		if threshold > 0 && lintResult.UsagePercentage < threshold {
			if !quiet {
				fmt.Fprintf(os.Stderr, "\nStrict mode: Usage percentage %.1f%% is below threshold %.1f%%\n",
					lintResult.UsagePercentage, threshold)
			}
			os.Exit(1)
		}
	} else if lintResult.ErrorCount > 0 {
		// Default "Soft Gate" mode: only errors fail the build
		os.Exit(1)
	}

	return nil
}
