package dbx_test

/*
 * File: pkg/dbx/options_test.go
 *
 * Purpose: unit testing for data client customization
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/tournabyte/webapi/pkg/dbx"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

func TestApplyMongoClientOption(t *testing.T) {
	opts := options.Client()

	t.Run("AppName", func(t *testing.T) {
		appname := "option-test-application"
		setter := dbx.MongoClientAppName(appname)

		setter(opts)

		assert.Equal(t, appname, *opts.AppName)
	})

	t.Run("OperationTimeout", func(t *testing.T) {
		timeout := 5 * time.Second
		setter := dbx.MongoClientOperationTimeout(timeout)

		setter(opts)

		assert.Equal(t, timeout, *opts.Timeout)
	})

	t.Run("HostList", func(t *testing.T) {
		hosts := []string{"mongo.example.com:27017", "mongo.example.org:27017"}
		setter := dbx.MongoClientHosts(hosts...)

		setter(opts)

		assert.Equal(t, hosts, opts.Hosts)
	})

	t.Run("MinPoolSize", func(t *testing.T) {
		var sz uint64 = 5
		setter := dbx.MinimumPoolSize(sz)

		setter(opts)

		assert.Equal(t, sz, *opts.MinPoolSize)
	})

	t.Run("MaxPoolSize", func(t *testing.T) {
		var sz uint64 = 50
		setter := dbx.MaximumPoolSize(sz)

		setter(opts)

		assert.Equal(t, sz, *opts.MaxPoolSize)
	})

	t.Run("Credential", func(t *testing.T) {
		username := "option-test-user"
		password := "option-test-pass"
		setter := dbx.MongoClientCredentials(username, password)

		setter(opts)

		assert.Equal(t, username, opts.Auth.Username)
		assert.Equal(t, password, opts.Auth.Password)
		assert.True(t, opts.Auth.PasswordSet, "The crednetial password set field was not set")
	})

	t.Run("DirectConnection", func(t *testing.T) {
		y := true
		setter := dbx.DirectConnection(y)

		setter(opts)

		assert.True(t, *opts.Direct, "The direct connection field was not set")
	})

	t.Run("ReplicaSetName", func(t *testing.T) {
		rs := "option-test-cluster"
		setter := dbx.ReplicaSet(rs)

		setter(opts)

		assert.Equal(t, rs, *opts.ReplicaSet)
	})

	t.Run("ConnectTimeout", func(t *testing.T) {
		timeout := 15 * time.Second
		setter := dbx.MongoClientConnectionTimeout(timeout)

		setter(opts)

		assert.Equal(t, timeout, *opts.ConnectTimeout)
	})

	t.Run("ReadConcern", func(t *testing.T) {
		rc := readconcern.Linearizable()
		setter := dbx.MongoClientReadConcern(rc)

		setter(opts)

		assert.Equal(t, rc, opts.ReadConcern)
	})

	t.Run("ReadPreference", func(t *testing.T) {
		rp := readpref.Primary()
		setter := dbx.MongoClientReadPreference(rp)

		setter(opts)

		assert.Equal(t, rp, opts.ReadPreference)
	})

	t.Run("WriteConcern", func(t *testing.T) {
		wc := writeconcern.Majority()
		setter := dbx.MongoClientWriteConcern(wc)

		setter(opts)

		assert.Equal(t, wc, opts.WriteConcern)
	})

	t.Run("ReadRetryPolicy", func(t *testing.T) {
		setter := dbx.MongoClientReadRetryPolicy(true)

		setter(opts)

		assert.True(t, *opts.RetryReads, "The read retry policy was not set")
	})

	t.Run("WriteRetryPolicy", func(t *testing.T) {
		setter := dbx.MongoClientWriteRetryPolicy(false)

		setter(opts)

		assert.False(t, *opts.RetryWrites, "The write retry policy was not set")
	})

	t.Run("TLS", func(t *testing.T) {
		config := tls.Config{
			RootCAs:    x509.NewCertPool(),
			ServerName: "mongodb.example.com",
			MinVersion: tls.VersionTLS13,
		}
		setter := dbx.MongoClientTLSConfig(&config)

		setter(opts)

		assert.Equal(t, &config, opts.TLSConfig)

	})

	t.Run("BSONOptions", func(t *testing.T) {
		flags := dbx.NilSliceAsEmpty | dbx.NilByteSliceAsEmpty | dbx.NilMapsAsEmpty
		setter := dbx.MongoClientBSONOptions(flags)

		setter(opts)

		assert.False(t, opts.BSONOptions.UseJSONStructTags, "The use JSON tags flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.ErrorOnInlineDuplicates, "The error on inline duplicates flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.IntMinSize, "The int min size flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.OmitZeroStruct, "The omit zero struct flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.OmitEmpty, "The omit empty flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.StringifyMapKeysWithFmt, "The stringify map key with fmt flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.AllowTruncatingDoubles, "The allow truncating doubles flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.BinaryAsSlice, "The binary as slice flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.DefaultDocumentM, "The default document M flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.DefaultDocumentMap, "The default document map flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.ObjectIDAsHexString, "The object ID as hex string flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.UseLocalTimeZone, "The use local timezone flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.ZeroMaps, "The zero maps flag was unexpectedly set")
		assert.False(t, opts.BSONOptions.ZeroStructs, "The zero structs flag was unexpectedly set")

		assert.True(t, opts.BSONOptions.NilSliceAsEmpty, "The nil slice as empty flag was unexpectedly unset")
		assert.True(t, opts.BSONOptions.NilMapAsEmpty, "The nil map as empty flag was unexpectedly unset")
		assert.True(t, opts.BSONOptions.NilByteSliceAsEmpty, "The nil byte slice as empty flag was unexpectedly unset")
	})
}

func TestApplyMongoFindOperationOption(t *testing.T) {
	opts := options.Find()

	t.Run("Projection", func(t *testing.T) {
		projection := bson.D{
			bson.E{Key: "some_wanted_field", Value: true},
			bson.E{Key: "some_unwanted_field", Value: false},
		}
		setter := dbx.FindProjection(projection...)

		setter(opts)

		assert.Greater(t, len(opts.List()), 0)
	})

	t.Run("Offset", func(t *testing.T) {
		var skip int64 = 5
		setter := dbx.FindOffset(skip)

		setter(opts)

		assert.Greater(t, len(opts.List()), 1)
	})

	t.Run("Cap", func(t *testing.T) {
		var limit int64 = 50
		setter := dbx.FindCap(limit)

		setter(opts)

		assert.Greater(t, len(opts.List()), 2)
	})

	t.Run("SortKey", func(t *testing.T) {
		sorting := bson.D{
			bson.E{Key: "an_important_field", Value: 1},
			bson.E{Key: "a_less_important_field", Value: -1},
		}
		setter := dbx.FindSortKey(sorting...)

		setter(opts)
		assert.Greater(t, len(opts.List()), 3)
	})
}

func TestApplyMongoFindOneOperationOption(t *testing.T) {
	opts := options.FindOne()

	t.Run("Projection", func(t *testing.T) {
		projection := bson.D{
			bson.E{Key: "some_wanted_field", Value: true},
			bson.E{Key: "some_unwanted_field", Value: false},
		}
		setter := dbx.FindOneProjection(projection...)

		setter(opts)

		assert.Greater(t, len(opts.List()), 0)
	})

	t.Run("Offset", func(t *testing.T) {
		var skip int64 = 5
		setter := dbx.FindOneOffset(skip)

		setter(opts)

		assert.Greater(t, len(opts.List()), 1)
	})

	t.Run("SortKey", func(t *testing.T) {
		sorting := bson.D{
			bson.E{Key: "an_important_field", Value: 1},
			bson.E{Key: "a_less_important_field", Value: -1},
		}
		setter := dbx.FindOneSortKey(sorting...)

		setter(opts)
		assert.Greater(t, len(opts.List()), 2)
	})
}

func TestApplyMongoInsertOperationOption(t *testing.T) {
	opts := options.InsertMany()

	t.Run("Validate", func(t *testing.T) {
		validate := true
		setter := dbx.ValidateInsertedDocuments(validate)

		setter(opts)

		assert.Greater(t, len(opts.List()), 0)
	})

	t.Run("FailFast", func(t *testing.T) {
		failfast := true
		setter := dbx.StopOnError(failfast)

		setter(opts)

		assert.Greater(t, len(opts.List()), 1)
	})

}

func TestApplyMongoInsertOneOperationOption(t *testing.T) {
	opts := options.InsertOne()
	validate := true
	setter := dbx.ValidateInsertedDocument(validate)

	setter(opts)

	assert.Greater(t, len(opts.List()), 0)
}

func TestApplyMongoUpdateOperationOption(t *testing.T) {
	opts := options.UpdateMany()

	t.Run("Validate", func(t *testing.T) {
		validate := true
		setter := dbx.ValidateUpdatedDocuments(validate)

		setter(opts)

		assert.Greater(t, len(opts.List()), 0)
	})

	t.Run("Upsert", func(t *testing.T) {
		upsert := false
		setter := dbx.DoInsertsOnNoMatchFound(upsert)

		setter(opts)

		assert.Greater(t, len(opts.List()), 1)
	})

	t.Run("ArrayFilters", func(t *testing.T) {
		arr := []any{
			bson.E{Key: "$lt", Value: bson.E{Key: "field_to_cmp", Value: 0}},
			bson.A{1, 3, 5, 7},
		}
		setter := dbx.UpdateArrayElementFilter(arr...)

		setter(opts)

		assert.Greater(t, len(opts.List()), 2)
	})
}

func TestApplyMongoUpdateOneOperationOption(t *testing.T) {
	opts := options.UpdateOne()

	t.Run("Validate", func(t *testing.T) {
		validate := true
		setter := dbx.ValidateUpdatedDocument(validate)

		setter(opts)

		assert.Greater(t, len(opts.List()), 0)
	})

	t.Run("Upsert", func(t *testing.T) {
		upsert := false
		setter := dbx.DoInsertOnNoMatchFound(upsert)

		setter(opts)

		assert.Greater(t, len(opts.List()), 1)
	})

	t.Run("ArrayFilters", func(t *testing.T) {
		arr := []any{
			bson.E{Key: "$lt", Value: bson.E{Key: "field_to_cmp", Value: 0}},
			bson.A{1, 3, 5, 7},
		}
		setter := dbx.UpdateOneArrayElementFilter(arr...)

		setter(opts)

		assert.Greater(t, len(opts.List()), 2)
	})
}

func TestApplyMinioClientOptions(t *testing.T) {
	opts := minio.Options{}

	t.Run("StaticCredentials", func(t *testing.T) {
		username := "option-test-user"
		password := "option-test-pass"
		setter := dbx.MinioStaticCredentials(username, password)

		setter(&opts)

		assert.NotNil(t, opts.Creds, "The client credentials were unexpectedly unset")
	})

	t.Run("SecureConnection", func(t *testing.T) {
		secure := true
		setter := dbx.MinioUseSecureConnection(secure)

		setter(&opts)

		assert.True(t, opts.Secure, "The secure connection flag was unexpectedly unset")
	})

	t.Run("MaxRetries", func(t *testing.T) {
		retries := 10
		setter := dbx.MinioMaxRetries(retries)

		setter(&opts)

		assert.Equal(t, retries, opts.MaxRetries)
	})
}

func TestApplyMinioPutOption(t *testing.T) {
	opts := minio.PutObjectOptions{}

	t.Run("MetadataValid", func(t *testing.T) {
		metadata := []string{"key1", "value1", "key2", "value2"}
		setter := dbx.PutObjectMetadata(metadata...)

		setter(&opts)

		assert.Equal(t, 2, len(opts.UserMetadata))
		assert.Contains(t, opts.UserMetadata, "key1")
		assert.Contains(t, opts.UserMetadata, "key2")
	})

	t.Run("MetadataInvalid", func(t *testing.T) {
		metadata := []string{"key1", "value1", "key2", "value2", "key3"}
		setter := dbx.PutObjectMetadata(metadata...)

		assert.Error(t, setter(&opts))
	})

	t.Run("TagsValid", func(t *testing.T) {
		tags := []string{"tag1", "value1", "tag2", "value2"}
		setter := dbx.PutObjectTags(tags...)

		setter(&opts)

		assert.Equal(t, 2, len(opts.UserTags))
		assert.Contains(t, opts.UserTags, "tag1")
		assert.Contains(t, opts.UserTags, "tag2")
	})

	t.Run("TagsInvalid", func(t *testing.T) {
		tags := []string{"tag1", "value1", "tag2", "value2", "tag3"}
		setter := dbx.PutObjectTags(tags...)

		assert.Error(t, setter(&opts))
	})

	t.Run("ContentType", func(t *testing.T) {
		contentType := "image/png"
		setter := dbx.PutObjectContentType(contentType)

		setter(&opts)

		assert.Equal(t, contentType, opts.ContentType)
	})
}

func TestApplyMinioDeleteOptions(t *testing.T) {
	opts := minio.RemoveObjectOptions{}

	t.Run("Forced", func(t *testing.T) {
		forced := true
		setter := dbx.DeleteObjectForced(forced)

		setter(&opts)

		assert.True(t, opts.ForceDelete, "The forced flag was unexpectedly unset")
	})

	t.Run("BypassGovernance", func(t *testing.T) {
		bypass := true
		setter := dbx.DeleteObjectBypassGovernancePolicy(bypass)

		setter(&opts)

		assert.True(t, opts.GovernanceBypass, "The bypass flag was unexpectedly unset")
	})
}
