package cssgen

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// parserState maintains context while parsing CSS
type parserState struct {
	currentLayer  string
	inferredLayer string               // From file path
	classes       map[string]*CSSClass // Use map to deduplicate during parsing
	fullContent   string               // For intent extraction
	config        Config               // Configuration for parsing
}

// ParseCSS parses CSS content and returns structured classes
func ParseCSS(content string, filename string, inferredLayer string, config Config) ([]*CSSClass, error) {
	state := &parserState{
		currentLayer:  "",
		inferredLayer: inferredLayer,
		classes:       make(map[string]*CSSClass),
		fullContent:   content,
		config:        config,
	}

	lexer := css.NewLexer(parse.NewInputString(content))

	for {
		tt, text := lexer.Next()
		if tt == css.ErrorToken {
			// ErrorToken at EOF is normal - just break
			break
		}

		// Track @layer declarations
		if tt == css.AtKeywordToken && string(text) == "@layer" {
			state.handleLayerDeclaration(lexer)
			continue
		}

		// Look for class selectors followed by { declarations }
		if tt == css.DelimToken && len(text) > 0 && text[0] == '.' {
			// This is a class selector
			state.handleClassRule(lexer, filename)
		}
	}

	// Extract intent if enabled
	if config.ExtractIntent {
		for _, class := range state.classes {
			class.Intent = extractIntent(content, class.Name)
		}
	}

	// Convert map to slice
	result := make([]*CSSClass, 0, len(state.classes))
	for _, class := range state.classes {
		result = append(result, class)
	}

	return result, nil
}

// parseFile reads and parses a single CSS file
func parseFile(path string, config Config) ([]*CSSClass, error) {
	// #nosec G304 - path comes from trusted configuration
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Infer layer from path if enabled
	inferredLayer := ""
	if config.LayerInferFromPath {
		inferredLayer = inferLayerFromPath(path, config.SourceDir)
	}

	return ParseCSS(string(content), path, inferredLayer, config)
}

// inferLayerFromPath extracts layer name from file path
// Pattern: layers/{layerName}/**/*.css → layerName
func inferLayerFromPath(filePath, sourceDir string) string {
	// Normalize paths by converting all backslashes to forward slashes
	// This handles Windows paths correctly on all platforms
	path := strings.ReplaceAll(filePath, "\\", "/")
	srcDir := strings.ReplaceAll(sourceDir, "\\", "/")

	// Normalize - ensure sourceDir doesn't end with /
	srcDir = strings.TrimSuffix(srcDir, "/")

	// Remove source dir prefix
	relPath := strings.TrimPrefix(path, srcDir)
	relPath = strings.TrimPrefix(relPath, "/")

	// Match pattern: layers/{layer}/...
	parts := strings.Split(relPath, "/")
	if len(parts) >= 2 && parts[0] == "layers" {
		// Extract layer name from second part
		layerName := parts[1]
		// Handle case where file IS the layer (e.g., layers/utilities.css)
		layerName = strings.TrimSuffix(layerName, ".css")
		return layerName
	}

	// Fallback for root-level files
	if strings.Contains(relPath, "base.css") {
		return "base"
	}
	if strings.Contains(relPath, "utilities.css") {
		return "utilities"
	}
	if strings.Contains(relPath, "reset.css") {
		return "reset"
	}

	return "n/a"
}

// handleLayerDeclaration processes @layer declarations
func (s *parserState) handleLayerDeclaration(lexer *css.Lexer) {
	// Read tokens until we hit { or ;
	var layerName string

	for {
		tt, text := lexer.Next()
		if tt == css.ErrorToken {
			break
		}

		if tt == css.IdentToken {
			layerName = string(text)
		}

		if tt == css.LeftBraceToken {
			// @layer name { ... }
			if layerName != "" {
				s.currentLayer = layerName
			}
			return
		}

		if tt == css.SemicolonToken {
			// @layer name1, name2;
			return
		}
	}
}

