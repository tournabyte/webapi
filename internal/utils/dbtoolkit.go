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

	"github.com/tournabyte/webapi/app"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver"
)

// Type `ConnectionOption` is a functional setter for the mongo-driver client options
type ConnectionOption func(*options.ClientOptions) error

// Type `FindOperationOption` is a functional setter for the mongo-driver find many options
type FindOperationOption func(*options.FindOptionsBuilder) error

// Type `InsertOperationOption` is a functional setter for the mongo-driver insert many options
type InsertOperationOption func(*options.InsertManyOptionsBuilder) error

// Type `UpdateOperationOption` is a functional setter for the mongo-driver update many options
type UpdateOperationOption func(*options.UpdateManyOptionsBuilder) error

// Type `DeleteOperationOption` is a functional setter for the mongo-driver delete many options
type DeleteOperationOption func(*options.DeleteManyOptionsBuilder) error

// Function `connectOptsWith` creates `options.ClientOptions` type with the given options and ensures the sanity of the configuration
// Parameters:
//   - ...opts: the configuration option functions to apply to the `options.ClientOptions` instance
//
// Returns:
//   - `*options.ClientOptions`: the validated mongo-driver client options (nil if an error occurred)
//   - `error`: the issue with connection configuration specified (nil if no validation issues are present)
func ConnectOptsWith(opts ...ConnectionOption) (*options.ClientOptions, error) {
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

// Function `findOptsWith` creates `options.FindOptionsBuilder` type with the given options
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

// Function `insertOptsWith` creates `options.InsertManyOptionsBuilder` type with the given options
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

// Function `updateOptsWith` creates `options.UpdateManyOptionsBuilder` type with the given options
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

// Function `deleteOptsWith` creates `options.DeleteManyOptionsBuilder` type with the given options
// Parameters:
//   - ...opts: the configuration option functions to apply to the `options.DeleteManyOptionsBuilder` instance
//
// Returns:
//   - `*options.DeleteManyOptionsBuilder`: the mongo-driver delete options lister (nil if an error occurred)
//   - `error`: the issue with the delete opts configuration specified (nil if no issue occurred)
func DeleteOptsWith(opts ...DeleteOperationOption) (*options.DeleteManyOptionsBuilder, error) {
	config := options.DeleteMany()

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Function `ConnectingApplicationName` provides a ConnectOption to set the connection's application name
// Parameters:
//   - name: the application name to use for the connection option
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's app name setting
func ConnectingApplicationName(name string) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetAppName(name)
		return nil
	}
}

// Function `OperationTimeout` provides the ConnectOption to set the connection's timeout setting
// Parameters:
//   - to: max time to wait for operations (only respected if operation context does not specify a deadline)
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's timeout policy
func OperationTimeout(to time.Duration) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetTimeout(to)
		return nil
	}
}

// Function `ConnectionHosts` provides the ConnectOption to set the connection's host lists
// Parameters:
//   - ...hosts: list of hosts for db servers to try for operations
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's host list
func ConnectionHosts(hosts ...string) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetHosts(hosts)
		return nil
	}
}

// Function `MinimumPoolSize` provides the ConnectOption to set the connection's min pool size
// Parameters:
//   - sz: least number of connection that can be in the connection pool at once
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's min pool size setting
func MinimumPoolSize(sz uint64) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetMinPoolSize(sz)
		return nil
	}
}

// Function `MaximumPoolSize` provides the ConnectOption to set the connection's max pool size
// Parameters:
//   - sz: most number of connection that can be in the connection pool at once
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's max pool size setting
func MaximumPoolSize(sz uint64) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetMaxPoolSize(sz)
		return nil
	}
}

// Function `ConnectionDeployment` provides the ConnectionOption to set the connection's internal deployment
// Parameters:
//   - d: the deployment to use for the connection
//
// Returns:
//   - `ConnectionOption`: closure to set the given `options.ClientOptions` instance's internal deployment setting
func ConnectionDeployment(d driver.Deployment) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.Deployment = d
		return nil
	}
}

// Function `ConnectionCredentials` provides the ConnectOption to set the connection's authentication details
// Parameters:
//   - username: the user to authenticate as to the db server
//   - password: the password to authenticate with when challenged by the db server
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's credential option
func ConnectionCredentials(username string, password string) ConnectionOption {
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

// Function `DirectConnection` provides the ConnectOption to indicate whether the connection is direct or not
// Parameters:
//   - connectDirectly: true to connect directly to a specific host
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's direct connection setting
func DirectConnection(connectDirectly bool) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetDirect(connectDirectly)
		return nil
	}
}

