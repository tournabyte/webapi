package utils

/*
 * File: internal/utils/s3toolkit.go
 *
 * Purpose: managing the lifecycle of the minio client instance
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"errors"
	"io"
	"net/url"
	"slices"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/tournabyte/webapi/app"
)

// Type `MinioClientOption` is a functional setter for the MinIO client options
type MinioClientOption func(*minio.Options) error

// Type `MinioPutOption` is a functional setter for the MinIO put operation options
type MinioPutOption func(*minio.PutObjectOptions) error

// Type `MinioDeleteOption` is a functional setter for the MinIO delete operation options
type MinioDeleteOption func(*minio.RemoveObjectOptions) error

// Function `MinioStaticCredentials` provides the option setter to utilized the provided static credentials
//
// Parameters:
//   - accessKey: the access ID for accessing the storage service
//   - secretKey: the secret for responding to authentication requests
//
// Returns:
//   - `MinioClientOption`: functional setter to use the provided credentials
func MinioStaticCredentials(accessKey string, secretKey string) MinioClientOption {
	return func(o *minio.Options) error {
		o.Creds = credentials.NewStaticV4(
			accessKey,
			secretKey,
			"",
		)
		return nil
	}
}

// Function `MinioUseSecureConnection` provides the option setter to utilize secure connections if enabled
//
// Parameters:
//   - secure: set the secure connection setting accordingly
//
// Returns:
//   - `MinioClientOption`: functional setter to use the secure conection option
func MinioUseSecureConnection(secure bool) MinioClientOption {
	return func(o *minio.Options) error {
		o.Secure = secure
		return nil
	}
}

// Function `MinioMaxRetries` provides the option setter to utilize the provided retry limit
//
// Parameters:
//   - retryLimit: the number of retries to allow
//
// Returns:
//   - `MinioClientOption`: functional setter to use the retry limit option
func MinioMaxRetries(retryLimit int) MinioClientOption {
	return func(o *minio.Options) error {
		o.MaxRetries = retryLimit
		return nil
	}
}

// Function `PutObjectMetadata` provies the option setter to utilize the provided object metadata
//
// Parameters:
//   - ...values: sequence of key-value pairs to use as metadata
//
// Returns:
//   - `MinioPutOption`: functional setter to use the user metadata option
func PutObjectMetadata(values ...string) MinioPutOption {
	return func(o *minio.PutObjectOptions) error {
		if len(values)%2 != 0 {
			return errors.New("received odd length input for user metadata option")
		}

		metadata := make(map[string]string)
		for pair := range slices.Chunk(values, 2) {
			key, value := pair[0], pair[1]
			metadata[key] = value
		}
		o.UserMetadata = metadata

		return nil
	}
}

// Function `PutObjectTags` provides the option setter to utilize the provided object tags
//
// Parameters:
//   - ...values: sequence of key-value pairs to use as tags
//
// Returns:
//   - `MinioPutOption`: functional setter to use the user tags option
func PutObjectTags(values ...string) MinioPutOption {
	return func(o *minio.PutObjectOptions) error {
		if len(values)%2 != 0 {
			return errors.New("received odd length input for user tags option")
		}

		tags := make(map[string]string)
		for pair := range slices.Chunk(values, 2) {
			key, value := pair[0], pair[1]
			tags[key] = value
		}
		o.UserTags = tags

		return nil
	}
}

// Function `PutObjectContentType` provides the option setter to utilize the provided content type label
//
// Parameters:
//   - contentType: the content type to use as the content label
//
// Returns:
//   - `MinioPutOption`: functional setter to use the content type label option
func PutObjectContentType(contentType string) MinioPutOption {
	return func(o *minio.PutObjectOptions) error {
		o.ContentType = contentType
		return nil
	}
}

// Function `DeleteObjectForced` provides the option setter to utilize the provided force delete option
//
// Parameters:
//   - force: true to issue a forced flag
//
// Returns:
//   - `MinioDeleteOption`: functional setter to use the force delete option
func DeleteObjectForced(force bool) MinioDeleteOption {
	return func(o *minio.RemoveObjectOptions) error {
		o.ForceDelete = force
		return nil
	}
}

// Function `DeleteObjectBypassGovernancePolicy` provides the option setter to utilize the provide governance bypass option
//
// Parameters:
//   - bypass: true to skip governance enforcement
//
// Returns:
//   - `MinioDeleteOption`: functional setter to use the bypass governance option
func DeleteObjectBypassGovernancePolicy(bypass bool) MinioDeleteOption {
	return func(o *minio.RemoveObjectOptions) error {
		o.GovernanceBypass = bypass
		return nil
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

// Function `MinioConnFromConfig` creates the MinioConnection instance reflecting the options presented in the provided application options
//
// Parameters:
//   - cfg: the application configuration to extract minio options from
//
// Returns:
//   - `MinioConnection`: minio client lifecycle manager on success
//   - `error`: reported issue on failure
func MinioConnFromConfig(cfg *app.ApplicationOptions) (*MinioConnection, error) {
	return NewMinioConnection(
		cfg.Filestore.Endpoint,
		MinioStaticCredentials(cfg.Filestore.AccessKey, cfg.Filestore.SecretKey),
		MinioUseSecureConnection(false),
		MinioMaxRetries(3),
	)
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
func NewMinioConnection(endpoint string, opts ...MinioClientOption) (*MinioConnection, error) {
	var cfg *minio.Options

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if conn, err := minio.New(endpoint, cfg); err != nil {
		return nil, err
	} else {
		conn.SetAppInfo("Tournabyte API", "v1")
		return &MinioConnection{client: conn, options: cfg}, nil
	}
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
func (conn *MinioConnection) Put(ctx context.Context, bucket string, key string, contents io.Reader, sz int64, opts ...MinioPutOption) (*minio.UploadInfo, error) {
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
func (conn *MinioConnection) Delete(ctx context.Context, bucket string, key string, opts ...MinioDeleteOption) error {
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
