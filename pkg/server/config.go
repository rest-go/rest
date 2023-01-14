package server

import "fmt"

type Config struct {
	Prefix string
	DB     DBConfig
	Auth   AuthConfig
}

func (c Config) String() string {
	return fmt.Sprintf("prefix: %s, db: %s, auth: %s", c.Prefix, c.DB, c.Auth)
}

type DBConfig struct {
	URL string
}

func (c DBConfig) String() string {
	return fmt.Sprintf("{url: %s}", c.URL)
}

type AuthConfig struct {
	Enabled bool
	Secret  string
}

func (c AuthConfig) String() string {
	return fmt.Sprintf("{enabled: %v, secret:xxx}", c.Enabled)
}
