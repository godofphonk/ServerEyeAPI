// Copyright (c) 2026 godofphonk
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
