package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yacobolo/cssgen"
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen"},
	Short:   "Generate Go constants from CSS files",
	Long: `Parse CSS files and generate type-safe Go constants with 1:1 mapping.
Each CSS class becomes exactly one Go constant.`,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return loadConfig(cmd)
	},
	RunE: runGenerate,
}

func init() {
	f := generateCmd.Flags()
	f.String("source", "web/ui/src/styles", "Source CSS directory")
	f.String("output-dir", "internal/web/ui", "Output directory for generated files")
	f.StringSlice("include", nil, "Glob patterns for CSS files to include")
	f.String("format", "markdown", "Generation format: markdown|compact")
	f.Int("property-limit", 5, "Max properties per category in comments")
	f.Bool("show-internal", false, "Show -webkit-* properties")
	f.Bool("extract-intent", true, "Parse @intent comments from CSS")
	f.Bool("infer-layer", true, "Infer layer from file path")
	f.Bool("lint", false, "Run linter after generation")
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	config := buildGenerateConfig()

	result, err := cssgen.Generate(config)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	quiet := getBoolWithFallback("quiet", "quiet", false)

	if !quiet {
		fmt.Printf("Generated files in %s\n", config.OutputDir)
		fmt.Printf("  Files scanned: %d\n", result.FilesScanned)
		fmt.Printf("  Classes generated: %d\n", result.ClassesGenerated)

		for _, w := range result.Warnings {
			fmt.Printf("  Warning: %s\n", w)
		}
	}

	// Run lint after generate if --lint flag set
	lint, _ := cmd.Flags().GetBool("lint")
	if lint {
		return runLint(config.OutputDir, config.PackageName)
	}

	return nil
}
