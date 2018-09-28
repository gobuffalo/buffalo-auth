package auth

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/buffalo/meta"
	"github.com/gobuffalo/flect"
	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/movinglater/gotools"
	"github.com/gobuffalo/genny/movinglater/plushgen"
	"github.com/gobuffalo/packr"
	"github.com/gobuffalo/plush"
	"github.com/pkg/errors"
)

// New actions/auth.go file configured to the specified providers.
func New(args []string) (*genny.Generator, error) {
	g := genny.New()

	fields, extraFields := extractFields(args)

	parts := append([]string{"db", "generate", "model"}, fields...)
	g.Command(exec.Command("buffalo", parts...))

	if err := g.Box(packr.NewBox("./templates")); err != nil {
		return g, errors.WithStack(err)
	}

	ctx := plush.NewContext()
	ctx.Set("app", meta.New("."))

	g.Transformer(plushgen.Transformer(ctx))
	g.RunFn(modifyUsersModel)
	g.RunFn(addPasswordFields)
	g.RunFn(addAppActions)
	g.RunFn(formFieldsFn(extraFields))

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

		path := filepath.Join("templates", "users", "new.html")
		f, err := os.Open(path)
		if err != nil {
			return err
		}

		gf := genny.NewFile(path, f)
		lines := strings.Split(gf.String(), "\n")
		ln := -1

		for index, line := range lines {
			if strings.Contains(line, `<%= f.InputTag("PasswordConfirmation", {type: "password"}) %>`) {
				ln = index
				break
			}
		}

		lines = append(lines[:ln], append(fieldInputs, lines[ln:]...)...)
		gf = genny.NewFile(path, strings.NewReader(strings.Join(lines, "\n")))
		return r.File(gf)
	}
}

func addAppActions(r *genny.Runner) error {
	f, err := os.Open(`actions/app.go`)
	if err != nil {
		return err
	}

	file := genny.NewFile(`actions/app.go`, f)
	file, err = gotools.AddInsideBlock(
		file,
		"if app == nil {",
		`app.Use(SetCurrentUser)`,
		`app.Use(Authorize)`,
		`app.GET("/users/new", UsersNew)`,
		`app.POST("/users", UsersCreate)`,
		`app.GET("/signin", AuthNew)`,
		`app.POST("/signin", AuthCreate)`,
		`app.DELETE("/signout", AuthDestroy)`,
		`app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthNew, AuthCreate)`,
	)

	if err != nil {
		return err
	}

	return r.File(file)
}

func addPasswordFields(r *genny.Runner) error {
	f, err := os.Open(`models/user.go`)
	if err != nil {
		return err
	}
	file := genny.NewFile(`models/user.go`, f)
	file, err = gotools.AddInsideBlock(
		file,
		"type User struct {",
		"Password string `json:\"-\" db:\"-\"`",
		"PasswordConfirmation string `json:\"-\" db:\"-\"`",
	)

	if err != nil {
		return err
	}

	return r.File(file)
}

func modifyUsersModel(r *genny.Runner) error {
	f, err := os.Open(`models/user.go`)
	if err != nil {
		return err
	}

	file := genny.NewFile(`models/user.go`, f)
	file = gotools.ReplaceBlockBody(file, `func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {`, validateCreateFuncBodyLiteral)
	file = gotools.ReplaceBlockBody(file, "func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {", validateFuncBodyLiteral)
	file = gotools.Append(createFuncLiteral)
	file, err = gotools.AddImport(file, "strings", "github.com/pkg/errors", "golang.org/x/crypto/bcrypt")

	return r.File(file)
}

const (
	createFuncLiteral = `
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
	}`

	validateFuncBodyLiteral = `
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
		), err`

	validateCreateFuncBodyLiteral = `
		var err error
		return validate.Validate(
			&validators.StringIsPresent{Field: u.Password, Name: "Password"},
			&validators.StringsMatch{Name: "Password", Field: u.Password, Field2: u.PasswordConfirmation, Message: "Password does not match confirmation"},
		), err`
)
