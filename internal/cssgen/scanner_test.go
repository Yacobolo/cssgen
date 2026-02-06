package cssgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindClassColumn(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		className string
		wantCol   int
	}{
		{
			name:      "single class",
			line:      `<div class="btn">`,
			className: "btn",
			wantCol:   13, // Position of 'b' in "btn"
		},
		{
			name:      "multiple classes - first",
			line:      `<div class="btn btn--primary">`,
			className: "btn",
			wantCol:   13,
		},
		{
			name:      "multiple classes - second",
			line:      `<div class="btn btn--primary">`,
			className: "btn--primary",
			wantCol:   17, // Position of 'b' in "btn--primary"
		},
		{
			name:      "with leading spaces",
			line:      `  <div class="btn btn--outline">`,
			className: "btn--outline",
			wantCol:   19, // Accounts for leading spaces
		},
		{
			name:      "single quotes",
			line:      `<div class='icon nav-item-icon'>`,
			className: "nav-item-icon",
			wantCol:   18,
		},
		{
			name:      "class not found",
			line:      `<div class="btn">`,
			className: "nonexistent",
			wantCol:   0, // Returns 0 to signal fallback needed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findClassColumn(tt.line, tt.className)
			require.Equal(t, tt.wantCol, got)
		})
	}
}

func TestIsTemplGenerated(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "standard templ generated (_templ.go)",
			path:     "internal/web/features/sidebar_templ.go",
			expected: true,
		},
		{
			name:     "alternate templ generated (.templ.go)",
			path:     "internal/web/features/sidebar.templ.go",
			expected: true,
		},
		{
			name:     "regular go file",
			path:     "internal/api/handlers.go",
			expected: false,
		},
		{
			name:     "templ source file",
			path:     "internal/web/features/sidebar.templ",
			expected: false,
		},
		{
			name:     "nested path with _templ.go",
			path:     "internal/web/features/common/components/datatable_templ.go",
			expected: true,
		},
		{
			name:     "file with templ in name but not generated",
			path:     "internal/templates/handler.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTemplGenerated(tt.path)
			require.Equal(t, tt.expected, got, "isTemplGenerated(%q)", tt.path)
		})
	}
}

func TestShouldSkipFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "skip templ generated",
			path:     "internal/web/sidebar_templ.go",
			expected: true,
		},
		{
			name:     "scan templ source",
			path:     "internal/web/sidebar.templ",
			expected: false,
		},
		{
			name:     "scan regular go",
			path:     "internal/api/handlers.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipFile(tt.path)
			require.Equal(t, tt.expected, got, "shouldSkipFile(%q)", tt.path)
		})
	}
}

// Integration test: Verify filtering works end-to-end
func TestExpandGlobPatternsFiltersGeneratedFiles(t *testing.T) {
	// This test requires actual .templ and _templ.go files to exist
	// It validates that the filtering actually works in practice

	patterns := []string{"internal/web/features/**/*.go"}
	files, err := expandGlobPatterns(patterns)
	require.NoError(t, err)

	// Verify no _templ.go files in results
	for _, file := range files {
		require.False(t, isTemplGenerated(file),
			"Found generated file in results: %s", file)
	}
}
