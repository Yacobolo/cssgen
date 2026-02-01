package cssgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGeneratedFile(t *testing.T) {
	tests := []struct {
		name              string
		content           string
		expectedConstants map[string]string
		expectedAllCSS    map[string]bool
	}{
		{
			name: "simple constants",
			content: `package ui

var AllCSSClasses = map[string]bool{
	"btn": true,
	"btn--primary": true,
	"app-sidebar": true,
}

const Btn = "btn"
const BtnPrimary = "btn btn--primary"
const AppSidebar = "app-sidebar"
`,
			expectedConstants: map[string]string{
				"Btn":        "btn",
				"BtnPrimary": "btn btn--primary",
				"AppSidebar": "app-sidebar",
			},
			expectedAllCSS: map[string]bool{
				"btn":          true,
				"btn--primary": true,
				"app-sidebar":  true,
			},
		},
		{
			name: "with comments",
			content: `package ui

var AllCSSClasses = map[string]bool{
	"btn": true,
	"btn--primary": true,
}

// Btn is a button class
const Btn = "btn"
const BtnPrimary = "btn btn--primary" // Primary button
`,
			expectedConstants: map[string]string{
				"Btn":        "btn",
				"BtnPrimary": "btn btn--primary",
			},
			expectedAllCSS: map[string]bool{
				"btn":          true,
				"btn--primary": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpfile, err := os.CreateTemp("", "test-*.go")
			require.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			_, err = tmpfile.WriteString(tt.content)
			require.NoError(t, err)
			tmpfile.Close()

			// Parse
			constants, allCSS, err := ParseGeneratedFile(tmpfile.Name())
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConstants, constants)
			assert.Equal(t, tt.expectedAllCSS, allCSS)
		})
	}
}

func TestExtractClassesFromLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected []ClassReference
	}{
		{
			name: "class attribute with single class",
			line: `<div class="app-sidebar">`,
			expected: []ClassReference{
				{FullClassValue: "app-sidebar", IsConstant: false},
			},
		},
		{
			name: "class attribute with multiple classes",
			line: `<button class="btn btn--primary btn--lg">`,
			expected: []ClassReference{
				{FullClassValue: "btn btn--primary btn--lg", IsConstant: false},
			},
		},
		{
			name: "class with ui constant",
			line: `<div class={ ui.AppSidebar }>`,
			expected: []ClassReference{
				{IsConstant: true, ConstName: "AppSidebar"},
			},
		},
		{
			name: "class with multiple ui constants",
			line: `<button class={ ui.Btn, ui.BtnPrimary, ui.BtnLg }>`,
			expected: []ClassReference{
				{IsConstant: true, ConstName: "Btn"},
				{IsConstant: true, ConstName: "BtnPrimary"},
				{IsConstant: true, ConstName: "BtnLg"},
			},
		},
		{
			name: "class with string literal in braces",
			line: `<div class={ "nav-group" }>`,
			expected: []ClassReference{
				{FullClassValue: "nav-group", IsConstant: false},
			},
		},
		{
			name: "templ.KV with string",
			line: `<div class={ templ.KV("nav-group--iconic", true) }>`,
			expected: []ClassReference{
				{FullClassValue: "nav-group--iconic", IsConstant: false},
			},
		},
		{
			name: "templ.Classes with mixed content",
			line: `<div class={ templ.Classes("btn", ui.BtnPrimary) }>`,
			expected: []ClassReference{
				{FullClassValue: "btn", IsConstant: false},
				{IsConstant: true, ConstName: "BtnPrimary"},
			},
		},
		{
			name: "templ.Classes with multi-class string",
			line: `<div class={ templ.Classes("btn btn--sm") }>`,
			expected: []ClassReference{
				{FullClassValue: "btn btn--sm", IsConstant: false},
			},
		},
		{
			name:     "comment line",
			line:     `// class="old-style"`,
			expected: []ClassReference{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractClassesFromLine(tt.line, 1, "test.templ")

			// Compare only relevant fields (ignore Location, LineContent)
			require.Len(t, result, len(tt.expected), "wrong number of results")

			for i, ref := range result {
				assert.Equal(t, tt.expected[i].FullClassValue, ref.FullClassValue, "FullClassValue mismatch at index %d", i)
				assert.Equal(t, tt.expected[i].IsConstant, ref.IsConstant, "isConstant mismatch at index %d", i)
				assert.Equal(t, tt.expected[i].ConstName, ref.ConstName, "constName mismatch at index %d", i)
			}
		})
	}
}

