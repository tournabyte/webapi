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
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewOptions[C any](opts ...OptionSetter[C]) (*C, error) {
	var config *C = new(C)

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Function `defaultFactory` provides a data structure implementing the `DriverClient` interface. This default option simply wraps the `mongo.Connect` function and returns the corresponding result
//
// Parameters:
//   - ...opts: the sequence of `options.ClientOptions` to configure the resulting connection
//
// Returns:
//   - `DriverClient`: data structure implementing the critical operations needed to manage a database connection lifecycle
//   - `error`: issue reported when attempting to create the database connection (nil if connection created successfully)
func defaultFactory(opts ...*options.ClientOptions) (*mongo.Client, error) {
	return mongo.Connect(opts...)
}

// Function signature for a mongo client factory
type DriverClientFactory func(...*options.ClientOptions) (*mongo.Client, error)

// Variable `clientFactory` points to the factory function that will be used to create mongo clients
var ClientFactory DriverClientFactory = defaultFactory

// Type `DatabaseConnection` wraps the `mongo.Client` object to make lifecycle management easier
//
// Members:
//   - client: pointer to the instance implementing the `mongo.Client` under management
//   - options: configuration object for creating the corresponding driver client instance
type DatabaseConnection struct {
	client  *mongo.Client
	options *options.ClientOptions
}

// Function `NewMongoConnection` creates the connection configured as specified by the `options` member and confirms the connection can be established
// Parameters:
//   - ...opts: variadic sequence of options to specify how the connection should be configured
//
// Returns:
//   - `DatabaseConnection`: the resulting established mongodb connection with client and configuration information available for lifecycle management
//   - `error`: the issue that occurred when attempting to create the specified connection (nil if no issue occurred)
func NewMongoConnection(opts ...OptionSetter[options.ClientOptions]) (*DatabaseConnection, error) {
	var conn DatabaseConnection
	var connectTimeout time.Duration

	if config, configErr := NewOptions(opts...); configErr != nil {
		return nil, configErr
	} else {
		connect, connectErr := ClientFactory(config)
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
func (db *DatabaseConnection) Disconnect(ctx context.Context) error {
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
func (db *DatabaseConnection) SetUpSession(ctx context.Context, opts ...options.Lister[options.SessionOptions]) (context.Context, error) {
	if sess, err := db.client.StartSession(opts...); err != nil {
		return nil, fmt.Errorf("session could not be started: %w", err)
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
func (db *DatabaseConnection) TearDownSession(ctx context.Context) error {
	if sess := mongo.SessionFromContext(ctx); sess == nil {
		return errors.New("no active session to close")
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
func (db *DatabaseConnection) BeginTransaction(ctx context.Context, opts ...options.Lister[options.TransactionOptions]) error {
	if sess := mongo.SessionFromContext(ctx); sess == nil {
		return errors.New("no active session to upgrade")

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
func (db *DatabaseConnection) CommitTransaction(ctx context.Context) error {
	if sess := mongo.SessionFromContext(ctx); sess == nil {
		return errors.New("no active transaction to commit")
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
func (db *DatabaseConnection) AbortTransaction(ctx context.Context) error {
	if sess := mongo.SessionFromContext(ctx); sess == nil {
		return errors.New("no active transaction to rollback")
	} else {
		return sess.AbortTransaction(ctx)
	}
}

// Type `MinioConnection` represents a connection to the MinIO service as the associated options
//
// Struct members:
//   - client: the managed client instance connecting to a MinIO instance
//   - options: the connection options used in client creation
type MinioConnection struct {
	client  *minio.Client
	options *minio.Options
}

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
			conn.SetAppInfo("Tournabyte API", "v1")
			return &MinioConnection{client: conn, options: cfg}, nil
		}
	}
}

// Function `(*MinioConnection).GetLink` retrieve a URL to download the object specified by the given bucket and key directly from the minio instance
//
// Parameters:
//   - ctx: context managing the lifecycle of the URL generation operation
//   - bucket: the bucket name containing the target object
//   - key: the unique identifier (within bucket) of the target object
//   - notValidAfter: the duration the generated URL should be valid for
//
// Returns:
//   - `*url.URL`: the generated URL to retrieve the specified object
//   - `error`: issue that occurred during URL generation (nil if the generation succeeded)
func (conn *MinioConnection) GetLink(ctx context.Context, bucket string, key string, notValidAfter time.Duration) (*url.URL, error) {
	return conn.client.PresignedGetObject(
		ctx,
		bucket,
		key,
		notValidAfter,
		url.Values{},
	)
}

// Function `(*MinioConnection).PutLink` retrieve a URL to upload an object to the specified bucket and key directly to the minio instance
//
// Parameters:
//   - ctx: context managing the lifecycle of the URL generation operation
//   - bucket: the bucket name containing the target object
//   - key: the unique identifier (within bucket) of the target object
//   - notValidAfter: the duration the generated URL should be valid for
//
// Returns:
//   - `*url.URL`: the generated URL to retrieve the specified object
//   - `error`: issue that occurred during URL generation (nil if the generation succeeded)
func (conn *MinioConnection) PutLink(ctx context.Context, bucket string, key string, notValidAfter time.Duration) (*url.URL, error) {
	return conn.client.PresignedPutObject(
		ctx,
		bucket,
		key,
		notValidAfter,
	)
}

// Function `(*MinioConnection).Put` uploads the contents to the minio connection instance under the specified bucket and key
//
// Parameters:
//   - ctx: context managing the lifecycle of the put operation
//   - bucket: name of the bucket the upload should target
//   - key: unique object key to identify the uploaded object
//   - contents: the stuff to upload
//   - sz: the size of the stuff being uploaded
//   - ...opts: options for modifying the upload behavior
//
// Returns:
//   - `minio.UploadInfo`: pointer to the information about the upload operation
//   - `error`: issue that occured during the upload operation (nil if operation succeeded)
func (conn *MinioConnection) Put(ctx context.Context, bucket string, key string, contents io.Reader, sz int64, opts ...OptionSetter[minio.PutObjectOptions]) (*minio.UploadInfo, error) {
	var cfg minio.PutObjectOptions
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	res, err := conn.client.PutObject(
		ctx,
		bucket,
		key,
		contents,
		sz,
		cfg,
	)

	return &res, err
}

// Function `(*MinioConnection).Delete` removes the object specified by the given bucket and key
//
// Parameters:
//   - ctx: context managing the lifecycle of the delete operation
//   - bucket: bucket name containing the object to delete
//   - key: unique identifier (within bucket) of object to delete
//   - ...opts: options for modifying the delete behavior
//
// Returns:
//   - `error`: issue that occurred during the delte operation (nil if the operation succeeded)
func (conn *MinioConnection) Delete(ctx context.Context, bucket string, key string, opts ...OptionSetter[minio.RemoveObjectOptions]) error {
	var cfg minio.RemoveObjectOptions
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return err
		}
	}

	return conn.client.RemoveObject(
		ctx,
		bucket,
		key,
		cfg,
	)
}
