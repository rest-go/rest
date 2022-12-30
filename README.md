# Rest

Turn any database to a restful interface
# Example

## Install

``` bash
go get github.com/shellfly/rest-go
```

## Run Rest server

``` bash
rest -addr :3000 -db.url sqlite://chibook.db
```

## Access restful apis

``` bash
curl "localhost:3000/artists?&artistid=in.(99,100)"
```

# Road map
- [x] CRUD
- [x] Limit tables
- [x] open db by url config
- [x] Page
- [x] eq.4, lt., gt.,gte. is.
- [x] ?select=f1,f2
- [x] order by
- [ ] json operations
- [ ] test
- [ ] log level

