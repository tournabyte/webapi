package cmd_test

/*
 * File: cmd/config_test.go
 *
 * Purpose: testing application configuration logic
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"bytes"
	"os"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tournabyte/webapi/cmd"
)

const configContentTemplate = `{
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
      "mongodb01.example.com",
      "mongodb02.example.com",
      "mongodb03.example.com"
    ],
    "username": "{{.mongoAccessFile}}",
    "password": "{{.mongoKeyFile}}"
  },
  "minio": {
    "endpoint": "minio.example.com:9000",
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
}
`

func setUpValidAppConfigEnv(t *testing.T, configContentTemplate string, cfgFilemode os.FileMode, testConfig *cmd.AppConfig) {
	t.Helper()

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

	configTemplate, err := template.New("configContent").Parse(configContentTemplate)
	require.NoError(t, err)

	configOutput := new(bytes.Buffer)
	err = configTemplate.Execute(configOutput, fdata)
	require.NoError(t, err)

	err = os.WriteFile(configPath, configOutput.Bytes(), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tokenKeyPath, []byte("token"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(dbAccessKey, []byte("dbuser"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(dbSecretKey, []byte("dbkey"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(s3AccessKey, []byte("s3ID"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(s3SecretKey, []byte("s3Key"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tlsCertKey, []byte("tls.certificate"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tlsChainKey, []byte("tls.keychain"), cfgFilemode)
	require.NoError(t, err)

	*testConfig = *cmd.NewAppConfig("json", "webapi", etcDir)
}

func setUpValidAppConfigEnvWithMissingConfigFile(t *testing.T, configContentTemplate string, cfgFilemode os.FileMode, testConfig *cmd.AppConfig) {
	t.Helper()

	etcDir := t.TempDir()
	// configPath := etcDir + "/webapi.json"

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

	configTemplate, err := template.New("configContent").Parse(configContentTemplate)
	require.NoError(t, err)

	configOutput := new(bytes.Buffer)
	err = configTemplate.Execute(configOutput, fdata)
	require.NoError(t, err)

	/* -- The configuration file could not be found
	err = os.WriteFile(configPath, configOutput.Bytes(), 0644)
	require.NoError(t, err)
	*/

	err = os.WriteFile(tokenKeyPath, []byte("token"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(dbAccessKey, []byte("dbuser"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(dbSecretKey, []byte("dbkey"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(s3AccessKey, []byte("s3ID"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(s3SecretKey, []byte("s3Key"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tlsCertKey, []byte("tls.certificate"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tlsChainKey, []byte("tls.keychain"), cfgFilemode)
	require.NoError(t, err)

	*testConfig = *cmd.NewAppConfig("json", "webapi", etcDir)

}

func setUpValidAppConfigEnvWithInvalidSecretPathValue(t *testing.T, configContentTemplate string, cfgFilemode os.FileMode, testConfig *cmd.AppConfig) {
	t.Helper()

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
		"certFile":        "",
		"chainFile":       "",
		"jwtkeyFile":      tokenKeyPath,
		"mongoAccessFile": dbAccessKey,
		"mongoKeyFile":    dbSecretKey,
		"minioIDFile":     s3AccessKey,
		"minioKeyFile":    s3SecretKey,
	}

	configTemplate, err := template.New("configContent").Parse(configContentTemplate)
	require.NoError(t, err)

	configOutput := new(bytes.Buffer)
	err = configTemplate.Execute(configOutput, fdata)
	require.NoError(t, err)

	err = os.WriteFile(configPath, configOutput.Bytes(), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tokenKeyPath, []byte("token"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(dbAccessKey, []byte("dbuser"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(dbSecretKey, []byte("dbkey"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(s3AccessKey, []byte("s3ID"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(s3SecretKey, []byte("s3Key"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tlsCertKey, []byte("tls.certificate"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tlsChainKey, []byte("tls.keychain"), cfgFilemode)
	require.NoError(t, err)

	*testConfig = *cmd.NewAppConfig("json", "webapi", etcDir)
}

func setUpValidAppConfigEnvWithMissingSecretFile(t *testing.T, configContentTemplate string, cfgFilemode os.FileMode, testConfig *cmd.AppConfig) {
	t.Helper()

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

	configTemplate, err := template.New("configContent").Parse(configContentTemplate)
	require.NoError(t, err)

	configOutput := new(bytes.Buffer)
	err = configTemplate.Execute(configOutput, fdata)
	require.NoError(t, err)

	err = os.WriteFile(configPath, configOutput.Bytes(), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tokenKeyPath, []byte("token"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(dbAccessKey, []byte("dbuser"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(dbSecretKey, []byte("dbkey"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(s3AccessKey, []byte("s3ID"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(s3SecretKey, []byte("s3Key"), cfgFilemode)
	require.NoError(t, err)

	/* -- Missing secrets will fail to resolve
	err = os.WriteFile(tlsCertKey, []byte("tls.certificate"), cfgFilemode)
	require.NoError(t, err)

	err = os.WriteFile(tlsChainKey, []byte("tls.keychain"), cfgFilemode)
	require.NoError(t, err)
	*/

	*testConfig = *cmd.NewAppConfig("json", "webapi", etcDir)
}

func TestAppConfigBindedSuccessfully(t *testing.T) {
	var testConfig cmd.AppConfig
	setUpValidAppConfigEnv(t, configContentTemplate, 0400, &testConfig)

	assert.NoError(t, testConfig.Bind())
}

func TestAppConfigLoadedFailure_FilesystemError(t *testing.T) {
	var testConfig cmd.AppConfig
	setUpValidAppConfigEnvWithMissingConfigFile(t, configContentTemplate, 0400, &testConfig)

	assert.Error(t, testConfig.Bind()) // TODO:specific error checks
}

func TestAppConfigLoadedFailure_FlagOverrideError(t *testing.T) {
	var testConfig cmd.AppConfig
	setUpValidAppConfigEnv(t, configContentTemplate, 0400, &testConfig)

	assert.Error(t, testConfig.Bind(cmd.OverrideFromFlag("badflag", nil))) //TODO: specific error check
}

func TestAppConfigLoadedFailure_UnmarshallingError(t *testing.T) {
	t.Run("SyntaxIssue", func(t *testing.T) {
		var testConfig cmd.AppConfig
		setUpValidAppConfigEnv(t, configContentTemplate+"]]]", 0400, &testConfig)

		assert.Error(t, testConfig.Bind()) //TODO: specific error check
	})

	t.Run("PathValueIssue", func(t *testing.T) {
		var testConfig cmd.AppConfig
		setUpValidAppConfigEnvWithInvalidSecretPathValue(t, configContentTemplate, 0400, &testConfig)

		assert.Error(t, testConfig.Bind()) // TODO: specific error check
	})

	t.Run("PathValiePermissionIssue", func(t *testing.T) {
		var testConfig cmd.AppConfig
		setUpValidAppConfigEnv(t, configContentTemplate, 0644, &testConfig)

		assert.Error(t, testConfig.Bind()) //TODO: specific error check
	})

	t.Run("PathValueNotExistsIssue", func(t *testing.T) {
		var testConfig cmd.AppConfig
		setUpValidAppConfigEnvWithMissingSecretFile(t, configContentTemplate, 0400, &testConfig)

		assert.Error(t, testConfig.Bind()) //TODO: specific error check
	})
}