// Function `ConnectTimeout` provides the ConnectOption to set the connection's timeout policy
// Parameters:
//   - to: the duration to wait before timing out connection attempts
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's timeout
func ConnectTimeout(to time.Duration) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetConnectTimeout(to)
		return nil
	}
}

// Function `HeartbeatInterval` provides the ConnectOption to set the connection's heartbeat interval
// Parameters:
//   - hbInterval: the duration to use between hearbeat health checks
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's heartbeat interval
func HeartbeatInterval(hbInterval time.Duration) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetHeartbeatInterval(hbInterval)
		return nil
	}
}

// Function `ConnectionReadConcern` provides the ConnectOption to set the connection's read concern level
// Parameters:
//   - rc: pointer to the read concern level to follow
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's app read concern level
func ConnectionReadConcern(rc *readconcern.ReadConcern) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetReadConcern(rc)
		return nil
	}
}

// Function `ConnectionReadPreference` provides the ConnectOption to set the connection's read prefrences
// Parameters:
//   - rp: pointer to the read preferences to follow
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's read preference
func ConnectionReadPreference(rp *readpref.ReadPref) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetReadPreference(rp)
		return nil
	}
}

// Function `ConnectionWriteConcern` provides the ConnectOption to set the connection's write concern level
// Parameters:
//   - wc: pointer to the write concern level to follow
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's write concern level
func ConnectionWriteConcern(wc *writeconcern.WriteConcern) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetWriteConcern(wc)
		return nil
	}
}

// Function `ConnectionReadRetryPolicy` provides the ConnectOption to set the connection's retry policy for reads
// Parameters:
//   - retry: true to indicate reads should be retried
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's read retry policy
func ConnectionReadRetryPolicy(retry bool) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetRetryReads(retry)
		return nil
	}
}

// Function `ConnectionWriteRetryPolicy` provides the ConnectOption to set the connection's retry policy for writes
// Parameters:
//   - retry: true to indicate writes should be retried
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's write rety policy
func ConnectionWriteRetryPolicy(retry bool) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetRetryWrites(retry)
		return nil
	}
}

// Function `ConnectionTLSConfig` provides the ConnectOption to set the connection's TLS configuration
// Parameters:
//   - tls: the TLS configuration to use
//
// Returns:
//   - `ConnectOption`: closure to set the given `options.ClientOptions` instance's TLS setting
func ConnectionTLSConfig(tls *tls.Config) ConnectionOption {
	return func(opts *options.ClientOptions) error {
		opts.SetTLSConfig(tls)
		return nil
	}
}

// Function `ProjectionSpecification` provides the FindOperationOption to specify fields to keep/discard in a find operation
// Parameters:
//   - specs: fields to keep or discard
//
// Returns:
//   - `FindOperationOption`: closure to set the given `options.FindOptionsBuilder` instance's projection setting
func ProjectionSpecification(selectors ...projectionSelector) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetProjection(merge(selectors...))
		return nil
	}
}

// Function `SkipFirstN` provides the FindOperationOption to specify the number of documents to skip in a find operation
// Parameters:
//   - skip: number of documents to skip
//
// Returns:
//   - `FindOperationOption`: closure to set the given `options.FindOptionsBuilder` instance's skip setting
func SkipFirstN(skip int64) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetSkip(skip)
		return nil
	}
}

// Function `MatchCountLimit` provides the FindOperationOption to specify the operation document limit
// Parameters:
//   - limit: number of document to limit resultset to
//
// Returns:
//   - `FindOperationOption`: closure to set the given `options.FindOptionsBuilder` instance's limit setting
func MatchCountLimit(limit int64) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
		opts.SetLimit(limit)
		return nil
	}
}

