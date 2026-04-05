package dbx

/*
 * File: pkg/dbx/errors.go
 *
 * Purpose: propogating driver errors and presentable failures
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"github.com/tournabyte/webapi/pkg/handlerutil"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Function `IsDuplicateKeyError` determines if the given error is a mongo duplicate key error
//
// Parameters:
//   - e: the error to classify
//
// Returns:
//   - `error`: the corresponding failure message (nil if the match condition is not satisfied)
//   - `bool`: whether the match condition of the given error was satisfied
func IsDuplicateKeyError(e error) (error, bool) {
	if mongo.IsDuplicateKeyError(e) {
		return handlerutil.ErrConstraintsNotSatisfied(
			handlerutil.NewDetail("constraint violation", "operation violation unique key constraint"),
		), true
	}
	return nil, false
}
