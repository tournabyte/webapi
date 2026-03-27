package app

import (
	"bytes"
	"log/slog"
	"os"
	"testing"
	"text/template"

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
	etcDir := t.TempDir()
	configPath := etcDir + "/webapi.json"

	secretsDir := t.TempDir()
	tokenKeyPath := secretsDir + "/signing_key"
	dbAccessKey := secretsDir + "/mongo_access_key"
	dbSecretKey := secretsDir + "/mongo_secret_key"
	s3AccessKey := secretsDir + "/minio_access_key"
	s3SecretKey := secretsDir + "/minio_secret_key"

	tlsCertKey := secretsDir + "/tls_certificate_file"
	tlsChainKey := secretsDir + "/tls_keychain_file"

	fdata := map[string]any{
		"certFile":        tlsCertKey,
		"chainFile":       tlsChainKey,
		"jwtkeyFile":      tokenKeyPath,
		"mongoAccessFile": dbAccessKey,
		"mongoKeyFile":    dbSecretKey,
		"minioIDFile":     s3AccessKey,
		"minioKeyFile":    s3SecretKey,
	}

	configContent := `{
  "serve": {
    "port": "3000",
    "security": {
      "useTLS": false,
      "certificateFile": "{{.certFile}}",
      "keychainFile": "{{.chainFile}}"
    },
    "sessions": {
      "signingAlgorithm": "HS256",
      "signingKeyFile": "{{.jwtkeyFile}}",
      "accessTokenTTL": "15m",
      "refreshTokenTTL": "72h",
      "tokenIssuer": "example.com",
      "tokenSubject": "Example Token Subject"
    }
  },
  "mongodb": {
    "hosts": [
      "mongodb01.tournabyte.com",
      "mongodb02.tournabyte.com",
      "mongodb03.tournabyte.com"
    ],
    "username": "{{.mongoAccessFile}}",
    "password": "{{.mongoKeyFile}}"
  },
  "minio": {
    "endpoint": "localhost:9000",
    "accessKey": "{{.minioIDFile}}",
    "secretKey": "{{.minioKeyFile}}"
  },
  "log": [
    {
      "level": "debug",
      "destination": [
        "std.out"
      ],
      "source": false,
      "json": false
    }
	]
}`
	configTemplate, err := template.New("configContent").Parse(configContent)
	require.NoError(t, err)

	configOutput := new(bytes.Buffer)
	err = configTemplate.Execute(configOutput, fdata)
	require.NoError(t, err)

	err = os.WriteFile(configPath, configOutput.Bytes(), 0644)
	require.NoError(t, err)

	err = os.WriteFile(tokenKeyPath, []byte("token"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(dbAccessKey, []byte("dbuser"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(dbSecretKey, []byte("dbkey"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(s3AccessKey, []byte("s3ID"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(s3SecretKey, []byte("s3Key"), 0600)
	require.NoError(t, err)

	err = os.WriteFile(tlsCertKey, []byte("tls.certificate"), 0400)
	require.NoError(t, err)

	err = os.WriteFile(tlsChainKey, []byte("tls.keychain"), 0400)
	require.NoError(t, err)

	// Create a test app config with the temp directory
	testConfig := NewAppConfig("json", "webapi", []string{etcDir})

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
