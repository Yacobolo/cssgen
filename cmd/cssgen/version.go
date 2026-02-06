package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=1.0.0" ./cmd/cssgen
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of cssgen",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("cssgen %s\n", version)
	},
}
