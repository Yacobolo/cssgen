package cssgen

import (
	"io"
	"os"
)

// DetermineOutputFormat selects the appropriate output format based on flags and environment
func DetermineOutputFormat(formatFlag string, quiet bool) OutputFormat {
	// Explicit -quiet flag wins (exit code only)
	if quiet {
		return OutputIssues // Issues only, will be suppressed in main.go
	}

	// Explicit format flag wins
	if formatFlag != "" {
		switch formatFlag {
		case "issues":
			return OutputIssues
		case "summary":
			return OutputSummary
		case "full":
			return OutputFull
		case "json":
			return OutputJSON
		case "markdown", "md":
			return OutputMarkdown
		default:
			// Invalid format, fall through to auto-detection
		}
	}

	// Smart defaults based on environment
	return DetermineDefaultOutputFormat()
}

// DetermineDefaultOutputFormat returns the default output format
// Following golangci-lint's UX: issues only by default (clean, fast, consistent everywhere)
func DetermineDefaultOutputFormat() OutputFormat {
	return OutputIssues
}

// WriteOutput writes the lint result in the specified format
func WriteOutput(w io.Writer, result *LintResult, format OutputFormat, config LintConfig) {
	// Show progress indicator if we scanned many files (stderr to avoid polluting output)
	if result.FilesScanned > 50 && format != OutputJSON && format != OutputMarkdown {
		os.Stderr.WriteString("🔍 Scanning complete\n")
	}

	switch format {
	case OutputIssues:
		// Issues only (golangci-lint format)
		reporter := NewReporter(w, config)
		reporter.PrintIssues(result.Issues)
		reporter.PrintSummary(*result)

	case OutputSummary:
		// Statistics and Quick Wins only (no individual issues)
		useColors := shouldUseColors(config)
		verboseReporter := NewVerboseReporter(w, useColors)
		verboseReporter.PrintStatistics(*result)
		verboseReporter.PrintAdoptionProgress(*result)
		verboseReporter.PrintQuickWins(*result)
		verboseReporter.PrintWarnings(*result)

	case OutputFull:
		// Everything: issues + statistics + quick wins
		reporter := NewReporter(w, config)
		reporter.PrintIssues(result.Issues)
		reporter.PrintSummary(*result)

		verboseReporter := NewVerboseReporter(w, reporter.UseColors())
		verboseReporter.PrintStatistics(*result)
		verboseReporter.PrintAdoptionProgress(*result)
		verboseReporter.PrintQuickWins(*result)
		verboseReporter.PrintWarnings(*result)

	case OutputJSON:
		// JSON export
		if err := WriteJSON(w, result); err != nil {
			// Log error but don't crash
			os.Stderr.WriteString("Error writing JSON: " + err.Error() + "\n")
		}

	case OutputMarkdown:
		// Markdown report
		if err := WriteMarkdown(w, result); err != nil {
			// Log error but don't crash
			os.Stderr.WriteString("Error writing Markdown: " + err.Error() + "\n")
		}
	}
}
