package core

/*
 * File: pkg/core/auth_test.go
 *
 * Purpose: unit tests for the authentication/authorization logic
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tournabyte/webapi/pkg/dbx"
	"github.com/tournabyte/webapi/pkg/handlerutil"
	"github.com/tournabyte/webapi/pkg/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/drivertest"
)

var (
	pingResponse = bson.D{{Key: "ok", Value: 1}}
	insertOk     = bson.D{{Key: "ok", Value: 1}, {Key: "n", Value: 1}}
	deleteOk     = bson.D{{Key: "ok", Value: 1}, {Key: "n", Value: 1}}
	findUserDoc  = bson.A{
		bson.M{
			"_id":           bson.NewObjectID(),
			"login_email":   "testuser@example.io",
			"password_hash": "$argon2id$v=19$m=16,t=2,p=1$YWJjZGVmZ2g$njexcOQ6BRt+mtozS/6LDg",
			"metadata": bson.M{
				"active":     true,
				"created_at": time.Now().UTC(),
				"updated_at": time.Now().UTC().Add(time.Hour),
			},
		},
	}
	findUserOk = bson.D{
		{Key: "ok", Value: 1},
		{Key: "cursor", Value: bson.D{
			{Key: "id", Value: int64(0)},
			{Key: "ns", Value: "tournabyte.users"},
			{Key: "firstBatch", Value: findUserDoc},
		}},
	}
	findSessDoc = bson.A{
		bson.M{
			"_id":              fmt.Sprintf("%x", sha256.Sum256(bytes.NewBufferString("abcdefg").Bytes())),
			"not_valid_before": time.Now().UTC().Add(-time.Hour),
			"not_valid_after":  time.Now().UTC().Add(time.Hour),
			"authorizes":       bson.NewObjectID(),
			"rotated":          false,
		},
	}
	findSessionOk = bson.D{
		{Key: "ok", Value: 1},
		{Key: "cursor", Value: bson.D{
			{Key: "id", Value: int64(0)},
			{Key: "ns", Value: "tournabyte.sessions"},
			{Key: "firstBatch", Value: findSessDoc},
		}},
	}
)

func setupWorkingUserCreationContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		insertOk,
		insertOk,
	)
	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func setupWorkingSessionRefreshContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findSessionOk,
		findUserOk,
		deleteOk,
		insertOk,
	)
	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx

}

func setupWorkingUserCreationWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
	t.Helper()
	space := handlerutil.DefaultWorkspace()
	body := models.AuthenticationRequest{
		Email:    "testuser@example.io",
		Password: "s3cr3tk3y",
	}

	space.Set(
		handlerutil.RequestBindings,
		handlerutil.Bindings{
			Body: func(a any) error {
				outVal := reflect.ValueOf(a)
				if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
					return handlerutil.ErrNotAddressable
				}

				valVal := reflect.ValueOf(body)
				if !valVal.Type().AssignableTo(outVal.Type().Elem()) {
					return handlerutil.ErrNotAssignable
				}
				outVal.Elem().Set(valVal)
				return nil
			},
		},
	)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(`1010101010101010101010101010101010101010101010101010101010101010`)}, nil)
	require.NoError(t, err)
	space.Set(authTokenOptionsKey, models.TokenOptions{
		Issuer:    "testissuer",
		Subject:   "testsubject",
		ExpiresIn: time.Minute,
		Signer:    signer,
	})
	space.Set(authSessionOptionsKey, models.SessionOptions{ExpiresIn: time.Hour})

	return &space
}

func setupWorkingSessionRefreshWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
	t.Helper()
	space := handlerutil.DefaultWorkspace()
	body := models.SessionID{
		RefreshToken: "abcdefg",
	}

	space.Set(
		handlerutil.RequestBindings,
		handlerutil.Bindings{
			Body: func(a any) error {
				outVal := reflect.ValueOf(a)
				if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
					return handlerutil.ErrNotAddressable
				}

				valVal := reflect.ValueOf(body)
				if !valVal.Type().AssignableTo(outVal.Type().Elem()) {
					return handlerutil.ErrNotAssignable
				}
				outVal.Elem().Set(valVal)
				return nil
			},
		},
	)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(`1010101010101010101010101010101010101010101010101010101010101010`)}, nil)
	require.NoError(t, err)
	space.Set(authTokenOptionsKey, models.TokenOptions{
		Issuer:    "testissuer",
		Subject:   "testsubject",
		ExpiresIn: time.Minute,
		Signer:    signer,
	})
	space.Set(authSessionOptionsKey, models.SessionOptions{ExpiresIn: time.Hour})

	return &space
}

func setupWorkingUserAuthenticationContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findUserOk,
		insertOk,
	)
	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func setupWorkingSessionCloseWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
	t.Helper()
	space := handlerutil.DefaultWorkspace()
	body := models.SessionID{
		RefreshToken: "abcdefg",
	}
	header := models.AuthorizationHeaderContent{
		Token: "abc.xyz",
	}

	space.Set(
		handlerutil.RequestBindings,
		handlerutil.Bindings{
			URI: func(a any) error {
				outVal := reflect.ValueOf(a)
				if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
					return handlerutil.ErrNotAddressable
				}

				valVal := reflect.ValueOf(body)
				if !valVal.Type().AssignableTo(outVal.Type().Elem()) {
					return handlerutil.ErrNotAssignable
				}
				outVal.Elem().Set(valVal)
				return nil
			},
			Headers: func(a any) error {
				outVal := reflect.ValueOf(a)
				if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
					return handlerutil.ErrNotAddressable
				}

				valVal := reflect.ValueOf(header)
				if !valVal.Type().AssignableTo(outVal.Type().Elem()) {
					return handlerutil.ErrNotAssignable
				}
				outVal.Elem().Set(valVal)
				return nil
			},
		},
	)

	return &space

}

func setupWorkingSessionCloseContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findSessionOk,
		deleteOk,
	)

	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func TestUserCreationPipeline(t *testing.T) {
	t.Run("UserCreatedSuccessfully", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := userCreationPipeline(setupWorkingUserCreationContext(t))
		var result models.AuthenticatedUser
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingUserCreationWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(userAuthorizationResponseKey, &result))

		assert.NotZero(t, result.ID)
		assert.NotZero(t, result.AccessToken)
		assert.NotZero(t, result.RefreshToken)

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})
}

func TestUserAuthenticationPipeline(t *testing.T) {
	t.Run("UserAuthenticatedSuccessfully", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := userAuthenticationPipeline(setupWorkingUserAuthenticationContext(t))
		var result models.AuthenticatedUser
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingUserCreationWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(userAuthorizationResponseKey, &result))

		assert.NotZero(t, result.ID)
		assert.NotZero(t, result.AccessToken)
		assert.NotZero(t, result.RefreshToken)

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})
}

func TestSessionRefreshPipeline(t *testing.T) {
	t.Run("SessionRefreshedSuccessfully", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := sessionRefreshPipeline(setupWorkingSessionRefreshContext(t))
		var result models.AuthenticatedUser
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingSessionRefreshWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(userAuthorizationResponseKey, &result))

		assert.NotZero(t, result.ID)
		assert.NotZero(t, result.AccessToken)
		assert.NotZero(t, result.RefreshToken)

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})
}

func TestSessionClosePipeline(t *testing.T) {
	t.Run("SessionClosedSuccessfully", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := sessionClosePipeline(setupWorkingSessionCloseContext(t))
		var result gin.H
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingSessionCloseWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(userLogoutResponseKey, &result))

		assert.NotZero(t, result)
		assert.Contains(t, result, "sessionClosed")

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})
}
