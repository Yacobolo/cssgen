package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetKoanf creates a fresh koanf instance for each test.
func resetKoanf() {
	k = koanf.New(".")
}

func TestConfigFileLoading(t *testing.T) {
	resetKoanf()

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".cssgen.yaml")
	configContent := `
package: custom-pkg
verbose: true

generate:
  source: custom/css
  output-dir: custom/output
  format: compact
  property-limit: 10

lint:
  strict: true
  threshold: 80.0
  paths:
    - "custom/**/*.templ"
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))
	require.NoError(t, loadConfigFromPath(configPath))

	assert.Equal(t, "custom-pkg", k.String("package"))
	assert.True(t, k.Bool("verbose"))
	assert.Equal(t, "custom/css", k.String("generate.source"))
	assert.Equal(t, "custom/output", k.String("generate.output-dir"))
	assert.Equal(t, "compact", k.String("generate.format"))
	assert.Equal(t, 10, k.Int("generate.property-limit"))
	assert.True(t, k.Bool("lint.strict"))
	assert.InDelta(t, 80.0, k.Float64("lint.threshold"), 0.01)
}

func TestConfigFileNotFound_UsesDefaults(t *testing.T) {
	resetKoanf()

	// Point to non-existent config — should not error
	require.NoError(t, loadConfigFromPath("/nonexistent/.cssgen.yaml"))

	// buildGenerateConfig should return defaults
	config := buildGenerateConfig()
	assert.Equal(t, "web/ui/src/styles", config.SourceDir)
	assert.Equal(t, "internal/web/ui", config.OutputDir)
	assert.Equal(t, "ui", config.PackageName)
	assert.Equal(t, "markdown", config.Format)
	assert.Equal(t, 5, config.PropertyLimit)
}

func TestEnvVarOverridesConfigFile(t *testing.T) {
	resetKoanf()

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".cssgen.yaml")
	configContent := `
generate:
  source: from-file
lint:
  strict: false
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Set env vars that should override config file
	t.Setenv("CSSGEN_GENERATE_SOURCE", "from-env")
	t.Setenv("CSSGEN_LINT_STRICT", "true")

	require.NoError(t, loadConfigFromPath(configPath))

	assert.Equal(t, "from-env", k.String("generate.source"))
	assert.True(t, k.Bool("lint.strict"))
}

func TestBuildGenerateConfig_Defaults(t *testing.T) {
	resetKoanf()

	config := buildGenerateConfig()
	assert.Equal(t, "web/ui/src/styles", config.SourceDir)
	assert.Equal(t, "internal/web/ui", config.OutputDir)
	assert.Equal(t, "ui", config.PackageName)
	assert.Equal(t, "markdown", config.Format)
	assert.Equal(t, 5, config.PropertyLimit)
	assert.False(t, config.ShowInternal)
	assert.True(t, config.ExtractIntent)
	assert.True(t, config.LayerInferFromPath)
	assert.Equal(t, []string{
		"layers/components/**/*.css",
		"layers/utilities.css",
		"layers/base.css",
	}, config.Includes)
}

func TestBuildLintConfig_Defaults(t *testing.T) {
	resetKoanf()

	config := buildLintConfig("/test/styles.gen.go")
	assert.Equal(t, "/test/styles.gen.go", config.GeneratedFile)
	assert.Equal(t, "ui", config.PackageName)
	assert.False(t, config.Strict)
	assert.InDelta(t, 0.0, config.Threshold, 0.01)
	assert.Equal(t, 0, config.MaxIssuesPerLinter)
	assert.True(t, config.PrintIssuedLines)
	assert.True(t, config.PrintLinterName)
	assert.Equal(t, []string{
		"internal/web/features/**/*.templ",
		"internal/web/features/**/*.go",
	}, config.ScanPaths)
}

func TestBuildGenerateConfig_FromConfigFile(t *testing.T) {
	resetKoanf()

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".cssgen.yaml")
	configContent := `
package: mypkg
generate:
  source: src/css
  output-dir: gen/out
  format: compact
  property-limit: 3
  include:
    - "**/*.css"
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))
	require.NoError(t, loadConfigFromPath(configPath))

	config := buildGenerateConfig()
	assert.Equal(t, "src/css", config.SourceDir)
	assert.Equal(t, "gen/out", config.OutputDir)
	assert.Equal(t, "mypkg", config.PackageName)
	assert.Equal(t, "compact", config.Format)
	assert.Equal(t, 3, config.PropertyLimit)
	assert.Equal(t, []string{"**/*.css"}, config.Includes)
}

func TestBuildLintConfig_FromConfigFile(t *testing.T) {
	resetKoanf()

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".cssgen.yaml")
	configContent := `
lint:
  strict: true
  threshold: 75.5
  paths:
    - "src/**/*.go"
  max-issues-per-linter: 10
  print-lines: false
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))
	require.NoError(t, loadConfigFromPath(configPath))

	config := buildLintConfig("/test/styles.gen.go")
	assert.True(t, config.Strict)
	assert.InDelta(t, 75.5, config.Threshold, 0.01)
	assert.Equal(t, []string{"src/**/*.go"}, config.ScanPaths)
	assert.Equal(t, 10, config.MaxIssuesPerLinter)
	assert.False(t, config.PrintIssuedLines)
}

func TestInitCommand_CreatesConfigFile(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	cmd := rootCmd
	cmd.SetArgs([]string{"init"})
	require.NoError(t, cmd.Execute())

	// Verify file was created
	data, err := os.ReadFile(".cssgen.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(data), "package: ui")
	assert.Contains(t, string(data), "generate:")
	assert.Contains(t, string(data), "lint:")
}

func TestInitCommand_RefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	// Create existing file
	require.NoError(t, os.WriteFile(".cssgen.yaml", []byte("existing"), 0644))

	cmd := rootCmd
	cmd.SetArgs([]string{"init"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInitCommand_ForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	// Create existing file
	require.NoError(t, os.WriteFile(".cssgen.yaml", []byte("existing"), 0644))

	cmd := rootCmd
	cmd.SetArgs([]string{"init", "--force"})
	require.NoError(t, cmd.Execute())

	data, err := os.ReadFile(".cssgen.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(data), "package: ui")
}

func TestVersionCommand(t *testing.T) {
	cmd := rootCmd
	cmd.SetArgs([]string{"version"})
	require.NoError(t, cmd.Execute())
}

func TestGetStringWithFallback(t *testing.T) {
	resetKoanf()

	// No keys set - should return default
	assert.Equal(t, "default", getStringWithFallback("flag-key", "config.key", "default"))
}

func TestGetBoolWithFallback(t *testing.T) {
	resetKoanf()

	// No keys set - should return default
	assert.False(t, getBoolWithFallback("flag-key", "config.key", false))
	assert.True(t, getBoolWithFallback("flag-key", "config.key", true))
}

func TestGetIntWithFallback(t *testing.T) {
	resetKoanf()

	// No keys set - should return default
	assert.Equal(t, 42, getIntWithFallback("flag-key", "config.key", 42))
}

func TestGetFloat64WithFallback(t *testing.T) {
	resetKoanf()

	// No keys set - should return default
	assert.InDelta(t, 3.14, getFloat64WithFallback("flag-key", "config.key", 3.14), 0.01)
}
