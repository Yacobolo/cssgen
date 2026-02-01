// Package cssgen provides CSS-to-Go code generation for type-safe CSS class constants.
package cssgen

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// AnalyzeClasses builds inheritance graph and resolves full class names
func AnalyzeClasses(classes []*CSSClass) error {
	// Build a map for quick lookup
	classMap := make(map[string]*CSSClass)
	for _, class := range classes {
		classMap[class.Name] = class
	}

	// Detect BEM patterns and link parent classes
	for _, class := range classes {
		base, isModifier := detectBEMPattern(class.Name)

		if isModifier && base != "" {
			// Link to parent class
			if parent, exists := classMap[base]; exists {
				class.ParentClass = parent
			}
		} else {
			// Standalone class (utility or base)
			class.IsUtility = !isModifier
		}

		// Generate Go name (initial)
		class.GoName = toGoName(class.Name)

		// NO FullClasses assignment - we use 1:1 mapping (Name directly)
		// ParentClass linking is kept for comment generation only
	}

	// Compute property diffs for modifiers
	for _, class := range classes {
		if class.ParentClass != nil {
			class.PropertyDiff = DiffProperties(class, class.ParentClass)
		}
	}

	// Resolve GoName collisions by adding numeric suffixes
	goNameMap := make(map[string][]*CSSClass)
	for _, class := range classes {
		goNameMap[class.GoName] = append(goNameMap[class.GoName], class)
	}

	for goName, classList := range goNameMap {
		if len(classList) > 1 {
			// Collision detected - add numeric suffix
			for i, class := range classList {
				if i > 0 {
					class.GoName = fmt.Sprintf("%s%d", goName, i+1)
				}
				// First one keeps the original name
			}
		}
	}

	return nil
}

// DiffProperties compares modifier properties to base class
func DiffProperties(modifier, base *CSSClass) *PropertyDiff {
	if base == nil {
		return nil
	}

	diff := &PropertyDiff{
		Added:     make(map[string]string),
		Changed:   make(map[string]string),
		Unchanged: []string{},
	}

	// Find changed and unchanged properties
	for name, modValue := range modifier.Properties {
		if baseValue, exists := base.Properties[name]; exists {
			if baseValue != modValue {
				diff.Changed[name] = modValue
			} else {
				diff.Unchanged = append(diff.Unchanged, name)
			}
		} else {
			diff.Added[name] = modValue
		}
	}

	// Sort unchanged for determinism
	sort.Strings(diff.Unchanged)

	return diff
}

// detectBEMPattern identifies base class from modifier naming
func detectBEMPattern(className string) (base string, isModifier bool) {
	// Standard BEM modifier: btn--primary
	if strings.Contains(className, "--") {
		parts := strings.Split(className, "--")
		return parts[0], true
	}

	// BEM element: card__header
	if strings.Contains(className, "__") {
		parts := strings.Split(className, "__")
		return parts[0], true
	}

	// No BEM pattern: utility class
	return "", false
}

// mergeConflicts handles duplicate class names across files
func mergeConflicts(classes []*CSSClass) ([]*CSSClass, []string) {
	classMap := make(map[string]*CSSClass)
	warnings := []string{}

	for _, class := range classes {
		existing, found := classMap[class.Name]

		if !found {
			classMap[class.Name] = class
			continue
		}

		// Merge properties
		for k, v := range class.Properties {
			existing.Properties[k] = v
		}

		// Merge pseudo-states
		for _, ps := range class.PseudoStates {
			if !contains(existing.PseudoStates, ps) {
				existing.PseudoStates = append(existing.PseudoStates, ps)
			}
		}

		// Warn about conflict
		warnings = append(warnings, fmt.Sprintf(
			"Duplicate class '%s' found in %s and %s - properties merged",
			class.Name, existing.SourceFile, class.SourceFile,
		))
	}

	// Convert map back to slice
	result := make([]*CSSClass, 0, len(classMap))
	for _, class := range classMap {
		result = append(result, class)
	}

	return result, warnings
}

// toGoName converts kebab-case to PascalCase
func toGoName(className string) string {
	// Remove leading dot if present
	name := strings.TrimPrefix(className, ".")

	// Remove leading underscore for internal classes but remember it
	isInternal := strings.HasPrefix(name, "_")
	name = strings.TrimPrefix(name, "_")

	// Split on - and __
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_'
	})

	// PascalCase
	for i, part := range parts {
		if len(part) > 0 {
			// Capitalize first letter
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
	}

	result := strings.Join(parts, "")

	// Add underscore prefix back for internal classes
	if isInternal {
		result = "_" + result
	}

	return result
}
