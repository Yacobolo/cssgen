package cssgen

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	ignore "github.com/sabhiram/go-gitignore"
)

// ClassReference represents a CSS class reference found in code
type ClassReference struct {
	ClassName      string       // Individual class: "btn--ghost" (DEPRECATED: use FullClassValue)
	FullClassValue string       // Full attribute: "btn btn--ghost btn--sm"
	Location       FileLocation // Where it was found
	IsConstant     bool         // true if using ui.Foo, false if "foo"
	ConstName      string       // "Foo" if IsConstant is true
	LineContent    string       // The full line for context
}

// FileLocation tracks where a class reference was found
type FileLocation struct {
	File   string
	Line   int
	Column int    // 1-based column (exact start of class name)
	Text   string // Full line content for source display
}

// ScanStats tracks file scanning statistics
type ScanStats struct {
	FilesDiscovered int // Total files found by glob patterns
	FilesScanned    int // Files actually scanned (after filtering)
	FilesSkipped    int // Files skipped due to filtering
}

// scanPattern represents a regex pattern for finding class references
type scanPattern struct {
	name    string
	regex   *regexp.Regexp
	isConst bool
}

var (
	// Patterns for finding CSS class references
	// Ordered from most specific to least specific
	patterns = []scanPattern{
		// Constant usage (ui.Foo)
		{
			name:    "ui package constant",
			regex:   regexp.MustCompile(`ui\.([A-Z][a-zA-Z0-9]*)`),
			isConst: true,
		},

		// Hardcoded strings in various contexts
		{
			name:    "class attribute with quotes",
			regex:   regexp.MustCompile(`class="([^"]+)"`),
			isConst: false,
		},
		{
			name:    "class with string literal in braces",
			regex:   regexp.MustCompile(`class=\{\s*"([^"]+)"`),
			isConst: false,
		},
		{
			name:    "templ.Classes with string",
			regex:   regexp.MustCompile(`templ\.Classes\(\s*"([^"]+)"`),
			isConst: false,
		},
		{
			name:    "templ.KV with string",
			regex:   regexp.MustCompile(`templ\.KV\(\s*"([^"]+)"`),
			isConst: false,
		},
		{
			name:    "ds.Class call",
			regex:   regexp.MustCompile(`ds\.Class\(\s*"([^"]+)"`),
			isConst: false,
		},
	}

	// Regex to detect templ.Classes and templ.KV with comma-separated values
	templClassesMulti = regexp.MustCompile(`templ\.Classes\(([^)]+)\)`)
	templKVMulti      = regexp.MustCompile(`templ\.KV\(([^)]+)\)`)

	// Comment patterns to skip
	commentPattern = regexp.MustCompile(`^\s*//`)

	// gitignore caching
	gitIgnoreCache *ignore.GitIgnore
	gitIgnoreOnce  sync.Once
)

// isTemplGenerated checks if a file is a templ-generated Go file
// Handles both _templ.go and .templ.go suffix variations
func isTemplGenerated(path string) bool {
	return strings.HasSuffix(path, "_templ.go") ||
		strings.HasSuffix(path, ".templ.go")
}

// loadGitIgnore loads the .gitignore file once (thread-safe)
// Gracefully degrades if .gitignore doesn't exist
func loadGitIgnore() *ignore.GitIgnore {
	gitIgnoreOnce.Do(func() {
		// Try to load .gitignore from current directory
		gi, err := ignore.CompileIgnoreFile(".gitignore")
		if err != nil {
			// Gracefully degrade - no .gitignore is fine
			gitIgnoreCache = nil
			return
		}
		gitIgnoreCache = gi
	})
	return gitIgnoreCache
}

// shouldSkipFile determines if a file should be excluded from scanning
// Returns true if the file should be skipped, false otherwise
//
// Two-layer filtering:
// 1. Pattern check (fast): Skip *_templ.go files
// 2. Gitignore check (professional): Skip gitignored files (only for relative paths)
func shouldSkipFile(path string) bool {
	// Layer 1: Fast pattern check for templ-generated files
	if isTemplGenerated(path) {
		return true
	}

	// Layer 2: Check against .gitignore if available
	// Only apply gitignore to relative paths (paths within the project)
	// Absolute paths (like /tmp/...) should not be affected by project gitignore
	if !filepath.IsAbs(path) {
		gi := loadGitIgnore()
		if gi != nil && gi.MatchesPath(path) {
			return true
		}
	}

	return false
}

