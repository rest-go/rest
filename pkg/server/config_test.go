package server

import "testing"

func TestConfig(t *testing.T) {
	config := Config{
		DB:   DBConfig{URL: "sqlite://my.db"},
		Auth: AuthConfig{Enabled: true},
	}
	t.Log(config.String())
}
