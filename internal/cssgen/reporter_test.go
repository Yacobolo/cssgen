package cssgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildCaretIndicator(t *testing.T) {
	reporter := &Reporter{}

	tests := []struct {
		name       string
		sourceLine string
		column     int
		want       string
	}{
		{
			name:       "spaces only",
			sourceLine: "  <div class=\"btn\">",
			column:     15,
			want:       "              ^", // 14 spaces + caret
		},
		{
			name:       "tabs and spaces",
			sourceLine: "\t\t<button class=\"icon\">",
			column:     17,
			want:       "\t\t              ^", // 2 tabs + 14 spaces + caret (column 17 in string)
		},
		{
			name:       "start of line",
			sourceLine: "class=\"btn\"",
			column:     1,
			want:       "^",
		},
		{
			name:       "column 0 fallback",
			sourceLine: "some line",
			column:     0,
			want:       "^",
		},
		{
			name:       "column beyond line length",
			sourceLine: "short",
			column:     100,
			want:       "     ^", // Pads to line length only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reporter.buildCaretIndicator(tt.sourceLine, tt.column)
			require.Equal(t, tt.want, got)
		})
	}
}
