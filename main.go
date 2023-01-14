package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/rest-go/auth"
	"github.com/rest-go/rest/pkg/server"
	"github.com/rest-go/rest/pkg/sqlx"
)

func parseFlags() *Config {
	addr := flag.String("addr", ":3000", "listen addr")
	url := flag.String("db.url", "", "db url")
	cfgPath := flag.String("config", "", "path to config file")

	// Actually parse the flags
	flag.Parse()
	cfg := &Config{}
	if *cfgPath != "" {
		var err error
		cfg, err = NewConfig(*cfgPath)
		if err != nil {
			log.Fatal(err)
		}
	}
	if cfg.Addr == "" {
		cfg.Addr = *addr
	}
	if *url != "" {
		cfg.DB.URL = *url
	}
	return cfg
}

func setupAuth() {
	authFlags := flag.NewFlagSet("foo", flag.ExitOnError)
	url := authFlags.String("db.url", "", "db url")
	if err := authFlags.Parse(os.Args[3:]); err != nil {
		log.Fatal("failed to parse flags", err)
	}
	if *url == "" {
		log.Fatal("db url is required to setup auth tables")
	}
	db, err := sqlx.Open(*url)
	if err != nil {
		log.Fatal("can't open db url, ", *url, err)
	}
	if err = auth.New(db, "").Setup(); err != nil {
		log.Fatal("setup auth error ", err)
	}
}

func main() {
	if len(os.Args) >= 3 {
		if os.Args[1] == "auth" && os.Args[2] == "setup" {
			setupAuth()
			return
		}
	}
	cfg := parseFlags()
	s := server.NewServer(&server.Config{DB: cfg.DB, Auth: cfg.Auth})
	log.Print("listen on addr: ", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, s)) //nolint:gosec // not handled for now
}
