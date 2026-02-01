// Package cssgen provides CSS class constant generation and linting.
//
// # Pure 1:1 Class Mapping
//
// The generator creates one Go constant for each CSS class with exact 1:1 mapping:
//
//   - CSS: .btn           → Go: Btn = "btn"
//   - CSS: .btn--brand    → Go: BtnBrand = "btn--brand"
//   - CSS: .card__header  → Go: CardHeader = "card__header"
//
// Usage in templates:
//
//	<button class={ ui.Btn, ui.BtnBrand }>Click</button>
//	// Produces: <button class="btn btn--brand">
//
//	<div class={ ui.Card, ui.CardHeader }>Content</div>
//	// Produces: <div class="card card__header">
//
// # Smart Linter with Greedy Token Matching
//
// When analyzing class="btn btn--brand", the linter:
//
//  1. Checks if exact match exists for full string
//  2. If no match, splits into tokens: ["btn", "btn--brand"]
//  3. Matches each token individually: btn → ui.Btn, btn--brand → ui.BtnBrand
//  4. Suggests: { ui.Btn, ui.BtnBrand }
//
// This results in predictable, accurate suggestions with zero "pollution."
//
// # Match Results
//
//   - Matched:    Constant available (e.g., "btn" → ui.Btn)
//   - Unmatched:  Class exists in CSS but no constant generated (e.g., utility classes)
//   - Invalid:    Class doesn't exist in CSS (typo or missing)
package cssgen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
)

// LintConfig holds linting configuration
type LintConfig struct {
	ScanPaths     []string // Patterns to scan (e.g., "internal/web/features/**/*.templ")
	GeneratedFile string   // Path to styles.gen.go
	PackageName   string   // "ui"
	Verbose       bool
	Strict        bool    // Exit with code 1 if issues found
	Threshold     float64 // Minimum adoption percentage (for -strict mode)

	// New golangci-style configuration
	MaxIssuesPerLinter int  // 0 = unlimited (default)
	MaxSameIssues      int  // 0 = unlimited (default)
	ShowStats          bool // Show statistics summary (auto-enabled with Verbose)
	PrintIssuedLines   bool // Show source lines with issues (default: true)
	PrintLinterName    bool // Show (csslint) suffix (default: true)
	UseColors          bool // Enable color output (default: auto-detect)
}

// LintResult contains linting analysis results
type LintResult struct {
	// Statistics
	TotalConstants        int     // 229
	ActuallyUsed          int     // Constants referenced via ui.ConstName (e.g., 0)
	AvailableForMigration int     // Constants that match hardcoded strings (e.g., 111)
	CompletelyUnused      int     // No usage, no matches (e.g., 118)
	UsagePercentage       float64 // Percentage of actually used constants (e.g., 0%)

	// Issues in golangci-lint format
	Issues           []Issue            // All issues found
	IssuesByCategory map[string][]Issue // Grouped by type for stats

	// Legacy detailed findings (used for verbose mode only)
	UnusedClasses    []UnusedClass
	HardcodedStrings []HardcodedString
	InvalidClasses   []InvalidClass // Classes that don't exist in CSS
	FilesScanned     int
	ClassesFound     int // Total hardcoded classes found
	ConstantsFound   int // Total ui.Foo references found
	ErrorCount       int // Count of invalid classes
	TruncatedCount   int // Issues removed due to limits

	// Summary
	Warnings    []string
	Suggestions []string
	QuickWins   QuickWinsSummary // Most frequently hardcoded classes
}

// UnusedClass represents a generated constant with no usage
type UnusedClass struct {
	ConstName string // "AppSidebar"
	CSSClass  string // "app-sidebar"
	Layer     string // "components"
	DefinedIn string // "styles.gen.go:123"
}

// InvalidClass represents a class that doesn't exist in CSS
type InvalidClass struct {
	ClassName   string       // "btn--outline"
	Location    FileLocation // Where it was found
	LineContent string       // Full line for context
}

// ClassificationResult categorizes a class reference
type ClassificationResult int

const (
	// ClassMatched indicates the class has a constant available (migration opportunity).
	ClassMatched ClassificationResult = iota
	// ClassBypassed indicates a valid CSS class with no constant needed (allowed).
	ClassBypassed
	// ClassZombie indicates the class doesn't exist in CSS (error).
	ClassZombie
)

// HardcodedString represents a CSS class string that could use a constant
type HardcodedString struct {
	FullClassValue string             // "btn btn--ghost btn--sm"
	Suggestion     ConstantSuggestion // Smart suggestion with analysis
	Location       FileLocation
	LineContent    string // Full line for context
}

// MatchType indicates how a class was matched to a constant
type MatchType int

// MatchType represents how a CSS class was matched.
const (
	MatchExact MatchType = iota // Direct match in ExactMap
	MatchNone                   // No match found
)

