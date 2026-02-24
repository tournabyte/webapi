package auth

/*
 * File: internal/domains/auth/model.go
 *
 * Purpose: data modeling for authentication and authorization
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

// Type `NewUserRequest` represents the minimum details required to create a user account
//
// Fields:
//   - Email: the email of the new user
//   - Password: the password of the new user (to be cryptographically secured before storing)
//   - DisplayName: the public display name of the new user
type NewUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"displayName" binding:"required,alphanum,min=4"`
}

// Type `LoginAttempt` represents the details needed to complete an authentication challenge
//
// Fields:
//   - AuthenticateAs: the email identifier to initiate the authentication challenge as
//   - Passphrase: the reponse to the authentication challenge
type LoginAttempt struct {
	AuthenticateAs string `json:"authenticateAs" binding:"required"`
	Passphrase     string `json:"passphrase" binding:"required"`
}

// Type `AuthenticatedUser` represents the response structure for successfully authenticating as a user
//
// Fields:
//   - ID: the new user ID
type AuthenticatedUser struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
