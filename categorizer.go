package cssgen

import (
	"sort"
	"strings"
)

// propertyCategories maps CSS property names to categories
var propertyCategories = map[string]PropertyCategory{
	// Visual
	"background":          CategoryVisual,
	"background-color":    CategoryVisual,
	"background-image":    CategoryVisual,
	"background-size":     CategoryVisual,
	"background-position": CategoryVisual,
	"background-repeat":   CategoryVisual,
	"color":               CategoryVisual,
	"border":              CategoryVisual,
	"border-color":        CategoryVisual,
	"border-radius":       CategoryVisual,
	"border-width":        CategoryVisual,
	"border-style":        CategoryVisual,
	"border-top":          CategoryVisual,
	"border-right":        CategoryVisual,
	"border-bottom":       CategoryVisual,
	"border-left":         CategoryVisual,
	"border-inline":       CategoryVisual,
	"border-block":        CategoryVisual,
	"box-shadow":          CategoryVisual,
	"opacity":             CategoryVisual,
	"outline":             CategoryVisual,
	"outline-color":       CategoryVisual,
	"outline-width":       CategoryVisual,
	"outline-style":       CategoryVisual,
	"fill":                CategoryVisual,
	"stroke":              CategoryVisual,

	// Layout
	"display":               CategoryLayout,
	"flex":                  CategoryLayout,
	"flex-direction":        CategoryLayout,
	"flex-wrap":             CategoryLayout,
	"flex-grow":             CategoryLayout,
	"flex-shrink":           CategoryLayout,
	"flex-basis":            CategoryLayout,
	"justify-content":       CategoryLayout,
	"align-items":           CategoryLayout,
	"align-self":            CategoryLayout,
	"align-content":         CategoryLayout,
	"gap":                   CategoryLayout,
	"row-gap":               CategoryLayout,
	"column-gap":            CategoryLayout,
	"grid":                  CategoryLayout,
	"grid-template-columns": CategoryLayout,
	"grid-template-rows":    CategoryLayout,
	"grid-template-areas":   CategoryLayout,
	"grid-column":           CategoryLayout,
	"grid-row":              CategoryLayout,
	"position":              CategoryLayout,
	"inset":                 CategoryLayout,
	"inset-block":           CategoryLayout,
	"inset-block-start":     CategoryLayout,
	"inset-block-end":       CategoryLayout,
	"inset-inline":          CategoryLayout,
	"inset-inline-start":    CategoryLayout,
	"inset-inline-end":      CategoryLayout,
	"top":                   CategoryLayout,
	"right":                 CategoryLayout,
	"bottom":                CategoryLayout,
	"left":                  CategoryLayout,
	"width":                 CategoryLayout,
	"height":                CategoryLayout,
	"inline-size":           CategoryLayout,
	"block-size":            CategoryLayout,
	"min-width":             CategoryLayout,
	"min-height":            CategoryLayout,
	"min-inline-size":       CategoryLayout,
	"min-block-size":        CategoryLayout,
	"max-width":             CategoryLayout,
	"max-height":            CategoryLayout,
	"max-inline-size":       CategoryLayout,
	"max-block-size":        CategoryLayout,
	"padding":               CategoryLayout,
	"padding-top":           CategoryLayout,
	"padding-right":         CategoryLayout,
	"padding-bottom":        CategoryLayout,
	"padding-left":          CategoryLayout,
	"padding-inline":        CategoryLayout,
	"padding-inline-start":  CategoryLayout,
	"padding-inline-end":    CategoryLayout,
	"padding-block":         CategoryLayout,
	"padding-block-start":   CategoryLayout,
	"padding-block-end":     CategoryLayout,
	"margin":                CategoryLayout,
	"margin-top":            CategoryLayout,
	"margin-right":          CategoryLayout,
	"margin-bottom":         CategoryLayout,
	"margin-left":           CategoryLayout,
	"margin-inline":         CategoryLayout,
	"margin-inline-start":   CategoryLayout,
	"margin-inline-end":     CategoryLayout,
	"margin-block":          CategoryLayout,
	"margin-block-start":    CategoryLayout,
	"margin-block-end":      CategoryLayout,
	"overflow":              CategoryLayout,
	"overflow-x":            CategoryLayout,
	"overflow-y":            CategoryLayout,
	"z-index":               CategoryLayout,
	"aspect-ratio":          CategoryLayout,
	"object-fit":            CategoryLayout,
	"object-position":       CategoryLayout,

	// Typography
	"font-family":          CategoryTypography,
	"font-size":            CategoryTypography,
	"font-weight":          CategoryTypography,
	"font-style":           CategoryTypography,
	"font-variant":         CategoryTypography,
	"font-variant-numeric": CategoryTypography,
	"line-height":          CategoryTypography,
	"letter-spacing":       CategoryTypography,
	"text-align":           CategoryTypography,
	"text-decoration":      CategoryTypography,
	"text-transform":       CategoryTypography,
	"text-overflow":        CategoryTypography,
	"white-space":          CategoryTypography,
	"word-break":           CategoryTypography,
	"word-wrap":            CategoryTypography,
	"hyphens":              CategoryTypography,

	// Effects
	"transition":                 CategoryEffects,
	"transition-property":        CategoryEffects,
	"transition-duration":        CategoryEffects,
	"transition-timing-function": CategoryEffects,
	"transition-delay":           CategoryEffects,
	"transform":                  CategoryEffects,
	"transform-origin":           CategoryEffects,
	"animation":                  CategoryEffects,
	"animation-name":             CategoryEffects,
	"animation-duration":         CategoryEffects,
	"animation-timing-function":  CategoryEffects,
	"animation-delay":            CategoryEffects,
	"animation-iteration-count":  CategoryEffects,
	"animation-direction":        CategoryEffects,
	"filter":                     CategoryEffects,
	"backdrop-filter":            CategoryEffects,
	"mix-blend-mode":             CategoryEffects,
	"clip-path":                  CategoryEffects,
	"mask":                       CategoryEffects,
}

