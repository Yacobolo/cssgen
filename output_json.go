package cssgen

import (
	"encoding/json"
	"io"
	"time"
)

// JSONOutput represents the structured JSON export schema
type JSONOutput struct {
	Version   string        `json:"version"`
	Timestamp string        `json:"timestamp"`
	Summary   JSONSummary   `json:"summary"`
	Stats     JSONStats     `json:"stats"`
	Issues    []JSONIssue   `json:"issues"`
	QuickWins JSONQuickWins `json:"quick_wins"`
}

// JSONSummary contains high-level issue counts
type JSONSummary struct {
	TotalIssues  int `json:"total_issues"`
	Errors       int `json:"errors"`
	Warnings     int `json:"warnings"`
	FilesScanned int `json:"files_scanned"`
}

// JSONStats contains adoption and usage statistics
type JSONStats struct {
	TotalConstants         int     `json:"total_constants"`
	ActuallyUsed           int     `json:"actually_used"`
	MigrationOpportunities int     `json:"migration_opportunities"`
	CompletelyUnused       int     `json:"completely_unused"`
	UsagePercentage        float64 `json:"usage_percentage"`
	HardcodedClasses       int     `json:"hardcoded_classes"`
	ConstantReferences     int     `json:"constant_references"`
}

// JSONIssue represents a single linting issue
type JSONIssue struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Linter   string `json:"linter"`
	Source   string `json:"source,omitempty"` // Optional source line
}

// JSONQuickWins contains migration opportunities
type JSONQuickWins struct {
	SingleClass []JSONQuickWin `json:"single_class"`
	MultiClass  []JSONQuickWin `json:"multi_class"`
}

// JSONQuickWin represents a high-impact refactoring opportunity
type JSONQuickWin struct {
	Class       string `json:"class"`
	Occurrences int    `json:"occurrences"`
	Suggestion  string `json:"suggestion"`
}

// WriteJSON writes the lint result as JSON
func WriteJSON(w io.Writer, result *LintResult) error {
	output := buildJSONOutput(result)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// buildJSONOutput converts LintResult to JSONOutput
func buildJSONOutput(result *LintResult) JSONOutput {
	// Count errors and warnings
	var errors, warnings int
	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		}
	}

	// Convert issues
	jsonIssues := make([]JSONIssue, len(result.Issues))
	for i, issue := range result.Issues {
		source := ""
		if len(issue.SourceLines) > 0 {
			source = issue.SourceLines[0]
		}
		jsonIssues[i] = JSONIssue{
			File:     issue.Pos.Filename,
			Line:     issue.Pos.Line,
			Column:   issue.Pos.Column,
			Severity: issue.Severity,
			Message:  issue.Text,
			Linter:   issue.FromLinter,
			Source:   source,
		}
	}

	// Convert quick wins
	singleClass := make([]JSONQuickWin, len(result.QuickWins.SingleClass))
	for i, win := range result.QuickWins.SingleClass {
		singleClass[i] = JSONQuickWin{
			Class:       win.ClassName,
			Occurrences: win.Occurrences,
			Suggestion:  win.Suggestion,
		}
	}

	multiClass := make([]JSONQuickWin, len(result.QuickWins.MultiClass))
	for i, win := range result.QuickWins.MultiClass {
		multiClass[i] = JSONQuickWin{
			Class:       win.ClassName,
			Occurrences: win.Occurrences,
			Suggestion:  win.Suggestion,
		}
	}

	return JSONOutput{
		Version:   "1.0",
		Timestamp: time.Now().Format(time.RFC3339),
		Summary: JSONSummary{
			TotalIssues:  len(result.Issues),
			Errors:       errors,
			Warnings:     warnings,
			FilesScanned: result.FilesScanned,
		},
		Stats: JSONStats{
			TotalConstants:         result.TotalConstants,
			ActuallyUsed:           result.ActuallyUsed,
			MigrationOpportunities: result.AvailableForMigration,
			CompletelyUnused:       result.CompletelyUnused,
			UsagePercentage:        result.UsagePercentage,
			HardcodedClasses:       result.ClassesFound,
			ConstantReferences:     result.ConstantsFound,
		},
		Issues: jsonIssues,
		QuickWins: JSONQuickWins{
			SingleClass: singleClass,
			MultiClass:  multiClass,
		},
	}
}