// handleClassRule processes a class selector and its declarations
func (s *parserState) handleClassRule(lexer *css.Lexer, filename string) {
	// At this point we've seen a '.', read the class name
	tt, classNameBytes := lexer.Next()
	if tt != css.IdentToken {
		return
	}

	firstClassName := string(classNameBytes)

	// Collect all class names and pseudo-states before the opening brace
	type selectorInfo struct {
		className    string
		pseudoStates []string
	}

	selectors := []selectorInfo{{className: firstClassName, pseudoStates: []string{}}}
	currentIdx := 0

	for {
		tt, text := lexer.Next()
		if tt == css.ErrorToken {
			return
		}

		// Handle additional class names in compound selectors (.foo.bar)
		if tt == css.DelimToken && len(text) > 0 && text[0] == '.' {
			tt2, className2 := lexer.Next()
			if tt2 == css.IdentToken {
				// Found another class in compound selector
				selectors = append(selectors, selectorInfo{
					className:    string(className2),
					pseudoStates: []string{},
				})
				currentIdx = len(selectors) - 1
			}
			continue
		}

		// Track pseudo-classes/elements
		if tt == css.ColonToken {
			tt2, text2 := lexer.Next()
			if tt2 == css.IdentToken {
				pseudoName := string(text2)

				// Handle functional pseudo-classes that contain selectors
				if pseudoName == "not" || pseudoName == "is" || pseudoName == "where" {
					// Next should be '('
					tt3, _ := lexer.Next()
					if tt3 == css.LeftParenthesisToken {
						// Look for class selectors inside parentheses
						depth := 1 // Track nesting level
						for depth > 0 {
							ttInner, textInner := lexer.Next()

							if ttInner == css.ErrorToken {
								break
							}

							// Track nested parentheses
							if ttInner == css.LeftParenthesisToken {
								depth++
							} else if ttInner == css.RightParenthesisToken {
								depth--
								if depth == 0 {
									break
								}
							}

							// Extract class selectors
							if ttInner == css.DelimToken && len(textInner) > 0 && textInner[0] == '.' {
								ttClass, classNameInner := lexer.Next()
								if ttClass == css.IdentToken {
									// Extract class from :not(.foo), :is(.bar), etc.
									selectors = append(selectors, selectorInfo{
										className:    string(classNameInner),
										pseudoStates: []string{},
									})
								}
							}
						}
					}
				} else {
					// Regular pseudo-class (:hover, :focus, etc.)
					selectors[currentIdx].pseudoStates = append(
						selectors[currentIdx].pseudoStates,
						":"+pseudoName,
					)
				}
			}
			continue
		}

		// Handle comma - start a new selector
		if tt == css.CommaToken {
			// Look for next class selector
			for {
				tt2, data2 := lexer.Next()
				if tt2 == css.ErrorToken || tt2 == css.LeftBraceToken {
					break
				}
				if tt2 == css.DelimToken && len(data2) > 0 && data2[0] == '.' {
					// Found another class
					tt3, className3 := lexer.Next()
					if tt3 == css.IdentToken {
						selectors = append(selectors, selectorInfo{
							className:    string(className3),
							pseudoStates: []string{},
						})
						currentIdx = len(selectors) - 1
						break
					}
				}
			}
			continue
		}

		if tt == css.LeftBraceToken {
			// Found the declaration block
			properties := s.extractDeclarations(lexer)

			// Apply properties to all collected selectors
			for _, sel := range selectors {
				// Create or update the base class (always)
				class, exists := s.classes[sel.className]
				if !exists {
					// Use explicit layer if set, otherwise use inferred layer
					layer := s.currentLayer
					if layer == "" && s.inferredLayer != "" {
						layer = s.inferredLayer
					}

					class = &CSSClass{
						Name:         sel.className,
						Layer:        layer,
						Properties:   make(map[string]string),
						PseudoStates: []string{},
						SourceFile:   filename,
						IsInternal:   strings.HasPrefix(sel.className, "_"),
					}
					s.classes[sel.className] = class
				}

				// If this selector has pseudo-states, track property changes
				if len(sel.pseudoStates) > 0 {
					// This is a pseudo-state variant (.btn:hover)
					for _, ps := range sel.pseudoStates {
						psp := PseudoStateProperties{
							PseudoState: ps,
							Changes:     make(map[string]string),
						}

						// Track which properties are different from base
						for prop, val := range properties {
							if baseVal, hasBase := class.Properties[prop]; !hasBase || baseVal != val {
								psp.Changes[prop] = val
							}
						}

						if len(psp.Changes) > 0 {
							class.PseudoStateProperties = append(
								class.PseudoStateProperties,
								psp,
							)
						}
					}

					// Also add to pseudo states list
					for _, ps := range sel.pseudoStates {
						if !contains(class.PseudoStates, ps) {
							class.PseudoStates = append(class.PseudoStates, ps)
						}
					}
				} else {
					// Regular class, merge properties
					for k, v := range properties {
						class.Properties[k] = v
					}
				}
			}

			return
		}
	}
}

