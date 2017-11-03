package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/gobuffalo/buffalo-auth/auth"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/makr"
	"github.com/spf13/cobra"
)

// authCmd generates a actions/auth.go file configured to the specified providers.
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Generates a full auth implementation",
	RunE: func(cmd *cobra.Command, args []string) error {
		pwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}

		g, err := auth.New()
		if err != nil {
			return err
		}
		return g.Run(".", makr.Data{
			"packagePath": packagePath(pwd),
		})
	},
}

func goPath(root string) string {
	gpMultiple := envy.GoPaths()
	path := ""

	for i := 0; i < len(gpMultiple); i++ {
		if strings.HasPrefix(root, filepath.Join(gpMultiple[i], "src")) {
			path = gpMultiple[i]
			break
		}
	}
	return path
}

func packagePath(rootPath string) string {
	gosrcpath := strings.Replace(filepath.Join(goPath(rootPath), "src"), "\\", "/", -1)
	rootPath = strings.Replace(rootPath, "\\", "/", -1)
	return strings.Replace(rootPath, gosrcpath+"/", "", 2)
}

func init() {
	RootCmd.AddCommand(authCmd)
}
