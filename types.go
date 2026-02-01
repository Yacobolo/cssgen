package cssgen

// PropertyDiff tracks changes between modifier and base
type PropertyDiff struct {
	Added     map[string]string // New properties in modifier
	Changed   map[string]string // Properties that override base
	Unchanged []string          // Properties inherited as-is
}

// PropertyCategory groups related CSS properties
type PropertyCategory string

// Property categories for organizing CSS properties
const (
	CategoryVisual     PropertyCategory = "Visual"
	CategoryLayout     PropertyCategory = "Layout"
	CategoryTypography PropertyCategory = "Typography"
	CategoryEffects    PropertyCategory = "Effects"
	CategoryTokens     PropertyCategory = "Tokens"
	CategoryInternal   PropertyCategory = "Internal"
)

// CategorizedProperty represents a property with its category
type CategorizedProperty struct {
	Name     string
	Value    string
	Category PropertyCategory
	IsToken  bool // True if value matches var(--ui-*)
}

// PseudoStateProperties tracks property changes in pseudo-states
type PseudoStateProperties struct {
	PseudoState string            // ":hover", ":focus", etc.
	Changes     map[string]string // Properties that change in this state
}

// CSSClass represents a parsed CSS class with full context
type CSSClass struct {
	Name   string // "btn--primary"
	GoName string // "BtnPrimary"
	// FullClasses field REMOVED - no longer needed with 1:1 mapping
	Layer                 string                  // "components"
	Properties            map[string]string       // CSS properties (cleaned)
	ParentClass           *CSSClass               // Link to base class (for comments/context only)
	PseudoStates          []string                // [":hover", ":focus"] - included in comments
	PseudoStateProperties []PseudoStateProperties // Property changes in pseudo-states
	PropertyDiff          *PropertyDiff           // Diff vs. parent class
	Intent                string                  // Human intent from @intent comment
	IsUtility             bool                    // True if atomic utility class (no BEM)
	IsInternal            bool                    // True if starts with _ (skip public const)
	SourceFile            string                  // For debugging/conflict resolution
}

// Layer represents a CSS cascade layer with priority
type Layer struct {
	Name    string
	Classes []*CSSClass
	Order   int // For priority (base=0, components=1, utilities=2)
}

// Config holds generator configuration
type Config struct {
	SourceDir          string   // "web/ui/src/styles"
	OutputDir          string   // "internal/web/ui" (output directory for generated files)
	PackageName        string   // "ui"
	Includes           []string // ["layers/components/**/*.css", "layers/utilities.css"]
	Verbose            bool     // Enable debug logging
	LayerInferFromPath bool     // Infer layer from file path (default: true)
	Format             string   // Output format: "markdown", "compact" (default: "markdown")
	PropertyLimit      int      // Max properties to show per category (default: 5)
	ShowInternal       bool     // Show -webkit-* properties (default: false)
	ExtractIntent      bool     // Parse @intent comments (default: true)
}

// GenerateResult contains generation stats
type GenerateResult struct {
	ClassesGenerated int
	FilesScanned     int
	IntentsExtracted int // Number of @intent comments extracted
	Warnings         []string
	Errors           []error
}

// OutputFormat represents the linter output format
type OutputFormat string

const (
	// OutputIssues shows only errors/warnings in golangci-lint format (CI-friendly)
	OutputIssues OutputFormat = "issues"
	// OutputSummary shows statistics and Quick Wins only (weekly reports)
	OutputSummary OutputFormat = "summary"
	// OutputFull shows issues + statistics + Quick Wins (interactive development)
	OutputFull OutputFormat = "full"
	// OutputJSON exports structured data in JSON format (tooling integration)
	OutputJSON OutputFormat = "json"
	// OutputMarkdown generates a Markdown report (shareable reports)
	OutputMarkdown OutputFormat = "markdown"
)
