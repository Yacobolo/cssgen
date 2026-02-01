package cssgen

// Issue represents a single linting violation in golangci-lint format
type Issue struct {
	FromLinter  string       `json:"FromLinter"`  // "csslint"
	Text        string       `json:"Text"`        // "invalid CSS class \"btn--outline\" not found in stylesheet"
	Severity    string       `json:"Severity"`    // "", "warning", "error"
	SourceLines []string     `json:"SourceLines"` // Lines of code with issue
	Pos         IssuePos     `json:"Pos"`         // File location
	LineRange   *LineRange   `json:"LineRange"`   // Optional range
	Replacement *Replacement `json:"Replacement"` // Optional fix suggestion
}

// IssuePos specifies the exact location of an issue
type IssuePos struct {
	Filename string `json:"Filename"` // "internal/web/features/scheduleview/pages/scheduleview.templ"
	Line     int    `json:"Line"`     // 35
	Column   int    `json:"Column"`   // 15 (1-based, exact start of invalid class)
}

// LineRange specifies a range of lines
type LineRange struct {
	From int `json:"From"`
	To   int `json:"To"`
}

// Replacement provides automated fix suggestion (future --fix flag)
type Replacement struct {
	NewText      string // "ui.Icon" or "btn--outlined"
	InlineLength int    // Length of text to replace
}

// IssueSeverity constants
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityInfo    = ""
)

// IssueType constants matching linter categories
const (
	IssueInvalidClass   = "invalid CSS class %q not found in stylesheet"
	IssueHardcodedClass = "hardcoded CSS class %q should use %s constant"
	IssueUnusedConstant = "exported constant %s is unused"
)
