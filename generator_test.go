package cssgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name          string
		css           string
		expectedCount int
		checkClasses  map[string]func(*testing.T, *CSSClass)
	}{
		{
			name:          "base class",
			css:           ".btn { color: red; }",
			expectedCount: 1,
			checkClasses: map[string]func(*testing.T, *CSSClass){
				"btn": func(t *testing.T, c *CSSClass) {
					assert.Equal(t, "btn", c.Name)
					assert.Equal(t, "red", c.Properties["color"])
				},
			},
		},
		{
			name: "bem modifier",
			css: `.btn { color: blue; }
			      .btn--primary { background: red; }`,
			expectedCount: 2,
			checkClasses: map[string]func(*testing.T, *CSSClass){
				"btn": func(t *testing.T, c *CSSClass) {
					assert.Equal(t, "btn", c.Name)
					assert.Equal(t, "blue", c.Properties["color"])
				},
				"btn--primary": func(t *testing.T, c *CSSClass) {
					assert.Equal(t, "btn--primary", c.Name)
					assert.Equal(t, "red", c.Properties["background"])
				},
			},
		},
		{
			name: "layer tracking",
			css: `@layer components {
				.card { padding: 1rem; }
			}`,
			expectedCount: 1,
			checkClasses: map[string]func(*testing.T, *CSSClass){
				"card": func(t *testing.T, c *CSSClass) {
					assert.Equal(t, "components", c.Layer)
				},
			},
		},
		{
			name:          "pseudo-states",
			css:           `.btn:hover { color: red; }`,
			expectedCount: 1,
			checkClasses: map[string]func(*testing.T, *CSSClass){
				"btn": func(t *testing.T, c *CSSClass) {
					assert.Contains(t, c.PseudoStates, ":hover")
				},
			},
		},
		{
			name:          "internal class",
			css:           `._internal { display: block; }`,
			expectedCount: 1,
			checkClasses: map[string]func(*testing.T, *CSSClass){
				"_internal": func(t *testing.T, c *CSSClass) {
					assert.True(t, c.IsInternal)
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				LayerInferFromPath: false,
				ExtractIntent:      false,
			}
			classes, err := ParseCSS(tt.css, "test.css", "", config)
			require.NoError(t, err)
			assert.Len(t, classes, tt.expectedCount)

			// Build map for easy lookup
			classMap := make(map[string]*CSSClass)
			for _, c := range classes {
				classMap[c.Name] = c
			}

			// Run custom checks
			for name, checkFn := range tt.checkClasses {
				class, exists := classMap[name]
				require.True(t, exists, "class %s not found", name)
				checkFn(t, class)
			}
		})
	}
}

func TestAnalyzer(t *testing.T) {
	tests := []struct {
		name       string
		input      []*CSSClass
		expectedGo map[string]string // Name -> GoName
	}{
		{
			name: "BEM modifier - 1:1 mapping",
			input: []*CSSClass{
				{Name: "btn"},
				{Name: "btn--primary"},
			},
			expectedGo: map[string]string{
				"btn":          "Btn",
				"btn--primary": "BtnPrimary",
			},
		},
		{
			name: "BEM element - 1:1 mapping",
			input: []*CSSClass{
				{Name: "card"},
				{Name: "card__header"},
			},
			expectedGo: map[string]string{
				"card":         "Card",
				"card__header": "CardHeader",
			},
		},
		{
			name: "utility class",
			input: []*CSSClass{
				{Name: "flex-center"},
			},
			expectedGo: map[string]string{
				"flex-center": "FlexCenter",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AnalyzeClasses(tt.input)
			require.NoError(t, err)

			for _, class := range tt.input {
				expectedGo := tt.expectedGo[class.Name]
				assert.Equal(t, expectedGo, class.GoName, "GoName for %s", class.Name)

				// With 1:1 mapping, we don't have FullClasses field anymore
				// Constants use class.Name directly
			}
		})
	}
}

