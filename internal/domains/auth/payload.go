package auth

/*
 * File: internal/domains/auth/payload.go
 *
 * Purpose: data modeling for HTTP handler request information
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

// Type `AuthorizationHeaderContent` represents a key:value pair specifically for the HTTP Authorization header
//
// Fields:
//   - Token: JWT value given as part of the HTTP authorization header
type AuthorizationHeaderContent struct {
	Token string `header:"Authorization" binding:"required"`
}

// Type `NewUserRequest` represents the minimum details required to create a user account
//
// Fields:
//   - Email: the email of the new user
//   - Password: the password of the new user (to be cryptographically secured before storing)
type AuthenticationRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}
