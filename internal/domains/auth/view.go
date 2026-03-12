package auth

/*
 * File: internal/domains/auth/view.go
 *
 * Purpose: data modeling for HTTP handler responses
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

// Type `AuthenticatedUser` represents the response structure for successfully authenticating as a user
//
// Fields:
//   - ID: the new user ID
//   - AccessToken: the JSON web token used for authorization to access protected resources
//   - RefreshToken: the token used for obtaining another access token once the current one has expired
type AuthenticatedUser struct {
	ID           string `json:"id"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
