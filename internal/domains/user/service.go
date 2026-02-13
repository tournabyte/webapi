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

func CreateUserRecord(ctx context.Context, conn *mongo.Client, userDetails *NewUser, dst *string) error {
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

	*dst = res.InsertedID.(bson.ObjectID).Hex()
	return nil
}

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

func FindUserPrimaryProfileById(ctx context.Context, idHex string, dst *PlayerProfile) error

func FindUserLoginByEmail(ctx context.Context, email string, dst *LoginCredentials) error

func UpdatePrimaryProfileDetails(ctx context.Context, newDisplayName *string, newAvatarKey *string, newBio *string) error

func UpdatePrimaryProfilePreferences(ctx context.Context, newLanguageSetting *string, newTimezoneSetting *string) error

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
