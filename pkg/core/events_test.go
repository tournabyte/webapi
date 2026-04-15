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

var (
	findEventDoc = bson.A{
		bson.M{
			"_id":         bson.NewObjectID(),
			"host":        bson.NewObjectID(),
			"status":      models.StatusPlanned,
			"name":        "Testing Tournament",
			"game":        "Rock-Paper-Scissors",
			"description": "A test tournament for rps",
		},
	}
	findEventOk = bson.D{
		{Key: "ok", Value: 1},
		{Key: "cursor", Value: bson.D{
			{Key: "id", Value: int64(0)},
			{Key: "ns", Value: "tournabyte.events"},
			{Key: "firstBatch", Value: findEventDoc},
		}},
	}
	findParticipantDoc = bson.A{
		bson.M{
			"_id":             bson.NewObjectID(),
			"display_name":    "Spock",
			"participates_in": findEventDoc[0].(bson.M)["_id"].(bson.ObjectID),
		},
	}
	findParticipantOk = bson.D{
		{Key: "ok", Value: 1},
		{Key: "cursor", Value: bson.D{
			{Key: "id", Value: int64(0)},
			{Key: "ns", Value: "tournabyte.participants"},
			{Key: "firstBatch", Value: findParticipantDoc},
		}},
	}
	listParticipantsDocs = bson.A{
		bson.M{
			"_id":             bson.NewObjectID(),
			"display_name":    "Rock",
			"participates_in": findEventDoc[0].(bson.M)["_id"].(bson.ObjectID),
		},
		bson.M{
			"_id":             bson.NewObjectID(),
			"display_name":    "Paper",
			"participates_in": findEventDoc[0].(bson.M)["_id"].(bson.ObjectID),
		},
		bson.M{
			"_id":             bson.NewObjectID(),
			"display_name":    "Scissors",
			"participates_in": findEventDoc[0].(bson.M)["_id"].(bson.ObjectID),
		},
	}
	listParticipantOk = bson.D{
		{Key: "ok", Value: 1},
		{Key: "cursor", Value: bson.D{
			{Key: "id", Value: int64(0)},
			{Key: "ns", Value: "tournabyte.participants"},
			{Key: "firstBatch", Value: listParticipantsDocs},
		}},
	}
	updateOneOk = bson.D{
		{Key: "ok", Value: 1},
		{Key: "n", Value: 1},         // matched count
		{Key: "nModified", Value: 1}, // modified count
	}
	deleteOneOk = bson.D{
		{Key: "ok", Value: 1},
		{Key: "n", Value: 1}, // matched count
	}
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

func setupWorkingEventLookupContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findEventOk,
	)

	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func setupWorkingEventModificationContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findEventOk,
		updateOneOk,
	)
	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func setupWorkingEventRemovalContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findEventOk,
		deleteOneOk,
	)
	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func setupWorkingCreateParticipantContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findEventOk,
		updateOneOk,
	)
	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func setupWorkingLookupParticipantContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findParticipantOk,
	)
	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func setupWorkingListParticipantsContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		listParticipantOk,
	)
	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func setupWorkingUpdateParticipantContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findEventOk,
		updateOneOk,
	)
	mockDb, err := dbx.NewMongoConnection(
		dbx.ConnectionDeployment(m),
	)
	require.NoError(t, err)

	ctx, err := mockDb.SetUpSession(context.Background())
	require.NoError(t, err)

	return ctx
}