// ClassAnalysis provides detailed breakdown of how a class was analyzed
type ClassAnalysis struct {
	ClassName  string    // "btn--ghost"
	Match      MatchType // MatchModifier
	Suggestion string    // "BtnGhost"
	Context    string    // "included in ui.BtnGhost"
}

// ConstantSuggestion contains the analysis and recommendation for a class string
type ConstantSuggestion struct {
	Constants        []string        // ["BtnGhost", "BtnSm"] (deduplicated)
	Analysis         []ClassAnalysis // Per-class breakdown (for verbose mode)
	HasUnmatched     bool            // True if some classes had no match
	UnmatchedClasses []string        // List of classes that didn't match any constant
	HasInvalid       bool            // Contains invalid (non-existent) classes
	InvalidClasses   []string        // List of invalid classes
}

// QuickWin represents a high-impact refactoring opportunity
type QuickWin struct {
	ClassName   string // "btn"
	Occurrences int    // 45
	Suggestion  string // "ui.Btn"
}

// QuickWinsSummary categorizes quick wins by refactoring pattern
type QuickWinsSummary struct {
	SingleClass []QuickWin // Single class: "btn" -> ui.Btn
	MultiClass  []QuickWin // Multiple classes: "btn btn--brand" -> { ui.Btn, ui.BtnBrand }
}

// CSSLookup provides fast lookups for CSS class -> constant mapping
type CSSLookup struct {
	// ExactMap: 1:1 mapping - "btn" -> "Btn", "btn--brand" -> "BtnBrand"
	ExactMap map[string]string

	// AllConstants: All constants (unchanged)
	AllConstants map[string]string // ConstName -> CSSClass

	// ConstantParts: Track which classes a constant contains
	// With 1:1 mapping, this is always single-element array
	ConstantParts map[string][]string

	// AllCSSClasses: All classes found in CSS (for static analysis)
	// Used to detect invalid class references (typos)
	AllCSSClasses map[string]bool
}

// Lint performs linting analysis on the codebase
func Lint(config LintConfig) (*LintResult, error) {
	// Step 1: Parse generated constants file
	constants, allCSSClasses, err := ParseGeneratedFile(config.GeneratedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated file: %w", err)
	}

	// Step 2: Build lookup maps
	lookup := buildLookupMaps(constants)
	lookup.AllCSSClasses = allCSSClasses

	// Step 3: Scan files for class references
	references, stats, err := ScanFiles(config.ScanPaths, config.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}
	_ = stats // Stats are printed in ScanFiles if verbose

	// Step 4: Count unique files
	filesScanned := countUniqueFiles(references)

	// Step 5: Analyze usage
	result := analyzeUsage(constants, references, lookup)
	result.FilesScanned = filesScanned

	// Step 6: Generate suggestions
	result.Suggestions = generateSuggestions(result)

	// Step 7: Apply issue limiting if configured
	if config.MaxIssuesPerLinter > 0 || config.MaxSameIssues > 0 {
		result.Issues, result.TruncatedCount = limitIssues(result.Issues, config)
	}

	return result, nil
}

