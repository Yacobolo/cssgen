package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a default .cssgen.yaml config file",
	Long:  `Create a .cssgen.yaml configuration file in the current directory with sensible defaults.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		force, _ := cmd.Flags().GetBool("force")

		if _, err := os.Stat(".cssgen.yaml"); err == nil && !force {
			return fmt.Errorf(".cssgen.yaml already exists (use --force to overwrite)")
		}

		if err := os.WriteFile(".cssgen.yaml", []byte(defaultConfig), 0644); err != nil {
			return fmt.Errorf("writing config file: %w", err)
		}

		fmt.Println("Created .cssgen.yaml")
		return nil
	},
}

const defaultConfig = `# cssgen configuration
# Docs: https://github.com/yacobolo/cssgen

# Shared settings
package: ui
verbose: false

# Generation settings
generate:
  source: web/ui/src/styles
  output-dir: internal/web/ui
  include:
    - "layers/components/**/*.css"
    - "layers/utilities.css"
    - "layers/base.css"
  format: markdown         # markdown | compact
  property-limit: 5
  show-internal: false
  extract-intent: true
  infer-layer: true

# Linting settings
lint:
  paths:
    - "internal/web/features/**/*.templ"
    - "internal/web/features/**/*.go"
  strict: false
  threshold: 0.0
  output-format: issues    # issues | summary | full | json | markdown
  max-issues-per-linter: 0 # 0 = unlimited
  max-same-issues: 0       # 0 = unlimited
  print-lines: true
  print-linter-name: true
`

func init() {
	initCmd.Flags().Bool("force", false, "Overwrite existing config file")
}
