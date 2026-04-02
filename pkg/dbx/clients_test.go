package dbx_test

/*
 * File: pkg/dbx/clients_test.go
 *
 * Purpose: unit testing for data client management
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tournabyte/webapi/pkg/dbx"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/drivertest"
)

var errOptionSetterFailure = errors.New("option setter failed")

func TestMongoClientOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		optAppName := "dbxtestcase"
		optHostList := []string{"127.0.0.1"}
		optDirect := true
		opts, err := dbx.NewOptions(
			dbx.MongoClientAppName(optAppName),
			dbx.MongoClientHosts(optHostList...),
			dbx.DirectConnection(optDirect),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)

		assert.Equal(t, optAppName, *opts.AppName)
		assert.Equal(t, len(optHostList), len(opts.Hosts))
		assert.True(t, *opts.Direct)
	})

	t.Run("BadOption", func(t *testing.T) {
		optAppName := "dbxtestcase"
		optHostList := []string{"127.0.0.1:27017", "127.0.0.1:27018"}
		optDirect := false
		opts, err := dbx.NewOptions(
			dbx.MongoClientAppName(optAppName),
			dbx.MongoClientHosts(optHostList...),
			func(co *options.ClientOptions) error { return errOptionSetterFailure },
			dbx.DirectConnection(optDirect),
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestFindOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.FindCap(5000),
			dbx.FindOffset(2500),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)
	})

	t.Run("BadOption", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.FindCap(5000),
			func(fob *options.FindOptionsBuilder) error { return errOptionSetterFailure },
			dbx.FindOffset(2500),
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestFindOneOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.FindOneOffset(9999),
			dbx.FindOneSortKey(bson.E{Key: "birth_date", Value: -1}),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)
	})

	t.Run("BadOption", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.FindOneOffset(9999),
			func(foob *options.FindOneOptionsBuilder) error { return errOptionSetterFailure },
			dbx.FindOneSortKey(bson.E{Key: "birth_date", Value: -1}),
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestInsertOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.ValidateInsertedDocuments(true),
			dbx.StopOnError(true),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)
	})

	t.Run("BadOption", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.ValidateInsertedDocuments(true),
			func(c *options.InsertManyOptionsBuilder) error { return errOptionSetterFailure },
			dbx.StopOnError(true),
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestInsertOneOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.ValidateInsertedDocument(true),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)
	})

	t.Run("BadOption", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.ValidateInsertedDocument(true),
			func(c *options.InsertOneOptionsBuilder) error { return errOptionSetterFailure },
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestUpdateOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.DoInsertsOnNoMatchFound(false),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)
	})

	t.Run("BadOption", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.DoInsertsOnNoMatchFound(false),
			func(c *options.UpdateManyOptionsBuilder) error { return errOptionSetterFailure },
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestUpdateOneOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.DoInsertOnNoMatchFound(false),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)
	})

	t.Run("BadOption", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.DoInsertOnNoMatchFound(false),
			func(c *options.UpdateOneOptionsBuilder) error { return errOptionSetterFailure },
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestMinioClientOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.MinioStaticCredentials("dbxtestcaseid", "dbxtestcasekey"),
			dbx.MinioMaxRetries(5),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)
	})

	t.Run("BadOption", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.MinioStaticCredentials("dbxtestcaseid", "dbxtestcasekey"),
			dbx.MinioMaxRetries(5),
			func(c *minio.Options) error { return errOptionSetterFailure },
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestPutObjectOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.PutObjectContentType("image/png"),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)
	})

	t.Run("BadOption", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.PutObjectContentType("image/png"),
			func(c *minio.PutObjectOptions) error { return errOptionSetterFailure },
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestRemoveObjectOptionsInstantiation(t *testing.T) {
	t.Run("ValidCombination", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.DeleteObjectForced(true),
		)

		assert.NoError(t, err)
		assert.NotNil(t, opts)
	})

	t.Run("BadOption", func(t *testing.T) {
		opts, err := dbx.NewOptions(
			dbx.DeleteObjectForced(true),
			func(c *minio.RemoveObjectOptions) error { return errOptionSetterFailure },
		)

		assert.Error(t, err) // TODO: specific error check
		assert.Nil(t, opts)
	})
}

func TestMongoConnectionLifecycle(t *testing.T) {

	t.Run("DefaultSettings", func(t *testing.T) {
		m := drivertest.NewMockDeployment(
			bson.D{{Key: "ok", Value: 1}},
		)

		db, err := dbx.NewMongoConnection(dbx.ConnectionDeployment(m))

		assert.NoError(t, err)
		assert.NotNil(t, db)
		defer db.Disconnect(context.Background())
	})

	t.Run("ModifiedSettings", func(t *testing.T) {
		m := drivertest.NewMockDeployment(
			bson.D{{Key: "ok", Value: 1}},
		)

		db, err := dbx.NewMongoConnection(
			dbx.ConnectionDeployment(m),
			dbx.MongoClientAppName("dbxtestcase"),
			dbx.MongoClientHosts("127.0.0.1:27017"),
			dbx.DirectConnection(true),
		)

		assert.NoError(t, err)
		assert.NotNil(t, db)
		defer db.Disconnect(context.Background())
	})

	t.Run("InvalidSettings", func(t *testing.T) {
		m := drivertest.NewMockDeployment(
			bson.D{{Key: "ok", Value: 1}},
		)

		db, err := dbx.NewMongoConnection(
			dbx.ConnectionDeployment(m),
			dbx.MongoClientAppName("dbxtestcase"),
			dbx.MongoClientHosts("127.0.0.1:27017", "127.0.0.1:27018"),
			dbx.DirectConnection(true),
		)

		assert.Error(t, err)
		assert.Nil(t, db)
	})

	t.Run("DeploymentUnreachable", func(t *testing.T) {
		m := drivertest.NewMockDeployment()

		db, err := dbx.NewMongoConnection(
			dbx.ConnectionDeployment(m),
			dbx.MongoClientAppName("dbxtestcase"),
			dbx.MongoClientHosts("127.0.0.1:27017"),
			dbx.DirectConnection(true),
		)

		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestMongoSessionLifecycle(t *testing.T) {
	var ctx = context.Background()
	var conn *dbx.MongoConnection

	t.Run("SetupConnection", func(t *testing.T) {
		m := drivertest.NewMockDeployment(
			bson.D{{Key: "ok", Value: 1}},
		)
		db, err := dbx.NewMongoConnection(
			dbx.ConnectionDeployment(m),
			dbx.MongoClientAppName("dbxtestcase"),
			dbx.MongoClientHosts("127.0.0.1:27017"),
			dbx.DirectConnection(true),
		)

		require.NoError(t, err)
		conn = db
	})

	t.Run("CreateAndTerminateSession", func(t *testing.T) {
		ctxWithSession, err := conn.SetUpSession(ctx)
		sess := mongo.SessionFromContext(ctxWithSession)

		require.NoError(t, err)
		require.NotNil(t, sess)

		assert.NoError(t, conn.TearDownSession(ctxWithSession))
	})

	t.Run("TeardownConnection", func(t *testing.T) {
		assert.NoError(t, conn.Disconnect(ctx))
	})
}

func TestMongoTxnLifecycle(t *testing.T) {
	var ctx = context.Background()
	var conn *dbx.MongoConnection

	t.Run("SetupConnection", func(t *testing.T) {
		m := drivertest.NewMockDeployment(
			bson.D{{Key: "ok", Value: 1}},
		)
		db, err := dbx.NewMongoConnection(
			dbx.ConnectionDeployment(m),
			dbx.MongoClientAppName("dbxtestcase"),
			dbx.MongoClientHosts("127.0.0.1:27017"),
			dbx.DirectConnection(true),
		)

		require.NoError(t, err)
		conn = db
	})

	t.Run("BeginAndCommitTransaction", func(t *testing.T) {
		ctxWithSession, err := conn.SetUpSession(ctx)
		require.NoError(t, err)
		defer conn.TearDownSession(ctxWithSession)

		require.NoError(t, conn.BeginTransaction(ctxWithSession))
		time.Sleep(time.Second)
		require.NoError(t, conn.CommitTransaction(ctxWithSession))
	})

	t.Run("BeginAndRollbackTransaction", func(t *testing.T) {
		ctxWithSession, err := conn.SetUpSession(ctx)
		require.NoError(t, err)
		defer conn.TearDownSession(ctxWithSession)

		require.NoError(t, conn.BeginTransaction(ctxWithSession))
		time.Sleep(time.Second)
		require.NoError(t, conn.AbortTransaction(ctxWithSession))
	})

	t.Run("TeardownConnection", func(t *testing.T) {
		assert.NoError(t, conn.Disconnect(ctx))
	})
}

func TestMinioConnectionLifecycle(t *testing.T) {
	conn, err := dbx.NewMinioConnection(
		"minio.example.com:9000",
		dbx.MinioMaxRetries(5),
		dbx.MinioUseSecureConnection(false),
		dbx.MinioStaticCredentials("dbxtestcaseid", "dbxtestcasekey"),
	)

	assert.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestDataClientsFromContext(t *testing.T) {
	var mongoConn *dbx.MongoConnection
	var minioConn *dbx.MinioConnection
	var setupErr error

	m := drivertest.NewMockDeployment(
		bson.D{{Key: "ok", Value: 1}},
	)
	mongoConn, setupErr = dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
		dbx.MongoClientAppName("dbxtestcase"),
		dbx.MongoClientHosts("127.0.0.1:27017"),
		dbx.DirectConnection(true),
	)

	require.NoError(t, setupErr)

	minioConn, setupErr = dbx.NewMinioConnection(
		"minio.example.com:9000",
		dbx.MinioMaxRetries(5),
		dbx.MinioUseSecureConnection(false),
		dbx.MinioStaticCredentials("dbxtestcaseid", "dbxtestcasekey"),
	)

	require.NoError(t, setupErr)

	t.Run("FoundMongoAndMinio", func(t *testing.T) {
		ctx := context.Background()
		ctx2, err := mongoConn.SetUpSession(ctx)
		require.NoError(t, err)

		ctx3 := minioConn.SetUpSession(ctx2)

		mongoSess, mongoErr := dbx.MongoFromContext(ctx3)
		require.NoError(t, mongoErr)
		assert.NotNil(t, mongoSess)

		minioSess, minioErr := dbx.MinioFromContext(ctx3)
		require.NoError(t, minioErr)
		assert.NotNil(t, minioSess)
	})

	t.Run("FoundMongoButNotMinio", func(t *testing.T) {
		ctx := context.Background()
		ctx2, err := mongoConn.SetUpSession(ctx)
		require.NoError(t, err)

		mongoSess, mongoErr := dbx.MongoFromContext(ctx2)
		require.NoError(t, mongoErr)
		assert.NotNil(t, mongoSess)

		minioSess, minioErr := dbx.MinioFromContext(ctx2)
		require.Error(t, minioErr)
		assert.Nil(t, minioSess)
	})

	t.Run("FoundMinioButNotMongo", func(t *testing.T) {
		ctx := context.Background()
		ctx2 := minioConn.SetUpSession(ctx)

		mongoSess, mongoErr := dbx.MongoFromContext(ctx2)
		require.Error(t, mongoErr)
		assert.Nil(t, mongoSess)

		minioSess, minioErr := dbx.MinioFromContext(ctx2)
		require.NoError(t, minioErr)
		assert.NotNil(t, minioSess)
	})

	t.Run("FoundNeitherMongoAndMinio", func(t *testing.T) {
		ctx := context.Background()

		minioSess, minioErr := dbx.MinioFromContext(ctx)
		require.Error(t, minioErr)
		assert.Nil(t, minioSess)

		mongoSess, mongoErr := dbx.MongoFromContext(ctx)
		require.Error(t, mongoErr)
		assert.Nil(t, mongoSess)
	})

}
