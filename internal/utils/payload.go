package utils

/*
 * File: internal/utils/payload.go
 *
 * Purpose: API layer helpers for API response structuring and construction
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"github.com/gin-gonic/gin"
)

// Function `RespondWithRequestedData` returns a JSON mapping that indicates a successful response
//
// Paramaters:
//   - data: the data to include in the `data` field of the JSON response body
//
// Returns:
//   - `gin.H`: the mapping that can be marshalled to JSON easily with the Gin web framework
//
// Encoding:
//
//	{
//		"ok": true,
//		"data": ...
//	}
func RespondWithRequestedData(data any) gin.H {
	return gin.H{
		"ok":   true,
		"data": data,
	}
}

// Function `RespondWithError` returns a JSON mapping that indicates an unsuccessful response
//
// Paramaters:
//   - error: the message to include in the `error.message` field of the JSON response body
//   - details: additional information regarding the error
//
// Returns:
//   - `gin.H`: the mapping that can be marshalled to JSON easily with the Gin web framework
//
// Encoding:
//
//	{
//		"ok": false,
//		"error": {
//			"message": "...",
//			"details": {...}
//		}
//	}
func RespondWithError(err error, details *map[string]string) gin.H {
	if details == nil {
		return gin.H{
			"ok": false,
			"error": gin.H{
				"message": err.Error(),
			},
		}
	}
	return gin.H{
		"ok": false,
		"error": gin.H{
			"message": err.Error(),
			"details": details,
		},
	}
}
