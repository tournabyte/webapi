package cmd

/*
 * File: cmd/config.go
 *
 * Purpose: define the applicatio nconfiguration options
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Variable `appConfig` holds the application configuration user for CLI invokations
var appConfig *AppConfig

// Type `AppConfig` manages the application configuration lifecyle using viper
//
// Struct members:
//   - Opts: a stand-alone instance of viper.Viper (not the package singleton)
type AppConfig struct {
	Cache   *viper.Viper
	Options ApplicationOptions
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
func NewAppConfig(cfgType string, cfgName string, cfgPaths ...string) *AppConfig {
	cfg := AppConfig{Cache: viper.New()}

	cfg.Cache.SetConfigName(cfgName)
	cfg.Cache.SetConfigType(cfgType)
	for _, p := range cfgPaths {
		cfg.Cache.AddConfigPath(p)
	}

	return &cfg
}

// Function `(*AppConfig).populateFromDisk` attempts to read in the configuration source from disk as specified during construction
//
// Returns:
//   - `error`: issue that occurred when populating configuration from disk (nil if population succeeded without issue)
func (cfg *AppConfig) populateFromDisk() error {
	return cfg.Cache.ReadInConfig()
}

// Function `(*AppConfig).populateFromOverride` applies the given flag override as a configuration source
//
// Returns:
//   - `error`: issue that occurred when populating configuration from flag override (nil if population succeeded without issue)
func (cfg *AppConfig) populateFromOverride(override FlagOverride) error {
	if override.OptionValue == nil {
		return fmt.Errorf("attempted to override configuration key `%s` with a nil flag", override.OptionKey)
	}
	return cfg.Cache.BindPFlag(override.OptionKey, override.OptionValue)
}

// Function `(*AppConfig).unmarshalOptions` attempts to map the internal viper config to a `ApplicationOptions` instance
//
// Returns
//   - `error`: error indicating a failure when unmarshalling (nil if unmarshalled successfully)
func (cfg *AppConfig) unmarshalOptions() error {
	if err := cfg.Cache.Unmarshal(&cfg.Options); err != nil {
		return err
	}
	if err := resolveFiles(&cfg.Options); err != nil {
		return err
	}
	return nil
}

// Function `Bind` initiates the binding of command line flags and configuration files into program memory
//
// Parameters:
//   - overrides: command specific flag overrides for this binding attempt
//
// Returns:
//   - `error`: issue that occurred when populating configuration (nil if population succeeded without issue)
func (cfg *AppConfig) Bind(overrides ...FlagOverride) error {
	if err := cfg.populateFromDisk(); err != nil {
		return fmt.Errorf("could not populate configuration from disk: %w", err)
	}

	for _, opt := range overrides {
		if err := cfg.populateFromOverride(opt); err != nil {
			return fmt.Errorf("could not populate from flag override: %w", err)
		}
	}

	if err := cfg.unmarshalOptions(); err != nil {
		return fmt.Errorf("could not unmarshal configuration: %w", err)
	}
	return nil
}

// Type `FlagOverride` represents a mapping from a command line flag to the configuration option key it corresponds to
//
// Members:
//   - OptionKey: the configuration key that will be overridden by a command line flag
//   - OptionValue: the flag that will contain the value to override
type FlagOverride struct {
	OptionKey   string
	OptionValue *pflag.Flag
}

// Function `OverrideFromFlag` creates a flag override from the specified option key and command line flag
//
// Parameters:
//   - key: the configuration option this flag will override
//   - flag: the flag that provides the value to override the configuration key
//
// Returns:
//   - `FlagOverride`: the override specification that can be processed during application option binding
func OverrideFromFlag(key string, flag *pflag.Flag) FlagOverride {
	return FlagOverride{
		OptionKey:   key,
		OptionValue: flag,
	}
}

// Type `ApplicationOptions` represents the configuration file structure using go struct syntax
//
// Members:
//   - Serve: the service options component of the config. See the `serviceOptions` struct for more details
//   - Log: the logging configuration(s) for the application. See the `loggingOptions` struct for more details
//   - RecordStore: the database options component of the config. See the `recordStorageOptions` struct for more details
//   - ObjectStore: the object store options component of the config. See the `objectStorageOptions` struct for more details
type ApplicationOptions struct {
	Serve       serviceOptions       `mapstructure:"serve"`
	Log         []loggingOptions     `mapstructure:"log"`
	RecordStore recordStorageOptions `mapstructure:"mongodb"`
	ObjectStore objectStorageOptions `mapstructure:"minio"`
}

// Type `serviceOptions` represents the configuration components pertaining to the service customization capabilities of the API server
//
// Members:
//   - Port: the port to listen on for incoming connections
//   - Security: option set pertaining to the security setting of the API server process
//   - Sessions: option set pertaining to the session configuration of the API server authorization process
type serviceOptions struct {
	Port     uint            `mapstructure:"port"`
	Security securityOptions `mapstructure:"security"`
	Sessions sessionOptions  `mapstructure:"sessions"`
}

// Type `securityOptions` represents the options available to configure security settings for the API server
//
// Members:
//   - TLSEnabled: indicates whether the API server should use TLS or not
//   - Certificate: /path/to/cert containing the server certificate to use (will be read during configuration unmarshalling)
//   - Keychain: /path/to/keychain containing the server keycahin to use (will be read during configuration unmarshalling)
type securityOptions struct {
	TLSEnabled  bool   `mapstructure:"useTLS"`
	Certificate string `mapstructure:"certificateFile" fromfile:"required,perm=0400"`
	Keychain    string `mapstructure:"keychainFile" fromfile:"required,perm=0400"`
}

// Type `sessionOptions` represents the options available to configure how the API server deals with token-based sessions that it issues to clients
//
// Members:
//   - Algorithm: the signing algorithm the API server should use
//   - SigningKey: the secret key the API server should use to encode and decode access tokens
//   - AccessTokenTTL: the duration that an access token should remain valid
//   - RefreshTokenTTL: the duration that a refresh token should remain valid
//   - Issuer: the value to include for the `iss` field of the access token before encoding (will be validated when decoding presented access tokens)
//   - Subject: the value to include for the `sub` field of the access token before encoding (will be validated when decoding presented access tokens)
type sessionOptions struct {
	Algorithm       string        `mapstructure:"signingAlgorithm"`
	SigningKey      string        `mapstructure:"signingKeyFile" fromfile:"required,perm=0600"`
	AccessTokenTTL  time.Duration `mapstructure:"accessTokenTTL"`
	RefreshTokenTTL time.Duration `mapstructure:"refreshTokenTTL"`
	Issuer          string        `mapstructure:"tokenIssuer"`
	Subject         string        `mapstructure:"tokenSubject"`
}

// Type `recordStorageOptions` represents the options available to configure how the API server stores structured records
//
// Members:
//   - Hosts: the list of hostnames to use for database access
//   - Username: /path/to/file containing the user to authenticate to the database as during operation (will be read during configuration unmarshalling)
//   - Password: /path/to/file containing the security response to use when challenged by the database to authenticate (will be read during configuration unmarshalling)
type recordStorageOptions struct {
	Hosts    []string `mapstructure:"hosts"`
	Username string   `mapstructure:"username" fromfile:"required,perm=0600"`
	Password string   `mapstructure:"password" fromfile:"required,perm=0600"`
}

// Type `objectStorageOptions` represents the options available to configure how the API server stores unstructured data
//
// Members:
//   - Endpoint: the location of the object storage solution
//   - AccessKey: /path/to/file containing the key used to claim access to the object storage service (will be read during configuration unmarshalling)
//   - SecretKey: /path/to/file containing the key used to response to authentication challenges from the object storage service (will be read during configuration unmarshalling)
type objectStorageOptions struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"accessKey" fromfile:"required,perm=0600"`
	SecretKey string `mapstructure:"secretKey" fromfile:"required,perm=0600"`
}

// Type `loggingOptions` represents the structured logging options component of the configuration file structure
//
// Struct members:
//   - Level: the minimum log severity level to be emitted
//   - Destination: the locations to send emitted records to
//   - UseJSON: boolean indicating to emit records as JSON or plaintext
//   - UseSource: boolean indicating to include source code location in emitted records
type loggingOptions struct {
	Level       string   `mapstructure:"level"`
	Destination []string `mapstructure:"destination"`
	UseJSON     bool     `mapstructure:"json"`
	UseSource   bool     `mapstructure:"source"`
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
			if len(path) == 0 {
				return fmt.Errorf("field `%s` was tagged to be resolved from a file, but did not contain a path value", fieldT.Name)
			}
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
				buf := bytes.NewBuffer(data)
				field.SetString(strings.TrimSpace(buf.String()))
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
