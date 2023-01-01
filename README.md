# Rest

Turn any database to a restful interface

# Install

``` bash
go install github.com/shellfly/rest
```

# Run Rest server

1. Postgres
``` bash
rest -db.url "postgres://user:passwd@localhost:5432/db?search_path=api"
```

2. MySQL
``` bash
rest -db.url "mysql://user:passwd@localhost:3306/db"
```

3. SQLite
``` bash
rest -db.url "sqlite://chinook.db"
```

## Use RESTful apis

1. Create
``` bash
curl -XPOST "localhost:3000/artists" -d '{"artistid":10000, "name": "Bruce Lee"}'
{"code":200,"msg":"success"}
```

2. Read
``` bash
curl "localhost:3000/artists?&artistid=eq.10000"
{"code":200,"msg":"success","data":[{"ArtistId":10000,"Name":"Bruce Lee"}]}
```

3. Update
``` bash
curl -XPUT "localhost:3000/artists?&artistid=eq.10000" -d '{"name": "Stephen Chow"}'

{"code":200,"msg":"successfully updated 1 rows"}
```

4. Delete
``` bash
curl -XDELETE "localhost:3000/artists?&artistid=eq.10000"
{"code":200,"msg":"successfully deleted 1 rows"}
```

# Road map
- [x] CRUD
- [x] Limit tables
- [x] open db by url config
- [x] Page
- [x] eq.4, lt., gt.,gte. is.
- [x] ?select=f1,f2
- [x] order by
- [ ] post many
- [ ] escape select
- [ ] post nested json
- [ ] json operations
- [ ] test
- [ ] log level

