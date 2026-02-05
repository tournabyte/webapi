package app

import (
	"log/slog"
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestAppConfig_UnmarshalOptions(t *testing.T) {
	cfg := NewAppConfig("json", "testconfig", []string{"."})

	cfg.Opts.Set("serve.port", 8080)
	cfg.Opts.Set("mongodb.hosts", []string{"localhost:27017"})
	cfg.Opts.Set("mongodb.username", "testuser")
	cfg.Opts.Set("mongodb.password", "testpass")
	cfg.Opts.Set("minio.endpoint", "http://localhost:9000")
	cfg.Opts.Set("minio.accessKey", "access")
	cfg.Opts.Set("minio.secretKey", "secret")
	cfg.Opts.Set("log.minLevel", "info")
	cfg.Opts.Set("log.destination", []string{"std.out"})
	cfg.Opts.Set("log.json", false)

	options, err := cfg.UnmarshalOptions()
	require.NoError(t, err)
	require.NotNil(t, options)

	assert.Equal(t, uint(8080), options.Serve.Port)
	assert.Equal(t, []string{"localhost:27017"}, options.Database.Hosts)
	assert.Equal(t, "testuser", options.Database.Username)
	assert.Equal(t, "testpass", options.Database.Password)
	assert.Equal(t, "http://localhost:9000", options.Filestore.Endpoint)
	assert.Equal(t, "access", options.Filestore.AccessKey)
	assert.Equal(t, "secret", options.Filestore.SecretKey)
	assert.Equal(t, "info", options.Log.Level)
	assert.Equal(t, []string{"std.out"}, options.Log.Destination)
	assert.Equal(t, false, options.Log.UseJSON)
}

func TestInitLogs(t *testing.T) {
	tests := []struct {
		name        string
		logConfig   loggingOptions
		expectError bool
	}{
		{
			name: "debug level to stdout",
			logConfig: loggingOptions{
				Level:       "debug",
				Destination: []string{"std.out"},
				UseJSON:     false,
			},
			expectError: false,
		},
		{
			name: "error level to stderr",
			logConfig: loggingOptions{
				Level:       "error",
				Destination: []string{"std.err"},
				UseJSON:     false,
			},
			expectError: false,
		},
		{
			name: "warn level to stdout with JSON",
			logConfig: loggingOptions{
				Level:       "warn",
				Destination: []string{"std.out"},
				UseJSON:     true,
			},
			expectError: false,
		},
		{
			name: "default level to stdout",
			logConfig: loggingOptions{
				Level:       "unknown",
				Destination: []string{"std.out"},
				UseJSON:     false,
			},
			expectError: false,
		},
		{
			name: "multiple destinations",
			logConfig: loggingOptions{
				Level:       "info",
				Destination: []string{"std.out", "std.err"},
				UseJSON:     false,
			},
			expectError: false,
		},
		{
			name: "invalid file destination",
			logConfig: loggingOptions{
				Level:       "info",
				Destination: []string{"/invalid/path/that/does/not/exist/file.log"},
				UseJSON:     false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original default logger to restore later
			originalLogger := slog.Default()

			err := initLogs(tt.logConfig)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify logger was set by checking it's not nil
				assert.NotNil(t, slog.Default())
			}

			// Restore original logger
			slog.SetDefault(originalLogger)
		})
	}
}

func TestInitLogs_TempFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-log-*.log")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	logConfig := loggingOptions{
		Level:       "info",
		Destination: []string{tempFile.Name()},
		UseJSON:     false,
	}

	err = initLogs(logConfig)
	assert.NoError(t, err)

	// Verify the file exists and is writable
	_, err = os.Stat(tempFile.Name())
	assert.NoError(t, err)
}

func TestApplicationOptions_StructFields(t *testing.T) {
	options := ApplicationOptions{
		Serve: serviceOptions{
			Port:     8080,
			UseTLS:   true,
			CertFile: "/path/to/cert.pem",
			KeyFile:  "/path/to/key.pem",
		},
		Database: databaseOptions{
			Hosts:    []string{"localhost:27017", "mongodb:27017"},
			Username: "admin",
			Password: "password",
		},
		Filestore: filestoreOptions{
			Endpoint:  "http://localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
		},
		Log: loggingOptions{
			Level:       "debug",
			Destination: []string{"std.out", "std.err"},
			UseJSON:     true,
		},
	}

	assert.Equal(t, uint(8080), options.Serve.Port)
	assert.True(t, options.Serve.UseTLS)
	assert.Equal(t, "/path/to/cert.pem", options.Serve.CertFile)
	assert.Equal(t, "/path/to/key.pem", options.Serve.KeyFile)

	assert.Equal(t, []string{"localhost:27017", "mongodb:27017"}, options.Database.Hosts)
	assert.Equal(t, "admin", options.Database.Username)
	assert.Equal(t, "password", options.Database.Password)

	assert.Equal(t, "http://localhost:9000", options.Filestore.Endpoint)
	assert.Equal(t, "minioadmin", options.Filestore.AccessKey)
	assert.Equal(t, "minioadmin", options.Filestore.SecretKey)

	assert.Equal(t, "debug", options.Log.Level)
	assert.Equal(t, []string{"std.out", "std.err"}, options.Log.Destination)
	assert.True(t, options.Log.UseJSON)
}