func setupWorkingRemoveParticipantContext(t *testing.T) context.Context {
	t.Helper()

	m := drivertest.NewMockDeployment(
		pingResponse,
		findEventOk,
		deleteOneOk,
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

func setupWorkingEventLookupWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
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

	req := models.EventID{
		ID: bson.NewObjectID().Hex(),
	}
	header := models.AuthorizationHeaderContent{
		Token: token,
	}

	space.Set(handlerutil.RequestBindings, handlerutil.Bindings{
		URI: func(a any) error {
			outVal := reflect.ValueOf(a)
			if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
				return handlerutil.ErrNotAddressable
			}

			valVal := reflect.ValueOf(req)
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

func setupWorkingEventModificationWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
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
		Me: findEventDoc[0].(bson.M)["host"].(bson.ObjectID).Hex(),
	}
	token, err := jwt.Signed(signer).Claims(cl1).Claims(cl2).Serialize()
	require.NoError(t, err)

	req := models.EventID{
		ID: findEventDoc[0].(bson.M)["_id"].(bson.ObjectID).Hex(),
	}
	header := models.AuthorizationHeaderContent{
		Token: token,
	}
	body := models.UpdateEventRequest{
		NewStatus: "CONCLUDED",
	}

	space.Set(handlerutil.RequestBindings, handlerutil.Bindings{
		URI: func(a any) error {
			outVal := reflect.ValueOf(a)
			if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
				return handlerutil.ErrNotAddressable
			}

			valVal := reflect.ValueOf(req)
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
	})

	space.Set(authTokenOptionsKey, tokenOpts)
	space.Set(models.ValidatorObjectKey, validator.New())

	return &space

}

func setupWorkingEventRemovalWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
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
		Me: findEventDoc[0].(bson.M)["host"].(bson.ObjectID).Hex(),
	}
	token, err := jwt.Signed(signer).Claims(cl1).Claims(cl2).Serialize()
	require.NoError(t, err)

	req := models.EventID{
		ID: findEventDoc[0].(bson.M)["_id"].(bson.ObjectID).Hex(),
	}
	header := models.AuthorizationHeaderContent{
		Token: token,
	}
	space.Set(handlerutil.RequestBindings, handlerutil.Bindings{
		URI: func(a any) error {
			outVal := reflect.ValueOf(a)
			if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
				return handlerutil.ErrNotAddressable
			}

			valVal := reflect.ValueOf(req)
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

func setupWorkingCreateParticipantWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
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
		Me: findEventDoc[0].(bson.M)["host"].(bson.ObjectID).Hex(),
	}
	token, err := jwt.Signed(signer).Claims(cl1).Claims(cl2).Serialize()
	require.NoError(t, err)

	req := models.EventID{
		ID: findEventDoc[0].(bson.M)["_id"].(bson.ObjectID).Hex(),
	}
	header := models.AuthorizationHeaderContent{
		Token: token,
	}
	body := models.CreateOrModifyParticipantRequest{
		DisplayName: "Spock",
	}

	space.Set(handlerutil.RequestBindings, handlerutil.Bindings{
		URI: func(a any) error {
			outVal := reflect.ValueOf(a)
			if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
				return handlerutil.ErrNotAddressable
			}

			valVal := reflect.ValueOf(req)
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
	})

	space.Set(authTokenOptionsKey, tokenOpts)
	space.Set(models.ValidatorObjectKey, validator.New())

	return &space
}

func setupWorkingListParticipantsWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
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
		Me: findEventDoc[0].(bson.M)["host"].(bson.ObjectID).Hex(),
	}
	token, err := jwt.Signed(signer).Claims(cl1).Claims(cl2).Serialize()
	require.NoError(t, err)

	req := models.EventID{
		ID: findEventDoc[0].(bson.M)["_id"].(bson.ObjectID).Hex(),
	}
	header := models.AuthorizationHeaderContent{
		Token: token,
	}

	space.Set(handlerutil.RequestBindings, handlerutil.Bindings{
		URI: func(a any) error {
			outVal := reflect.ValueOf(a)
			if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
				return handlerutil.ErrNotAddressable
			}

			valVal := reflect.ValueOf(req)
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

func setupWorkingLookupParticipantWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
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
		Me: findEventDoc[0].(bson.M)["host"].(bson.ObjectID).Hex(),
	}
	token, err := jwt.Signed(signer).Claims(cl1).Claims(cl2).Serialize()
	require.NoError(t, err)

	player := models.ParticipantID{
		EID: findEventDoc[0].(bson.M)["_id"].(bson.ObjectID).Hex(),
		PID: findParticipantDoc[0].(bson.M)["_id"].(bson.ObjectID).Hex(),
	}
	header := models.AuthorizationHeaderContent{
		Token: token,
	}

	space.Set(handlerutil.RequestBindings, handlerutil.Bindings{
		URI: func(a any) error {
			outVal := reflect.ValueOf(a)
			if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
				return handlerutil.ErrNotAddressable
			}

			valVal := reflect.ValueOf(player)
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

func setupWorkingUpdateParticipantWorkspace(t *testing.T) *handlerutil.HandlerWorkspace {
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
		Me: findEventDoc[0].(bson.M)["host"].(bson.ObjectID).Hex(),
	}
	token, err := jwt.Signed(signer).Claims(cl1).Claims(cl2).Serialize()
	require.NoError(t, err)

	uri := models.ParticipantID{
		PID: findParticipantDoc[0].(bson.M)["_id"].(bson.ObjectID).Hex(),
		EID: findEventDoc[0].(bson.M)["_id"].(bson.ObjectID).Hex(),
	}
	body := models.CreateOrModifyParticipantRequest{
		DisplayName: "THE_SPOCK_RULEZ",
	}
	header := models.AuthorizationHeaderContent{
		Token: token,
	}

	space.Set(handlerutil.RequestBindings, handlerutil.Bindings{
		URI: func(a any) error {
			outVal := reflect.ValueOf(a)
			if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
				return handlerutil.ErrNotAddressable
			}

			valVal := reflect.ValueOf(uri)
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

func TestEventLookupPipeline(t *testing.T) {
	t.Run("EventLookupSuccessful", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := eventRetreivalPipeline(setupWorkingEventLookupContext(t))
		var result models.EventRecord
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingEventLookupWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(eventRecordKey, &result))

		assert.NotZero(t, result.ID)
		assert.NotZero(t, result.Host)
		assert.NotZero(t, result.Name)
		assert.NotZero(t, result.Game)
		assert.NotZero(t, result.Description)

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})
}

func TestEventUpdatePipeline(t *testing.T) {
	t.Run("EventUpdatedSuccessfully", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := eventModificiationPipeline(setupWorkingEventModificationContext(t))
		var result models.EventID
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingEventModificationWorkspace(t)

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

func TestEventDeletePipeline(t *testing.T) {
	t.Run("EventDeleteSuccessfully", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := eventDeletionPipeline(setupWorkingEventRemovalContext(t))
		var result models.EventID
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingEventRemovalWorkspace(t)

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

func TestEventParticipantPipeline(t *testing.T) {
	t.Run("CreateParticipantSuccessfully", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := createParticipantPipeline(setupWorkingCreateParticipantContext(t))
		var result models.ParticipantID
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingCreateParticipantWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(participatIDResponseKey, &result))

		assert.NotZero(t, result.EID)
		assert.NotZero(t, result.PID)

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})

	t.Run("LookupEventParticipants", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := listParticipantsPipeline(setupWorkingListParticipantsContext(t))
		var result []models.EventParticipant
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingListParticipantsWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(participantListRecordsKey, &result))

		assert.Equal(t, len(listParticipantsDocs), len(result))

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})

	t.Run("LookupEventParticipantByID", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := getParticipantPipeline(setupWorkingLookupParticipantContext(t))
		var result models.EventParticipant
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingLookupParticipantWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(participantRecordKey, &result))

		assert.NotZero(t, result.ID)
		assert.NotZero(t, result.DisplayName)
		assert.NotZero(t, result.ParticipatesIn)

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})

	t.Run("ModifyEventParticipant", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := updateParticipantPipeline(setupWorkingUpdateParticipantContext(t))
		var result models.ParticipantID
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingUpdateParticipantWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(participatIDResponseKey, &result))

		assert.NotZero(t, result.EID)
		assert.NotZero(t, result.PID)

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})

	t.Run("DeleteEventParticipant", func(t *testing.T) {
		pCtx, pCancel, pIn, pOut := removeParticipantPipeline(setupWorkingRemoveParticipantContext(t))
		var result models.ParticipantID
		defer close(pIn)
		defer pCancel(nil)

		pIn <- setupWorkingLookupParticipantWorkspace(t)

		after, ok := <-pOut
		require.True(t, ok, "Reading value from pipeline exit channel failed")
		require.NoError(t, after.Get(participatIDResponseKey, &result))

		assert.NotZero(t, result.EID)
		assert.NotZero(t, result.PID)

		select {
		case <-pCtx.Done():
			require.NoError(t, context.Cause(pCtx))
		default:
		}
	})
}