// ScanFiles scans files matching the given patterns for CSS class references
func ScanFiles(scanPatterns []string, verbose bool) ([]ClassReference, ScanStats, error) {
	files, stats, err := expandGlobPatternsWithStats(scanPatterns)
	if err != nil {
		return nil, stats, err
	}

	// Print one-line summary in verbose mode
	if verbose && stats.FilesSkipped > 0 {
		println("✓ Scanned", stats.FilesScanned, "files (skipped", stats.FilesSkipped, "generated/ignored files)")
	}

	var allRefs []ClassReference
	for _, file := range files {
		refs, err := scanFile(file)
		if err != nil {
			// Log warning but continue
			continue
		}
		allRefs = append(allRefs, refs...)
	}

	return allRefs, stats, nil
}

// expandGlobPatterns expands glob patterns to actual file paths
func expandGlobPatterns(patterns []string) ([]string, error) {
	var allFiles []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			// Deduplicate and only include files (not directories)
			if !seen[match] {
				info, err := os.Stat(match)
				if err == nil && !info.IsDir() {
					// Apply two-layer filtering
					if !shouldSkipFile(match) {
						allFiles = append(allFiles, match)
						seen[match] = true
					}
				}
			}
		}
	}

	return allFiles, nil
}

// expandGlobPatternsWithStats expands globs and tracks statistics
// Used when verbose output is enabled
func expandGlobPatternsWithStats(patterns []string) ([]string, ScanStats, error) {
	var allFiles []string
	seen := make(map[string]bool)
	stats := ScanStats{}

	for _, pattern := range patterns {
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			return nil, stats, err
		}

		for _, match := range matches {
			if !seen[match] {
				info, err := os.Stat(match)
				if err == nil && !info.IsDir() {
					stats.FilesDiscovered++

					if shouldSkipFile(match) {
						stats.FilesSkipped++
					} else {
						allFiles = append(allFiles, match)
						seen[match] = true
						stats.FilesScanned++
					}
				}
			}
		}
	}

	return allFiles, stats, nil
}

// scanFile scans a single file for CSS class references
func scanFile(filePath string) ([]ClassReference, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var refs []ClassReference
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		lineRefs := extractClassesFromLine(line, lineNum, filePath)
		refs = append(refs, lineRefs...)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return refs, nil
}

// findClassColumn locates the exact column where className starts within line
// For multi-class strings like "btn btn--sm", finds the first token
func findClassColumn(line string, fullClassString string) int {
	// For multi-class strings, find the anchor (first token)
	tokens := strings.Fields(fullClassString)
	searchTarget := fullClassString
	if len(tokens) > 0 {
		searchTarget = tokens[0] // Use first token as anchor
	}

	// Strategy 1: Look for the class name in a class= attribute first
	classAttrIdx := strings.Index(line, "class=")
	if classAttrIdx != -1 {
		// Find opening quote
		quoteIdx := strings.IndexAny(line[classAttrIdx:], `"'`)
		if quoteIdx != -1 {
			searchStart := classAttrIdx + quoteIdx + 1

			// Find the search target within the attribute
			classesStr := line[searchStart:]
			endQuote := strings.IndexAny(classesStr, `"'`)
			if endQuote != -1 {
				classesStr = classesStr[:endQuote]
			}

			// Find our target token
			idx := strings.Index(classesStr, searchTarget)
			if idx != -1 {
				return searchStart + idx + 1 // 1-based column
			}
		}
	}

	// Strategy 2: Search for pattern in quotes
	searchPattern := `"` + searchTarget + `"`
	idx := strings.Index(line, searchPattern)
	if idx != -1 {
		return idx + 2 // +1 for 1-based, +1 to skip quote
	}

	// Strategy 3: Direct search
	idx = strings.Index(line, searchTarget)
	if idx != -1 {
		return idx + 1
	}

	// Fallback
	return 0
}

