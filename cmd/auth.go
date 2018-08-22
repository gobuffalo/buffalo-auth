package cmd

import (
	"context"

	"github.com/gobuffalo/buffalo-auth/genny/auth"
	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/movinglater/gotools"
	"github.com/pkg/errors"
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

		g, err := auth.New(args)
		if err != nil {
			return err
		}

		r.With(g)

		g, err = gotools.GoFmt(r.Root)
		if err != nil {
			return errors.WithStack(err)
		}
		r.With(g)

		return r.Run()
	},
}

func init() {
	authCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "run the generator without creating files or running commands")
	RootCmd.AddCommand(authCmd)
}
