package cssgen

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Reporter handles formatting and outputting linting results
type Reporter struct {
	w               io.Writer
	useColors       bool
	printLines      bool
	printLinterName bool
}

// NewReporter creates a new reporter with the given configuration
func NewReporter(w io.Writer, config LintConfig) *Reporter {
	return &Reporter{
		w:               w,
		useColors:       shouldUseColors(config),
		printLines:      config.PrintIssuedLines,
		printLinterName: config.PrintLinterName,
	}
}

// shouldUseColors determines if colors should be enabled
func shouldUseColors(config LintConfig) bool {
	// Explicit flag wins
	if config.UseColors {
		return true
	}

	// Check for FORCE_COLOR environment variable (GitHub Actions, etc.)
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// GitHub Actions supports colors
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return true
	}

	// Auto-detect TTY
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}

	return false
}

// PrintIssues outputs issues in golangci-lint format
func (r *Reporter) PrintIssues(issues []Issue) {
	// Sort issues by file, then line, then column
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Pos.Filename != issues[j].Pos.Filename {
			return issues[i].Pos.Filename < issues[j].Pos.Filename
		}
		if issues[i].Pos.Line != issues[j].Pos.Line {
			return issues[i].Pos.Line < issues[j].Pos.Line
		}
		return issues[i].Pos.Column < issues[j].Pos.Column
	})

	// Print each issue
	for _, issue := range issues {
		r.printIssue(issue)
	}
}

// printIssue formats a single issue in golangci-lint style
func (r *Reporter) printIssue(issue Issue) {
	// Format: file:line:col: message (linter)
	location := fmt.Sprintf("%s:%d:%d:", issue.Pos.Filename, issue.Pos.Line, issue.Pos.Column)

	linterSuffix := ""
	if r.printLinterName {
		linterSuffix = fmt.Sprintf(" (%s)", issue.FromLinter)
	}

	// Print main issue line
	fmt.Fprintf(r.w, "%s %s%s\n",
		RenderStyle(StyleCyan, location, r.useColors),
		issue.Text,
		RenderStyle(StyleGray, linterSuffix, r.useColors))

	// Print source lines with caret indicator
	if r.printLines && len(issue.SourceLines) > 0 {
		for _, line := range issue.SourceLines {
			fmt.Fprintf(r.w, "\t%s\n", line)
		}

		// Print caret indicator
		caret := r.buildCaretIndicator(issue.SourceLines[0], issue.Pos.Column)
		fmt.Fprintf(r.w, "\t%s\n", RenderStyle(StyleYellow, caret, r.useColors))
	}
}

// buildCaretIndicator creates the "^" indicator aligned with the column
// CRITICAL: Handles tabs vs spaces correctly for perfect alignment
func (r *Reporter) buildCaretIndicator(sourceLine string, column int) string {
	if column <= 0 {
		return "^"
	}

	// Extract the prefix up to the column (0-based index = column - 1)
	prefixLen := column - 1
	if prefixLen > len(sourceLine) {
		prefixLen = len(sourceLine)
	}

	prefix := sourceLine[:prefixLen]

	// Build padding that matches tabs/spaces in the prefix
	var padding strings.Builder
	for _, ch := range prefix {
		if ch == '\t' {
			padding.WriteRune('\t')
		} else {
			padding.WriteRune(' ')
		}
	}

	return padding.String() + "^"
}

// PrintSummary outputs the issue count summary
func (r *Reporter) PrintSummary(result LintResult) {
	totalIssues := len(result.Issues)
	truncated := result.TruncatedCount

	// Count by severity
	var errors, warnings int
	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		}
	}

	fmt.Fprintln(r.w, "")

	// Show severity breakdown if we have both types
	if errors > 0 && warnings > 0 {
		if truncated > 0 {
			fmt.Fprintf(r.w, "%s (%s, %s; %s truncated):\n",
				pluralizeCount(totalIssues, "issue", "issues"),
				pluralizeCount(errors, "error", "errors"),
				pluralizeCount(warnings, "warning", "warnings"),
				pluralizeCount(truncated, "issue", "issues"))
		} else {
			fmt.Fprintf(r.w, "%s (%s, %s):\n",
				pluralizeCount(totalIssues, "issue", "issues"),
				pluralizeCount(errors, "error", "errors"),
				pluralizeCount(warnings, "warning", "warnings"))
		}
	} else {
		// Only one type of issue or none
		if truncated > 0 {
			fmt.Fprintf(r.w, "%s (%s truncated):\n",
				pluralizeCount(totalIssues, "issue", "issues"),
				pluralizeCount(truncated, "issue", "issues"))
		} else {
			fmt.Fprintf(r.w, "%s:\n", pluralizeCount(totalIssues, "issue", "issues"))
		}
	}

	// Group by linter
	linterCounts := make(map[string]int)
	for _, issue := range result.Issues {
		linterCounts[issue.FromLinter]++
	}

	// Print linter breakdown
	for linter, count := range linterCounts {
		fmt.Fprintf(r.w, "* %s: %d\n", linter, count)
	}

	// Print helpful hint if there are issues
	if totalIssues > 0 {
		fmt.Fprintln(r.w, "")
		fmt.Fprintln(r.w, RenderStyle(StyleGray, "Hint: Run with --output-format full to see statistics and Quick Wins", r.useColors))
	}
}

// pluralizeCount returns a formatted string with count and singular/plural form
func pluralizeCount(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// UseColors returns whether colors are enabled
func (r *Reporter) UseColors() bool {
	return r.useColors
}
