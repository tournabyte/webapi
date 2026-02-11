package utils

/*
 * File: internal/utils/payload_test.go
 *
 * Purpose: unit tests for API layer response builders
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseDataEncoding(t *testing.T) {
	t.Run("StructuredData", func(t *testing.T) {
		data := struct {
			ANumber    int    `json:"aNumber"`
			SomeString string `json:"someString"`
		}{ANumber: 55, SomeString: "Hello!"}

		body := RespondWithRequestedData(data)
		encoded, encodeErr := json.Marshal(body)

		assert.NoError(t, encodeErr)
		assert.Contains(t, string(encoded), `"ok":true`)
		assert.Contains(t, string(encoded), `"aNumber":55`)
		assert.Contains(t, string(encoded), `"someString":"Hello!"`)
	})

	t.Run("Mapping", func(t *testing.T) {
		data := map[string]any{"aNumber": 55, "someString": "Hello!"}

		body := RespondWithRequestedData(data)
		encoded, encodeErr := json.Marshal(body)

		assert.NoError(t, encodeErr)
		assert.Contains(t, string(encoded), `"ok":true`)
		assert.Contains(t, string(encoded), `"aNumber":55`)
		assert.Contains(t, string(encoded), `"someString":"Hello!"`)
	})

	t.Run("Slice", func(t *testing.T) {
		data := []string{"A", "B", "C", "D"}

		body := RespondWithRequestedData(data)
		encoded, encodeErr := json.Marshal(body)

		assert.NoError(t, encodeErr)
		assert.Contains(t, string(encoded), `"ok":true`)
		assert.Contains(t, string(encoded), `["A","B","C","D"]`)
	})
}

func TestResponseErrorEncoding(t *testing.T) {
	t.Run("APIFailure", func(t *testing.T) {
		err := ResourceNotFound()

		body := RespondWithError(err, nil)
		encoded, encodeErr := json.Marshal(body)

		assert.NoError(t, encodeErr)
		assert.Contains(t, string(encoded), `"ok":false`)
		assert.Contains(t, string(encoded), `"message":"requested resource does not exist"`)
		assert.NotContains(t, string(encoded), `"details":`)
	})

	t.Run("APIFailureWithDetails", func(t *testing.T) {
		err := ResourceNotFound("user", "no user with ID 1")

		body := RespondWithError(err, &err.Details)
		encoded, encodeErr := json.Marshal(body)

		assert.NoError(t, encodeErr)
		assert.Contains(t, string(encoded), `"ok":false`)
		assert.Contains(t, string(encoded), `"message":"requested resource does not exist"`)
		assert.Contains(t, string(encoded), `"details":{"user":"no user with ID 1"}`)
	})

	t.Run("ArbitraryError", func(t *testing.T) {
		err := errors.New("custom error")

		body := RespondWithError(err, nil)
		encoded, encodeErr := json.Marshal(body)

		assert.NoError(t, encodeErr)
		assert.Contains(t, string(encoded), `"ok":false`)
		assert.Contains(t, string(encoded), `"message":"custom error"`)
		assert.NotContains(t, string(encoded), `"details":`)
	})
}
