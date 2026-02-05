package app

import (
	"log/slog"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitServeCmd(t *testing.T) {
	serveCmd := initServeCmd()

	assert.NotNil(t, serveCmd)
	assert.Equal(t, "serve", serveCmd.Use)
	assert.Equal(t, "start the Tournabyte API webserver", serveCmd.Short)
	assert.NotNil(t, serveCmd.RunE)

	// Test that flags are properly set
	assert.True(t, serveCmd.Flags().Lookup("port") != nil)
	assert.True(t, serveCmd.Flags().Lookup("dbhosts") != nil)
	assert.True(t, serveCmd.Flags().Lookup("dbuser") != nil)
	assert.True(t, serveCmd.Flags().Lookup("dbpass") != nil)
	assert.True(t, serveCmd.Flags().Lookup("s3url") != nil)
	assert.True(t, serveCmd.Flags().Lookup("s3access") != nil)
	assert.True(t, serveCmd.Flags().Lookup("s3secret") != nil)

	// Test default flag values
	port, err := serveCmd.Flags().GetInt("port")
	assert.NoError(t, err)
	assert.Equal(t, 8080, port)

	dbhosts, err := serveCmd.Flags().GetStringSlice("dbhosts")
	assert.NoError(t, err)
	assert.Equal(t, []string{"localhost:27017"}, dbhosts)
}

func TestDoServe(t *testing.T) {
	// Save original logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	// Save original appOpts and restore it after test
	originalAppOpts := appOpts
	defer func() { appOpts = originalAppOpts }()

	tests := []struct {
		name    string
		cmd     *cobra.Command
		args    []string
		wantErr bool
	}{
		{
			name:    "basic serve command",
			cmd:     &cobra.Command{Use: "serve"},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "serve command with args",
			cmd:     &cobra.Command{Use: "serve"},
			args:    []string{"arg1", "arg2"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up proper configuration for testing that won't cause port binding conflicts
			// In tests, we'll create a temporary app config with a different port
			testConfig := NewAppConfig("json", "appconf", []string{"/nonexistent/path"})

			// Save original appConfig and restore it after test
			originalAppConfig := appConfig
			defer func() { appConfig = originalAppConfig }()
			appConfig = testConfig

			// Mock app options with a non-conflicting port to avoid "address already in use"
			mockOpts := &ApplicationOptions{
				Serve: serviceOptions{
					Port:   9999, // Use a different port to avoid conflicts
					UseTLS: false,
				},
				Log: loggingOptions{
					Level:       "info",
					Destination: []string{"std.out"},
					UseJSON:     false,
				},
				Database: databaseOptions{
					Hosts:    []string{"localhost:27017"},
					Username: "",
					Password: "",
				},
				Filestore: filestoreOptions{
					Endpoint:  "",
					AccessKey: "",
					SecretKey: "",
				},
			}

			// Manually set appOpts to avoid nil pointer dereference
			appOpts = mockOpts

			err := doServe(tt.cmd, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRootCmd(t *testing.T) {
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "tbyte-webapi", rootCmd.Use)
	assert.Equal(t, "controls the webapi for the Tournabyte platform", rootCmd.Short)
	assert.NotNil(t, rootCmd.PersistentPreRunE)
	assert.NotNil(t, appConfig)

	// Test that serve command is added as subcommand
	commands := rootCmd.Commands()
	assert.Len(t, commands, 1)
	assert.Equal(t, "serve", commands[0].Name())
}

func TestInit(t *testing.T) {
	// Test that init() was called and added the serve command
	// Since init() runs automatically when the package is loaded,
	// we can verify the state it should have created
	assert.NotNil(t, rootCmd)

	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Name() == "serve" {
			found = true
			break
		}
	}
	assert.True(t, found, "serve command should be added by init()")
}

func TestInitAppContext_WithValidConfig(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	tempConfigPath := tempDir + "/appconf.json"

	configContent := `{
		"serve": {
			"port": 8080,
			"useTLS": false
		},
		"log": {
			"minLevel": "info",
			"destination": ["std.out"],
			"json": false
		},
		"mongodb": {
			"hosts": ["localhost:27017"],
			"username": "test",
			"password": "test"
		},
		"minio": {
			"endpoint": "http://localhost:9000",
			"accessKey": "minioadmin",
			"secretKey": "minioadmin"
		}
	}`

	err := os.WriteFile(tempConfigPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create a test app config with the temp directory
	testConfig := NewAppConfig("json", "appconf", []string{tempDir})

	// Save original appConfig and restore it after test
	originalAppConfig := appConfig
	defer func() { appConfig = originalAppConfig }()
	appConfig = testConfig

	// Create a test command
	cmd := &cobra.Command{Use: "test"}
	args := []string{}

	// Save original logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	err = initAppContext(cmd, args)
	assert.NoError(t, err)
}

func TestInitAppContext_WithInvalidConfig(t *testing.T) {
	// Create app config that points to non-existent directory
	testConfig := NewAppConfig("json", "nonexistent", []string{"/nonexistent/path"})

	// Save original appConfig and restore it after test
	originalAppConfig := appConfig
	defer func() { appConfig = originalAppConfig }()
	appConfig = testConfig

	cmd := &cobra.Command{Use: "test"}
	args := []string{}

	err := initAppContext(cmd, args)
	assert.Error(t, err)
}

func TestAppConfig_GlobalVariable(t *testing.T) {
	// Test that the global appConfig variable is properly initialized
	assert.NotNil(t, appConfig)
	assert.NotNil(t, appConfig.Opts)
}

func TestExecute_Function(t *testing.T) {
	// Save original logger to restore later
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	// Save original root command and args
	originalRootCmd := rootCmd
	defer func() { rootCmd = originalRootCmd }()

	// Create a mock command that always succeeds
	mockCmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	rootCmd = mockCmd

	// Execute should not panic
	assert.NotPanics(t, func() {
		Execute()
	})
}

func TestServeCmd_FlagsBinding(t *testing.T) {
	serveCmd := initServeCmd()

	// Test that flags can be set and retrieved
	err := serveCmd.Flags().Set("port", "9000")
	assert.NoError(t, err)

	port, err := serveCmd.Flags().GetInt("port")
	assert.NoError(t, err)
	assert.Equal(t, 9000, port)

	err = serveCmd.Flags().Set("dbuser", "testuser")
	assert.NoError(t, err)

	dbuser, err := serveCmd.Flags().GetString("dbuser")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", dbuser)

	err = serveCmd.Flags().Set("dbhosts", "host1:27017,host2:27017")
	assert.NoError(t, err)

	dbhosts, err := serveCmd.Flags().GetStringSlice("dbhosts")
	assert.NoError(t, err)
	assert.Equal(t, []string{"host1:27017", "host2:27017"}, dbhosts)
}
