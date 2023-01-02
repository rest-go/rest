package main

import (
	"flag"
	"log"
	"net/http"
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
		cfg.DB.Url = *url
	}
	return cfg
}

func main() {
	cfg := parseFlags()

	s := NewService(cfg.DB.Url, cfg.DB.Tables...)
	log.Print("listen on addr: ", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, s))
}
