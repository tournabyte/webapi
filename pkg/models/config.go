package models

/*
 * File: pkg/models/config.go
 *
 * Purpose: define the service configuration options
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import "time"

// Type `ApplicationOptions` represents the configuration file structure using go struct syntax
//
// Members:
//   - Serve: the service options component of the config. See the `serviceOptions` struct for more details
//   - Log: the logging configuration(s) for the application. See the `loggingOptions` struct for more details
//   - RecordStore: the database options component of the config. See the `recordStorageOptions` struct for more details
//   - ObjectStore: the object store options component of the config. See the `objectStorageOptions` struct for more details
type ApplicationOptions struct {
	Serve       serviceOptions       `mapstructure:"serve"`
	Log         loggingOptions       `mapstructure:"log"`
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
	Outputs []string `mapstructure:"destinations"`
	Prefix  string   `mapstructure:"prefix"`
	Flags   int      `mapstructure:"flags"`
}
