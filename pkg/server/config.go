package server

import "fmt"

type Config struct {
	DB   DBConfig
	Auth AuthConfig
	Cors CorsConfig
}

func (c Config) String() string {
	return fmt.Sprintf("db: %s, auth: %s, cors: %s", c.DB, c.Auth, c.Cors)
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

type CorsConfig struct {
	Enabled bool
	Origins []string
}

func (c CorsConfig) String() string {
	return fmt.Sprintf("{enabled: %v, origins: %v}", c.Enabled, c.Origins)
}
