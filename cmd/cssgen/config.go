package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	"github.com/yacobolo/cssgen/internal/cssgen"
)

var k = koanf.New(".")

// activeCmd holds the cobra command that was executed, used to check
// whether a flag was explicitly set on the command line.
var activeCmd *cobra.Command

// loadConfig loads configuration with precedence: flags > env > file > defaults.
// It must be called after cobra parses flags (in PreRunE or RunE).
func loadConfig(cmd *cobra.Command) error {
	activeCmd = cmd

	// Resolve config file path from flag
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = ".cssgen.yaml"
	}

	// Load config file and env vars
	if err := loadConfigFromPath(configPath); err != nil {
		return err
	}

	// 3. CLI flags (highest precedence — only flags that were explicitly set)
	// Merge flags from the specific command and its parent (root) flags.
	// The koanf instance (k) is passed so posflag can skip flags whose
	// keys already exist from the config file / env providers.
	if err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil); err != nil {
		return fmt.Errorf("loading command flags: %w", err)
	}

	return nil
}

// loadConfigFromPath loads configuration from a file and environment variables.
// This is separated from loadConfig to allow testing without a cobra command.
func loadConfigFromPath(configPath string) error {
	// 1. Config file (lowest precedence among providers)
	if _, err := os.Stat(configPath); err == nil {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			return fmt.Errorf("loading config file %s: %w", configPath, err)
		}
	}

	// 2. Environment variables (CSSGEN_* prefix)
	if err := k.Load(env.Provider("CSSGEN_", ".", func(s string) string {
		// CSSGEN_GENERATE_SOURCE -> generate.source
		// CSSGEN_LINT_STRICT -> lint.strict
		// CSSGEN_VERBOSE -> verbose
		return strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, "CSSGEN_")),
			"_", ".",
		)
	}), nil); err != nil {
		return fmt.Errorf("loading environment variables: %w", err)
	}

	return nil
}

// buildGenerateConfig constructs the library's Config struct from koanf state.
func buildGenerateConfig() cssgen.Config {
	config := cssgen.Config{
		SourceDir:          getStringWithFallback("source", "generate.source", "web/ui/src/styles"),
		OutputDir:          getStringWithFallback("output-dir", "generate.output-dir", "internal/web/ui"),
		PackageName:        getStringWithFallback("package", "package", "ui"),
		Verbose:            getBoolWithFallback("verbose", "verbose", false),
		Format:             getStringWithFallback("format", "generate.format", "markdown"),
		PropertyLimit:      getIntWithFallback("property-limit", "generate.property-limit", 5),
		ShowInternal:       getBoolWithFallback("show-internal", "generate.show-internal", false),
		ExtractIntent:      getBoolWithFallback("extract-intent", "generate.extract-intent", true),
		LayerInferFromPath: getBoolWithFallback("infer-layer", "generate.infer-layer", true),
	}

	// Handle includes: check flag key first, then config key
	if includes := k.Strings("include"); len(includes) > 0 {
		config.Includes = includes
	} else if includes := k.Strings("generate.include"); len(includes) > 0 {
		config.Includes = includes
	} else {
		config.Includes = []string{
			"layers/components/**/*.css",
			"layers/utilities.css",
			"layers/base.css",
		}
	}

	return config
}

// buildLintConfig constructs the library's LintConfig struct from koanf state.
func buildLintConfig(generatedFile string) cssgen.LintConfig {
	// Handle paths: check flag key first, then config key
	var scanPaths []string
	if paths := k.Strings("paths"); len(paths) > 0 {
		scanPaths = paths
	} else if paths := k.Strings("lint.paths"); len(paths) > 0 {
		scanPaths = paths
	} else {
		scanPaths = []string{
			"internal/web/features/**/*.templ",
			"internal/web/features/**/*.go",
		}
	}

	return cssgen.LintConfig{
		GeneratedFile:      generatedFile,
		PackageName:        getStringWithFallback("package", "package", "ui"),
		ScanPaths:          scanPaths,
		Verbose:            getBoolWithFallback("verbose", "verbose", false),
		Strict:             getBoolWithFallback("strict", "lint.strict", false),
		Threshold:          getFloat64WithFallback("threshold", "lint.threshold", 0.0),
		MaxIssuesPerLinter: getIntWithFallback("max-issues-per-linter", "lint.max-issues-per-linter", 0),
		MaxSameIssues:      getIntWithFallback("max-same-issues", "lint.max-same-issues", 0),
		ShowStats:          true,
		PrintIssuedLines:   getBoolWithFallback("print-lines", "lint.print-lines", true),
		PrintLinterName:    getBoolWithFallback("print-linter-name", "lint.print-linter-name", true),
		UseColors:          getBoolWithFallback("color", "color", false),
	}
}

// flagChanged reports whether the given flag was explicitly set on the command line.
func flagChanged(flagKey string) bool {
	if activeCmd == nil {
		return false
	}
	if f := activeCmd.Flags().Lookup(flagKey); f != nil {
		return f.Changed
	}
	// Check inherited (persistent) flags from parent commands
	if f := activeCmd.InheritedFlags().Lookup(flagKey); f != nil {
		return f.Changed
	}
	return false
}

// getStringWithFallback checks the flag key (only if explicitly set on CLI),
// then the config file key, then returns the default.
func getStringWithFallback(flagKey, configKey, defaultVal string) string {
	// CLI flag takes precedence — but only if explicitly changed
	if flagChanged(flagKey) {
		if v := k.String(flagKey); v != "" {
			return v
		}
	}
	if v := k.String(configKey); v != "" {
		return v
	}
	return defaultVal
}

// getBoolWithFallback checks the flag key (only if explicitly set on CLI),
// then the config file key, then returns the default.
func getBoolWithFallback(flagKey, configKey string, defaultVal bool) bool {
	if flagChanged(flagKey) {
		return k.Bool(flagKey)
	}
	if k.Exists(configKey) {
		return k.Bool(configKey)
	}
	return defaultVal
}

// getIntWithFallback checks the flag key (only if explicitly set on CLI),
// then the config file key, then returns the default.
func getIntWithFallback(flagKey, configKey string, defaultVal int) int {
	if flagChanged(flagKey) {
		return k.Int(flagKey)
	}
	if k.Exists(configKey) {
		return k.Int(configKey)
	}
	return defaultVal
}

// getFloat64WithFallback checks the flag key (only if explicitly set on CLI),
// then the config file key, then returns the default.
func getFloat64WithFallback(flagKey, configKey string, defaultVal float64) float64 {
	if flagChanged(flagKey) {
		return k.Float64(flagKey)
	}
	if k.Exists(configKey) {
		return k.Float64(configKey)
	}
	return defaultVal
}
