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
//   - specs: fields to keep or discard
//
// Returns:
//   - `FindOperationOption`: closure to set the given `options.FindOptionsBuilder` instance's projection setting
func FindProjectSpec(selectors ...projectionSelector) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetProjection(merge(selectors...))
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
func FindSortKey(sortBy ...sortKey) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetSort(merge(sortBy...))
		return nil
	}
}

// Function `FindOneProjectSpec` provides the FindOneOperationOption to specify fields to keep/discard in a find operation
// Parameters:
//   - specs: fields to keep or discard
//
// Returns:
//   - `FindOneOperationOption`: closure to set the given `options.FindOneOptionsBuilder` instance's projection setting
func FindOneProjectSpec(selectors ...projectionSelector) FindOneOperationOption {
	return func(opts *options.FindOneOptionsBuilder) error {
		opts.SetProjection(merge(selectors...))
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
func FindOneSortKey(sortBy ...sortKey) FindOneOperationOption {
	return func(opts *options.FindOneOptionsBuilder) error {
		opts.SetSort(merge(sortBy...))
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
func UpdateArrayElementFilter(filters ...filterCondition) UpdateOperationOption {
	return func(opts *options.UpdateManyOptionsBuilder) error {
		var af []any
		for _, f := range filters {
			af = append(af, f())
		}
		opts.SetArrayFilters(af)
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
func UpdateOneArrayElementFilter(filters ...filterCondition) UpdateOneOperationOption {
	return func(opts *options.UpdateOneOptionsBuilder) error {
		var af []any
		for _, f := range filters {
			af = append(af, f())
		}
		opts.SetArrayFilters(af)
		return nil
	}
}

const (
	SortAscending  = 1
	SortDescending = -1
)

// Enumeration on projection selection states
const (
	ProjectionDiscard = 0
	ProjectionRetain  = 1
)

// Enumeration of filter condition operators
const (
	FilterGreaterThan          = "$gt"
	FilterGreaterThanOrEqualTo = "$gte"
	FilterLessThan             = "$lt"
	FilterLessThanOrEqualTo    = "$lte"
	FilterValueInArray         = "$in"
	FilterValueNotInArray      = "$nin"
	FilterLogicalAnd           = "$and"
	FilterLogicalOr            = "$or"
	FilterLogicalNot           = "$not"
	FilterFieldExists          = "$exists"
)

// Enumeration of update operator tokens
const (
	UpdateSetValue       = "$set"
	UpdateIncrementValue = "$inc"
	UpdateDecrementValue = "$dec"
	UpdateMultiplyValue  = "$mul"
)

// Type `sortKey` represents a sorting instruction indicating a field and an ordering direction
type sortKey func() bson.E

// Type `projectionSelector` represents an instruction to include or exclude a field from a result
type projectionSelector func() bson.E

// Type `filterCondition` represents an condition a document should meet to be included in the result set
type filterCondition func() bson.E

// Type `updateInstruction` represents an change that should be applied to a matching document
type updateInstruction func() bson.E

// Function `Asc` creates a sortKey indicating results should be sorted by the named field with lesser items coming before greater items
// Parameters:
//   - field: the name of the field the sorting should apply to
//
// Returns:
//   - `sortKey`: closure for generating the corresponding bson for the mongo-driver
func Asc(field string) sortKey {
	return func() bson.E {
		return bson.E{
			Key: field, Value: SortAscending,
		}
	}
}

// Function `Des` creates a sortKey indicating results should be sorted by the named field with greater items coming before lesser items
// Parameters:
//   - field: the name of the field the sorting should apply to
//
// Returns:
//   - `sortKey`: closure for generating the corresponding bson for the mongo-driver
func Des(field string) sortKey {
	return func() bson.E {
		return bson.E{
			Key: field, Value: SortDescending,
		}
	}
}

// Function `Discard` creates a projectionSelector indicating the results should exclude the named field
// Parameters:
//   - field: the name of the field to exclude from the results
//
// Returns:
//   - `projectionSelector`: closure for generating the corresponding bson for the mongo-driver
func Discard(field string) projectionSelector {
	return func() bson.E {
		return bson.E{
			Key: field, Value: ProjectionDiscard,
		}
	}
}

// Function `Retain` creates a projectionSelector indicating the results should include the named field
// Parameters:
//   - field: the name of the field to include from the results
//
// Returns:
//   - `projectionSelector`: closure for generating the corresponding bson for the mongo-driver
func Retain(field string) projectionSelector {
	return func() bson.E {
		return bson.E{
			Key: field, Value: ProjectionRetain,
		}
	}
}

// Function `Eq` creates a filterCondition indicating a document should contain a field with a value matching exactly as specified
// Parameters:
//   - field: the name of the field to check for equality
//   - value: the value to check against to determine equality
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func Eq(field string, value any) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: field, Value: value,
		}
	}
}

// Function `Gt` creates a filterCondition indicating a document should contain a field with a value that is comparably greater than the value specified
// Parameters:
//   - field: the name of the field to check for inequality
//   - minValue: the value to compare against to determine inequality
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func Gt(field string, minValue any) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: field, Value: bson.E{
				Key: FilterGreaterThan, Value: minValue,
			},
		}
	}
}

