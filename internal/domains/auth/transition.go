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

	"github.com/alexedwards/argon2id"
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
	in.record.ID = bson.NewObjectID()
	in.record.LoginEmail = in.request.Email
	in.record.Metadata = utils.InitialMetadata()
	if hash, err := argon2id.CreateHash(in.request.Password, argon2id.DefaultParams); err != nil {
		return in, err
	} else {
		in.record.PasswordHash = hash
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
		InsertOne(ctx, in.record, cfg)

	if opErr != nil {
		return in, opErr
	}

	in.response.ID = in.record.ID.Hex()
	return in, nil
}