// Function `SortResultsBy` provides the FindOperationOption to specify the operation's sort key
// Parameters:
//   - sortBy: key to sort the resultset by
//
// Returns:
//   - `FindOperationOption`: closure to set the given `options.FindOptionsBuilder` instance's sort setting
func SortResultsBy(sortBy ...sortKey) FindOperationOption {
	return func(opts *options.FindOptionsBuilder) error {
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

// Function `AllowInsertOnNoMatchFound` enables upsert behavior for cases when the filter does not find a matching document
// Parameters:
//   - upsert: true to allow upsertion, false to not apply updates
//
// Returns:
//   - `UpdateOperationOption`: closure to set the given `options.UpdateManyOptionsBuilder` instance upsert setting
func AllowInsertOnNoMatchFound(upsert bool) UpdateOperationOption {
	return func(opts *options.UpdateManyOptionsBuilder) error {
		opts.SetUpsert(upsert)
		return nil
	}
}

// Function `ArrayElementFilter` specifies which elements of the array to apply an update to
// Parameters:
//   - ...filters: matching conditions to determine if an update should apply to an array element
//
// Returns:
//   - `UpdateOperationOption`: closure to set the given `options.UpdateManyOptionsBuilder` array filter setting
func ArrayElementFilter(filters ...filterCondition) UpdateOperationOption {
	return func(opts *options.UpdateManyOptionsBuilder) error {
		var af []any
		for _, f := range filters {
			af = append(af, f())
		}
		opts.SetArrayFilters(af)
		return nil
	}
}

// Type `sortingOrder` represents the order a sorting operation will take
type sortingOrder int

// Type `projectionState` represents the action to be taken on a field within a resultset
type projectionState int

// Type `filterConditionOperator` represents an operator to determine if a document meets a filter condition
type filterConditionOperator string

// Type `updateOperator` represents an operator to determine how a document will be modified
type updateOperator string

const (
	SortAscending  sortingOrder = 1
	SortDescending sortingOrder = -1
)

// Enumeration on projection selection states
const (
	ProjectionDiscard projectionState = 0
	ProjectionRetain  projectionState = 1
)

// Enumeration of filter condition operators
const (
	FilterGreaterThan          filterConditionOperator = "$gt"
	FilterGreaterThanOrEqualTo filterConditionOperator = "$gte"
	FilterLessThan             filterConditionOperator = "$lt"
	FilterLessThanOrEqualTo    filterConditionOperator = "$lte"
	FilterValueInArray         filterConditionOperator = "$in"
	FilterValueNotInArray      filterConditionOperator = "$nin"
	FilterLogicalAnd           filterConditionOperator = "$and"
	FilterLogicalOr            filterConditionOperator = "$or"
	FilterLogicalNot           filterConditionOperator = "$not"
	FilterFieldExists          filterConditionOperator = "$exists"
)

// Enumeration of update operator tokens
const (
	UpdateSetValue       updateOperator = "$set"
	UpdateIncrementValue updateOperator = "$inc"
	UpdateDecrementValue updateOperator = "$dec"
	UpdateMultiplyValue  updateOperator = "$mul"
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
				Key: string(FilterGreaterThan), Value: minValue,
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
				Key: string(FilterLessThan), Value: maxValue,
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
				Key: string(FilterGreaterThanOrEqualTo), Value: minValueIncluded,
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
				Key: string(FilterLessThanOrEqualTo), Value: maxValueIncluded,
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
				Key: string(FilterValueInArray), Value: values,
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
				Key: string(FilterValueNotInArray), Value: values,
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
			Key: string(FilterLogicalAnd), Value: clauses,
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
			Key: string(FilterLogicalOr), Value: clauses,
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
			Key: string(FilterLogicalNot), Value: condition(),
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
				Key: string(FilterFieldExists), Value: true,
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
				Key: string(FilterFieldExists), Value: false,
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
			Key: string(UpdateSetValue), Value: bson.E{
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
			Key: string(UpdateIncrementValue), Value: bson.E{
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
			Key: string(UpdateDecrementValue), Value: bson.E{
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
			Key: string(UpdateMultiplyValue), Value: bson.E{
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

// Function `ConnectionFromConfig` creates the connection configured by the given application configuration
// Parameters:
//   - cfg: the configuration options to use
//
// Returns:
//   - `DatabaseConnection` the resulting establish mongodb connection with client and configuration information available
//   - `error`: the issue that occurred when attempting to create the specified connection (nil if no issue occurred)
func ConnectionFromConfig(cfg *app.ApplicationOptions) (*DatabaseConnection, error) {
	return NewConnection(
		ConnectingApplicationName("Tournabyte API"),
		ConnectionCredentials(cfg.Database.Username, cfg.Database.Password),
		ConnectionHosts(cfg.Database.Hosts...),
	)
}

// Function `NewConnection` creates the connection configured as specified by the `options` member and confirms the connection can be established
// Parameters:
//   - ...opts: variadic sequence of options to specify how the connection should be configured
//
// Returns:
//   - `DatabaseConnection`: the resulting established mongodb connection with client and configuration information available for lifecycle management
//   - `error`: the issue that occurred when attempting to create the specified connection (nil if no issue occurred)
func NewConnection(opts ...ConnectionOption) (*DatabaseConnection, error) {
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
