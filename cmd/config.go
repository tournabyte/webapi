package cmd

/*
 * File: cmd/config.go
 *
 * Purpose: defining the application configuration options for the Tournabyte webapi
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Type `ApplicationOptions` represents the configuration file structure using go struct syntax
//
// Struct members:
//   - Serve: the service options components of the config. See `serviceOptions` for more details
//   - Database: the db options component of the config. See `databaseOptions` for more details
//   - Filestore: the file storage options component of the config. See `filestoreOptions` for more details
type ApplicationOptions struct {
	Serve     serviceOptions   `mapstructure:"serve"`
	Log       loggingOptions   `mapstructure:"log"`
	Database  databaseOptions  `mapstructure:"mongodb"`
	Filestore filestoreOptions `mapstructure:"minio"`
}

// Type `serviceOptions` represents the webapi server options component of the configuration file structure
//
// Struct members:
//   - Port: the port to listen on for incoming connection
//   - UseTLS: indicates whether the server will use HTTP or HTTPS
//   - CertFile: path to the certificate file for HTTPS operation
//   - KeyFile: path to the keychain file for HTTPS operation
type serviceOptions struct {
	Port     uint   `mapstructure:"port"`
	UseTLS   bool   `mapstructure:"useTLS"`
	CertFile string `mapstructure:"certificateFile"`
	KeyFile  string `mapstructure:"keychainFile"`
}

// Type `databaseOptions` represents the webapi database options component of the configuration file structure
//
// Struct members:
//   - Hosts: list of hostnames to use for database access (list b/c of mongodb driver options)
//   - Username: the user to authenticate to the database as during operation
//   - Password: the security response to use when challenged by the database authentication request
type databaseOptions struct {
	Hosts    []string `mapstructure:"hosts"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
}

// Type `filestoreOptions` represents the webapi file storage options component of the configuration file structure
//
// Struct members:
//   - Endpoint: the location of the file storage solution (typically a URL)
//   - AccessKey: the key used to claim access to the file storage service
//   - SecretKey: the key used to respond to authentication challenges from the file storage service
type filestoreOptions struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"accessKey"`
	SecretKey string `mapstructure:"secretKey"`
}

// Type `loggingOptions` represents the webapi structured logging options component of the configuration file structure
//
// Struct members:
//   - Level: the minimum log severity level to be emitted
//   - Destination: the location to send emitted records to
type loggingOptions struct {
	Level       int    `mapstructure:"minLevel"`
	Destination string `mapstructure:"destination"`
}

// Type `AppConfig` manages the application configuration lifecyle using viper
//
// Struct members:
//   - Opts: a stand-alone instance of viper.Viper (not the package singleton)
type AppConfig struct {
	Opts *viper.Viper
}

// Function `NewAppConfig` constructs an app config with given options
//
// Parameters:
//   - cfgType: the type of configuration this AppConfig expects to process
//   - cfgName: the name of the configuration file (without file extension) this AppConfig expects to find
//   - cfgPaths: the list of directories this AppConfig will search for the named configuration file
//
// Returns:
//   - `*AppConfig`: pointer to the application configuration source manager that is ready to populate
func NewAppConfig(cfgType string, cfgName string, cfgPaths []string) *AppConfig {
	cfg := AppConfig{Opts: viper.New()}

	cfg.Opts.SetConfigName(cfgName)
	cfg.Opts.SetConfigType(cfgType)
	for _, p := range cfgPaths {
		cfg.Opts.AddConfigPath(p)
	}

	return &cfg
}

// Function `(*AppConfig).PopulateFromFile` attempts to read in the configuration source
//
// Returns:
//   - `error`: error indicating a problem with populating the internal viper config (nil if successfully populated)
func (cfg *AppConfig) PopulateFromFile() error {
	return cfg.Opts.ReadInConfig()
}

// Function `(*AppConfig).PopulateFromFlagset` applies the supplied flags as a configuration source
//
// Parameters:
//   - flags: the flagset to apply as a configuration source
//   - renames: a mapping of configuration options to flag names to resolve naming differences
//
// Returns:
//   - `error`: error indicating a problem with populating the internal viper config (nil if successfully populated)
func (cfg *AppConfig) PopulateFromFlagset(flags *pflag.FlagSet, renames map[string]string) error {
	for optName, flagName := range renames {
		err := cfg.Opts.BindPFlag(optName, flags.Lookup(flagName))
		if err != nil {
			return err
		}
	}
	return nil
}

// Function `(*AppConfig).UnmarshalOptions` attempts to map the internal viper config to a `ApplicationOptions` instance
//
// Returns
//   - `*ApplicationOptions`: pointer to the successfully unmarshalled options structure
//   - `error`: error indicating a failure when unmarshalling (nil if unmarshalled successfully)
func (cfg *AppConfig) UnmarshalOptions() (*ApplicationOptions, error) {
	var options ApplicationOptions
	if err := cfg.Opts.Unmarshal(&options); err != nil {
		return nil, err
	}
	return &options, nil
}