// Function `Lt` creates a filterCondition indicating a document should contain a field with a value that is comparably less than the value specified
// Parameters:
//   - field: the name of the field to check for inequality
//   - maxValue: the value to compare against to determine inequality
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func Lt(field string, maxValue any) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: field, Value: bson.E{
				Key: FilterLessThan, Value: maxValue,
			},
		}
	}
}

// Function `Gte` creates a filterCondition indicating a document should contain a field with a value that is comparably greater than or equal to the value specified
// Parameters:
//   - field: the name of the field to check for inequality
//   - minValueIncluded: the value to compare against to determine inequality
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func Gte(field string, minValueIncluded any) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: field, Value: bson.E{
				Key: FilterGreaterThanOrEqualTo, Value: minValueIncluded,
			},
		}
	}
}

// Function `Lte` creates a filterCondition indicating a document should contain a field with a value that is comparably less than or equal to the value specified
// Parameters:
//   - field: the name of the field to check for inequality
//   - maxValueIncluded: the value to compare against to determine inequality
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func Lte(field string, maxValueIncluded any) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: field, Value: bson.E{
				Key: FilterLessThanOrEqualTo, Value: maxValueIncluded,
			},
		}
	}
}

// Function `In` creates a filterCondition indicates a document should contain a field with a value that is contained in the given values list
// Parameters:
//   - field: the name of the field to check for inclusion
//   - values...: the set to compare against to determine inclusion
//
// Returns:
//   - `filterCondition`: closure to generating the corresponding bson for the mongo-driver
func In(field string, values ...any) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: field, Value: bson.E{
				Key: FilterValueInArray, Value: values,
			},
		}
	}
}

// Function `NotIn` creates a filterCondition indicates a document should contain a field with a value that is not contained in the given values list
// Parameters:
//   - field: the name of the field to check for exclusion
//   - values...: the set to compare against to determine exclusion
//
// Returns:
//   - `filterCondition`: closure to generating the corresponding bson for the mongo-driver
func NotIn(field string, values ...any) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: field, Value: bson.E{
				Key: FilterValueNotInArray, Value: values,
			},
		}
	}
}

// Function `And` creates a filterCondition merging an arbitrary number of filterConditions and indicates a document should match ALL of the conditions provided
// Parameters:
//   - conditions...: the conditions to merge into a logical and clause
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func And(conditions ...filterCondition) filterCondition {
	return func() bson.E {
		clauses := bson.A{}
		for _, cond := range conditions {
			clauses = append(clauses, cond())
		}
		return bson.E{
			Key: FilterLogicalAnd, Value: clauses,
		}
	}
}

// Function `Or` creates a filterCondition merging an arbitrary number of filterConditions and indicates a document should match ANY of the conditions provided
// Parameters:
//   - conditions...: the conditions to merge into a logical or clause
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func Or(conditions ...filterCondition) filterCondition {
	return func() bson.E {
		clauses := bson.A{}
		for _, cond := range conditions {
			clauses = append(clauses, cond())
		}
		return bson.E{
			Key: FilterLogicalOr, Value: clauses,
		}
	}
}

// Function `Not` creates a filterCondition that inverts the given filterCondition and indicates a document should match the opposite of the condition provided
// Parameters:
//   - condition: the condition to logically invert
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func Not(condition filterCondition) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: FilterLogicalNot, Value: condition(),
		}
	}
}

// Function `Exists` creates a filterCondition that asserts that the given field is part of a document
// Parameters:
//   - field: the field that should exist in a document
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func Exists(field string) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: field, Value: bson.E{
				Key: FilterFieldExists, Value: true,
			},
		}
	}
}

