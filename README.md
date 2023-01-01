# Rest

Rest serves a fully RESTful API from any existing database, PostgreSQL, MySQL and SQLite are supported for now, other database that has [a driver in Golang](https://github.com/golang/go/wiki/SQLDrivers) could be added without much efforts in theory, and might be added in the future.

# Guide
## Install

``` bash
go install github.com/shellfly/rest
```

## Run rest server



``` bash
# PG
rest -db.url "postgres://user:passwd@localhost:5432/db?search_path=api"

# MySQL
rest -db.url "mysql://user:passwd@localhost:3306/db"

# SQLite
rest -db.url "sqlite://chinook.db"
```

## Use apis

``` bash
# Create an artist
curl -XPOST "localhost:3000/artists" -d '{"artistid":10000, "name": "Bruce Lee"}'

# Fetch one artist
curl -XGET "localhost:3000/artists?&artistid=eq.10000&singular"

# Fetch many artists
curl -XGET "localhost:3000/artists"

# Fetch count
curl -XGET "localhost:3000/artists?count"

# Update
curl -XPUT "localhost:3000/artists?&artistid=eq.10000" -d '{"name": "Stephen Chow"}'

# Delete
curl -XDELETE "localhost:3000/artists?&artistid=eq.10000"
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
- [ ] escape select
- [ ] post nested json
- [ ] json operations
- [ ] test
- [ ] log level

