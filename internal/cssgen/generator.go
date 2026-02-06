package cssgen

import (
	"fmt"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

// Generate is the main entry point
func Generate(config Config) (*GenerateResult, error) {
	result := &GenerateResult{}

	// 1. Scan CSS files
	files, err := scanCSSFiles(config.SourceDir, config.Includes)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}
	result.FilesScanned = len(files)

	if config.Verbose {
		fmt.Printf("Found %d CSS files\n", len(files))
	}

	// 2. Parse all files
	classes, warnings, err := processFiles(files, config)
	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}
	result.Warnings = warnings

	// Count intents extracted
	for _, class := range classes {
		if class.Intent != "" {
			result.IntentsExtracted++
		}
	}

	if config.Verbose {
		fmt.Printf("Parsed %d classes\n", len(classes))
	}

	// 3. Analyze BEM patterns and build inheritance
	if err := AnalyzeClasses(classes); err != nil {
		return nil, fmt.Errorf("analyze failed: %w", err)
	}

	// 4. Merge conflicts
	classes, conflicts := mergeConflicts(classes)
	result.Warnings = append(result.Warnings, conflicts...)

	// 5. Filter internal classes
	publicClasses := make([]*CSSClass, 0, len(classes))
	for _, class := range classes {
		if !class.IsInternal {
			publicClasses = append(publicClasses, class)
		}
	}
	result.ClassesGenerated = len(publicClasses)

	if config.Verbose {
		fmt.Printf("Generated %d public constants (%d internal classes filtered)\n",
			len(publicClasses), len(classes)-len(publicClasses))
	}

	// 6. Generate Go file
	// Pass both public classes for constants AND all classes for AllCSSClasses map
	if err := WriteGoFile(publicClasses, classes, config, *result); err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	return result, nil
}

// scanCSSFiles finds all CSS files matching includes
func scanCSSFiles(sourceDir string, includes []string) ([]string, error) {
	var files []string

	for _, pattern := range includes {
		// Combine source dir with pattern
		fullPattern := filepath.Join(sourceDir, pattern)

		// Use doublestar for ** glob support
		matches, err := doublestar.FilepathGlob(fullPattern)
		if err != nil {
			return nil, fmt.Errorf("glob pattern %q: %w", pattern, err)
		}

		files = append(files, matches...)
	}

	// Remove duplicates
	seen := make(map[string]bool)
	unique := make([]string, 0, len(files))
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			unique = append(unique, f)
		}
	}

	return unique, nil
}

// processFiles parses all CSS files
func processFiles(files []string, config Config) ([]*CSSClass, []string, error) {
	var allClasses []*CSSClass
	var warnings []string

	for _, file := range files {
		if config.Verbose {
			fmt.Printf("Parsing %s\n", file)
		}

		classes, err := parseFile(file, config)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to parse %s: %v", file, err))
			continue
		}

		allClasses = append(allClasses, classes...)
	}

	return allClasses, warnings, nil
}
