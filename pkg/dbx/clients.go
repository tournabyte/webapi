package dbx

/*
 * File: pkg/dbx/clients.go
 *
 * Purpose: client managing logic for MongoDB and MinIO clients
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Errors for attempting to access invalid session pointers
var (
	ErrInvalidSessConfig = errors.New("invalid session configuration")
	ErrNoSessInCtx       = errors.New("provided context did not have a session value")
)

// Function `NewOptions` creates a config object with the provided option setter functions
//
// Type Parameters:
//   - C: the config object to be created and configured by setters
//
// Parameters:
//   - ...opts: variadic sequence of option setters to apply to the created configuration object
//
// Returns:
//   - `*C`: pointer to the created and configured configuration object
//   - `error`: issue that occurred during object configuration (nil if no issue occurred)
func NewOptions[C any](opts ...OptionSetter[C]) (*C, error) {
	var config *C = new(C)

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Function `MongoFromContext` retrieves the pointer to the `mongo.Session` pointer value within the provided context
//
// Parameters:
//   - ctx: the context to retrieve the `mongo.Session` from (if it exists)
//
// Returns:
//   - `*mongo.Session`: pointer to the session stored within the context
//   - `error`: the error that occurred when attempting to read the session value
func MongoFromContext(ctx context.Context) (*mongo.Session, error) {
	if sess := mongo.SessionFromContext(ctx); sess == nil {
		return nil, ErrNoSessInCtx
	} else {
		return sess, nil
	}
}

// Function `MinioFromContext` retrieves the pointer to the `minio.Client` pointer value within the provided context
//
// Parameters:
//   - ctx: the context to retrieve the `minio.Client` from (if it exists)
//
// Returns:
//   - `*minio.Client`: pointer to the client stored within the context
//   - `error`: the error that occurred when attempting to read the session value
func MinioFromContext(ctx context.Context) (*minio.Client, error) {
	val := ctx.Value(minioSessionKey{})
	if val == nil {
		return nil, ErrNoSessInCtx
	}

	sess, ok := val.(*minio.Client)
	if !ok {
		return nil, ErrNoSessInCtx
	}

	return sess, nil
}

// Type `MongoConnection` wraps the `mongo.Client` object to make lifecycle management easier
//
// Members:
//   - client: pointer to the instance implementing the `mongo.Client` under management
//   - options: configuration object for creating the corresponding driver client instance
type MongoConnection struct {
	client  *mongo.Client
	options *options.ClientOptions
}

// Function `NewMongoConnection` creates the connection configured as specified by the `options` member and confirms the connection can be established
// Parameters:
//   - ...opts: variadic sequence of options to specify how the connection should be configured
//
// Returns:
//   - `*MongoConnection`: the resulting established mongodb connection with client and configuration information available for lifecycle management
//   - `error`: the issue that occurred when attempting to create the specified connection (nil if no issue occurred)
func NewMongoConnection(opts ...OptionSetter[options.ClientOptions]) (*MongoConnection, error) {
	var conn MongoConnection
	var connectTimeout time.Duration

	if config, configErr := NewOptions(opts...); configErr != nil {
		return nil, configErr
	} else {
		connect, connectErr := mongo.Connect(config)
		if connectErr != nil {
			return nil, connectErr
		}

		conn.client = connect
		conn.options = config

		if config.ConnectTimeout == nil {
			connectTimeout = 5 * time.Second
		} else {
			connectTimeout = *config.ConnectTimeout
		}
		ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
		defer cancel()

		if pingErr := connect.Ping(ctx, config.ReadPreference); pingErr != nil {
			connect.Disconnect(context.Background())
			return nil, pingErr
		}

		return &conn, nil
	}

}

// Function `DatabaseConnection.Disconnect` initiates the disconnect sequence for the underlying mongo-driver client instance
// Parameters:
//   - ctx: the context managing the lifetime of the disconnect sequence
//
// Returns:
//   - `error`: the issue that occurred during the disconnect sequence (nil if no issue occurred)
func (db *MongoConnection) Disconnect(ctx context.Context) error {
	return db.client.Disconnect(ctx)
}

// Function `DatabaseConnection.SetUpSession` initializes the context session for the given context
//
// Parameters:
//   - ctx: the parent context to add the session to
//
// Returns:
//   - `context.Context`: a child context containing a `mongo.Session` for future use
//   - `error`: issue that occurred during session setup (nil if no issue occurred)
func (db *MongoConnection) SetUpSession(ctx context.Context, opts ...options.Lister[options.SessionOptions]) (context.Context, error) {
	if sess, err := db.client.StartSession(opts...); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSessConfig, err)
	} else {
		return mongo.NewSessionContext(ctx, sess), nil
	}
}

// Function `DatabaseConnection.TearDownSession` ends the context session for the gien context. It is useful to defer this function after obtaining a session context from `DatabaseConnection.SetUpSession`
//
// Parameters:
//   - ctx: the context containing a session value that needs to be closed
//
// Returns:
//   - `error`: issue that occurred when attempting to wrap up the session (nil if no issue occurred)
func (db *MongoConnection) TearDownSession(ctx context.Context) error {
	if sess := mongo.SessionFromContext(ctx); sess == nil {
		return ErrNoSessInCtx
	} else {
		sess.EndSession(ctx)
		return nil
	}
}

// Function `DatabaseConnection.BeginTransaction` starts a transaction on the given session context.
//
// Parameters:
//   - ctx: the context containing a session value that will be upgraded to a transaction
//
// Returns:
//   - `error`: issue that occured when initializing a transaction (nil if no issue occurred)
func (db *MongoConnection) BeginTransaction(ctx context.Context, opts ...options.Lister[options.TransactionOptions]) error {
	if sess := mongo.SessionFromContext(ctx); sess == nil {
		return ErrNoSessInCtx

	} else {
		return sess.StartTransaction(opts...)
	}
}

// Function `DatabaseConnection.CommitTransaction` commits the currently running transaction
//
// Parameters:
//   - ctx: the context containing a session value with an active transaction
//
// Returns:
//   - `error`: issue that occurred during transaction commit (nil if no issue occurred)
func (db *MongoConnection) CommitTransaction(ctx context.Context) error {
	if sess := mongo.SessionFromContext(ctx); sess == nil {
		return ErrNoSessInCtx
	} else {
		return sess.CommitTransaction(ctx)
	}
}

// Function `DatabaseConnection.AbortTransaction` aborts the currently running transaction
//
// Parameters:
//   - ctx: the context containing a session value with an active transaction
//
// Returns:
//   - `error`: issue that occurred during transaction abort (nil if no issue occurred)
func (db *MongoConnection) AbortTransaction(ctx context.Context) error {
	if sess := mongo.SessionFromContext(ctx); sess == nil {
		return ErrNoSessInCtx
	} else {
		return sess.AbortTransaction(ctx)
	}
}

// Type `MinioConnection` represents a connection to the MinIO service as the associated options
//
// Members:
//   - client: the managed client instance connecting to a MinIO instance
//   - options: the connection options used in client creation
type MinioConnection struct {
	client  *minio.Client
	options *minio.Options
}

// Type `minioSessionKey` is an internal context key for minio clients attached to a context
type minioSessionKey struct{}

// Function `NewMinioConnection` creates the MinioConnection instance from the given endpoint and connection option setters
//
// Parameters:
//   - endpoint: the URL targeting the MinIO service instance
//   - ...opts: variadic list of minio client option setters to customize client behavior
//
// Returns:
//   - `MinioConnection`: pointer to the minio client lifecycle manager
//   - `error`: issue that prevented client construction (nil if construction was successful)
func NewMinioConnection(endpoint string, opts ...OptionSetter[minio.Options]) (*MinioConnection, error) {

	if cfg, err := NewOptions(opts...); err != nil {
		return nil, err
	} else {
		if conn, err := minio.New(endpoint, cfg); err != nil {
			return nil, err
		} else {
			return &MinioConnection{client: conn, options: cfg}, nil
		}
	}
}

// Function `(*MinioConnection).SetApplicationInfo` sets application info on the underlying client
//
// Parameters:
//   - name: the name portion of the application info
//   - version: the version portion of the application info
func (conn *MinioConnection) SetApplicationInfo(name string, version string) {
	conn.client.SetAppInfo(name, version)
}

// Function `(*MinioConnection).SetUpSession` initializes a minio client session within the given context
//
// Parameters:
//   - ctx: the context to set up the session within
//
// Returns:
//   - `context.Context`: the context with the client as a key/value pair
func (conn *MinioConnection) SetUpSession(ctx context.Context) context.Context {
	return context.WithValue(ctx, minioSessionKey{}, conn.client)
}
