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

package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateServerID(t *testing.T) {
	id1 := GenerateServerID()
	id2 := GenerateServerID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.True(t, len(id1) >= 11) // srv_ + 8 hex chars
	assert.True(t, len(id2) >= 11)
	assert.True(t, strings.HasPrefix(id1, "srv_"))
	assert.True(t, strings.HasPrefix(id2, "srv_"))
}

func TestGenerateServerKey(t *testing.T) {
	key1 := GenerateServerKey()
	key2 := GenerateServerKey()

	assert.NotEmpty(t, key1)
	assert.NotEmpty(t, key2)
	assert.NotEqual(t, key1, key2)
	assert.True(t, len(key1) >= 20) // key_ + 16 hex chars
	assert.True(t, len(key2) >= 20)
	assert.True(t, strings.HasPrefix(key1, "key_"))
	assert.True(t, strings.HasPrefix(key2, "key_"))
}

func TestGenerateSecretKey(t *testing.T) {
	key1 := GenerateSecretKey()
	key2 := GenerateSecretKey()

	assert.NotEmpty(t, key1)
	assert.NotEmpty(t, key2)
	assert.NotEqual(t, key1, key2)
	assert.True(t, len(key1) >= 23) // secret_ + 16 hex chars
	assert.True(t, len(key2) >= 23)
	assert.True(t, strings.HasPrefix(key1, "secret_"))
	assert.True(t, strings.HasPrefix(key2, "secret_"))
}

func TestValidateSecretKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid key",
			key:     "valid_secret_key_123",
			wantErr: false,
		},
		{
			name:    "too short",
			key:     "short",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			key:     "invalid@key",
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateServerID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid id",
			id:      "srv_12345678",
			wantErr: false,
		},
		{
			name:    "too short",
			id:      "srv_",
			wantErr: true,
		},
		{
			name:    "missing prefix",
			id:      "12345678",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			id:      "srv_123@456",
			wantErr: true,
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateServerKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid key",
			key:     "key_1234567890abcdef",
			wantErr: false,
		},
		{
			name:    "too short",
			key:     "key_",
			wantErr: true,
		},
		{
			name:    "missing prefix",
			key:     "1234567890abcdef",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			key:     "key_123@456",
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
