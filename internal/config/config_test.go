package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GetAddr(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "default host and port",
			cfg: Config{
				Host: "0.0.0.0",
				Port: 8080,
			},
			want: "0.0.0.0:8080",
		},
		{
			name: "custom host and port",
			cfg: Config{
				Host: "127.0.0.1",
				Port: 9000,
			},
			want: "127.0.0.1:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.GetAddr()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoad(t *testing.T) {
	// Set required environment variables
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_URL", "postgres://test:pass@localhost:5432/testdb")

	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DATABASE_URL")
	}()

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-secret", cfg.JWTSecret)
	assert.Equal(t, "postgres://test:pass@localhost:5432/testdb", cfg.DatabaseURL)
}
