<p align="center"><img src="https://github.com/gobuffalo/buffalo/blob/master/logo.svg" width="360"></p>

# Auth Generator for Buffalo

[![Tests](https://github.com/gobuffalo/buffalo-auth/actions/workflows/tests.yml/badge.svg)](https://github.com/gobuffalo/buffalo-auth/actions/workflows/tests.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/gobuffalo/buffalo-auth.svg)](https://pkg.go.dev/github.com/gobuffalo/buffalo-auth)
[![Go Report Card](https://goreportcard.com/badge/github.com/gobuffalo/buffalo-auth)](https://goreportcard.com/report/github.com/gobuffalo/buffalo-auth)

## Installation

```console
$ buffalo plugins install github.com/gobuffalo/buffalo-auth
```

## Usage

To generate a basic username / password authentication you can run:

```console
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

```console
$ buffalo generate auth first_name last_name notes:text
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
