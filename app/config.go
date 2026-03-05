package app

/*
 * File: app/config.go
 *
 * Purpose: defining the application configuration options for the Tournabyte webapi
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Variable `appOpts` holds the unmarshalled application configuration singleton
var appOpts *ApplicationOptions = &ApplicationOptions{}

// Type `ApplicationOptions` represents the configuration file structure using go struct syntax
//
// Struct members:
//   - Serve: the service options components of the config. See `serviceOptions` for more details
//   - Database: the db options component of the config. See `databaseOptions` for more details
//   - Filestore: the file storage options component of the config. See `filestoreOptions` for more details
type ApplicationOptions struct {
	Serve     serviceOptions   `mapstructure:"serve"`
	Log       []loggingOptions `mapstructure:"log"`
	Database  databaseOptions  `mapstructure:"mongodb"`
	Filestore filestoreOptions `mapstructure:"minio"`
}

// Function `resolveFiles` recursively resolves the fields tagged with the `fromfile` tag by treating the existing value as a filepath and overwriting the value with the file's contents
//
// Parameters:
//   - v: the structure to walk recursively and revolve files of
//
// Returns:
//   - `error`: issue that occurred during file resolution (nil if no issue occurred)
func resolveFiles(v any) error {
	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Pointer {
		return errors.New("pointer required to resolve file-like fields")
	}

	return resolve(val.Elem())
}

// Function `resolve` replaces the given value (which should be a filepath) with the file's contents
//
// Parameters:
//   - val: the value to resolve
//
// Returns:
//   - `error`: issue that occurred during resolution (nil if no issue occurred)
func resolve(val reflect.Value) error {
	if val.Kind() != reflect.Struct {
		return nil
	}

	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldT := t.Field(i)

		if tag, exists := fieldT.Tag.Lookup("fromfile"); exists {
			var required bool
			var mode os.FileMode

			if r, m, err := parseFromFileTag(tag); err != nil {
				return err
			} else {
				required = *r
				mode = *m
			}

			if field.Kind() != reflect.String {
				return errors.New("can only resolve fields of type string")
			}

			path := field.String()
			if info, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) && !required {
					return nil
				}
				return err
			} else {
				perms := info.Mode().Perm()

				if perms > mode {
					return fmt.Errorf("permissions on `%s` are too open (expected <= %o)", path, mode)
				}
			}

			if data, err := os.ReadFile(path); err != nil {
				return err
			} else {
				field.SetString(string(data))
				continue
			}
		}

		switch field.Kind() {
		case reflect.Struct:
			if err := resolve(field); err != nil {
				return err
			}

		case reflect.Pointer:
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				if err := resolve(field); err != nil {
					return err
				}
			}

		case reflect.Slice, reflect.Array:
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.Kind() == reflect.Struct {
					if err := resolve(elem); err != nil {
						return err
					}
				}

				if elem.Kind() == reflect.Pointer && elem.Elem().Kind() == reflect.Struct {
					if err := resolve(elem.Elem()); err != nil {
						return err
					}
				}
			}
		}

	}

	return nil
}

// Function `parseFromFileTag` parses the given tag and extracts the required and file mode information contained within it
//
// Parameters:
//   - tag: the tag to parse
//
// Returns:
//   - `*bool`: whether or not the file path must exists
//   - `*os.FileMode`: the minimum permission allowed for the file
//   - `error`: issue that occurred during parsing (nil if no issue occurred)
func parseFromFileTag(tag string) (*bool, *os.FileMode, error) {
	var required bool = false
	var mode os.FileMode
	var parseErr error = nil

	for part := range strings.SplitSeq(tag, ",") {
		part = strings.TrimSpace(part)
		switch {
		case part == "required":
			required = true
		case strings.HasPrefix(part, "perm="):
			if val, err := strconv.ParseUint(strings.TrimPrefix(part, "perm="), 8, 32); err != nil {
				parseErr = err
			} else {
				mode = os.FileMode(val)
			}
		}
	}

	return &required, &mode, parseErr
}

// Type `serviceOptions` represents the webapi server options component of the configuration file structure
//
// Struct members:
//   - Port: the port to listen on for incoming connection
//   - UseTLS: indicates whether the server will use HTTP or HTTPS
//   - CertFile: path to the certificate file for HTTPS operation
//   - KeyFile: path to the keychain file for HTTPS operation
//   - Tokens: options supporting the web token capabilities of the server
type serviceOptions struct {
	Port     uint            `mapstructure:"port"`
	Security securityOptions `mapstructure:"security"`
	Sessions sessionOpts     `mapstructure:"sessions"`
}

type sessionOpts struct {
	Algorithm       string        `mapstructure:"signingAlgorithm"`
	SigningKey      string        `mapstructure:"signingKeyFile" fromfile:"required,perm=0600"`
	AccessTokenTTL  time.Duration `mapstructure:"accessTokenTTL"`
	RefreshTokenTTL time.Duration `mapstructure:"refreshTokenTTL"`
	Issuer          string        `mapstructure:"tokenIssuer"`
	Subject         string        `mapstructure:"tokenSubject"`
}

type securityOptions struct {
	TLSEnabled  bool   `mapstructure:"useTLS"`
	Certificate string `mapstructure:"certificateFile" fromfile:"required,perm=0400"`
	Keychain    string `mapstructure:"keychainFile" fromfile:"required,perm=0400"`
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
	Password string   `mapstructure:"password" fromfile:"required,perm=0600"`
}

// Type `filestoreOptions` represents the webapi file storage options component of the configuration file structure
//
// Struct members:
//   - Endpoint: the location of the file storage solution (typically a URL)
//   - AccessKey: the key used to claim access to the file storage service
//   - SecretKey: the key used to respond to authentication challenges from the file storage service
type filestoreOptions struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"accessKey" fromfile:"required,perm=0600"`
	SecretKey string `mapstructure:"secretKey" fromfile:"required,perm=0600"`
}

// Type `loggingOptions` represents the webapi structured logging options component of the configuration file structure
//
// Struct members:
//   - Level: the minimum log severity level to be emitted
//   - Destination: the locations to send emitted records to
//   - UseJSON: boolean indicating to emit records as JSON or plaintext
type loggingOptions struct {
	Level       string   `mapstructure:"level"`
	Destination []string `mapstructure:"destination"`
	UseJSON     bool     `mapstructure:"json"`
	UseSource   bool     `mapstructure:"source"`
}

// Function `initLogs` initializes structured logging for the application
//
// Parameters:
//   - logConfig: a `loggingOptions` instances sourced from the application configuration
//
// Returns:
//   - `error`: issue with logging setup (if any)
func initLogs(logConfigs ...loggingOptions) error {
	var handlers []slog.Handler

	for _, cfg := range logConfigs {
		if h, err := makeHandler(cfg); err != nil {
			return err
		} else {
			handlers = append(handlers, h)
		}
	}

	slog.SetDefault(
		slog.New(slog.NewMultiHandler(handlers...)),
	)
	return nil
}

// Function `makeHandler` creates the logging record handler corresponding to the given configuration object
//
// Parameters:
//   - cfg: the logging configuration object to customize the resulting handler
//
// Returns:
//   - `slog.Handler`: the handler to process logging records
//   - `error`: the issue with producing the handler (nil if handler created successfully)
func makeHandler(cfg loggingOptions) (slog.Handler, error) {
	var level slog.Level
	var outputs []io.Writer
	var handler slog.Handler
	var opts slog.HandlerOptions

	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "error":
		level = slog.LevelError
	case "warn":
		level = slog.LevelWarn
	case "info":
		level = slog.LevelInfo
	default:
		return nil, errors.New("Invalid logging level provided")
	}

	for _, dst := range cfg.Destination {
		switch dst {
		case "std.out":
			outputs = append(outputs, os.Stdout)
		case "std.err":
			outputs = append(outputs, os.Stderr)
		default:
			if f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
				return nil, err
			} else {
				outputs = append(outputs, f)
			}
		}
	}

	opts = slog.HandlerOptions{Level: level, AddSource: cfg.UseSource}
	if cfg.UseJSON {
		handler = slog.NewJSONHandler(io.MultiWriter(outputs...), &opts)
	} else {
		handler = slog.NewTextHandler(io.MultiWriter(outputs...), &opts)
	}
	return handler, nil
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
	slog.Debug("New application configuration instance: ", slog.String("type", cfgType), slog.String("name", cfgName), slog.Any("search paths", cfgPaths))
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
	slog.Debug("Populate configuration from file...")
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
	slog.Debug("Populating configuration from flagset...")
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
//   - `error`: error indicating a failure when unmarshalling (nil if unmarshalled successfully)
func (cfg *AppConfig) UnmarshalOptions() error {
	slog.Debug("Unmarshalling configuration options...")
	if err := cfg.Opts.Unmarshal(appOpts); err != nil {
		slog.Error("Could not unmarshall configuration", slog.String("reason", err.Error()))
		return err
	}
	if err := resolveFiles(appOpts); err != nil {
		slog.Error("Could not resolve configuration file paths with their contents", slog.String("reason", err.Error()))
		return err
	}
	return nil
}
