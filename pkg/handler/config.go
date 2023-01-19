package handler

import "fmt"

type Config struct {
	DB   DBConfig
	Auth AuthConfig
}

func (c Config) String() string {
	return fmt.Sprintf("db: %s, auth: %s", c.DB, c.Auth)
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
