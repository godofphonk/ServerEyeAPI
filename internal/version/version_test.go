package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	assert.NotEmpty(t, version)
	assert.Contains(t, version, "1.0.0")
	assert.Contains(t, version, "commit:")
	assert.Contains(t, version, "built:")
	assert.Contains(t, version, "go:")
}

func TestVersionVariables(t *testing.T) {
	assert.Equal(t, "1.0.0", Version)
	assert.NotEmpty(t, GitCommit)
	assert.NotEmpty(t, BuildTime)
	assert.NotEmpty(t, GoVersion)
}