func TestBuildLookupMaps(t *testing.T) {
	constants := map[string]string{
		"Btn":            "btn",
		"BtnPrimary":     "btn--primary",
		"BtnDanger":      "btn--danger",
		"AppSidebar":     "app-sidebar",
		"NavGroup":       "nav-group",
		"NavGroupIconic": "nav-group--iconic",
	}

	lookup := buildLookupMaps(constants)

	// Test exact matches (1:1 mapping - each class maps to exactly one constant)
	assert.Equal(t, "Btn", lookup.ExactMap["btn"])
	assert.Equal(t, "BtnPrimary", lookup.ExactMap["btn--primary"])
	assert.Equal(t, "BtnDanger", lookup.ExactMap["btn--danger"])
	assert.Equal(t, "AppSidebar", lookup.ExactMap["app-sidebar"])
	assert.Equal(t, "NavGroup", lookup.ExactMap["nav-group"])
	assert.Equal(t, "NavGroupIconic", lookup.ExactMap["nav-group--iconic"])

	// Test ConstantParts (with 1:1 mapping, always single-element)
	assert.Equal(t, []string{"btn"}, lookup.ConstantParts["Btn"])
	assert.Equal(t, []string{"btn--primary"}, lookup.ConstantParts["BtnPrimary"])
	assert.Equal(t, []string{"nav-group--iconic"}, lookup.ConstantParts["NavGroupIconic"])
}

