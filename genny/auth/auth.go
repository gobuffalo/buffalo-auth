package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo/meta"
	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/movinglater/attrs"
	"github.com/gobuffalo/genny/movinglater/gotools"
	"github.com/gobuffalo/genny/movinglater/plushgen"
	"github.com/gobuffalo/packr"
	"github.com/gobuffalo/plush"
	"github.com/pkg/errors"
)

func extraAttrs(args []string) []string {
	var names = map[string]string{
		"email":    "email",
		"password": "password",
		"id":       "id",
	}

	var result = []string{}
	for _, field := range args {
		at, _ := attrs.Parse(field)
		field = at.Name.Underscore().String()

		if names[field] != "" {
			continue
		}

		names[field] = field
		result = append(result, field)
	}

	return result
}

var fields attrs.Attrs

// New actions/auth.go file configured to the specified providers.
func New(args []string) (*genny.Generator, error) {
	g := genny.New()

	var err error
	fields, err = attrs.ParseArgs(extraAttrs(args)...)
	if err != nil {
		return g, errors.WithStack(err)
	}

	if err := g.Box(packr.NewBox(filepath.Join("..", "auth", "templates"))); err != nil {
		return g, errors.WithStack(err)
	}

	ctx := plush.NewContext()
	ctx.Set("app", meta.New("."))
	ctx.Set("attrs", fields)

	g.Transformer(plushgen.Transformer(ctx))
	g.Transformer(genny.NewTransformer(".html", newUserHTMLTransformer))
	g.Transformer(genny.NewTransformer(".fizz", migrationsTransformer))

	g.RunFn(func(r *genny.Runner) error {
		path := filepath.Join("actions", "app.go")
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		gf := genny.NewFile(path, file)

		gf, err = gotools.AddInsideBlock(
			gf,
			`if app == nil {`,
			`app.Use(SetCurrentUser)`,
			`app.Use(Authorize)`,
			`app.GET("/users/new", UsersNew)`,
			`app.POST("/users", UsersCreate)`,
			`app.GET("/signin", AuthNew)`,
			`app.POST("/signin", AuthCreate)`,
			`app.DELETE("/signout", AuthDestroy)`,
			`app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthNew, AuthCreate)`,
		)

		return r.File(gf)
	})

	return g, nil
}

func newUserHTMLTransformer(f genny.File) (genny.File, error) {
	if f.Name() != filepath.Join("templates", "users", "new.html") {
		return f, nil
	}

	fieldInputs := []string{}
	for _, field := range fields {
		name := field.Name.Proper()
		fieldInputs = append(fieldInputs, fmt.Sprintf(`<%%= f.InputTag("%v", {}) %%>`, name))
	}

	lines := strings.Split(f.String(), "\n")
	ln := -1

	for index, line := range lines {
		if strings.Contains(line, `<%= f.InputTag("PasswordConfirmation"`) {
			ln = index + 1
			break
		}
	}

	lines = append(lines[:ln], append(fieldInputs, lines[ln:]...)...)
	b := strings.NewReader(strings.Join(lines, "\n"))
	return genny.NewFile(f.Name(), b), nil
}

func migrationsTransformer(f genny.File) (genny.File, error) {
	t := time.Now()
	parts := strings.Split(f.Name(), string(filepath.Separator))
	p := parts[len(parts)-1]
	return genny.NewFile(filepath.Join("migrations", fmt.Sprintf("%d_%s", t.UnixNano(), p)), f), nil
}
