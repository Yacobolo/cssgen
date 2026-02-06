package cssgen

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetermineOutputFormat(t *testing.T) {
	tests := []struct {
		name       string
		formatFlag string
		quiet      bool
		envVars    map[string]string
		expected   OutputFormat
	}{
		{
			name:       "explicit quiet flag",
			formatFlag: "",
			quiet:      true,
			expected:   OutputIssues,
		},
		{
			name:       "explicit issues format",
			formatFlag: "issues",
			quiet:      false,
			expected:   OutputIssues,
		},
		{
			name:       "explicit summary format",
			formatFlag: "summary",
			quiet:      false,
			expected:   OutputSummary,
		},
		{
			name:       "explicit full format",
			formatFlag: "full",
			quiet:      false,
			expected:   OutputFull,
		},
		{
			name:       "explicit json format",
			formatFlag: "json",
			quiet:      false,
			expected:   OutputJSON,
		},
		{
			name:       "explicit markdown format",
			formatFlag: "markdown",
			quiet:      false,
			expected:   OutputMarkdown,
		},
		{
			name:       "markdown shorthand (md)",
			formatFlag: "md",
			quiet:      false,
			expected:   OutputMarkdown,
		},
		{
			name:       "default format is issues (no auto-detection)",
			formatFlag: "",
			quiet:      false,
			envVars:    map[string]string{},
			expected:   OutputIssues,
		},
		{
			name:       "quiet overrides format flag",
			formatFlag: "full",
			quiet:      true,
			expected:   OutputIssues,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			result := DetermineOutputFormat(tt.formatFlag, tt.quiet)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteJSON(t *testing.T) {
	// Create a sample lint result
	result := &LintResult{
		TotalConstants:        100,
		ActuallyUsed:          20,
		AvailableForMigration: 30,
		CompletelyUnused:      50,
		UsagePercentage:       20.0,
		FilesScanned:          10,
		ClassesFound:          150,
		ConstantsFound:        20,
		ErrorCount:            5,
		Issues: []Issue{
			{
				FromLinter:  "csslint",
				Text:        "invalid CSS class \"foo\" not found in stylesheet",
				Severity:    SeverityError,
				SourceLines: []string{`<div class="foo">`},
				Pos: IssuePos{
					Filename: "test.templ",
					Line:     10,
					Column:   12,
				},
			},
			{
				FromLinter:  "csslint",
				Text:        "hardcoded CSS class \"bar\" should use ui.Bar constant",
				Severity:    SeverityWarning,
				SourceLines: []string{`<div class="bar">`},
				Pos: IssuePos{
					Filename: "test.templ",
					Line:     20,
					Column:   12,
				},
			},
		},
		QuickWins: QuickWinsSummary{
			SingleClass: []QuickWin{
				{ClassName: "btn", Occurrences: 45, Suggestion: "ui.Btn"},
			},
			MultiClass: []QuickWin{
				{ClassName: "btn btn--brand", Occurrences: 10, Suggestion: "{ ui.Btn, ui.BtnBrand }"},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteJSON(&buf, result)
	require.NoError(t, err)

	// Parse JSON to verify structure
	var output JSONOutput
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "1.0", output.Version)
	assert.NotEmpty(t, output.Timestamp)

	// Verify summary
	assert.Equal(t, 2, output.Summary.TotalIssues)
	assert.Equal(t, 1, output.Summary.Errors)
	assert.Equal(t, 1, output.Summary.Warnings)
	assert.Equal(t, 10, output.Summary.FilesScanned)

	// Verify stats
	assert.Equal(t, 100, output.Stats.TotalConstants)
	assert.Equal(t, 20, output.Stats.ActuallyUsed)
	assert.Equal(t, 30, output.Stats.MigrationOpportunities)
	assert.Equal(t, 50, output.Stats.CompletelyUnused)
	assert.InDelta(t, 20.0, output.Stats.UsagePercentage, 0.01)
	assert.Equal(t, 150, output.Stats.HardcodedClasses)
	assert.Equal(t, 20, output.Stats.ConstantReferences)

	// Verify issues
	require.Len(t, output.Issues, 2)
	assert.Equal(t, "test.templ", output.Issues[0].File)
	assert.Equal(t, 10, output.Issues[0].Line)
	assert.Equal(t, 12, output.Issues[0].Column)
	assert.Equal(t, "error", output.Issues[0].Severity)
	assert.Equal(t, "csslint", output.Issues[0].Linter)
	assert.Contains(t, output.Issues[0].Source, "foo")

	// Verify quick wins
	require.Len(t, output.QuickWins.SingleClass, 1)
	assert.Equal(t, "btn", output.QuickWins.SingleClass[0].Class)
	assert.Equal(t, 45, output.QuickWins.SingleClass[0].Occurrences)
	assert.Equal(t, "ui.Btn", output.QuickWins.SingleClass[0].Suggestion)

	require.Len(t, output.QuickWins.MultiClass, 1)
	assert.Equal(t, "btn btn--brand", output.QuickWins.MultiClass[0].Class)
	assert.Equal(t, 10, output.QuickWins.MultiClass[0].Occurrences)
}

func TestWriteMarkdown(t *testing.T) {
	result := &LintResult{
		TotalConstants:        100,
		ActuallyUsed:          80,
		AvailableForMigration: 10,
		CompletelyUnused:      10,
		UsagePercentage:       80.0,
		FilesScanned:          10,
		ClassesFound:          150,
		ConstantsFound:        20,
		ErrorCount:            0,
		Issues: []Issue{
			{
				Severity: SeverityWarning,
				Text:     `hardcoded CSS class "btn" should use ui.Btn constant`,
			},
			{
				Severity: SeverityWarning,
				Text:     `hardcoded CSS class "icon" should use ui.Icon constant`,
			},
		},
		QuickWins: QuickWinsSummary{
			SingleClass: []QuickWin{
				{ClassName: "icon", Occurrences: 28, Suggestion: "ui.Icon"},
				{ClassName: "flex-fill", Occurrences: 6, Suggestion: "ui.FlexFill"},
			},
			MultiClass: []QuickWin{
				{ClassName: "btn btn--ghost", Occurrences: 5, Suggestion: "{ ui.Btn, ui.BtnGhost }"},
			},
		},
		Suggestions: []string{
			"Import the ui package",
			"Replace hardcoded strings",
		},
	}

	var buf bytes.Buffer
	err := WriteMarkdown(&buf, result)
	require.NoError(t, err)

	markdown := buf.String()

	// Verify markdown structure
	assert.Contains(t, markdown, "# CSS Linter Report")
	assert.Contains(t, markdown, "## Executive Summary")
	assert.Contains(t, markdown, "## 🎯 Quick Wins")
	assert.Contains(t, markdown, "## 📊 Detailed Statistics")
	assert.Contains(t, markdown, "## ✅ Recommendations")
	// No errors section when there are no errors
	assert.NotContains(t, markdown, "## ❌ Errors")

	// Verify content
	assert.Contains(t, markdown, "**Total Issues** | 2 (0 errors, 2 warnings)")
	assert.Contains(t, markdown, "**Files Scanned** | 10")
	assert.Contains(t, markdown, "**Adoption Rate** | 80.0%")
	assert.Contains(t, markdown, "**Constants Used** | 80 / 100")

	// Verify quick wins table
	assert.Contains(t, markdown, "`icon` | 28 | `ui.Icon`")
	assert.Contains(t, markdown, "`flex-fill` | 6 | `ui.FlexFill`")
	assert.Contains(t, markdown, "`btn btn--ghost` | 5")

	// Verify status badge (80% + no errors = Excellent)
	assert.Contains(t, markdown, "🟢 Excellent")

	// Verify recommendations
	assert.Contains(t, markdown, "Import the ui package")
	assert.Contains(t, markdown, "Replace hardcoded strings")

	// Verify footer
	assert.Contains(t, markdown, "*Generated by cssgen linter v1.0*")
}

func TestMarkdownStatusBadges(t *testing.T) {
	tests := []struct {
		name            string
		errorCount      int
		usagePercentage float64
		expectedStatus  string
		expectedEmoji   string
	}{
		{
			name:            "excellent (no errors, 80%+)",
			errorCount:      0,
			usagePercentage: 85.0,
			expectedStatus:  "🟢 Excellent",
		},
		{
			name:            "good progress (no errors, 50-79%)",
			errorCount:      0,
			usagePercentage: 65.0,
			expectedStatus:  "🟡 Good Progress",
		},
		{
			name:            "needs attention (errors present)",
			errorCount:      5,
			usagePercentage: 90.0,
			expectedStatus:  "🔴 Needs Attention",
		},
		{
			name:            "needs attention (low adoption)",
			errorCount:      0,
			usagePercentage: 20.0,
			expectedStatus:  "🔴 Needs Attention",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &LintResult{
				ErrorCount:      tt.errorCount,
				UsagePercentage: tt.usagePercentage,
				TotalConstants:  100,
				ActuallyUsed:    int(tt.usagePercentage),
				FilesScanned:    10,
			}

			var buf bytes.Buffer
			err := WriteMarkdown(&buf, result)
			require.NoError(t, err)

			markdown := buf.String()
			assert.Contains(t, markdown, tt.expectedStatus)
		})
	}
}

func TestWriteOutput_AllFormats(t *testing.T) {
	result := &LintResult{
		TotalConstants:  100,
		ActuallyUsed:    20,
		UsagePercentage: 20.0,
		FilesScanned:    10,
		ErrorCount:      1,
		Issues: []Issue{
			{
				FromLinter:  "csslint",
				Text:        "test issue",
				Severity:    SeverityError,
				SourceLines: []string{"test line"},
				Pos: IssuePos{
					Filename: "test.templ",
					Line:     1,
					Column:   1,
				},
			},
		},
		QuickWins: QuickWinsSummary{
			SingleClass: []QuickWin{
				{ClassName: "btn", Occurrences: 10, Suggestion: "ui.Btn"},
			},
		},
	}

	config := LintConfig{
		PrintIssuedLines: true,
		PrintLinterName:  true,
		UseColors:        false,
	}

	tests := []struct {
		name           string
		format         OutputFormat
		expectedInside []string
	}{
		{
			name:   "issues format",
			format: OutputIssues,
			expectedInside: []string{
				"test.templ:1:1:",
				"test issue",
				"1 issue",
			},
		},
		{
			name:   "summary format",
			format: OutputSummary,
			expectedInside: []string{
				"CSS Linter Statistics",
				"Total Constants:",
				"Quick Wins",
			},
		},
		{
			name:   "full format",
			format: OutputFull,
			expectedInside: []string{
				"test.templ:1:1:",
				"1 issue",
				"CSS Linter Statistics",
				"Quick Wins",
			},
		},
		{
			name:   "json format",
			format: OutputJSON,
			expectedInside: []string{
				`"version"`,
				`"summary"`,
				`"stats"`,
				`"issues"`,
			},
		},
		{
			name:   "markdown format",
			format: OutputMarkdown,
			expectedInside: []string{
				"# CSS Linter Report",
				"## Executive Summary",
				"## 🎯 Quick Wins",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			WriteOutput(&buf, result, tt.format, config)

			output := buf.String()
			for _, expected := range tt.expectedInside {
				assert.Contains(t, output, expected,
					"Format %s should contain %q", tt.format, expected)
			}
		})
	}
}

func TestExtractClassNameFromMessage(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{
			message:  `invalid CSS class "btn--outline" not found in stylesheet`,
			expected: "btn--outline",
		},
		{
			message:  `hardcoded CSS class "data-table" should use ui.DataTable constant`,
			expected: "data-table",
		},
		{
			message:  "no quotes in this message",
			expected: "",
		},
		{
			message:  `only "one quote`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := extractClassNameFromMessage(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJSONOutputSchema(t *testing.T) {
	// Verify JSON output follows the schema exactly
	result := &LintResult{
		TotalConstants:        232,
		ActuallyUsed:          0,
		AvailableForMigration: 118,
		CompletelyUnused:      114,
		UsagePercentage:       0.0,
		FilesScanned:          13,
		ClassesFound:          225,
		ConstantsFound:        0,
		ErrorCount:            12,
		Issues:                []Issue{},
		QuickWins: QuickWinsSummary{
			SingleClass: []QuickWin{},
			MultiClass:  []QuickWin{},
		},
	}

	var buf bytes.Buffer
	err := WriteJSON(&buf, result)
	require.NoError(t, err)

	// Parse and verify all required fields exist
	var output map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Top-level fields
	assert.Contains(t, output, "version")
	assert.Contains(t, output, "timestamp")
	assert.Contains(t, output, "summary")
	assert.Contains(t, output, "stats")
	assert.Contains(t, output, "issues")
	assert.Contains(t, output, "quick_wins")

	// Summary fields
	summary := output["summary"].(map[string]interface{})
	assert.Contains(t, summary, "total_issues")
	assert.Contains(t, summary, "errors")
	assert.Contains(t, summary, "warnings")
	assert.Contains(t, summary, "files_scanned")

	// Stats fields
	stats := output["stats"].(map[string]interface{})
	assert.Contains(t, stats, "total_constants")
	assert.Contains(t, stats, "actually_used")
	assert.Contains(t, stats, "migration_opportunities")
	assert.Contains(t, stats, "completely_unused")
	assert.Contains(t, stats, "usage_percentage")
	assert.Contains(t, stats, "hardcoded_classes")
	assert.Contains(t, stats, "constant_references")

	// Quick wins fields
	quickWins := output["quick_wins"].(map[string]interface{})
	assert.Contains(t, quickWins, "single_class")
	assert.Contains(t, quickWins, "multi_class")
}

func TestMarkdownEscaping(t *testing.T) {
	// Verify markdown properly escapes pipe characters in suggestions
	result := &LintResult{
		TotalConstants:  100,
		ActuallyUsed:    20,
		UsagePercentage: 20.0,
		FilesScanned:    10,
		QuickWins: QuickWinsSummary{
			MultiClass: []QuickWin{
				{
					ClassName:   "btn btn--brand",
					Occurrences: 5,
					Suggestion:  "{ ui.Btn | ui.BtnBrand }", // Contains pipe
				},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteMarkdown(&buf, result)
	require.NoError(t, err)

	markdown := buf.String()

	// Verify pipes are escaped
	assert.Contains(t, markdown, "\\|", "Pipes should be escaped in markdown tables")
}
