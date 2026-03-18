package utils

/*
 * File: internal/utils/dbtoolkit.go
 *
 * Purpose: utilities for managing the lifecycle of the mongodb driver client
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/carlmjohnson/truthy"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver"
)

// Type `MongoOperationFunc` is an alias for functions that execute mongodb operations
type MongoOperationFunc func(context.Context, *mongo.Client) error

// Type `MongoClientOption` is a functional setter for the mongo-driver client options
type MongoClientOption func(*options.ClientOptions) error

// Type `FindOperationOption` is a functional setter for the mongo-driver find many options
type FindOperationOption func(*options.FindOptionsBuilder) error

// Type `FindOneOperationOption` is a functional setter for the mongo-driver find one options
type FindOneOperationOption func(*options.FindOneOptionsBuilder) error

// Type `InsertOperationOption` is a functional setter for the mongo-driver insert many options
type InsertOperationOption func(*options.InsertManyOptionsBuilder) error

// Type `InsertOneOperationOption` is a functional setter for the mongo-drive insert one options
type InsertOneOperationOption func(*options.InsertOneOptionsBuilder) error

// Type `UpdateOperationOption` is a functional setter for the mongo-driver update many options
type UpdateOperationOption func(*options.UpdateManyOptionsBuilder) error

// Type `UpdateOneOperationOption` is a functional setter for the mongo-driver update one options
type UpdateOneOperationOption func(*options.UpdateOneOptionsBuilder) error

// Type `DeleteOperationOption` is a functional setter for the mongo-driver delete many options
type DeleteOperationOption func(*options.DeleteManyOptionsBuilder) error

// Function `ConnectOptsWith` creates `options.ClientOptions` type with the given options and ensures the sanity of the configuration
// Parameters:
//   - ...opts: the configuration option functions to apply to the `options.ClientOptions` instance
//
// Returns:
//   - `*options.ClientOptions`: the validated mongo-driver client options (nil if an error occurred)
//   - `error`: the issue with connection configuration specified (nil if no validation issues are present)
func ConnectOptsWith(opts ...MongoClientOption) (*options.ClientOptions, error) {
	config := options.Client()

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Function `FindOptsWith` creates `options.FindOptionsBuilder` type with the given options
// Parameters:
//   - ...opts: the configuration option functions to apply to the `options.FindOptionsBuilder` instance
//
// Returns:
//   - `*options.FindOptionsBuilder`: the mongo-driver find options lister (nil if an error occurred)
//   - `error`: the issue with the findopts confiuration specified (nil if no issue occurred)
func FindOptsWith(opts ...FindOperationOption) (*options.FindOptionsBuilder, error) {
	config := options.Find()

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Function `FindOneOptsWith` creates `options.FindOneOptionsBuilder` type with the given options
// Parameters:
//   - ...opts: the configuration option function to apply to the `options.FindOneOptionsBuilder` instance
//
// Returns:
//   - `*options.FindOneOptionsBuilder`: the mongo-driver find one options lister (nil if an error occurred)
//   - `error`: the issue with the findopts configuration specified (nil if no issue occurred)
func FindOneOptsWith(opts ...FindOneOperationOption) (*options.FindOneOptionsBuilder, error) {
	config := options.FindOne()

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Function `InsertOptsWith` creates `options.InsertManyOptionsBuilder` type with the given options
// Parameters:
//   - ...opts: the configuration option functions to apply to the `options.InsertManyOptionsBuilder` instance
//
// Returns
//   - `*options.InsertManyOptionsBuilder`: the mongo-driver insert options lister (nil if an error occurred)
//   - `error`: the issue with the insertopts configuration specified (nil if no issue occurred)
func InsertOptsWith(opts ...InsertOperationOption) (*options.InsertManyOptionsBuilder, error) {
	config := options.InsertMany()

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Function `InsertOneOptsWith` creates `options.InsertOneOptionsBuilder` type with the given options
// Parameters:
//   - ...opts: the configuration option functions to apply to the `options.InsertOneOptionsBuilder` instance
//
// Returns:
//   - `*options.InsertOneOptionsBuilder`: the mongo-driver insert one options lister (nil if an error occurred)
//   - `error`: the issue with the insert opts configuration specified (nil if no issue occurred)
func InsertOneOptsWith(opts ...InsertOneOperationOption) (*options.InsertOneOptionsBuilder, error) {
	config := options.InsertOne()

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Function `UpdateOptsWith` creates `options.UpdateManyOptionsBuilder` type with the given options
// Parameters
//   - ...opts: the configuration option functions to apply to the `options.UpdateManyOptionsBuilder` instance
//
// Returns
//   - `*options.UpdateManyOptionsBuilder`: the mongo-driver update options lister (nil if an error occurred)
//   - `error`: the issue with the update opts configuration specified (nil if no issue occurred)
func UpdateOptsWith(opts ...UpdateOperationOption) (*options.UpdateManyOptionsBuilder, error) {
	config := options.UpdateMany()

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Function `UpdateOneOptsWith` creates `options.UpdateOneOptionsBuilder` type with the given options
// Parameters
//   - ...opts: the configuration option functions to apply to the `options.UpdateOneOptionsBuilder` instance
//
// Returns
//   - `*options.UpdateOneOptionsBuilder`: the mongo-driver update options lister (nil if an error occurred)
//   - `error`: the issue with the update opts configuration specified (nil if no issue occurred)
func UpdateOneOptsWith(opts ...UpdateOneOperationOption) (*options.UpdateOneOptionsBuilder, error) {
	config := options.UpdateOne()

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Function `MongoClientAppName` provides a MongoClientOption to set the connection's application name
// Parameters:
//   - name: the application name to use for the connection option
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's app name setting
func MongoClientAppName(name string) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetAppName(name)
		return nil
	}
}

// Function `MongoClientOperationTimeout` provides the MongoClientOption to set the connection's timeout setting
// Parameters:
//   - to: max time to wait for operations (only respected if operation context does not specify a deadline)
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's timeout policy
func MongoClientOperationTimeout(to time.Duration) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetTimeout(to)
		return nil
	}
}

// Function `MongoClientHosts` provides the MongoClientOption to set the connection's host lists
// Parameters:
//   - ...hosts: list of hosts for db servers to try for operations
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's host list
func MongoClientHosts(hosts ...string) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetHosts(hosts)
		return nil
	}
}

// Function `MinimumPoolSize` provides the MongoClientOption to set the connection's min pool size
// Parameters:
//   - sz: least number of connection that can be in the connection pool at once
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's min pool size setting
func MinimumPoolSize(sz uint64) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetMinPoolSize(sz)
		return nil
	}
}

// Function `MaximumPoolSize` provides the MongoClientOption to set the connection's max pool size
// Parameters:
//   - sz: most number of connection that can be in the connection pool at once
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's max pool size setting
func MaximumPoolSize(sz uint64) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetMaxPoolSize(sz)
		return nil
	}
}

// Function `ConnectionDeployment` provides the MongoClientOption to set the connection's internal deployment
// Parameters:
//   - d: the deployment to use for the connection
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's internal deployment setting
func ConnectionDeployment(d driver.Deployment) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.Deployment = d
		return nil
	}
}

// Function `MongoClientCredentials` provides the MongoClientOption to set the connection's authentication details
// Parameters:
//   - username: the user to authenticate as to the db server
//   - password: the password to authenticate with when challenged by the db server
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's credential option
func MongoClientCredentials(username string, password string) MongoClientOption {
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

// Function `DirectConnection` provides the MongoClientOption to indicate whether the connection is direct or not
// Parameters:
//   - connectDirectly: true to connect directly to a specific host
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's direct connection setting
func DirectConnection(connectDirectly bool) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetDirect(connectDirectly)
		return nil
	}
}

// Function `MongoClientConnectionTimeout` provides the MongoClientOption to set the connection's timeout policy
// Parameters:
//   - to: the duration to wait before timing out connection attempts
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's timeout
func MongoClientConnectionTimeout(to time.Duration) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetConnectTimeout(to)
		return nil
	}
}

// Function `MongoClientReadConcern` provides the MongoClientOption to set the connection's read concern level
// Parameters:
//   - rc: pointer to the read concern level to follow
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's app read concern level
func MongoClientReadConcern(rc *readconcern.ReadConcern) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetReadConcern(rc)
		return nil
	}
}

// Function `MongoClientReadPreference` provides the MongoClientOption to set the connection's read prefrences
// Parameters:
//   - rp: pointer to the read preferences to follow
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's read preference
func MongoClientReadPreference(rp *readpref.ReadPref) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetReadPreference(rp)
		return nil
	}
}

// Function `MongoClientWriteConcern` provides the MongoClientOption to set the connection's write concern level
// Parameters:
//   - wc: pointer to the write concern level to follow
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's write concern level
func MongoClientWriteConcern(wc *writeconcern.WriteConcern) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetWriteConcern(wc)
		return nil
	}
}

// Function `MongoClientReadRetryPolicy` provides the MongoClientOption to set the connection's retry policy for reads
// Parameters:
//   - retry: true to indicate reads should be retried
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's read retry policy
func MongoClientReadRetryPolicy(retry bool) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetRetryReads(retry)
		return nil
	}
}

// Function `MongoClientWriteRetryPolicy` provides the MongoClientOption to set the connection's retry policy for writes
// Parameters:
//   - retry: true to indicate writes should be retried
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's write rety policy
func MongoClientWriteRetryPolicy(retry bool) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetRetryWrites(retry)
		return nil
	}
}

// Function `MongoClientTLSConfig` provides the MongoClientOption to set the connection's TLS configuration
// Parameters:
//   - tls: the TLS configuration to use
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's TLS setting
func MongoClientTLSConfig(tls *tls.Config) MongoClientOption {
	return func(opts *options.ClientOptions) error {
		opts.SetTLSConfig(tls)
		return nil
	}
}

// Function`MongoClientBSONOptions` provides the MongoClientOption to set the connection's BSON options
// Parameters:
//   - flags: integer indicating which options to set as true
//
// Returns:
//   - `MongoClientOption`: closure to set the given `options.ClientOptions` instance's BSON settings
func MongoClientBSONOptions(flags uint32) MongoClientOption {
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

// Function `FindProjectSpec` provides the FindOperationOption to specify fields to keep/discard in a find operation
// Parameters:
//   - selectors: fields to keep or discard
//
// Returns:
//   - `FindOperationOption`: closure to set the given `options.FindOptionsBuilder` instance's projection setting
func FindProjectSpec(selectors ...bson.E) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetProjection(mergeToMap(selectors...))
		return nil
	}
}

// Function `FindOffset` provides the FindOperationOption to specify the number of documents to skip in a find operation
// Parameters:
//   - skip: number of documents to skip
//
// Returns:
//   - `FindOperationOption`: closure to set the given `options.FindOptionsBuilder` instance's skip setting
func FindOffset(skip int64) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetSkip(skip)
		return nil
	}
}

// Function `FindResultLimit` provides the FindOperationOption to specify the operation document limit
// Parameters:
//   - limit: number of document to limit resultset to
//
// Returns:
//   - `FindOperationOption`: closure to set the given `options.FindOptionsBuilder` instance's limit setting
func FindResultLimit(limit int64) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetLimit(limit)
		return nil
	}
}

// Function `FindSortKey` provides the FindOperationOption to specify the operation's sort key
// Parameters:
//   - sortBy: key to sort the resultset by
//
// Returns:
//   - `FindOperationOption`: closure to set the given `options.FindOptionsBuilder` instance's sort setting
func FindSortKey(sortBy ...bson.E) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetSort(mergeToSeq(sortBy...))
		return nil
	}
}

// Function `FindOneProjectSpec` provides the FindOneOperationOption to specify fields to keep/discard in a find operation
// Parameters:
//   - specs: fields to keep or discard
//
// Returns:
//   - `FindOneOperationOption`: closure to set the given `options.FindOneOptionsBuilder` instance's projection setting
func FindOneProjectSpec(selectors ...bson.E) FindOneOperationOption {
	return func(opts *options.FindOneOptionsBuilder) error {
		opts.SetProjection(mergeToMap(selectors...))
		return nil
	}
}

// Function `FindOneOffset` provides the FindOneOperationOption to specify the number of documents to skip in a find operation
// Parameters:
//   - skip: number of documents to skip
//
// Returns:
//   - `FindOneOperationOption`: closure to set the given `options.FindOneOptionsBuilder` instance's skip setting
func FindOneOffset(skip int64) FindOneOperationOption {
	return func(opts *options.FindOneOptionsBuilder) error {
		opts.SetSkip(skip)
		return nil
	}
}

// Function `FindOneSortKey` provides the FindOneOperationOption to specify the operation's sort key
// Parameters:
//   - sortBy: key to sort the resultset by
//
// Returns:
//   - `FindOneOperationOption`: closure to set the given `options.FindOneOptionsBuilder` instance's sort setting
func FindOneSortKey(sortBy ...bson.E) FindOneOperationOption {
	return func(opts *options.FindOneOptionsBuilder) error {
		opts.SetSort(mergeToSeq(sortBy...))
		return nil
	}
}

// Function `ValidateInsertedDocuments` enforces document validation for the insert operation
// Parameters:
//   - validate: true to enforce validation rules, false to bypass
//
// Returns:
//   - `InsertOperationOption`: closure to set the given `options.InsertManyOptionsBuilder` instance validation bypass setting
func ValidateInsertedDocuments(validate bool) InsertOperationOption {
	return func(opts *options.InsertManyOptionsBuilder) error {
		opts.SetBypassDocumentValidation(!validate)
		return nil
	}
}

// Function `StopOnError` enables a fail fast policy when an insertion error occurs
// Parameters:
//   - failfast: true to stop writes after a failure, false will allow writes after a failure
//
// Returns:
//   - `InsertOperationOption`: closure to set the given `options.InsertManyOptionsBuilder` instance ordered field
func StopOnError(failfast bool) InsertOperationOption {
	return func(opts *options.InsertManyOptionsBuilder) error {
		opts.SetOrdered(failfast)
		return nil
	}
}

// Function `ValidateInsertedDocument` enforces document validation for the insert operation
// Parameters:
//   - validate: true to enforce validation rules, false to bypass
//
// Returns:
//   - `InsertOneOperationOption`: closure to set the given `options.InsertOneOptionsBuilder` instance validation bypass setting
func ValidateInsertedDocument(validate bool) InsertOneOperationOption {
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
//   - `UpdateOperationOption`: closure to set the given `options.UpdateManyOptionsBuilder` instance validation bypass setting
func ValidateUpdatedDocuments(validate bool) UpdateOperationOption {
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
//   - `UpdateOperationOption`: closure to set the given `options.UpdateManyOptionsBuilder` instance upsert setting
func DoInsertsOnNoMatchFound(upsert bool) UpdateOperationOption {
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
//   - `UpdateOperationOption`: closure to set the given `options.UpdateManyOptionsBuilder` array filter setting
func UpdateArrayElementFilter(filters ...any) UpdateOperationOption {
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
//   - `UpdateOneOperationOption`: closure to set the given `options.UpdateOneOptionsBuilder` instance validation bypass setting
func ValidateUpdatedDocument(validate bool) UpdateOneOperationOption {
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
//   - `UpdateOneOperationOption`: closure to set the given `options.UpdateOneOptionsBuilder` instance upsert setting
func DoInsertOnNoMatchFound(upsert bool) UpdateOneOperationOption {
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
//   - `UpdateOperationOption`: closure to set the given `options.UpdateOneOptionsBuilder` array filter setting
func UpdateOneArrayElementFilter(filters ...any) UpdateOneOperationOption {
	return func(opts *options.UpdateOneOptionsBuilder) error {
		opts.SetArrayFilters(filters)
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

// Type `DocumentMetadata` represents common metatdata fields that can be attached to a document
//
// Fields:
//   - Active: indicates the document should be accounted for the active counts
//   - CreatedAt: the timestamp of document creation
//   - UpdatedAt: the timestamp of document modification
type DocumentMetadata struct {
	Active    bool      `bson:"active"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// Function `InitialMetadata` constructs a `DocumentMetadata` instance corresponding to a newly created active document
//
// Returns:
//   - `DocumentMetadata`: the metadata corresponding to a newly created active document
func InitialMetadata() DocumentMetadata {
	now := time.Now().UTC()
	return DocumentMetadata{
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Function `(*DocumentMetadata).ToggleActive` updates the active field of the metadata to the opposite boolean value
func (dm *DocumentMetadata) ToggleActive() {
	dm.Active = !dm.Active
}

// Function `(*DocumentMetadata).MarkUpdated` updates the updated at field of the metadata to the current timestamp
func (dm *DocumentMetadata) MarkUpdated() {
	dm.UpdatedAt = time.Now().UTC()
}

// Type `QueryContext` represents a mongodb database/collection pair that can be used to specify a database or collection target
//
// Fields:
//   - Database: the database name
//   - Collection: the collection name (within the database)
type QueryContext struct {
	Database   string
	Collection string
}

// Function `NewQueryContext` creates query context instances for reference when specifying database operations
//
// Parameters:
//   - db: the name of the database to target
//   - coll: the name of the collection (within the database) to target
//
// Returns:
//   - `QueryContext`: the database/collection pair for easy reference
func NewQueryContext(db string, coll string) QueryContext {
	return QueryContext{
		Database:   db,
		Collection: coll,
	}
}

// Function `defaultFactory` provides a data structure implementing the `DriverClient` interface
// This default option simply wraps the `mongo.Connect` function and returns the corresponding result
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
// Struct Members:
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
func NewMongoConnection(opts ...MongoClientOption) (*DatabaseConnection, error) {
	var conn DatabaseConnection
	var connectTimeout time.Duration

	if config, configErr := ConnectOptsWith(opts...); configErr != nil {
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

// Function `DatabaseConnection.WithSession` produces a handler function closure that sets up a mongo session within the request context
// Parameters:
//   - txn: is this session also a transaction?
//
// Returns:
//   - `gin.HandlerFunc`: a closure capable of taking part of a handlers chain
func (db *DatabaseConnection) WithSession(txn bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess, err := db.client.StartSession()
		if err != nil {
			RespondWithError(c, ErrUpstreamDataUnavailable)
			return
		}

		sessCtx := mongo.NewSessionContext(c.Request.Context(), sess)
		c.Request = c.Request.WithContext(sessCtx)
		defer sess.EndSession(sessCtx)
		if txn && sess.StartTransaction() != nil {
			RespondWithError(c, ErrUpstreamDataUnavailable)
			return
		}

		c.Next()

		if txn {
			if truthy.ValueSlice(c.Errors) {
				sess.AbortTransaction(sessCtx)
			} else {
				sess.CommitTransaction(sessCtx)
			}
		}
	}
}
