package auth

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobuffalo/attrs"
	"github.com/gobuffalo/genny/v2"
	"github.com/gobuffalo/genny/v2/gogen"
	"github.com/gobuffalo/genny/v2/plushgen"
	"github.com/gobuffalo/meta"
	"github.com/gobuffalo/plush/v4"
)

//go:embed templates
var templates embed.FS

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
		return g, fmt.Errorf("could not parse arguments: %w", err)
	}

	sub, err := fs.Sub(templates, "templates")
	if err != nil {
		return g, fmt.Errorf("failed to get subtree of templates: %w", err)
	}
	if err := g.FS(sub); err != nil {
		return g, fmt.Errorf("failed to add subtree: %w", err)
	}

	ctx := plush.NewContext()
	ctx.Set("app", meta.New("."))
	ctx.Set("attrs", fields)

	g.Transformer(plushgen.Transformer(ctx))
	g.Transformer(genny.NewTransformer(".html", newUserHTMLTransformer))
	g.Transformer(genny.Replace(".html", ".plush.html"))
	g.Transformer(genny.NewTransformer(".fizz", migrationsTransformer(time.Now())))

	g.RunFn(func(r *genny.Runner) error {
		path := filepath.Join("actions", "app.go")

		f, err := r.FindFile(path)
		if err != nil {
			return fmt.Errorf("setup auth: %w", err)
		}

		expressions := []string{
			``,
			`// NOTE: this block should go before any resources`,
			`// that need to be protected by buffalo-goth!`,
			`//AuthMiddlewares`,
			`app.Use(SetCurrentUser)`,
			`app.Use(Authorize)`,
			``,
			`//Routes for Auth`,
			`auth := app.Group("/auth")`,
			`auth.GET("/", AuthLanding)`,
			`auth.GET("/new", AuthNew)`,
			`auth.POST("/", AuthCreate)`,
			`auth.DELETE("/", AuthDestroy)`,
			`auth.Middleware.Skip(Authorize, AuthLanding, AuthNew, AuthCreate)`,
			``,
			`//Routes for User registration`,
			`users := app.Group("/users")`,
			`users.GET("/new", UsersNew)`,
			`users.POST("/", UsersCreate)`,
			`users.Middleware.Remove(Authorize)`,
			``,
		}

		f, err = gogen.AddInsideBlock(f, "appOnce.Do(func() {", expressions...)
		if err != nil {
			if strings.Contains(err.Error(), "could not find desired block") {
				// TODO: remove this block some day soon
				// add this block for compatibility with the apps built with
				// the old version of Buffalo CLI (v0.18.8 or older)
				f, err = gogen.AddInsideBlock(f, "if app == nil {", expressions...)
				if err != nil {
					if err != nil {
						return fmt.Errorf("could not add a code block: %w", err)
					} else {
						r.Logger.Warnf("This app was built with CLI v0.18.8 or older. See https://gobuffalo.io/documentation/known-issues/#cli-v0.18.8")
					}
				}
			} else {
				return fmt.Errorf("could not add a code block: %w", err)
			}
		}

		return r.File(f)
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
		fieldInputs = append(fieldInputs, fmt.Sprintf(`      <%%= f.InputTag("%v", {}) %%>`, name))
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

func migrationsTransformer(t time.Time) genny.TransformerFn {
	v := t.UTC().Format("20060102150405")
	return func(f genny.File) (genny.File, error) {
		p := filepath.Base(f.Name())
		return genny.NewFile(filepath.Join("migrations", fmt.Sprintf("%s_%s", v, p)), f), nil
	}
}