// categorizeProperty determines the category of a CSS property
func categorizeProperty(name string) PropertyCategory {
	// Check exact match
	if cat, exists := propertyCategories[name]; exists {
		return cat
	}

	// Check prefixes for internal properties
	if strings.HasPrefix(name, "-webkit-") ||
		strings.HasPrefix(name, "-moz-") ||
		strings.HasPrefix(name, "-ms-") ||
		strings.HasPrefix(name, "-o-") {
		return CategoryInternal
	}

	// Check for flex-* and grid-* properties (catch-all for flex/grid)
	if strings.HasPrefix(name, "flex-") || strings.HasPrefix(name, "grid-") {
		return CategoryLayout
	}

	// Check for border-* properties
	if strings.HasPrefix(name, "border-") {
		return CategoryVisual
	}

	// Check for padding-* and margin-* properties
	if strings.HasPrefix(name, "padding-") || strings.HasPrefix(name, "margin-") {
		return CategoryLayout
	}

	// Default to Layout for unknown properties
	return CategoryLayout
}

// isTokenValue checks if a value uses design tokens
func isTokenValue(value string) bool {
	return strings.Contains(value, "var(--ui-")
}

// categorizeProperties groups properties by category
func categorizeProperties(props map[string]string) map[PropertyCategory][]CategorizedProperty {
	result := make(map[PropertyCategory][]CategorizedProperty)

	for name, value := range props {
		cat := categorizeProperty(name)
		prop := CategorizedProperty{
			Name:     name,
			Value:    value,
			Category: cat,
			IsToken:  isTokenValue(value),
		}
		result[cat] = append(result[cat], prop)
	}

	// Sort properties within each category
	for cat := range result {
		sort.Slice(result[cat], func(i, j int) bool {
			return result[cat][i].Name < result[cat][j].Name
		})
	}

	return result
}
