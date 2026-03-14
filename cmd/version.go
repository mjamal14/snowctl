package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Set via ldflags at build time.
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print snowctl version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("snowctl %s\n", version)
		fmt.Printf("  commit:    %s\n", commit)
		fmt.Printf("  built:     %s\n", buildDate)
		fmt.Printf("  go:        %s\n", runtime.Version())
		fmt.Printf("  os/arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