// ParseGeneratedFile reads styles.gen.go and all related split files (styles_*.gen.go)
// and extracts constant definitions and AllCSSClasses
func ParseGeneratedFile(path string) (map[string]string, map[string]bool, error) {
	constants := make(map[string]string)
	allCSSClasses := make(map[string]bool)

	// Parse main file and all split files in the same directory
	dir := filepath.Dir(path)
	pattern := filepath.Join(dir, "styles*.gen.go")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, nil, fmt.Errorf("glob pattern error: %w", err)
	}

	// If no files found via glob, try the provided path directly
	if len(files) == 0 {
		files = []string{path}
	}

	fset := token.NewFileSet()
	for _, filePath := range files {
		file, err := parser.ParseFile(fset, filePath, nil, 0)
		if err != nil {
			// Skip files that can't be parsed (might be in progress)
			continue
		}

		// Walk the AST and find const and var declarations
		ast.Inspect(file, func(n ast.Node) bool {
			if genDecl, ok := n.(*ast.GenDecl); ok {
				// Parse constants (existing logic)
				if genDecl.Tok == token.CONST {
					for _, spec := range genDecl.Specs {
						if vspec, ok := spec.(*ast.ValueSpec); ok {
							if len(vspec.Names) > 0 && len(vspec.Values) > 0 {
								name := vspec.Names[0].Name
								if lit, ok := vspec.Values[0].(*ast.BasicLit); ok {
									value := strings.Trim(lit.Value, `"`)
									constants[name] = value
								}
							}
						}
					}
				}

				// Parse AllCSSClasses var (NEW)
				if genDecl.Tok == token.VAR {
					for _, spec := range genDecl.Specs {
						if vspec, ok := spec.(*ast.ValueSpec); ok {
							if len(vspec.Names) > 0 && vspec.Names[0].Name == "AllCSSClasses" {
								// Parse the map literal
								if comp, ok := vspec.Values[0].(*ast.CompositeLit); ok {
									for _, elt := range comp.Elts {
										if kv, ok := elt.(*ast.KeyValueExpr); ok {
											if lit, ok := kv.Key.(*ast.BasicLit); ok {
												className := strings.Trim(lit.Value, `"`)
												allCSSClasses[className] = true
											}
										}
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}

	return constants, allCSSClasses, nil
}

// buildLookupMaps creates reverse lookup maps for fast searching
func buildLookupMaps(constants map[string]string) *CSSLookup {
	lookup := &CSSLookup{
		ExactMap:      make(map[string]string),
		AllConstants:  constants,
		ConstantParts: make(map[string][]string),
	}

	// With 1:1 mapping, this is trivial
	for constName, cssValue := range constants {
		// Direct 1:1 mapping
		lookup.ExactMap[cssValue] = constName

		// ConstantParts is now always single-element (1:1)
		lookup.ConstantParts[constName] = []string{cssValue}
	}

	return lookup
}

// analyzeUsage compares constants with found references
func analyzeUsage(constants map[string]string, references []ClassReference, lookup *CSSLookup) *LintResult {
	result := &LintResult{
		TotalConstants: len(constants),
	}

	// Track which constants are actually used (via ui.ConstName)
	actuallyUsed := make(map[string]bool)
	// Track which constants have migration opportunities (match hardcoded strings)
	availableForMigration := make(map[string]bool)

	var hardcodedStrings []HardcodedString
	var invalidClasses []InvalidClass
	var issues []Issue

	for _, ref := range references {
		if ref.IsConstant {
			// This is a ui.Foo reference - actually used in code
			actuallyUsed[ref.ConstName] = true
			result.ConstantsFound++
		} else {
			// This is a hardcoded string
			result.ClassesFound++

			// Use smart solver with full class value
			suggestion := ResolveBestConstants(ref.FullClassValue, lookup)

			// Track invalid classes and create error issues
			if suggestion.HasInvalid {
				for _, invalidClass := range suggestion.InvalidClasses {
					invalidClasses = append(invalidClasses, InvalidClass{
						ClassName:   invalidClass,
						Location:    ref.Location,
						LineContent: ref.LineContent,
					})
					result.ErrorCount++

					// Find the exact column for this specific invalid class
					column := findClassColumn(ref.Location.Text, invalidClass)
					if column == 0 {
						column = ref.Location.Column // fallback to original column
					}

					// Create error issue
					issues = append(issues, Issue{
						FromLinter:  "csslint",
						Text:        fmt.Sprintf(IssueInvalidClass, invalidClass),
						Severity:    SeverityError,
						SourceLines: []string{ref.Location.Text},
						Pos: IssuePos{
							Filename: ref.Location.File,
							Line:     ref.Location.Line,
							Column:   column,
						},
					})
				}
			}

			if len(suggestion.Constants) > 0 {
				// Mark suggested constants as "available for migration"
				// but NOT as "actually used"
				for _, constName := range suggestion.Constants {
					availableForMigration[constName] = true
				}

				hs := HardcodedString{
					FullClassValue: ref.FullClassValue,
					Suggestion:     suggestion,
					Location:       ref.Location,
					LineContent:    ref.LineContent,
				}
				hardcodedStrings = append(hardcodedStrings, hs)

				// NEW: Create WARNING issue for hardcoded strings (unless internal class or has invalid classes)
				// Skip warning if the suggestion contains invalid classes (already reported as error)
				if !hasInternalClasses(ref.FullClassValue) && !suggestion.HasInvalid {
					column := findClassColumn(ref.Location.Text, ref.FullClassValue)
					if column == 0 {
						column = ref.Location.Column // fallback to original column
					}

					suggestionText := formatSuggestion(suggestion)
					issues = append(issues, Issue{
						FromLinter:  "csslint",
						Text:        fmt.Sprintf(IssueHardcodedClass, ref.FullClassValue, suggestionText),
						Severity:    SeverityWarning,
						SourceLines: []string{ref.Location.Text},
						Pos: IssuePos{
							Filename: ref.Location.File,
							Line:     ref.Location.Line,
							Column:   column,
						},
					})
				}
			}
		}
	}

	result.ActuallyUsed = len(actuallyUsed)
	result.AvailableForMigration = len(availableForMigration)
	result.CompletelyUnused = result.TotalConstants - result.ActuallyUsed - result.AvailableForMigration
	result.HardcodedStrings = hardcodedStrings
	result.InvalidClasses = invalidClasses

	if result.TotalConstants > 0 {
		result.UsagePercentage = float64(result.ActuallyUsed) / float64(result.TotalConstants) * 100
	}

	// Combine actually used and available for migration to find what's used/referenced
	allUsedOrReferenced := make(map[string]bool)
	for k := range actuallyUsed {
		allUsedOrReferenced[k] = true
	}
	for k := range availableForMigration {
		allUsedOrReferenced[k] = true
	}

	// Find unused constants (constants with no usage and no migration opportunities)
	result.UnusedClasses = findUnusedConstants(constants, allUsedOrReferenced)

	// Generate quick wins
	result.QuickWins = generateQuickWins(hardcodedStrings)

	// Store issues
	result.Issues = issues

	// Group issues by category
	result.IssuesByCategory = make(map[string][]Issue)
	for _, issue := range issues {
		result.IssuesByCategory[issue.Severity] = append(result.IssuesByCategory[issue.Severity], issue)
	}

	return result
}

// findConstantSuggestion finds the best constant match for a CSS class
// With 1:1 mapping, this is a simple exact lookup
func findConstantSuggestion(className string, lookup *CSSLookup) string {
	// 1:1 lookup - simple and fast!
	if constName, exists := lookup.ExactMap[className]; exists {
		return constName
	}

	return ""
}

// classifyClass determines if a class is valid, has a constant, or is invalid
func classifyClass(className string, lookup *CSSLookup) ClassificationResult {
	// Check if class exists in CSS
	if !lookup.AllCSSClasses[className] {
		return ClassZombie // ERROR: Class doesn't exist
	}

	// Check if we have a constant for it (1:1 mapping)
	if _, exists := lookup.ExactMap[className]; exists {
		return ClassMatched // Has exact constant
	}

	// Valid CSS class, but no constant (e.g., _internal, utility)
	return ClassBypassed
}

// ResolveBestConstants analyzes a full class string and returns the optimal constant combination.
//
// Algorithm (Greedy Token Matching):
//  1. Try exact match for entire string
//  2. Split into tokens and match each individually (1:1 mapping)
//  3. No deduplication needed with 1:1 mapping
//
// Example:
//
//	Input:  "btn btn--brand"
//	Output: ConstantSuggestion{
//	  Type: SuggestionMultiClass,
//	  Constants: ["Btn", "BtnBrand"],  // Each token maps exactly
//	  Analysis: [
//	    {ClassName: "btn", Match: MatchExact, Suggestion: "Btn"},
//	    {ClassName: "btn--brand", Match: MatchExact, Suggestion: "BtnBrand"},
//	  ],
//	}
func ResolveBestConstants(classString string, lookup *CSSLookup) ConstantSuggestion {
	classes := strings.Fields(classString)

	// Step 1: Try exact match for entire string first
	if constName, exists := lookup.ExactMap[classString]; exists {
		return ConstantSuggestion{
			Constants: []string{constName},
			Analysis:  nil, // No breakdown needed
		}
	}

	// Step 2: Greedy token matching - split and match each individually
	var suggestions []string
	var analysis []ClassAnalysis
	var unmatchedClasses []string
	var invalidClasses []string

	for _, class := range classes {
		classAnalysis := ClassAnalysis{ClassName: class}

		// Classify the class
		classification := classifyClass(class, lookup)

		switch classification {
		case ClassZombie:
			// ERROR: Class doesn't exist in CSS
			classAnalysis.Match = MatchNone
			invalidClasses = append(invalidClasses, class)
			unmatchedClasses = append(unmatchedClasses, class)

		case ClassBypassed:
			// Valid CSS, no constant - silently allow
			classAnalysis.Match = MatchNone
			classAnalysis.Context = "valid CSS (no constant)"

		case ClassMatched:
			// 1:1 lookup - simple and fast!
			if constName, exists := lookup.ExactMap[class]; exists {
				suggestions = append(suggestions, constName)
				classAnalysis.Match = MatchExact
				classAnalysis.Suggestion = constName
			}
		}

		analysis = append(analysis, classAnalysis)
	}

	// With 1:1 mapping, NO deduplication needed!
	// Each class token maps to exactly one constant

	return ConstantSuggestion{
		Constants:        suggestions,
		Analysis:         analysis,
		HasUnmatched:     len(unmatchedClasses) > 0,
		UnmatchedClasses: unmatchedClasses,
		HasInvalid:       len(invalidClasses) > 0,
		InvalidClasses:   invalidClasses,
	}
}

// formatSuggestion converts a ConstantSuggestion to a human-readable string
func formatSuggestion(s ConstantSuggestion) string {
	if len(s.Constants) == 0 {
		return "(no suggestion)"
	}

	if len(s.Constants) == 1 {
		return "ui." + s.Constants[0]
	}

	// Multiple constants: { ui.Btn, ui.BtnBrand }
	parts := make([]string, len(s.Constants))
	for i, c := range s.Constants {
		parts[i] = "ui." + c
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// isInternalClass checks if a class name starts with underscore
// These are treated as intentional "escape hatches" and not warned about
func isInternalClass(className string) bool {
	return strings.HasPrefix(className, "_")
}

// hasInternalClasses checks if any class in a space-separated string is internal
func hasInternalClasses(fullClassValue string) bool {
	classes := strings.Fields(fullClassValue)
	for _, cls := range classes {
		if isInternalClass(cls) {
			return true
		}
	}
	return false
}

// findUnusedConstants identifies constants with 0 references
func findUnusedConstants(constants map[string]string, usedConsts map[string]bool) []UnusedClass {
	var unused []UnusedClass

	for constName, cssValue := range constants {
		if !usedConsts[constName] {
			unused = append(unused, UnusedClass{
				ConstName: constName,
				CSSClass:  cssValue,
				Layer:     inferLayer(cssValue), // Simple heuristic
			})
		}
	}

	// Sort by name
	sort.Slice(unused, func(i, j int) bool {
		return unused[i].ConstName < unused[j].ConstName
	})

	return unused
}

// inferLayer attempts to guess the layer from the CSS class name
func inferLayer(cssClass string) string {
	// Simple heuristics
	if strings.HasPrefix(cssClass, "text-") ||
		strings.HasPrefix(cssClass, "bg-") ||
		strings.HasPrefix(cssClass, "flex-") ||
		strings.HasPrefix(cssClass, "grid-") {
		return "utilities"
	}

	if strings.Contains(cssClass, "btn") ||
		strings.Contains(cssClass, "card") ||
		strings.Contains(cssClass, "modal") ||
		strings.Contains(cssClass, "nav") ||
		strings.Contains(cssClass, "table") {
		return "components"
	}

	return "base"
}

// generateQuickWins identifies the most frequently hardcoded classes
func generateQuickWins(hardcodedStrings []HardcodedString) QuickWinsSummary {
	singleClass := make(map[string]int)
	multiClass := make(map[string]int)
	suggestionMap := make(map[string]string)

	for _, hs := range hardcodedStrings {
		// Skip suggestions with unmatched classes - they're not "quick wins"
		if hs.Suggestion.HasUnmatched {
			continue
		}

		classes := strings.Fields(hs.FullClassValue)

		if len(classes) == 1 && len(hs.Suggestion.Constants) == 1 {
			// Single-class exact match
			singleClass[hs.FullClassValue]++
			suggestionMap[hs.FullClassValue] = "ui." + hs.Suggestion.Constants[0]
		} else if len(classes) > 1 && len(hs.Suggestion.Constants) > 1 {
			// Multi-class pattern (only if ALL classes matched)
			multiClass[hs.FullClassValue]++

			constList := make([]string, len(hs.Suggestion.Constants))
			for i, c := range hs.Suggestion.Constants {
				constList[i] = "ui." + c
			}
			suggestionMap[hs.FullClassValue] = "{ " + strings.Join(constList, ", ") + " }"
		}
	}

	return QuickWinsSummary{
		SingleClass: sortByFrequency(singleClass, suggestionMap),
		MultiClass:  sortByFrequency(multiClass, suggestionMap),
	}
}

// sortByFrequency converts frequency map to sorted QuickWin slice
func sortByFrequency(freq map[string]int, suggestions map[string]string) []QuickWin {
	var wins []QuickWin

	for className, count := range freq {
		if suggestion, ok := suggestions[className]; ok {
			wins = append(wins, QuickWin{
				ClassName:   className,
				Occurrences: count,
				Suggestion:  suggestion,
			})
		}
	}

	// Sort by occurrences (descending)
	sort.Slice(wins, func(i, j int) bool {
		return wins[i].Occurrences > wins[j].Occurrences
	})

	// Limit to top 10
	if len(wins) > 10 {
		wins = wins[:10]
	}

	return wins
}

// generateSuggestions creates actionable recommendations
func generateSuggestions(result *LintResult) []string {
	var suggestions []string

	if len(result.HardcodedStrings) > 0 {
		suggestions = append(suggestions, "Import the ui package in template files: import \"go-scheduler/internal/web/ui\"")
		suggestions = append(suggestions, "Replace hardcoded strings with constants (see Quick Wins below)")
	}

	if result.CompletelyUnused > 50 {
		suggestions = append(suggestions, "Consider removing unused constants or adding them to templates")
	}

	if result.UsagePercentage < 20 {
		suggestions = append(suggestions, "Low adoption detected - start with Quick Wins for maximum impact")
	}

	return suggestions
}

// countUniqueFiles counts unique files in references
func countUniqueFiles(references []ClassReference) int {
	files := make(map[string]bool)
	for _, ref := range references {
		files[ref.Location.File] = true
	}
	return len(files)
}

// PrintLintReport formats and prints the lint report
func PrintLintReport(result *LintResult, w io.Writer, verbose bool) {
	// Color setup
	cyan := color.New(color.FgCyan, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	// Header
	fmt.Fprintln(w, "")
	cyan.Fprintln(w, "CSS Constants Lint Report")
	cyan.Fprintln(w, "=========================")

	// ERRORS SECTION (NEW - Show first, most important)
	if len(result.InvalidClasses) > 0 {
		fmt.Fprintln(w, "")
		red.Fprintln(w, "❌ ERRORS - Invalid CSS Classes")
		fmt.Fprintln(w, "-------------------------------")
		fmt.Fprintln(w, "These classes are used in templates but don't exist in any CSS file:")

		// Group by class name
		invalidByClass := make(map[string][]FileLocation)
		for _, invalid := range result.InvalidClasses {
			invalidByClass[invalid.ClassName] = append(
				invalidByClass[invalid.ClassName],
				invalid.Location,
			)
		}

		fmt.Fprintln(w, "")
		for className, locations := range invalidByClass {
			fmt.Fprintf(w, "• %q (%d occurrence%s)\n",
				className,
				len(locations),
				pluralize(len(locations)))

			// Show first 3 locations
			limit := 3
			if len(locations) < limit {
				limit = len(locations)
			}
			for i := 0; i < limit; i++ {
				fmt.Fprintf(w, "  📁 %s:%d\n",
					GetRelativePath(locations[i].File),
					locations[i].Line)
			}
			if len(locations) > limit {
				fmt.Fprintf(w, "  ... and %d more location%s\n",
					len(locations)-limit,
					pluralize(len(locations)-limit))
			}
			fmt.Fprintln(w, "")
		}
	}

	// Statistics
	cyan.Fprintln(w, "📊 STATISTICS")
	fmt.Fprintln(w, "-------------")

	// Add error count to stats (if any)
	if result.ErrorCount > 0 {
		fmt.Fprintf(w, "❌ Errors Found:        %d\n", result.ErrorCount)
	}

	fmt.Fprintf(w, "Total Constants:        %d\n", result.TotalConstants)
	fmt.Fprintf(w, "Actually Used:          %d (%.1f%%)\n", result.ActuallyUsed, result.UsagePercentage)
	fmt.Fprintf(w, "Migration Opportunities: %d\n", result.AvailableForMigration)
	fmt.Fprintf(w, "Completely Unused:      %d\n", result.CompletelyUnused)
	fmt.Fprintf(w, "Files Scanned:          %d\n", result.FilesScanned)
	fmt.Fprintf(w, "Hardcoded Classes:      %d\n", result.ClassesFound)
	fmt.Fprintf(w, "Constant References:    %d\n", result.ConstantsFound)

	// Adoption progress bar
	fmt.Fprintln(w, "")
	cyan.Fprintln(w, "📈 ADOPTION PROGRESS")
	fmt.Fprintln(w, "-------------------")
	printProgressBar(w, result.UsagePercentage)

	// Quick Wins
	if len(result.QuickWins.SingleClass) > 0 || len(result.QuickWins.MultiClass) > 0 {
		printQuickWins(w, result.QuickWins, green)
	}

	// Warnings summary (compact by default)
	if len(result.UnusedClasses) > 0 || len(result.HardcodedStrings) > 0 {
		fmt.Fprintln(w, "")
		yellow.Fprintln(w, "⚠️  WARNINGS")
		fmt.Fprintln(w, "-----------")

		if len(result.UnusedClasses) > 0 {
			fmt.Fprintf(w, "• %d unused constants (never imported)\n", len(result.UnusedClasses))
		}

		if len(result.HardcodedStrings) > 0 {
			fmt.Fprintf(w, "• %d hardcoded class strings found\n", len(result.HardcodedStrings))
		}
	}

	// Detailed output only in verbose mode
	if verbose {
		// Unused constants
		if len(result.UnusedClasses) > 0 {
			fmt.Fprintln(w, "")
			yellow.Fprintln(w, "🔍 UNUSED CONSTANTS (Detailed)")
			fmt.Fprintln(w, "-----------------------------")

			// Group by layer
			byLayer := make(map[string][]UnusedClass)
			for _, unused := range result.UnusedClasses {
				byLayer[unused.Layer] = append(byLayer[unused.Layer], unused)
			}

			for layer, classes := range byLayer {
				fmt.Fprintf(w, "\nLayer: %s (%d unused)\n", layer, len(classes))
				for _, cls := range classes {
					fmt.Fprintf(w, "  • %-20s → \"%s\"\n", cls.ConstName, cls.CSSClass)
				}
			}
		}

		// Hardcoded strings (show first 20 with detailed analysis)
		if len(result.HardcodedStrings) > 0 {
			printHardcodedStringsVerbose(w, result.HardcodedStrings, yellow)
		}
	} else {
		// Compact mode: show first 5 hardcoded strings
		if len(result.HardcodedStrings) > 0 {
			printHardcodedStringsCompact(w, result.HardcodedStrings, 5, yellow)
		}
	}

	// Recommendations
	if len(result.Suggestions) > 0 {
		fmt.Fprintln(w, "")
		green.Fprintln(w, "✅ RECOMMENDATIONS")
		fmt.Fprintln(w, "------------------")
		for i, suggestion := range result.Suggestions {
			fmt.Fprintf(w, "%d. %s\n", i+1, suggestion)
		}
	}

	// Summary
	fmt.Fprintln(w, "")
	if result.ErrorCount > 0 {
		red.Fprintf(w, "❌ BUILD FAILED: %d invalid CSS class%s found. Fix these errors before deploying.\n",
			result.ErrorCount,
			pluralize(result.ErrorCount))
	} else if result.UsagePercentage >= 80 {
		green.Fprintf(w, "✓ Excellent adoption! %.1f%% of constants are in use.\n", result.UsagePercentage)
	} else if result.UsagePercentage >= 50 {
		yellow.Fprintf(w, "⚠ Moderate adoption. %.1f%% of constants are in use. Focus on Quick Wins to improve.\n", result.UsagePercentage)
	} else {
		red.Fprintf(w, "⚠ Low adoption. Only %.1f%% of constants are in use. Start with Quick Wins for maximum impact.\n", result.UsagePercentage)
	}

	if !verbose && (len(result.UnusedClasses) > 0 || len(result.HardcodedStrings) > 0) {
		fmt.Fprintln(w, "\nRun with -verbose for detailed breakdown")
	}

	fmt.Fprintln(w, "")
}

// printQuickWins prints the quick wins section with categorization
func printQuickWins(w io.Writer, summary QuickWinsSummary, green *color.Color) {
	fmt.Fprintln(w, "")
	green.Fprintln(w, "🎯 QUICK WINS")
	fmt.Fprintln(w, "-------------")

	if len(summary.SingleClass) > 0 {
		fmt.Fprintln(w, "\nHigh Confidence (Single Class - Direct Replace):")
		for i, win := range summary.SingleClass {
			if i >= 10 {
				break
			}
			fmt.Fprintf(w, "%d. \"%s\" - %d occurrences → Use %s\n",
				i+1, win.ClassName, win.Occurrences, win.Suggestion)
		}
	}

	if len(summary.MultiClass) > 0 {
		fmt.Fprintln(w, "\nMigration Opportunities (Multi-Class Consolidation):")
		for i, win := range summary.MultiClass {
			if i >= 10 {
				break
			}
			fmt.Fprintf(w, "%d. \"%s\" - %d occurrences → Use %s\n",
				i+1, win.ClassName, win.Occurrences, win.Suggestion)
		}
	}
}

// printHardcodedStringsCompact prints a compact summary of hardcoded strings
func printHardcodedStringsCompact(w io.Writer, hardcodedStrings []HardcodedString, limit int, yellow *color.Color) {
	fmt.Fprintln(w, "")
	yellow.Fprintln(w, "⚠️  HARDCODED STRINGS")
	fmt.Fprintln(w, "-------------------")
	fmt.Fprintf(w, "Found %d hardcoded class strings\n", len(hardcodedStrings))

	if limit > 0 && len(hardcodedStrings) > limit {
		fmt.Fprintf(w, "Showing first %d (use -verbose for full list)\n", limit)
		hardcodedStrings = hardcodedStrings[:limit]
	}

	for _, hs := range hardcodedStrings {
		fmt.Fprintf(w, "\n📁 %s:%d\n", GetRelativePath(hs.Location.File), hs.Location.Line)
		fmt.Fprintf(w, "   Found: %s\n", hs.FullClassValue)

		if len(hs.Suggestion.Constants) == 1 {
			fmt.Fprintf(w, "   Suggestion: Use ui.%s\n", hs.Suggestion.Constants[0])
		} else if len(hs.Suggestion.Constants) > 1 {
			constList := make([]string, len(hs.Suggestion.Constants))
			for i, c := range hs.Suggestion.Constants {
				constList[i] = "ui." + c
			}
			suggestion := fmt.Sprintf("{ %s }", strings.Join(constList, ", "))
			if hs.Suggestion.HasUnmatched {
				fmt.Fprintf(w, "   Suggestion: Replace with %s ⚠️  (loses: %s)\n",
					suggestion, strings.Join(hs.Suggestion.UnmatchedClasses, ", "))
			} else {
				fmt.Fprintf(w, "   Suggestion: Replace with %s\n", suggestion)
			}
		} else if hs.Suggestion.HasUnmatched {
			fmt.Fprintf(w, "   ⚠️  No matching constants for: %s\n",
				strings.Join(hs.Suggestion.UnmatchedClasses, ", "))
		}
	}
}

// printHardcodedStringsVerbose prints detailed analysis of hardcoded strings
func printHardcodedStringsVerbose(w io.Writer, hardcodedStrings []HardcodedString, yellow *color.Color) {
	fmt.Fprintln(w, "")
	yellow.Fprintln(w, "🔍 HARDCODED STRINGS (Detailed Analysis)")
	fmt.Fprintln(w, "---------------------------------------")

	limit := 20
	if len(hardcodedStrings) < limit {
		limit = len(hardcodedStrings)
	}

	for i := 0; i < limit; i++ {
		hs := hardcodedStrings[i]
		fmt.Fprintf(w, "\n📁 %s:%d\n", GetRelativePath(hs.Location.File), hs.Location.Line)
		fmt.Fprintf(w, "   Found: %s\n", hs.FullClassValue)

		if len(hs.Suggestion.Constants) == 1 {
			// Single constant - simple one-line format
			fmt.Fprintf(w, "   Suggestion: Use ui.%s\n", hs.Suggestion.Constants[0])
		} else if len(hs.Suggestion.Constants) > 1 {
			// Multiple constants - show analysis breakdown
			fmt.Fprintln(w, "   Analysis:")
			for _, analysis := range hs.Suggestion.Analysis {
				switch analysis.Match {
				case MatchExact:
					fmt.Fprintf(w, "     • %q → ui.%s ✅\n", analysis.ClassName, analysis.Suggestion)
				case MatchNone:
					// Check if it's invalid or just bypassed
					if hs.Suggestion.HasInvalid && contains(hs.Suggestion.InvalidClasses, analysis.ClassName) {
						fmt.Fprintf(w, "     • %q → ❌ INVALID (doesn't exist in CSS)\n", analysis.ClassName)
					} else if analysis.Context == "valid CSS (no constant)" {
						fmt.Fprintf(w, "     • %q → Valid CSS (no constant needed)\n", analysis.ClassName)
					} else {
						fmt.Fprintf(w, "     • %q → No constant available\n", analysis.ClassName)
					}
				}
			}

			// Only show replacement suggestion if no invalid classes
			if len(hs.Suggestion.Constants) > 0 && !hs.Suggestion.HasInvalid {
				constList := make([]string, len(hs.Suggestion.Constants))
				for i, c := range hs.Suggestion.Constants {
					constList[i] = "ui." + c
				}
				suggestion := fmt.Sprintf("{ %s }", strings.Join(constList, ", "))
				if hs.Suggestion.HasUnmatched {
					fmt.Fprintf(w, "   ⚠️  Partial Match: Replace with %s\n", suggestion)
					fmt.Fprintf(w, "   ⚠️  WARNING: This will lose the following classes: %s\n",
						strings.Join(hs.Suggestion.UnmatchedClasses, ", "))
				} else {
					fmt.Fprintf(w, "   Suggestion: Replace with %s\n", suggestion)
				}
			} else if hs.Suggestion.HasInvalid {
				fmt.Fprintf(w, "   ❌ Cannot suggest replacement - contains invalid class%s: %s\n",
					pluralize(len(hs.Suggestion.InvalidClasses)),
					strings.Join(hs.Suggestion.InvalidClasses, ", "))
			} else if hs.Suggestion.HasUnmatched {
				fmt.Fprintf(w, "   ⚠️  No matching constants available\n")
			}
		} else {
			// No constants - show unmatched analysis
			if hs.Suggestion.HasUnmatched {
				fmt.Fprintf(w, "   ⚠️  No matching constants available\n")
			}
		}

		fmt.Fprintf(w, "   Context: %s\n", hs.LineContent)
	}

	if len(hardcodedStrings) > limit {
		fmt.Fprintf(w, "\n... and %d more\n", len(hardcodedStrings)-limit)
	}
}

// printProgressBar prints a visual progress bar
func printProgressBar(w io.Writer, percentage float64) {
	barWidth := 20
	filled := int(percentage / 100 * float64(barWidth))

	fmt.Fprint(w, "[")
	for i := 0; i < barWidth; i++ {
		if i < filled {
			fmt.Fprint(w, "█")
		} else {
			fmt.Fprint(w, "░")
		}
	}
	fmt.Fprintf(w, "] %.1f%%\n", percentage)
}

// pluralize returns "s" if count != 1
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// limitIssues applies max-issues-per-linter and max-same-issues constraints
func limitIssues(issues []Issue, config LintConfig) ([]Issue, int) {
	originalCount := len(issues)

	// Apply max-issues-per-linter
	if config.MaxIssuesPerLinter > 0 && len(issues) > config.MaxIssuesPerLinter {
		issues = issues[:config.MaxIssuesPerLinter]
	}

	// Apply max-same-issues (deduplication by message text)
	if config.MaxSameIssues > 0 {
		issues = deduplicateSameIssues(issues, config.MaxSameIssues)
	}

	truncatedCount := originalCount - len(issues)
	return issues, truncatedCount
}

// deduplicateSameIssues limits how many times the same message appears
func deduplicateSameIssues(issues []Issue, maxSame int) []Issue {
	messageCounts := make(map[string]int)
	var filtered []Issue

	for _, issue := range issues {
		count := messageCounts[issue.Text]
		if count < maxSame {
			filtered = append(filtered, issue)
			messageCounts[issue.Text]++
		}
	}

	return filtered
}