// extractClassesFromLine extracts all CSS class references from a line
func extractClassesFromLine(line string, lineNum int, file string) []ClassReference {
	// Skip comments
	if commentPattern.MatchString(line) {
		return nil
	}

	var refs []ClassReference

	// Check if line contains templ.Classes or templ.KV - use specialized handlers
	hasTemplClasses := strings.Contains(line, "templ.Classes(")
	hasTemplKV := strings.Contains(line, "templ.KV(")

	if hasTemplClasses {
		refs = append(refs, extractFromTemplClasses(line, lineNum, file)...)
	}
	if hasTemplKV {
		refs = append(refs, extractFromTemplKV(line, lineNum, file)...)
	}

	// If we handled templ functions, skip standard pattern matching for those
	// to avoid duplicates
	if hasTemplClasses || hasTemplKV {
		// templ functions already handled, don't apply other patterns
		// to avoid duplicates
		return refs
	}

	// Standard pattern matching for other cases
	for _, pattern := range patterns {
		matches := pattern.regex.FindAllStringSubmatchIndex(line, -1)
		for _, match := range matches {
			if len(match) < 4 {
				continue
			}

			captured := line[match[2]:match[3]]

			ref := ClassReference{
				Location: FileLocation{
					File:   file,
					Line:   lineNum,
					Column: match[0] + 1, // 1-indexed (relative to original line)
					Text:   strings.TrimSpace(line),
				},
				LineContent: strings.TrimSpace(line),
				IsConstant:  pattern.isConst,
			}

			if pattern.isConst {
				// ui.Foo -> Foo
				ref.ConstName = captured
			} else {
				// Hardcoded string: Store FULL value, not split
				ref.FullClassValue = captured
			}

			refs = append(refs, ref)
		}
	}

	return refs
}

// extractFromTemplClasses extracts class names from templ.Classes(...) calls
// Handles: templ.Classes("foo", "bar", ui.Baz, templ.KV(...))
func extractFromTemplClasses(line string, lineNum int, file string) []ClassReference {
	var refs []ClassReference

	matches := templClassesMulti.FindAllStringSubmatchIndex(line, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		content := line[match[2]:match[3]]
		refs = append(refs, parseTemplArguments(content, lineNum, file, line)...)
	}

	return refs
}

// extractFromTemplKV extracts class names from templ.KV(...) calls
// Handles: templ.KV("foo", condition)
func extractFromTemplKV(line string, lineNum int, file string) []ClassReference {
	var refs []ClassReference

	matches := templKVMulti.FindAllStringSubmatchIndex(line, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		content := line[match[2]:match[3]]
		// For KV, only the first argument is the class name
		parts := splitTemplArgs(content)
		if len(parts) > 0 {
			refs = append(refs, parseTemplArguments(parts[0], lineNum, file, line)...)
		}
	}

	return refs
}

// parseTemplArguments parses arguments inside templ functions
// Handles: "foo", ui.Bar, "baz qux"
func parseTemplArguments(args string, lineNum int, file string, fullLine string) []ClassReference {
	var refs []ClassReference

	// Split by commas (simple approach - doesn't handle nested parens)
	parts := splitTemplArgs(args)

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check if it's a ui constant
		if strings.HasPrefix(part, "ui.") {
			constName := strings.TrimPrefix(part, "ui.")
			refs = append(refs, ClassReference{
				Location: FileLocation{
					File:   file,
					Line:   lineNum,
					Column: strings.Index(fullLine, part) + 1,
					Text:   strings.TrimSpace(fullLine),
				},
				LineContent: strings.TrimSpace(fullLine),
				IsConstant:  true,
				ConstName:   constName,
			})
			continue
		}

		// Check if it's a string literal
		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			classStr := strings.Trim(part, `"`)
			// Store full class value instead of splitting
			refs = append(refs, ClassReference{
				Location: FileLocation{
					File:   file,
					Line:   lineNum,
					Column: strings.Index(fullLine, classStr) + 1,
					Text:   strings.TrimSpace(fullLine),
				},
				LineContent:    strings.TrimSpace(fullLine),
				IsConstant:     false,
				FullClassValue: classStr,
			})
		}
	}

	return refs
}

// splitTemplArgs splits comma-separated arguments
// Simple splitter - doesn't handle nested function calls
func splitTemplArgs(s string) []string {
	var parts []string
	var current strings.Builder
	parenDepth := 0

	for _, r := range s {
		switch r {
		case '(':
			parenDepth++
			current.WriteRune(r)
		case ')':
			parenDepth--
			current.WriteRune(r)
		case ',':
			if parenDepth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// GetRelativePath returns a relative path from the current working directory
func GetRelativePath(absPath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return absPath
	}

	rel, err := filepath.Rel(cwd, absPath)
	if err != nil {
		return absPath
	}

	return rel
}
