package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/rest-go/auth"
	"github.com/rest-go/rest/pkg/log"
	"github.com/rest-go/rest/pkg/server"
)

const defaultAddr = ":3000"

func parseConfig() *Config {
	addr := flag.String("addr", "", "listen addr")
	url := flag.String("db.url", "", "database url")
	cfgPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg := &Config{}
	if *cfgPath != "" {
		var err error
		cfg, err = NewConfig(*cfgPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *addr != "" {
		cfg.Addr = *addr
	} else if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}
	if *url != "" {
		cfg.DB.URL = *url
	}
	return cfg
}

func main() {
	cfg := parseConfig()
	restServer := server.New(&cfg.DB, server.EnableAuth(cfg.Auth.Enabled))
	if cfg.Auth.Enabled {
		log.Info("auth is enabled")
		authHandler, err := auth.NewHandler(cfg.DB.URL, []byte(cfg.Auth.Secret))
		if err != nil {
			log.Fatal("initialize auth error ", err)
		}
		http.Handle("/auth/", authHandler)

		middleware := auth.NewMiddleware([]byte(cfg.Auth.Secret))
		http.Handle("/", middleware(restServer))
	} else {
		http.Handle("/", restServer)
	}

	s := &http.Server{
		Addr:              cfg.Addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Info("listen on addr: ", cfg.Addr)
	log.Fatal(s.ListenAndServe())
}
