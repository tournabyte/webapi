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
	"context"
	"reflect"
	"testing"
	"time"

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
