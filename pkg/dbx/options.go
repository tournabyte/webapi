package dbx

/*
 * File: pkg/dbx/options.go
 *
 * Purpose: options for customizing clients for MongoDB and MinIO
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"crypto/tls"
	"errors"
	"slices"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver"
)

// Constants refer to bit flags for the `options.BSONOptions` fields
const (
	UseJSONTags uint32 = 0b0001 << iota
	ErrOnInlineDuplicates
	UseIntMinSize
	NilMapsAsEmpty
	NilSliceAsEmpty
	NilByteSliceAsEmpty
	OmitZeroStruct
	OmitEmpty
	StringifyMapKeyWithFmt
	AllowTruncatingDoubles
	BinaryAsSlice
	DefaultDocumentM
	DefaultDocumentMap
	ObjectIDAsHexString
	UseLocalTimezone
	ZeroMaps
	ZeroStructs
)

// Type `OptionSetter[T] is a generic function pointer that modifies an options type struct and returns an error`
type OptionSetter[T any] func(*T) error

// Function `MongoClientAppName` provides a OptionSetter[options.ClientOptions] to set the connection's application name
//
// Parameters:
//   - name: the application name to use for the connection option
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's app name setting
func MongoClientAppName(name string) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetAppName(name)
		return nil
	}
}

// Function `MongoClientOperationTimeout` provides the OptionSetter[options.ClientOptions] to set the connection's timeout setting
//
// Parameters:
//   - to: max time to wait for operations (only respected if operation context does not specify a deadline)
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's timeout policy
func MongoClientOperationTimeout(to time.Duration) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetTimeout(to)
		return nil
	}
}

// Function `MongoClientHosts` provides the OptionSetter[options.ClientOptions] to set the connection's host lists
//
// Parameters:
//   - ...hosts: list of hosts for db servers to try for operations
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's host list
func MongoClientHosts(hosts ...string) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetHosts(hosts)
		return nil
	}
}

// Function `MinimumPoolSize` provides the OptionSetter[options.ClientOptions] to set the connection's min pool size
//
// Parameters:
//   - sz: least number of connection that can be in the connection pool at once
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's min pool size setting
func MinimumPoolSize(sz uint64) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetMinPoolSize(sz)
		return nil
	}
}

// Function `MaximumPoolSize` provides the OptionSetter[options.ClientOptions] to set the connection's max pool size
//
// Parameters:
//   - sz: most number of connection that can be in the connection pool at once
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's max pool size setting
func MaximumPoolSize(sz uint64) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetMaxPoolSize(sz)
		return nil
	}
}

// Function `ConnectionDeployment` provides the OptionSetter[options.ClientOptions] to set the connection's internal deployment
//
// Parameters:
//   - d: the deployment to use for the connection
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's internal deployment setting
func ConnectionDeployment(d driver.Deployment) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.Deployment = d
		return nil
	}
}

// Function `MongoClientCredentials` provides the OptionSetter[options.ClientOptions] to set the connection's authentication details
//
// Parameters:
//   - username: the user to authenticate as to the db server
//   - password: the password to authenticate with when challenged by the db server
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's credential option
func MongoClientCredentials(username string, password string) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		creds := options.Credential{
			Username:    username,
			Password:    password,
			PasswordSet: true,
		}

		opts.SetAuth(creds)
		return nil
	}
}

// Function `DirectConnection` provides the OptionSetter[options.ClientOptions] to indicate whether the connection is direct or not
//
// Parameters:
//   - connectDirectly: true to connect directly to a specific host
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's direct connection setting
func DirectConnection(connectDirectly bool) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetDirect(connectDirectly)
		return nil
	}
}

// Function `ReplicaSet` provides the OptionSetter[options.ClientOptions] to indicate the connection should target a replica set
//
// Parameters:
//   - replicaSetName: name of the replica set to target
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's replica set setting
func ReplicaSet(replicaSetName string) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetReplicaSet(replicaSetName)
		return nil
	}
}

// Function `MongoClientConnectionTimeout` provides the OptionSetter[options.ClientOptions] to set the connection's timeout policy
//
// Parameters:
//   - to: the duration to wait before timing out connection attempts
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's timeout
func MongoClientConnectionTimeout(to time.Duration) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetConnectTimeout(to)
		return nil
	}
}

// Function `MongoClientReadConcern` provides the OptionSetter[options.ClientOptions] to set the connection's read concern level
//
// Parameters:
//   - rc: pointer to the read concern level to follow
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's app read concern level
func MongoClientReadConcern(rc *readconcern.ReadConcern) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetReadConcern(rc)
		return nil
	}
}

// Function `MongoClientReadPreference` provides the OptionSetter[options.ClientOptions] to set the connection's read prefrences
//
// Parameters:
//   - rp: pointer to the read preferences to follow
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's read preference
func MongoClientReadPreference(rp *readpref.ReadPref) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetReadPreference(rp)
		return nil
	}
}

// Function `MongoClientWriteConcern` provides the OptionSetter[options.ClientOptions] to set the connection's write concern level
//
// Parameters:
//   - wc: pointer to the write concern level to follow
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's write concern level
func MongoClientWriteConcern(wc *writeconcern.WriteConcern) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetWriteConcern(wc)
		return nil
	}
}

// Function `MongoClientReadRetryPolicy` provides the OptionSetter[options.ClientOptions] to set the connection's retry policy for reads
//
// Parameters:
//   - retry: true to indicate reads should be retried
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's read retry policy
func MongoClientReadRetryPolicy(retry bool) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetRetryReads(retry)
		return nil
	}
}

// Function `MongoClientWriteRetryPolicy` provides the OptionSetter[options.ClientOptions] to set the connection's retry policy for writes
//
// Parameters:
//   - retry: true to indicate writes should be retried
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's write rety policy
func MongoClientWriteRetryPolicy(retry bool) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetRetryWrites(retry)
		return nil
	}
}

// Function `MongoClientTLSConfig` provides the OptionSetter[options.ClientOptions] to set the connection's TLS configuration
//
// Parameters:
//   - tls: the TLS configuration to use
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's TLS setting
func MongoClientTLSConfig(tls *tls.Config) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		opts.SetTLSConfig(tls)
		return nil
	}
}

// Function`MongoClientBSONOptions` provides the OptionSetter[options.ClientOptions] to set the connection's BSON options
//
// Parameters:
//   - flags: integer indicating which options to set as true
//
// Returns:
//   - `OptionSetter[options.ClientOptions]`: closure to set the given `options.ClientOptions` instance's BSON settings
func MongoClientBSONOptions(flags uint32) OptionSetter[options.ClientOptions] {
	return func(opts *options.ClientOptions) error {
		bsonOpts := options.BSONOptions{
			UseJSONStructTags:       flags&UseJSONTags > 0,
			ErrorOnInlineDuplicates: flags&ErrOnInlineDuplicates > 0,
			IntMinSize:              flags&UseIntMinSize > 0,
			NilMapAsEmpty:           flags&NilMapsAsEmpty > 0,
			NilSliceAsEmpty:         flags&NilSliceAsEmpty > 0,
			NilByteSliceAsEmpty:     flags&NilByteSliceAsEmpty > 0,
			OmitZeroStruct:          flags&OmitZeroStruct > 0,
			OmitEmpty:               flags&OmitEmpty > 0,
			StringifyMapKeysWithFmt: flags&StringifyMapKeyWithFmt > 0,
			AllowTruncatingDoubles:  flags&AllowTruncatingDoubles > 0,
			BinaryAsSlice:           flags&BinaryAsSlice > 0,
			DefaultDocumentM:        flags&DefaultDocumentM > 0,
			DefaultDocumentMap:      flags&DefaultDocumentMap > 0,
			ObjectIDAsHexString:     flags&ObjectIDAsHexString > 0,
			UseLocalTimeZone:        flags&UseLocalTimezone > 0,
			ZeroMaps:                flags&ZeroMaps > 0,
			ZeroStructs:             flags&ZeroStructs > 0,
		}
		opts.SetBSONOptions(&bsonOpts)
		return nil
	}
}

// Function `FindProjection` provides the OptionSetter[options.FindOptionsBuilder] to specify fields to keep/discard in a find operation
//
// Parameters:
//   - selectors: fields to keep or discard
//
// Returns:
//   - `OptionSetter[options.FindOptionsBuilder]`: closure to set the given `options.FindOptionsBuilder` instance's projection setting
func FindProjection(selectors ...bson.E) OptionSetter[options.FindOptionsBuilder] {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetProjection(mergeToMap(selectors...))
		return nil
	}
}

// Function `FindOffset` provides the OptionSetter[options.FindOptionsBuilder] to specify the number of documents to skip in a find operation
//
// Parameters:
//   - skip: number of documents to skip
//
// Returns:
//   - `OptionSetter[options.FindOptionsBuilder]`: closure to set the given `options.FindOptionsBuilder` instance's skip setting
func FindOffset(skip int64) OptionSetter[options.FindOptionsBuilder] {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetSkip(skip)
		return nil
	}
}

// Function `FindCap` provides the OptionSetter[options.FindOptionsBuilder] to specify the operation document limit
//
// Parameters:
//   - limit: number of document to limit resultset to
//
// Returns:
//   - `OptionSetter[options.FindOptionsBuilder]`: closure to set the given `options.FindOptionsBuilder` instance's limit setting
func FindCap(limit int64) OptionSetter[options.FindOptionsBuilder] {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetLimit(limit)
		return nil
	}
}

// Function `FindSortKey` provides the OptionSetter[options.FindOptionsBuilder] to specify the operation's sort key
//
// Parameters:
//   - sortBy: key to sort the resultset by
//
// Returns:
//   - `OptionSetter[options.FindOptionsBuilder]`: closure to set the given `options.FindOptionsBuilder` instance's sort setting
func FindSortKey(sortBy ...bson.E) OptionSetter[options.FindOptionsBuilder] {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetSort(mergeToSeq(sortBy...))
		return nil
	}
}

// Function `FindOneProjection` provides the OptionSetter[options.FindOneOptionsBuilder] to specify fields to keep/discard in a find operation
//
// Parameters:
//   - specs: fields to keep or discard
//
// Returns:
//   - `OptionSetter[options.FindOneOptionsBuilder]`: closure to set the given `options.FindOneOptionsBuilder` instance's projection setting
func FindOneProjection(selectors ...bson.E) OptionSetter[options.FindOneOptionsBuilder] {
	return func(opts *options.FindOneOptionsBuilder) error {
		opts.SetProjection(mergeToMap(selectors...))
		return nil
	}
}

// Function `FindOneOffset` provides the OptionSetter[options.FindOneOptionsBuilder] to specify the number of documents to skip in a find operation
//
// Parameters:
//   - skip: number of documents to skip
//
// Returns:
//   - `OptionSetter[options.FindOneOptionsBuilder]`: closure to set the given `options.FindOneOptionsBuilder` instance's skip setting
func FindOneOffset(skip int64) OptionSetter[options.FindOneOptionsBuilder] {
	return func(opts *options.FindOneOptionsBuilder) error {
		opts.SetSkip(skip)
		return nil
	}
}

// Function `FindOneSortKey` provides the OptionSetter[options.FindOneOptionsBuilder] to specify the operation's sort key
//
// Parameters:
//   - sortBy: key to sort the resultset by
//
// Returns:
//   - `OptionSetter[options.FindOneOptionsBuilder]`: closure to set the given `options.FindOneOptionsBuilder` instance's sort setting
func FindOneSortKey(sortBy ...bson.E) OptionSetter[options.FindOneOptionsBuilder] {
	return func(opts *options.FindOneOptionsBuilder) error {
		opts.SetSort(mergeToSeq(sortBy...))
		return nil
	}
}

// Function `ValidateInsertedDocuments` enforces document validation for the insert operation
//
// Parameters:
//   - validate: true to enforce validation rules, false to bypass
//
// Returns:
//   - `OptionSetter[options.InsertManyOptionsBuilder]`: closure to set the given `options.InsertManyOptionsBuilder` instance validation bypass setting
func ValidateInsertedDocuments(validate bool) OptionSetter[options.InsertManyOptionsBuilder] {
	return func(opts *options.InsertManyOptionsBuilder) error {
		opts.SetBypassDocumentValidation(!validate)
		return nil
	}
}

// Function `StopOnError` enables a fail fast policy when an insertion error occurs
//
// Parameters:
//   - failfast: true to stop writes after a failure, false will allow writes after a failure
//
// Returns:
//   - `OptionSetter[options.InsertManyOptionsBuilder]`: closure to set the given `options.InsertManyOptionsBuilder` instance ordered field
func StopOnError(failfast bool) OptionSetter[options.InsertManyOptionsBuilder] {
	return func(opts *options.InsertManyOptionsBuilder) error {
		opts.SetOrdered(failfast)
		return nil
	}
}

// Function `ValidateInsertedDocument` enforces document validation for the insert operation
//
// Parameters:
//   - validate: true to enforce validation rules, false to bypass
//
// Returns:
//   - `OptionSetter[options.InsertOneOptionsBuilder]`: closure to set the given `options.InsertOneOptionsBuilder` instance validation bypass setting
func ValidateInsertedDocument(validate bool) OptionSetter[options.InsertOneOptionsBuilder] {
	return func(opts *options.InsertOneOptionsBuilder) error {
		opts.SetBypassDocumentValidation(!validate)
		return nil
	}
}

// Function `ValidateUpdatedDocuments` enforces document validation for the update operation
// Parameters:
//   - validate: true to enforce validation rules, false to bypass
//
// Returns:
//   - `OptionSetter[options.UpdateManyOptionsBuilder]`: closure to set the given `options.UpdateManyOptionsBuilder` instance validation bypass setting
func ValidateUpdatedDocuments(validate bool) OptionSetter[options.UpdateManyOptionsBuilder] {
	return func(opts *options.UpdateManyOptionsBuilder) error {
		opts.SetBypassDocumentValidation(!validate)
		return nil
	}
}

// Function `DoInsertsOnNoMatchFound` enables upsert behavior for cases when the filter does not find matching documents
// Parameters:
//   - upsert: true to allow upsertion, false to not apply updates
//
// Returns:
//   - `OptionSetter[options.UpdateManyOptionsBuilder]`: closure to set the given `options.UpdateManyOptionsBuilder` instance upsert setting
func DoInsertsOnNoMatchFound(upsert bool) OptionSetter[options.UpdateManyOptionsBuilder] {
	return func(opts *options.UpdateManyOptionsBuilder) error {
		opts.SetUpsert(upsert)
		return nil
	}
}

// Function `UpdateArrayElementFilter` specifies which elements of the array to apply an update to
// Parameters:
//   - ...filters: matching conditions to determine if an update should apply to an array element
//
// Returns:
//   - `OptionSetter[options.UpdateManyOptionsBuilder]`: closure to set the given `options.UpdateManyOptionsBuilder` array filter setting
func UpdateArrayElementFilter(filters ...any) OptionSetter[options.UpdateManyOptionsBuilder] {
	return func(opts *options.UpdateManyOptionsBuilder) error {
		opts.SetArrayFilters(filters)
		return nil
	}
}

// Function `ValidateUpdatedDocument` enforces document validation for the update one operation
// Parameters:
//   - validate: true to enforce validation rules, false to bypass
//
// Returns:
//   - `OptionSetter[options.UpdateOneOptionsBuilder]`: closure to set the given `options.UpdateOneOptionsBuilder` instance validation bypass setting
func ValidateUpdatedDocument(validate bool) OptionSetter[options.UpdateOneOptionsBuilder] {
	return func(opts *options.UpdateOneOptionsBuilder) error {
		opts.SetBypassDocumentValidation(!validate)
		return nil
	}
}

// Function `DoInsertOnNoMatchFound` enables upsert behavior for cases when the filter does not find a matching document
// Parameters:
//   - upsert: true to allow upsertion, false to not apply updates
//
// Returns:
//   - `OptionSetter[options.UpdateOneOptionsBuilder]`: closure to set the given `options.UpdateOneOptionsBuilder` instance upsert setting
func DoInsertOnNoMatchFound(upsert bool) OptionSetter[options.UpdateOneOptionsBuilder] {
	return func(opts *options.UpdateOneOptionsBuilder) error {
		opts.SetUpsert(upsert)
		return nil
	}
}

// Function `UpdateOneArrayElementFilter` specifies which elements of the array to apply an update to
// Parameters:
//   - ...filters: matching conditions to determine if an update should apply to an array element
//
// Returns:
//   - `OptionSetter[options.UpdateOneOptionsBuilder]`: closure to set the given `options.UpdateOneOptionsBuilder` array filter setting
func UpdateOneArrayElementFilter(filters ...any) OptionSetter[options.UpdateOneOptionsBuilder] {
	return func(opts *options.UpdateOneOptionsBuilder) error {
		opts.SetArrayFilters(filters)
		return nil
	}
}

// Function `MinioStaticCredentials` provides the option setter to utilized the provided static credentials
//
// Parameters:
//   - accessKey: the access ID for accessing the storage service
//   - secretKey: the secret for responding to authentication requests
//
// Returns:
//   - `OptionSetter[minio.Optioons]`: functional setter to use the provided credentials
func MinioStaticCredentials(accessKey string, secretKey string) OptionSetter[minio.Options] {
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
//   - `OptionSetter[minio.Optioons]`: functional setter to use the secure conection option
func MinioUseSecureConnection(secure bool) OptionSetter[minio.Options] {
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
//   - `OptionSetter[minio.Optioons]`: functional setter to use the retry limit option
func MinioMaxRetries(retryLimit int) OptionSetter[minio.Options] {
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
//   - `OptionSetter[minio.PutObjectOptions]`: functional setter to use the user metadata option
func PutObjectMetadata(values ...string) OptionSetter[minio.PutObjectOptions] {
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
//   - `OptionSetter[minio.PutObjectOptions]`: functional setter to use the user tags option
func PutObjectTags(values ...string) OptionSetter[minio.PutObjectOptions] {
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
//   - `OptionSetter[minio.PutObjectOptions]`: functional setter to use the content type label option
func PutObjectContentType(contentType string) OptionSetter[minio.PutObjectOptions] {
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
//   - `OptionSetter[minio.RemoveObjectOptions]`: functional setter to use the force delete option
func DeleteObjectForced(force bool) OptionSetter[minio.RemoveObjectOptions] {
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
//   - `OptionSetter[minio.RemoveObjectOptions]`: functional setter to use the bypass governance option
func DeleteObjectBypassGovernancePolicy(bypass bool) OptionSetter[minio.RemoveObjectOptions] {
	return func(o *minio.RemoveObjectOptions) error {
		o.GovernanceBypass = bypass
		return nil
	}
}

// Function `mergeToMap` takes a sequence of `bson.E` instances and turns them into an associative array with chaining duplicate keys
//
// Parameters:
//   - elements: the elements to merge
//
// Returns:
//   - `bson.M`: the mapped version of the element sequence
func mergeToMap(elements ...bson.E) bson.M {
	var m bson.M = make(bson.M)

	for _, e := range elements {
		if existing, exists := m[e.Key]; exists {
			switch v := existing.(type) {
			case bson.A:
				m[e.Key] = append(v, e.Value)
			default:
				m[e.Key] = bson.A{v, e.Value}
			}
		} else {
			m[e.Key] = e.Value
		}
	}

	return m
}

// Function `mergeToSeq` takes a sequence of `bson.E` elements an orders them into a `bson.D`
//
// Parameters:
//   - elements: the elements to merge
//
// Returns:
//   - `bson.D`: the ordered sequence of elements
func mergeToSeq(elements ...bson.E) bson.D {
	var l bson.D

	for _, e := range elements {
		l = append(l, e)
	}

	return l
}