// Function `NotExists` creates a filterCondition that asserts that the given field is not part of a document
// Parameters:
//   - field: the field that should not exist in a document
//
// Returns:
//   - `filterCondition`: closure for generating the corresponding bson for the mongo-driver
func NotExists(field string) filterCondition {
	return func() bson.E {
		return bson.E{
			Key: field, Value: bson.E{
				Key: FilterFieldExists, Value: false,
			},
		}
	}
}

// Function `Set` creates an updateInstruction that sets the named field to the given value
// Parameters:
//   - field: the field that should be updated
//   - value: the value that should be used for the update
//
// Returns:
//   - `updateInstruction`: closure for generating the corresponding bson for the mongo-driver
func Set(field string, value any) updateInstruction {
	return func() bson.E {
		return bson.E{
			Key: UpdateSetValue, Value: bson.E{
				Key: field, Value: value,
			},
		}
	}
}

// Function `Increment` creates an updateInstruction that increments the named field's value by a given step
// Parameters:
//   - field: the field that should be updated
//   - step: the amount of the increment to take place
//
// Returns:
//   - `updateInstruction`: closure for generating the corresponding bson for the mongo-driver
func Increment(field string, step any) updateInstruction {
	return func() bson.E {
		return bson.E{
			Key: UpdateIncrementValue, Value: bson.E{
				Key: field, Value: step,
			},
		}
	}
}

// Function `Decrement` creates an updateInstruction that decrements the named field's value by a given step
// Parameters:
//   - field: the field that should be updated
//   - step: the amount of the decrement to take place
//
// Returns:
//   - `updateInstruction`: closure for generating the corresponding bson for the mongo-driver
func Decrement(field string, step any) updateInstruction {
	return func() bson.E {
		return bson.E{
			Key: UpdateDecrementValue, Value: bson.E{
				Key: field, Value: step,
			},
		}
	}
}

// Function `Scale` creates an updateInstruction that multiplies the named field's value by a given step
// Parameters:
//   - field: the field that should be updated
//   - step: the level of scale to take place
//
// Returns:
//   - `updateInstruction`: closure for generating the corresponding bson for the mongo-driver
func Scale(field string, step any) updateInstruction {
	return func() bson.E {
		return bson.E{
			Key: UpdateMultiplyValue, Value: bson.E{
				Key: field, Value: step,
			},
		}
	}
}

// Function `merge[T]` takes a variable number of callable items that produce `bson.E`s and merges them into a `bson.D`
// Type Parameters:
//   - T: a type who's underlying type is a parameterless function returning a `bson.E` instance
//
// Parameters:
//   - items...: the items to be merged
//
// Returns:
//   - `bson.D`: the sequence of items collected from the variadic argument
func merge[T ~func() bson.E](items ...T) bson.D {
	var itemsList bson.D

	for _, item := range items {
		itemsList = append(itemsList, item())
	}

	return itemsList
}

// Function `Directives[T]` takes a variable number of items that produce `bson.E`s and presents them back as a native go slice
// Type Parameters:
//   - T: a type who's underlying type is a parameterless function returning a `bson.E` instance
//
// Parameters:
//   - items...: the items to be included in the resulting slice
//
// Returns:
//   - `[]T` the sequence of items from the argument list presented as a slice
func Directives[T ~func() bson.E](items ...T) []T {
	return items
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

// Function `DatabaseConnection.Client` retrieves a reference to the underlying mongo-driver client instance
// Returns:
//   - `mongo.Client`: the underlying mongo-driver client instance
func (db *DatabaseConnection) Client() *mongo.Client {
	return db.client
}

// Function `DatabaseConnection.WithSession` initiates a client session and executes the sequence of operation functions within the created session
// Parameters:
//   - ctx: the context managing the lifetime of the session
//   - ...ops: sequence of `MongoOperationFunc`s to execute within the session
//
// Returns:
//   - `error`: issue encountered with session or first operation to fail
func (db *DatabaseConnection) WithSession(ctx context.Context, ops ...MongoOperationFunc) error {
	if sess, err := db.Client().StartSession(); err != nil {
		return err
	} else {
		defer sess.EndSession(ctx)
		for _, op := range ops {
			if err := op(ctx, sess.Client()); err != nil {
				return err
			}
		}
		return nil
	}
}
