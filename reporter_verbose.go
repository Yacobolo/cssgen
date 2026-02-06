package cssgen

import (
	"fmt"
	"io"
)

// VerboseReporter handles detailed statistics and suggestions
type VerboseReporter struct {
	w         io.Writer
	useColors bool
}

// NewVerboseReporter creates a verbose reporter
func NewVerboseReporter(w io.Writer, useColors bool) *VerboseReporter {
	return &VerboseReporter{
		w:         w,
		useColors: useColors,
	}
}

// PrintStatistics outputs detailed linting statistics
func (r *VerboseReporter) PrintStatistics(result LintResult) {
	fmt.Fprintln(r.w, "")
	fmt.Fprintln(r.w, RenderStyle(StyleCyan, "CSS Linter Statistics", r.useColors))
	fmt.Fprintln(r.w, "------------------------")

	fmt.Fprintf(r.w, "Total Constants:         %d\n", result.TotalConstants)
	fmt.Fprintf(r.w, "Actually Used:           %d (%.1f%%)\n", result.ActuallyUsed, result.UsagePercentage)
	fmt.Fprintf(r.w, "Migration Opportunities: %d\n", result.AvailableForMigration)
	fmt.Fprintf(r.w, "Completely Unused:       %d\n", result.CompletelyUnused)
	fmt.Fprintf(r.w, "Files Scanned:           %d\n", result.FilesScanned)
	fmt.Fprintf(r.w, "Hardcoded Classes:       %d\n", result.ClassesFound)
	fmt.Fprintf(r.w, "Constant References:     %d\n", result.ConstantsFound)
}

// PrintAdoptionProgress shows visual progress bar
func (r *VerboseReporter) PrintAdoptionProgress(result LintResult) {
	fmt.Fprintln(r.w, "")
	fmt.Fprintln(r.w, RenderStyle(StyleCyan, "Adoption Progress", r.useColors))
	fmt.Fprintln(r.w, "-------------------")
	printProgressBar(r.w, result.UsagePercentage)
}

// PrintQuickWins shows migration opportunities
func (r *VerboseReporter) PrintQuickWins(result LintResult) {
	if len(result.QuickWins.SingleClass) == 0 && len(result.QuickWins.MultiClass) == 0 {
		return
	}

	fmt.Fprintln(r.w, "")
	fmt.Fprintln(r.w, RenderStyle(StyleGreen, "Quick Wins", r.useColors))
	fmt.Fprintln(r.w, "-------------")

	if len(result.QuickWins.SingleClass) > 0 {
		fmt.Fprintln(r.w, "\nHigh Confidence (Single Class - Direct Replace):")
		for i, win := range result.QuickWins.SingleClass {
			if i >= 10 {
				break
			}
			fmt.Fprintf(r.w, "%d. \"%s\" - %d occurrences → Use %s\n",
				i+1, win.ClassName, win.Occurrences, win.Suggestion)
		}
	}

	if len(result.QuickWins.MultiClass) > 0 {
		fmt.Fprintln(r.w, "\nMigration Opportunities (Multi-Class Consolidation):")
		for i, win := range result.QuickWins.MultiClass {
			if i >= 10 {
				break
			}
			fmt.Fprintf(r.w, "%d. \"%s\" - %d occurrences → Use %s\n",
				i+1, win.ClassName, win.Occurrences, win.Suggestion)
		}
	}
}

// PrintWarnings shows linter warnings
func (r *VerboseReporter) PrintWarnings(result LintResult) {
	if len(result.Warnings) == 0 {
		return
	}

	fmt.Fprintln(r.w, "")
	fmt.Fprintln(r.w, RenderStyle(StyleYellow, "Warnings", r.useColors))
	fmt.Fprintln(r.w, "-----------")

	for _, warning := range result.Warnings {
		fmt.Fprintf(r.w, "• %s\n", warning)
	}
}
