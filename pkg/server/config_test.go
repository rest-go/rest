package server

import "testing"

func TestConfig(t *testing.T) {
	config := Config{
		DB:   DBConfig{URL: "sqlite://my.db"},
		Auth: AuthConfig{Enabled: true},
		Cors: CorsConfig{Enabled: true, Origins: []string{"example1.com", "example2.com"}},
	}
	t.Log(config.String())
}