// extractDeclarations reads property: value pairs until }
func (s *parserState) extractDeclarations(lexer *css.Lexer) map[string]string {
	props := make(map[string]string)

	var currentProp string
	var currentVal []string

	for {
		tt, text := lexer.Next()

		if tt == css.ErrorToken || tt == css.RightBraceToken {
			// Save last property
			if currentProp != "" && len(currentVal) > 0 {
				props[currentProp] = strings.TrimSpace(strings.Join(currentVal, ""))
			}
			break
		}

		switch {
		case tt == css.IdentToken && currentProp == "":
			// Start of property name
			currentProp = string(text)
		case tt == css.ColonToken && currentProp != "":
			// Separator between property and value
			continue
		case tt == css.SemicolonToken:
			// End of declaration
			if currentProp != "" && len(currentVal) > 0 {
				props[currentProp] = strings.TrimSpace(strings.Join(currentVal, ""))
			}
			currentProp = ""
			currentVal = nil
		case currentProp != "":
			// Part of the value
			currentVal = append(currentVal, string(text))
		}
	}

	return props
}

// cleanProperties formats properties as single-line comment
func cleanProperties(props map[string]string) string {
	if len(props) == 0 {
		return ""
	}

	// Sort keys for determinism
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Format: { key: value; key2: value2; }
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", k, props[k]))
	}

	result := "{ " + strings.Join(parts, "; ") + "; }"

	// Truncate for IDE tooltip readability
	if len(result) > 120 {
		result = result[:117] + "..."
	}

	return result
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// extractIntent looks for @intent comments above CSS rules
func extractIntent(content string, className string) string {
	lines := strings.Split(content, "\n")

	// Find the line with the class definition
	classLine := -1
	searchPattern := fmt.Sprintf(".%s", className)
	for i, line := range lines {
		if strings.Contains(line, searchPattern) && strings.Contains(line, "{") {
			classLine = i
			break
		}
	}

	if classLine == -1 {
		return ""
	}

	// Look backwards for @intent comment (max 10 lines)
	for i := classLine - 1; i >= 0 && i >= classLine-10; i-- {
		line := strings.TrimSpace(lines[i])

		// Stop at empty line or non-comment
		if line == "" || (!strings.HasPrefix(line, "/*") && !strings.HasPrefix(line, "*") && !strings.HasPrefix(line, "//")) {
			break
		}

		// Extract @intent directive
		if strings.Contains(line, "@intent") {
			// Extract text after @intent
			parts := strings.SplitN(line, "@intent", 2)
			if len(parts) == 2 {
				intent := strings.TrimSpace(parts[1])
				// Clean up comment markers
				intent = strings.TrimPrefix(intent, "*")
				intent = strings.TrimSuffix(intent, "*/")
				intent = strings.TrimSpace(intent)
				return intent
			}
		}
	}

	return ""
}