func TestFindConstantSuggestion(t *testing.T) {
	constants := map[string]string{
		"Btn":        "btn",
		"BtnPrimary": "btn--primary",
		"AppSidebar": "app-sidebar",
	}
	lookup := buildLookupMaps(constants)

	tests := []struct {
		className string
		expected  string
	}{
		{"btn", "Btn"},                 // Exact match (single class)
		{"app-sidebar", "AppSidebar"},  // Exact match (single class)
		{"btn--primary", "BtnPrimary"}, // Modifier match
		{"unknown-class", ""},          // No match
	}

	for _, tt := range tests {
		t.Run(tt.className, func(t *testing.T) {
			result := findConstantSuggestion(tt.className, lookup)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyzeUsage(t *testing.T) {
	constants := map[string]string{
		"Btn":        "btn",
		"BtnPrimary": "btn btn--primary",
		"AppSidebar": "app-sidebar",
		"Unused":     "unused-class",
	}
	lookup := buildLookupMaps(constants)

	references := []ClassReference{
		{FullClassValue: "btn", IsConstant: false, Location: FileLocation{File: "test.templ", Line: 1}},
		{FullClassValue: "btn", IsConstant: false, Location: FileLocation{File: "test.templ", Line: 2}},
		{IsConstant: true, ConstName: "AppSidebar", Location: FileLocation{File: "test.templ", Line: 3}},
	}

	result := analyzeUsage(constants, references, lookup)

	assert.Equal(t, 4, result.TotalConstants)
	assert.Equal(t, 1, result.ActuallyUsed)              // AppSidebar (actually used via ui.AppSidebar)
	assert.Equal(t, 1, result.AvailableForMigration)     // Btn (matches hardcoded "btn")
	assert.Equal(t, 2, result.CompletelyUnused)          // BtnPrimary and Unused
	assert.Equal(t, 2, result.ClassesFound)              // Two hardcoded "btn"
	assert.Equal(t, 1, result.ConstantsFound)            // One ui.AppSidebar
	assert.InDelta(t, 25.0, result.UsagePercentage, 0.1) // 1/4 = 25% (only actually used)

	// Check hardcoded strings
	require.Len(t, result.HardcodedStrings, 2)
	assert.Equal(t, "btn", result.HardcodedStrings[0].FullClassValue)
	assert.Equal(t, []string{"Btn"}, result.HardcodedStrings[0].Suggestion.Constants)

	// Check unused classes
	require.Len(t, result.UnusedClasses, 2)
	unusedNames := []string{result.UnusedClasses[0].ConstName, result.UnusedClasses[1].ConstName}
	assert.Contains(t, unusedNames, "BtnPrimary")
	assert.Contains(t, unusedNames, "Unused")

	// Check quick wins
	require.NotEmpty(t, result.QuickWins.SingleClass)
	assert.Equal(t, "btn", result.QuickWins.SingleClass[0].ClassName)
	assert.Equal(t, 2, result.QuickWins.SingleClass[0].Occurrences)
	assert.Equal(t, "ui.Btn", result.QuickWins.SingleClass[0].Suggestion)
}

func TestScanFile(t *testing.T) {
	content := `package test

import "go-scheduler/internal/web/ui"

// This is a comment with class="ignored"
templ Component() {
	<div class="app-sidebar">
		<button class={ ui.Btn, ui.BtnPrimary }>Click</button>
		<span class={ templ.KV("active", true) }>Active</span>
	</div>
}
`

	tmpfile, err := os.CreateTemp("", "test-*.templ")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	refs, err := scanFile(tmpfile.Name())
	require.NoError(t, err)

	// Should find:
	// - "app-sidebar" (line 7)
	// - ui.Btn (line 8)
	// - ui.BtnPrimary (line 8)
	// - "active" (line 9)
	// Should NOT find comment on line 5

	require.Len(t, refs, 4)

	// Find each reference
	var foundAppSidebar, foundBtn, foundBtnPrimary, foundActive bool
	for _, ref := range refs {
		if ref.FullClassValue == "app-sidebar" && !ref.IsConstant {
			foundAppSidebar = true
		}
		if ref.IsConstant && ref.ConstName == "Btn" {
			foundBtn = true
		}
		if ref.IsConstant && ref.ConstName == "BtnPrimary" {
			foundBtnPrimary = true
		}
		if ref.FullClassValue == "active" && !ref.IsConstant {
			foundActive = true
		}
	}

	assert.True(t, foundAppSidebar, "should find app-sidebar")
	assert.True(t, foundBtn, "should find ui.Btn")
	assert.True(t, foundBtnPrimary, "should find ui.BtnPrimary")
	assert.True(t, foundActive, "should find active")
}

func TestExpandGlobPatterns(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "glob-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := []string{
		"file1.templ",
		"file2.go",
		"subdir/file3.templ",
		"subdir/file4.go",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		require.NoError(t, err)
		err = os.WriteFile(path, []byte("test"), 0644)
		require.NoError(t, err)
	}

	// Test glob pattern
	pattern := filepath.Join(tmpDir, "**/*.templ")
	matches, err := expandGlobPatterns([]string{pattern})
	require.NoError(t, err)

	// Should find file1.templ and subdir/file3.templ
	assert.Len(t, matches, 2)

	// Verify they're all .templ files
	for _, match := range matches {
		assert.True(t, strings.HasSuffix(match, ".templ"))
	}
}

func TestInferLayer(t *testing.T) {
	tests := []struct {
		cssClass string
		expected string
	}{
		{"text-primary", "utilities"},
		{"bg-danger", "utilities"},
		{"flex-row", "utilities"},
		{"grid-cols-3", "utilities"},
		{"btn", "components"},
		{"btn-primary", "components"},
		{"card", "components"},
		{"modal", "components"},
		{"nav-item", "components"},
		{"table-row", "components"},
		{"body", "base"},
		{"container", "base"},
	}

	for _, tt := range tests {
		t.Run(tt.cssClass, func(t *testing.T) {
			result := inferLayer(tt.cssClass)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateQuickWins(t *testing.T) {
	constants := map[string]string{
		"Btn":        "btn",
		"AppSidebar": "app-sidebar",
		"NavItem":    "nav-item",
		"DataTable":  "data-table",
		"Badge":      "badge",
	}
	lookup := buildLookupMaps(constants)

	// Create hardcoded strings with different frequencies
	hardcodedStrings := []HardcodedString{
		{FullClassValue: "btn", Suggestion: ResolveBestConstants("btn", lookup)},
		{FullClassValue: "btn", Suggestion: ResolveBestConstants("btn", lookup)},
		{FullClassValue: "btn", Suggestion: ResolveBestConstants("btn", lookup)},
		{FullClassValue: "data-table", Suggestion: ResolveBestConstants("data-table", lookup)},
		{FullClassValue: "data-table", Suggestion: ResolveBestConstants("data-table", lookup)},
		{FullClassValue: "nav-item", Suggestion: ResolveBestConstants("nav-item", lookup)},
	}

	summary := generateQuickWins(hardcodedStrings)

	// Should be sorted by occurrences (descending)
	require.Len(t, summary.SingleClass, 3)
	assert.Equal(t, "btn", summary.SingleClass[0].ClassName)
	assert.Equal(t, 3, summary.SingleClass[0].Occurrences)
	assert.Equal(t, "data-table", summary.SingleClass[1].ClassName)
	assert.Equal(t, 2, summary.SingleClass[1].Occurrences)
	assert.Equal(t, "nav-item", summary.SingleClass[2].ClassName)
	assert.Equal(t, 1, summary.SingleClass[2].Occurrences)
}

func TestLintEndToEnd(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "lint-e2e-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create generated file
	generatedContent := `package ui

var AllCSSClasses = map[string]bool{
	"btn": true,
	"btn--primary": true,
	"app-sidebar": true,
	"unused": true,
}

const Btn = "btn"
const BtnPrimary = "btn btn--primary"
const AppSidebar = "app-sidebar"
const UnusedClass = "unused"
`
	generatedFile := filepath.Join(tmpDir, "styles.gen.go")
	err = os.WriteFile(generatedFile, []byte(generatedContent), 0644)
	require.NoError(t, err)

	// Create template file with mixed usage
	templContent := `package test

templ Page() {
	<div class="app-sidebar">
		<button class="btn">Click</button>
		<button class={ ui.BtnPrimary }>Submit</button>
	</div>
}
`
	templFile := filepath.Join(tmpDir, "page.templ")
	err = os.WriteFile(templFile, []byte(templContent), 0644)
	require.NoError(t, err)

	// Run linter
	config := LintConfig{
		GeneratedFile: generatedFile,
		PackageName:   "ui",
		ScanPaths:     []string{filepath.Join(tmpDir, "*.templ")},
		Verbose:       false,
	}

	result, err := Lint(config)
	require.NoError(t, err)

	// Verify results
	assert.Equal(t, 4, result.TotalConstants)
	assert.Equal(t, 1, result.ActuallyUsed)              // BtnPrimary (actually used via ui.BtnPrimary)
	assert.Equal(t, 2, result.AvailableForMigration)     // Btn and AppSidebar (match hardcoded strings)
	assert.Equal(t, 1, result.CompletelyUnused)          // UnusedClass
	assert.InDelta(t, 25.0, result.UsagePercentage, 0.1) // 1/4 = 25% (only actually used)

	// Should find 2 hardcoded strings: "app-sidebar", "btn"
	assert.Equal(t, 2, result.ClassesFound)

	// Should find 1 constant reference: ui.BtnPrimary
	assert.Equal(t, 1, result.ConstantsFound)

	// Should have hardcoded suggestions
	require.NotEmpty(t, result.HardcodedStrings)

	// Should have unused classes
	require.Len(t, result.UnusedClasses, 1)
	assert.Equal(t, "UnusedClass", result.UnusedClasses[0].ConstName)
}

func TestResolveBestConstants(t *testing.T) {
	// 1:1 mapping constants
	constants := map[string]string{
		"Btn":         "btn",
		"BtnGhost":    "btn--ghost",
		"BtnSm":       "btn--sm",
		"BtnOutlined": "btn--outlined",
		"Icon":        "icon",
		"Page":        "page",
	}
	lookup := buildLookupMaps(constants)
	lookup.AllCSSClasses = map[string]bool{
		"btn":           true,
		"btn--ghost":    true,
		"btn--sm":       true,
		"btn--outlined": true,
		"icon":          true,
		"page":          true,
	}

	tests := []struct {
		name              string
		input             string
		expectedConstants []string
		expectAnalysis    bool
		expectUnmatched   bool
		expectedUnmatched []string
	}{
		{
			name:              "exact single class",
			input:             "icon",
			expectedConstants: []string{"Icon"},
			expectAnalysis:    false,
			expectUnmatched:   false,
		},
		{
			name:              "multi-class 1:1 mapping",
			input:             "btn btn--ghost btn--sm",
			expectedConstants: []string{"Btn", "BtnGhost", "BtnSm"}, // All three map individually
			expectAnalysis:    true,
			expectUnmatched:   false,
			expectedUnmatched: []string{},
		},
		{
			name:              "base class only",
			input:             "btn",
			expectedConstants: []string{"Btn"},
			expectAnalysis:    false,
			expectUnmatched:   false,
		},
		{
			name:              "modifier without base",
			input:             "btn--ghost",
			expectedConstants: []string{"BtnGhost"},
			expectAnalysis:    false,
			expectUnmatched:   false,
		},
		{
			name:              "no match - invalid class",
			input:             "unknown-class",
			expectedConstants: []string{},
			expectAnalysis:    true,
			expectUnmatched:   true,
			expectedUnmatched: []string{"unknown-class"}, // Invalid (doesn't exist in CSS)
		},
		{
			name:              "partial match - non-existent modifier (invalid)",
			input:             "btn btn--sm btn--outline", // btn--outline doesn't exist in CSS
			expectedConstants: []string{"Btn", "BtnSm"},   // Btn and BtnSm match (1:1)
			expectAnalysis:    true,
			expectUnmatched:   true,
			expectedUnmatched: []string{"btn--outline"}, // Invalid (doesn't exist in CSS)
		},
		{
			name:              "partial match - multiple invalid",
			input:             "btn btn--fake btn--invalid",
			expectedConstants: []string{"Btn"},
			expectAnalysis:    true,
			expectUnmatched:   true,
			expectedUnmatched: []string{"btn--fake", "btn--invalid"}, // Both invalid (don't exist in CSS)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveBestConstants(tt.input, lookup)

			assert.ElementsMatch(t, tt.expectedConstants, result.Constants)

			if tt.expectAnalysis {
				assert.NotNil(t, result.Analysis)
			} else {
				assert.Nil(t, result.Analysis)
			}

			assert.Equal(t, tt.expectUnmatched, result.HasUnmatched, "HasUnmatched mismatch")
			if tt.expectUnmatched {
				assert.ElementsMatch(t, tt.expectedUnmatched, result.UnmatchedClasses, "UnmatchedClasses mismatch")
			}
		})
	}
}

// TestDeduplicateConstants removed - no longer needed with 1:1 mapping
// With 1:1 mapping, each class token maps to exactly one constant,
// so there's nothing to deduplicate.

func TestClassifyClass(t *testing.T) {
	constants := map[string]string{
		"Btn":         "btn",
		"BtnPrimary":  "btn--primary",
		"BtnOutlined": "btn--outlined",
	}
	lookup := buildLookupMaps(constants)
	lookup.AllCSSClasses = map[string]bool{
		"btn":           true,
		"btn--primary":  true,
		"btn--outlined": true,
		"_internal":     true,
	}

	tests := []struct {
		name     string
		class    string
		expected ClassificationResult
	}{
		{
			name:     "matched with constant",
			class:    "btn",
			expected: ClassMatched,
		},
		{
			name:     "matched modifier",
			class:    "btn--primary",
			expected: ClassMatched,
		},
		{
			name:     "bypassed internal class",
			class:    "_internal",
			expected: ClassBypassed,
		},
		{
			name:     "zombie - typo",
			class:    "btn--outline", // Should be "btn--outlined"
			expected: ClassZombie,
		},
		{
			name:     "zombie - non-existent",
			class:    "fake-class",
			expected: ClassZombie,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyClass(tt.class, lookup)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInvalidClassDetection(t *testing.T) {
	constants := map[string]string{
		"Btn":         "btn",
		"BtnOutlined": "btn btn--outlined",
		"BtnSm":       "btn btn--sm",
	}

	allCSSClasses := map[string]bool{
		"btn":           true,
		"btn--outlined": true,
		"btn--sm":       true,
	}

	lookup := buildLookupMaps(constants)
	lookup.AllCSSClasses = allCSSClasses

	// Test typo detection
	result := ResolveBestConstants("btn btn--sm btn--outline", lookup)

	assert.True(t, result.HasInvalid, "Should detect invalid class")
	assert.Equal(t, []string{"btn--outline"}, result.InvalidClasses)
	assert.True(t, result.HasUnmatched)
}

func TestLintWithInvalidClasses(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint-invalid-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create generated file with AllCSSClasses
	generatedContent := `package ui

var AllCSSClasses = map[string]bool{
	"btn": true,
	"btn--outlined": true,
	"btn--sm": true,
}

const Btn = "btn"
const BtnOutlined = "btn btn--outlined"
const BtnSm = "btn btn--sm"
`
	generatedFile := filepath.Join(tmpDir, "styles.gen.go")
	err = os.WriteFile(generatedFile, []byte(generatedContent), 0644)
	require.NoError(t, err)

	// Create template with typo
	templContent := `package test

templ Page() {
	<button class="btn btn--sm btn--outline">Click</button>
}
`
	templFile := filepath.Join(tmpDir, "page.templ")
	err = os.WriteFile(templFile, []byte(templContent), 0644)
	require.NoError(t, err)

	// Run linter
	config := LintConfig{
		GeneratedFile: generatedFile,
		PackageName:   "ui",
		ScanPaths:     []string{filepath.Join(tmpDir, "*.templ")},
	}

	result, err := Lint(config)
	require.NoError(t, err)

	// Verify error detection
	assert.Equal(t, 1, result.ErrorCount)
	require.Len(t, result.InvalidClasses, 1)
	assert.Equal(t, "btn--outline", result.InvalidClasses[0].ClassName)
}

func TestHardcodedClassWarnings(t *testing.T) {
	tests := []struct {
		name             string
		cssContent       string
		templContent     string
		expectedWarnings int
		expectedErrors   int
		checkWarningText string
	}{
		{
			name:       "single hardcoded class with exact match",
			cssContent: `.btn { color: blue; }`,
			templContent: `package test
templ Button() {
	<button class="btn">Click</button>
}`,
			expectedWarnings: 1,
			expectedErrors:   0,
			checkWarningText: `hardcoded CSS class "btn" should use ui.Btn`,
		},
		{
			name: "multi-class hardcoded with BEM",
			cssContent: `.btn { color: blue; }
.btn--primary { background: red; }`,
			templContent: `package test
templ Button() {
	<button class="btn btn--primary">Click</button>
}`,
			expectedWarnings: 1,
			expectedErrors:   0,
			checkWarningText: `should use { ui.Btn, ui.BtnPrimary }`,
		},
		{
			name:       "internal class should not warn",
			cssContent: `._debug { color: red; }`,
			templContent: `package test
templ Debug() {
	<div class="_debug">Test</div>
}`,
			expectedWarnings: 0,
			expectedErrors:   0,
		},
		{
			name: "mixed internal and normal - should not warn",
			cssContent: `.btn { color: blue; }
._internal { display: none; }`,
			templContent: `package test
templ Mixed() {
	<div class="btn _internal">Test</div>
}`,
			expectedWarnings: 0,
			expectedErrors:   0,
		},
		{
			name:       "invalid class should be error not warning",
			cssContent: `.btn { color: blue; }`,
			templContent: `package test
templ Invalid() {
	<button class="btn--invalid">Click</button>
}`,
			expectedWarnings: 0,
			expectedErrors:   1,
		},
		{
			name:       "using ui constant should have no issues",
			cssContent: `.btn { color: blue; }`,
			templContent: `package test
import "go-scheduler/internal/web/ui"
templ Correct() {
	<button class={ ui.Btn }>Click</button>
}`,
			expectedWarnings: 0,
			expectedErrors:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir for test files
			tmpDir := t.TempDir()

			// Write CSS file
			cssFile := filepath.Join(tmpDir, "styles.css")
			err := os.WriteFile(cssFile, []byte(tt.cssContent), 0644)
			require.NoError(t, err)

			// Write templ file
			templFile := filepath.Join(tmpDir, "test.templ")
			err = os.WriteFile(templFile, []byte(tt.templContent), 0644)
			require.NoError(t, err)

			// Generate constants file
			genFile := filepath.Join(tmpDir, "styles.gen.go")
			config := Config{
				SourceDir:   tmpDir,
				OutputDir:   tmpDir,
				PackageName: "ui",
				Includes:    []string{"*.css"},
			}
			_, err = Generate(config)
			require.NoError(t, err)

			// Run linter
			lintConfig := LintConfig{
				ScanPaths:     []string{templFile},
				GeneratedFile: genFile,
				PackageName:   "ui",
				Strict:        true, // Enable warnings for hardcoded classes
			}
			result, err := Lint(lintConfig)
			require.NoError(t, err)

			// Count warnings and errors
			var warnings, errors int
			for _, issue := range result.Issues {
				switch issue.Severity {
				case SeverityWarning:
					warnings++
				case SeverityError:
					errors++
				}
			}

			// Verify counts
			assert.Equal(t, tt.expectedWarnings, warnings,
				"Expected %d warnings, got %d", tt.expectedWarnings, warnings)
			assert.Equal(t, tt.expectedErrors, errors,
				"Expected %d errors, got %d", tt.expectedErrors, errors)

			// Check warning text if expected
			if tt.checkWarningText != "" {
				found := false
				for _, issue := range result.Issues {
					if issue.Severity == SeverityWarning &&
						strings.Contains(issue.Text, tt.checkWarningText) {
						found = true
						break
					}
				}
				assert.True(t, found,
					"Expected to find warning containing: %s", tt.checkWarningText)
			}
		})
	}
}

func TestInternalClassExemption(t *testing.T) {
	tests := []struct {
		name      string
		className string
		expected  bool
	}{
		{"underscore prefix", "_debug", true},
		{"underscore in middle", "nav_item", false},
		{"regular class", "btn", false},
		{"BEM modifier", "btn--primary", false},
		{"multiple underscores", "__private", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isInternalClass(tt.className)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestHasInternalClasses(t *testing.T) {
	tests := []struct {
		name           string
		fullClassValue string
		expected       bool
	}{
		{"all internal", "_debug _private", true},
		{"mixed with internal", "btn _debug", true},
		{"no internal", "btn btn--primary", false},
		{"single internal", "_test", true},
		{"single normal", "btn", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasInternalClasses(tt.fullClassValue)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFormatSuggestion(t *testing.T) {
	tests := []struct {
		name     string
		input    ConstantSuggestion
		expected string
	}{
		{
			name: "exact match single constant",
			input: ConstantSuggestion{
				Constants: []string{"Btn"},
			},
			expected: "ui.Btn",
		},
		{
			name: "multi-class suggestion",
			input: ConstantSuggestion{
				Constants: []string{"BtnGhost", "BtnSm"},
			},
			expected: "{ ui.BtnGhost, ui.BtnSm }",
		},
		{
			name: "no suggestion",
			input: ConstantSuggestion{
				Constants: []string{},
			},
			expected: "(no suggestion)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSuggestion(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
