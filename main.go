package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rest-go/auth"

	"github.com/rest-go/rest/pkg/log"
	"github.com/rest-go/rest/pkg/server"
	"github.com/rest-go/rest/pkg/sql"
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
	if flag.Arg(0) == "auth" {
		authCmd(cfg.DB.URL)
		return
	}

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

func authCmd(url string) {
	if len(flag.Args()) == 1 {
		log.Fatal("rest auth setup/user/policy")
	}

	db, err := sql.Open(url)
	if err != nil {
		log.Fatal(err)
	}
	switch flag.Args()[1] {
	case "setup":
		username, password, err := auth.Setup(db)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("admin user created, username: %s, password: %s", username, password)
	case "user":
		if err := userCmd(db, flag.Args()[2:]); err != nil {
			log.Fatal(err)
		}
	case "policy":
		if err := policyCmd(db, flag.Args()[2:]); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("rest auth setup/user/policy")
	}
}

func userCmd(db *sql.DB, args []string) error {
	if len(args) == 0 {
		return errors.New("rest auth user list/add")
	}
	ctx, cancel := context.WithTimeout(context.Background(), sql.DefaultTimeout)
	defer cancel()

	switch args[0] {
	case "list":
		objects, err := db.FetchData(ctx, "SELECT id, username, is_admin FROM auth_users")
		if err != nil {
			return err
		}
		fmt.Println("id | username | is_admin ")
		for _, object := range objects {
			fmt.Printf("%d | %s | %t\n", object["id"], object["username"], object["is_admin"])
		}
	case "add":
		if len(args) < 4 { //nolint:gomnd
			return errors.New("rest auth user add <username> <password> <is_admin>")
		}
		username, password := args[1], args[2]
		hashedPasswd, err := auth.HashPassword(password)
		if err != nil {
			return err
		}
		isAdmin, _ := strconv.ParseBool(args[3])
		_, err = db.ExecQuery(ctx,
			"INSERT INTO auth_users (username, password, is_admin) VALUES (?,?,?)",
			username, hashedPasswd, isAdmin,
		)
		if err != nil {
			return err
		}
		fmt.Println("use is added into database")
	}
	return nil
}

func policyCmd(db *sql.DB, args []string) error {
	if len(args) == 0 {
		return errors.New("rest auth policy list/add")
	}
	ctx, cancel := context.WithTimeout(context.Background(), sql.DefaultTimeout)
	defer cancel()

	switch args[0] {
	case "list":
		objects, err := db.FetchData(ctx, "SELECT * FROM auth_policies")
		if err != nil {
			return err
		}
		fmt.Println("table_name | action | expression| description ")
		for _, object := range objects {
			fmt.Printf("%s | %s | %s | %s\n",
				object["table_name"], object["action"],
				object["expression"], object["description"],
			)
		}
	case "add":
		if len(args) < 5 { //nolint:gomnd
			return errors.New("rest auth policy add <table_name> <action> <expression> <description>")
		}
		_, err := db.ExecQuery(ctx,
			"INSERT INTO auth_policies (table_name, action, expression, description) VALUES (?,?,?,?)",
			args[1], args[2], args[3], args[4],
		)
		if err != nil {
			return err
		}
		fmt.Println("policy is added into database")
	}
	return nil
}
