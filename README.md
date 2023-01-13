# Rest
Logo

![ci](https://github.com/rest-go/rest/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/rest-go/rest/branch/main/graph/badge.svg?token=T38FWXMVY1)](https://codecov.io/gh/rest-go/rest)

Rest serves a fully RESTful API from any SQL database, PostgreSQL, MySQL, and SQLite are supported for now.

Visit https://rest-go.com for the full documentation, examples, and guides.

## Getting Started

### Start Rest in Docker
run the server and connect to an existing database
``` bash
# connect to postgres
docker run -p 3000:3000 restgo/rest -db.url "postgres://user:passwd@localhost:5432/db"

# connect to sqlite file with volume
docker run -p 3000:3000 -v $(pwd):/data restgo/rest -db.url "sqlite:///data/my.db"
```

### Use API
Assume there is a `todos` table in the database with `id`, and `title` fields:

``` bash
# Create a todo item
curl -XPOST "localhost:3000/todos" -d '{"title": "setup api server", "done": false}'

# Read
curl -XGET "localhost:3000/todos/1"

# Update
curl -XPUT "localhost:3000/todos/1" -d '{"title": "setup api server", "done": true}'

# Delete
curl -XDELETE "localhost:3000/todos/1"
```

## Use the binary

### Precompiled binaries
Precompiled binaries for released versions are available on the [Releases](https://github.com/rest-go/rest/releases) page](https://github.com/rest-go/rest/releases), download it to your local machine, and run it directly is the fastest way to use Rest.

### Go install
If you are familiar with Golang, you can use go install
``` bash
go install github.com/rest-go/rest
```

### Run server
``` bash
rest -db.url "mysql://username:password@tcp(localhost:3306)/db"
```

## Use it as a Go library
It also works to embed the rest server into an existing Go HTTP server

``` bash
go get github.com/rest-go/rest
```

``` go
package main

import (
	"log"
	"net/http"

	"github.com/rest-go/rest/pkg/server"
)

func main() {
	s := server.NewServer("sqlite://my.db")
	http.Handle("/", s)
	// or with prefix
	// http.Handle("/admin", s.WithPrefix("/admin"))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Features
- [x] CRUD
- [x] Page
- [x] operators (eq, lt, gt, etc.)
- [x] select fields
- [x] order by
- [x] count
- [x] json 
## Road map
- [ ] Authentication and Authorization
- [ ] Resource Embedding(one-to-one, one-to-many)
- [ ] Logical operators(and, or)
- [ ] Open API
- [ ] Web UI management
