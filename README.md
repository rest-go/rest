# Rest
Logo

![ci](https://github.com/shellfly/rest-go/actions/workflows/ci.yml/badge.svg)

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
rest -db.url "mysql://user:passwd@localhost:3306/db"

# SQLite
rest -db.url "sqlite://chinook.db"
```


### Docker image

``` bash
docker run --name rest -d -p 127.0.0.1:3000:3000 shellfly/rest -db.url "mysql://user:passwd@host:port/db"
```

## Use RESTFul API

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

### JSON

``` bash
# POST json
curl -XPOST "localhost:3000/people" -d '{"id":1, "json_data": {"blood_type":"A-", "phones":[{"country_code":61, "number":"919-929-5745"}]}}'

# Fetch json field
curl "http://localhost:3000/people?select=id,json_data->>blood_type,json_data->>phones"
```

# Using Rest as a Go Library

``` go
package main

func main(){
    
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
- [ ] common types(int, char, timestamp, decimal)
- [ ] test
- [ ] log level
- [ ] security
- [ ] comment/documentation
- [x] json (postgres, operations, nested post/get)
- [ ] json (mysql & sqlite)
- [ ] Resource Embedding(one,many)
- [ ] Logical operators(or, and is already in code)
- [ ] escape field name
- [ ] application/x-www-form-urlencoded
- [ ] open api
- [ ] web management
