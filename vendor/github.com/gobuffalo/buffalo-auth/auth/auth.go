package auth

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/gobuffalo/buffalo-auth/transformer"
	"github.com/gobuffalo/buffalo/generators"
	"github.com/gobuffalo/makr"
	"github.com/gobuffalo/packr"
	"github.com/markbates/inflect"
)

// New actions/auth.go file configured to the specified providers.
func New(args []string) (*makr.Generator, error) {
	g := makr.New()
	files, err := generators.FindByBox(packr.NewBox("../auth/templates"))
	if err != nil {
		return nil, err
	}

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

	commandParts := append([]string{"db", "generate", "model"}, fields...)
	g.Add(makr.NewCommand(exec.Command("buffalo", commandParts...)))

	for _, f := range files {
		g.Add(makr.NewFile(f.WritePath, f.Body))
	}

	g.Add(&makr.Func{
		Should: func(data makr.Data) bool { return true },
		Runner: func(root string, data makr.Data) error {

			fieldInputs := []string{}
			for _, field := range extraFields {
				name := inflect.Camelize(field)
				fieldInputs = append(fieldInputs, fmt.Sprintf(`<%%= f.InputTag("%v", {}) %%>`, name))
			}

			tr := transformer.NewTransformer("templates/users/new.html")
			tr.AppendAfter(`<%= f.InputTag("PasswordConfirmation", {type: "password"}) %>`, fieldInputs...)

			return nil
		},
	})

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

	g.Add(&makr.Func{
		Should: func(data makr.Data) bool { return true },
		Runner: func(root string, data makr.Data) error {
			tr := transformer.NewTransformer("models/user.go")
			tr.AppendToBlock("type User struct {", []string{
				"Password string `json:\"-\" db:\"-\"`",
				"PasswordConfirmation string `json:\"-\" db:\"-\"`",
			}...)

			return nil
		},
	})

	g.Add(&makr.Func{
		Should: func(data makr.Data) bool { return true },
		Runner: func(root string, data makr.Data) error {
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
		},
	})

	g.Fmt(".")
	return g, nil
}
