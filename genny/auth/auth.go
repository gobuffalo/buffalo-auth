package auth

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/gobuffalo/buffalo-auth/transformer"
	"github.com/gobuffalo/buffalo/generators"
	"github.com/gobuffalo/buffalo/meta"
	"github.com/gobuffalo/flect"
	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/movinglater/plushgen"
	"github.com/gobuffalo/packr"
	"github.com/gobuffalo/plush"
	"github.com/pkg/errors"
)

// New actions/auth.go file configured to the specified providers.
func New(args []string) (*genny.Generator, error) {
	g := genny.New()

	fields, extraFields := extractFields(args)

	commandParts := append([]string{"db", "generate", "model"}, fields...)
	g.Command(exec.Command("buffalo", commandParts...))

	if err := g.Box(packr.NewBox("./templates")); err != nil {
		return g, errors.WithStack(err)
	}

	ctx := plush.NewContext()
	ctx.Set("app", meta.New("."))

	g.Transformer(plushgen.Transformer(ctx))
	g.RunFn(formFieldsFn(extraFields))
	g.RunFn(addAppActions)
	g.RunFn(addPasswordFields)
	g.RunFn(modifyUsersModel)

	return g, nil
}

func extractFields(args []string) ([]string, []string) {
	fields := []string{"user", "email", "password_hash"}
	extraFields := []string{}
	for _, field := range args {
		fieldName := strings.Split(field, ":")[0]
		if strings.Contains(strings.Join(fields, "\n"), fieldName) {
			continue
		}

		fields = append(fields, field)
		extraFields = append(extraFields, field)
	}

	return fields, extraFields
}

func formFieldsFn(extraFields []string) genny.RunFn {
	return func(r *genny.Runner) error {
		fieldInputs := []string{}
		for _, field := range extraFields {
			name := flect.Capitalize(flect.Camelize(field))
			fieldInputs = append(fieldInputs, fmt.Sprintf(`<%%= f.InputTag("%v", {}) %%>`, name))
		}

		tr := transformer.NewTransformer("templates/users/new.html")
		tr.AppendAfter(`<%= f.InputTag("PasswordConfirmation", {type: "password"}) %>`, fieldInputs...)

		return nil
	}
}

func addAppActions(r *genny.Runner) error {
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
}

func addPasswordFields(r *genny.Runner) error {
	tr := transformer.NewTransformer("models/user.go")
	tr.AppendToBlock("type User struct {", []string{
		"Password string `json:\"-\" db:\"-\"`",
		"PasswordConfirmation string `json:\"-\" db:\"-\"`",
	}...)

	return nil
}

func modifyUsersModel(r *genny.Runner) error {
	tr := transformer.NewTransformer("models/user.go")
	tr.SetBlockBody("func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {", `
			var err error
			return validate.Validate(
				&validators.StringIsPresent{Field: u.Password, Name: "Password"},
				&validators.StringsMatch{Name: "Password", Field: u.Password, Field2: u.PasswordConfirmation, Message: "Password does not match confirmation"},
			), err
		`)

	tr.SetBlockBody("func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {", `
			var err error
			return validate.Validate(
				&validators.StringIsPresent{Field: u.Email, Name: "Email"},
				&validators.StringIsPresent{Field: u.PasswordHash, Name: "PasswordHash"},
				// check to see if the email address is already taken:
				&validators.FuncValidator{
					Field:   u.Email,
					Name:    "Email",
					Message: "%s is already taken",
					Fn: func() bool {
						var b bool
						q := tx.Where("email = ?", u.Email)
						if u.ID != uuid.Nil {
							q = q.Where("id != ?", u.ID)
						}
						b, err = q.Exists(u)
						if err != nil {
							return false
						}
						return !b
					},
				},
			), err
		`)

	tr.Append(`
			// Create wraps up the pattern of encrypting the password and
			// running validations. Useful when writing tests.
			func (u *User) Create(tx *pop.Connection) (*validate.Errors, error) {
				u.Email = strings.ToLower(strings.TrimSpace(u.Email))
				ph, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
				if err != nil {
					return validate.NewErrors(), errors.WithStack(err)
				}
				u.PasswordHash = string(ph)
				return tx.ValidateAndCreate(u)
			}

		`)

	tr.AddImports("\"strings\"", "\"github.com/pkg/errors\"", "\"golang.org/x/crypto/bcrypt\"")
	return nil
}
