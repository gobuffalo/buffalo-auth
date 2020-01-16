package cmd

import (
	"context"

	"github.com/gobuffalo/buffalo-auth/genny/auth"
	"github.com/gobuffalo/genny/v2"
	"github.com/gobuffalo/genny/v2/gogen"
	"github.com/spf13/cobra"
)

var dryRun bool

// authCmd generates a actions/auth.go file configured to the specified providers.
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Generates a full auth implementation",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := genny.WetRunner(context.Background())
		if dryRun {
			r = genny.DryRunner(context.Background())
		}

		if err := r.WithNew(auth.New(args)); err != nil {
			return err
		}

		if err := r.WithNew(gogen.Fmt(r.Root)); err != nil {
			return err
		}

		return r.Run()
	},
}

func init() {
	authCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "run the generator without creating files or running commands")
	RootCmd.AddCommand(authCmd)
}