func TestEndToEnd(t *testing.T) {
	// Create temp CSS file
	cssContent := `@layer components {
		.btn { color: red; }
		.btn--primary { background: blue; }
	}`

	tmpDir := t.TempDir()
	cssFile := filepath.Join(tmpDir, "test.css")
	require.NoError(t, os.WriteFile(cssFile, []byte(cssContent), 0644))

	// Generate
	outputFile := filepath.Join(tmpDir, "styles.gen.go")
	config := Config{
		SourceDir:          tmpDir,
		OutputDir:          tmpDir,
		PackageName:        "ui",
		Includes:           []string{"*.css"},
		LayerInferFromPath: true,
		ExtractIntent:      false,
		Format:             "markdown",
		PropertyLimit:      5,
	}

	result, err := Generate(config)
	require.NoError(t, err)
	assert.Equal(t, 1, result.FilesScanned)
	assert.Equal(t, 2, result.ClassesGenerated)

	// Verify output file exists
	_, err = os.Stat(outputFile)
	require.NoError(t, err)

	// Verify output content (main file has AllCSSClasses map)
	output, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	outputStr := string(output)
	assert.Contains(t, outputStr, "package ui")
	assert.Contains(t, outputStr, `"btn": true`)
	assert.Contains(t, outputStr, `"btn--primary": true`)

	// Constants are in split file (styles_test.gen.go)
	splitFile := filepath.Join(tmpDir, "styles_test.gen.go")
	splitOutput, err := os.ReadFile(splitFile)
	require.NoError(t, err)

	splitStr := string(splitOutput)
	assert.Contains(t, splitStr, "package ui")
	assert.Contains(t, splitStr, "const Btn =")
	assert.Contains(t, splitStr, "const BtnPrimary =")
	assert.Contains(t, splitStr, `Btn = "btn"`)
	assert.Contains(t, splitStr, `BtnPrimary = "btn--primary"`) // 1:1 mapping
	assert.Contains(t, splitStr, "@layer components")
}

func TestBEMDetection(t *testing.T) {
	tests := []struct {
		name      string
		className string
		wantBase  string
		wantMod   bool
	}{
		{"modifier", "btn--primary", "btn", true},
		{"element", "card__header", "card", true},
		{"base", "btn", "", false},
		{"utility", "flex-center", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, isMod := detectBEMPattern(tt.className)
			assert.Equal(t, tt.wantBase, base)
			assert.Equal(t, tt.wantMod, isMod)
		})
	}
}

func TestToGoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"btn", "Btn"},
		{"btn--primary", "BtnPrimary"},
		{"card__header", "CardHeader"},
		{"flex-center", "FlexCenter"},
		{"_internal", "_Internal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toGoName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeConflicts(t *testing.T) {
	classes := []*CSSClass{
		{
			Name:       "btn",
			Properties: map[string]string{"color": "red"},
			SourceFile: "file1.css",
		},
		{
			Name:       "btn",
			Properties: map[string]string{"background": "blue"},
			SourceFile: "file2.css",
		},
	}

	merged, warnings := mergeConflicts(classes)

	assert.Len(t, merged, 1)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "Duplicate class")

	// Properties should be merged
	btn := merged[0]
	assert.Equal(t, "red", btn.Properties["color"])
	assert.Equal(t, "blue", btn.Properties["background"])
}

func TestParserWithTestdata(t *testing.T) {
	testFiles := []struct {
		file          string
		expectedCount int
	}{
		{"simple.css", 1},
		{"bem.css", 3},
		{"layers.css", 1},
		{"complex.css", 4}, // btn, btn--primary, btn--secondary, _internal-base
	}

	for _, tt := range testFiles {
		t.Run(tt.file, func(t *testing.T) {
			path := filepath.Join("testdata", tt.file)
			config := Config{
				SourceDir:          "testdata",
				LayerInferFromPath: false,
				ExtractIntent:      false,
			}
			classes, err := parseFile(path, config)
			require.NoError(t, err)
			assert.Len(t, classes, tt.expectedCount, "Expected %d classes in %s", tt.expectedCount, tt.file)
		})
	}
}

