package auth

/*
 * File: internal/domains/auth/service.go
 *
 * Purpose: implementing authentication and authorization business logic as functions
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"crypto/rand"
	"log/slog"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/carlmjohnson/truthy"
	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/tournabyte/webapi/internal/domains/user"
	"github.com/tournabyte/webapi/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Register(ctx *gin.Context, src *NewUserRequest) utils.MongoOperationFunc {
	account := user.NewAccountFromRequest(ctx, src.Email, src.DisplayName, src.Password)
	slog.Debug("New account instance", slog.String("email", src.Email), slog.String("display name", src.DisplayName))

	return func(timeout context.Context, conn *mongo.Client) error {
		if truthy.ValueSlice(ctx.Errors) {
			slog.ErrorContext(timeout, "abort mongo operation: request context contains errors", slog.String("errors", ctx.Errors.String()))
			return ErrInvalidUserRegistration
		}

		cfg, cfgErr := utils.InsertOneOptsWith(utils.ValidateInsertedDocument(true))
		if cfgErr != nil {
			slog.ErrorContext(timeout, "abort mongo operation: error configuring operation options", slog.String("error", cfgErr.Error()))
			return cfgErr
		}

		_, opErr := conn.
			Database(`tournabyte`).
			Collection(`users`).
			InsertOne(timeout, account, cfg)
		if opErr != nil {
			slog.ErrorContext(timeout, "abort mongo operation: error executing operation", slog.String("error", opErr.Error()))
			return opErr
		}

		slog.DebugContext(timeout, "mongo operation completed: account record inserted", slog.String("id", account.ID.Hex()))
		return nil
	}
}

func Authenticate(ctx *gin.Context, attempt string, signer jose.Signer, oracle *user.LoginCredentials, dst *AuthenticatedUser) utils.MongoOperationFunc {
	return func(timeout context.Context, conn *mongo.Client) error {

		passwordMatches(ctx, attempt, oracle.PasswordHash)
		access, refresh := makeSessionTokens(ctx, dst.ID, signer)
		session := user.SecureSessionToken(ctx, refresh)

		if truthy.ValueSlice(ctx.Errors) {
			return ErrInvalidSession
		}

		cfg, cfgErr := utils.UpdateOneOptsWith(utils.ValidateUpdatedDocument(true), utils.DoInsertOnNoMatchFound(false))
		if cfgErr != nil {
			return cfgErr
		}

		id, idErr := bson.ObjectIDFromHex(dst.ID)
		if idErr != nil {
			return idErr
		}

		_, opErr := conn.
			Database(`tournabyte`).
			Collection(`users`).
			UpdateByID(
				timeout,
				id,
				bson.M{"$push": bson.M{"active_sessions": session}}, // #TODO: write as composed directive to avoid raw BSON
				cfg,
			)
		if opErr != nil {
			return opErr
		}

		dst.AccessToken = access
		dst.RefreshToken = refresh
		return nil
	}
}

func FindLoginDetails(ctx *gin.Context, email string, creds *user.LoginCredentials, dst *AuthenticatedUser) utils.MongoOperationFunc {
	var account user.FullAccountDetails

	return func(timeout context.Context, conn *mongo.Client) error {
		if truthy.ValueSlice(ctx.Errors) {
			return ErrInvalidSession
		}

		cfg, cfgErr := utils.FindOneOptsWith(
			utils.FindOneProjectSpec(utils.Retain("login_credentials"), utils.Retain("_id")),
		)
		if cfgErr != nil {
			return cfgErr
		}

		opErr := conn.
			Database(`tournabyte`).
			Collection(`users`).
			FindOne(
				timeout,
				utils.Directives(
					utils.Eq("login_credentials.email", email),
				),
				cfg,
			).Decode(&account)
		if opErr != nil {
			return opErr
		}

		creds.Email = account.Credentials.Email
		creds.PasswordHash = account.Credentials.PasswordHash
		dst.ID = account.ID.Hex()
		return nil
	}
}

func makeSessionTokens(ctx *gin.Context, userid string, signer jose.Signer) (string, string) {
	issueTime := time.Now().UTC()
	public := jwt.Claims{
		Issuer:    "api.tournabyte.com",
		Subject:   "Tournabyte API authorization",
		Expiry:    jwt.NewNumericDate(issueTime.Add(10 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(issueTime),
		NotBefore: jwt.NewNumericDate(issueTime.Add(15 * time.Second)),
	}

	custom := utils.AuthenticationTokenClaims{
		Owner: userid,
	}

	raw, err := jwt.Signed(signer).Claims(public).Claims(custom).Serialize()
	if err != nil {
		ctx.Error(err)
	}
	return raw, rand.Text()
}

func passwordMatches(ctx *gin.Context, provided string, stored string) {
	match, err := argon2id.ComparePasswordAndHash(provided, stored)
	if err != nil {
		ctx.Error(err)
	}
	if !match {
		ctx.Error(ErrWrongCredentials)
	}
}
