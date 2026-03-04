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

/*
// Function `CreateUserRecord` creates a new user record using the provided database connection and initial user details
//
// Parameters:
//   - ctx: the context controlling the lifetime of the user creation operation
//   - conn: the mongodb driver client to execute the user creation operation on
//   - userDetails: the initial user details to include in the newly created user record
//   - inserted: the location to write the newly created user ID to
//
// Returns:
//   - `error`: issue that occurred during the user creation operation (nil if no issue occurred)
func CreateUserRecord(ctx context.Context, conn *mongo.Client, signer jose.Signer, userDetails *NewUserRequest, inserted *AuthenticatedUser) error {
	slog.Info("Creating new user record")
	var account user.FullAccountDetails
	var primaryProfile user.PlayerProfile
	var loginDetails user.LoginCredentials

	account.ID = bson.NewObjectID()
	inserted.ID = account.ID.Hex()
	if err := makeSession(inserted, signer); err != nil {
		return err
	}
	var refresh user.ActiveSession
	refresh.NotValidBefore = time.Now().UTC()
	refresh.NotValidAfter = refresh.NotValidBefore.Add(72 * time.Hour)
	if refreshHash, err := argon2id.CreateHash(inserted.RefreshToken, argon2id.DefaultParams); err != nil {
		return err
	} else {
		refresh.TokenHash = refreshHash
		account.Sessions = append(account.Sessions, refresh)
	}
	if err := secureCredentials(userDetails.Email, userDetails.Password, &loginDetails); err != nil {
		return err
	}
	basicProfile(userDetails.DisplayName, &primaryProfile)
	account.Credentials = loginDetails
	account.PrimaryProfile = primaryProfile
	account.CreatedAt = time.Now().UTC()
	account.UpdatedAt = time.Now().UTC()

	res, err := conn.
		Database(`tournabyte`).
		Collection(`users`).
		InsertOne(
			ctx,
			&account,
		)

	if err != nil {
		return err
	}

	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		inserted.ID = id.Hex()
	} else {
		id := fmt.Sprint(res.InsertedID)
		id = strings.TrimPrefix(id, `ObjectID("`)
		id = strings.TrimSuffix(id, `")`)
		inserted.ID = id
	}

	inserted.Email = account.Credentials.Email
	inserted.DisplayName = account.PrimaryProfile.DisplayName

	return nil
}

// Function `ValidateLoginCredentials` validates the provides credentials to authenticate the specified user
//
// Parameters:
//   - ctx: the context controlling the lifetime of the credential check
//   - conn: the mongodb driver client to execute this credential check on
//   - tokenSigner: the JWT signing tool
//   - attempt: the authentication challenge response to verify
//   - user: the location to write authenticated details to
func ValidateLoginCredentials(ctx context.Context, conn *mongo.Client, tokenSigner jose.Signer, attempt *LoginAttempt, dst *AuthenticatedUser) error {
	slog.Info("Performing authentication challenge")
	var account user.FullAccountDetails

	err := conn.
		Database(`tournabyte`).
		Collection(`users`).
		FindOne(
			ctx,
			bson.D{{Key: "login_credentials.email", Value: attempt.AuthenticateAs}},
		).
		Decode(&account)

	if err != nil {
		slog.Error("Could not retrieve account details for user", slog.String("userID", attempt.AuthenticateAs), slog.String("err", err.Error()))
		return err
	}

	if match, err := argon2id.ComparePasswordAndHash(attempt.Passphrase, account.Credentials.PasswordHash); err != nil {
		slog.Error("Failed to compare password and hash")
		return err
	} else if !match {
		slog.Error("Password and hash comparison did not match")
		return utils.ErrInvalidLoginAttempt
	} else {
		slog.Debug("Authentication successful, creating session...")
		dst.ID = account.ID.Hex()
		dst.Email = account.Credentials.Email
		dst.DisplayName = account.PrimaryProfile.DisplayName
		if err := makeSession(dst, tokenSigner); err != nil {
			return err
		}
		var refresh user.ActiveSession
		refresh.NotValidBefore = time.Now().UTC()
		refresh.NotValidAfter = refresh.NotValidBefore.Add(72 * time.Hour)
		if refreshHash, err := argon2id.CreateHash(dst.RefreshToken, argon2id.DefaultParams); err != nil {
			return err
		} else {
			refresh.TokenHash = refreshHash
		}

		res, err := conn.
			Database(`tournabyte`).
			Collection(`users`).
			UpdateByID(
				ctx,
				account.ID,
				bson.M{"$push": bson.M{"active_sessions": refresh}},
			)
		if err != nil {
			return err
		} else if res.ModifiedCount != 1 {
			return utils.ErrInvalidLoginAttempt
		} else {
			return nil
		}
	}
}

func secureCredentials(email string, passwd string, creds *user.LoginCredentials) error {
	if hash, err := argon2id.CreateHash(passwd, argon2id.DefaultParams); err != nil {
		return err
	} else {
		creds.Email = email
		creds.PasswordHash = hash
		return nil
	}
}

func basicProfile(displayName string, profile *user.PlayerProfile) {
	profile.DisplayName = displayName
	profile.Preferences = user.ProfileSettings{Language: "en", Timezone: time.UTC.String()}
	profile.CreatedAt = time.Now().UTC()
	profile.UpdatedAt = time.Now().UTC()
}

func makeSession(user *AuthenticatedUser, signer jose.Signer) error {
	issueTime := time.Now().UTC()
	public := jwt.Claims{
		Issuer:    "api.tournabyte.com",
		Subject:   "Tournabyte API authorization",
		Expiry:    jwt.NewNumericDate(issueTime.Add(10 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(issueTime),
		NotBefore: jwt.NewNumericDate(issueTime.Add(15 * time.Second)),
	}

	custom := utils.AuthenticationTokenClaims{
		Owner: user.ID,
	}

	if raw, err := jwt.Signed(signer).Claims(public).Claims(custom).Serialize(); err != nil {
		return err
	} else {
		user.AccessToken = raw
		user.RefreshToken = rand.Text()
		return nil
	}
}
*/
