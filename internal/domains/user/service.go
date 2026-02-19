package user

/*
 * File: internal/domains/user/service.go
 *
 * Purpose: implementing user management business logic as functions
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/tournabyte/webapi/internal/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

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
func CreateUserRecord(ctx context.Context, conn *mongo.Client, userDetails *NewUserRequest, inserted *AuthenticatedUser) error {
	slog.Info("Creating new user record")
	var account FullAccountDetails
	var primaryProfile PlayerProfile
	var loginDetails LoginCredentials

	account.ID = bson.NewObjectID()
	account.Sessions = make([]ActiveSession, 0)
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
		return nil
	}

	id := fmt.Sprint(res.InsertedID)
	id = strings.TrimPrefix(id, `ObjectID("`)
	id = strings.TrimSuffix(id, `")`)
	inserted.ID = id

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
func ValidateLoginCredentials(ctx context.Context, conn *mongo.Client, tokenSigner jose.Signer, attempt *LoginAttempt, user *AuthenticatedUser) error {
	slog.Info("Performing authentication challenge")
	var account FullAccountDetails

	err := conn.
		Database(`tournabyte`).
		Collection(`users`).
		FindOne(
			ctx,
			bson.D{{Key: "login_credentials.email", Value: attempt.AuthenticateAs}},
		).
		Decode(&account)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			slog.Error("Attempted to authenciate as non-existent user", slog.String("userID", attempt.AuthenticateAs))
			return utils.NotAuthorized()
		}
		slog.Error("Could not retrieve account details for user", slog.String("userID", attempt.AuthenticateAs), slog.String("err", err.Error()))
		return err
	}

	if match, err := argon2id.ComparePasswordAndHash(attempt.Passphrase, account.Credentials.PasswordHash); err != nil {
		slog.Error("Failed to compare password and hash")
		return utils.TryAgainLater()
	} else if !match {
		slog.Error("Password and hash comparison did not match")
		return utils.NotAuthorized()
	} else {
		slog.Debug("Authentication successful, creating session...")
		user.ID = account.ID.Hex()
		user.Email = account.Credentials.Email
		user.DisplayName = account.PrimaryProfile.DisplayName
		if err := makeSession(user, tokenSigner); err != nil {
			return err
		}
		var refresh ActiveSession
		refresh.NotValidBefore = time.Now().UTC()
		refresh.NotValidAfter = refresh.NotValidBefore.Add(72 * time.Hour)
		if refreshHash, err := argon2id.CreateHash(user.RefreshToken, argon2id.DefaultParams); err != nil {
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
			return utils.TryAgainLater()
		} else {
			return nil
		}
	}
}

// Function `FindUserPrimaryProfileByEmail` retrieves the primary profile of the user identified by the given email
//
// Parameters:
//   - ctx: the context controlling the lifetime of the profile lookup operation
//   - conn: the mongodb driver client to execute this profile lookup operation on
//   - email: the email that identifies the profile being looked up
//   - dst: location to write the profile details to
//
// Returns:
//   - `error`: issue that occurred during the profile lookup operation (nil is no issue occurred)
func FindUserPrimaryProfileByEmail(ctx context.Context, conn *utils.DatabaseConnection, email string, dst *PlayerProfile) error {
	cfg, err := utils.FindOptsWith(utils.ProjectionSpecification(utils.Retain(`primary_profile`)))
	if err != nil {
		return err
	}

	cur, err := conn.
		Client().
		Database(`tournabyte`).
		Collection(`users`).
		Find(ctx, utils.Eq(`email`, email), cfg)

	if err != nil {
		return err
	}

	defer cur.Close(ctx)
	if !cur.Next(ctx) {
		return mongo.ErrNoDocuments
	}
	return cur.Decode(dst)
}

// func FindUserPrimaryProfileById(ctx context.Context, idHex string, dst *PlayerProfile) error

// func FindUserLoginByEmail(ctx context.Context, email string, dst *LoginCredentials) error

// func UpdatePrimaryProfileDetails(ctx context.Context, newDisplayName *string, newAvatarKey *string, newBio *string) error

// func UpdatePrimaryProfilePreferences(ctx context.Context, newLanguageSetting *string, newTimezoneSetting *string) error

func secureCredentials(email string, passwd string, creds *LoginCredentials) error {
	if hash, err := argon2id.CreateHash(passwd, argon2id.DefaultParams); err != nil {
		return err
	} else {
		creds.Email = email
		creds.PasswordHash = hash
		return nil
	}
}

func basicProfile(displayName string, profile *PlayerProfile) {
	profile.DisplayName = displayName
	profile.Preferences = ProfileSettings{Language: "en", Timezone: time.UTC.String()}
	profile.CreatedAt = time.Now().UTC()
	profile.UpdatedAt = time.Now().UTC()
}

func makeSession(user *AuthenticatedUser, signer jose.Signer) error {
	public := jwt.Claims{
		Issuer:   "example.com",
		Subject:  "Tournabyte API authorization",
		Expiry:   jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}

	custom := AuthenticationTokenClaims{
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
