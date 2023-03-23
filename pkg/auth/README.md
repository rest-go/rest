# Auth package


Auth is a RESTFul Authentication and Authorization package for Golang HTTP apps.

It handles the common tasks of registration, logging in, logging out, JWT token generation, and JWT token verification.

## Usage
import `auth` to your app, create `auth.Handler` and `auth.Middleware` based on requirements.
```go
package main

import (
	"log"
	"net/http"

	"github.com/rest-go/rest/pkg/auth"
)

func handle(w http.ResponseWriter, req *http.Request) {
	user := auth.GetUser(req)
	if user.IsAnonymous() {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	dbURL := "sqlite://my.db"
	jwtSecret := "my secret"
	authHandler, err := auth.NewHandler(dbURL, []byte(jwtSecret))
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/auth/", authHandler)

	middleware := auth.NewMiddleware([]byte(jwtSecret))
	http.Handle("/", middleware(http.HandlerFunc(handle)))
	log.Fatal(http.ListenAndServe(":8000", nil)) //nolint:gosec
}
```

## Setup database

Send a `POST` request to `/auth/setup` to set up database tables for users. This
will also create an admin user account and return the username and password in
the response.

```bash
$ curl -XPOST "localhost:8000/auth/setup"
```

## Auth handler

The `Auth` struct implements the `http.Hanlder` interface and provides the below endpoints for user management.

1. Register

```bash
$ curl  -XPOST "localhost:8000/auth/register" -d '{"username":"hello", "password": "world"}'
```

2. Login

```bash
$ curl  -XPOST "localhost:8000/auth/login" -d '{"username":"hello", "password": "world"}'
```

3. Logout

Currently, the authentication mechanism is based on JWT token only, logout is a no-op on the
server side, and the client should clear the token by itself.

```bash
$ curl  -XPOST "localhost:8000/auth/logout"
```

## Auth middleware and `GetUser`

Auth middleware will parse JWT token in the HTTP header, and when successful,
set the user in the request context, the `GetUser` method can be used to get the
user from the request.

``` go
user := auth.GetUser(req)
```

