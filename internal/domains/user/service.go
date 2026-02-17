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
	"log/slog"
	"time"

	"github.com/alexedwards/argon2id"
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
func CreateUserRecord(ctx context.Context, conn *mongo.Client, userDetails *NewUser, inserted *string) error {
	slog.Info("Creating new user record")
	var account FullAccountDetails
	var primaryProfile PlayerProfile
	var loginDetails LoginCredentials

	account.ID = bson.NewObjectID().Hex()
	secureCredentials(userDetails.Email, userDetails.Password, &loginDetails)
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

	*inserted = res.InsertedID.(bson.ObjectID).Hex() // Panics if ID is not ObjectID (to be handled by HTTP handler)
	return nil
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

func secureCredentials(email string, passwd string, creds *LoginCredentials) {
	if hash, err := argon2id.CreateHash(passwd, argon2id.DefaultParams); err != nil {
		panic(err)
	} else {
		creds.Email = email
		creds.PasswordHash = hash
	}
}

func basicProfile(displayName string, profile *PlayerProfile) {
	profile.DisplayName = displayName
	profile.Preferences = ProfileSettings{Language: "en", Timezone: time.UTC.String()}
	profile.CreatedAt = time.Now().UTC()
	profile.UpdatedAt = time.Now().UTC()
}
