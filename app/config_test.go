package app

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewAppConfig(t *testing.T) {
	cfg := NewAppConfig("json", "testconfig", []string{"/test/path1", "/test/path2"})

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Opts)
}

func TestAppConfig_PopulateFromFlagset(t *testing.T) {
	cfg := NewAppConfig("json", "testconfig", []string{"."})

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Int("port", 8080, "Port")
	flags.String("dbuser", "", "Database user")

	renames := map[string]string{
		"serve.port":       "port",
		"mongodb.username": "dbuser",
	}

	err := cfg.PopulateFromFlagset(flags, renames)
	assert.NoError(t, err)

	flags.Set("port", "9000")
	flags.Set("dbuser", "testuser")

	assert.Equal(t, 9000, cfg.Opts.GetInt("serve.port"))
	assert.Equal(t, "testuser", cfg.Opts.GetString("mongodb.username"))
}
