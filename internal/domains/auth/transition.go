package auth

/*
 * File: internal/domains/auth/transition.go
 *
 * Purpose: step-wise definitions to update an input state into a resulting output state
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/tournabyte/webapi/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func bindAuthenticationRequest(in AuthenticationHandlerState) (AuthenticationHandlerState, error) {
	if err := in.bindings.BindBodyAsJSON(&in.request); err != nil {
		return in, err
	}
	return in, nil
}

func deriveAccountRecordFromRequest(in AuthenticationHandlerState) (AuthenticationHandlerState, error) {
	in.account.ID = bson.NewObjectID()
	in.account.LoginEmail = in.request.Email
	in.account.Metadata = utils.InitialMetadata()
	if hash, err := argon2id.CreateHash(in.request.Password, argon2id.DefaultParams); err != nil {
		return in, err
	} else {
		in.account.PasswordHash = hash
		return in, nil
	}
}

func saveAccountRecord(in AuthenticationHandlerState) (AuthenticationHandlerState, error) {
	ctx := context.Background()
	cfg, cfgErr := utils.InsertOneOptsWith(utils.ValidateInsertedDocument(true))
	if cfgErr != nil {
		return in, cfgErr
	}

	if in.mongosess == nil {
		return in, mongo.ErrWrongClient
	}

	_, opErr := in.mongosess.Client().
		Database(UserAccountQueryContext.Database).
		Collection(UserAccountQueryContext.Collection).
		InsertOne(ctx, in.account, cfg)

	if opErr != nil {
		return in, opErr
	}

	in.response.ID = in.account.ID.Hex()
	return in, nil
}

func findAccountRecord(in AuthenticationHandlerState) (AuthenticationHandlerState, error) {
	ctx := context.Background()
	cfg, cfgErr := utils.FindOneOptsWith()
	if cfgErr != nil {
		return in, cfgErr
	}

	if in.mongosess == nil {
		return in, mongo.ErrWrongClient
	}

	filter := bson.D{
		bson.E{Key: "login_email", Value: in.request.Email},
	}

	opErr := in.mongosess.Client().
		Database(UserAccountQueryContext.Database).
		Collection(UserAccountQueryContext.Collection).
		FindOne(ctx, filter, cfg).
		Decode(&in.account)

	if opErr != nil {
		return in, opErr
	}

	return in, nil
}

func validateCredentials(in AuthenticationHandlerState) (AuthenticationHandlerState, error) {
	if match, err := argon2id.ComparePasswordAndHash(in.request.Password, in.account.PasswordHash); err != nil {
		return in, err
	} else if !match {
		return in, utils.ErrInvalidLoginAttempt
	} else {
		return in, nil
	}
}

func createAccessToken(in AuthenticationHandlerState) (AuthenticationHandlerState, error) {
	opts := in.access
	issueTime := time.Now().UTC()
	public := jwt.Claims{
		Issuer:    opts.Issuer,
		Subject:   opts.Subject,
		Expiry:    jwt.NewNumericDate(issueTime.Add(opts.ExpiresIn)),
		IssuedAt:  jwt.NewNumericDate(issueTime),
		NotBefore: jwt.NewNumericDate(issueTime),
	}

	custom := AuthorizationTokenClaims{
		Owner: in.account.ID.Hex(),
	}

	if raw, err := jwt.Signed(opts.Signer).Claims(public).Claims(custom).Serialize(); err != nil {
		return in, err
	} else {
		in.response.AccessToken = raw
		return in, err
	}
}

func createRefreshToken(in AuthenticationHandlerState) (AuthenticationHandlerState, error) {
	raw := rand.Text()
	in.response.RefreshToken = raw
	return in, nil
}

func createSessionRecord(in AuthenticationHandlerState) (AuthenticationHandlerState, error) {
	opts := in.refresh
	now := time.Now().UTC()

	if hash, err := argon2id.CreateHash(in.response.RefreshToken, argon2id.DefaultParams); err != nil {
		return in, err
	} else {
		sess := UserSession{
			ID:             hash,
			Authorizes:     in.account.ID,
			NotValidBefore: now,
			NotValidAfter:  now.Add(opts.ExpiresIn),
			Rotated:        false,
		}
		in.session = sess
		return in, nil
	}
}

func saveSessionRecord(in AuthenticationHandlerState) (AuthenticationHandlerState, error) {
	ctx := context.Background()
	cfg, cfgErr := utils.InsertOneOptsWith(utils.ValidateInsertedDocument(true))
	if cfgErr != nil {
		return in, cfgErr
	}

	if in.mongosess == nil {
		return in, mongo.ErrWrongClient
	}

	_, opErr := in.mongosess.Client().
		Database(UserSessionQueryContext.Database).
		Collection(UserSessionQueryContext.Collection).
		InsertOne(ctx, in.session, cfg)

	if opErr != nil {
		return in, opErr
	}

	return in, nil
}
