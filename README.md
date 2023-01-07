# Rest
Logo

![ci](https://github.com/shellfly/rest-go/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/github/shellfly/rest-go/branch/main/graph/badge.svg?token=DT5Q3DYNXP)](https://codecov.io/github/shellfly/rest-go)

Rest serves a fully RESTful API from any database, PostgreSQL, MySQL and SQLite are supported for now.

Visit https://rest-go.com for the full documentation, examples and guides.

# Guide
## Install

There are various ways of installing Rest.

### Precompiled binaries
Precompiled binaries for released versions are available in the [release page](). Using the latest production release binary is the recommended way of installing Rest. See the Installing chapter in the documentation for all the details.

### Go install

``` bash
go install github.com/shellfly/rest
```

### Run rest server
``` bash
# PG
rest -db.url "postgres://user:passwd@localhost:5432/db?search_path=api"

# MySQL
rest -db.url "mysql://user:passwd@tcp(localhost:3306)/db"

# SQLite
rest -db.url "sqlite://chinook.db"
```

## Use API

``` bash
# Create an artist
curl -XPOST "localhost:3000/artists" -d '{"artistid":10000, "name": "Bruce Lee"}'

# Read an artist
curl -XGET "localhost:3000/artists?&artistid=eq.10000"

# Update
curl -XPUT "localhost:3000/artists?&artistid=eq.10000" -d '{"name": "Stephen Chow"}'

# Delete
curl -XDELETE "localhost:3000/artists?&artistid=eq.10000"
```

### Docker image

``` bash
docker run --name rest -d -p 127.0.0.1:3000:3000 shellfly/rest -db.url "mysql://user:passwd@tcp(host:port)/db"
docker run --name rest -d -p 127.0.0.1:3000:3000 shellfly/rest -db.url "mysql://user:passwd@host:port/db"
```

### JSON

``` bash
# POST json
curl -XPOST "localhost:3000/people" -d '{"id":1, "json_data": {"blood_type":"A-", "phones":[{"country_code":61, "number":"919-929-5745"}]}}'

# Fetch json field
curl "http://localhost:3000/people?select=id,json_data->>blood_type,json_data->>phones"
```

# Use rest as a Go library
It also works to embed rest server into an existing Go http server

``` go
package main

import (
	"log"
	"net/http"

	"github.com/shellfly/rest/pkg/server"
)

func main() {
	s := server.NewServer("sqlite://chinook.db")
	http.Handle("/", s)
	// or with prefix
	// http.Handle("/admin", s.WithPrefix("/admin"))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

# Features
- [x] CRUD
- [x] Limit tables
- [x] Page
- [x] eq.,lt., gt., is., like.
- [x] ?select=f1,f2
- [x] order by
- [x] count
- [x] post many
- [x] debug output sql & args
- [x] common types(int, bool, char, timestamp, decimal)
- [x] test
- [ ] security sql
- [ ] auth(http & jwt)
- [ ] comment/documentation
- [x] json (postgres, operations, nested post/get)
  - [ ] quote
- [ ] json (mysql & sqlite)
- [ ] test for different db (github action)
# Road map
- [ ] Resource Embedding(one,many)
- [ ] Logical operators(or, and is already in code)
- [ ] escape field name
- [ ] application/x-www-form-urlencoded
- [ ] open api
- [ ] web management