func TestLayerInference(t *testing.T) {
	tests := []struct {
		path      string
		sourceDir string
		expected  string
	}{
		{"web/ui/src/styles/layers/components/badge.css", "web/ui/src/styles", "components"},
		{"web/ui/src/styles/layers/utilities.css", "web/ui/src/styles", "utilities"},
		{"web/ui/src/styles/layers/base.css", "web/ui/src/styles", "base"},
		{"web/ui/src/styles/reset.css", "web/ui/src/styles", "reset"},
		{"web/ui/src/custom.css", "web/ui/src/styles", "n/a"},
		{"layers/components/forms/input.css", ".", "components"},
		{"/absolute/path/layers/tokens/colors.css", "/absolute/path", "tokens"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := inferLayerFromPath(tt.path, tt.sourceDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLayerInferenceWindowsPaths(t *testing.T) {
	// Test that Windows-style paths (with backslashes) work correctly
	// filepath.ToSlash should convert them
	tests := []struct {
		name      string
		path      string
		sourceDir string
		expected  string
	}{
		{
			name:      "windows components path",
			path:      `C:\projects\web\ui\src\styles\layers\components\badge.css`,
			sourceDir: `C:\projects\web\ui\src\styles`,
			expected:  "components",
		},
		{
			name:      "windows utilities path",
			path:      `C:\projects\web\ui\src\styles\layers\utilities.css`,
			sourceDir: `C:\projects\web\ui\src\styles`,
			expected:  "utilities",
		},
		{
			name:      "windows with subdirectory",
			path:      `C:\code\styles\layers\components\forms\input.css`,
			sourceDir: `C:\code\styles`,
			expected:  "components",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferLayerFromPath(tt.path, tt.sourceDir)
			assert.Equal(t, tt.expected, result, "Path: %s, SourceDir: %s", tt.path, tt.sourceDir)
		})
	}
}

func TestPropertyCategorization(t *testing.T) {
	tests := []struct {
		property string
		expected PropertyCategory
	}{
		{"background", CategoryVisual},
		{"color", CategoryVisual},
		{"border-radius", CategoryVisual},
		{"display", CategoryLayout},
		{"flex-direction", CategoryLayout},
		{"padding", CategoryLayout},
		{"margin-top", CategoryLayout},
		{"font-size", CategoryTypography},
		{"line-height", CategoryTypography},
		{"transition", CategoryEffects},
		{"transform", CategoryEffects},
		{"-webkit-font-smoothing", CategoryInternal},
		{"grid-template-columns", CategoryLayout},
		{"border-top-color", CategoryVisual},
		{"flex-grow", CategoryLayout},
	}

	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			result := categorizeProperty(tt.property)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenDetection(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"var(--ui-color-primary)", true},
		{"var(--ui-space-md)", true},
		{"#ff0000", false},
		{"1rem", false},
		{"var(--custom)", false}, // Not --ui-*
		{"rgba(0,0,0,0.5)", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := isTokenValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPropertyDiffing(t *testing.T) {
	base := &CSSClass{
		Name: "btn",
		Properties: map[string]string{
			"display":    "inline-flex",
			"padding":    "1rem",
			"background": "transparent",
		},
	}

	modifier := &CSSClass{
		Name: "btn--primary",
		Properties: map[string]string{
			"display":    "inline-flex", // unchanged
			"padding":    "1rem",        // unchanged
			"background": "blue",        // changed
			"color":      "white",       // added
		},
		ParentClass: base,
	}

	diff := DiffProperties(modifier, base)

	assert.Len(t, diff.Changed, 1)
	assert.Equal(t, "blue", diff.Changed["background"])

	assert.Len(t, diff.Added, 1)
	assert.Equal(t, "white", diff.Added["color"])

	assert.Len(t, diff.Unchanged, 2)
	assert.Contains(t, diff.Unchanged, "display")
	assert.Contains(t, diff.Unchanged, "padding")
}

func TestPseudoStateProperties(t *testing.T) {
	css := `
.btn {
	background: transparent;
	color: black;
}

.btn:hover {
	background: blue;
	/* color stays the same */
}

.btn:focus {
	outline: 2px solid blue;
}
`

	config := Config{}
	classes, err := ParseCSS(css, "test.css", "components", config)
	require.NoError(t, err)

	// Find btn class
	var btn *CSSClass
	for _, c := range classes {
		if c.Name == "btn" {
			btn = c
			break
		}
	}
	require.NotNil(t, btn)

	// Should have 2 pseudo-state property sets
	assert.Len(t, btn.PseudoStateProperties, 2)

	// Check :hover changes
	var hoverProps *PseudoStateProperties
	for i := range btn.PseudoStateProperties {
		if btn.PseudoStateProperties[i].PseudoState == ":hover" {
			hoverProps = &btn.PseudoStateProperties[i]
			break
		}
	}
	require.NotNil(t, hoverProps)
	assert.Equal(t, "blue", hoverProps.Changes["background"])
	assert.NotContains(t, hoverProps.Changes, "color") // unchanged

	// Check :focus changes
	var focusProps *PseudoStateProperties
	for i := range btn.PseudoStateProperties {
		if btn.PseudoStateProperties[i].PseudoState == ":focus" {
			focusProps = &btn.PseudoStateProperties[i]
			break
		}
	}
	require.NotNil(t, focusProps)
	assert.Equal(t, "2px solid blue", focusProps.Changes["outline"])
}

func TestIntentExtraction(t *testing.T) {
	css := `
/* @intent Primary brand badge for key status indicators */
.badge--brand {
	background: var(--ui-color-primary);
}

/* Regular comment, no intent */
.badge--secondary {
	background: var(--ui-color-secondary);
}

// @intent Inline comment style
.badge--info {
	background: var(--ui-color-info);
}
`

	config := Config{ExtractIntent: true}
	classes, err := ParseCSS(css, "test.css", "components", config)
	require.NoError(t, err)

	// Build map for easy lookup
	classMap := make(map[string]*CSSClass)
	for _, c := range classes {
		classMap[c.Name] = c
	}

	// Test brand badge
	brandBadge, exists := classMap["badge--brand"]
	require.True(t, exists)
	assert.Equal(t, "Primary brand badge for key status indicators", brandBadge.Intent)

	// Test secondary badge (no intent)
	secondaryBadge, exists := classMap["badge--secondary"]
	require.True(t, exists)
	assert.Empty(t, secondaryBadge.Intent)

	// Test info badge (inline comment style)
	infoBadge, exists := classMap["badge--info"]
	require.True(t, exists)
	assert.Equal(t, "Inline comment style", infoBadge.Intent)
}

// TestCompoundSelectors tests extraction of classes from compound selectors (.foo.bar)
func TestCompoundSelectors(t *testing.T) {
	tests := []struct {
		name        string
		css         string
		wantClasses []string
	}{
		{
			name:        "compound selector - two classes",
			css:         `.nav-item--with-icon.nav-item--active { background: red; }`,
			wantClasses: []string{"nav-item--with-icon", "nav-item--active"},
		},
		{
			name:        "compound selector - three classes",
			css:         `.foo.bar.baz { color: blue; }`,
			wantClasses: []string{"foo", "bar", "baz"},
		},
		{
			name:        "compound with pseudo-class",
			css:         `.nav-item--with-icon.nav-item--active:hover { background: blue; }`,
			wantClasses: []string{"nav-item--with-icon", "nav-item--active"},
		},
		{
			name:        "mobile modifier compound",
			css:         `.app-sidebar--mobile.app-sidebar--open { transform: translateX(0); }`,
			wantClasses: []string{"app-sidebar--mobile", "app-sidebar--open"},
		},
		{
			name:        "badge dot compound",
			css:         `.badge--dot.badge--success { color: green; }`,
			wantClasses: []string{"badge--dot", "badge--success"},
		},
		{
			name:        "comma-separated with compound",
			css:         `.foo.bar, .baz.qux { color: red; }`,
			wantClasses: []string{"foo", "bar", "baz", "qux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{ExtractIntent: false}
			classes, err := ParseCSS(tt.css, "test.css", "", config)
			require.NoError(t, err)

			got := make([]string, len(classes))
			for i, c := range classes {
				got[i] = c.Name
			}

			// Sort for consistent comparison
			require.ElementsMatch(t, tt.wantClasses, got, "Extracted classes don't match expected")
		})
	}
}

// TestFunctionalPseudoClasses tests extraction from :not(), :is(), :where()
func TestFunctionalPseudoClasses(t *testing.T) {
	tests := []struct {
		name        string
		css         string
		wantClasses []string
	}{
		{
			name:        "functional pseudo :not()",
			css:         `.nav-group:not(.nav-group--iconic) { display: none; }`,
			wantClasses: []string{"nav-group", "nav-group--iconic"},
		},
		{
			name:        "functional pseudo :is() with single class",
			css:         `.item:is(.active) { border: 1px solid; }`,
			wantClasses: []string{"item", "active"},
		},
		{
			name:        "functional pseudo :is() with multiple classes",
			css:         `.item:is(.active, .focused) { border: 1px solid; }`,
			wantClasses: []string{"item", "active", "focused"},
		},
		{
			name:        "functional pseudo :where()",
			css:         `.btn:where(.primary, .secondary) { padding: 1rem; }`,
			wantClasses: []string{"btn", "primary", "secondary"},
		},
		{
			name:        "complex :not() with pseudo-state",
			css:         `.btn:hover:not(:disabled) { color: red; }`,
			wantClasses: []string{"btn"}, // :disabled is not a class
		},
		{
			name:        "nested functional pseudo",
			css:         `.app-sidebar--collapsed .nav-group:not(.nav-group--iconic) .nav-group-items { display: none; }`,
			wantClasses: []string{"app-sidebar--collapsed", "nav-group", "nav-group--iconic", "nav-group-items"},
		},
		{
			name:        "data table sort icons",
			css:         `.data-table__sort-icon--asc:not(.data-table__sort-icon--active) { opacity: 0.5; }`,
			wantClasses: []string{"data-table__sort-icon--asc", "data-table__sort-icon--active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{ExtractIntent: false}
			classes, err := ParseCSS(tt.css, "test.css", "", config)
			require.NoError(t, err)

			got := make([]string, len(classes))
			for i, c := range classes {
				got[i] = c.Name
			}

			require.ElementsMatch(t, tt.wantClasses, got, "Extracted classes don't match expected")
		})
	}
}

// TestRealWorldCSS tests with actual CSS patterns from the project
func TestRealWorldCSS(t *testing.T) {
	css := `
	.nav-item--with-icon.nav-item--active {
		background: var(--ui-color-secondary-container);
		color: var(--ui-color-secondary-container-on);
	}

	.nav-item--with-icon.nav-item--active:hover {
		background: var(--ui-color-secondary-container);
		color: var(--ui-color-secondary-container-on);
	}

	.nav-item--compact.nav-item--active {
		border-inline-start-color: var(--ui-color-primary);
		color: var(--ui-color-primary);
	}

	.app-sidebar--mobile.app-sidebar--open {
		transform: translateX(0);
	}

	.app-sidebar--collapsed .nav-group:not(.nav-group--iconic) .nav-group-items {
		display: none;
	}

	.badge--dot.badge--success {
		background: var(--ui-color-success);
	}

	.btn:hover:not(:disabled) {
		background: var(--ui-color-surface-container-high);
	}
	`

	config := Config{ExtractIntent: false}
	classes, err := ParseCSS(css, "sidebar.css", "components", config)
	require.NoError(t, err)

	classNames := make([]string, len(classes))
	for i, c := range classes {
		classNames[i] = c.Name
	}

	// Critical compound selector classes that were previously missed
	assert.Contains(t, classNames, "nav-item--with-icon", "Should extract first class from compound selector")
	assert.Contains(t, classNames, "nav-item--active", "Should extract second class from compound selector")
	assert.Contains(t, classNames, "nav-item--compact", "Should extract from different compound selector")
	assert.Contains(t, classNames, "app-sidebar--mobile", "Should extract mobile modifier")
	assert.Contains(t, classNames, "app-sidebar--open", "Should extract open state")
	assert.Contains(t, classNames, "app-sidebar--collapsed", "Should extract from descendant selector")
	assert.Contains(t, classNames, "nav-group", "Should extract base class")
	assert.Contains(t, classNames, "nav-group--iconic", "Should extract class from :not()")
	assert.Contains(t, classNames, "nav-group-items", "Should extract child class")
	assert.Contains(t, classNames, "badge--dot", "Should extract badge modifier")
	assert.Contains(t, classNames, "badge--success", "Should extract badge variant")
	assert.Contains(t, classNames, "btn", "Should extract button base")

	// Verify we got all expected classes (should be 12 unique classes)
	// Note: Some classes appear multiple times in CSS but should only be extracted once
	assert.GreaterOrEqual(t, len(classNames), 12, "Should extract at least 12 unique classes from real-world CSS")
}
