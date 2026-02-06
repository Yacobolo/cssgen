package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cssgen",
	Short: "CSS constant generator and linter for Go/templ projects",
	Long: `Type-safe CSS class constants with 1:1 mapping.
Each CSS class becomes exactly one Go constant.
Composition happens in templates: { ui.Btn, ui.BtnPrimary }`,
	// Default behavior: run generate when no subcommand is given.
	// We must call loadConfig here because PreRunE of generateCmd
	// is not triggered when delegating via rootCmd.RunE.
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := loadConfig(cmd); err != nil {
			return err
		}
		return runGenerate(generateCmd, nil)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Global persistent flags (inherited by all subcommands)
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress all output (exit code only)")
	rootCmd.PersistentFlags().String("package", "ui", "Go package name")
	rootCmd.PersistentFlags().Bool("color", false, "Force color output")
	rootCmd.PersistentFlags().String("config", ".cssgen.yaml", "Config file path")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(versionCmd)
}
