package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "v1.5.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "display version of buffalo-auth",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("buffalo-auth", Version)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
