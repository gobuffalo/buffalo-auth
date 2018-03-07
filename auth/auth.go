package auth

import (
	"os/exec"
	"path/filepath"

	"github.com/gobuffalo/buffalo/generators"
	"github.com/gobuffalo/makr"
)

// New actions/auth.go file configured to the specified providers.
func New() (*makr.Generator, error) {
	g := makr.New()
	files, err := generators.FindByBox(filepath.Join("github.com", "gobuffalo", "buffalo-auth", "auth"))
	if err != nil {
		return nil, err
	}

	g.Add(makr.NewCommand(exec.Command("buffalo", "db", "generate", "model", "user", "email", "password_hash")))

	for _, f := range files {
		g.Add(makr.NewFile(f.WritePath, f.Body))
	}

	g.Add(&makr.Func{
		Should: func(data makr.Data) bool { return true },
		Runner: func(root string, data makr.Data) error {
			return generators.AddInsideAppBlock(
				`app.Use(SetCurrentUser)`,
				`app.Use(Authorize)`,
				`app.GET("/users/new", UsersNew)`,
				`app.POST("/users", UsersCreate)`,
				`app.GET("/signin", AuthNew)`,
				`app.POST("/signin", AuthCreate)`,
				`app.DELETE("/signout", AuthDestroy)`,
				`app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthNew, AuthCreate)`,
			)
		},
	})
	g.Add(makr.NewCommand(makr.GoGet("github.com/markbates/goth/...")))
	g.Fmt(".")
	return g, nil
}
