package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/rest-go/auth"
	"github.com/rest-go/rest/pkg/log"
	"github.com/rest-go/rest/pkg/server"
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

func main() {
	cfg := parseFlags()
	restSrv := server.New(&cfg.DB, server.EnableAuth(cfg.Auth.Enabled))
	if cfg.Auth.Enabled {
		log.Info("auth is enabled")
		restAuth, err := auth.New(cfg.DB.URL, []byte(cfg.Auth.Secret))
		if err != nil {
			log.Fatal("initialize auth error ", err)
		}
		http.Handle("/auth/", restAuth)
		http.Handle("/", restAuth.Middleware(restSrv))
	} else {
		http.Handle("/", restSrv)
	}

	log.Info("listen on addr: ", cfg.Addr)
	server := &http.Server{
		Addr:              cfg.Addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
