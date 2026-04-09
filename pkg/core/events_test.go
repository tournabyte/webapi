package core

/*
 * File: pkg/core/event_test.go
 *
 * Purpose: unit tests for the event management logic
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
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tournabyte/webapi/pkg/dbx"
	"github.com/tournabyte/webapi/pkg/handlerutil"
	"github.com/tournabyte/webapi/pkg/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/drivertest"
)

func setupWorkingEventCreationContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
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

func setupWorkingEventCreationWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
	t.Helper()
	space := handlerutil.DefaultWorkspace()
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(`1010101010101010101010101010101010101010101010101010101010101010`)}, nil)
	require.NoError(t, err)
	tokenOpts := models.TokenOptions{
		Subject:   "testsubject",
		Issuer:    "testissuer",
		Signer:    signer,
		ExpiresIn: 5 * time.Minute,
		Key:       `1010101010101010101010101010101010101010101010101010101010101010`,
		Algorithm: "HS256",
	}
	cl1 := jwt.Claims{
		Subject:   tokenOpts.Subject,
		Issuer:    tokenOpts.Issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		Expiry:    jwt.NewNumericDate(time.Now().Add(tokenOpts.ExpiresIn)),
	}
	cl2 := models.AuthorizationTokenClaims{
		Me: bson.NewObjectID().Hex(),
	}
	token, err := jwt.Signed(signer).Claims(cl1).Claims(cl2).Serialize()
	require.NoError(t, err)

	body := models.CreateEventRequest{
		Name:        "Testing tournament",
		Game:        "Rock-Paper-Scissors",
		Description: "Test tournament for RPS",
	}
	header := models.AuthorizationHeaderContent{
		Token: token,
	}

	space.Set(handlerutil.RequestBindings, handlerutil.Bindings{
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
	})

	space.Set(authTokenOptionsKey, tokenOpts)
	space.Set(models.ValidatorObjectKey, validator.New())

	return &space
}

func TestEventCreationPipeline(t *testing.T) {
	t.Run("EventCreatedSuccessfully", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := eventCreationPipeline(setupWorkingEventCreationContext(t))
		var result models.EventID
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingEventCreationWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(eventIDResponseKey, &result))

		assert.NotZero(t, result.ID)

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})
}
