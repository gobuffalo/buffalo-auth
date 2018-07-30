# Auth Generator for Buffalo

## Installation

```bash
$ go get -u github.com/gobuffalo/buffalo-auth
```

## Usage

To generate a basic username / password authentication you can run:

```bash
$ buffalo generate auth
```

This will do:

- Generate User authentication actions in `actions/auth.go`:
  - AuthNew
  - AuthCreate
  - AuthDestroy

- Generate User signup actions in `actions/users.go`:
  - UsersNew
  - UsersCreate

- Generate User model and migration ( model will be in `models/user.go`):

- Generate Auth Middlewares
  - SetCurrentUser
  - Authorize

- Add actions and middlewares in `app.go`:
  - [GET] /users/new -> UsersNew
  - [POST] /users -> UsersCreate
  - [GET] /signin -> AuthNew
  - [POST] /signin -> AuthCreate
  - [DELETE] /signout -> AuthDestroy

- Use middlewares for all your actions and skip
  - HomeHandler
  - UsersNew
  - UsersCreate
  - AuthNew
  - AuthCreate

### User model Fields

Sometimes you would want to add extra fields to the user model, to do so, you can pass those to the auth command and use the pop notation for those fields, for example:

```bash
$ buffalo auth first_name last_name phone_number notes:text
```

Will generate a User model (inside `models/user.go`) that looks like:

```go
type User struct {
  ID                   uuid.UUID `json:"id" db:"id"`
  CreatedAt            time.Time `json:"created_at" db:"created_at"`
  UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
  Email                string    `json:"email" db:"email"`
  PasswordHash         string    `json:"password_hash" db:"password_hash"`
  FirstName            string    `json:"first_name" db:"first_name"`
  LastName             string    `json:"last_name" db:"last_name"`
  Notes                string    `json:"notes" db:"notes"`
  Password             string    `json:"-" db:"-"`
  PasswordConfirmation string    `json:"-" db:"-"`
}
```