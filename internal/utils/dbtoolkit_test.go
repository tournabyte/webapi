package utils

/*
 * File: internal/utils/dbtoolkit_test.go
 *
 * Purpose: unit tests for the dbtoolkit utilities
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

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/drivertest"
)

func TestConnectWithDefaultSetting(t *testing.T) {
	m := drivertest.NewMockDeployment(
		bson.D{{Key: "ok", Value: 1}},
	)

	conn, err := NewMongoConnection(ConnectionDeployment(m))

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Disconnect(context.Background())
}

func TestConnectWithSpecifiedSetting(t *testing.T) {
	m := drivertest.NewMockDeployment(
		bson.D{{Key: "ok", Value: 1}},
	)

	conn, err := NewMongoConnection(
		ConnectionDeployment(m),
		MongoClientAppName("dbxtestcase"),
		MongoClientHosts("127.0.0.1"),
		DirectConnection(true),
	)

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Disconnect(context.Background())
}

func TestConnectWithInvalidSettingCombination(t *testing.T) {
	conn, err := NewMongoConnection(
		MongoClientAppName("dbxtestcase"),
		MongoClientHosts("127.0.0.1:27017", "127.0.0.1:27000"),
		DirectConnection(true),
	)

	assert.Nil(t, conn)
	assert.Errorf(t, err, "a direct connection cannot be made if multiple hosts are specified")
}

func TestConnectWithTimedOutPing(t *testing.T) {
	m := drivertest.NewMockDeployment()

	conn, err := NewMongoConnection(
		ConnectionDeployment(m),
		MongoClientAppName("dbxtestcase"),
		MongoClientHosts("127.0.0.1"),
		DirectConnection(true),
		MongoClientConnectionTimeout(2*time.Second),
	)

	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestConnectWithClientCreationFailure(t *testing.T) {
	badFactory := func(opts ...*options.ClientOptions) (*mongo.Client, error) {
		return nil, errors.New("client creation failure")
	}

	original := ClientFactory
	ClientFactory = badFactory
	defer func() {
		ClientFactory = original
	}()

	conn, err := NewMongoConnection(
		MongoClientAppName("dbxtestcase"),
		MongoClientHosts("127.0.0.1"),
		DirectConnection(true),
	)

	assert.Errorf(t, err, "client creation failure")
	assert.Nil(t, conn)
}
